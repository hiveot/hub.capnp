# Cap'n proto definition for thing description (TD) documents
@0xc2181f7117220fb9;

using Go = import "/go.capnp";
$Go.package("svc");
$Go.import("github.com/hiveot/hub.capnp/go/svc");


struct TD {
    # TD containing a Thing Description document
    # This retains the TD itself in JSON encoding as per W3C specification. If a TD is needed
    # for service operations that might change.
    
    id @0 :Text;
    # ID of the TD
    
    tdJson @1 :Text;
    # The JSON encoded TD
}

