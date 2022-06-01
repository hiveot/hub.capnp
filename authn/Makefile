# Makefile to build and test the WoST Hub authn plugin
DIST_FOLDER=./dist
INSTALL_HOME=~/bin/wosthub
.DEFAULT_GOAL := help

all: authn authservice  ## Build auth and authservice
	@echo "> Build successful. The executable(s) '$?' can be found in $(DIST_FOLDER)/bin"

authn: ## Build authentication CLI to generate certificates, set password and set ACLs
	mkdir -p $(DIST_FOLDER)/bin
	go build -o $(DIST_FOLDER)/bin/$@ ./cmd/$@/main.go

authservice: ## Build authentication service to run at startup
	go build -o $(DIST_FOLDER)/bin/$@ ./cmd/$@/main.go

clean: ## Clean distribution files
	go mod tidy
	go clean -cache -testcache
	rm -f $(DIST_FOLDER)/bin/*

help: ## Show this help
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

install:  all ## Install the auth app
	mkdir -p $(INSTALL_HOME)/bin
	mkdir -p $(INSTALL_HOME)/config
	cp -f $(DIST_FOLDER)/bin/* $(INSTALL_HOME)/bin
	cp -n $(DIST_FOLDER)/config/* $(INSTALL_HOME)/config/

test: all  ## Run tests (stop on first error, don't run parallel)
	go test -race -failfast -p 1 -cover ./...

upgrade:
	go get -u all
	go mod tidy
