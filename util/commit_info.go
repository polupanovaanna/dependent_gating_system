package util

import (
	"log"

	"golang.org/x/net/context"
)

type Server struct {
}

func (s *Server) Translate(ctx context.Context, in *CommitInfo) (*ServerResponse, error) {
	log.Printf("Receive message body from client: %s", in.HeadHash)
	return &ServerResponse{Response: "Evetything is ok!"}, nil
}
