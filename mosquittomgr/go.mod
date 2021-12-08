module github.com/wostzone/hub/mosquittomgr

go 1.16

require (
	github.com/alexedwards/argon2id v0.0.0-20210511081203-7d35d68092b8
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/wostzone/hub/auth v0.0.0-20211107034644-2b0b1bf17e2c
	github.com/wostzone/hub/lib/client v0.0.0-20211107002258-e0347d7dc2cd
	github.com/wostzone/hub/lib/serve v0.0.0-20211107033214-4ab42d646d60
)

replace github.com/wostzone/hub/lib/serve => ../lib/serve

replace github.com/wostzone/hub/lib/client => ../lib/client

replace github.com/wostzone/hub/auth => ../auth