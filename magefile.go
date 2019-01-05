// +build mage

package main

import (
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var deps = []string{
	"github.com/golangci/golangci-lint/cmd/golangci-lint",
}

// Test run go test.
func Test() {
	sh.RunV("go", "test", "./...")
}

// Deps install dependency tools.
func Deps() {
	for _, dep := range deps {
		sh.RunV("go", "install", dep)
	}
}

// Lint run linter.
func Lint() {
	mg.Deps(Deps)
	sh.RunV("golangci-lint", "run")
}
