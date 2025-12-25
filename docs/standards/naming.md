# Naming Protocol

Standard names for animations and assets to assist LLM generation.

---

## Animation States

| State | Description |
|-------|-------------|
| `idle` | Default standing/inactive |
| `walk` | Walking movement |
| `run` | Running movement |
| `attack` | Attack animation |
| `hurt` | Taking damage |
| `death` | Death animation |
| `jump` | Jump animation |
| `fall` | Falling animation |

## Direction Suffixes

| Suffix | Direction |
|--------|-----------|
| `_up` | Facing up |
| `_down` | Facing down |
| `_left` | Facing left |
| `_right` | Facing right |

## Frame Numbering

Use 2-digit zero-padded: `_01`, `_02`, `_03`

---

## Examples

```
player_idle_01.png
player_walk_right_01.png
player_walk_right_02.png
enemy_attack_down_01.png
```

## Sprite Sheets

```
player_spritesheet.png
player_spritesheet.json  # Frame data
```
