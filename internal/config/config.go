package config

import (
	"fmt"
	"forester/internal/ptr"
	"os"
	"path/filepath"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

var config struct {
	App struct {
		Port int `env:"PORT" env-default:"8000" env-description:"HTTP port of the API service"`
	} `env-prefix:"APP_"`
	Database struct {
		Host        string        `env:"HOST" env-default:"localhost" env-description:"main database hostname"`
		Port        uint16        `env:"PORT" env-default:"5432" env-description:"main database port"`
		Name        string        `env:"NAME" env-default:"provisioning" env-description:"main database name"`
		User        string        `env:"USER" env-default:"postgres" env-description:"main database username"`
		Password    string        `env:"PASSWORD" env-default:"" env-description:"main database password"`
		MinConn     int32         `env:"MIN_CONN" env-default:"2" env-description:"connection pool minimum size"`
		MaxConn     int32         `env:"MAX_CONN" env-default:"50" env-description:"connection pool maximum size"`
		MaxIdleTime time.Duration `env:"MAX_IDLE_TIME" env-default:"15m" env-description:"connection pool idle time (time interval syntax)"`
		MaxLifetime time.Duration `env:"MAX_LIFETIME" env-default:"2h" env-description:"connection pool total lifetime (time interval syntax)"`
		LogLevel    string        `env:"LOG_LEVEL" env-default:"info" env-description:"logging level of database logs"`
	} `env-prefix:"DATABASE_"`
	Logging struct {
		Level    string `env:"LEVEL" env-default:"info" env-description:"logger level (trace, debug, info, warn, error, fatal, panic)"`
		Stdout   bool   `env:"STDOUT" env-default:"true" env-description:"logger standard output, disabled in clowder by default, stdout is still used if there is no other writer"`
		MaxField int    `env:"MAX_FIELD" env-default:"0" env-description:"logger maximum field length (dev only)"`
	} `env-prefix:"LOGGING_"`
	Images struct {
		Directory string `env:"DIRECTORY" env-default:"images" env-description:"absolute path to directory with images"`
	} `env-prefix:"IMAGES_"`
}

// Config shortcuts
var (
	Application = &config.App
	Database    = &config.Database
	Logging     = &config.Logging
	Images      = &config.Images
)

// Initialize loads configuration from provided .env files, the first existing file wins.
func Initialize(configFiles ...string) error {
	var loaded bool
	for _, configFile := range configFiles {
		if _, err := os.Stat(configFile); err == nil {
			// if config file exists, load it (also loads environmental variables)
			err := cleanenv.ReadConfig(configFile, &config)
			if err != nil {
				return fmt.Errorf("config read error: %w", err)
			}
			loaded = true
		}
	}

	if !loaded {
		// otherwise use only environmental variables instead
		err := cleanenv.ReadEnv(&config)
		if err != nil {
			return fmt.Errorf("config environment parse error: %w", err)
		}
	}

	// validate
	var err error
	config.Images.Directory, err = filepath.Abs(config.Images.Directory)
	if err != nil {
		return fmt.Errorf("image directory config error: %w", err)
	}

	return nil
}

func HelpText() (string, error) {
	text, err := cleanenv.GetDescription(&config, ptr.To(""))
	if err != nil {
		return "", fmt.Errorf("cannot generate help text: %w", err)
	}
	return text, nil
}
