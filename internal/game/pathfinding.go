package game

import (
	"container/heap"
)

// PathNode represents a node in the pathfinding grid.
type PathNode struct {
	X, Y      int
	Walkable  bool
	G, H, F   float64
	Parent    *PathNode
	heapIndex int
}

// PathGrid is the grid used for pathfinding.
type PathGrid struct {
	Width  int
	Height int
	Nodes  [][]*PathNode
}

// NewPathGrid creates a new pathfinding grid.
func NewPathGrid(width, height int) *PathGrid {
	grid := &PathGrid{
		Width:  width,
		Height: height,
		Nodes:  make([][]*PathNode, height),
	}

	for y := range height {
		grid.Nodes[y] = make([]*PathNode, width)
		for x := range width {
			grid.Nodes[y][x] = &PathNode{
				X:        x,
				Y:        y,
				Walkable: true,
			}
		}
	}

	return grid
}

// GetNode returns the node at the given position.
func (g *PathGrid) GetNode(x, y int) *PathNode {
	if x < 0 || x >= g.Width || y < 0 || y >= g.Height {
		return nil
	}

	return g.Nodes[y][x]
}

// SetWalkable sets whether a tile is walkable.
func (g *PathGrid) SetWalkable(x, y int, walkable bool) {
	if node := g.GetNode(x, y); node != nil {
		node.Walkable = walkable
	}
}

// Reset resets all pathfinding data for a new search.
func (g *PathGrid) Reset() {
	for y := 0; y < g.Height; y++ {
		for x := 0; x < g.Width; x++ {
			node := g.Nodes[y][x]
			node.G = 0
			node.H = 0
			node.F = 0
			node.Parent = nil
			node.heapIndex = -1
		}
	}
}

// Point represents a 2D integer point.
type Point struct {
	X, Y int
}

// FindPath uses A* to find the shortest path between two points.
func (g *PathGrid) FindPath(startX, startY, endX, endY int) []Point {
	g.Reset()

	start := g.GetNode(startX, startY)
	end := g.GetNode(endX, endY)

	if start == nil || end == nil || !start.Walkable || !end.Walkable {
		return nil
	}

	openHeap := &nodeHeap{}
	heap.Init(openHeap)

	closed := make(map[*PathNode]bool)

	start.G = 0
	start.H = heuristic(start, end)
	start.F = start.G + start.H
	heap.Push(openHeap, start)

	for openHeap.Len() > 0 {
		item := heap.Pop(openHeap)

		current, ok := item.(*PathNode)
		if !ok {
			continue
		}

		if current == end {
			return reconstructPath(end)
		}

		closed[current] = true

		// Check all 8 neighbors
		neighbors := g.getNeighbors(current)
		for _, neighbor := range neighbors {
			if closed[neighbor] || !neighbor.Walkable {
				continue
			}

			// Calculate movement cost (diagonal = 1.414, straight = 1)
			dx := abs(neighbor.X - current.X)
			dy := abs(neighbor.Y - current.Y)

			moveCost := 1.0
			if dx+dy == 2 {
				moveCost = 1.414
			}

			tentativeG := current.G + moveCost

			if neighbor.heapIndex == -1 {
				// Not in open set
				neighbor.G = tentativeG
				neighbor.H = heuristic(neighbor, end)
				neighbor.F = neighbor.G + neighbor.H
				neighbor.Parent = current
				heap.Push(openHeap, neighbor)
			} else if tentativeG < neighbor.G {
				// Better path found
				neighbor.G = tentativeG
				neighbor.F = neighbor.G + neighbor.H
				neighbor.Parent = current
				heap.Fix(openHeap, neighbor.heapIndex)
			}
		}
	}

	return nil // No path found
}

// getNeighbors returns all valid neighbors (including diagonals).
func (g *PathGrid) getNeighbors(node *PathNode) []*PathNode {
	var neighbors []*PathNode

	dirs := [][2]int{
		{-1, -1},
		{0, -1},
		{1, -1},
		{-1, 0},
		{1, 0},
		{-1, 1},
		{0, 1},
		{1, 1},
	}

	for _, d := range dirs {
		nx, ny := node.X+d[0], node.Y+d[1]
		if n := g.GetNode(nx, ny); n != nil {
			neighbors = append(neighbors, n)
		}
	}

	return neighbors
}

// heuristic calculates the estimated distance to the goal (octile distance).
func heuristic(a, b *PathNode) float64 {
	dx := float64(abs(a.X - b.X))
	dy := float64(abs(a.Y - b.Y))

	return dx + dy + (1.414-2)*minFloat(dx, dy)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}

	return x
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}

	return b
}

func reconstructPath(end *PathNode) []Point {
	var path []Point

	current := end
	for current != nil {
		path = append([]Point{{X: current.X, Y: current.Y}}, path...)
		current = current.Parent
	}

	return path
}

// nodeHeap implements heap.Interface for A* open set.
type nodeHeap []*PathNode

func (h nodeHeap) Len() int           { return len(h) }
func (h nodeHeap) Less(i, j int) bool { return h[i].F < h[j].F }
func (h nodeHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].heapIndex = i
	h[j].heapIndex = j
}

func (h *nodeHeap) Push(x any) {
	n, ok := x.(*PathNode)
	if !ok {
		return
	}

	n.heapIndex = len(*h)
	*h = append(*h, n)
}

func (h *nodeHeap) Pop() any {
	old := *h
	n := len(old)
	node := old[n-1]
	node.heapIndex = -1
	*h = old[0 : n-1]

	return node
}
