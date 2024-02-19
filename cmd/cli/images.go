package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"forester/internal/api/ctl"
)

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
	fmt.Fprintln(w, "Image ID\tImage Name\tKind")
	for _, img := range images {
		fmt.Fprintf(w, "%d\t%s\t%s\n", img.ID, img.Name, ctl.ImageIntToKind(img.Kind))
	}
	w.Flush()

	return nil
}
