# Makefile to build the Hub WASM for JS clients
DIST_FOLDER=../../dist/bin
INSTALL_HOME=~/bin/hiveot
.DEFAULT_GOAL := help

.FORCE:

all: wasm

wasm: ## Compile go to wasm
	GOOS=js GOARCH=wasm go build -o ${DIST_FOLDER}/hapi.wasm ./main.go

help: ## Show this help
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

test: ## Run tests
	GOOS=js GOARCH=wasm go test -failfast  -cover ./...

upgrade: ## Upgrade packages (use with care)
	go get -u all
	go mod tidy
