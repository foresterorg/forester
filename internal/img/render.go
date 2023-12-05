package img

import (
	"context"
	"embed"
	"fmt"
	"forester/internal/config"
	"forester/internal/version"
	"io"
	"strconv"
	"text/template"
)

//go:embed *.tmpl.*
var templatesFS embed.FS

var templates *template.Template

func init() {
	var err error
	templates, err = template.ParseFS(templatesFS, "*.tmpl.*")
	if err != nil {
		panic(err)
	}
}

type BootISOParams struct {
	BaseHost   string
	BaseURL    string
	SyslogPort string
	Version    string
	ImageID    int64
	ImageDir   string
}

// Generates BIOS/EFI common boot ISO: https://fedoraproject.org/wiki/Changes/BIOSBootISOWithGrub2
func renderGenerateBootISO(ctx context.Context, w io.Writer, imgID int64, imgDir string) error {
	p := BootISOParams{
		BaseHost:   config.BaseHost(),
		BaseURL:    config.BaseURL(),
		Version:    version.BuildTag,
		SyslogPort: strconv.Itoa(config.Application.SyslogPort),
		ImageID:    imgID,
		ImageDir:   imgDir,
	}
	err := templates.ExecuteTemplate(w, "genboot.tmpl.sh", p)
	if err != nil {
		return fmt.Errorf("error executing template: %w", err)
	}

	return nil
}
