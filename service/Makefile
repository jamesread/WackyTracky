cicd: build
        rm wacky-tracky-server

build:
	go build

codestyle:
	go fmt ./...
	go vet ./...
	gofmt -w .
	go mod tidy
