module github.com/wostzone/hub/launcher

go 1.14

require (
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/wostzone/hub/certs v0.0.0-20220302002533-d6b81ae191e8
	github.com/wostzone/hub/lib/client v0.0.0-20211107002258-e0347d7dc2cd
)

// replace github.com/wostzone/hub/launcher => ./
replace github.com/wostzone/hub/lib/serve => ../lib/serve

replace github.com/wostzone/hub/lib/client => ../lib/client
