#!/bin/bash



# compile the test service capnp file
capnp compile -I${GOPATH}/src/capnproto.org/go/capnp/std -I../../../hub.capnp/capnp/hubapi -ogo:./ ./CapTestService.capnp
