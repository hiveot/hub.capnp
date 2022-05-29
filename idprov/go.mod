module github.com/wostzone/hub/idprov

go 1.14

require (
	github.com/gorilla/mux v1.8.0
	github.com/grandcat/zeroconf v1.0.0
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.1
	github.com/wostzone/hub/certs v0.0.0-00010101000000-000000000000
	github.com/wostzone/wost-go v0.0.0-00010101000000-000000000000
)

replace github.com/wostzone/hub/certs => ../certs

replace github.com/wostzone/wost-go => ../../wost-go
