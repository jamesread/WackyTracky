OPENAPI_SRC := ../docs/source/api/openapi.yaml
OPENAPI_EMBED := internal/apidocs/openapi.yaml

cicd: build
	rm wacky-tracky-server

$(OPENAPI_EMBED): $(OPENAPI_SRC)
	cp $(OPENAPI_SRC) $(OPENAPI_EMBED)

build: $(OPENAPI_EMBED)
	go build

test: $(OPENAPI_EMBED)
	go test ./...

codestyle:
	go fmt ./...
	go vet ./...
	gofmt -w .
	go mod tidy
