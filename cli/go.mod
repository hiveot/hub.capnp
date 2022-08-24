module github.com/wostzone/hub/launcher

go 1.17

require (
	github.com/go-cmd/cmd v1.4.1
	github.com/sirupsen/logrus v1.9.0
	github.com/stretchr/testify v1.8.0
	github.com/urfave/cli/v2 v2.11.2
	github.com/wostzone/hub/svc v0.0.0-00010101000000-000000000000
	github.com/wostzone/wost-go v0.0.0-20220604012454-a45ed192e850
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/wostzone/wost.grpc/go v0.0.0-00010101000000-000000000000 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	golang.org/x/crypto v0.0.0-20220817201139-bc19a97f63c8 // indirect
	golang.org/x/net v0.0.0-20220822230855-b0a4917ee28c // indirect
	golang.org/x/sys v0.0.0-20220823224334-20c2bfdbfe24 // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20220822174746-9e6da59bd2fc // indirect
	google.golang.org/grpc v1.49.0 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
	gopkg.in/square/go-jose.v2 v2.6.0 // indirect
)

replace github.com/wostzone/hub/svc => ../svc

replace github.com/wostzone/wost-go => ../../wost-go

replace github.com/wostzone/wost.grpc/go => ../../wost.grpc/go
