# Makefile to build and test the WoST Hub
# To build including generating certificates: make all
DIST_FOLDER=./dist
INSTALL_HOME=~/bin/wosthub
PKG_NAME="certs"

.DEFAULT_GOAL := help

.PHONY: 

all: certs  ## Build cert management CLI

install:  all ## Install the certs app 
	mkdir -p $(INSTALL_HOME)/bin
	mkdir -p $(INSTALL_HOME)/config
	mkdir -p $(INSTALL_HOME)/logs
	cp $(DIST_FOLDER)/bin/* $(INSTALL_HOME)/bin
	#cp -n $(DIST_FOLDER)/config/* $(INSTALL_HOME)/config/  

dist: clean   ## Build binary distribution tarball 
		tar -czf $(PKG_NAME) -C $(DIST_FOLDER) .

test: all .PHONY ## Run tests sequentially
	go test -race -failfast -p 1 -cover ./...

clean: ## Clean distribution files
	go mod tidy
	go clean -cache -testcache
	rm -f test/certs/*
	rm -f test/logs/*
	rm -f $(DIST_FOLDER)/bin/*
	mkdir -p $(DIST_FOLDER)/bin
	mkdir -p $(DIST_FOLDER)/config

certs: ## Build certs CLI to generate certificates
	mkdir -p $(DIST_FOLDER)/bin
	go build -o $(DIST_FOLDER)/bin/$@ ./cmd/$@/main.go
	@echo "> SUCCESS. The executable '$@' can be found in $(DIST_FOLDER)/bin/$@"


upgrade:
	go get -u all
	go mod tidy

help: ## Show this help
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
