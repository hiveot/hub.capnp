module github.com/wostzone/hub/mosquittomgr

go 1.16

require (
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.1
	github.com/wostzone/hub/authn v0.0.0-00010101000000-000000000000
	github.com/wostzone/hub/authz v0.0.0-00010101000000-000000000000
	github.com/wostzone/wost-go v0.0.0-20220526055823-29600e2bc990
)

replace github.com/wostzone/hub/authn => ../authn

replace github.com/wostzone/hub/authz => ../authz

replace github.com/wostzone/wost-go => ../../wost-go
