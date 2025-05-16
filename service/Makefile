default:
	go build

codestyle:
	go fmt ./...
	go vet ./...
	gofmt -w .
	go mod tidy
