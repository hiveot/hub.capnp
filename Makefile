# Makefile to build and test the HiveOT Hub launcher
DIST_FOLDER=./dist
INSTALL_HOME=~/bin/hiveot
.DEFAULT_GOAL := help

.FORCE: 

all: certs directory gateway history provisioning hubcli    ## Build all services

certs: .FORCE ## Build the certificate management service
	go build -o $(DIST_FOLDER)/bin/$@ ./pkg/certservice/*.go
	@echo "> Build successful. The executable '$@' can be found in $(DIST_FOLDER)/bin/$@"

directory: .FORCE ## Build the Thing directory service
	go build -o $(DIST_FOLDER)/bin/$@ ./pkg/directorystore/main.go
	@echo "> Build successful. The executable '$@' can be found in $(DIST_FOLDER)/bin/$@"

gatewy: .FORCE ## Build the Hub gateway
	go build -o $(DIST_FOLDER)/bin/$@ ./pkg/gateway/main.go
	@echo "> Build successful. The executable '$@' can be found in $(DIST_FOLDER)/bin/$@"

history: .FORCE ## Build the Thing value history service
	go build -o $(DIST_FOLDER)/bin/$@ ./pkg/svc/historystore/main.go
	@echo "> Build successful. The executable '$@' can be found in $(DIST_FOLDER)/bin/$@"

hubcli: .FORCE ## Build Hub CLI
	go build -o $(DIST_FOLDER)/bin/$@ ./cmd/hubcli/main.go
	@echo "> SUCCESS. The executable '$@' can be found in $(BIN_FOLDER)/$@"

provisioning: .FORCE ## Build Hub provisioning service
	go build -o $(DIST_FOLDER)/bin/$@ ./pkg/svc/provisioningservice/main.go
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
