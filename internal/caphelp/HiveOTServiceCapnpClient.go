package caphelp

import (
	"context"

	"capnproto.org/go/capnp/v3"

	"github.com/hiveot/hub.capnp/go/hubapi"
)

// HiveOTServiceCapnpClient provide the POGS client connection to the capnp server of HiveOTService
type HiveOTServiceCapnpClient struct {
	capability hubapi.CapHiveOTService
}

// GetCapability obtains the capability with the given name.
// The caller must release the capability when done.
func (cl *HiveOTServiceCapnpClient) GetCapability(ctx context.Context,
	clientID string, clientType string, capabilityName string, args []string) (
	capabilityRef capnp.Client, err error) {

	method, release := cl.capability.GetCapability(ctx,
		func(params hubapi.CapHiveOTService_getCapability_Params) error {
			_ = params.SetClientID(clientID)
			_ = params.SetClientType(clientType)
			_ = params.SetCapabilityName(capabilityName)
			if args != nil {
				err = params.SetArgs(MarshalStringList(args))
			}
			return err
		})
	defer release()
	// return a future. Caller must release
	//capability = method.Cap().AddRef()

	// Just return the actual capability instead of a future, so the error is obtained if it isn't available.
	// Would be nice to return the future but this is an infrequent call anyways.
	resp, err := method.Struct()
	if err == nil {
		capability := resp.Cap().AddRef()
		capabilityRef = capability
	}
	return capabilityRef, err
}

// ListCapabilities lists the available capabilities of the service
// Returns a list of capabilities that can be obtained through the service
func (cl *HiveOTServiceCapnpClient) ListCapabilities(
	ctx context.Context, clientType string) (infoList []CapabilityInfo, err error) {

	infoList = make([]CapabilityInfo, 0)
	method, release := cl.capability.ListCapabilities(ctx,
		func(params hubapi.CapHiveOTService_listCapabilities_Params) error {
			err = params.SetClientType(clientType)
			return err
		})
	defer release()
	resp, err := method.Struct()
	if err == nil {
		infoListCapnp, err2 := resp.InfoList()
		if err = err2; err == nil {
			infoList = UnmarshalCapabilities(infoListCapnp)
		}
	}
	return infoList, err
}

// Stop the service and release its resources
func (cl *HiveOTServiceCapnpClient) Stop(ctx context.Context) error {
	cl.capability.Release()
	return nil
}

func NewHiveOTServiceCapnpClient(capability hubapi.CapHiveOTService) *HiveOTServiceCapnpClient {
	cl := &HiveOTServiceCapnpClient{capability: capability}
	return cl
}
