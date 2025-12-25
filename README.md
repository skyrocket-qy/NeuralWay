# AI-Generation Game Framework

A Go-based ECS (Entity-Component-System) game framework designed for AI-assisted game development with **full multi-platform support**, while AI can 100% joined, it also can use by developer easily.
Each package is component-based, you can choose to use or not.

## Goals

1. **Code-based** — AI can 100% participate, no GUI tools required
2. **Multi-platform** — Windows, Linux, macOS, Android, iOS, Web (WASM)
3. **Scalable performance** — From mobile phones to high-end PCs
4. **ECS Architecture** — Clean separation of data and logic
5. **AI-based project evolution** — Self evolve project by AI

---

## Quick Start

### Prerequisites

- Go 1.21+
- For mobile builds: `gomobile` (run `make mobile-init`)
- For Windows cross-compile: MinGW (`x86_64-w64-mingw32-gcc`)
- For Android: Android SDK + NDK
- For iOS: macOS with Xcode

### Run an Example

```bash
# Run the survivor game
make run-survivor

# Run other examples
make run-snake
make run-pong
make run-breakout
```

---

## Multi-Platform Building

### Build an Example for Any Platform

Use the universal build script to build any example game for any platform:

```bash
# Build for current OS
./scripts/build-example.sh survivor

# Build for specific platforms
./scripts/build-example.sh survivor wasm      # Web browser
./scripts/build-example.sh survivor windows   # Windows .exe
./scripts/build-example.sh survivor linux     # Linux binary
./scripts/build-example.sh survivor darwin    # macOS binary
./scripts/build-example.sh survivor android   # Android APK
./scripts/build-example.sh survivor ios       # iOS framework (macOS only)

# Build for ALL platforms
./scripts/build-example.sh survivor all
```

Output files are placed in `dist/<example-name>/<platform>/`.

### Platform-Specific Details

#### Desktop (Windows, Linux, macOS)

```bash
# Using Makefile
make build-windows   # Requires MinGW on Linux/macOS
make build-linux
make build-darwin

# Using script
./scripts/build-example.sh mygame windows
./scripts/build-example.sh mygame linux
./scripts/build-example.sh mygame darwin
```

**Windows cross-compilation from Linux/macOS:**
```bash
# Install MinGW
# Ubuntu/Debian:
sudo apt-get install mingw-w64
# macOS:
brew install mingw-w64
```

#### WebAssembly (Browser)

```bash
# Build
./scripts/build-example.sh survivor wasm

# Serve locally
cd dist/survivor/wasm
python3 -m http.server 8080
# Open http://localhost:8080
```

The WASM build creates:
- `<game>.wasm` — The compiled game
- `wasm_exec.js` — Go runtime for WASM
- `index.html` — Ready-to-use web page

#### Android

```bash
# Initialize gomobile (one-time setup)
make mobile-init

# Build APK
./scripts/build-example.sh survivor android
# Output: dist/survivor/android/survivor.apk
```

**Requirements:**
- Android SDK + NDK
- `ANDROID_HOME` environment variable set
- gomobile (`go install golang.org/x/mobile/cmd/gomobile@latest`)

#### iOS

```bash
# macOS only
./scripts/build-example.sh survivor ios
# Output: dist/survivor/ios/survivor.xcframework
```

**Requirements:**
- macOS with Xcode installed
- Valid Apple Developer account (for device deployment)

---

## Project Structure

```
├── cmd/game/           # Main game entry point
├── examples/           # Example games (snake, pong, survivor, etc.)
│   └── survivor/
│       ├── main.go     # Game code
│       ├── assets/     # Game assets (sprites, sounds)
│       └── web/        # Pre-built web version
├── internal/
│   ├── components/     # ECS Components (Position, Velocity, Sprite, etc.)
│   ├── systems/        # ECS Systems (Render, Movement, Collision, etc.)
│   ├── engine/         # Core engine functionality
│   └── assets/         # Asset loading utilities
├── scripts/
│   ├── build-example.sh    # Multi-platform build script
│   ├── build-mobile.sh     # Mobile build helper
│   └── wasm/               # WASM templates
└── Makefile            # Build commands
```

---

## Developing a New Game

### 1. Create Your Game Directory

```bash
mkdir -p examples/mygame/assets
touch examples/mygame/main.go
```

### 2. Basic Game Template

```go
package main

import (
    "log"
    "github.com/hajimehoshi/ebiten/v2"
)

const (
    screenWidth  = 800
    screenHeight = 600
)

type Game struct {
    // Your game state here
}

func NewGame() *Game {
    return &Game{}
}

func (g *Game) Update() error {
    // Handle input and update game logic (60 FPS)
    return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
    // Render your game
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
    return screenWidth, screenHeight
}

func main() {
    ebiten.SetWindowSize(screenWidth, screenHeight)
    ebiten.SetWindowTitle("My Game")
    
    if err := ebiten.RunGame(NewGame()); err != nil {
        log.Fatal(err)
    }
}
```

### 3. Add to Makefile

```makefile
run-mygame:
	go run ./examples/mygame
```

### 4. Build for All Platforms

```bash
./scripts/build-example.sh mygame all
```

---

## Working with Sprites

### Standard Chroma Key Background Removal

When creating sprite assets, use **Pure Magenta (#FF00FF)** as the background color for automatic transparency removal. This follows the 16-bit game development convention:

| Color Name   | Hex (8-bit) | Usage |
|-------------|-------------|-------|
| Pure Magenta | #FF00FF    | Most common, rarely appears in sprites |
| Bright Green | #00FF00    | Alternative for purple/pink sprites |
| Pure Black   | #000000    | Use with caution |

The framework's `removeBackground()` function automatically detects and removes these key colors.

### Loading Sprites with Transparency

```go
import (
    "embed"
    "image"
    "bytes"
    _ "image/png"
    
    "github.com/hajimehoshi/ebiten/v2"
)

//go:embed assets/*.png
var assets embed.FS

func loadSprite(filename string) *ebiten.Image {
    data, _ := assets.ReadFile(filename)
    img, _, _ := image.Decode(bytes.NewReader(data))
    img = removeBackground(img)  // Remove magenta background
    return ebiten.NewImageFromImage(img)
}
```

---

## Using ECS Components

```go
import (
    "td/internal/components"
    "github.com/mlange-42/ark/ecs"
)

// Create a world
world := ecs.NewWorld()

// Create entities with components
entity := world.NewEntity()
world.Add(entity, 
    components.Position{X: 100, Y: 100},
    components.Velocity{X: 1, Y: 0},
    components.NewSprite(myImage),
)
```

### Available Components

| Component    | Description                          |
|-------------|--------------------------------------|
| `Position`  | 2D position (X, Y float64)           |
| `Velocity`  | 2D velocity (X, Y float64)           |
| `Sprite`    | Renderable image with scale/offset   |
| `Collider`  | Bounding box with collision layers   |
| `Health`    | Current/Max health                   |
| `Tag`       | Entity categorization marker         |

---

## Example Games

| Game          | Description          | Run Command              | Platforms |
|---------------|----------------------|--------------------------|-----------|
| Survivor      | Vampire Survivors-style | `make run-survivor`    | All |
| Snake         | Classic snake        | `make run-snake`         | All |
| Pong          | Two-player pong      | `make run-pong`          | All |
| Breakout      | Brick breaker        | `make run-breakout`      | All |
| Flappy        | Flappy bird clone    | `make run-flappy`        | All |
| 2048          | Puzzle 2048          | `make run-2048`          | All |
| Minesweeper   | Classic minesweeper  | `make run-minesweeper`   | All |
| Roguelike     | Dungeon crawler      | `make run-roguelike`     | All |

---

## Build Commands Reference

### Development

```bash
make run              # Run main game
make run-<example>    # Run specific example
make test             # Run tests
make lint             # Run linter
make clean            # Clean build artifacts
```

### Desktop Builds

```bash
make build              # Current platform
make build-darwin       # macOS
make build-windows      # Windows (requires MinGW)
make build-linux        # Linux
make build-all-desktop  # All desktop platforms
```

### WebAssembly

```bash
make build-wasm         # Standard Go WASM
make build-wasm-tiny    # TinyGo (smaller binary)
make serve-wasm         # Serve locally
```

### Mobile

```bash
make mobile-init        # Initialize gomobile (run once)
make build-android      # Android AAR library
make build-android-apk  # Android APK
make build-ios          # iOS framework
```

### Release Package

```bash
make dist VERSION=1.0.0  # Creates all platform packages
```

---

## Tips for AI-Assisted Development

1. **Keep games in `examples/`** — Each game is self-contained
2. **Use the ECS pattern** — Separate data (components) from logic (systems)
3. **Test frequently** — Use `go run ./examples/yourname`
4. **Use magenta backgrounds** — For sprite transparency (#FF00FF)
5. **Check multi-platform builds** — Test WASM before deploying
6. **Embed assets** — Use `//go:embed` for portable builds

---

## Troubleshooting

### WASM not loading

Ensure you're serving via HTTP (not file://):
```bash
cd dist/mygame/wasm
python3 -m http.server 8080
```

### MinGW not found (Windows cross-compile)

```bash
# Ubuntu/Debian
sudo apt-get install mingw-w64

# macOS
brew install mingw-w64
```

### Android build fails

1. Ensure `ANDROID_HOME` is set
2. Run `make mobile-init` first
3. Install Android NDK

### iOS build fails

- Requires macOS with Xcode
- Run `xcode-select --install` if needed