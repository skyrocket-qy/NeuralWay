package systems

import (
	"container/heap"
	"math"

	"github.com/mlange-42/ark/ecs"
	"td/internal/components"
)

// PathNode represents a node in pathfinding.
type PathNode struct {
	X, Y    int
	G, H, F float64 // G=cost from start, H=heuristic, F=G+H
	Parent  *PathNode
	index   int // For heap
}

// PathHeap implements heap.Interface for A*.
type PathHeap []*PathNode

func (h PathHeap) Len() int           { return len(h) }
func (h PathHeap) Less(i, j int) bool { return h[i].F < h[j].F }
func (h PathHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *PathHeap) Push(x any) {
	n, ok := x.(*PathNode)
	if !ok {
		return
	}

	n.index = len(*h)
	*h = append(*h, n)
}

func (h *PathHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[0 : n-1]

	return item
}

// NavGrid represents a navigation grid for pathfinding.
type NavGrid struct {
	Width, Height int
	CellSize      float64
	Walkable      [][]bool
	Costs         [][]float64 // Movement cost per cell (1.0 = normal)
}

// NewNavGrid creates a navigation grid.
func NewNavGrid(width, height int, cellSize float64) *NavGrid {
	walkable := make([][]bool, height)

	costs := make([][]float64, height)
	for y := range walkable {
		walkable[y] = make([]bool, width)

		costs[y] = make([]float64, width)
		for x := range walkable[y] {
			walkable[y][x] = true
			costs[y][x] = 1.0
		}
	}

	return &NavGrid{
		Width:    width,
		Height:   height,
		CellSize: cellSize,
		Walkable: walkable,
		Costs:    costs,
	}
}

// SetWalkable sets whether a cell is walkable.
func (g *NavGrid) SetWalkable(x, y int, walkable bool) {
	if x >= 0 && x < g.Width && y >= 0 && y < g.Height {
		g.Walkable[y][x] = walkable
	}
}

// SetCost sets the movement cost for a cell.
func (g *NavGrid) SetCost(x, y int, cost float64) {
	if x >= 0 && x < g.Width && y >= 0 && y < g.Height {
		g.Costs[y][x] = cost
	}
}

// IsWalkable returns true if a cell is walkable.
func (g *NavGrid) IsWalkable(x, y int) bool {
	if x < 0 || x >= g.Width || y < 0 || y >= g.Height {
		return false
	}

	return g.Walkable[y][x]
}

// WorldToGrid converts world coordinates to grid coordinates.
func (g *NavGrid) WorldToGrid(wx, wy float64) (int, int) {
	return int(wx / g.CellSize), int(wy / g.CellSize)
}

// GridToWorld converts grid coordinates to world coordinates (center of cell).
func (g *NavGrid) GridToWorld(gx, gy int) (float64, float64) {
	return (float64(gx) + 0.5) * g.CellSize, (float64(gy) + 0.5) * g.CellSize
}

// PathfindingSystem provides A* pathfinding.
type PathfindingSystem struct {
	Grid     *NavGrid
	MaxNodes int // Maximum nodes to explore (prevents infinite loops)
}

// NewPathfindingSystem creates a pathfinding system.
func NewPathfindingSystem(grid *NavGrid) *PathfindingSystem {
	return &PathfindingSystem{
		Grid:     grid,
		MaxNodes: 1000,
	}
}

// Path represents a found path.
type Path struct {
	Points [][2]float64 // World coordinates
	Valid  bool
	Length float64
}

// FindPath finds a path from start to end using A*.
func (p *PathfindingSystem) FindPath(startX, startY, endX, endY float64) *Path {
	grid := p.Grid

	// Convert to grid coords
	sx, sy := grid.WorldToGrid(startX, startY)
	ex, ey := grid.WorldToGrid(endX, endY)

	// Check endpoints
	if !grid.IsWalkable(sx, sy) || !grid.IsWalkable(ex, ey) {
		return &Path{Valid: false}
	}

	// Already at destination
	if sx == ex && sy == ey {
		wx, wy := grid.GridToWorld(ex, ey)

		return &Path{
			Points: [][2]float64{{wx, wy}},
			Valid:  true,
		}
	}

	// A* algorithm
	openSet := &PathHeap{}
	heap.Init(openSet)

	closedSet := make(map[int64]bool)
	nodeMap := make(map[int64]*PathNode)

	startNode := &PathNode{X: sx, Y: sy, G: 0, H: heuristic(sx, sy, ex, ey)}
	startNode.F = startNode.G + startNode.H
	heap.Push(openSet, startNode)
	nodeMap[coordKey(sx, sy)] = startNode

	nodesExplored := 0

	for openSet.Len() > 0 && nodesExplored < p.MaxNodes {
		nodesExplored++

		item := heap.Pop(openSet)

		current, ok := item.(*PathNode)
		if !ok {
			continue
		}

		// Found the goal
		if current.X == ex && current.Y == ey {
			return p.reconstructPath(current)
		}

		closedSet[coordKey(current.X, current.Y)] = true

		// Check neighbors (8 directions)
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				if dx == 0 && dy == 0 {
					continue
				}

				nx, ny := current.X+dx, current.Y+dy

				if !grid.IsWalkable(nx, ny) {
					continue
				}

				// Prevent diagonal movement through walls
				if dx != 0 && dy != 0 {
					if !grid.IsWalkable(current.X+dx, current.Y) ||
						!grid.IsWalkable(current.X, current.Y+dy) {
						continue
					}
				}

				key := coordKey(nx, ny)
				if closedSet[key] {
					continue
				}

				// Calculate cost (diagonal = 1.414)
				moveCost := 1.0
				if dx != 0 && dy != 0 {
					moveCost = 1.414
				}

				moveCost *= grid.Costs[ny][nx]

				tentativeG := current.G + moveCost

				neighbor, exists := nodeMap[key]
				if !exists {
					neighbor = &PathNode{X: nx, Y: ny}
					neighbor.H = heuristic(nx, ny, ex, ey)
					neighbor.G = tentativeG
					neighbor.F = neighbor.G + neighbor.H
					neighbor.Parent = current
					heap.Push(openSet, neighbor)
					nodeMap[key] = neighbor
				} else if tentativeG < neighbor.G {
					neighbor.G = tentativeG
					neighbor.F = neighbor.G + neighbor.H
					neighbor.Parent = current
					heap.Fix(openSet, neighbor.index)
				}
			}
		}
	}

	return &Path{Valid: false}
}

// reconstructPath builds the path from goal to start.
func (p *PathfindingSystem) reconstructPath(goal *PathNode) *Path {
	path := &Path{Valid: true}

	// Collect nodes from goal to start
	nodes := make([]*PathNode, 0)

	current := goal
	for current != nil {
		nodes = append(nodes, current)
		current = current.Parent
	}

	// Reverse to get start-to-goal order
	path.Points = make([][2]float64, len(nodes))
	for i := 0; i < len(nodes); i++ {
		node := nodes[len(nodes)-1-i]
		wx, wy := p.Grid.GridToWorld(node.X, node.Y)
		path.Points[i] = [2]float64{wx, wy}
	}

	// Calculate length
	for i := 1; i < len(path.Points); i++ {
		dx := path.Points[i][0] - path.Points[i-1][0]
		dy := path.Points[i][1] - path.Points[i-1][1]
		path.Length += math.Sqrt(dx*dx + dy*dy)
	}

	return path
}

func heuristic(x1, y1, x2, y2 int) float64 {
	dx := float64(x2 - x1)
	dy := float64(y2 - y1)
	// Euclidean distance
	return math.Sqrt(dx*dx + dy*dy)
}

func coordKey(x, y int) int64 {
	return int64(x)<<32 | int64(uint32(y))
}

// SmoothPath reduces nodes in a path using line-of-sight.
func (p *PathfindingSystem) SmoothPath(path *Path) *Path {
	if !path.Valid || len(path.Points) < 3 {
		return path
	}

	smoothed := &Path{Valid: true}
	smoothed.Points = append(smoothed.Points, path.Points[0])

	current := 0
	for current < len(path.Points)-1 {
		// Find furthest visible point
		furthest := current + 1
		for next := current + 2; next < len(path.Points); next++ {
			if p.hasLineOfSight(path.Points[current][0], path.Points[current][1],
				path.Points[next][0], path.Points[next][1]) {
				furthest = next
			}
		}

		smoothed.Points = append(smoothed.Points, path.Points[furthest])
		current = furthest
	}

	return smoothed
}

// hasLineOfSight checks if there's clear line between two world points.
func (p *PathfindingSystem) hasLineOfSight(x1, y1, x2, y2 float64) bool {
	// Bresenham-like line check
	gx1, gy1 := p.Grid.WorldToGrid(x1, y1)
	gx2, gy2 := p.Grid.WorldToGrid(x2, y2)

	dx := abs(gx2 - gx1)
	dy := abs(gy2 - gy1)

	sx := 1
	if gx1 > gx2 {
		sx = -1
	}

	sy := 1
	if gy1 > gy2 {
		sy = -1
	}

	err := dx - dy

	for {
		if !p.Grid.IsWalkable(gx1, gy1) {
			return false
		}

		if gx1 == gx2 && gy1 == gy2 {
			break
		}

		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			gx1 += sx
		}

		if e2 < dx {
			err += dx
			gy1 += sy
		}
	}

	return true
}

func abs(x int) int {
	if x < 0 {
		return -x
	}

	return x
}

// PathFollower helps an entity follow a path.
type PathFollower struct {
	Path         *Path
	CurrentIndex int
	Speed        float64
	Threshold    float64 // Distance to consider waypoint reached
}

// NewPathFollower creates a path follower.
func NewPathFollower(speed float64) *PathFollower {
	return &PathFollower{
		Speed:     speed,
		Threshold: 5.0,
	}
}

// SetPath sets a new path to follow.
func (pf *PathFollower) SetPath(path *Path) {
	pf.Path = path
	pf.CurrentIndex = 0
}

// GetNextMove returns direction to move toward next waypoint.
func (pf *PathFollower) GetNextMove(currentX, currentY float64) (dx, dy float64, reachedEnd bool) {
	if pf.Path == nil || !pf.Path.Valid || pf.CurrentIndex >= len(pf.Path.Points) {
		return 0, 0, true
	}

	target := pf.Path.Points[pf.CurrentIndex]
	dx = target[0] - currentX
	dy = target[1] - currentY
	dist := math.Sqrt(dx*dx + dy*dy)

	if dist <= pf.Threshold {
		pf.CurrentIndex++

		return pf.GetNextMove(currentX, currentY)
	}

	// Normalize and scale by speed
	dx = (dx / dist) * pf.Speed
	dy = (dy / dist) * pf.Speed

	return dx, dy, false
}

// IsFinished returns true if path is complete.
func (pf *PathFollower) IsFinished() bool {
	return pf.Path == nil || pf.CurrentIndex >= len(pf.Path.Points)
}

// Navigation component for entities.
type Navigation struct {
	TargetX, TargetY float64
	Path             *Path
	CurrentWaypoint  int
	Speed            float64
	RecalcInterval   float64 // How often to recalculate path
	RecalcTimer      float64
	Stopped          bool
}

// NavigationSystem manages entity pathfinding.
type NavigationSystem struct {
	Pathfinding *PathfindingSystem
	navFilter   *ecs.Filter2[components.Position, Navigation]
}

// NewNavigationSystem creates a navigation system.
func NewNavigationSystem(world *ecs.World, pathfinding *PathfindingSystem) *NavigationSystem {
	return &NavigationSystem{
		Pathfinding: pathfinding,
		navFilter:   ecs.NewFilter2[components.Position, Navigation](world),
	}
}

// Update updates all navigating entities.
func (s *NavigationSystem) Update(world *ecs.World, dt float64) {
	query := s.navFilter.Query()
	for query.Next() {
		pos, nav := query.Get()

		if nav.Stopped {
			continue
		}

		// Recalculate path if needed
		nav.RecalcTimer -= dt
		if nav.Path == nil || nav.RecalcTimer <= 0 {
			nav.Path = s.Pathfinding.FindPath(pos.X, pos.Y, nav.TargetX, nav.TargetY)
			nav.CurrentWaypoint = 0
			nav.RecalcTimer = nav.RecalcInterval
		}

		if nav.Path == nil || !nav.Path.Valid || nav.CurrentWaypoint >= len(nav.Path.Points) {
			continue
		}

		// Move toward current waypoint
		target := nav.Path.Points[nav.CurrentWaypoint]
		dx := target[0] - pos.X
		dy := target[1] - pos.Y
		dist := math.Sqrt(dx*dx + dy*dy)

		if dist < 5.0 {
			nav.CurrentWaypoint++

			continue
		}

		// Move
		moveX := (dx / dist) * nav.Speed * dt
		moveY := (dy / dist) * nav.Speed * dt
		pos.X += moveX
		pos.Y += moveY
	}
}
