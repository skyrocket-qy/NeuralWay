# Actions Reference

AI agent action interface for controlling game entities.

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

### MoveAction
Move entity by delta.
```go
ai.MoveAction{Entity: e, DX: 10, DY: 0}
```

### SetPositionAction
Set absolute position.
```go
ai.SetPositionAction{Entity: e, X: 100, Y: 200}
```

### SetVelocityAction
Set entity velocity.
```go
ai.SetVelocityAction{Entity: e, VX: 5, VY: -3}
```

### RemoveEntityAction
Remove entity from world.
```go
ai.RemoveEntityAction{Entity: e}
```

### SetStateAction
Update AIMetadata visual state.
```go
ai.SetStateAction{Entity: e, State: "attacking"}
```

---

## ActionExecutor

```go
executor := ai.NewActionExecutor(&world)

// Single action
executor.Execute(action)

// Batch actions
errors := executor.ExecuteBatch([]Action{action1, action2})
```

---

## State Export

### Full Export
```go
exporter := ai.NewStateExporter(&world)
json := exporter.ExportJSON(&world, tick)
markdown := exporter.ExportMarkdown(&world, tick)
```

### Compact Export (90% fewer tokens)
```go
compact := ai.NewCompactExporter(&world)
state := compact.Export(&world, tick)
// "T42|P:100,200|E:50,75;H:80/100"
```
