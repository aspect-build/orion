module github.com/aspect-build/aspect-gazelle/runner

replace github.com/aspect-build/aspect-gazelle => ../

go 1.24.5

require (
	github.com/EngFlow/gazelle_cc v0.1.0 // NOTE: keep in sync with MODULE.bazel
	github.com/bazel-contrib/rules_python/gazelle v0.0.0-20250715122526-cab415d82ebb // NOTE: keep in sync with MODULE.bazel
	github.com/bazelbuild/bazel-gazelle v0.45.1-0.20250924144014-2de7b829fef1 // NOTE: keep in sync with MODULE.bazel
	github.com/bazelbuild/buildtools v0.0.0-20250715102656-62b9413b08bb
	go.opentelemetry.io/otel v1.37.0
	go.opentelemetry.io/otel/trace v1.37.0
	golang.org/x/term v0.33.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/Masterminds/semver/v3 v3.4.0 // indirect
	github.com/a8m/envsubst v1.4.3 // indirect
	github.com/alecthomas/participle/v2 v2.1.4 // indirect
	github.com/bmatcuk/doublestar/v4 v4.9.1 // indirect
	github.com/dimchansky/utfbom v1.1.1 // indirect
	github.com/elliotchance/orderedmap v1.8.0 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/fatih/color v1.18.0
	github.com/go-git/gcfg v1.5.1-0.20230307220236-3a3c6141e376 // indirect
	github.com/go-git/go-billy/v5 v5.6.2 // indirect
	github.com/go-git/go-git/v5 v5.16.2
	github.com/go-ini/ini v1.67.0 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/goccy/go-yaml v1.18.0 // indirect
	github.com/itchyny/gojq v0.12.17 // indirect
	github.com/itchyny/timefmt-go v0.1.6 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/jinzhu/copier v0.4.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/msolo/jsonr v0.0.0-20231023064044-62fbfc3a0313 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2
	github.com/smacker/go-tree-sitter v0.0.0-20240827094217-dd81d9e9be82 // indirect
	github.com/yuin/gopher-lua v1.1.1 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel/metric v1.37.0 // indirect
	go.starlark.net v0.0.0-20250717191651-336a4b3a6d1d // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/mod v0.26.0 // indirect
	golang.org/x/sync v0.16.0 // indirect
	golang.org/x/text v0.27.0 // indirect
	golang.org/x/tools/go/vcs v0.1.0-deprecated // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
)

require (
	github.com/aspect-build/aspect-gazelle v0.0.0-00010101000000-000000000000
	github.com/bazel-contrib/rules_jvm v0.30.0
)

require (
	github.com/bazelbuild/rules_go v0.56.1 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/hashicorp/golang-lru v1.0.2 // indirect
	github.com/magiconair/properties v1.8.10 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mikefarah/yq/v4 v4.46.1 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/rs/zerolog v1.33.0 // indirect
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250204164813-702378808489 // indirect
	google.golang.org/grpc v1.70.0 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
	gopkg.in/op/go-logging.v1 v1.0.0-20160211212156-b2cb9fa56473 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
