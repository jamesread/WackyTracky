default: generate service frontend

service:
	$(MAKE) -wC service

frontend:
	$(MAKE) -wC frontend

generate:
	$(MAKE) -C protocol

docs:
	$(MAKE) -C docs

it: integration-tests

integration-tests:
	$(MAKE) -C integration-tests

test: service
	$(MAKE) -wC service test

go-tools:
	go install "github.com/bufbuild/buf/cmd/buf"
	go install "github.com/fzipp/gocyclo/cmd/gocyclo"
	go install "github.com/go-critic/go-critic/cmd/gocritic"
	go install "google.golang.org/protobuf/cmd/protoc-gen-go"

certs:
	openssl req -x509 -newkey rsa:4096 -nodes -keyout wt.key -out wt.crt -days 365

# Marketing screenshots (server must be running on http://localhost:8080; optional REPO_COMMON)
marketing-screenshots:
	python3 var/marketing/scripts/take-marketing-screenshots.py

.PHONY: default service frontend generate docs integration-tests test go-tools certs marketing-screenshots
