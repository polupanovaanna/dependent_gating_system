package util

import (
	"os"
	"os/exec"
	"strings"
)

func ClearAll() {
	DirSetup()
	RunCommand("rm -f patch*")
	RunCommand("rm -f -r nodes")
	RunCommand("mkdir nodes")
}

func DirSetup() {
	err := os.Chdir("/home/anna/go_git_actions")
	CheckErr(err, "Error returning to current directory")
}

func RunCommand(commandLine string) string {
	args := strings.Split(commandLine, " ")
	command := exec.Command(args[0], args[1:]...)
	var out strings.Builder
	var errOut strings.Builder
	command.Stdout = &out
	command.Stderr = &errOut
	err := command.Run()
	CheckErr(err, "Error executing command: "+commandLine+"\n Error log: "+errOut.String())
	DirSetup()
	return out.String()
}
