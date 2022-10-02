package capnp4POGS

import (
	"capnproto.org/go/capnp/v3"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/provisioning"
)

// OobSecretsPOGS2Capnp converts a list of OOB secrets from POGS to Capnp
func OobSecretsPOGS2Capnp(secrets []provisioning.OOBSecret) hubapi.OOBSecret_List {

	_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))
	secretListCapnp, _ := hubapi.NewOOBSecret_List(seg, int32(len(secrets)))
	for i, secret := range secrets {
		_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))
		secretCapnp, _ := hubapi.NewOOBSecret(seg)
		_ = secretCapnp.SetDeviceID(secret.DeviceID)
		_ = secretCapnp.SetOobSecret(secret.OobSecret)
		_ = secretListCapnp.Set(i, secretCapnp)
	}

	return secretListCapnp
}

// OobSecretsCapnp2POGS converts a list of OOB secrets from Capnp to POGS
func OobSecretsCapnp2POGS(secretListCapnp hubapi.OOBSecret_List) []provisioning.OOBSecret {
	secretListPOGS := make([]provisioning.OOBSecret, secretListCapnp.Len())
	for i := 0; i < secretListCapnp.Len(); i++ {
		secretCapnp := secretListCapnp.At(i)
		deviceID, _ := secretCapnp.DeviceID()
		oobSecret, _ := secretCapnp.OobSecret()
		secret := provisioning.OOBSecret{
			DeviceID:  deviceID,
			OobSecret: oobSecret,
		}
		secretListPOGS[i] = secret
	}

	return secretListPOGS
}
