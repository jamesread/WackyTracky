package runtimeconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jamesread/golure/pkg/listenaddr"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	log "github.com/sirupsen/logrus"
)

// RequiredMigration is the sql-migrate id the IAM SQLite database must have applied.
const RequiredMigration = "1.iam.sql"

type config struct {
	ListenAddress      string          `koanf:"listenAddress"`
	Database           *databaseConfig `koanf:"database"`
	IAM                *iamConfig      `koanf:"iam"`
	Auth               map[string]any  `koanf:"auth"`
	WallpaperDirectory string          `koanf:"wallpaperDirectory"`
	Cert               string          `koanf:"cert"`
	Key                string          `koanf:"key"`
	ShowFooter         bool            `koanf:"showFooter"`
	ShowNewVersions    bool            `koanf:"showNewVersions"`
	ShowVersionNumber  bool            `koanf:"showVersionNumber"`
}

type databaseConfig struct {
	Driver   string `koanf:"driver"`
	Database string `koanf:"database"`
	Hostname string `koanf:"hostname"`
	Username string `koanf:"username"`
	Password string `koanf:"password"`
	Port     int    `koanf:"port"`
}

type iamConfig struct {
	Database string `koanf:"database"`
}

var (
	RuntimeConfig = defaultConfig()
	configFile    string
)

func defaultConfig() config {
	return config{
		Database: &databaseConfig{
			Driver:   "nodriver",
			Hostname: "localhost",
			Username: "user",
			Password: "pass",
			Port:     7687,
		},
		IAM: &iamConfig{
			Database: "./data/iam.sqlite",
		},
		Auth:               map[string]any{},
		ListenAddress:      "0.0.0.0:8080",
		WallpaperDirectory: "/opt/wallpaper/",
		Cert:               "wt.crt",
		Key:                "wt.key",
		ShowFooter:         true,
		ShowNewVersions:    true,
		ShowVersionNumber:  true,
	}
}

// ConfigFileUsed returns the path of the loaded config.yaml, if any.
func ConfigFileUsed() string {
	return configFile
}

// IAMDatabasePath returns the configured IAM SQLite path, with a safe default.
func IAMDatabasePath() string {
	if RuntimeConfig.IAM != nil && RuntimeConfig.IAM.Database != "" {
		return RuntimeConfig.IAM.Database
	}
	return "./data/iam.sqlite"
}

// Load reads config.yaml (koanf), applies env overrides, and resolves listen address.
func Load() error {
	k := koanf.New(".")
	setKoanfDefaults(k)
	if err := loadConfigFile(k); err != nil {
		return err
	}
	cfg := defaultConfig()
	if err := k.Unmarshal("", &cfg); err != nil {
		return fmt.Errorf("config unmarshal: %w", err)
	}
	applyEnvOverrides(&cfg)
	if err := resolveListenAddress(&cfg); err != nil {
		return err
	}
	ensureNestedDefaults(&cfg)
	RuntimeConfig = cfg
	return nil
}

func loadConfigFile(k *koanf.Koanf) error {
	path, err := findConfigFile()
	if err != nil {
		return err
	}
	if path == "" {
		log.Info("Config file not found; using defaults")
		return nil
	}
	if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
		return fmt.Errorf("read config %s: %w", path, err)
	}
	configFile = path
	log.WithField("file", path).Info("Starting to read a config")
	return nil
}

func setKoanfDefaults(k *koanf.Koanf) {
	_ = k.Set("listenAddress", "0.0.0.0:8080")
	_ = k.Set("database.driver", "nodriver")
	_ = k.Set("database.hostname", "localhost")
	_ = k.Set("database.username", "user")
	_ = k.Set("database.password", "pass")
	_ = k.Set("database.port", 7687)
	_ = k.Set("iam.database", "./data/iam.sqlite")
	_ = k.Set("wallpaperDirectory", "/opt/wallpaper/")
	_ = k.Set("cert", "wt.crt")
	_ = k.Set("key", "wt.key")
	_ = k.Set("showFooter", true)
	_ = k.Set("showNewVersions", true)
	_ = k.Set("showVersionNumber", true)
	_ = k.Set("auth", map[string]any{})
}

func findConfigFile() (string, error) {
	candidates := []string{
		"config.yaml",
		filepath.Join("..", "config.yaml"),
		filepath.Join("/config", "config.yaml"),
	}
	for _, c := range candidates {
		path, err := resolveConfigCandidate(c)
		if err != nil {
			return "", err
		}
		if path != "" {
			return path, nil
		}
	}
	return "", nil
}

func resolveConfigCandidate(c string) (string, error) {
	info, err := os.Stat(c)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	if info.IsDir() {
		return "", nil
	}
	abs, err := filepath.Abs(c)
	if err != nil {
		return c, nil
	}
	return abs, nil
}

func applyEnvOverrides(cfg *config) {
	ensureConfigStructs(cfg)
	applyDatabaseEnv(cfg.Database)
	applyIAMEnv(cfg.IAM)
}

func ensureConfigStructs(cfg *config) {
	if cfg.Database == nil {
		cfg.Database = &databaseConfig{}
	}
	if cfg.IAM == nil {
		cfg.IAM = &iamConfig{}
	}
}

func applyDatabaseEnv(db *databaseConfig) {
	setEnvString(&db.Driver, "DATABASE_DRIVER")
	setEnvString(&db.Database, "DATABASE_DATABASE")
	setEnvString(&db.Hostname, "DATABASE_HOSTNAME")
	setEnvString(&db.Username, "DATABASE_USERNAME")
	setEnvString(&db.Password, "DATABASE_PASSWORD")
	setEnvInt(&db.Port, "DATABASE_PORT")
}

func applyIAMEnv(iam *iamConfig) {
	setEnvString(&iam.Database, "IAM_DATABASE")
}

func setEnvString(dst *string, envKey string) {
	if v := strings.TrimSpace(os.Getenv(envKey)); v != "" {
		*dst = v
	}
}

func setEnvInt(dst *int, envKey string) {
	v := strings.TrimSpace(os.Getenv(envKey))
	if v == "" {
		return
	}
	if n, err := strconv.Atoi(v); err == nil {
		*dst = n
	}
}

func resolveListenAddress(cfg *config) error {
	if strings.TrimSpace(os.Getenv("PORT")) == "" {
		if strings.TrimSpace(cfg.ListenAddress) == "" {
			cfg.ListenAddress = "0.0.0.0:8080"
		}
		return nil
	}
	addr, err := listenaddr.AvailableListenAddr()
	if err != nil {
		return fmt.Errorf("resolve listen address: %w", err)
	}
	cfg.ListenAddress = addr
	return nil
}

func ensureNestedDefaults(cfg *config) {
	if cfg.Auth == nil {
		cfg.Auth = map[string]any{}
	}
	if cfg.Database == nil {
		cfg.Database = &databaseConfig{Driver: "nodriver"}
	}
	if cfg.IAM == nil {
		cfg.IAM = &iamConfig{Database: "./data/iam.sqlite"}
	}
}
