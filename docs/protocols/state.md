# State Inspection Protocol

Standard format for engine-to-AI world snapshots.

---

## Full JSON Format

```json
{
  "tick": 42,
  "entities": [
    {
      "id": 1,
      "type": "player",
      "description": "Hero character",
      "position": [100, 200],
      "velocity": [5, 0],
      "state": "walking",
      "health": [80, 100],
      "tags": ["controllable"]
    }
  ],
  "summary": "Tick 42: 1 player(s), 3 enemy(s)"
}
```

---

## Compact Format

For minimal token usage (~90% savings):

```
T{tick}|{entities}
```

Entity format: `{type_code}:{x},{y}[;{extras}]`

| Code | Type |
|------|------|
| P | player |
| E | enemy |
| I | item |
| X | projectile |
| U | ui |
| F | effect |

Example: `T42|P:100,200;H:80/100|E:50,75`

---

## Implementation

```go
exporter := ai.NewStateExporter(&world)
json := exporter.ExportJSON(&world, tick)

compact := ai.NewCompactExporter(&world)
state := compact.Export(&world, tick)
```
