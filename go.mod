module github.com/aspect-build/silo

go 1.18

require (
	aspect.build/talkie v0.0.0-00010101000000-000000000000
	github.com/bazelbuild/rules_go v0.34.0
	github.com/golang/protobuf v1.5.2
	github.com/onsi/ginkgo/v2 v2.1.4
	github.com/onsi/gomega v1.20.0
	github.com/sirupsen/logrus v1.9.0
	google.golang.org/genproto v0.0.0-20220812140447-cec7f5303424
	google.golang.org/grpc v1.48.0
	google.golang.org/protobuf v1.28.1
	sigs.k8s.io/kind v0.14.0
)

require (
	github.com/BurntSushi/toml v1.0.0 // indirect
	github.com/alessio/shellescape v1.4.1 // indirect
	github.com/evanphx/json-patch/v5 v5.6.0 // indirect
	github.com/google/go-cmp v0.5.8 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/pelletier/go-toml v1.9.4 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/spf13/cobra v1.4.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/net v0.0.0-20220812174116-3211cb980234 // indirect
	golang.org/x/sys v0.0.0-20220811171246-fbc7d0a398ab // indirect
	golang.org/x/text v0.3.7 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)

replace aspect.build/talkie => ./talkie
