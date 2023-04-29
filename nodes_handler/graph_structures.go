package nodes_handler

import (
	"container/heap"
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
	Priority     int             //priority of Node due to targets affected
	Index        int             //index of Node in Priority queue
}

type NodesPriorityQueue []*Node

func (pq NodesPriorityQueue) Len() int { return len(pq) }

func (pq NodesPriorityQueue) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return pq[i].Priority > pq[j].Priority
}

func (pq NodesPriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Index = i
	pq[j].Index = j
}

func (pq *NodesPriorityQueue) Push(x any) {
	n := len(*pq)
	item := x.(*Node)
	item.Index = n
	*pq = append(*pq, item)
}

func (pq *NodesPriorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.Index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

// update modifies the priority and value of an Item in the queue.
func (pq *NodesPriorityQueue) update(node *Node, priority int) {
	node.Priority = priority
	heap.Fix(pq, node.Index)
}

func getPriorityBranch(root *Node, node *Node) BranchDir {
	var rootPatches = strings.Join(root.PatchApplied, "")
	var nodePatches = strings.Join(node.PatchApplied, "")

	if strings.Contains(nodePatches, rootPatches) {
		return Right
	} else {
		return Left
	}

}
