module github.com/wostzone/hub/logger

go 1.14

require (
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/wostzone/hub/lib/client v0.0.0-20211107044856-0202be1adf0c
)

replace github.com/wostzone/hub/lib/client => ../lib/client
