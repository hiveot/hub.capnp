module github.com/wostzone/hub

go 1.14

require (
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/wostzone/hubclient-go v0.0.0-00010101000000-000000000000
	github.com/wostzone/hubserve-go v0.0.0-20210907050346-343a1e9f8ad6
	golang.org/x/sys v0.0.0-20210510120138-977fb7262007 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

replace github.com/wostzone/hubclient-go => ../hubclient-go
