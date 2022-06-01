module github.com/wostzone/hub/authn

go 1.18

require (
	github.com/alexedwards/argon2id v0.0.0-20211130144151-3585854a6387
	github.com/docopt/docopt-go v0.0.0-20180111231733-ee0de3bc6815
	github.com/fsnotify/fsnotify v1.5.4
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/gorilla/mux v1.8.0
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.1
	github.com/wostzone/wost-go v0.0.0-00010101000000-000000000000
	golang.org/x/crypto v0.0.0-20220525230936-793ad666bf5e
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rs/cors v1.8.2 // indirect
	golang.org/x/net v0.0.0-20220526153639-5463443f8c37 // indirect
	golang.org/x/sys v0.0.0-20220520151302-bc2c85ada10a // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/wostzone/wost-go => ../../wost-go
