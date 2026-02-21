package runtimeconfig

type config struct {
	ListenAddress      string
	Database           *databaseConfig
	WallpaperDirectory string
	Cert               string
	Key                string
}

type databaseConfig struct {
	Driver   string
	Database string // path for drivers like todotxt
	Hostname string
	Username string
	Password string
	Port     int
}

var RuntimeConfig = config{
	Database: &databaseConfig{
		Driver:   "nodriver",
		Hostname: "localhost",
		Username: "user",
		Password: "pass",
		Port:     7687,
	},
	ListenAddress:      "0.0.0.0:8080",
	WallpaperDirectory: "/opt/wallpaper/",
	Cert:               "wt.crt",
	Key:                "wt.key",
}
