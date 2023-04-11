package util

import (
	"os/exec"
	"strings"
)

func RunCommand(commandLine string) string {
	args := strings.Split(commandLine, " ")
	command := exec.Command(args[0], args[1:]...)
	var out strings.Builder
	command.Stdout = &out
	err := command.Run()
	CheckErr(err, "Error executing command: "+commandLine)
	return out.String()
}
