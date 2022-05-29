module github.com/wostzone/hub/authn

go 1.16

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

replace github.com/wostzone/wost-go => ../../wost-go
