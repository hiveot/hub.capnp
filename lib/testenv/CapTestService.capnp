# Cap'n proto definition for the test service
@0x9f9160fd7c4ae45f;

using Go = import "/go.capnp";
$Go.package("testsvc");
$Go.import("github.com/hiveot/hub/internal/captest");

using Resolver = import "/Resolver.capnp";


interface CapTestService {
    capMethod1 @0 (clientID :Text, authType :Text) -> (capabilit :CapMethod1Service);
    # obtain the capability to run test method1.
    # FIXME: there should be no need to name result property 'capability'
}

interface CapMethod1Service {
    method1 @0 () -> (forYou :Text);
    # method1 has a message for you
}
