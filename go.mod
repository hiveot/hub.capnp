module github.com/hiveot/hub

go 1.18

require (
	capnproto.org/go/capnp/v3 v3.0.0-alpha.5
	github.com/hiveot/hub.capnp v0.0.0-00010101000000-000000000000
	github.com/hiveot/hub.go v0.0.0-20220604012454-a45ed192e850
	github.com/ohler55/ojg v1.14.4
	github.com/sirupsen/logrus v1.9.0
	github.com/stretchr/testify v1.8.0
	github.com/struCoder/pidusage v0.2.1
	github.com/urfave/cli/v2 v2.11.2
	go.mongodb.org/mongo-driver v1.10.1
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/araddon/dateparse v0.0.0-20210429162001-6b43995a97de // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/montanaflynn/stats v0.6.6 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.1 // indirect
	github.com/xdg-go/stringprep v1.0.3 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	github.com/youmark/pkcs8 v0.0.0-20201027041543-1326539a0a0a // indirect
	golang.org/x/crypto v0.0.0-20220826181053-bd7e27e6170d // indirect
	golang.org/x/sync v0.0.0-20220819030929-7fc1605a5dde // indirect
	golang.org/x/sys v0.0.0-20220825204002-c680a09ffe64 // indirect
	golang.org/x/text v0.3.7 // indirect
)

replace github.com/hiveot/hub.go => ../hub.go

replace github.com/hiveot/hub.capnp => ../hub.capnp
