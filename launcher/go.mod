module github.com/wostzone/hub/launcher

go 1.14

require (
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.1
	github.com/wostzone/hub/certs v0.0.0-00010101000000-000000000000
	github.com/wostzone/hub/lib/client v0.0.0-20220515042304-a67a4a917e3b
	github.com/wostzone/wost-go v0.0.0-00010101000000-000000000000
)

replace github.com/wostzone/hub/certs => ../certs

replace github.com/wostzone/wost-go => ../../wost-go
