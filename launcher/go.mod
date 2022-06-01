module github.com/wostzone/hub/launcher

go 1.18

require (
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.1
	github.com/wostzone/hub/certs v0.0.0-20220601182859-0028a8d31b61
	github.com/wostzone/wost-go v0.0.0-20220601182858-860c54605a83
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/sys v0.0.0-20220520151302-bc2c85ada10a // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/wostzone/hub/certs => ../certs

replace github.com/wostzone/wost-go => ../../wost-go
