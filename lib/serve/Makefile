
FORCE: help

all: FORCE ## This is a library, nothing to build

test: FORCE ## Run tests 
	go test -failfast -v ./pkg/...

clean: ## Clean distribution files
	go mod tidy
	go clean
	rm -f test/certs/*
	rm -f test/logs/*

#deps: ## Build GO dependencies 
#		$(GODEP)

upgrade: ## Upgrade the dependencies to the latest version. Use with care.
	go get -u all
	go mod tidy


help: ## Show this help
		@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
