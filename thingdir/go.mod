module github.com/wostzone/hub/thingdir

go 1.14

require (
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/grandcat/zeroconf v1.0.0
	github.com/imdario/mergo v0.3.12
	github.com/nakabonne/tstorage v0.3.5
	github.com/ohler55/ojg v1.12.9
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/wostzone/hub/authn v0.0.0-00010101000000-000000000000
	github.com/wostzone/hub/authz v0.0.0-00010101000000-000000000000
	github.com/wostzone/hub/lib/client v0.0.0-20211108021100-1227a372e631
	github.com/wostzone/hub/lib/serve v0.0.0-20211107033214-4ab42d646d60
)

replace github.com/wostzone/hub/authn => ../authn

replace github.com/wostzone/hub/authz => ../authz

replace github.com/wostzone/hub/lib/serve => ../lib/serve

replace github.com/wostzone/hub/lib/client => ../lib/client
