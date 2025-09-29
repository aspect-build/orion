package main

import (
	"github.com/aspect-build/aspect-gazelle/runner"
)

func main() {
	c := runner.New()

	c.AddLanguage(runner.JavaScript)
	c.AddLanguage(runner.Orion)
	c.AddLanguage(runner.Bzl)
	c.AddLanguage(runner.Go)
	c.AddLanguage(runner.Protobuf)
	c.AddLanguage(runner.Python)
	c.AddLanguage(runner.CC)

	_, err := c.Test()
	if err != nil {
		panic(err)
	}
}
