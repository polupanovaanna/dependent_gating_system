package util

import (
	"bytes"
	"github.com/bluekeyes/go-gitdiff/gitdiff"
	"github.com/go-git/go-git/v5"
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

	for _, f := range files {
		//TODO here need to work with docker
		err := os.RemoveAll(dir)
		_, err = git.PlainClone(dir, false, &git.CloneOptions{
			URL: "https://github.com/polupanovaanna/github_actions_test_project.git",
		})
		CheckErr(err, "Error when uploading git repository: %s")

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
	//TODO correct check of build with container

	err = os.Chdir("tmp/")
	CheckErr(err, "failed to find tmp directory")

	args := strings.Split(in.CommandLine, " ")
	cmd := exec.Command(args[0], args[1:]...)

	err = cmd.Run()
	CheckErr(err, "There are possible conflicts. Pull request could not be merged!")

	return &ServerResponse{Response: "Evetything is ok!"}, nil
}
