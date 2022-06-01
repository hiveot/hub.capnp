# Makefile to build and test the WoST Hub logger plugin
DIST_FOLDER=./dist
INSTALL_HOME=~/bin/wosthub
.DEFAULT_GOAL := help

all: logger ## Build package with binary distribution and config
	@echo "> Build successful. The executable(s) '$?' can be found in $(DIST_FOLDER)/bin"

logger: ## Build package logger
	go build -o $(DIST_FOLDER)/bin/$@ ./cmd/logger/main.go

clean: ## Clean distribution files
	go mod tidy
	go clean -cache -testcache
	rm -f $(DIST_FOLDER)/bin/*

help: ## Show this help
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

install: all ## Install the plugin into ~/bin/wost/bin and config
	mkdir -p $(INSTALL_HOME)/bin
	mkdir -p $(INSTALL_HOME)/config
	cp $(DIST_FOLDER)/bin/* $(INSTALL_HOME)/bin/
	cp -n $(DIST_FOLDER)/config/* $(INSTALL_HOME)/config/

test: all  ## Run tests (stop on first error, don't run parallel)
	go test -race -failfast -p 1 -cover ./...

upgrade: ## Upgrade packages
	go get -u all
	go mod tidy
