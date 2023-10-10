package config

import (
	"fmt"
	"forester/internal/ptr"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"golang.org/x/exp/slog"
)

var config struct {
	App struct {
		Port            int           `env:"PORT" env-default:"8000" env-description:"HTTP port of the API service"`
		SyslogPort      int           `env:"SYSLOG_PORT" env-default:"8514" env-description:"syslog TCP and UDP port"`
		Hostname        string        `env:"HOSTNAME" env-default:"" env-description:"hostname of the service"`
		InstallDuration time.Duration `env:"INSTALL_DURATION" env-default:"1h" env-description:"duration for which the service initiates provisioning after acquire"`
	} `env-prefix:"APP_"`
	Database struct {
		Host        string        `env:"HOST" env-default:"localhost" env-description:"main database hostname"`
		Port        uint16        `env:"PORT" env-default:"5432" env-description:"main database port"`
		Name        string        `env:"NAME" env-default:"forester" env-description:"main database name"`
		User        string        `env:"USER" env-default:"postgres" env-description:"main database username"`
		Password    string        `env:"PASSWORD" env-default:"" env-description:"main database password"`
		MinConn     int32         `env:"MIN_CONN" env-default:"2" env-description:"connection pool minimum size"`
		MaxConn     int32         `env:"MAX_CONN" env-default:"50" env-description:"connection pool maximum size"`
		MaxIdleTime time.Duration `env:"MAX_IDLE_TIME" env-default:"15m" env-description:"connection pool idle time (time interval syntax)"`
		MaxLifetime time.Duration `env:"MAX_LIFETIME" env-default:"2h" env-description:"connection pool total lifetime (time interval syntax)"`
		LogLevel    string        `env:"LOG_LEVEL" env-default:"trace" env-description:"logging level of database logs"`
	} `env-prefix:"DATABASE_"`
	Logging struct {
		Level     string `env:"LEVEL" env-default:"debug" env-description:"logger level (debug, info, warn, error)"`
		Syslog    bool   `env:"SYSLOG" env-default:"false" env-description:"write Anaconda syslog data into application log"`
		SyslogDir string `env:"SYSLOG_DIR" env-default:"logs" env-description:"absolute path to directory with syslog files"`
	} `env-prefix:"LOGGING_"`
	Images struct {
		Directory string `env:"DIR" env-default:"images" env-description:"absolute path to directory with images"`
		BootId    int64  `env:"BOOT_ID" env-default:"1" env-description:"boot shim/grub from image DB ID"`
	} `env-prefix:"IMAGES_"`
}

// Config shortcuts
var (
	Application = &config.App
	Database    = &config.Database
	Logging     = &config.Logging
	Images      = &config.Images
)

var Hostname string

func init() {
	var err error
	Hostname, err = os.Hostname()
	if err != nil {
		panic(err)
	}
}

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

	// print key configuration values
	slog.Debug("image configuration",
		"path", config.Images.Directory,
		"boot_id", config.Images.BootId,
	)
	slog.Debug("syslog configuration",
		"enabled", config.Logging.Syslog,
		"directory", config.Logging.SyslogDir,
	)

	return nil
}

func HelpText() string {
	text, err := cleanenv.GetDescription(&config, ptr.To(""))
	if err != nil {
		panic(err)
	}
	return text
}

func ParsedLoggingLevel() slog.Level {
	switch Logging.Level {
	case "debug", "DEBUG":
		return slog.LevelDebug
	case "info", "INFO":
		return slog.LevelInfo
	case "warn", "warning", "WARN", "WARNING":
		return slog.LevelWarn
	case "error", "ERROR":
		return slog.LevelError
	}

	return slog.LevelDebug
}

func BootPath() string {
	return path.Join(config.Images.Directory, strconv.FormatInt(config.Images.BootId, 10))
}

func BaseURL() string {
	return fmt.Sprintf("http://%s:%d", Hostname, config.App.Port)
}

func BaseHost() string {
	return Hostname
}
