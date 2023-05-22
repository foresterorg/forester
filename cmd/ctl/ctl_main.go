package main

import (
	"context"
	"flag"
	"fmt"
	"forester/internal/config"
	"forester/internal/img"
	"forester/internal/logging"
	"log"
	"os"

	"golang.org/x/exp/slog"
)

var (
	_debug = flag.Bool("G", false, "debug level logging")
)

func printe(msg string) {
	_, _ = fmt.Fprint(os.Stderr, msg+"\n")
	os.Exit(1)
}

func printef(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

func main() {

	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		printe("Please specify a subcommand")
	}
	cmd, args := args[0], args[1:]

	if *_debug {
		logging.Initialize(slog.LevelDebug)
	} else {
		logging.Initialize(slog.LevelInfo)
	}

	switch cmd {
	case "image":
		image(args)
	default:
		printef("Unrecognized command %q. Use -h to get help.", cmd)
	}

	err := config.Initialize()
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	file, err := os.Open("f37-truncated.iso")
	if err != nil {
		panic(err)
	}
	err = img.UploadImage(ctx, 1, file)
	if err != nil {
		panic(err)
	}
}

func registerGlobalFlags(fset *flag.FlagSet) {
	flag.VisitAll(func(f *flag.Flag) {
		fset.Var(f.Value, f.Name, f.Usage)
	})
}

func image(args []string) {
	fset := flag.NewFlagSet("image", flag.ExitOnError)
	registerGlobalFlags(fset)

	var (
		_    = fset.Bool("u", false, "upload image")
		list = fset.Bool("l", false, "list images")
		del  = fset.Bool("d", false, "delete image")
		_    = fset.String("f", "", "image file name")
	)

	err := fset.Parse(args)
	if err != nil {
		log.Fatal("Unable to parse image flags")
	}
	args = fset.Args()
	slog.Info("Image called", "list", list, "del", del)
}
