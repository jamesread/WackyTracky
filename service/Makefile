cicd: build
	rm wacky-tracky-server

build:
	go build

test:
	go test ./...

codestyle:
	go fmt ./...
	go vet ./...
	gofmt -w .
	go mod tidy
