package tmpl

import (
	"embed"
	"fmt"
	"forester/internal/config"
	"forester/internal/version"
	"io"
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
		BaseURL: config.BaseURL(),
		Version: version.BuildTag,
	}
}

func Render(w io.Writer, name string, params any) error {
	err := templates.ExecuteTemplate(w, name, params)
	if err != nil {
		return fmt.Errorf("error executing template: %w", err)
	}

	return nil
}

func RenderGrubBootstrap(w io.Writer) error {
	params := commonParams()

	return Render(w, "grub_bootstrap.tmpl.txt", params)
}

func RenderGrubKernel(w io.Writer, params GrubKernelParams) error {
	params.CommonParams = commonParams()

	return Render(w, "grub_kernel.tmpl.txt", params)
}

func RenderGrubError(w io.Writer, params GrubErrorParams) error {
	params.CommonParams = commonParams()

	return Render(w, "grub_error.tmpl.txt", params)
}

func RenderKickstartDiscover(w io.Writer) error {
	params := commonParams()

	return Render(w, "ks_discover.tmpl.txt", params)
}

func RenderKickstartInstall(w io.Writer, params KickstartParams) error {
	params.CommonParams = commonParams()

	return Render(w, "ks_install.tmpl.txt", params)
}

func RenderKickstartError(w io.Writer, params KickstartErrorParams) error {
	params.CommonParams = commonParams()

	return Render(w, "ks_error.tmpl.txt", params)
}
