# Makefile to build and test the HiveOT Hub launcher
BIN_FOLDER=./dist/bin
SERVICE_FOLDER=$(BIN_FOLDER)/services
DIST_FOLDER=./dist
INSTALL_HOME=~/bin/hiveot
.DEFAULT_GOAL := help

.FORCE: 

all: hubcli launcher services   ## Build all

services: certs directory history provisioning state ## Build all services

certs: .FORCE ## Build the certificate management service
	go build -o $(SERVICE_FOLDER)/$@ ./pkg/certs/cmd/main.go

directory: .FORCE ## Build the Thing directory store
	go build -o $(SERVICE_FOLDER)/$@ ./pkg/directory/cmd/main.go

gateway: .FORCE ## Build the Hub gateway
	go build -o $(SERVICE_FOLDER)/$@ ./pkg/gateway/cmd/main.go

history: .FORCE ## Build the Thing value history store
	go build -o $(SERVICE_FOLDER)/$@ ./pkg/history/cmd/main.go

hubcli: .FORCE ## Build Hub CLI
	go build -o $(BIN_FOLDER)/$@ ./cmd/hubcli/main.go

launcher: .FORCE ## Build the Hub Service Launcher
	go build -o $(BIN_FOLDER)/$@ ./pkg/launcher/cmd/main.go

provisioning: .FORCE ## Build Hub provisioning service
	go build -o $(SERVICE_FOLDER)/$@ ./pkg/provisioning/cmd/main.go

state: .FORCE ## Build State store
	go build -o $(SERVICE_FOLDER)/$@ ./pkg/state/cmd/main.go

clean: ## Clean distribution files
	go mod tidy
	go clean -cache -testcache
	rm -f $(BIN_FOLDER)
	rm -f $(SERVICE_FOLDER)
	mkdir -p $(BIN_FOLDER)
	mkdir -p $(SERVICE_FOLDER)

help: ## Show this help
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

install:  all ## build and install the services
	mkdir -p $(INSTALL_HOME)/$(BIN_FOLDER)
	mkdir -p $(INSTALL_HOME)/$(SERVICES_FOLDER)
	mkdir -p $(INSTALL_HOME)/config
	mkdir -p $(INSTALL_HOME)/certs
	mkdir -p $(INSTALL_HOME)/stores
	cp $(BIN_FOLDER)/bin/* $(INSTALL_HOME)/$(BIN_FOLDER)
	cp -n $(DIST_FOLDER)/config/* $(INSTALL_HOME)/config/

test: all  ## Run tests (stop on first error, don't run parallel)
	go test -race -failfast -p 1 -cover ./...

upgrade:
	go get -u all
	go mod tidy
