package tmpl

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"text/template"

	"forester/internal/config"
	"forester/internal/version"
)

//go:embed *.tmpl.*
var templatesFS embed.FS

var templates *template.Template

func init() {
	var err error
	templates, err = template.New("").Funcs(template.FuncMap{
		"MakeSlice": MakeSlice,
	}).ParseFS(templatesFS, "*.tmpl.*")
	if err != nil {
		panic(err)
	}
}

func commonParams() *CommonParams {
	return &CommonParams{
		BaseHost:   config.BaseHost(),
		BaseURL:    config.BaseURL(),
		Version:    version.BuildTag,
		SyslogPort: strconv.Itoa(config.Application.SyslogPort),
	}
}

func Render(ctx context.Context, w io.Writer, name string, params any) error {
	var lb bytes.Buffer
	mw := io.MultiWriter(w, &lb)
	slog.DebugContext(ctx, "rendering template", "name", name, "params", params)
	err := templates.ExecuteTemplate(mw, name, params)
	if err != nil {
		return fmt.Errorf("error executing template: %w", err)
	}
	slog.DebugContext(ctx, lb.String(), "template", true)

	return nil
}

func RenderIpxeBootstrap(ctx context.Context, w io.Writer) error {
	params := commonParams()
	return Render(ctx, w, "bootstrap_ipxe.tmpl.txt", params)
}

func RenderGrubBootstrap(ctx context.Context, w io.Writer) error {
	params := commonParams()
	return Render(ctx, w, "bootstrap_grub.tmpl.txt", params)
}

func RenderGrubKernel(ctx context.Context, w io.Writer, params BootKernelParams) error {
	params.CommonParams = commonParams()

	return Render(ctx, w, "grub_kernel.tmpl.txt", params)
}

func RenderIpxeKernel(ctx context.Context, w io.Writer, params BootKernelParams) error {
	params.CommonParams = commonParams()

	return Render(ctx, w, "ipxe_kernel.tmpl.txt", params)
}

func RenderBootError(ctx context.Context, w io.Writer, params BootErrorParams) error {
	params.CommonParams = commonParams()

	return Render(ctx, w, "grub_error.tmpl.txt", params)
}

func RenderKickstartDiscover(ctx context.Context, w io.Writer, params KickstartParams) error {
	params.CommonParams = commonParams()

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

func RenderDhcpConf(ctx context.Context, w io.Writer, name, format string, params DhcpParams) error {
	params.CommonParams = commonParams()

	return Render(ctx, w, fmt.Sprintf("%s_%s.tmpl.txt", name, format), params)
}
