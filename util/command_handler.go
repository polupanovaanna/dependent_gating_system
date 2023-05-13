package util

import (
	"os"
	"os/exec"
	"strings"
)

func ClearAll() {
	DirSetup()

	_, err := RunCommand("rm -f patch*")
	CheckErr(err, "error while clearing old patches")

	_, err = RunCommand("rm -f -r nodes")
	CheckErr(err, "error while clearing old nodes content")

	_, err = RunCommand("mkdir nodes")
	CheckErr(err, "error while creating nodes directory")
}

func DirSetup() {
	dirname, err := os.UserHomeDir()
	CheckErr(err, "user doesn't have home directory")

	err = os.Chdir(dirname + "/dependent_gating_system")
	CheckErr(err, "Error returning to current directory")
}

func RunCommand(commandLine string) (string, error) {
	var args = strings.Split(commandLine, " ")
	var command = exec.Command(args[0], args[1:]...)

	var out strings.Builder
	var errOut strings.Builder

	command.Stdout = &out
	command.Stderr = &errOut
	err := command.Run()

	DirSetup()

	return out.String(), err
}
