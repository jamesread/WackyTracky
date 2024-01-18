package singleFrontend

import (
	"context"
	log "github.com/sirupsen/logrus"

	"google.golang.org/grpc"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	gw "github.com/wacky-tracky/wacky-tracky-server/gen/grpc"
	"google.golang.org/protobuf/encoding/protojson"
)

func startRestGateway() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption("application/json+pretty", &runtime.HTTPBodyMarshaler{
			Marshaler: &runtime.JSONPb{
				MarshalOptions: protojson.MarshalOptions{
					UseProtoNames:   false,
					Indent:          "\t",
					Multiline:       true,
					EmitUnpopulated: true,
				},
			},
		}),
	)

	opts := []grpc.DialOption{grpc.WithInsecure()}

	err := gw.RegisterWackyTrackyClientServiceHandlerFromEndpoint(ctx, mux, "0.0.0.0:8083", opts)

	if err != nil {
		log.Errorf("Register REST: %v", err)
	}

	return http.ListenAndServe("0.0.0.0:8082", prettier(allowCors(mux)))
}

// https://grpc-ecosystem.github.io/grpc-gateway/docs/mapping/customizing_your_gateway/#pretty-print-json-responses-when-queried-with-pretty
func prettier(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// checking Values as map[string][]string also catches ?pretty and ?pretty=
		// r.URL.Query().Get("pretty") would not.
		if _, ok := r.URL.Query()["pretty"]; ok {
			r.Header.Set("Accept", "application/json+pretty")
		}
		h.ServeHTTP(w, r)
	})
}

func allowCors(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		h.ServeHTTP(w, r)
	})
}
