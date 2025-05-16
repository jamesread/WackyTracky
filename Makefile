default: service frontend

service:
	$(MAKE) -wC service

frontend:
	$(MAKE) -wC frontend

go-tools:
	go install "github.com/bufbuild/buf/cmd/buf"
	go install "github.com/fzipp/gocyclo/cmd/gocyclo"
	go install "github.com/go-critic/go-critic/cmd/gocritic"
	go install "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway"
	go install "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2"
	go install "google.golang.org/grpc/cmd/protoc-gen-go-grpc"
	go install "google.golang.org/protobuf/cmd/protoc-gen-go"

certs:
	openssl req -x509 -newkey rsa:4096 -nodes -keyout wt.key -out wt.crt -days 365

.PHONY: default service frontend go-tools certs
