bk:
	git add .
	git commit -m update
	git push

# =============================================================================
# Development
# =============================================================================

run:
	go run ./cmd/game

run-snake:
	go run ./examples/snake

run-pong:
	go run ./examples/pong

run-breakout:
	go run ./examples/breakout

run-slots:
	go run ./examples/slots

run-flappy:
	go run ./examples/flappy

run-2048:
	go run ./examples/puzzle_2048

run-minesweeper:
	go run ./examples/minesweeper

run-cookie:
	go run ./examples/cookie_clicker

run-shooter:
	go run ./examples/space_shooter

run-platformer:
	go run ./examples/platformer

run-match3:
	go run ./examples/match3

run-blackjack:
	go run ./examples/blackjack

run-agar:
	go run ./examples/agar

run-rts:
	go run ./examples/mini_rts

run-rpg:
	go run ./examples/rpg_battle

run-roguelike:
	go run ./examples/roguelike

run-survivor:
	go run ./examples/survivor

# =============================================================================
# Example Game Builds (use scripts/build-example.sh for more options)
# =============================================================================

# Build survivor for all platforms
build-survivor:
	./scripts/build-example.sh survivor desktop

build-survivor-wasm:
	./scripts/build-example.sh survivor wasm

build-survivor-all:
	./scripts/build-example.sh survivor all

# Generic example build (usage: make build-example EXAMPLE=survivor PLATFORM=wasm)
EXAMPLE ?= survivor
PLATFORM ?= desktop

build-example:
	./scripts/build-example.sh $(EXAMPLE) $(PLATFORM)

# Serve any example as WASM
serve-example:
	cd dist/$(EXAMPLE)/wasm && python3 -m http.server 8080

test:
	go test ./...

lint:
	go vet ./...

clean:
	rm -rf dist/

# =============================================================================
# Desktop Builds
# =============================================================================

build: build-darwin
	@echo "Built for current platform"

build-darwin:
	@mkdir -p dist/bin
	go build -o dist/bin/game-darwin ./cmd/game

build-windows:
	@mkdir -p dist/bin
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -o dist/bin/game.exe ./cmd/game

build-linux:
	@mkdir -p dist/bin
	GOOS=linux GOARCH=amd64 go build -o dist/bin/game-linux ./cmd/game

build-all-desktop: build-darwin build-windows build-linux
	@echo "Built for all desktop platforms"

# =============================================================================
# WebAssembly Builds
# =============================================================================

# Standard Go WASM (larger binary, full compatibility)
build-wasm:
	@mkdir -p dist/wasm
	GOOS=js GOARCH=wasm go build -o dist/wasm/game.wasm ./cmd/game
	@# Go 1.25+ uses lib/wasm, older uses misc/wasm
	@if [ -f "$$(go env GOROOT)/lib/wasm/wasm_exec.js" ]; then \
		cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" dist/wasm/; \
	else \
		cp "$$(go env GOROOT)/misc/wasm/wasm_exec.js" dist/wasm/; \
	fi
	cp scripts/wasm/index.html dist/wasm/
	@echo "WASM build complete. Serve dist/wasm/ with a web server."

# TinyGo WASM (smaller binary, some limitations)
build-wasm-tiny:
	@mkdir -p dist/wasm-tiny
	tinygo build -o dist/wasm-tiny/game.wasm -target wasm ./cmd/game
	cp scripts/wasm/wasm_exec_tiny.js dist/wasm-tiny/wasm_exec.js
	cp scripts/wasm/index.html dist/wasm-tiny/
	@echo "TinyGo WASM build complete. Serve dist/wasm-tiny/ with a web server."

# Serve WASM build locally
serve-wasm:
	cd dist/wasm && python3 -m http.server 8080

# =============================================================================
# Mobile Builds (requires gomobile)
# =============================================================================

# Initialize gomobile (run once)
mobile-init:
	go install golang.org/x/mobile/cmd/gomobile@latest
	gomobile init

# Android AAR library
build-android:
	@mkdir -p dist/android
	gomobile bind -target=android -androidapi 21 -o dist/android/game.aar ./cmd/game
	@echo "Android AAR built at dist/android/game.aar"

# Android APK (requires Android SDK)
build-android-apk:
	@mkdir -p dist/android
	gomobile build -target=android -androidapi 21 -o dist/android/game.apk ./cmd/game
	@echo "Android APK built at dist/android/game.apk"

# iOS framework
build-ios:
	@mkdir -p dist/ios
	gomobile bind -target=ios -o dist/ios/Game.xcframework ./cmd/game
	@echo "iOS framework built at dist/ios/Game.xcframework"

# =============================================================================
# Release Packaging
# =============================================================================

VERSION ?= dev

dist: build-all-desktop build-wasm
	@mkdir -p dist/releases
	cd dist/bin && zip ../releases/game-$(VERSION)-darwin.zip game-darwin
	cd dist/bin && zip ../releases/game-$(VERSION)-windows.zip game.exe
	cd dist/bin && zip ../releases/game-$(VERSION)-linux.zip game-linux
	cd dist/wasm && zip ../releases/game-$(VERSION)-wasm.zip *
	@echo "Release packages created in dist/releases/"

# =============================================================================
# Help
# =============================================================================

help:
	@echo "AI-Generation ECS Game Framework - Build Commands"
	@echo ""
	@echo "Development:"
	@echo "  make run             - Run the main game"
	@echo "  make run-<example>   - Run specific example (e.g., run-survivor)"
	@echo "  make test            - Run tests"
	@echo "  make lint            - Run linter"
	@echo ""
	@echo "Example Builds (Multi-Platform):"
	@echo "  make build-example EXAMPLE=survivor PLATFORM=wasm"
	@echo "  make build-survivor-wasm    - Build survivor for web"
	@echo "  make build-survivor-all     - Build survivor for all platforms"
	@echo "  make serve-example EXAMPLE=survivor  - Serve WASM locally"
	@echo ""
	@echo "  Or use the script directly:"
	@echo "    ./scripts/build-example.sh survivor wasm"
	@echo ""
	@echo "Desktop Builds:"
	@echo "  make build           - Build for current platform"
	@echo "  make build-darwin    - Build for macOS"
	@echo "  make build-windows   - Build for Windows (requires mingw)"
	@echo "  make build-linux     - Build for Linux"
	@echo ""
	@echo "WebAssembly:"
	@echo "  make build-wasm      - Build main game with standard Go"
	@echo "  make build-wasm-tiny - Build with TinyGo (smaller)"
	@echo "  make serve-wasm      - Serve WASM build locally"
	@echo ""
	@echo "Mobile:"
	@echo "  make mobile-init     - Initialize gomobile (run once)"
	@echo "  make build-android   - Build Android AAR"
	@echo "  make build-android-apk - Build Android APK"
	@echo "  make build-ios       - Build iOS framework (macOS only)"
	@echo ""
	@echo "Release:"
	@echo "  make dist VERSION=1.0.0 - Create release packages"

.PHONY: run test lint clean build build-darwin build-windows build-linux \
        build-wasm build-wasm-tiny serve-wasm mobile-init build-android \
        build-android-apk build-ios dist help bk \
        build-survivor build-survivor-wasm build-survivor-all \
        build-example serve-example

golint:
	golangci-lint run ./... 