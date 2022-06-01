module github.com/wostzone/hub/authz

go 1.18

require (
	github.com/docopt/docopt-go v0.0.0-20180111231733-ee0de3bc6815
	github.com/fsnotify/fsnotify v1.5.4
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.1
	github.com/wostzone/wost-go v0.0.0-20220526055823-29600e2bc990
	gopkg.in/yaml.v3 v3.0.1
)

replace github.com/wostzone/wost-go => ../../wost-go
