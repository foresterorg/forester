package main

import (
	"context"
	"errors"
	"fmt"
	"forester/internal/api/ctl"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

type systemRegisterCmd struct {
	Name          string            `arg:"-n"`
	HwAddrs       []string          `arg:"-m,required,separate"`
	Facts         map[string]string `arg:"-f"`
	ApplianceName string            `arg:"-a"`
	UID           string            `arg:"-u"`
}

type systemShowCmd struct {
	Pattern string `arg:"positional,required" placeholder:"MAC_OR_NAME"`
}

type systemListCmd struct {
	DisplayFacts []string `args:"-f,separate"`
	Limit        int64    `arg:"-m" default:"100"`
	Offset       int64    `arg:"-o" default:"0"`
}

type systemKickstartCmd struct {
	Pattern string `arg:"positional,required" placeholder:"MAC_OR_NAME"`
}

type systemLogsCmd struct {
	Pattern  string `arg:"positional,required" placeholder:"MAC_OR_NAME"`
	Download string `arg:"-d" help:"download a log" placeholder:"f-X-XXXX.log"`
	Last     bool   `arg:"-l" help:"show last log"`
}

type systemSshCmd struct {
	Pattern string `arg:"positional,required" placeholder:"MAC_OR_NAME"`
}

type systemRenameCmd struct {
	Pattern string `arg:"positional,required" placeholder:"MAC_OR_NAME"`
	Name    string `arg:"-n,required" placeholder:"NEW_SYSTEM_NAME"`
}

type systemDeployCmd struct {
	Pattern     string   `arg:"positional,required" placeholder:"MAC_OR_NAME"`
	Image       string   `arg:"-i,required"`
	Snippets    []string `arg:"-s,separate"`
	TextSnippet string   `arg:"-x"`
	Kickstart   string   `arg:"-k" placeholder:"KS_OVERRIDE_CONTENTS"`
	Comment     string   `arg:"-c"`
	Duration    string   `arg:"-d" default:"3h"`
}

type systemBootNetworkCmd struct {
	Pattern string `arg:"positional,required" placeholder:"MAC_OR_NAME"`
}

type systemBootLocalCmd struct {
	Pattern string `arg:"positional,required" placeholder:"MAC_OR_NAME"`
}

type emptyCmd struct{}

type systemCmd struct {
	Register    *systemRegisterCmd    `arg:"subcommand:register" help:"register system"`
	List        *systemListCmd        `arg:"subcommand:list" help:"list systems"`
	Show        *systemShowCmd        `arg:"subcommand:show" help:"show system"`
	Rename      *systemRenameCmd      `arg:"subcommand:rename" help:"rename existing system"`
	Deploy      *systemDeployCmd      `arg:"subcommand:deploy" help:"deploy an image to a system"`
	Acquire     *emptyCmd             `arg:"subcommand:acquire" help:"acquire system (deprecated)"`
	Release     *emptyCmd             `arg:"subcommand:release" help:"release system (deprecated)"`
	Kickstart   *systemKickstartCmd   `arg:"subcommand:kickstart" help:"show system kickstart"`
	Logs        *systemLogsCmd        `arg:"subcommand:logs" help:"show installation log history"`
	Ssh         *systemSshCmd         `arg:"subcommand:ssh" help:"ssh to anaconda during installation"`
	BootNetwork *systemBootNetworkCmd `arg:"subcommand:bootnet" help:"reset (hard reboot) system and boot from network"`
	BootLocal   *systemBootLocalCmd   `arg:"subcommand:bootlocal" help:"reset (hard reboot) system and boot from local drive"`
}

func systemRegister(ctx context.Context, cmdArgs *systemRegisterCmd) error {
	client := ctl.NewSystemServiceClient(args.URL, http.DefaultClient)
	sys := ctl.NewSystem{
		Name:          cmdArgs.Name,
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
	fmt.Fprintln(w, "ID\tName\tHw Addresses\tFacts")
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
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", line.ID, line.Name, a, strings.Join(factCol, " "))
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

func downloadLog(path, uuid string) error {
	url := fmt.Sprintf("%s/logs/%s", path, uuid)
	err := download(url, os.Stdout)
	if err != nil {
		return fmt.Errorf("cannot fetch %s: %w", url, err)
	}
	return nil
}

func systemLogs(ctx context.Context, cmdArgs *systemLogsCmd) error {
	if cmdArgs.Download != "" {
		err := downloadLog(args.URL, cmdArgs.Download)
		if err != nil {
			return fmt.Errorf("cannot download: %w", err)
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
		var lastEntry *ctl.LogEntry
		for _, le := range entries {
			if le.Size > 0 {
				lastEntry = le
				fmt.Fprintf(w, "%s\t%s\t%s\t%d\n", le.CreatedAt.Local().Format(time.DateTime), le.ModifiedAt.Local().Format(time.DateTime), le.Path, le.Size)
			}
		}
		w.Flush()

		if cmdArgs.Last && lastEntry != nil {
			err := downloadLog(args.URL, lastEntry.Path)
			if err != nil {
				return fmt.Errorf("cannot download: %w", err)
			}
		}
	}
	return nil
}

func systemSsh(ctx context.Context, cmdArgs *systemSshCmd) error {
	client := ctl.NewSystemServiceClient(args.URL, http.DefaultClient)
	sys, err := client.Find(ctx, cmdArgs.Pattern)
	if err != nil {
		return fmt.Errorf("cannot find system: %w", err)
	}
	fmt.Printf("ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no forester-%d@IP\nPASSWORD: %s\n", sys.ID, "INSTALLATION UUID")

	return nil
}

func systemRename(ctx context.Context, cmdArgs *systemRenameCmd) error {
	client := ctl.NewSystemServiceClient(args.URL, http.DefaultClient)
	err := client.Rename(ctx, cmdArgs.Pattern, cmdArgs.Name)
	if err != nil {
		return fmt.Errorf("cannot rename system: %w", err)
	}

	return nil
}

var ErrAcquireReleaseDeprecated = errors.New("acquire/release was deprecated, use 'forester-cli deploy' instead")

func systemAcquire(ctx context.Context, cmdArgs *emptyCmd) error {
	return ErrAcquireReleaseDeprecated
}

func systemRelease(ctx context.Context, cmdArgs *emptyCmd) error {
	return ErrAcquireReleaseDeprecated
}

func systemDeploy(ctx context.Context, cmdArgs *systemDeployCmd) error {
	dur, err := time.ParseDuration(cmdArgs.Duration)
	if err != nil {
		return fmt.Errorf("cannot parse duration: %w", err)
	}

	client := ctl.NewSystemServiceClient(args.URL, http.DefaultClient)
	err = client.Deploy(ctx, cmdArgs.Pattern, cmdArgs.Image, cmdArgs.Snippets, cmdArgs.TextSnippet, cmdArgs.Kickstart, cmdArgs.Comment, time.Now().Add(dur))
	if err != nil {
		return fmt.Errorf("cannot deploy system: %w", err)
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
