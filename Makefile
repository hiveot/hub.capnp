# Makefile to build the cap'n proto interface of Hub core services
CAPNP_GO=capnp compile "-I$(GOPATH)/src/capnproto.org/go/capnp/std" -ogo:./go/ --src-prefix=capnp/
.DEFAULT_GOAL := help

.FORCE:

# Capnproto RPC. This needs go-capnproto2 installed
svc: .FORCE ## Compile cap'n proto (testing capnp)
	$(CAPNP_GO)  ./capnp/svc/CertSvc.capnp 
	$(CAPNP_GO)  ./capnp/svc/EventHistory.capnp 
	$(CAPNP_GO)  ./capnp/svc/PropertyStore.capnp 
	$(CAPNP_GO)  ./capnp/svc/Provisioning.capnp 
	$(CAPNP_GO)  ./capnp/svc/ThingDirectory.capnp 
	$(CAPNP_GO)  ./capnp/svc/Gateway.capnp 


go: svc  ## Compile hub files for go
	cd go && go mod tidy



help: ## Show this help
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
