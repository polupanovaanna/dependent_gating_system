package main

import (
	"fmt"
	"github_actions/util"

	"log"
	"net"

	"google.golang.org/grpc"
)

func main() {

	fmt.Println("Server is starting...")

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 9000))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := util.Server{}

	grpcServer := grpc.NewServer()

	util.RegisterCommitDataServer(grpcServer, &s)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}
}
