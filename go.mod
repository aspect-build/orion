module github.com/aspect-build/silo

go 1.18

require (
	github.com/aspect-build/talkie v0.0.0-00010101000000-000000000000
	github.com/bazelbuild/rules_go v0.34.0
	github.com/onsi/ginkgo/v2 v2.1.4
	github.com/onsi/gomega v1.20.0
)

require (
	github.com/google/go-cmp v0.5.8 // indirect
	golang.org/x/net v0.0.0-20220812174116-3211cb980234 // indirect
	golang.org/x/sys v0.0.0-20220811171246-fbc7d0a398ab // indirect
	golang.org/x/text v0.3.7 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/aspect-build/talkie => ./talkie
