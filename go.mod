module github.com/aspect-build/silo

go 1.20

require (
	github.com/Masterminds/semver/v3 v3.2.1
	github.com/aspect-build/talkie v0.0.0-00010101000000-000000000000
	github.com/bazel-contrib/rules_jvm v0.13.0
	github.com/bazelbuild/bazel-gazelle v0.31.1
	github.com/bazelbuild/bazelisk v1.17.0
	github.com/bazelbuild/buildtools v0.0.0-20230510134650-37bd1811516d
	github.com/bazelbuild/rules_go v0.40.0
	github.com/bmatcuk/doublestar/v4 v4.6.0
	github.com/buildbarn/bb-remote-execution v0.0.0-20230616145815-13b8a35f4aeb
	github.com/emirpasic/gods v1.18.1
	github.com/evanw/esbuild v0.18.10
	github.com/fatih/color v1.15.0
	github.com/go-test/deep v1.1.0
	github.com/golang/mock v1.6.0
	github.com/golang/protobuf v1.5.3
	github.com/google/triage-party v1.4.0
	github.com/hashicorp/go-hclog v1.5.0
	github.com/hashicorp/go-plugin v1.4.10
	github.com/manifoldco/promptui v0.9.0
	github.com/mattn/go-isatty v0.0.19
	github.com/mitchellh/go-homedir v1.1.0
	github.com/msolo/jsonr v0.0.0-20230325054138-b14a608f43e2
	github.com/onsi/ginkgo/v2 v2.11.0
	github.com/onsi/gomega v1.27.8
	github.com/pkg/browser v0.0.0-20210911075715-681adbf594b8
	github.com/pmezard/go-difflib v1.0.0
	github.com/rabbitmq/amqp091-go v1.8.1
	github.com/sabhiram/go-gitignore v0.0.0-20210923224102-525f6e181f06
	github.com/sirupsen/logrus v1.9.3
	github.com/slack-go/slack v0.12.2
	github.com/smacker/go-tree-sitter v0.0.0-20230501083651-a7d92773b3aa
	github.com/spf13/cobra v1.7.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.16.0
	github.com/uptrace/bun v1.1.14
	github.com/uptrace/bun/dialect/pgdialect v1.1.14
	github.com/uptrace/bun/dialect/sqlitedialect v1.1.14
	github.com/uptrace/bun/driver/pgdriver v1.1.14
	github.com/uptrace/bun/driver/sqliteshim v1.1.14
	golang.org/x/exp v0.0.0-20230626212559-97b1e661b5df
	golang.org/x/mod v0.11.0
	golang.org/x/sync v0.3.0
	google.golang.org/genproto v0.0.0-20230626202813-9b080da550b3
	google.golang.org/genproto/googleapis/api v0.0.0-20230626202813-9b080da550b3
	google.golang.org/grpc v1.56.1
	google.golang.org/protobuf v1.31.0
	gopkg.in/yaml.v3 v3.0.1
	sigs.k8s.io/yaml v1.3.0
)

require github.com/buildbarn/bb-storage v0.0.0-20230517082930-1eb2bad58bdb // indirect

require (
	cloud.google.com/go/compute v1.19.1 // indirect
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	cloud.google.com/go/longrunning v0.4.1 // indirect
	github.com/GoogleCloudPlatform/cloudsql-proxy v0.0.0-20200501161113-5e9e23d7cb91 // indirect
	github.com/bazelbuild/remote-apis v0.0.0-20230411132548-35aee1c4a425 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bgentry/go-netrc v0.0.0-20140422174119-9fd32a8b3d3d // indirect
	github.com/buildbarn/bb-clientd v0.0.0-20230414074313-677f2e45487b
	github.com/buildbarn/bb-clientd/internal/mock v0.0.0-00010101000000-000000000000 // indirect
	github.com/buildbarn/bb-storage/internal/mock v0.0.0-00010101000000-000000000000 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/chzyer/readline v1.5.1 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-sql-driver/mysql v1.6.0 // indirect
	github.com/go-task/slim-sprig v0.0.0-20230315185526-52ccab3ef572 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/google/btree v1.1.2 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/go-github/v33 v33.0.0 // indirect
	github.com/google/go-jsonnet v0.19.1 // indirect
	github.com/google/go-querystring v1.0.0 // indirect
	github.com/google/pprof v0.0.0-20221118152302-e6195bd50e26 // indirect
	github.com/google/s2a-go v0.1.3 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.2.3 // indirect
	github.com/googleapis/gax-go/v2 v2.8.0 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-retryablehttp v0.6.6 // indirect
	github.com/hashicorp/go-version v1.6.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hashicorp/yamux v0.1.1 // indirect
	github.com/hokaccha/go-prettyjson v0.0.0-20190818114111-108c894c2c0e // indirect
	github.com/imjasonmiller/godice v0.1.2 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jmoiron/sqlx v1.3.5 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-sqlite3 v1.14.16 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/mitchellh/go-testing-interface v1.14.1 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/oklog/run v1.1.0 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/pelletier/go-toml/v2 v2.0.8 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/prometheus/client_golang v1.15.1 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.42.0 // indirect
	github.com/prometheus/procfs v0.9.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/spf13/afero v1.9.5 // indirect
	github.com/spf13/cast v1.5.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/subosito/gotenv v1.4.2 // indirect
	github.com/tmthrgd/go-hex v0.0.0-20190904060850-447a3041c3bc // indirect
	github.com/vmihailenco/msgpack/v5 v5.3.5 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/xanzy/go-gitlab v0.36.0 // indirect
	go.opencensus.io v0.24.0 // indirect
	golang.org/x/crypto v0.9.0 // indirect
	golang.org/x/net v0.10.0 // indirect
	golang.org/x/oauth2 v0.8.0 // indirect
	golang.org/x/sys v0.9.0 // indirect
	golang.org/x/text v0.9.0 // indirect
	golang.org/x/time v0.1.0 // indirect
	golang.org/x/tools v0.9.3 // indirect
	google.golang.org/api v0.122.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230530153820-e85fd2cbaebc // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/klog/v2 v2.90.1 // indirect
	lukechampine.com/uint128 v1.3.0 // indirect
	mellium.im/sasl v0.3.1 // indirect
	modernc.org/cc/v3 v3.40.0 // indirect
	modernc.org/ccgo/v3 v3.16.13 // indirect
	modernc.org/libc v1.22.6 // indirect
	modernc.org/mathutil v1.5.0 // indirect
	modernc.org/memory v1.5.0 // indirect
	modernc.org/opt v0.1.3 // indirect
	modernc.org/sqlite v1.22.1 // indirect
	modernc.org/strutil v1.1.3 // indirect
	modernc.org/token v1.1.0 // indirect
)

replace github.com/aspect-build/talkie => ./talkie

replace github.com/buildbarn/bb-clientd/internal/mock => ./third_party/hack/github.com/buildbarn/bb-clientd/internal/mock

replace github.com/buildbarn/bb-storage/internal/mock => ./third_party/hack/github.com/buildbarn/bb-storage/internal/mock

replace github.com/gordonklaus/ineffassign => github.com/gordonklaus/ineffassign v0.0.0-20230610083614-0e73809eb601

replace mvdan.cc/gofumpt => mvdan.cc/gofumpt v0.5.0
