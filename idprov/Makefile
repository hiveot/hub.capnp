.DEFAULT_GOAL := help
DIST_FOLDER=./dist
INSTALL_HOME=~/bin/wosthub

FORCE: help

all: oob idprov ## make IDProv server and oob utility

clean: ## Clean distribution files
	go mod tidy
	go clean
	rm -f test/certs/*
	rm -f test/client/*
	rm -f test/logs/*

test: FORCE ## Run tests (stop on first error, don't run parallel)
	go test -race -failfast -p 1 -v ./pkg/...

oob: ## Build the oob utility
	mkdir -p $(DIST_FOLDER)/bin
	go build -o $(DIST_FOLDER)/bin/$@ ./cmd/$@/main.go

idprov: ## Build idprov server plugin
	mkdir -p $(DIST_FOLDER)/bin
	go build -o $(DIST_FOLDER)/bin/$@ ./cmd/$@/main.go


install: all ## Install idprov server and oob into ~/bin/wost/bin
	mkdir -p $(INSTALL_HOME)/bin
	mkdir -p $(INSTALL_HOME)/config
	mkdir -p $(INSTALL_HOME)/logs
	cp $(DIST_FOLDER)/bin/* $(INSTALL_HOME)/bin/
	cp -n $(DIST_FOLDER)/config/* $(INSTALL_HOME)/config/

upgrade:
	go get -u all
	go mod tidy

help: ## Show this help
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
