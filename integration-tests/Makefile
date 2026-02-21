default: test-install test-run

test-install:
	npm install --no-fund

test-run:
	npx mocha tests --recursive -t 15000

.PHONY: default test-install test-run
