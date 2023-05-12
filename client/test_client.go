package main

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"

	"github_actions/commitinfo"
)

func CheckErr(err error, msg string) {
	if err != nil {
		log.Fatalf(msg, err)
	}
}

func main() {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial("localhost:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))

	CheckErr(err, "did not connect: %s")

	defer conn.Close()

	c := commitinfo.NewCommitDataClient(conn)

	//reproduce the main steps with the test repository
	response, _ := c.Translate(context.Background(), &commitinfo.CommitInfo{
		HeadHash:    "dfc984f2f8cebb0f4bc6a460843086a00c405444",
		CommandLine: "bazel build //a_app:a_build //b_app:b_build //c_app:c_build",
		CommitDiff: "diff --git a/a_app/a.go b/a_app/a.go\n" +
			"index f056313..91f7693 100644\n" +
			"--- a/a_app/a.go\n" +
			"+++ b/a_app/a.go\n" +
			"@@ -4,4 +4,5 @@ import \"fmt\"\n" +
			" \n func main() {\n" +
			" \tfmt.Println(\"hello from a\")\n" +
			"+\tfmt.Println(\"hello from a again!!!\")\n }\n",
		AffectedTargets: "//a_app:a.go\n  //a_app:a_build"})
	log.Println(response)

	response, _ = c.Translate(context.Background(), &commitinfo.CommitInfo{
		HeadHash:    "dfc984f2f8cebb0f4bc6a460843086a00c405444",
		CommandLine: "bazel build //a_app:a_build //b_app:b_build //c_app:c_build",
		CommitDiff: "diff --git a/b_app/b.go b/b_app/b.go\n" +
			"index ce7a909..7be6585 100644\n" +
			"--- a/b_app/b.go\n" +
			"+++ b/b_app/b.go\n" +
			"@@ -4,4 +4,5 @@ import \"fmt\"\n" +
			" \n func main() {\n" +
			" \tfmt.Println(\"hello from b\")\n" +
			"+\tfmt.Println(\"hello from b again!!\")\n }\n",
		AffectedTargets: "//b_app:b_build\n  //b_app:b.go"})
	log.Println(response)

	response, _ = c.Translate(context.Background(), &commitinfo.CommitInfo{
		HeadHash:    "5226b93119ca12c51c30b97642cff726993e14aa",
		CommandLine: "bazel build //a_app:a_build //b_app:b_build //c_app:c_build",
		CommitDiff: "diff --git a/c_app/c.go b/c_app/c.go\n" +
			"index f1c60d0..148de42 100644\n" +
			"--- a/c_app/c.go\n" +
			"+++ b/c_app/c.go\n" +
			"@@ -4,8 +4,8 @@ import \"fmt\"\n" +
			" \n func main() {\n" +
			" \tfmt.Println(\"hello from c\")\n" +
			"-\tvar first = 1\n" +
			"-\tfirst = first * 2\n" +
			"-\tfmt.Println(first)\n" +
			"+\tvar second = 1\n" +
			"+\tsecond = second * 2\n" +
			"+\tfmt.Println(second)\n" +
			" \n }\n",
		AffectedTargets: "//c_app:c.go\n  //c_app:c_build"})
	log.Println(response)

	response, _ = c.Translate(context.Background(), &commitinfo.CommitInfo{
		HeadHash:    "5226b93119ca12c51c30b97642cff726993e14aa",
		CommandLine: "bazel build //a_app:a_build //b_app:b_build //c_app:c_build",
		CommitDiff: "diff --git a/c_app/c.go b/c_app/c.go\n" +
			"index f1c60d0..9e016b5 100644\n" +
			"--- a/c_app/c.go\n" +
			"+++ b/c_app/c.go\n" +
			"@@ -6,6 +6,7 @@ func main() {\n" +
			" \tfmt.Println(\"hello from c\")\n" +
			" \tvar first = 1\n" +
			" \tfirst = first * 2\n" +
			"+\tfirst = first * 3\n" +
			" \tfmt.Println(first)\n \n }\n",
		AffectedTargets: "//c_app:c.go\n  //c_app:c_build"})
	log.Println(response)
}
