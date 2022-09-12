package certcli_test

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"path"
	"testing"
	"time"

	"capnproto.org/go/capnp/v3/rpc"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	svc2 "github.com/hiveot/hub.capnp/go/capnp/svc"
	pb "github.com/hiveot/hub.grpc/go/svc"
	"github.com/hiveot/hub/cmd/hubcli/certcli"

	"github.com/hiveot/hub.go/pkg/certsclient"
	"github.com/hiveot/hub/pkg/svc/certsvc/selfsigned"
	"github.com/hiveot/hub/pkg/svc/certsvc/service"
)

func TestGetCommands(t *testing.T) {
	cmd := certcli.GetCertCommands("./")
	assert.NotNil(t, cmd)
}

func TestCreateCA(t *testing.T) {
	tempFolder := path.Join(os.TempDir(), "certcli-test")
	force := true
	sanName := "test"
	//_ = os.MkdirAll(certsFolder, 0700)
	//_ = os.Chdir(tempFolder)

	err := certcli.HandleCreateCACert(tempFolder, sanName, 0, force)
	assert.NoError(t, err)

	certPath := path.Join(tempFolder, service.DefaultCaCertFile)
	assert.FileExists(t, certPath)

	_ = os.RemoveAll(tempFolder)
}

func TestCreateCA_ErrorExists(t *testing.T) {
	tempFolder := path.Join(os.TempDir(), "certcli-test")
	// create the cert
	force := true
	err := certcli.HandleCreateCACert(tempFolder, "test", 0, force)
	assert.NoError(t, err)

	// error cert exists
	force = false
	err = certcli.HandleCreateCACert(tempFolder, "test", 0, force)
	assert.Error(t, err)

	// error key exists
	os.Remove(path.Join(tempFolder, service.DefaultCaCertFile))
	force = false
	err = certcli.HandleCreateCACert(tempFolder, "test", 0, force)
	assert.Error(t, err)

	_ = os.RemoveAll(tempFolder)
}

func TestCreateCA_FolderDoesntExists(t *testing.T) {
	tempFolder := path.Join(os.TempDir(), "certcli-test")
	_ = os.RemoveAll(tempFolder)

	force := false
	err := certcli.HandleCreateCACert(tempFolder, "test", 0, force)
	assert.Error(t, err)
}

func TestCreateClientCert(t *testing.T) {
	clientID := "client"
	keyFile := ""
	const count = 10000

	tempFolder := path.Join(os.TempDir(), "certcli-test")
	err := certcli.HandleCreateCACert(tempFolder, "test", 0, true)
	assert.NoError(t, err)

	// create the cert
	t1 := time.Now()
	for i := 0; i < count; i++ {
		err = certcli.HandleCreateClientCert(tempFolder, clientID, keyFile, 0)
		assert.NoError(t, err)
	}
	d1 := time.Since(t1)
	fmt.Printf("HandleCreateClientCert duration of %d calls: %d msec\n", count, d1.Milliseconds())

	// direct call without cli to measure perfrmance
	caCert, caKey, err := selfsigned.CreateHubCA(1)
	assert.NoError(t, err)
	privKey := certsclient.CreateECDSAKeys()

	t2 := time.Now()
	for i := 0; i < count; i++ {
		cert, err := selfsigned.CreateClientCert(
			clientID, certsclient.OUClient, &privKey.PublicKey, caCert, caKey, service.DefaultClientCertDurationDays)
		assert.NoError(t, err)
		assert.NotNil(t, cert)
	}
	d2 := time.Since(t2)
	fmt.Printf("CreateClientCert duration of %d calls: %d msec\n", count, d2.Milliseconds())

	// missing key file
	// keyFile = "missingkeyfile.pem"
	// err = certcli.HandleCreateClientCert(tempFolder, clientID, keyFile, 0)
	// assert.Error(t, err)

	_ = os.RemoveAll(tempFolder)
}

// create a cert using a running gRPC service
func TestCreateClientCertGRPC(t *testing.T) {
	clientID := "client"
	const count = 10000

	address := "unix:///tmp/certsvc.socket"
	// address = "localhost:8881"

	fmt.Println("Connecting to: ", address)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()
	conn, _ := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	certClient := pb.NewCertServiceClient(conn)

	privKey := certsclient.CreateECDSAKeys()
	pubKeyPEM, err := certsclient.PublicKeyToPEM(&privKey.PublicKey)
	if err != nil {
		log.Fatal(err)
	}

	// create the cert
	t1 := time.Now()
	for i := 0; i < count; i++ {

		args := &pb.CreateClientCert_Args{ClientID: clientID, PubKeyPEM: pubKeyPEM}
		resp, err := certClient.CreateClientCert(ctx, args)
		_ = resp
		assert.NoError(t, err)
	}
	d1 := time.Since(t1)
	fmt.Printf("HandleCreateClientCert duration of %d calls: %d msec\n", count, d1.Milliseconds())

}

// create a cert using a running capnp service
func TestCreateClientCertCapnp(t *testing.T) {
	clientID := "client"
	const count = 10000

	// test creating the cert
	network := "unix"
	address := "/tmp/certsvc.socket"
	clientSideConn, err := net.Dial(network, address)
	transport := rpc.NewStreamTransport(clientSideConn)
	clientConn := rpc.NewConn(transport, nil)
	defer clientConn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	certClient := svc2.CertServiceCap(clientConn.Bootstrap(ctx))

	privKey := certsclient.CreateECDSAKeys()
	if err != nil {
		log.Fatal(err)
	}
	pubKeyPEM, err := certsclient.PublicKeyToPEM(&privKey.PublicKey)
	if err != nil {
		log.Fatal(err)
	}

	t1 := time.Now()
	for i := 0; i < count; i++ {
		// create the client capability (eg message)
		resp, release := certClient.CreateClientCert(ctx,
			func(params svc2.CertServiceCap_createClientCert_Params) error {
				fmt.Println("CertServiceCap_createClientCert_Params")
				err = params.SetClientID(clientID)
				err = params.SetPubKeyPEM(pubKeyPEM)
				return err
			})
		defer release()

		result, err := resp.Struct()
		if err != nil {
			log.Fatalf("error getting response struct: %v", err)
		}
		cp, err := result.CertPEM()
		assert.NoError(t, err)
		assert.NotNil(t, cp)
		cap, err := result.CaCertPEM()
		assert.NoError(t, err)
		assert.NotNil(t, cap)
	}
	d1 := time.Since(t1)
	fmt.Printf("duration of %d calls: %d msec\n", count, d1.Milliseconds())

}

func TestCreateDeviceCert(t *testing.T) {
	deviceID := "urn:publisher:device1"
	keyFile := ""
	tempFolder := path.Join(os.TempDir(), "certcli-test")
	err := certcli.HandleCreateCACert(tempFolder, "test", 0, true)
	assert.NoError(t, err)

	err = certcli.HandleCreateDeviceCert(tempFolder, deviceID, keyFile, 0)
	assert.NoError(t, err)

	// missing key file
	keyFile = "missingkeyfile.pem"
	err = certcli.HandleCreateDeviceCert(tempFolder, deviceID, keyFile, 0)
	assert.Error(t, err)

	_ = os.RemoveAll(tempFolder)
}

func TestCreateServiceCert(t *testing.T) {
	serviceID := "service25"
	keyFile := ""
	tempFolder := path.Join(os.TempDir(), "certcli-test")
	err := certcli.HandleCreateCACert(tempFolder, "test", 0, true)
	assert.NoError(t, err)

	err = certcli.HandleCreateServiceCert(tempFolder, serviceID, "127.0.0.1", keyFile, 0)
	assert.NoError(t, err)

	_ = os.RemoveAll(tempFolder)
}

func TestCreateServiceCertWithKey(t *testing.T) {
	serviceID := "service25"
	tempFolder := path.Join(os.TempDir(), "certcli-test")
	keyFile := path.Join(tempFolder, serviceID+".pem")
	err := certcli.HandleCreateCACert(tempFolder, "test", 0, true)
	assert.NoError(t, err)

	privKey := certsclient.CreateECDSAKeys()
	err = certsclient.SaveKeysToPEM(privKey, keyFile)
	assert.NoError(t, err)
	// use a valid key
	err = certcli.HandleCreateServiceCert(tempFolder, serviceID, "", keyFile, 0)
	assert.NoError(t, err)

	// missing key file
	keyFile2 := path.Join(tempFolder, "keydoesntexist.pem")
	err = certcli.HandleCreateServiceCert(tempFolder, serviceID, "", keyFile2, 0)
	assert.Error(t, err)

	_ = os.RemoveAll(tempFolder)
}

func TestCreateServiceCertMissingCA(t *testing.T) {
	serviceID := "service25"
	keyFile := ""
	tempFolder := path.Join(os.TempDir(), "certcli-test")
	_ = os.RemoveAll(tempFolder)

	err := certcli.HandleCreateServiceCert(tempFolder, serviceID, "", keyFile, 1)
	assert.Error(t, err)

	_ = os.RemoveAll(tempFolder)
}

func TestShowCertInfo(t *testing.T) {
	tempFolder := path.Join(os.TempDir(), "certcli-test")
	err := certcli.HandleCreateCACert(tempFolder, "test", 0, true)
	assert.NoError(t, err)
	certFile := path.Join(tempFolder, service.DefaultCaCertFile)

	err = certcli.HandleShowCertInfo(certFile)
	assert.NoError(t, err)

	_ = os.RemoveAll(tempFolder)
}
