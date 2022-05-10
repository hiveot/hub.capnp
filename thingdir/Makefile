DIST_FOLDER=./dist
INSTALL_HOME=~/bin/wosthub
.DEFAULT_GOAL := help

FORCE: help

all: thingdir ## Build package with binary distribution and config

test: FORCE ## Run tests (stop on first error, don't run parallel)
	go test -race -failfast -p 1 -v ./pkg/...

install: all  ## Install the plugin into ~/bin/wost/bin and config
	cp dist/bin/* $(INSTALL_HOME)/bin/
	cp -n dist/config/* $(INSTALL_HOME)/config/

clean: ## Clean distribution files
	go mod tidy
	go clean
	rm -f $(DIST_FOLDER)/certs/*
	rm -f $(DIST_FOLDER)/logs/*
	rm -f $(DIST_FOLDER)/bin/*
	rm -f $(DIST_FOLDER)/arm/*

thingdir: ## Build thingdir-pb plugin
	mkdir -p $(DIST_FOLDER)/bin
	go build -o $(DIST_FOLDER)/bin/$@ ./cmd/$@/main.go
	@echo "> SUCCESS. Plugin '$@' can be found at $(DIST_FOLDER)/bin/$@ and $(DIST_FOLDER)/arm/$@"

upgrade:
	go get -u all
	go mod tidy

help: ## Show this help
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
