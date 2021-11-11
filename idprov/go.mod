module github.com/wostzone/hub/idprov

go 1.14

require (
	github.com/gorilla/mux v1.8.0
	github.com/grandcat/zeroconf v1.0.0
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/wostzone/hub/lib/client v0.0.0-20211107002258-e0347d7dc2cd
	github.com/wostzone/hub/lib/serve v0.0.0-20211107033214-4ab42d646d60
)

replace github.com/wostzone/hub/lib/serve => ../lib/serve

replace github.com/wostzone/hub/lib/client => ../lib/client
