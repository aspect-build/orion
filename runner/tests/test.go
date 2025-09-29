package main

import (
	"log"
	"os"

	"github.com/aspect-build/aspect-gazelle/runner"
)

func main() {
	log.SetPrefix("gazelle: ")
	log.SetFlags(0) // don't print timestamps

	if wd := os.Getenv("BUILD_WORKSPACE_DIRECTORY"); wd != "" {
		_ = os.Chdir(wd)
	}

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
		log.Fatal(err)
	}
}
