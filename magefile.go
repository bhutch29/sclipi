// +build mage

package main

import (
	"github.com/magefile/mage/sh"
)

var Default = Simulate

func Build() error {
	version, err := getGitVersion()
	if err != nil {
		return err
	}
	return sh.Run("go", "build", "-ldflags", "-X main.version="+version, ".")
}

func BuildWindows() error {
	env := map[string]string{"GOOS": "windows", "GOARCH": "amd64"}
	version, err := getGitVersion()
	if err != nil {
		return err
	}
	return sh.RunWith(env, "go", "build", "-ldflags", "-X main.version="+version, ".")
}

func Install() error {
	version, err := getGitVersion()
	if err != nil {
		return err
	}
	return sh.Run("go", "install", "-v", "-ldflags", "-X main.version="+version, ".")
}

func Test() error {
	return sh.RunV("go", "test", "-v")
}

func Bench() error {
	return sh.RunV("go", "test", "-bench", ".")
}

func Cover() error {
	return sh.RunV("go", "test", "-cover")
}

func Simulate() error {
	return sh.RunV("go", "run", ".", "-s", "-q")
}

func Clean() {
	sh.Rm("sclipi")
	sh.Rm("sclipi.exe")
}

func getGitVersion() (string, error) {
	version, err := sh.Output("git", "describe", "--always", "--long", "--dirty")
	if err != nil {
		return "", err
	}
	return version, nil
}
