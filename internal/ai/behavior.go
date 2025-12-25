package ai

import "github.com/mlange-42/ark/ecs"

// NodeStatus represents the result of a behavior tree node tick.
type NodeStatus int

const (
	// Running means the node is still executing.
	Running NodeStatus = iota
	// Success means the node completed successfully.
	Success
	// Failure means the node failed.
	Failure
)

// BTContext provides context for behavior tree execution.
type BTContext struct {
	World  *ecs.World
	Entity ecs.Entity
	Data   map[string]any // Blackboard for sharing data
}

// NewBTContext creates a new behavior tree context.
func NewBTContext(world *ecs.World, entity ecs.Entity) *BTContext {
	return &BTContext{
		World:  world,
		Entity: entity,
		Data:   make(map[string]any),
	}
}

// Node is the interface for all behavior tree nodes.
type Node interface {
	Tick(ctx *BTContext) NodeStatus
}

// ============================================================================
// Composite Nodes
// ============================================================================

// Sequence runs children in order until one fails.
// Returns Success if all succeed, Failure if any fails.
type Sequence struct {
	Children []Node
	current  int
}

// NewSequence creates a sequence node.
func NewSequence(children ...Node) *Sequence {
	return &Sequence{Children: children}
}

func (n *Sequence) Tick(ctx *BTContext) NodeStatus {
	for n.current < len(n.Children) {
		status := n.Children[n.current].Tick(ctx)
		if status == Running {
			return Running
		}

		if status == Failure {
			n.current = 0 // Reset for next tick

			return Failure
		}

		n.current++
	}

	n.current = 0

	return Success
}

// Selector runs children until one succeeds.
// Returns Success if any succeeds, Failure if all fail.
type Selector struct {
	Children []Node
	current  int
}

// NewSelector creates a selector node.
func NewSelector(children ...Node) *Selector {
	return &Selector{Children: children}
}

func (n *Selector) Tick(ctx *BTContext) NodeStatus {
	for n.current < len(n.Children) {
		status := n.Children[n.current].Tick(ctx)
		if status == Running {
			return Running
		}

		if status == Success {
			n.current = 0

			return Success
		}

		n.current++
	}

	n.current = 0

	return Failure
}

// ============================================================================
// Decorator Nodes
// ============================================================================

// Inverter inverts the result of its child.
type Inverter struct {
	Child Node
}

// NewInverter creates an inverter node.
func NewInverter(child Node) *Inverter {
	return &Inverter{Child: child}
}

func (n *Inverter) Tick(ctx *BTContext) NodeStatus {
	status := n.Child.Tick(ctx)
	if status == Success {
		return Failure
	}

	if status == Failure {
		return Success
	}

	return Running
}

// Repeater repeats its child a set number of times.
type Repeater struct {
	Child Node
	Times int
	count int
}

// NewRepeater creates a repeater node.
func NewRepeater(child Node, times int) *Repeater {
	return &Repeater{Child: child, Times: times}
}

func (n *Repeater) Tick(ctx *BTContext) NodeStatus {
	if n.count >= n.Times {
		n.count = 0

		return Success
	}

	status := n.Child.Tick(ctx)
	if status == Success {
		n.count++
		if n.count >= n.Times {
			n.count = 0

			return Success
		}

		return Running
	}

	if status == Failure {
		n.count = 0

		return Failure
	}

	return Running
}

// ============================================================================
// Leaf Nodes
// ============================================================================

// Condition checks a predicate.
type Condition struct {
	Check func(ctx *BTContext) bool
}

// NewCondition creates a condition node.
func NewCondition(check func(ctx *BTContext) bool) *Condition {
	return &Condition{Check: check}
}

func (n *Condition) Tick(ctx *BTContext) NodeStatus {
	if n.Check(ctx) {
		return Success
	}

	return Failure
}

// ActionNode executes a function.
type ActionNode struct {
	Do func(ctx *BTContext) NodeStatus
}

// NewActionNode creates an action node.
func NewActionNode(do func(ctx *BTContext) NodeStatus) *ActionNode {
	return &ActionNode{Do: do}
}

func (n *ActionNode) Tick(ctx *BTContext) NodeStatus {
	return n.Do(ctx)
}

// SucceedNode always returns Success.
type SucceedNode struct{}

func (n *SucceedNode) Tick(ctx *BTContext) NodeStatus {
	return Success
}

// FailNode always returns Failure.
type FailNode struct{}

func (n *FailNode) Tick(ctx *BTContext) NodeStatus {
	return Failure
}

// ============================================================================
// Behavior Tree Runner
// ============================================================================

// BehaviorTree wraps a root node for execution.
type BehaviorTree struct {
	Root Node
}

// NewBehaviorTree creates a behavior tree with the given root.
func NewBehaviorTree(root Node) *BehaviorTree {
	return &BehaviorTree{Root: root}
}

// Tick runs one tick of the behavior tree.
func (bt *BehaviorTree) Tick(ctx *BTContext) NodeStatus {
	return bt.Root.Tick(ctx)
}
