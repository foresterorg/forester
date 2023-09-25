package main

import (
	"context"
	"errors"
	"fmt"
	"forester/internal/api/ctl"
	"forester/internal/config"
	"forester/internal/logging"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	arg "github.com/alexflint/go-arg"
	"golang.org/x/exp/slog"
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

type systemAcquireCmd struct {
	Pattern   string `arg:"positional,required" placeholder:"MAC_OR_NAME"`
	ImageName string `arg:"-i,required"`
	Comment   string `arg:"-c"`
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
	BootNetwork *systemBootNetworkCmd `arg:"subcommand:bootnet" help:"reset (hard reboot) system and boot from network"`
	BootLocal   *systemBootLocalCmd   `arg:"subcommand:bootlocal" help:"reset (hard reboot) system and boot from local drive"`
}

type imageUploadCmd struct {
	ImageFile string `arg:"positional,required" placeholder:"IMAGE_FILE"`
	Name      string `arg:"-n,required"`
}

type imageShowCmd struct {
	ImageName string `arg:"positional,required" placeholder:"NAME"`
}

type imageListCmd struct {
	Limit  int64 `arg:"-m" default:"100"`
	Offset int64 `arg:"-o" default:"0"`
}

type imageCmd struct {
	Upload *imageUploadCmd `arg:"subcommand:upload" help:"upload image"`
	Show   *imageShowCmd   `arg:"subcommand:show" help:"show image"`
	List   *imageListCmd   `arg:"subcommand:list" help:"list images"`
}

type applianceCreateCmd struct {
	Name string `arg:"-n,required"`
	Kind string `arg:"-k" default:"libvirt"`
	URI  string `arg:"-u" default:"unix:///var/run/libvirt/libvirt-sock"`
}

type applianceListCmd struct {
	Limit  int64 `arg:"-m" default:"100"`
	Offset int64 `arg:"-o" default:"0"`
}

type applianceEnlistCmd struct {
	Name          string `arg:"positional,required" placeholder:"APPLIANCE_NAME"`
	SystemPattern string `arg:"-n" placeholder:"REGEXP_SYSTEM_PATTERN" default:".*"`
}

type applianceCmd struct {
	Create *applianceCreateCmd `arg:"subcommand:create" help:"create appliance"`
	List   *applianceListCmd   `arg:"subcommand:list" help:"list appliances"`
	Enlist *applianceEnlistCmd `arg:"subcommand:enlist" help:"enlist systems of appliance"`
}

var args struct {
	Image     *imageCmd     `arg:"subcommand:image" help:"image related commands"`
	System    *systemCmd    `arg:"subcommand:system" help:"system related commands"`
	Appliance *applianceCmd `arg:"subcommand:appliance" help:"appliance related commands"`
	URL       string        `default:"http://localhost:8000"`
	Config    string        `default:"config/forester.env"`
	Quiet     bool
	Verbose   bool
	Debug     bool
}

func kindToInt(kind string) int16 {
	switch strings.ToLower(kind) {
	case "libvirt":
		return 1
	case "redfish":
		return 2
	default:
		panic(fmt.Sprintf("unknown kind: %s", kind))
	}
}

func intToKind(kind int16) string {
	switch kind {
	case 1:
		return "libvirt"
	case 2:
		return "redfish"
	default:
		panic(fmt.Sprintf("unknown kind: %d", kind))
	}
}

func main() {
	parser := arg.MustParse(&args)
	if parser.Subcommand() == nil {
		parser.Fail("missing subcommand")
	}

	if args.Debug {
		logging.Initialize(slog.LevelDebug)
	} else if args.Verbose {
		logging.Initialize(slog.LevelInfo)
	} else if args.Quiet {
		logging.Initialize(slog.LevelError)
	} else {
		logging.Initialize(slog.LevelWarn)
	}

	ctx := context.Background()
	err := config.Initialize(args.Config)
	if err != nil {
		panic(err)
	}

	switch {
	case args.Image != nil:
		if cmd := args.Image.Upload; cmd != nil {
			err = imageUpload(ctx, cmd)
		} else if cmd := args.Image.Show; cmd != nil {
			err = imageShow(ctx, cmd)
		} else if cmd := args.Image.List; cmd != nil {
			err = imageList(ctx, cmd)
		} else {
			_ = parser.FailSubcommand("unknown subcommand", "image")
		}
	case args.System != nil:
		if cmd := args.System.Register; cmd != nil {
			err = systemRegister(ctx, cmd)
		} else if cmd := args.System.Show; cmd != nil {
			err = systemShow(ctx, cmd)
		} else if cmd := args.System.List; cmd != nil {
			err = systemList(ctx, cmd)
		} else if cmd := args.System.Acquire; cmd != nil {
			err = systemAcquire(ctx, cmd)
		} else if cmd := args.System.Release; cmd != nil {
			err = systemRelease(ctx, cmd)
		} else if cmd := args.System.BootNetwork; cmd != nil {
			err = systemBootNetwork(ctx, cmd)
		} else if cmd := args.System.BootLocal; cmd != nil {
			err = systemBootLocal(ctx, cmd)
		} else {
			_ = parser.FailSubcommand("unknown subcommand", "system")
		}
	case args.Appliance != nil:
		if cmd := args.Appliance.Create; cmd != nil {
			err = applianceCreate(ctx, cmd)
		} else if cmd := args.Appliance.List; cmd != nil {
			err = applianceList(ctx, cmd)
		} else if cmd := args.Appliance.Enlist; cmd != nil {
			err = applianceEnlist(ctx, cmd)
		} else {
			_ = parser.FailSubcommand("unknown subcommand", "appliance")
		}
	default:
		parser.Fail("missing subcommand")
	}

	if err != nil {
		if args.Debug {
			panic(err)
		} else {
			fmt.Fprintf(os.Stderr, "ERROR: %s\nCommand returned an error, use -d or --debug for more info\n\n", err.Error())
		}
	}
}

func newTabWriter() *tabwriter.Writer {
	return tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
}

var ErrUploadNot200 = errors.New("upload error")

func uploadURL(mainURL, newPath string) (string, error) {
	newURL, err := url.Parse(mainURL)
	if err != nil {
		return "", fmt.Errorf("cannot parse URL %s: %w", mainURL, err)
	}
	newURL.Path = newPath

	return newURL.String(), nil
}

func imageUpload(ctx context.Context, cmdArgs *imageUploadCmd) error {
	client := ctl.NewImageServiceClient(args.URL, http.DefaultClient)
	_, uploadPath, err := client.Create(ctx, &ctl.Image{
		Name: cmdArgs.Name,
	})
	if err != nil {
		return fmt.Errorf("cannot create image: %w", err)
	}

	file, err := os.Open(cmdArgs.ImageFile)
	if err != nil {
		return fmt.Errorf("cannot open image: %w", err)
	}
	defer file.Close()

	uploadURL, err := uploadURL(args.URL, uploadPath)
	if err != nil {
		return fmt.Errorf("cannot upload image: %w", err)
	}

	r, err := http.NewRequest("PUT", uploadURL, file)
	if err != nil {
		return fmt.Errorf("cannot create upload request: %w", err)
	}
	fi, err := file.Stat()
	if err != nil {
		return fmt.Errorf("cannot stat file: %w", err)
	}
	r.Header.Set("Content-Type", "application/octet-stream")
	r.Header.Set("Content-Size", strconv.FormatInt(fi.Size(), 10))
	uploadClient := &http.Client{}
	res, err := uploadClient.Do(r)
	if err != nil {
		return fmt.Errorf("cannot send data: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("server returned %d: %w", res.StatusCode, ErrUploadNot200)
	}

	return nil
}

func imageShow(ctx context.Context, cmdArgs *imageShowCmd) error {
	client := ctl.NewImageServiceClient(args.URL, http.DefaultClient)
	result, err := client.Find(ctx, cmdArgs.ImageName)
	if err != nil {
		return fmt.Errorf("cannot find: %w", err)
	}

	w := newTabWriter()
	fmt.Fprintln(w, "Attribute\tValue")
	fmt.Fprintf(w, "%s\t%d\n", "ID", result.ID)
	fmt.Fprintf(w, "%s\t%s\n", "Name", result.Name)
	w.Flush()

	return nil
}

func imageList(ctx context.Context, cmdArgs *imageListCmd) error {
	client := ctl.NewImageServiceClient(args.URL, http.DefaultClient)
	images, err := client.List(ctx, cmdArgs.Limit, cmdArgs.Offset)
	if err != nil {
		return fmt.Errorf("cannot list images: %w", err)
	}

	w := newTabWriter()
	fmt.Fprintln(w, "Image ID\tImage Name")
	for _, img := range images {
		fmt.Fprintf(w, "%d\t%s\n", img.ID, img.Name)
	}
	w.Flush()

	return nil
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
	if result.Acquired && result.ImageID != nil {
		fmt.Fprintf(w, "%s\t%s\n", "Acquired at", result.AcquiredAt.Format(time.ANSIC))
		fmt.Fprintf(w, "%s\t%d\n", "Image ID", *result.ImageID)
	}
	for _, mac := range result.HwAddrs {
		fmt.Fprintf(w, "%s\t%s\n", "MAC", mac)
	}
	if result.Appliance != nil && result.Appliance.Name != "" {
		fmt.Fprintf(w, "%s\t%s\n", "Appliance Name", result.Appliance.Name)
		fmt.Fprintf(w, "%s\t%s\n", "Appliance Kind", intToKind(result.Appliance.Kind))
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

func systemAcquire(ctx context.Context, cmdArgs *systemAcquireCmd) error {
	client := ctl.NewSystemServiceClient(args.URL, http.DefaultClient)
	err := client.Acquire(ctx, cmdArgs.Pattern, cmdArgs.ImageName, cmdArgs.Comment)
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

func applianceCreate(ctx context.Context, cmdArgs *applianceCreateCmd) error {
	client := ctl.NewApplianceServiceClient(args.URL, http.DefaultClient)
	err := client.Create(ctx, cmdArgs.Name, kindToInt(cmdArgs.Kind), cmdArgs.URI)
	if err != nil {
		return fmt.Errorf("cannot create appliance: %w", err)
	}

	return nil
}

func applianceList(ctx context.Context, cmdArgs *applianceListCmd) error {
	client := ctl.NewApplianceServiceClient(args.URL, http.DefaultClient)
	appliances, err := client.List(ctx, cmdArgs.Limit, cmdArgs.Offset)
	if err != nil {
		return fmt.Errorf("cannot list appliances: %w", err)
	}

	w := newTabWriter()
	fmt.Fprintln(w, "ID\tName\tKind\tURI")
	for _, a := range appliances {
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", a.ID, a.Name, intToKind(a.Kind), a.URI)
	}
	w.Flush()

	return nil
}

func applianceEnlist(ctx context.Context, cmdArgs *applianceEnlistCmd) error {
	client := ctl.NewApplianceServiceClient(args.URL, http.DefaultClient)
	err := client.Enlist(ctx, cmdArgs.Name, cmdArgs.SystemPattern)
	if err != nil {
		return fmt.Errorf("cannot enlist systems: %w", err)
	}

	return nil
}
