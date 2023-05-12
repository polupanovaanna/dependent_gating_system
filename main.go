package main

import (
	"fmt"
	"github_actions/commitinfo"
	"log"
	"net"

	"google.golang.org/grpc"
)

func main() {
	log.Println("Server is starting...")

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", 8080))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := commitinfo.Server{}

	grpcServer := grpc.NewServer()

	commitinfo.RegisterCommitDataServer(grpcServer, &s)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}
}
