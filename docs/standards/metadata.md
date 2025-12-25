# Metadata Protocol

JSON sidecar files for embedding game logic hints into assets.

---

## File Format

For asset `sprite.png`, create `sprite.json`:

```json
{
  "type": "character",
  "tags": ["player", "animated"],
  "origin": [16, 32],
  "hitbox": { "x": 4, "y": 16, "w": 24, "h": 32 },
  "animations": {
    "idle": { "frames": [0, 1], "fps": 4, "loop": true },
    "walk": { "frames": [2, 3, 4, 5], "fps": 8, "loop": true }
  }
}
```

---

## Fields

| Field | Type | Description |
|-------|------|-------------|
| `type` | string | `character`, `item`, `tile`, `effect` |
| `tags` | string[] | Searchable labels |
| `origin` | [x,y] | Pivot point for rotation |
| `hitbox` | object | Collision rectangle |
| `animations` | object | Animation definitions |

## Animation Object

| Field | Type | Description |
|-------|------|-------------|
| `frames` | int[] | Frame indices |
| `fps` | int | Frames per second |
| `loop` | bool | Loop animation |
