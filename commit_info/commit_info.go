package commit_info

import (
	"github.com/emirpasic/gods/utils"

	"github_actions/nodes_handler"
	"github_actions/util"
	"golang.org/x/net/context"
	"log"
	"os"
)

var NodeCounter = 0 //global
var MainGraph = nodes_handler.Graph{}

type Server struct {
	UnimplementedCommitDataServer
}

func (s *Server) Translate(ctx context.Context, in *CommitInfo) (*ServerResponse, error) {
	log.Printf("Receive message body from client: %s", in.HeadHash)
	log.Println("Affected targets " + in.AffectedTargets)
	util.DirSetup()

	if NodeCounter == 0 {
		util.ClearAll()
		MainGraph.Init(in.CommandLine)
	}

	NodeCounter += 1
	diffPath := "patch" + utils.ToString(NodeCounter) + ".diff"

	file, err := os.Create(diffPath)
	util.CheckErr(err, "Error while creating file "+diffPath)

	_, err = file.WriteString(in.CommitDiff)
	_, err = file.WriteString("\n")
	util.CheckErr(err, "Error while file writing")

	err = file.Close()
	util.CheckErr(err, "Error while file saving")

	MainGraph.AddNodes(NodeCounter, MainGraph.Root)

	err = MainGraph.Run(MainGraph.Root)

	if err != nil {
		return &ServerResponse{Response: "The pull request could not be merged: " + err.Error()}, err
	}

	return &ServerResponse{Response: "Everything is ok!"}, nil
}
