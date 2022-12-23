# Makefile to build the cap'n proto interface of Hub core services
CAPNP_GO=capnp compile "-I$(GOPATH)/src/capnproto.org/go/capnp/std" -ogo:./go/ --src-prefix=capnp/
.DEFAULT_GOAL := help

.FORCE:

# Capnproto RPC. This needs go-capnproto2 installed
go: .FORCE ## Compile cap'n proto to go (testing capnp)
	$(CAPNP_GO)  ./capnp/hubapi/Resolver.capnp
	$(CAPNP_GO)  ./capnp/hubapi/Authn.capnp
	$(CAPNP_GO)  ./capnp/hubapi/Authz.capnp
	$(CAPNP_GO)  ./capnp/hubapi/Bucket.capnp
	$(CAPNP_GO)  ./capnp/hubapi/Certs.capnp
	$(CAPNP_GO)  ./capnp/hubapi/Directory.capnp
	$(CAPNP_GO)  ./capnp/hubapi/Gateway.capnp
	$(CAPNP_GO)  ./capnp/hubapi/History.capnp
	$(CAPNP_GO)  ./capnp/hubapi/Launcher.capnp
	$(CAPNP_GO)  ./capnp/hubapi/Provisioning.capnp
	$(CAPNP_GO)  ./capnp/hubapi/PubSub.capnp
	$(CAPNP_GO)  ./capnp/hubapi/State.capnp
	$(CAPNP_GO)  ./capnp/hubapi/Thing.capnp
	$(CAPNP_GO)  ./capnp/vocab/HiveVocabulary.capnp
	$(CAPNP_GO)  ./capnp/vocab/WoTVocabulary.capnp
	cd go && go mod tidy


help: ## Show this help
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
