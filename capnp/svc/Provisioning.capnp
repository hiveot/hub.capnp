# Cap'n proto definition for provisioning service
@0x9579ece206ee504b;

using Go = import "/go.capnp";
$Go.package("svc");
$Go.import("github.com/hiveot/hub.capnp/go/svc");

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

    validUntil @4 :Text;
    # ISO8601 timestamp until which the provisioning request is valid  

    timestamp @5 :Text; 
    # ISO8601 timestamp when the request was received. This is set by the service.
}


struct ProvisionResponse {
    # Struct holding the response to a provisioning request

    pending @0 :Bool;
    # the request is pending approval, wait retrySec to retry

    clientCertPEM @1 :Text;
    # The issued client certificate if approved, in PEM format

    caCertPEM @2 :Text;
    # CA's certificate used to sign the request, in PEM format

    retrySec @3 :Int32  = 3600;
    # Optional delay for retrying the request in seconds in case status is pending
}

interface ProvisioningService {
    # Provisioning service for issuing certificates to IoT devices and services

    addOOBSecret @0 (oobSecrets:List(OOBSecret)) -> ();
    # Add a list of OOB secrets for automated provisioning

    approveRequest @1 (deviceID:Text) -> ();
    # Approve a pending request for the given device ID

    getPendingRequests @2 () -> (requests :List(ProvisionRequest));
    # GetPendingRequests returns a list of pending requests

    refreshProvisioning @3 (deviceID:Text, pubKeyPEM:Text) -> (provResp :ProvisionResponse);
    # Refresh the provisioning and return a new certificate.
    # This will only succeed if the request is made with a valid certificate.

    submitProvisioningRequest @4 (provRequest:ProvisionRequest) -> (provResp:ProvisionResponse);
    # IoT device submits a provisioning request with the MD5 hash of the out-of-band secret 
    # If the deviceID and MD5 hash of the secret match with previously uploaded secrets then the 
    # request will be approved immediately.
}

