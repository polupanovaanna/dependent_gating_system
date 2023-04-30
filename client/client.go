package main

import (
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"

	"github_actions/commit_info"
)

func CheckErr(err error, msg string) {
	if err != nil {
		log.Fatalf(msg, err)
	}
}

func main() {

	//client is now unused and may be deleted later

	//url, directory := os.Args[1], os.Args[2]

	var conn *grpc.ClientConn
	conn, err := grpc.Dial("localhost:8080", grpc.WithInsecure())

	CheckErr(err, "did not connect: %s")

	defer conn.Close()

	c := commit_info.NewCommitDataClient(conn)

	/*r, err := git.PlainClone(directory, false, &git.CloneOptions{
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

			response, err := c.Translate(context.Background(), &commit_info.CommitInfo{HeadHash: masterHeadRef.Hash().String(),
				CommitDiff: patch.String()})
			CheckErr(err, "Error when translating info to request_handler: %s")
			fmt.Println(response)
			fmt.Println("branch: ", patch.String())
		}
		return nil
	}) //iterating branches*/

	response, err := c.Translate(context.Background(), &commit_info.CommitInfo{
		HeadHash:        "dfc984f2f8cebb0f4bc6a460843086a00c405444",
		CommandLine:     "bazel build //a_app:a_build",
		CommitDiff:      "diff --git a/a_app/a.go b/a_app/a.go\nindex f056313..91f7693 100644\n--- a/a_app/a.go\n+++ b/a_app/a.go\n@@ -4,4 +4,5 @@ import \"fmt\"\n \n func main() {\n \tfmt.Println(\"hello from a\")\n+\tfmt.Println(\"hello from a again!!!\")\n }\n",
		AffectedTargets: "a.go"})
	fmt.Println(response)

	response, err = c.Translate(context.Background(), &commit_info.CommitInfo{
		HeadHash:        "dfc984f2f8cebb0f4bc6a460843086a00c405444",
		CommandLine:     "bazel build //a_app:a_build",
		CommitDiff:      "diff --git a/b_app/b.go b/b_app/b.go\nindex ce7a909..7be6585 100644\n--- a/b_app/b.go\n+++ b/b_app/b.go\n@@ -4,4 +4,5 @@ import \"fmt\"\n \n func main() {\n \tfmt.Println(\"hello from b\")\n+\tfmt.Println(\"hello from b again!!\")\n }\n",
		AffectedTargets: "a.go"})
	fmt.Println(response)

	CheckErr(err, "Error when processing git info: %s")

}
