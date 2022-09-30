# Cap'n proto definition for provisioning service
@0x9579ece206ee504b;

using Go = import "/go.capnp";
$Go.package("hubapi");
$Go.import("github.com/hiveot/hub.capnp/go/hubapi");

struct OOBSecret {
    # OOBSecret holds a device's Out Of Band secret for automated provisioning
    # If the deviceID and MD5 hash of the secret match with the request it will be approved immediately
    
    deviceID @0 :Text;
    # The unique device ID or MAC address

    oobSecret @1 :Text;
    # The OOB secret of the device, 
}


struct ProvisionRequest {
    # ProvisionRequest holds the data of a provisioning request

    deviceID @0 :Text;
    # deviceID must be unique for the local network. 

    mac @1 :Text;
    # mac is the IoT device/service MAC address

    secretMD5 @2 :Text;
    # secretMD5 is the MD5 hash of the out-of-band secret if available. Use "" for manual approval

    pubKeyPEM @3 :Text;
    # pubKeyPEM is the public key of the IoT device, used to generate the client certificate.
}


struct ProvisionResponse {
    # Struct holding the response to a provisioning request

    deviceID @0 :Text;
    # deviceID of the request

    caCertPEM @1 :Text;
    # CA's certificate used to handle the request, in PEM format.

    clientCertPEM @2 :Text;
    # The issued client certificate, if approved, in PEM format.
    # This is empty if the request is pending.

    pending @3 :Bool;
    # The request is pending approval, wait the recommended retrySec before retrying.

    pubKeyPEM @4 :Text;
    # pubKeyPEM is the public key of the IoT device, used to generate the client certificate.

    requestTime @6 :Text;
    # ISO8601 timestamp when the request was received. This is set by the service.

    retrySec @5 :Int32  = 3600;
    # Recommended delay before retrying if the request is pending
}

interface CapProvisioning {
# Capabilities for provisioning of IoT devices

    capProvisionManagement @0 () -> (cap :CapProvisionManagement);
    # getManagementCapability provides the capability to manage provisioning requests

    capProvisionRequest @1 () ->(cap :CapProvisionRequest);
    # getRequestCapability provides the capability to provision IoT devices
}

interface CapProvisionManagement {
# Capability to manage provisioning requests and OOB secrets

    addOOBSecret @0 (oobSecrets:List(OOBSecret)) -> ();
    # Add a list of OOB secrets for automated pre-approved provisioning

    approveRequest @1 (deviceID:Text) -> ();
    # Approve a pending request for the given device ID

    getApprovedRequests @2 () -> (requests :List(ProvisionResponse));
    # GetApprovedRequests returns a list of provisioned devices

    getPendingRequests @3 () -> (requests :List(ProvisionResponse));
    # GetPendingRequests returns a list of pending provisioning requests
}

interface CapProvisionRequest {
# Capability to issue provisioning requests and certificate renewal

    submitProvisioningRequest @0 (provRequest:ProvisionRequest) -> (provResp:ProvisionResponse);
    # IoT device submits a provisioning request with the MD5 hash of the out-of-band secret
    # If the deviceID and MD5 hash of the secret match with previously uploaded secrets then the
    # request will be approved immediately.

    refreshDeviceCert @1 (deviceID:Text, pubKeyPEM:Text) -> (provResp :ProvisionResponse);
    # Refresh the device certificate and return a certificate with a new expiry date
    # This will only succeed if the request is made with a valid certificate.

}
