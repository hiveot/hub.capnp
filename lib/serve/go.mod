module github.com/wostzone/hub/lib/serve

go 1.14

require (
	github.com/fsnotify/fsnotify v1.5.1
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/gorilla/mux v1.8.0
	github.com/grandcat/zeroconf v1.0.0
	github.com/kr/pretty v0.3.0 // indirect
	github.com/rogpeppe/go-internal v1.8.0 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/wostzone/hub/lib/client v0.0.0-20211107002258-e0347d7dc2cd
	github.com/wostzone/hubclient-go v0.0.0-20211108021100-1227a372e631 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)

replace github.com/wostzone/hub/lib/client => ../client
