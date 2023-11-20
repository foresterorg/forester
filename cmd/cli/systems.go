package main

import (
	"context"
	"fmt"
	"forester/internal/api/ctl"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

type systemRegisterCmd struct {
	HwAddrs       []string          `arg:"-m,required"`
	Facts         map[string]string `arg:"-f"`
	ApplianceName string            `arg:"-a"`
	UID           string            `arg:"-u"`
}

type systemShowCmd struct {
	Pattern string `arg:"positional,required" placeholder:"MAC_OR_NAME"`
}

type systemListCmd struct {
	DisplayFacts []string `args:"-f"`
	Limit        int64    `arg:"-m" default:"100"`
	Offset       int64    `arg:"-o" default:"0"`
}

type systemKickstartCmd struct {
	Pattern string `arg:"positional,required" placeholder:"MAC_OR_NAME"`
}

type systemLogsCmd struct {
	Pattern  string `arg:"positional,required" placeholder:"MAC_OR_NAME"`
	Download string `arg:"-d"`
}

type systemAcquireCmd struct {
	Pattern  string   `arg:"positional,required" placeholder:"MAC_OR_NAME"`
	Image    string   `arg:"-i,required"`
	Snippets []string `arg:"-s"`
	Comment  string   `arg:"-c"`
}

type systemReleaseCmd struct {
	Pattern string `arg:"positional,required" placeholder:"MAC_OR_NAME"`
}

type systemBootNetworkCmd struct {
	Pattern string `arg:"positional,required" placeholder:"MAC_OR_NAME"`
}

type systemBootLocalCmd struct {
	Pattern string `arg:"positional,required" placeholder:"MAC_OR_NAME"`
}

type systemCmd struct {
	Register    *systemRegisterCmd    `arg:"subcommand:register" help:"register system"`
	List        *systemListCmd        `arg:"subcommand:list" help:"list systems"`
	Show        *systemShowCmd        `arg:"subcommand:show" help:"show system"`
	Acquire     *systemAcquireCmd     `arg:"subcommand:acquire" help:"acquire system"`
	Release     *systemReleaseCmd     `arg:"subcommand:release" help:"release system"`
	Kickstart   *systemKickstartCmd   `arg:"subcommand:kickstart" help:"show system kickstart"`
	Logs        *systemLogsCmd        `arg:"subcommand:logs" help:"show installation log history"`
	BootNetwork *systemBootNetworkCmd `arg:"subcommand:bootnet" help:"reset (hard reboot) system and boot from network"`
	BootLocal   *systemBootLocalCmd   `arg:"subcommand:bootlocal" help:"reset (hard reboot) system and boot from local drive"`
}

func systemRegister(ctx context.Context, cmdArgs *systemRegisterCmd) error {
	client := ctl.NewSystemServiceClient(args.URL, http.DefaultClient)
	sys := ctl.NewSystem{
		HwAddrs:       cmdArgs.HwAddrs,
		Facts:         cmdArgs.Facts,
		ApplianceName: &cmdArgs.ApplianceName,
		UID:           &cmdArgs.UID,
	}
	err := client.Register(ctx, &sys)
	if err != nil {
		return fmt.Errorf("cannot register system: %w", err)
	}

	return nil
}

func systemShow(ctx context.Context, cmdArgs *systemShowCmd) error {
	client := ctl.NewSystemServiceClient(args.URL, http.DefaultClient)
	result, err := client.Find(ctx, cmdArgs.Pattern)
	if err != nil {
		return fmt.Errorf("cannot find: %w", err)
	}

	w := newTabWriter()
	fmt.Fprintln(w, "Attribute\tValue")
	fmt.Fprintf(w, "%s\t%d\n", "ID", result.ID)
	fmt.Fprintf(w, "%s\t%s\n", "Name", result.Name)
	fmt.Fprintf(w, "%s\t%t\n", "Acquired", result.Acquired)
	fmt.Fprintf(w, "%s\t%s\n", "Install UUID", result.InstallUUID)
	if result.Acquired && result.ImageID != nil {
		fmt.Fprintf(w, "%s\t%s\n", "Acquired at", result.AcquiredAt.Format(time.ANSIC))
		fmt.Fprintf(w, "%s\t%d\n", "Image ID", *result.ImageID)
	}
	for _, mac := range result.HwAddrs {
		fmt.Fprintf(w, "%s\t%s\n", "MAC", mac)
	}
	if result.Appliance != nil && result.Appliance.Name != "" {
		fmt.Fprintf(w, "%s\t%s\n", "Appliance Name", result.Appliance.Name)
		fmt.Fprintf(w, "%s\t%s\n", "Appliance Kind", ctl.ApplianceIntToKind(result.Appliance.Kind))
		fmt.Fprintf(w, "%s\t%s\n", "Appliance URI", result.Appliance.URI)
	}
	if result.UID != nil {
		fmt.Fprintf(w, "%s\t%s\n", "UID", *result.UID)
	}
	if len(result.Facts) > 0 {
		keys := make([]string, 0, len(result.Facts))

		for k := range result.Facts {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		fmt.Fprintln(w, "\nFact\tValue")
		for _, k := range keys {
			fmt.Fprintf(w, "%s\t%s\n", k, result.Facts[k])
		}
	}
	w.Flush()

	return nil
}

func systemList(ctx context.Context, cmdArgs *systemListCmd) error {
	client := ctl.NewSystemServiceClient(args.URL, http.DefaultClient)
	result, err := client.List(ctx, cmdArgs.Limit, cmdArgs.Offset)
	if err != nil {
		return fmt.Errorf("cannot register system: %w", err)
	}

	if len(cmdArgs.DisplayFacts) == 0 {
		cmdArgs.DisplayFacts = []string{
			"redfish_manufacturer",
			"redfish_model",
			"system-manufacturer",
			"system-product-name",
		}
	}

	w := newTabWriter()
	fmt.Fprintln(w, "ID\tName\tHw Addresses\tAcquired\tFacts")
	for _, line := range result {
		a := line.HwAddrs[0]
		if len(line.HwAddrs) > 1 {
			a = fmt.Sprintf("%s (%d)", a, len(line.HwAddrs))
		}
		var factCol []string
		for _, fn := range cmdArgs.DisplayFacts {
			if f, ok := line.Facts[fn]; ok {
				factCol = append(factCol, f)
			}
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%t\t%s\n", line.ID, line.Name, a, line.Acquired, strings.Join(factCol, " "))
	}
	w.Flush()

	return nil
}

func systemKickstart(ctx context.Context, cmdArgs *systemKickstartCmd) error {
	client := ctl.NewSystemServiceClient(args.URL, http.DefaultClient)
	body, err := client.Kickstart(ctx, cmdArgs.Pattern)
	if err != nil {
		return fmt.Errorf("cannot render kickstart: %w", err)
	}

	fmt.Print(body)
	return nil
}

func systemLogs(ctx context.Context, cmdArgs *systemLogsCmd) error {
	if cmdArgs.Download != "" {
		// download log file
		url := fmt.Sprintf("%s/logs/%s", args.URL, cmdArgs.Download)
		err := download(url, os.Stdout)
		if err != nil {
			return fmt.Errorf("cannot fetch %s: %w", url, err)
		}
	} else {
		// get a listing
		client := ctl.NewSystemServiceClient(args.URL, http.DefaultClient)
		entries, err := client.Logs(ctx, cmdArgs.Pattern)
		if err != nil {
			return fmt.Errorf("cannot fetch logs: %w", err)
		}

		w := newTabWriter()
		fmt.Fprintln(w, "Created\tModified\tName\tSize")
		for _, le := range entries {
			fmt.Fprintf(w, "%s\t%s\t%s\t%d\n", le.CreatedAt.Local().Format(time.DateTime), le.ModifiedAt.Local().Format(time.DateTime), le.Path, le.Size)
		}
		w.Flush()
	}
	return nil
}

func systemAcquire(ctx context.Context, cmdArgs *systemAcquireCmd) error {
	client := ctl.NewSystemServiceClient(args.URL, http.DefaultClient)
	err := client.Acquire(ctx, cmdArgs.Pattern, cmdArgs.Image, cmdArgs.Comment, cmdArgs.Snippets)
	if err != nil {
		return fmt.Errorf("cannot acquire system: %w", err)
	}

	return nil
}

func systemRelease(ctx context.Context, cmdArgs *systemReleaseCmd) error {
	client := ctl.NewSystemServiceClient(args.URL, http.DefaultClient)
	err := client.Release(ctx, cmdArgs.Pattern)
	if err != nil {
		return fmt.Errorf("cannot release system: %w", err)
	}

	return nil
}

func systemBootNetwork(ctx context.Context, cmdArgs *systemBootNetworkCmd) error {
	client := ctl.NewSystemServiceClient(args.URL, http.DefaultClient)
	err := client.BootNetwork(ctx, cmdArgs.Pattern)
	if err != nil {
		return fmt.Errorf("cannot reset system: %w", err)
	}

	return nil
}

func systemBootLocal(ctx context.Context, cmdArgs *systemBootLocalCmd) error {
	client := ctl.NewSystemServiceClient(args.URL, http.DefaultClient)
	err := client.BootLocal(ctx, cmdArgs.Pattern)
	if err != nil {
		return fmt.Errorf("cannot reset system: %w", err)
	}

	return nil
}
