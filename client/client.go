package main

import (
	"fmt"
	"github.com/go-git/go-git/v5/plumbing"
	"log"
	"os"

	"github.com/go-git/go-git/v5"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github_actions/util"
)

func CheckErr(err error, msg string) {
	if err != nil {
		log.Fatalf(msg, err)
	}
}

func main() {

	url, directory := os.Args[1], os.Args[2]
	//TODO тут должна быть какая-то валидация урла кажется

	var conn *grpc.ClientConn
	conn, err := grpc.Dial("localhost:8080", grpc.WithInsecure())

	CheckErr(err, "did not connect: %s")

	defer conn.Close()

	c := util.NewCommitDataClient(conn)

	r, err := git.PlainClone(directory, false, &git.CloneOptions{
		URL: url,
	})
	CheckErr(err, "Error when uploading git repository: %s")

	masterHeadRef, err := r.Head() //master HEAD hash
	masterHeadCommit, _ := r.CommitObject(masterHeadRef.Hash())

	refIter, _ := r.References()

	refIter.ForEach(func(ref *plumbing.Reference) error {

		if ref.Name().IsRemote() {
			branchCommit, _ := r.CommitObject(ref.Hash())
			patch, _ := masterHeadCommit.Patch(branchCommit)

			response, err := c.Translate(context.Background(), &util.CommitInfo{HeadHash: masterHeadRef.Hash().String(),
				CommitDiff: patch.String()})
			CheckErr(err, "Error when translating info to server: %s")
			fmt.Println(response)
			fmt.Println("branch: ", patch.String())
		}
		return nil
	}) //iterating branches

	//response, err := c.Translate(context.Background(), &util.CommitInfo{HeadHash: masterHeadRef.String()})

	CheckErr(err, "Error when processing git info: %s")

	//log.Printf("Response from server: %s", response)

}
