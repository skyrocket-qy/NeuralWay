# Action Space Protocol

Unified verbs for AI agent game control.

---

## Action Interface

```go
type Action interface {
    Execute(world *ecs.World) error
    Name() string
}
```

---

## Built-in Actions

| Action | Parameters | Description |
|--------|------------|-------------|
| `MoveAction` | Entity, DX, DY | Move by delta |
| `SetPositionAction` | Entity, X, Y | Set absolute position |
| `SetVelocityAction` | Entity, VX, VY | Set velocity |
| `RemoveEntityAction` | Entity | Remove from world |
| `SetStateAction` | Entity, State | Update visual state |

---

## Execution

```go
executor := ai.NewActionExecutor(&world)

// Single action
executor.Execute(ai.MoveAction{Entity: e, DX: 10, DY: 0})

// Batch actions
errors := executor.ExecuteBatch([]ai.Action{
    ai.SetVelocityAction{Entity: e, VX: 5, VY: 0},
    ai.SetStateAction{Entity: e, State: "running"},
})
```

---

## Custom Actions

```go
type MyAction struct {
    Target ecs.Entity
}

func (a MyAction) Name() string { return "my_action" }
func (a MyAction) Execute(world *ecs.World) error {
    // Custom logic
    return nil
}
```
