package service

type ICertService interface {
	CreateClientCert(clientID string, pubKeyPEM string) (
		certPEM string, caCertPEM string, err error)

	CreateDeviceCert(clientID string, pubKeyPEM string) (
		certPEM string, caCertPEM string, err error)

	CreateServiceCert(serviceID string, pubKeyPEM string, names []string) (
		certPEM string, caCertPEM string, err error)
}
