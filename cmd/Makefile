# Makefile to build and test the Hub CLI
BIN_FOLDER=./bin
INSTALL_HOME=~/bin/wosthub

.DEFAULT_GOAL := help

.FORCE:

all: certs launcher  ## Build cert management CLI
	@echo "> Build successful. The executable(s) '$?' can be found in $(DIST_FOLDER)/bin"

certs: .FORCE  ## Build certs CLI to generate certificates
	go build -o $(BIN_FOLDER)/$@ ./$@/main.go
	@echo "> SUCCESS. The executable '$@' can be found in $(BIN_FOLDER)/$@"

launcher: .FORCE ## Build Hub services launcher
	go build -o $(BIN_FOLDER)/$@ ./$@/main.go
	@echo "> SUCCESS. The executable '$@' can be found in $(BIN_FOLDER)/$@"


clean: ## Clean distribution files
	mkdir -p $(BIN_FOLDER)
	go mod tidy
	go clean -cache -testcache
	rm -f $(BIN_FOLDER)/*

help: ## Show this help
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

install:  all ## Install the binaries in the $(INSTALL_HOME)/bin installation folder
	mkdir -p $(INSTALL_HOME)/bin
	mkdir -p $(INSTALL_HOME)/config
	cp $(BIN_FOLDER)/* $(INSTALL_HOME)/bin
	cp $(BIN_FOLDER)/* $(INSTALL_HOME)/bin
	cp -n launcher/launcher.yaml $(INSTALL_HOME)/config/

test: all  ## Run tests (stop on first error, don't run parallel)
	go test -race -failfast -p 1 -cover ./...

upgrade:
	go get -u all
	go mod tidy
