module github.com/aspect-build/silo

go 1.19

require (
	aspect.build/cli v0.7.0
	github.com/aspect-build/talkie v0.0.0-00010101000000-000000000000
	github.com/bazelbuild/bazel-gazelle v0.26.0
	github.com/bazelbuild/buildtools v0.0.0-20220907133145-b9bfff5d7f91
	github.com/bazelbuild/rules_go v0.35.0
	github.com/bmatcuk/doublestar v1.3.4
	github.com/emirpasic/gods v1.18.1
	github.com/evanw/esbuild v0.15.7
	github.com/go-test/deep v1.0.8
	github.com/golang/mock v1.6.0
	github.com/hashicorp/go-plugin v1.4.5
	github.com/manifoldco/promptui v0.9.0
	github.com/onsi/ginkgo/v2 v2.1.6
	github.com/onsi/gomega v1.20.2
	github.com/sirupsen/logrus v1.9.0
	google.golang.org/genproto v0.0.0-20220909194730-69f6226f97e5
	google.golang.org/grpc v1.49.0
	google.golang.org/protobuf v1.28.1
	gopkg.in/yaml.v2 v2.4.0
	sigs.k8s.io/kind v0.15.0
)

require (
	github.com/BurntSushi/toml v1.2.0 // indirect
	github.com/alessio/shellescape v1.4.1 // indirect
	github.com/bazelbuild/bazelisk v1.14.0 // indirect
	github.com/bgentry/go-netrc v0.0.0-20140422174119-9fd32a8b3d3d // indirect
	github.com/chzyer/readline v1.5.1 // indirect
	github.com/evanphx/json-patch/v5 v5.6.0 // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-cmp v0.5.8 // indirect
	github.com/hashicorp/go-hclog v1.3.0 // indirect
	github.com/hashicorp/go-version v1.6.0 // indirect
	github.com/hashicorp/yamux v0.1.1 // indirect
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/kr/pretty v0.3.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.16 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/go-testing-interface v1.14.1 // indirect
	github.com/oklog/run v1.1.0 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/spf13/cobra v1.5.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/net v0.0.0-20220907135653-1e95f45603a7 // indirect
	golang.org/x/sys v0.0.0-20220908164124-27713097b956 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/tools v0.1.12 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)

replace github.com/aspect-build/talkie => ./talkie
