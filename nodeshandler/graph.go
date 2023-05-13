package nodeshandler

import (
	"bytes"
	"errors"
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

var guard = make(chan int, 8) // max goroutines
var errC = make(chan error)   // return value

type ExecutionStatus int64
type BranchDir int64

var nodesPQ = pq.NewPriorityQueue()

type GraphHandler interface {
	init(command string)
	addNodes(patch int, cur *Node)
	runNode(this *Node)
	deleteNode(this *Node)
	run(node *Node)
}

func (g *Graph) Init(command string) {
	g.Command = command
	g.Root = nil
}

func (g *Graph) AddNodes(patch int, cur *Node) {
	if g.Root != nil { // and cur != nil
		// left
		if cur.Left != nil {
			g.AddNodes(patch, cur.Left)
		} else {
			// add left
			patchAppliedCopy := make([]string, len(cur.PatchApplied))
			_ = copy(patchAppliedCopy, cur.PatchApplied)
			changes := patchAppliedCopy[:len(patchAppliedCopy)-1] //remove last change
			changes = append(changes, utils.ToString(patch))
			var priority = getPriority(changes)

			log.Println("left node: " + strings.Join(changes, ""))
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

			log.Println("right node: " + strings.Join(changes, ""))
			var right = Node{Status: Ready, Path: "nodes/node" + strings.Join(changes, "") + "/",
				PatchApplied: changes, Left: nil, Right: nil, Parent: cur}

			nodesPQ.Insert(&right, priority)
			cur.Right = &right
		}
	} else { //define root
		var changes = []string{utils.ToString(patch)}

		log.Println("root node: " + strings.Join(changes, ""))
		var root = Node{Status: Ready, Path: "nodes/node" + strings.Join(changes, "") + "/",
			PatchApplied: changes, Left: nil, Right: nil, Parent: nil}

		nodesPQ.Insert(&root, 0)
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
	var mutex sync.Mutex

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

	_, err = util.RunCommand("docker pull polupanovaanna/github_actions_test_project:main")

	if err != nil {
		return errors.New("error while pulling docker image")
	}

	containerID, err := util.RunCommand("docker create polupanovaanna/github_actions_test_project:main")

	if err != nil {
		return errors.New("error while creating docker client")
	}

	containerID = containerID[:len(containerID)-1]

	_, err = util.RunCommand("docker cp " + containerID + ":/app " + dir)
	_, err = util.RunCommand("docker stop " + containerID)
	_, err = util.RunCommand("sudo chmod -R 777 " + dir)

	//apply patch
	for _, patchNum := range this.PatchApplied {
		var diffPatch = "patch" + patchNum + ".diff"

		mutex.Lock()
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
		mutex.Unlock()

		for _, targetFile := range files {
			mutex.Lock()
			file, err := os.OpenFile(dir+targetFile.OldName, os.O_CREATE|os.O_APPEND, os.ModePerm)

			if err != nil {
				return errors.New("Error while opening " + targetFile.OldName)
			}

			var output bytes.Buffer
			err = gitdiff.Apply(&output, file, targetFile)

			if err != nil {
				return errors.New("Error while applying changes " + targetFile.OldName)
			}

			err = file.Close()
			if err != nil {
				return errors.New("Error while closing " + targetFile.OldName)
			}

			err = ioutil.WriteFile(dir+targetFile.OldName, output.Bytes(), 0)
			if err != nil {
				return errors.New("Error while writing to file " + targetFile.OldName)
			}
			mutex.Unlock()
		} // patch is successfully applied
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

		return errors.New("there are possible conflicts. Pull request could not be merged")
	}

	this.Status = Successful

	return nil
}

func decreaseGuard() {
	select {
	case <-guard:
	default:
	}
}

func (g *Graph) Run(node *Node) error {
	cnt := 0

	if node == nil {
		decreaseGuard()

		return nil
	}

	if node.Status == Ready {
		log.Print("start run Node " + utils.ToString(node.PatchApplied) + "\n")
		err := g.runNode(node)
		log.Print("end run Node " + utils.ToString(node.PatchApplied) + "\n")
		decreaseGuard()

		return err
	}

	var priorityNode = nodesPQ.PopHighest()

	var nodeDir = getPriorityBranch(node, priorityNode.(*Node))

	if nodeDir == Left {
		if node.Status == Successful || node.Status == Failed {
			cnt++

			guard <- 1
			go g.RunLeft(node)
		}

		if node.Status == Successful {
			cnt++

			guard <- 1
			go g.RunRight(node)
		}
	} else {
		if node.Status == Successful {
			cnt++

			guard <- 1
			go g.RunRight(node)
		}

		if node.Status == Successful || node.Status == Failed {
			cnt++

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

	return nil
}

func (g *Graph) RunLeft(node *Node) {
	log.Print("start run Left Node " + utils.ToString(node.PatchApplied) + "\n")
	err := g.Run(node.Left)
	log.Print("end run Left Node " + utils.ToString(node.PatchApplied) + "\n")
	errC <- err
}

func (g *Graph) RunRight(node *Node) {
	log.Print("start run Right Node " + utils.ToString(node.PatchApplied) + "\n")
	err := g.Run(node.Right)
	log.Print("end run Right Node " + utils.ToString(node.PatchApplied) + "\n")
	errC <- err
}
