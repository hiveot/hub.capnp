module github.com/hiveot/hub

go 1.19

require (
	capnproto.org/go/capnp/v3 v3.0.0-alpha.24
	github.com/alexedwards/argon2id v0.0.0-20211130144151-3585854a6387
	github.com/araddon/dateparse v0.0.0-20210429162001-6b43995a97de
	github.com/cockroachdb/pebble v0.0.0-20230201230940-fcc9067e90e2
	github.com/fsnotify/fsnotify v1.6.0
	github.com/gobwas/ws v1.1.0
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/google/uuid v1.3.0
	github.com/hiveot/hub.capnp v0.1.0-alpha
	github.com/samber/lo v1.37.0
	github.com/sirupsen/logrus v1.9.0
	github.com/stretchr/testify v1.8.1
	github.com/struCoder/pidusage v0.2.1
	github.com/thanhpk/randstr v1.0.4
	github.com/tidwall/btree v1.6.0
	github.com/urfave/cli/v2 v2.24.2
	go.etcd.io/bbolt v1.3.7
	go.mongodb.org/mongo-driver v1.11.1
	go.uber.org/zap v1.24.0
	golang.org/x/crypto v0.5.0
	golang.org/x/sys v0.4.0
	gopkg.in/fsnotify.v1 v1.4.7
	gopkg.in/square/go-jose.v2 v2.6.0
	gopkg.in/yaml.v3 v3.0.1
	zenhack.net/go/websocket-capnp v0.0.0-20230122013820-cb32f4dfbb0b
)

require (
	github.com/DataDog/zstd v1.5.2 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/cockroachdb/errors v1.9.1 // indirect
	github.com/cockroachdb/logtags v0.0.0-20230118201751-21c54148d20b // indirect
	github.com/cockroachdb/redact v1.1.3 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/getsentry/sentry-go v0.17.0 // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/klauspost/compress v1.15.15 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/montanaflynn/stats v0.7.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.14.0 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.39.0 // indirect
	github.com/prometheus/procfs v0.9.0 // indirect
	github.com/rogpeppe/go-internal v1.9.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	github.com/youmark/pkcs8 v0.0.0-20201027041543-1326539a0a0a // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/multierr v1.9.0 // indirect
	golang.org/x/exp v0.0.0-20230131160201-f062dba9d201 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/text v0.6.0 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
)

replace github.com/hiveot/hub.capnp => ../hub.capnp

replace capnproto.org/go/capnp/v3 => ../../go-capnproto2
