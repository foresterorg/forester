package main

import (
	"context"
	"errors"
	"fmt"
	"forester/internal/api/ctl"
	"forester/internal/config"
	"forester/internal/logging"
	"net/http"
	"os"
	"strconv"
	"text/tabwriter"

	arg "github.com/alexflint/go-arg"
	"golang.org/x/exp/slog"
)

type imageUploadCmd struct {
	ImageFile string `arg:"positional" placeholder:"IMAGE_FILE"`
	Name      string `arg:"-n"`
}

type imageListCmd struct {
	Limit  int64 `arg:"-m" default:"100"`
	Offset int64 `arg:"-o" default:"0"`
}

type imageCmd struct {
	Upload *imageUploadCmd `arg:"subcommand:upload" help:"upload image"`
	List   *imageListCmd   `arg:"subcommand:list" help:"list images"`
}

var args struct {
	Image   *imageCmd `arg:"subcommand:image" help:"image related commands"`
	URL     string    `arg:"-u" default:"http://localhost:8000"`
	Config  string    `arg:"-c" default:"config/forester.env"`
	Quiet   bool      `arg:"-q"`
	Verbose bool      `arg:"-v"`
	Debug   bool      `arg:"-d"`
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
		} else if cmd := args.Image.List; cmd != nil {
			err = imageList(ctx, cmd)
		} else {
			_ = parser.FailSubcommand("missing image subcommand", "image")
		}
	default:
		parser.Fail("missing subcommand")
	}

	if err != nil {
		panic(err)
	}
}

var ErrUploadNot200 = errors.New("upload error")

func imageUpload(ctx context.Context, cmdArgs *imageUploadCmd) error {
	client := ctl.NewImageServiceClient(args.URL, http.DefaultClient)
	_, uploadURL, err := client.Create(ctx, &ctl.Image{
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

func imageList(ctx context.Context, cmdArgs *imageListCmd) error {

	client := ctl.NewImageServiceClient(args.URL, http.DefaultClient)
	images, err := client.List(ctx, 1000, 0)
	if err != nil {
		return fmt.Errorf("cannot list images: %w", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 5, ' ', 0)
	fmt.Fprintln(w, "Image ID\tImage Name")
	for _, img := range images {
		fmt.Fprintf(w, "%d\t%s\n", img.ID, img.Name)
	}
	w.Flush()

	return nil
}
