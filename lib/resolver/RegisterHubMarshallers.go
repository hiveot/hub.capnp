package resolver

import (
	"github.com/hiveot/hub/pkg/authn"
	"github.com/hiveot/hub/pkg/authn/capnpclient"
	"github.com/hiveot/hub/pkg/authz"
	capnpclient2 "github.com/hiveot/hub/pkg/authz/capnpclient"
	"github.com/hiveot/hub/pkg/certs"
	capnpclient3 "github.com/hiveot/hub/pkg/certs/capnpclient"
	"github.com/hiveot/hub/pkg/directory"
	capnpclient4 "github.com/hiveot/hub/pkg/directory/capnpclient"
	"github.com/hiveot/hub/pkg/gateway"
	capnpclient5 "github.com/hiveot/hub/pkg/gateway/capnpclient"
	"github.com/hiveot/hub/pkg/history"
	capnpclient6 "github.com/hiveot/hub/pkg/history/capnpclient"
	"github.com/hiveot/hub/pkg/provisioning"
	capnpclient7 "github.com/hiveot/hub/pkg/provisioning/capnpclient"
	"github.com/hiveot/hub/pkg/pubsub"
	capnpclient8 "github.com/hiveot/hub/pkg/pubsub/capnpclient"
	"github.com/hiveot/hub/pkg/state"
	capnpclient9 "github.com/hiveot/hub/pkg/state/capnpclient"
)

// Simple helper to register all Hub included marshallers

func RegisterHubMarshallers() {
	// authn
	RegisterCapnpMarshaller[authn.IAuthnService](capnpclient.NewAuthnCapnpClient, "")
	RegisterCapnpMarshaller[authn.IUserAuthn](capnpclient.NewUserAuthnCapnpClient, "")
	RegisterCapnpMarshaller[authn.IManageAuthn](capnpclient.NewManageAuthnCapnpClient, "")
	// authz
	RegisterCapnpMarshaller[authz.IAuthz](capnpclient2.NewAuthzCapnpClient, "")
	RegisterCapnpMarshaller[authz.IClientAuthz](capnpclient2.NewClientAuthzCapnpClient, "")
	RegisterCapnpMarshaller[authz.IManageAuthz](capnpclient2.NewManageAuthzCapnpClient, "")
	RegisterCapnpMarshaller[authz.IVerifyAuthz](capnpclient2.NewVerifyAuthzCapnpClient, "")
	// certs
	RegisterCapnpMarshaller[certs.ICerts](capnpclient3.NewCertsCapnpClient, "")
	RegisterCapnpMarshaller[certs.IDeviceCerts](capnpclient3.NewDeviceCertsCapnpClient, "")
	RegisterCapnpMarshaller[certs.IServiceCerts](capnpclient3.NewServiceCertsCapnpClient, "")
	RegisterCapnpMarshaller[certs.IUserCerts](capnpclient3.NewUserCertsCapnpClient, "")
	RegisterCapnpMarshaller[certs.IVerifyCerts](capnpclient3.NewVerifyCertsCapnpClient, "")
	// directory
	RegisterCapnpMarshaller[directory.IDirectory](capnpclient4.NewDirectoryCapnpClient, "")
	RegisterCapnpMarshaller[directory.IReadDirectory](capnpclient4.NewReadDirectoryCapnpClient, "")
	RegisterCapnpMarshaller[directory.IUpdateDirectory](capnpclient4.NewUpdateDirectoryCapnpClient, "")
	// gateway
	RegisterCapnpMarshaller[gateway.IGatewaySession](capnpclient5.NewGatewaySessionCapnpClient, "")
	// history
	RegisterCapnpMarshaller[history.IHistoryService](capnpclient6.NewHistoryCapnpClient, "")
	RegisterCapnpMarshaller[history.IAddHistory](capnpclient6.NewAddHistoryCapnpClient, "")
	RegisterCapnpMarshaller[history.IManageRetention](capnpclient6.NewManageRetentionCapnpClient, "")
	RegisterCapnpMarshaller[history.IReadHistory](capnpclient6.NewReadHistoryCapnpClient, "")
	// provisioning
	RegisterCapnpMarshaller[provisioning.IProvisioning](capnpclient7.NewProvisioningCapnpClient, "")
	RegisterCapnpMarshaller[provisioning.IManageProvisioning](capnpclient7.NewManageProvisioningCapnpClient, "")
	RegisterCapnpMarshaller[provisioning.IRefreshProvisioning](capnpclient7.NewRefreshProvisioningCapnpClient, "")
	RegisterCapnpMarshaller[provisioning.IRequestProvisioning](capnpclient7.NewRequestProvisioningCapnpClient, "")
	// pubsub
	RegisterCapnpMarshaller[pubsub.IPubSubService](capnpclient8.NewPubSubCapnpClient, "")
	RegisterCapnpMarshaller[pubsub.IDevicePubSub](capnpclient8.NewDevicePubSubCapnpClient, "")
	RegisterCapnpMarshaller[pubsub.IServicePubSub](capnpclient8.NewServicePubSubCapnpClient, "")
	RegisterCapnpMarshaller[pubsub.IUserPubSub](capnpclient8.NewUserPubSubCapnpClient, "")
	// state
	RegisterCapnpMarshaller[state.IStateService](capnpclient9.NewStateCapnpClient, "")
	RegisterCapnpMarshaller[state.IClientState](capnpclient9.NewClientStateCapnpClient, "")
}
