package nodeshandler

import (
	"github_actions/util"
	"log"
	"os"
	"strconv"
	"strings"
)

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

const (
	Right BranchDir = iota
	Left
)

type Node struct {
	Status       ExecutionStatus //current status
	Path         string          //path to local storage
	PatchApplied []string        //numbers of patches
	Left         *Node           //left child
	Right        *Node           //right child
	Parent       *Node           //parent
}

func getPriorityBranch(root *Node, node *Node) BranchDir {
	var rootPatches = strings.Join(root.PatchApplied, "")

	var nodePatches = strings.Join(node.PatchApplied, "")

	if strings.Contains(nodePatches, rootPatches) {
		return Right
	}

	return Left
}

func getTargetsFromPatch(patchNumber string) map[string]struct{} {
	var targets = make(map[string]struct{})

	var diffPatch = "patch" + patchNumber + ".diff"

	util.DirSetup()
	var patchB, err = os.ReadFile(diffPatch)
	util.CheckErr(err, "Failed patch reading")

	var patch = string(patchB)
	var lines = strings.Split(patch, "\n")

	for i := 0; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "--- a/") {
			targets[strings.TrimPrefix(lines[i], "--- a/")] = struct{}{}
		}
	}

	//returns set with affected files
	return targets
}

func getPriority(changes []string) float64 {
	var initTargets = getTargetsFromPatch(changes[0]) //get targets from first patch
	var priority = 1

	for i := 1; i < len(changes); i++ {
		var patchTargets = getTargetsFromPatch(changes[i])
		for target := range patchTargets {
			if _, exists := initTargets[target]; exists {
				priority++
			} else {
				initTargets[target] = struct{}{}
			}
		}
	}
	log.Println("Got priority for the node " + strings.Join(changes, "") + ": " + strconv.Itoa(priority))

	return float64(priority)
}
