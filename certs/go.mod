module github.com/wostzone/hub/certs

go 1.16

replace github.com/wostzone/hub/lib/client => ../lib/client

replace github.com/wostzone/hub/lib/serve => ../lib/serve

require (
	github.com/docopt/docopt-go v0.0.0-20180111231733-ee0de3bc6815
	github.com/sirupsen/logrus v1.8.1
	github.com/wostzone/hub/lib/client v0.0.0-20211107002258-e0347d7dc2cd
	github.com/wostzone/hub/lib/serve v0.0.0-20211107033214-4ab42d646d60
)
