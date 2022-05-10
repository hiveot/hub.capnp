DIST_FOLDER=./dist
INSTALL_HOME=~/bin/wosthub
.DEFAULT_GOAL := help

.PHONY: 

all: logger ## Build package with binary distribution and config

install: all ## Install the plugin into ~/bin/wost/bin and config
	cp $(DIST_FOLDER)/bin/* $(INSTALL_HOME)/bin/
	cp -n $(DIST_FOLDER)/config/* $(INSTALL_HOME)/config/

logger:
	go build -o $(DIST_FOLDER)/bin/$@ ./cmd/logger/main.go
	@echo "> SUCCESS. Plugin '$@' can be found at $(DIST_FOLDER)/bin/$@"

test: .PHONY
	go test -race -failfast -p 1 -v ./...

clean: ## Clean distribution files
	go mod tidy
	go clean
	rm -f $(DIST_FOLDER)/certs/*
	rm -f $(DIST_FOLDER)/logs/*
	rm -f $(DIST_FOLDER)/bin/*
	rm -f ./test/certs/*
	rm -f ./test/logs/*
	rm -f ./test/bin/*
	

upgrade: ## Upgrade packages
	go get -u all
	go mod tidy

help: ## Show this help
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
