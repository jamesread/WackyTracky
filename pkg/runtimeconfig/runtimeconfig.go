package runtimeconfig

type config struct {
	ListenAddressGrpc               string
	ListenAddressRest               string
	ListenAddressWebUI              string
	ListenAddressSingleHTTPFrontend string
	DB                              string
}

var RuntimeConfig = config{
	DB:                              "",
	ListenAddressGrpc:               "0.0.0.0:8083",
	ListenAddressRest:               "0.0.0.0:8082",
	ListenAddressWebUI:              "0.0.0.0:8084",
	ListenAddressSingleHTTPFrontend: "0.0.0.0:8080",
}
