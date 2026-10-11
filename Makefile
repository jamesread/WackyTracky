default: generate service frontend

service:
	$(MAKE) -wC service

frontend:
	$(MAKE) -wC frontend ci

generate:
	$(MAKE) -C protocol

docs:
	$(MAKE) -C docs ci
	./docs/node_modules/.bin/antora antora-playbook.yml --to-dir build/site

it: integration-tests

integration-tests:
	$(MAKE) -C integration-tests

test: service
	$(MAKE) -wC service test
	$(MAKE) -wC frontend test

lint: codestyle

codestyle:
	$(MAKE) -wC service codestyle

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

.PHONY: default service frontend generate docs integration-tests test lint codestyle go-tools certs marketing-screenshots
