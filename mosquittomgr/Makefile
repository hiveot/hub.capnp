# Makefile to build and test mosquittomg hub plugin
DIST_FOLDER=./dist
INSTALL_HOME=~/bin/wosthub
.DEFAULT_GOAL := help

FORCE: help

all: mosquittomgr mosqauth ## make the manager app and mosquitto authenticator

install:  all ## Install the mosquitto manager app and plugin into INSTALL_HOME folder
	mkdir -p $(INSTALL_HOME)/bin
	mkdir -p $(INSTALL_HOME)/config
	mkdir -p $(INSTALL_HOME)/logs
	cp $(DIST_FOLDER)/bin/* $(INSTALL_HOME)/bin
	cp $(DIST_FOLDER)/config/mosquitto.conf.template $(INSTALL_HOME)/config/  
	cp -n $(DIST_FOLDER)/config/* $(INSTALL_HOME)/config/  

clean: ## Clean distribution files
	go mod tidy
	go clean
	rm -f test/certs/*
	rm -f test/client/*
	rm -f test/logs/*
	rm -f $(DIST_FOLDER)/certs/*
	rm -f $(DIST_FOLDER)/logs/*
	rm -f $(DIST_FOLDER)/bin/*
	rm -f $(DIST_FOLDER)/config/mosquitto.conf
	mkdir -p $(DIST_FOLDER)/bin
	mkdir -p $(DIST_FOLDER)/certs
	mkdir -p $(DIST_FOLDER)/config
	mkdir -p $(DIST_FOLDER)/logs


test: mosqauth FORCE ## Run tests (stop on first error, don't run parallel)
	go test -failfast -p 1 -v ./internal/...

mosquittomgr: mosqauth ## Build mosquitto configuration manager
	mkdir -p $(DIST_FOLDER)/bin
	go build -o $(DIST_FOLDER)/bin/$@ ./cmd/$@/main.go
	
mosqauth: ## Build mosquitto auth plugin for use by the Hub
	mkdir -p $(DIST_FOLDER)/bin
	cd cmd/mosqauth && make
	@echo "> SUCCESS. The executable '$@' can be found in $(DIST_FOLDER)/bin/$@"

upgrade:
	go get -u all
	go mod tidy


help: ## Show this help
		@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
