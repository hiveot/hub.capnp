module github.com/wostzone/hub/lib/client

go 1.16

require (
	github.com/eclipse/paho.mqtt.golang v1.3.5
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/grandcat/zeroconf v1.0.0
	github.com/miekg/dns v1.1.43 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519 // indirect
	golang.org/x/net v0.0.0-20211105192438-b53810dc28af
	golang.org/x/sys v0.0.0-20211106132015-ebca88c72f68 // indirect
	gopkg.in/square/go-jose.v2 v2.6.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

replace github.com/wostzone/hub/lib/client/pkg/mqttclient => ./pkg/mqttclient
