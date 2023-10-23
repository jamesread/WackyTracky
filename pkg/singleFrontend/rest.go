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
		/*
			runtime.WithMetadata(func(ctx, context.Context, request *http.Request) metadata.MD {
				return md
			}
		*/
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.HTTPBodyMarshaler{
			Marshaler: &runtime.JSONPb{
				MarshalOptions: protojson.MarshalOptions{
					UseProtoNames:   true,
					EmitUnpopulated: true,
				},
			},
		}),
	)

	opts := []grpc.DialOption{grpc.WithInsecure()}

	err := gw.RegisterWackyTrackyClientApiHandlerFromEndpoint(ctx, mux, "0.0.0.0:8083", opts)

	if err != nil {
		log.Errorf("Register REST: %v", err)
	}

	return http.ListenAndServe("0.0.0.0:8082", allowCors(mux))
}

func allowCors(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		h.ServeHTTP(w, r)
	})
}
