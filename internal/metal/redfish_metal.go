package metal

import (
	"context"
	"fmt"
	"forester/internal/config"
	"forester/internal/logging"
	"forester/internal/model"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
	"golang.org/x/exp/slog"
	"math"
	"regexp"
	"strconv"
)

type RedfishMetal struct {
	Manual bool
}

func configFromApp(ctx context.Context, app *model.Appliance) gofish.ClientConfig {
	sw := logging.SlogDualWriter{Logger: slog.Default(), Level: slog.LevelInfo, Context: ctx}
	return gofish.ClientConfig{
		Endpoint:   app.URI,
		Insecure:   true,
		DumpWriter: sw,
	}
}

func (m RedfishMetal) Enlist(ctx context.Context, app *model.Appliance, pattern string) ([]*EnlistResult, error) {
	config := configFromApp(ctx, app)

	c, err := gofish.Connect(config)
	if err != nil {
		return nil, fmt.Errorf("redfish error: %w", err)
	}
	defer c.Logout()

	rSystems, err := c.Service.Systems()
	if err != nil {
		return nil, fmt.Errorf("redfish error: %w", err)
	}

	rg, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("cannot compile regular expression '%s': %w", pattern, err)
	}

	var result []*EnlistResult
	for _, rSystem := range rSystems {
		if rg.MatchString(rSystem.ID) {
			slog.DebugContext(ctx, "found redfish system",
				"id", rSystem.ID,
				"uuid", rSystem.UUID,
				"oid", rSystem.ODataID,
			)

			var addrs []string
			interfaces, err := rSystem.EthernetInterfaces()
			if err != nil {
				return nil, fmt.Errorf("redfish error: %w", err)
			}
			for _, iface := range interfaces {
				addrs = append(addrs, iface.MACAddress)
			}

			facts := map[string]string{
				"redfish_oid":             rSystem.ODataID,
				"redfish_model":           rSystem.Model,
				"redfish_name":            rSystem.Name,
				"redfish_description":     rSystem.Description,
				"redfish_asset_tag":       rSystem.AssetTag,
				"redfish_manufacturer":    rSystem.Manufacturer,
				"redfish_part_number":     rSystem.PartNumber,
				"redfish_serial_number":   rSystem.SerialNumber,
				"redfish_sku":             rSystem.SKU,
				"redfish_pcie_dev_count":  strconv.Itoa(rSystem.PCIeDevicesCount),
				"redfish_memory_bytes":    strconv.Itoa(int(float64(rSystem.MemorySummary.TotalSystemMemoryGiB) * math.Pow(2, 30))),
				"redfish_processor_model": rSystem.ProcessorSummary.Model,
				"redfish_processor_count": strconv.Itoa(rSystem.ProcessorSummary.Count),
				"redfish_processor_cores": strconv.Itoa(rSystem.ProcessorSummary.LogicalProcessorCount),
			}

			er := &EnlistResult{
				HwAddrs: addrs,
				Facts:   facts,
				UID:     rSystem.UUID,
			}

			result = append(result, er)
		} else {
			slog.DebugContext(ctx, "redfish system does not match the pattern",
				"id", rSystem.ID,
				"uuid", rSystem.UUID,
				"oid", rSystem.ODataID,
			)
		}
	}

	return result, nil
}

func (m RedfishMetal) BootNetwork(ctx context.Context, system *model.SystemAppliance) error {
	if m.Manual {
		return nil
	}

	cfg := configFromApp(ctx, &system.Appliance)

	c, err := gofish.Connect(cfg)
	if err != nil {
		return fmt.Errorf("redfish error: %w", err)
	}
	defer c.Logout()

	rSystems, err := c.Service.Systems()
	if err != nil {
		return fmt.Errorf("redfish error: %w", err)
	}

	uri := fmt.Sprintf("%s/boot/shim.efi", config.BaseURL())
	if len(system.HwAddrs) > 0 {
		uri = fmt.Sprintf("%s/boot/%s/shim.efi", config.BaseURL(), system.HwAddrs[0].String())
	} else {
		slog.WarnContext(ctx, "no mac address found for system", "system_id", system.System.ID)
	}

	bootOverride := redfish.Boot{
		BootSourceOverrideTarget:  redfish.UefiHTTPBootSourceOverrideTarget,
		BootSourceOverrideEnabled: redfish.OnceBootSourceOverrideEnabled,
		HTTPBootURI:               uri,
	}

	for _, rSystem := range rSystems {
		if rSystem.UUID == *system.UID {
			slog.DebugContext(ctx, "found redfish system", "id", rSystem.ID, "uuid", rSystem.UUID, "uid", *system.UID)
			err := rSystem.SetBoot(bootOverride)
			if err != nil {
				return fmt.Errorf("redfish error: %w", err)
			}
			if rSystem.PowerState == redfish.OnPowerState {
				err = rSystem.Reset(redfish.PowerCycleResetType)
			} else {
				err = rSystem.Reset(redfish.OnResetType)
			}
			if err != nil {
				return fmt.Errorf("redfish error: %w", err)
			}
		} else {
			slog.DebugContext(ctx, "checking redfish system", "id", rSystem.ID, "uuid", rSystem.UUID, "uid", *system.UID)
		}
	}

	return nil
}

func (m RedfishMetal) BootLocal(ctx context.Context, system *model.SystemAppliance) error {
	if m.Manual {
		return nil
	}

	slog.InfoContext(ctx, "noop operation", "function", "BootLocal")
	return nil
}
