package runtimeconfig

type config struct {
	ListenAddressGrpc               string
	ListenAddressRest               string
	ListenAddressWebUI              string
	ListenAddressSingleHTTPFrontend string
	Database                        *databaseConfig
	WallpaperDirectory              string
}

type databaseConfig struct {
	Driver   string
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
	ListenAddressGrpc:               "0.0.0.0:8083",
	ListenAddressRest:               "0.0.0.0:8082",
	ListenAddressWebUI:              "0.0.0.0:8084",
	ListenAddressSingleHTTPFrontend: "0.0.0.0:8080",
	WallpaperDirectory:              "/opt/wallpaper/",
}
