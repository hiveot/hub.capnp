# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GODEP=$(GOCMD) get
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GORUN=$(GOCMD) run

# Build and Packaging parameters
DIST_FOLDER=./dist
X64_DIST_FOLDER=$(DIST_FOLDER)/bin
ARM_DIST_FOLDER=$(DIST_FOLDER)/arm
PKG_NAME=wost-gateway.tgz
.DEFAULT_GOAL := help

.PHONY: help

all: FORCE ## Build package with binary distribution and config
all: clean gateway 

dist: clean x64  ## Build binary distribution including config
		tar -czf $(PKG_NAME) -C $(DIST_FOLDER) .

test: FORCE ## Run tests (todo fix this)
		$(GOTEST) -v ./pkg/...

clean: ## Clean distribution files
	$(GOCLEAN)
	rm -f test/certs/*
	rm -f test/logs/*
	rm -f $(DIST_FOLDER)/certs/*
	rm -f $(DIST_FOLDER)/logs/*
	rm -f $(X64_DIST_FOLDER)/*
	rm -f $(ARM_DIST_FOLDER)/*
	rm -f debug $(PKG_NAME)
	mkdir -p $(X64_DIST_FOLDER)
	mkdir -p $(ARM_DIST_FOLDER)
	mkdir -p $(DIST_FOLDER)/certs
	mkdir -p $(DIST_FOLDER)/config
	mkdir -p $(DIST_FOLDER)/logs

#deps: ## Build GO dependencies 
#		$(GODEP)

#upgrade: ## Upgrade the dependencies to the latest version. Use with care.
#		go fix

#prof: ## Run application with CPU and memory profiling
#	  $(GORUN) main.go -cpuprofile=cpu.prof -memprofile=mem.prof


# recorder:
# 	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(X64_DIST_FOLDER)/$@ plugins/$@/main.go
# 	GOOS=linux GOARCH=arm $(GOBUILD) -o $(ARM_DIST_FOLDER)/$@ plugins/$@/main.go
# 	@echo "> SUCCESS. Plugin '$@' can be found at $(X64_DIST_FOLDER)/$@ and $(ARM_DIST_FOLDER)/$@"

# gateway: FORCE ## Build gateway for amd64 and arm targets
# 	@echo "building for $(ARCH)"
# 	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(X64_DIST_FOLDER)/gateway cmd/main.go
# 	GOOS=linux GOARCH=arm $(GOBUILD) -o $(ARM_DIST_FOLDER)/gateway cmd/main.go
# 	@echo "> SUCCESS. The Gateway executable '$@' can be found in $(X64_DIST_FOLDER)/$@ and $(ARM_DIST_FOLDER)/$@"


#docker: ## Build gateway for Docker target (TODO untested)
#		docker run --rm -it -v "$(GOPATH)":/go -w /go/src/bitbucket.org/rsohlich/makepost golang:latest go build -o "$(BINARY_AMD64)" -v

help: ## Show this help
		@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

FORCE:
