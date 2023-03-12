# Makefile to build and test the HiveOT Hub launcher
BIN_FOLDER=./dist/bin
CAPNP_SRC=./api/capnp
CAPNP_GO=capnp compile "-I$(GOPATH)/src/capnproto.org/go/capnp/std" -ogo:./api/go/ --src-prefix=api/capnp/
SERVICE_FOLDER=$(BIN_FOLDER)/services
DIST_FOLDER=./dist
INSTALL_HOME=~/bin/hiveot
.DEFAULT_GOAL := help

.FORCE: 

all: api hub  ## Build APIs, CLI, Hub services

hub: hubcli launcher services   ## Build hub services and cli

api: hubapi-go  ## Build the hub api for all languages (currently only golang)

hubapi-go: .FORCE  ## Build the golang API from capnp
	$(CAPNP_GO)  $(CAPNP_SRC)/hubapi/Resolver.capnp
	$(CAPNP_GO)  $(CAPNP_SRC)/hubapi/Authn.capnp
	$(CAPNP_GO)  $(CAPNP_SRC)/hubapi/Authz.capnp
	$(CAPNP_GO)  $(CAPNP_SRC)/hubapi/Bucket.capnp
	$(CAPNP_GO)  $(CAPNP_SRC)/hubapi/Certs.capnp
	$(CAPNP_GO)  $(CAPNP_SRC)/hubapi/Directory.capnp
	$(CAPNP_GO)  $(CAPNP_SRC)/hubapi/Gateway.capnp
	$(CAPNP_GO)  $(CAPNP_SRC)/hubapi/History.capnp
	$(CAPNP_GO)  $(CAPNP_SRC)/hubapi/Launcher.capnp
	$(CAPNP_GO)  $(CAPNP_SRC)/hubapi/Provisioning.capnp
	$(CAPNP_GO)  $(CAPNP_SRC)/hubapi/PubSub.capnp
	$(CAPNP_GO)  $(CAPNP_SRC)/hubapi/State.capnp
	$(CAPNP_GO)  $(CAPNP_SRC)/hubapi/Thing.capnp
	$(CAPNP_GO)  $(CAPNP_SRC)/vocab/HiveVocabulary.capnp
	$(CAPNP_GO)  $(CAPNP_SRC)/vocab/WoTVocabulary.capnp
	go mod tidy

services: authn authz certs directory gateway history provisioning pubsub resolver state ## Build all services

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

pubsub: .FORCE ## Build the pubsub messaging service
	go build -o $(SERVICE_FOLDER)/$@ ./pkg/$@/cmd/main.go

resolver: .FORCE ## Build resolver service
	go build -o $(SERVICE_FOLDER)/$@ ./pkg/$@/cmd/main.go

state: .FORCE ## Build State store
	go build -o $(SERVICE_FOLDER)/$@ ./pkg/$@/cmd/main.go
	cp ./pkg/$@/config/*.yaml $(DIST_FOLDER)/config

clean: ## Clean distribution files
	go clean -cache -testcache -modcache
	rm -rf $(DIST_FOLDER)
	mkdir -p $(BIN_FOLDER)
	mkdir -p $(SERVICE_FOLDER)
	mkdir -p $(DIST_FOLDER)/certs
	mkdir -p $(DIST_FOLDER)/config
	mkdir -p $(DIST_FOLDER)/logs
	mkdir -p $(DIST_FOLDER)/run
	mkdir -p $(DIST_FOLDER)/stores
	go mod tidy
	go get all

help: ## Show this help
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

setup: ## Setup the capnp build environment
	go get capnproto.org/go/capnp/v3
	go install capnproto.org/go/capnp/v3/capnpc-go@latest
	#GO111MODULE=off go get -u capnproto.org/go/capnp/v3/

install:  hub  ## build and install the services
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

test: hub  ## Run tests (stop on first error, don't run parallel)
	go test -race -failfast -p 1 ./...

upgrade:
	go get -u all
	go mod tidy
