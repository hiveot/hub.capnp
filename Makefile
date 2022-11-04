# Makefile to build and test the HiveOT Hub launcher
BIN_FOLDER=./dist/bin
SERVICE_FOLDER=$(BIN_FOLDER)/services
DIST_FOLDER=./dist
INSTALL_HOME=~/bin/hiveot
.DEFAULT_GOAL := help

.FORCE: 

all: hubcli launcher services   ## Build all

services: authn authz certs directory history provisioning state ## Build all services

authz: .FORCE ## Build the authorization service
	go build -o $(SERVICE_FOLDER)/$@ ./pkg/$@/cmd/main.go

authn: .FORCE ## Build the authentication service
	go build -o $(SERVICE_FOLDER)/$@ ./pkg/$@/cmd/main.go

certs: .FORCE ## Build the certificate management service
	go build -o $(SERVICE_FOLDER)/$@ ./pkg/$@/cmd/main.go

directory: .FORCE ## Build the Thing directory store
	go build -o $(SERVICE_FOLDER)/$@ ./pkg/$@/cmd/main.go

gateway: .FORCE ## Build the Hub gateway
	go build -o $(SERVICE_FOLDER)/$@ ./pkg/$@/cmd/main.go

history: .FORCE ## Build the Thing value history store
	go build -o $(SERVICE_FOLDER)/$@ ./pkg/$@/cmd/main.go

hubcli: .FORCE ## Build Hub CLI
	go build -o $(BIN_FOLDER)/$@ ./cmd/$@/main.go

launcher: .FORCE ## Build the Hub Service Launcher
	go build -o $(BIN_FOLDER)/$@ ./pkg/$@/cmd/main.go
	cp ./pkg/$@/config/*.yaml $(DIST_FOLDER)/config

provisioning: .FORCE ## Build Hub provisioning service
	go build -o $(SERVICE_FOLDER)/$@ ./pkg/$@/cmd/main.go

state: .FORCE ## Build State store
	go build -o $(SERVICE_FOLDER)/$@ ./pkg/$@/cmd/main.go
	cp ./pkg/$@/config/*.yaml $(DIST_FOLDER)/config

clean: ## Clean distribution files
	go mod tidy
	go clean -cache -testcache
	rm -rf $(DIST_FOLDER)
	mkdir -p $(BIN_FOLDER)
	mkdir -p $(SERVICE_FOLDER)
	mkdir -p $(DIST_FOLDER)/certs
	mkdir -p $(DIST_FOLDER)/config
	mkdir -p $(DIST_FOLDER)/logs
	mkdir -p $(DIST_FOLDER)/run
	mkdir -p $(DIST_FOLDER)/stores

help: ## Show this help
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

install:  all ## build and install the services
	mkdir -p $(INSTALL_HOME)/bin
	mkdir -p $(INSTALL_HOME)/bin/services
	mkdir -p $(INSTALL_HOME)/certs
	mkdir -p $(INSTALL_HOME)/config
	mkdir -p $(INSTALL_HOME)/logs
	mkdir -p $(INSTALL_HOME)/stores
	mkdir -p $(INSTALL_HOME)/run
	cp -a $(BIN_FOLDER)/* $(INSTALL_HOME)/bin
	cp -a $(SERVICE_FOLDER)/* $(INSTALL_HOME)/bin/services
	cp -n $(DIST_FOLDER)/config/* $(INSTALL_HOME)/config/

test: all  ## Run tests (stop on first error, don't run parallel)
	# temp disable race test until races are fixed
	#go test -race -failfast -p 1 ./...
	go test -failfast -p 1  ./...

upgrade:
	go get -u all
	go mod tidy
