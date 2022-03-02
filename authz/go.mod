module github.com/wostzone/hub/authz

go 1.16

require (
	github.com/docopt/docopt-go v0.0.0-20180111231733-ee0de3bc6815
	github.com/fsnotify/fsnotify v1.5.1
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/wostzone/hub/lib/client v0.0.0-20211107002258-e0347d7dc2cd
	github.com/wostzone/hub/lib/serve v0.0.0-20211107033214-4ab42d646d60
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

replace github.com/wostzone/hub/lib/client => ../lib/client

replace github.com/wostzone/hub/lib/serve => ../lib/serve
