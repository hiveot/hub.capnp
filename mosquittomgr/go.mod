module github.com/wostzone/hub/mosquittomgr

go 1.18

require (
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.1
	github.com/wostzone/hub/authn v0.0.0-20220601182859-0028a8d31b61
	github.com/wostzone/hub/authz v0.0.0-20220601182859-0028a8d31b61
	github.com/wostzone/wost-go v0.0.0-20220601182858-860c54605a83
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/eclipse/paho.mqtt.golang v1.3.5 // indirect
	github.com/fsnotify/fsnotify v1.5.4 // indirect
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rs/cors v1.8.2 // indirect
	golang.org/x/net v0.0.0-20220531201128-c960675eff93 // indirect
	golang.org/x/sys v0.0.0-20220520151302-bc2c85ada10a // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/wostzone/hub/authn => ../authn

replace github.com/wostzone/hub/authz => ../authz

replace github.com/wostzone/wost-go => ../../wost-go
