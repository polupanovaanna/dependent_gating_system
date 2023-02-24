package main

import (
	"log"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github_actions/util"
)

func main() {

	var conn *grpc.ClientConn
	conn, err := grpc.Dial(":9000", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := util.NewCommitDataClient(conn)

	response, err := c.Translate(context.Background(), &util.CommitInfo{HeadHash: "Here is the hash"})
	if err != nil {
		log.Fatalf("Error when calling SayHello: %s", err)
	}
	log.Printf("Response from server: %s", response.Response)

}
