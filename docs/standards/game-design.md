# Game Design Standards - Underneath TD

## Display Settings

### Resolution Options
The game supports 5 resolution presets with 1080p as the default:

| Resolution | Name | Notes |
|------------|------|-------|
| 1024×576 | 16:9 | Smallest supported |
| 1280×720 | 720p | Good for smaller screens |
| 1366×768 | Standard | Common laptop resolution |
| 1600×900 | 900p | Mid-range |
| 1920×1080 | 1080p | **Default**, optimal experience |

### Fullscreen Mode
- Toggle available via settings menu
- Keyboard shortcut: Left/Right arrows or F key
- State persists in settings

### Settings Navigation
Settings screen supports both keyboard and mouse:
- **Keyboard**: Up/Down to select row, Left/Right to change value
- **Mouse**: Click < > buttons next to each setting

---

## Map System

### Tile Size
- Tiles are **48×48 pixels**
- Provides good balance between detail and map visibility

### Map Sizes
Each map has a unique size for gameplay variety:

| Map | Grid Size | Total Tiles |
|-----|-----------|-------------|
| Forest Path | 20×18 | 360 tiles |
| Castle Siege | 25×20 | 500 tiles |
| Fire Fortress | 30×25 | 750 tiles |
| Ice Cavern | 35×30 | 1,050 tiles |
| Void Realm | 40×30 | 1,200 tiles |

> Target range: **350 to 1,200 tiles** per map

### Camera System
Large maps require scrollable viewport:
- **WASD / Arrow Keys**: Pan camera
- **Mouse Edge Scrolling**: Move cursor to screen edge
- Camera automatically clamps to map boundaries
- Camera speed: 400 pixels/second

---

## Hero System

### Hero Deployment
During deploy phase:
1. Heroes appear in sidebar **HEROES** section (undeployed only)
2. Drag hero from sidebar to map
3. Drop on valid **grass tiles** (green highlight = valid, red = invalid)
4. Heroes disappear from sidebar once placed
5. Heroes persist through wave phase

### Valid Placement
Heroes can be placed on:
- Grass tiles (`TileGrass`)
- Tiles marked as buildable (`CanBuild = true`)
- Not on path, stone, or water tiles
- Not overlapping other heroes

### Starter Skill Gems
Each hero class receives a default skill gem on creation:

| Class | Starter Gem | Tags |
|-------|-------------|------|
| Marauder | Heavy Strike | attack, melee, physical |
| Ranger | Split Arrow | attack, bow, projectile |
| Witch | Fireball | spell, fire, projectile, aoe |
| Duelist | Cleave | attack, physical, aoe |
| Templar | Arc | spell, lightning, chain |
| Shadow | Poison Arrow | attack, chaos, projectile, dot |
| Scion | Ice Nova | spell, cold, aoe |

---

## UI/UX Standards

### Sidebar Layout
The right sidebar (200px wide) contains sections in order:
1. **CONTROL** - Wave, Gold, Core HP stats
2. **DEPLOYABLES** - Tower cards for placement
3. **HEROES** - Undeployed heroes for placement
4. **SELECTED** - Info panel for clicked unit

### Color Conventions
- **Green**: Valid placement, success states
- **Red**: Invalid placement, damage, enemies
- **Gold/Yellow**: Currency, rewards, highlights
- **Cyan/Blue**: Core, ES (energy shield), mana

### Mouse Responsiveness
- Custom cursor uses `ebiten.CursorPosition()` directly for zero-lag rendering
- Cached `mouse.X/Y` values used for game logic only

---

## Technical Notes

### Hero Data Storage
- Heroes are stored in `g.Player.Heroes` (player profile)
- **NOT** in `g.Heroes` (game session)
- All hero loops must use `g.Player.Heroes`

### Camera-Aware Rendering
All map elements must apply camera offset:
```go
screenX := worldX - g.CameraX
screenY := worldY - g.CameraY
```

### Camera-Aware Input
Mouse clicks on map must convert to world coordinates:
```go
worldX := float64(screenX) + g.CameraX
worldY := float64(screenY) + g.CameraY
```
