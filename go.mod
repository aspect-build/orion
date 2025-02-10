module github.com/aspect-build/silo

go 1.23.0

require (
	github.com/bazelbuild/bazel-gazelle v0.42.0 // NOTE: keep in sync with fetch.bzl
	github.com/bazelbuild/buildtools v0.0.0-20240918101019-be1c24cc9a44
	github.com/bazelbuild/rules_go v0.52.0 // indirect; NOTE: keep in sync with fetch.bzl
	github.com/bmatcuk/doublestar/v4 v4.7.1
	github.com/emirpasic/gods v1.18.1
	github.com/fatih/color v1.18.0 // indirect
	github.com/go-git/go-git/v5 v5.13.0
	github.com/itchyny/gojq v0.12.16
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	github.com/smacker/go-tree-sitter v0.0.0-20240827094217-dd81d9e9be82
	github.com/yargevad/filepathx v1.0.0
	go.starlark.net v0.0.0-20240725214946-42030a7cedce
	golang.org/x/sync v0.10.0
	google.golang.org/protobuf v1.34.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require (
	github.com/mikefarah/yq/v4 v4.45.1
	golang.org/x/term v0.28.0
	gopkg.in/op/go-logging.v1 v1.0.0-20160211212156-b2cb9fa56473
)

require (
	github.com/a8m/envsubst v1.4.2 // indirect
	github.com/alecthomas/assert/v2 v2.7.0 // indirect
	github.com/alecthomas/participle/v2 v2.1.1 // indirect
	github.com/dimchansky/utfbom v1.1.1 // indirect
	github.com/elliotchance/orderedmap v1.7.1 // indirect
	github.com/go-git/gcfg v1.5.1-0.20230307220236-3a3c6141e376 // indirect
	github.com/go-git/go-billy/v5 v5.6.0 // indirect
	github.com/goccy/go-json v0.10.4 // indirect
	github.com/goccy/go-yaml v1.13.3 // indirect
	github.com/itchyny/timefmt-go v0.1.6 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/jinzhu/copier v0.4.0 // indirect
	github.com/yuin/gopher-lua v1.1.1 // indirect
	golang.org/x/text v0.21.0 // indirect
	golang.org/x/tools/go/vcs v0.1.0-deprecated // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
)

require (
	github.com/chzyer/readline v1.5.1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/magiconair/properties v1.8.9 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/pelletier/go-toml/v2 v2.2.3 // indirect
	golang.org/x/mod v0.22.0 // indirect
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
)

replace github.com/gordonklaus/ineffassign => github.com/gordonklaus/ineffassign v0.1.0

replace mvdan.cc/gofumpt => mvdan.cc/gofumpt v0.6.0
