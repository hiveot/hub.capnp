# Makefile to build and test the HiveOT Hub launcher
DIST_FOLDER=./dist
INSTALL_HOME=~/bin/hiveot
.DEFAULT_GOAL := help

.FORCE: 

all: certservice directorystore historystore provisioning hubcli gateway  ## Build all services

certs: .FORCE ## Build the certificate management service
	go build -o $(DIST_FOLDER)/bin/$@ ./pkg/certservice/cmd/main.go

directory: .FORCE ## Build the Thing directory store
	go build -o $(DIST_FOLDER)/bin/$@ ./pkg/directory/cmd/main.go

gateway: .FORCE ## Build the Hub gateway
	go build -o $(DIST_FOLDER)/bin/$@ ./pkg/gateway/main.go

history: .FORCE ## Build the Thing value history store
	go build -o $(DIST_FOLDER)/bin/$@ ./pkg/history/main.go

hubcli: .FORCE ## Build Hub CLI
	go build -o $(DIST_FOLDER)/bin/$@ ./cmd/hubcli/main.go

provisioning: .FORCE ## Build Hub provisioning service
	go build -o $(DIST_FOLDER)/bin/$@ ./pkg/provisioning/main.go



clean: ## Clean distribution files
	go mod tidy
	go clean -cache -testcache
	rm -f $(DIST_FOLDER)/bin/*

help: ## Show this help
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

install:  all ## build and install the services
	mkdir -p $(INSTALL_HOME)/bin
	mkdir -p $(INSTALL_HOME)/config
	cp $(DIST_FOLDER)/bin/* $(INSTALL_HOME)/bin
	cp -n $(DIST_FOLDER)/config/* $(INSTALL_HOME)/config/

test: all  ## Run tests (stop on first error, don't run parallel)
	go test -race -failfast -p 1 -cover ./...

upgrade:
	go get -u all
	go mod tidy
