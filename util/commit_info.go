package util

import (
	"log"

	"golang.org/x/net/context"
)

type Server struct {
}

func (s *Server) Translate(ctx context.Context, in *CommitInfo) (*ServerResponse, error) {
	log.Printf("Receive message body from client: %s", in.HeadHash)
	// вот тут мне не то, чтобы все понятно, потому что у нас получается есть всегда на сервере
	// master версия репозитория

	return &ServerResponse{Response: "Evetything is ok!"}, nil
}
