package tmpl

import (
	"context"
	"embed"
	"fmt"
	"forester/internal/config"
	"forester/internal/version"
	"golang.org/x/exp/slog"
	"io"
	"strconv"
	"text/template"
)
import _ "embed"

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

func commonParams() *CommonParams {
	return &CommonParams{
		BaseHost:   config.Hostname,
		BaseURL:    config.BaseURL(),
		Version:    version.BuildTag,
		SyslogPort: strconv.Itoa(config.Application.SyslogPort),
	}
}

func Render(ctx context.Context, w io.Writer, name string, params any) error {
	slog.DebugContext(ctx, "rendering emplate", "name", name, "params", params)
	err := templates.ExecuteTemplate(w, name, params)
	if err != nil {
		return fmt.Errorf("error executing template: %w", err)
	}

	return nil
}

func RenderGrubBootstrap(ctx context.Context, w io.Writer) error {
	params := commonParams()

	return Render(ctx, w, "grub_bootstrap.tmpl.txt", params)
}

func RenderGrubKernel(ctx context.Context, w io.Writer, params GrubKernelParams) error {
	params.CommonParams = commonParams()

	return Render(ctx, w, "grub_kernel.tmpl.txt", params)
}

func RenderGrubError(ctx context.Context, w io.Writer, params GrubErrorParams) error {
	params.CommonParams = commonParams()

	return Render(ctx, w, "grub_error.tmpl.txt", params)
}

func RenderKickstartDiscover(ctx context.Context, w io.Writer) error {
	params := commonParams()

	return Render(ctx, w, "ks_discover.tmpl.txt", params)
}

func RenderKickstartInstall(ctx context.Context, w io.Writer, params KickstartParams) error {
	params.CommonParams = commonParams()

	return Render(ctx, w, "ks_install.tmpl.txt", params)
}

func RenderKickstartError(ctx context.Context, w io.Writer, params KickstartErrorParams) error {
	params.CommonParams = commonParams()

	return Render(ctx, w, "ks_error.tmpl.txt", params)
}
