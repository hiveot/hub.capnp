module svc

go 1.18

require (
	github.com/dapr/go-sdk v1.5.0
	github.com/ohler55/ojg v1.14.3
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.4
	github.com/wostzone/wost-go v0.0.0-20220604012454-a45ed192e850
	github.com/wostzone/wost.grpc/go v0.0.0-00010101000000-000000000000
	google.golang.org/protobuf v1.28.0
)

require (
	github.com/dapr/dapr v1.8.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/net v0.0.0-20220621193019-9d032be2e588 // indirect
	golang.org/x/sys v0.0.0-20220520151302-bc2c85ada10a // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20220622171453-ea41d75dfa0f // indirect
	google.golang.org/grpc v1.48.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/wostzone/wost-go => ../../wost-go

replace github.com/wostzone/wost.grpc/go => ../../wost.grpc/go
