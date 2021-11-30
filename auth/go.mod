module github.com/wostzone/hub/auth

go 1.16

require (
	github.com/alexedwards/argon2id v0.0.0-20210511081203-7d35d68092b8
	github.com/docopt/docopt-go v0.0.0-20180111231733-ee0de3bc6815
	github.com/fsnotify/fsnotify v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/rs/cors v1.8.0 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/wostzone/hub/lib/client v0.0.0-20211107002258-e0347d7dc2cd
	github.com/wostzone/hub/lib/serve v0.0.0-20211107033214-4ab42d646d60
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

replace github.com/wostzone/hub/lib/client => ../lib/client

replace github.com/wostzone/hub/lib/serve => ../lib/serve
