# Makefile to build and test the WoST Hub
# To build including generating certificates: make all
DIST_FOLDER=./dist
PKG_NAME=wost-hub.tgz
INSTALL_HOME=~/bin/wost
.DEFAULT_GOAL := help

.FORCE: 

all: hub core addons  ## Build the hub and plugins

install:  all ## Install the hub into ~/bin/wost/bin and config
	mkdir -p $(INSTALL_HOME)/bin
	mkdir -p $(INSTALL_HOME)/config
	mkdir -p $(INSTALL_HOME)/logs
	cp $(DIST_FOLDER)/bin/* ~/bin/wost/bin/
	cp -n $(DIST_FOLDER)/config/* ~/bin/wost/config/  

dist: clean   ## Build binary distribution tarball 
		tar -czf $(PKG_NAME) -C $(DIST_FOLDER) .

test: hub .FORCE ## Run hub test
	go test -failfast -p 1 -v ./...

clean: ## Clean distribution files
	go mod tidy
	go clean
	rm -f test/certs/*
	rm -f test/logs/*
	rm -f $(DIST_FOLDER)/certs/*
	rm -f $(DIST_FOLDER)/logs/*
	rm -f $(DIST_FOLDER)/bin/*
	rm -f debug $(PKG_NAME)
	mkdir -p $(DIST_FOLDER)/bin
	mkdir -p $(DIST_FOLDER)/certs
	mkdir -p $(DIST_FOLDER)/config
	mkdir -p $(DIST_FOLDER)/logs

hub: ## Build WoST Hub
	go build -o $(DIST_FOLDER)/bin/$@ ./cmd/$@/main.go
	@echo "> SUCCESS. The executable '$@' can be found in $(DIST_FOLDER)/bin/$@"

help: ## Show this help
		@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'


# Core plugins to build 
core: hubauth idprov-go mosquittomgr thingdir-go ## Core plugins to build

# Addon plugins
addons: logger owserver-pb  ## Addon plugins to build

hubauth: .FORCE ## plugin authentication utility
	make -C ../$@ all
	cp ../$@/$(DIST_FOLDER)/bin/* $(DIST_FOLDER)/bin
	cp ../$@/$(DIST_FOLDER)/config/* $(DIST_FOLDER)/config

idprov-go: .FORCE ## plugin provisioning service
	make -C ../$@ all
	cp ../$@/$(DIST_FOLDER)/bin/* $(DIST_FOLDER)/bin
	cp ../$@/$(DIST_FOLDER)/config/* $(DIST_FOLDER)/config

logger: .FORCE ## plugin simple message file logger
	make -C ../$@ all
	cp ../$@/$(DIST_FOLDER)/bin/* $(DIST_FOLDER)/bin
	cp ../$@/$(DIST_FOLDER)/config/* $(DIST_FOLDER)/config

mosquittomgr: .FORCE ## plugin mosquitto manager
	make -C ../$@ all
	cp ../$@/$(DIST_FOLDER)/bin/* $(DIST_FOLDER)/bin
	cp ../$@/$(DIST_FOLDER)/config/* $(DIST_FOLDER)/config

owserver-pb: .FORCE ## plugin owserver protocol binding
	make -C ../$@ all
	cp ../$@/$(DIST_FOLDER)/bin/* $(DIST_FOLDER)/bin
	cp ../$@/$(DIST_FOLDER)/config/* $(DIST_FOLDER)/config

thingdir-go: .FORCE ## plugin Thing Directory
	make -C ../$@ all
	cp ../$@/$(DIST_FOLDER)/bin/* $(DIST_FOLDER)/bin
	cp ../$@/$(DIST_FOLDER)/config/* $(DIST_FOLDER)/config
