# Makefile to build and test the WoST Core plugins 
DIST_FOLDER=../dist
INSTALL_HOME=~/bin/wost
.DEFAULT_GOAL := help

.FORCE: 


install: .FORCE ## Install the plugins as git submodules
	git submodule init
	git submodule update


# Core plugins to build 
all: hubauth idprov-go mosquittomgr thingdir-go ## Build core plugins

hubauth: .FORCE ## plugin authentication utility
	make -C ./$@ all
	cp -a ./$@/dist/bin/* $(DIST_FOLDER)/bin
	cp -a ./$@/dist/config/* $(DIST_FOLDER)/config

idprov-go: .FORCE ## plugin provisioning service
	make -C ./$@ all
	cp -a ./$@/dist/bin/* $(DIST_FOLDER)/bin
	cp -a ./$@/dist/config/* $(DIST_FOLDER)/config

mosquittomgr: .FORCE ## plugin mosquitto manager
	make -C ./$@ all
	cp -a ./$@/dist/bin/* $(DIST_FOLDER)/bin
	cp -a ./$@/dist/config/* $(DIST_FOLDER)/config

thingdir-go: .FORCE ## plugin Thing Directory
	make -C ./$@ all
	cp -a ./$@/dist/bin/* $(DIST_FOLDER)/bin
	cp -a ./$@/dist/config/* $(DIST_FOLDER)/config



help: ## Show this help
		@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
