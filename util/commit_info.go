package util

import (
	"bytes"
	"github.com/bluekeyes/go-gitdiff/gitdiff"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

type Server struct {
}

func CheckErr(err error, msg string) {
	if err != nil {
		log.Fatalf(msg, err)
	}
}

func runCommand(commandLine string) string {
	args := strings.Split(commandLine, " ")
	command := exec.Command(args[0], args[1:]...)
	var out strings.Builder
	command.Stdout = &out
	err := command.Run()
	CheckErr(err, "Error executing command: "+commandLine)
	return out.String()
}

func (s *Server) Translate(ctx context.Context, in *CommitInfo) (*ServerResponse, error) {
	log.Printf("Receive message body from client: %s", in.HeadHash)

	dir := "tmp/"

	file, err := os.Create("branch_patch.diff")

	_, err = file.WriteString(in.CommitDiff)
	_, err = file.WriteString("\n")
	CheckErr(err, "Error while file writing")
	err = file.Close()
	CheckErr(err, "Error while file saving")

	patch, err := os.Open("branch_patch.diff")
	files, _, err := gitdiff.Parse(patch)
	err = patch.Close()
	CheckErr(err, "Failed patch reading")

	//ctx_docker := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	CheckErr(err, "Error while creating docker client")
	defer cli.Close()

	runCommand("docker pull polupanovaanna/github_actions_test_project:main")
	containerId := runCommand("docker create polupanovaanna/github_actions_test_project:main")
	containerId = containerId[:len(containerId)-1]
	runCommand("docker cp " + containerId + ":/app " + dir) //TODO unique number of local mount
	runCommand("docker stop " + containerId)
	runCommand("sudo chmod -R 777 " + dir)

	for _, f := range files {

		file, err := os.OpenFile(dir+f.OldName, os.O_CREATE|os.O_APPEND, os.ModePerm)
		CheckErr(err, "Error while opening "+f.OldName)

		var output bytes.Buffer
		err = gitdiff.Apply(&output, file, f)
		CheckErr(err, "Error while applying changes "+f.OldName)

		err = file.Close()
		CheckErr(err, "Error while closing "+f.OldName)

		err = ioutil.WriteFile(dir+f.OldName, output.Bytes(), 0)
		CheckErr(err, "Error while writing to file "+f.OldName)

	}
	// patch is successfully applied

	err = os.Chdir("tmp/")
	CheckErr(err, "failed to find tmp directory")

	args := strings.Split(in.CommandLine, " ")
	cmd := exec.Command(args[0], args[1:]...)

	err = cmd.Run()
	CheckErr(err, "There are possible conflicts. Pull request could not be merged!")

	return &ServerResponse{Response: "Evetything is ok!"}, nil
}
