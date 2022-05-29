# Makefile to build and test the WoST Hub core services
DIST_FOLDER=./dist
PKG_NAME=wosthub.tgz
INSTALL_HOME=~/bin/wosthub
.DEFAULT_GOAL := help

.FORCE: 

all: authn authz certs idprov launcher logger mosquittomgr thingdir  ## Build the launcher and core plugins

install:  all ## Install the launcher into ~/bin/wost/bin and config
	mkdir -p $(INSTALL_HOME)/bin
	mkdir -p $(INSTALL_HOME)/config
	mkdir -p $(INSTALL_HOME)/logs
	cp $(DIST_FOLDER)/bin/* $(INSTALL_HOME)/bin/
	cp -n $(DIST_FOLDER)/config/* $(INSTALL_HOME)/config/  

dist: clean   ## Build binary distribution tarball 
		tar -czf $(PKG_NAME) -C $(DIST_FOLDER) .

clean: ## Clean distribution files
	rm -f $(DIST_FOLDER)/bin/*
	rm -f $(DIST_FOLDER)/config/*
	mkdir -p $(DIST_FOLDER)/bin
	mkdir -p $(DIST_FOLDER)/config

authn authz certs idprov launcher logger mosquittomgr thingdir:  .FORCE ## Build Hub services 
	make -C $@ all
	cp $@/dist/bin/* dist/bin
	cp $@/dist/config/* dist/config


test: clean ## Test all plugins
	make -C authn test
	make -C authz test
	make -C certs test
	make -C idprov test
	make -C launcher test
	make -C logger test
	make -C mosquittomgr test
	make -C thingdir test


help: ## Show this help
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'


addons: logger owserver-pb   ## Build addon plugins
