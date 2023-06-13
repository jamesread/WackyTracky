package runtimeconfig

type config struct {
	ListenAddressGrpc string
	ListenAddressRest string
	ListenAddressWebUI string
	ListenAddressSingleHTTPFrontend string
}

var RuntimeConfig = config{
	ListenAddressGrpc: "localhost:8083",
	ListenAddressRest: "localhost:8082",
	ListenAddressWebUI: "localhost:8084",
	ListenAddressSingleHTTPFrontend: "0.0.0.0:8080",
};


