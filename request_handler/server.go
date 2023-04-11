package main

import (
	"fmt"
	"github_actions/commit_info"

	"log"
	"net"

	"google.golang.org/grpc"
)

func main() {

	fmt.Println("Server is starting...")

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", 8080))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := commit_info.Server{}

	grpcServer := grpc.NewServer()

	commit_info.RegisterCommitDataServer(grpcServer, &s)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}
}
