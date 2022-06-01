module github.com/wostzone/hub/thingdir

go 1.18

require (
	github.com/gorilla/mux v1.8.0
	github.com/grandcat/zeroconf v1.0.0
	github.com/imdario/mergo v0.3.12
	github.com/ohler55/ojg v1.12.9
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.1
	github.com/wostzone/hub/authn v0.0.0-00010101000000-000000000000
	github.com/wostzone/hub/authz v0.0.0-00010101000000-000000000000
	github.com/wostzone/wost-go v0.0.0-20220526055823-29600e2bc990
)

require (
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/eclipse/paho.mqtt.golang v1.3.5 // indirect
	github.com/fsnotify/fsnotify v1.5.4 // indirect
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/miekg/dns v1.1.43 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rs/cors v1.8.2 // indirect
	golang.org/x/net v0.0.0-20220526153639-5463443f8c37 // indirect
	golang.org/x/sys v0.0.0-20220520151302-bc2c85ada10a // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/wostzone/hub/authn => ../authn

replace github.com/wostzone/hub/authz => ../authz

replace github.com/wostzone/wost-go => ../../wost-go
