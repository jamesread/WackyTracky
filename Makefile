default: generate service frontend

service:
	$(MAKE) -wC service

frontend:
	$(MAKE) -wC frontend

generate:
	$(MAKE) -C protocol

docs:
	$(MAKE) -C docs

integration-tests:
	$(MAKE) -C integration-tests

go-tools:
	go install "github.com/bufbuild/buf/cmd/buf"
	go install "github.com/fzipp/gocyclo/cmd/gocyclo"
	go install "github.com/go-critic/go-critic/cmd/gocritic"
	go install "google.golang.org/protobuf/cmd/protoc-gen-go"

certs:
	openssl req -x509 -newkey rsa:4096 -nodes -keyout wt.key -out wt.crt -days 365

.PHONY: default service frontend generate docs integration-tests go-tools certs
