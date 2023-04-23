package nodes_handler

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/bluekeyes/go-gitdiff/gitdiff"
	"github.com/docker/docker/client"
	"github.com/emirpasic/gods/utils"
	"github_actions/util"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

var sem = make(chan int, 8) //max goroutines
var errC = make(chan error) //max goroutines

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
			patchAppliedCopy := make([]string, len(cur.PatchApplied))
			_ = copy(patchAppliedCopy, cur.PatchApplied)
			changes := patchAppliedCopy[:len(patchAppliedCopy)-1] //remove last change
			changes = append(changes, utils.ToString(patch))
			fmt.Print("left node: " + strings.Join(changes, "") + "\n")
			var left = Node{Status: Ready, Path: "nodes/node" + strings.Join(changes, "") + "/",
				PatchApplied: changes, Left: nil, Right: nil, Parent: cur}
			cur.Left = &left
		}
		//right
		if cur.Right != nil {
			g.AddNodes(patch, cur.Right)
		} else {
			//add right
			patchAppliedCopy := make([]string, len(cur.PatchApplied))
			_ = copy(patchAppliedCopy, cur.PatchApplied)
			var changes = append(patchAppliedCopy, utils.ToString(patch))
			fmt.Print("right node: " + strings.Join(changes, "") + "\n")
			var right = Node{Status: Ready, Path: "nodes/node" + strings.Join(changes, "") + "/",
				PatchApplied: changes, Left: nil, Right: nil, Parent: cur}
			cur.Right = &right
		}
	} else { //define root
		var changes = []string{utils.ToString(patch)}
		fmt.Print("root node: " + strings.Join(changes, "") + "\n")
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
	util.DirSetup()

	this.Status = Running
	var dir = this.Path
	log.Print("Current dir: " + dir)

	//upload up-to-date docker image
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	util.CheckErr(err, "Error while creating docker client")
	defer cli.Close()

	util.RunCommand("docker pull polupanovaanna/github_actions_test_project:main")
	containerId := util.RunCommand("docker create polupanovaanna/github_actions_test_project:main")
	containerId = containerId[:len(containerId)-1]

	util.RunCommand("docker cp " + containerId + ":/app " + dir)
	util.RunCommand("docker stop " + containerId)
	util.RunCommand("sudo chmod -R 777 " + dir) //TODO fix that

	//apply patch
	for _, patchNum := range this.PatchApplied {
		var diffPatch = "patch" + patchNum + ".diff"

		util.DirSetup()
		patch, err := os.Open(diffPatch)
		util.CheckErr(err, "Failed patch opening")

		files, _, err := gitdiff.Parse(patch)
		util.CheckErr(err, "Failed patch parsing")

		err = patch.Close()
		util.CheckErr(err, "Failed patch closing")

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
	err = os.Chdir(dir)
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
	cnt := 0
	if node != nil {

		if node.Status == Successful || node.Status == Failed {
			fmt.Print(1)
			cnt += 1
			go func() {
				fmt.Print("start run Left Node " + utils.ToString(node.PatchApplied) + "\n")
				err := g.Run(node.Left)
				fmt.Print("end run Left Node " + utils.ToString(node.PatchApplied) + "\n")
				errC <- err
			}()
			fmt.Print(2)
		}
		if node.Status == Successful {
			cnt += 1
			go func() {
				fmt.Print("start run Right Node " + utils.ToString(node.PatchApplied) + "\n")
				err := g.Run(node.Right)
				fmt.Print("end run Right Node " + utils.ToString(node.PatchApplied) + "\n")
				errC <- err
			}()
		}
		if node.Status == Ready {
			fmt.Print("start run Node " + utils.ToString(node.PatchApplied) + "\n")
			err := g.runNode(node)
			fmt.Print("end run Node " + utils.ToString(node.PatchApplied) + "\n")
			return err
		}

		var err error
		for i := 0; i < cnt; i++ {
			err = <-errC
			if err != nil {
				return err
			}
		}
	}

	return nil
}
