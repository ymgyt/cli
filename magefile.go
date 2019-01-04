// +build mage

package main

import (
	// "github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Test run go test.
func Test() {
	sh.RunV("go", "test", "./...")
}
