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

// Tidy add/remove depenedencies.
func Tidy() {
	sh.RunV("go", "mod", "tidy")
}

// Deps install dependency tools.
func Deps() {
	for _, dep := range deps {
		sh.RunV("go", "install", dep)
	}
}

// Lint run linter.
func Lint() {
	sh.RunV("golangci-lint", "run", "--enable-all", "--disable=scopelint,lll,maligned")
}

type Coverage mg.Namespace

func (Coverage) Func() {
	defer coverage()()
	sh.RunV("go", "tool", "cover", "-func=cover.out")
}

func (Coverage) HTML() {
	defer coverage()()
	sh.RunV("go", "tool", "cover", "-html=cover.out")
}

func coverage() func() {
	sh.RunV("go", "test", "-coverprofile=cover.out", "./...")
	return func() {
		sh.Run("rm", "cover.out")
	}
}
