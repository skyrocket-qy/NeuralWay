---
description: Build and Validate Dev Survivor (WASM)
---

# Validate Dev Survivor

This workflow builds the game to WASM, serves it locally, and runs an automated browser test.

1. **Build WASM**
   ```bash
   cd examples/survivor
   GOOS=js GOARCH=wasm go build -o web/survivor.wasm main.go
   ```

2. **Serve**
   ```bash
   cd examples/survivor/web
   python3 -m http.server 8082
   ```

3. **Validate** (Manual or Agent)
   - Open `http://localhost:8082`
   - Press SPACE to start.
   - Use WASD to move.
