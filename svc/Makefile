# Makefile to build and test the WoST Hub launcher
DIST_FOLDER=../dist
INSTALL_HOME=~/bin/wosthub
.DEFAULT_GOAL := help

.FORCE: 

all: certsvc  historystore thingstore   ## Build all services

certsvc: .FORCE ## Build Hub certificate management service
	go build -o $(DIST_FOLDER)/bin/$@ ./certsvc/main.go
	@echo "> Build successful. The executable '$@' can be found in $(DIST_FOLDER)/bin/$@"

historystore: .FORCE ## Build Hub history-store service
	go build -o $(DIST_FOLDER)/bin/$@ ./historystore/main.go
	@echo "> Build successful. The executable '$@' can be found in $(DIST_FOLDER)/bin/$@"


thingstore: .FORCE ## Build Hub thingstore service
	go build -o $(DIST_FOLDER)/bin/$@ ./thingstore/main.go
	@echo "> Build successful. The executable '$@' can be found in $(DIST_FOLDER)/bin/$@"

clean: ## Clean distribution files
	go mod tidy
	go clean -cache -testcache
	rm -f $(DIST_FOLDER)/bin/*

help: ## Show this help
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

install:  all ## Install the service
	mkdir -p $(INSTALL_HOME)/bin
	mkdir -p $(INSTALL_HOME)/config
	cp $(DIST_FOLDER)/bin/* $(INSTALL_HOME)/bin
	cp -n $(DIST_FOLDER)/config/* $(INSTALL_HOME)/config/

test: all  ## Run tests (stop on first error, don't run parallel)
	go test -race -failfast -p 1 -cover ./...

upgrade:
	go get -u all
	go mod tidy
