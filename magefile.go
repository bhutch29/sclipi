// +build mage

package main

import (
	"github.com/magefile/mage/sh"
)

var Default = Install

func Build() error {
	return sh.Run("go", "build", ".")
}

func BuildWindows() error {
	env := map[string]string{"GOOS": "windows", "GOARCH": "amd64"}
	return sh.RunWith(env, "go", "build", ".")
}

func Install() error {
	version, err := sh.Output("git", "describe", "--always", "--long", "--dirty")
	if err != nil {
		return err
	}
	return sh.Run("go", "install", "-v", "-ldflags", "-X main.version="+version, ".")
}

func Test() error {
	return sh.RunV("go", "test", "-v")
}

func Clean() {
	sh.Rm("sclipi")
	sh.Rm("sclipi.exe")
}
