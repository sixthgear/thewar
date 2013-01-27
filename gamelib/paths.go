package gamelib

import (
	"container/heap"
)

type Node struct {
	index    int
	f, g     int
	cameFrom *Node
}

type PriorityQueue []*Node

func (pq PriorityQueue) Len() int           { return len(pq) }
func (pq PriorityQueue) Less(i, j int) bool { return pq[i].f < pq[j].f }
func (pq PriorityQueue) Swap(i, j int)      { pq[i], pq[j] = pq[j], pq[i] }
func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	(*pq) = (*pq)[0 : n+1]
	(*pq)[n] = x.(*Node)
}
func (pq *PriorityQueue) Pop() interface{} {
	n := len(*pq)
	item := (*pq)[n-1]
	*pq = (*pq)[0 : n-1]
	return item
}

func FindPath(m *Map, start *Hex, goal *Hex) []int {

	g := func(node *Hex) int { return 1 - TMOD[start.Unit.Type][node.TerrainType].MOV }
	h := func(node *Hex) int { return m.Distance(node, goal) }

	if start.Unit.Type == OBJ_AIRCRAFT {
		return []int{goal.Index, start.Index}
	}

	path := make([]int, 0)
	queue := make(PriorityQueue, 0, len(m.grid))
	openset := make(map[int]*Node, len(m.grid))
	closedset := make(map[int]*Node, len(m.grid))

	if g(goal) < 9999 {
		openset[start.Index] = &Node{start.Index, 0, h(start), nil}
		heap.Push(&queue, openset[start.Index])
	}

	for len(queue) > 0 {
		// take the lowest cost path from the open queue
		current := heap.Pop(&queue).(*Node)
		delete(openset, current.index)

		// check if this is the goal and rebuild the path
		if current.index == goal.Index {
			for n := current; n.index != start.Index; n = n.cameFrom {
				path = append(path, n.index)
			}
			return append(path, start.Index)
		}

		// add the.Index to our list of expanded nodes	
		closedset[current.index] = current

		// get list of neighbors
		neighbors := m.Neighbors(m.Index(current.index))
		for _, n := range neighbors {
			_, inClosed := closedset[n.Index]
			_, inOpen := openset[n.Index]
			tentative_g := current.g + g(&n)
			if inClosed || tentative_g >= 9999 {
				continue
			}
			if !inOpen || tentative_g <= openset[n.Index].g {

				if !inOpen {
					openset[n.Index] = new(Node)
				}
				openset[n.Index].index = n.Index
				openset[n.Index].cameFrom = current
				openset[n.Index].g = tentative_g
				openset[n.Index].f = tentative_g + h(&n)

				if !inOpen {
					heap.Push(&queue, openset[n.Index])
				} else {
					// TODO: Right now it is potentially possible for suboptimal paths
					// because we do not modify the heap order in the queue if we come 
					// back and modify a node in the open set. We should remove the
					// element from the heap and repush it if possible
				}
			}
		}
	}

	// no path found!
	return append(path, start.Index, start.Index)
}
