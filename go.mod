module forester

go 1.20

require (
	github.com/alexflint/go-arg v1.4.3
	github.com/digitalocean/go-libvirt v0.0.0-20221205150000-2939327a8519
	github.com/djherbis/atime v1.1.0
	github.com/djherbis/times v1.6.0
	github.com/georgysavva/scany/v2 v2.0.0
	github.com/go-chi/chi/v5 v5.0.10
	github.com/go-chi/render v1.0.3
	github.com/google/go-cmp v0.6.0
	github.com/google/uuid v1.4.0
	github.com/hooklift/iso9660 v1.0.0
	github.com/ilyakaznacheev/cleanenv v1.5.0
	github.com/jackc/pgx/v5 v5.5.0
	github.com/jackc/tern/v2 v2.1.1
	github.com/stmcginnis/gofish v0.15.0
	github.com/stretchr/testify v1.8.4
	github.com/thanhpk/randstr v1.0.6
	golang.org/x/exp v0.0.0-20231127185646-65229373498e
	gopkg.in/mcuadros/go-syslog.v2 v2.3.0
	libvirt.org/go/libvirtxml v1.9008.0
)

require golang.org/x/net v0.19.0 // indirect

require (
	github.com/BurntSushi/toml v1.3.2 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.2.1 // indirect
	github.com/Masterminds/sprig/v3 v3.2.3 // indirect
	github.com/ajg/form v1.5.1 // indirect
	github.com/alexflint/go-scalar v1.2.0 // indirect
	github.com/c4milo/gotoolkit v0.0.0-20190525173301-67483a18c17a // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/hooklift/assert v0.1.0 // indirect
	github.com/huandu/xstrings v1.4.0 // indirect
	github.com/imdario/mergo v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20231201235250-de7065d80cb9 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/pin/tftp/v3 v3.1.0
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/shopspring/decimal v1.3.1 // indirect
	github.com/spf13/cast v1.6.0 // indirect
	golang.org/x/crypto v0.16.0 // indirect
	golang.org/x/sync v0.5.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	olympos.io/encoding/edn v0.0.0-20201019073823-d3554ca0b0a3 // indirect
)

// https://github.com/darccio/mergo/issues/248
replace github.com/imdario/mergo => dario.cat/mergo v1.0.0
