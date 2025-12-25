# Agent Development Guidelines for Underneath

## Development Workflow

> [!IMPORTANT]
> **After EVERY code edit**, you MUST rebuild to WASM and test in browser.

### Build and Test Cycle

```bash
# 1. Build WASM (from project root)
cd /Users/zelin/project/td/examples/underneath && ./build_wasm.sh

# 2. Start web server
cd /Users/zelin/project/td/examples/underneath/web && python3 -m http.server 8080

# 3. Open browser and test
# Navigate to http://localhost:8080
```

### Testing Requirements

After each change, verify:
1. **Build succeeds** - No compilation errors
2. **Game loads** - Main menu appears
3. **Feature works** - Test the specific change you made
4. **No regressions** - Basic gameplay still functions

### Test Documentation

After testing, update:
- `docs/testing/test_plan.md` - Mark new tests as passed
- `docs/testing/test_walkthrough.md` - Document findings with screenshots

---

## Quick Reference

| Action | Command |
|--------|---------|
| Build Desktop | `go build ./examples/underneath` |
| Build WASM | `cd examples/underneath && ./build_wasm.sh` |
| Run Desktop | `go run ./examples/underneath` |
| Run Web | `cd examples/underneath/web && python3 -m http.server 8080` |

---

## Code Standards

1. **File headers** - Add comment explaining file purpose
2. **Input handling** - Put in `update*` functions, NOT `draw*` functions
3. **Error handling** - Graceful fallbacks, no crashes
4. **Sound effects** - Use `PlaySound()` for user feedback

---

## Project Structure

```
examples/underneath/
├── main.go          # Entry point
├── game.go          # Game state machine
├── logic.go         # Update logic
├── ui.go            # Drawing utilities
├── hub.go           # Management Hub UI
├── management.go    # Equipment/Gem UI
├── audio.go         # Sound system
├── assets.go        # Asset loading
├── docs/            # Documentation
│   ├── design/      # Architecture docs
│   └── testing/     # Test plans
└── web/             # WASM files
```
