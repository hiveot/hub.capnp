module github.com/wostzone/hub/thingdir

go 1.14

require (
	github.com/gorilla/mux v1.8.0
	github.com/grandcat/zeroconf v1.0.0
	github.com/imdario/mergo v0.3.12
	github.com/ohler55/ojg v1.12.9
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.1
	github.com/wostzone/hub/authn v0.0.0-00010101000000-000000000000
	github.com/wostzone/hub/authz v0.0.0-00010101000000-000000000000
	github.com/wostzone/hub/lib/client v0.0.0-20220515042304-a67a4a917e3b
	github.com/wostzone/wost-go v0.0.0-20220526055823-29600e2bc990
)

replace github.com/wostzone/hub/authn => ../authn

replace github.com/wostzone/hub/authz => ../authz

replace github.com/wostzone/wost-go => ../../wost-go
