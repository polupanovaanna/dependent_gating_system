package nodes_handler

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/bluekeyes/go-gitdiff/gitdiff"
	"github.com/docker/docker/client"
	"github.com/emirpasic/gods/utils"
	pq "github.com/kyroy/priority-queue"
	"github_actions/util"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
)

var guard = make(chan int, 8) //max goroutines
var errC = make(chan error)   //return value

type ExecutionStatus int64
type BranchDir int64

var nodesPQ = pq.NewPriorityQueue()

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
			var priority = getPriority(changes)

			fmt.Print("left node: " + strings.Join(changes, "") + "\n")
			var left = Node{Status: Ready, Path: "nodes/node" + strings.Join(changes, "") + "/",
				PatchApplied: changes, Left: nil, Right: nil, Parent: cur}

			nodesPQ.Insert(&left, priority)
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
			var priority = getPriority(changes)

			fmt.Print("right node: " + strings.Join(changes, "") + "\n")
			var right = Node{Status: Ready, Path: "nodes/node" + strings.Join(changes, "") + "/",
				PatchApplied: changes, Left: nil, Right: nil, Parent: cur}

			nodesPQ.Insert(&right, priority)
			cur.Right = &right
		}
	} else { //define root
		var changes = []string{utils.ToString(patch)}

		fmt.Print("root node: " + strings.Join(changes, "") + "\n")
		var root = Node{Status: Ready, Path: "nodes/node" + strings.Join(changes, "") + "/",
			PatchApplied: changes, Left: nil, Right: nil, Parent: nil}

		nodesPQ.Insert(&root, 10)
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
	var m sync.Mutex
	util.DirSetup()

	this.Status = Running
	var dir = this.Path
	log.Print("Current dir: " + dir)

	//upload up-to-date docker image
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return errors.New("Error while creating docker client")
	}
	defer cli.Close()

	util.RunCommand("docker pull polupanovaanna/github_actions_test_project:main")
	containerId := util.RunCommand("docker create polupanovaanna/github_actions_test_project:main")
	containerId = containerId[:len(containerId)-1]

	util.RunCommand("docker cp " + containerId + ":/app " + dir)
	util.RunCommand("docker stop " + containerId)
	util.RunCommand("sudo chmod -R 777 " + dir)

	//apply patch
	for _, patchNum := range this.PatchApplied {
		var diffPatch = "patch" + patchNum + ".diff"

		m.Lock()
		util.DirSetup()
		patch, err := os.Open(diffPatch)
		if err != nil {
			return errors.New("Failed patch opening")
		}

		files, _, err := gitdiff.Parse(patch)
		if err != nil {
			return errors.New("Failed patch parsing")
		}

		err = patch.Close()
		if err != nil {
			return errors.New("Failed patch closing")
		}
		m.Unlock()

		for _, f := range files {
			m.Lock()
			file, err := os.OpenFile(dir+f.OldName, os.O_CREATE|os.O_APPEND, os.ModePerm)
			if err != nil {
				return errors.New("Error while opening " + f.OldName)
			}

			var output bytes.Buffer
			err = gitdiff.Apply(&output, file, f)
			if err != nil {
				return errors.New("Error while applying changes " + f.OldName)
			}

			err = file.Close()
			if err != nil {
				return errors.New("Error while closing " + f.OldName)
			}

			err = ioutil.WriteFile(dir+f.OldName, output.Bytes(), 0)
			if err != nil {
				return errors.New("Error while writing to file " + f.OldName)
			}
			m.Unlock()
		}
		// patch is successfully applied
	}
	err = os.Chdir(dir)
	if err != nil {
		return errors.New("failed to find directory" + dir)
	}

	args := strings.Split(g.Command, " ")
	cmd := exec.Command(args[0], args[1:]...)

	err = cmd.Run()

	if err != nil {
		this.Status = Failed
		if this.Right != nil {
			g.deleteNode(this.Right)
		}
		return errors.New("There are possible conflicts. Pull request could not be merged!")
	}

	this.Status = Successful
	return nil
}

func decreaseGuard() {
	select {
	case _, _ = <-guard:
	default:
		fmt.Println("")
	}
}

func (g *Graph) Run(node *Node) error {
	cnt := 0
	if node != nil {
		if node.Status == Ready {
			fmt.Print("start run Node " + utils.ToString(node.PatchApplied) + "\n")
			err := g.runNode(node)
			fmt.Print("end run Node " + utils.ToString(node.PatchApplied) + "\n")
			decreaseGuard()
			return err
		}

		var priorityNode = nodesPQ.PopHighest()
		var nodeDir = getPriorityBranch(node, priorityNode.(*Node))

		if nodeDir == Left {
			if node.Status == Successful || node.Status == Failed {
				cnt += 1
				guard <- 1
				go g.RunLeft(node)
			}
			if node.Status == Successful {
				cnt += 1
				guard <- 1
				go g.RunRight(node)
			}
		} else {
			if node.Status == Successful {
				cnt += 1
				guard <- 1
				go g.RunRight(node)
			}
			if node.Status == Successful || node.Status == Failed {
				cnt += 1
				guard <- 1
				go g.RunLeft(node)
			}
		}

		decreaseGuard()

		var err error
		for i := 0; i < cnt; i++ {
			err = <-errC
			if err != nil {
				return err
			}
		}
	}

	decreaseGuard()
	return nil
}

func (g *Graph) RunLeft(node *Node) {
	fmt.Print("start run Left Node " + utils.ToString(node.PatchApplied) + "\n")
	err := g.Run(node.Left)
	fmt.Print("end run Left Node " + utils.ToString(node.PatchApplied) + "\n")
	errC <- err
}

func (g *Graph) RunRight(node *Node) {
	fmt.Print("start run Right Node " + utils.ToString(node.PatchApplied) + "\n")
	err := g.Run(node.Right)
	fmt.Print("end run Right Node " + utils.ToString(node.PatchApplied) + "\n")
	errC <- err
}
