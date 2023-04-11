package nodes_handler

import (
	"bytes"
	"errors"
	"github.com/bluekeyes/go-gitdiff/gitdiff"
	"github.com/docker/docker/client"
	"github.com/emirpasic/gods/utils"
	"github_actions/util"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

type ExecutionStatus int64

type Graph struct {
	Command string
	Root    *Node //root Node
}

const (
	Ready ExecutionStatus = iota
	Running
	Successful
	Failed
)

type Node struct {
	Status       ExecutionStatus //current status
	Path         string          //path to local storage
	PatchApplied []string        //numbers of patches
	Left         *Node           //left child
	Right        *Node           //right child
	Parent       *Node           //parent
}

type GraphHandler interface {
	init(command string)
	addNodes(patch int, cur *Node)
	runNode(this *Node)
	deleteNode(this *Node)
	run(node *Node) //need to be changed
}

func (g *Graph) Init(command string) {
	g.Command = command
	g.Root = nil
}

func (g *Graph) AddNodes(patch int, cur *Node) {
	if g.Root != nil { //and cur != nil
		//left
		if cur.Left != nil {
			g.AddNodes(patch, cur.Left)
		} else {
			//add left
			var changes = cur.PatchApplied[:len(cur.PatchApplied)-1] //remove last change
			changes = append(changes, utils.ToString(patch))
			var left = Node{Status: Ready, Path: "nodes/node/" + strings.Join(changes, "") + "/",
				PatchApplied: changes, Left: nil, Right: nil, Parent: cur}
			cur.Left = &left
		}
		//right
		if cur.Right != nil {
			g.AddNodes(patch, cur.Right)
		} else {
			//add right
			var changes = append(cur.PatchApplied, utils.ToString(patch))
			var right = Node{Status: Ready, Path: "nodes/node" + strings.Join(changes, "") + "/",
				PatchApplied: changes, Left: nil, Right: nil, Parent: cur}
			cur.Right = &right
		}
	} else { //define root
		var changes = []string{utils.ToString(patch)}
		var root = Node{Status: Ready, Path: "nodes/node" + strings.Join(changes, "") + "/",
			PatchApplied: changes, Left: nil, Right: nil, Parent: nil}
		g.Root = &root
	}

}

func (g *Graph) deleteNode(this *Node) {
	if this.Parent != nil {
		if this.Parent.Left == this {
			this.Parent.Left = nil
		} else {
			this.Parent.Right = nil
		}
	}
}

func (g *Graph) runNode(this *Node) error {
	this.Status = Running
	var dir = this.Path

	for _, patchNum := range this.PatchApplied {
		var diffPatch = "patch" + patchNum + ".diff"

		patch, err := os.Open(diffPatch)
		files, _, err := gitdiff.Parse(patch)
		err = patch.Close()
		util.CheckErr(err, "Failed patch reading")

		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		util.CheckErr(err, "Error while creating docker client")
		defer cli.Close()

		util.RunCommand("docker pull polupanovaanna/github_actions_test_project:main")
		containerId := util.RunCommand("docker create polupanovaanna/github_actions_test_project:main")
		containerId = containerId[:len(containerId)-1]
		util.RunCommand("docker cp " + containerId + ":/app " + dir)
		util.RunCommand("docker stop " + containerId)
		util.RunCommand("sudo chmod -R 777 " + dir) //TODO fix that

		for _, f := range files {
			file, err := os.OpenFile(dir+f.OldName, os.O_CREATE|os.O_APPEND, os.ModePerm)
			util.CheckErr(err, "Error while opening "+f.OldName)

			var output bytes.Buffer
			err = gitdiff.Apply(&output, file, f)
			util.CheckErr(err, "Error while applying changes "+f.OldName)

			err = file.Close()
			util.CheckErr(err, "Error while closing "+f.OldName)

			err = ioutil.WriteFile(dir+f.OldName, output.Bytes(), 0)
			util.CheckErr(err, "Error while writing to file "+f.OldName)
		}
		// patch is successfully applied
	}

	err := os.Chdir(dir)
	util.CheckErr(err, "failed to find directory"+dir)

	args := strings.Split(g.Command, " ")
	cmd := exec.Command(args[0], args[1:]...)

	err = cmd.Run()

	if err != nil {
		this.Status = Failed
		g.deleteNode(this.Right)
		return errors.New("There are possible conflicts. Pull request could not be merged!")
	}

	this.Status = Successful
	return nil
}

func (g *Graph) Run(node *Node) error {
	if node != nil {
		if node.Status == Successful {
			err := g.Run(node.Left)
			if err != nil {
				return err
			}
			err = g.Run(node.Right)
			return err
		}
		if node.Status == Failed {
			err := g.Run(node.Left)
			return err
		}
		if node.Status == Ready {
			err := g.runNode(node)
			return err
		}
	}
	return nil
}
