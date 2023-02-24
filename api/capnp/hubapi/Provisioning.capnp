# Cap'n proto definition for provisioning service
@0x9579ece206ee504b;

using Go = import "/go.capnp";
$Go.package("hubapi");
$Go.import("github.com/hiveot/hub/api/go/hubapi");

const provisioningServiceName :Text = "provisioning";


struct OOBSecret {
    # OOBSecret holds a device's Out Of Band secret for automated provisioning
    # If the deviceID and MD5 hash of the secret match with the request it will be approved immediately
    
    deviceID @0 :Text;
    # The unique device ID or MAC address

    oobSecret @1 :Text;
    # The OOB secret of the device, 
}

struct ProvisionStatus {
    # Struct holding the status of a provisioning request

    deviceID @0 :Text;
    # deviceID of the request

    caCertPEM @1 :Text;
    # CA's certificate for future secure TLS connections, in PEM format.

    clientCertPEM @2 :Text;
    # The issued client certificate, if approved, in PEM format.
    # This is empty if the request is pending.

    pending @3 :Bool;
    # The request is pending approval, wait retrySec seconds before retrying.

    pubKeyPEM @4 :Text;
    # pubKeyPEM is the public key of the IoT device, used to generate the client certificate.

    requestTime @6 :Text;
    # ISO8601 timestamp when the request was received. Used to expire requests.

    retrySec @5 :Int32  = 3600;
    # Recommended delay before retrying if the request is pending
}

const capNameManageProvisioning :Text = "capManageProvisioning";
const capNameRequestProvisioning :Text = "capRequestProvisioning";
const capNameRefreshProvisioning :Text = "capRefreshProvisioning";

interface CapProvisioning {
# Capabilities for provisioning of IoT devices

    capManageProvisioning @0 (clientID :Text) -> (cap :CapManageProvisioning);
    # getManagementCapability provides the capability to manage provisioning requests

    capRequestProvisioning @1 (clientID :Text) ->(cap :CapRequestProvisioning);
    # getRequestCapability provides the capability to provision IoT devices

    capRefreshProvisioning @2 (clientID :Text) ->(cap :CapRefreshProvisioning);
    # getRequestCapability provides the capability to provision IoT devices
    # The request must be made with a valid certificate and is only valid for a matching deviceID.
}

interface CapManageProvisioning {
# Capability to manage provisioning requests and OOB secrets

    addOOBSecrets @0 (oobSecrets:List(OOBSecret)) -> ();
    # Add a list of OOB secrets for automated pre-approved provisioning

    approveRequest @1 (deviceID:Text) -> ();
    # Approve a pending request for the given device ID

    getApprovedRequests @2 () -> (requests :List(ProvisionStatus));
    # GetApprovedRequests returns a list of provisioned devices

    getPendingRequests @3 () -> (requests :List(ProvisionStatus));
    # GetPendingRequests returns a list of pending provisioning requests
}

interface CapRequestProvisioning {
# Capability to request IoT device provisioning and receive a certificate

    submitProvisioningRequest @0 (deviceID :Text, md5Secret :Text, pubKeyPEM :Text) -> (status:ProvisionStatus);
    # IoT device submits a provisioning request with the MD5 hash of the out-of-band secret
    # If the deviceID and MD5 hash of the secret match with previously uploaded secrets then the
    # request will be approved immediately, otherwise a pending request will be created.

}

interface CapRefreshProvisioning {
# Capability to refresh an existing IoT device certificate
# This is only available to IoT devices with an existing valid certificate.

    refreshDeviceCert @0 (certPEM:Text) -> (status :ProvisionStatus);
    # Refresh the device certificate and return a certificate with a new expiry date
	# If the certificate is still valid this will succeed. If the certificate is expired
	# then the request must be approved by the administrator.
}
