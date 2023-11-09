package main

import (
	"context"
	"fmt"
	"forester/internal/config"
	"forester/internal/logging"
	arg "github.com/alexflint/go-arg"
	"golang.org/x/exp/slog"
	"os"
	"text/tabwriter"
)

var args struct {
	Image     *imageCmd     `arg:"subcommand:image" help:"image related commands"`
	Snippet   *snippetCmd   `arg:"subcommand:snippet" help:"snippet related commands"`
	System    *systemCmd    `arg:"subcommand:system" help:"system related commands"`
	Appliance *applianceCmd `arg:"subcommand:appliance" help:"appliance related commands"`
	URL       string        `default:"http://localhost:8000"`
	Config    string        `default:"config/forester.env"`
	Quiet     bool
	Verbose   bool
	Debug     bool
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
	case args.Snippet != nil:
		if cmd := args.Snippet.Create; cmd != nil {
			err = snippetCreate(ctx, cmd)
		} else if cmd := args.Snippet.List; cmd != nil {
			err = snippetList(ctx, cmd)
		} else if cmd := args.Snippet.Edit; cmd != nil {
			err = snippetEdit(ctx, cmd)
		} else if cmd := args.Snippet.Delete; cmd != nil {
			err = snippetDelete(ctx, cmd)
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
