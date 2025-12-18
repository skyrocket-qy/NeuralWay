bk:
	git add .
	git commit -m update
	git push

# =============================================================================
# Development
# =============================================================================

run:
	go run ./cmd/game

test:
	go test ./...

lint:
	go vet ./...

clean:
	rm -rf bin/ dist/

# =============================================================================
# Desktop Builds
# =============================================================================

build: build-darwin
	@echo "Built for current platform"

build-darwin:
	@mkdir -p bin
	go build -o bin/game-darwin ./cmd/game

build-windows:
	@mkdir -p bin
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -o bin/game.exe ./cmd/game

build-linux:
	@mkdir -p bin
	GOOS=linux GOARCH=amd64 go build -o bin/game-linux ./cmd/game

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
	cd bin && zip ../dist/releases/game-$(VERSION)-darwin.zip game-darwin
	cd bin && zip ../dist/releases/game-$(VERSION)-windows.zip game.exe
	cd bin && zip ../dist/releases/game-$(VERSION)-linux.zip game-linux
	cd dist/wasm && zip ../releases/game-$(VERSION)-wasm.zip *
	@echo "Release packages created in dist/releases/"

# =============================================================================
# Help
# =============================================================================

help:
	@echo "AI-Generation ECS Game Framework - Build Commands"
	@echo ""
	@echo "Development:"
	@echo "  make run             - Run the game"
	@echo "  make test            - Run tests"
	@echo "  make lint            - Run linter"
	@echo ""
	@echo "Desktop Builds:"
	@echo "  make build           - Build for current platform"
	@echo "  make build-darwin    - Build for macOS"
	@echo "  make build-windows   - Build for Windows (requires mingw)"
	@echo "  make build-linux     - Build for Linux"
	@echo ""
	@echo "WebAssembly:"
	@echo "  make build-wasm      - Build with standard Go"
	@echo "  make build-wasm-tiny - Build with TinyGo (smaller)"
	@echo "  make serve-wasm      - Serve WASM build locally"
	@echo ""
	@echo "Mobile:"
	@echo "  make mobile-init     - Initialize gomobile"
	@echo "  make build-android   - Build Android AAR"
	@echo "  make build-ios       - Build iOS framework"
	@echo ""
	@echo "Release:"
	@echo "  make dist VERSION=1.0.0 - Create release packages"

.PHONY: run test lint clean build build-darwin build-windows build-linux \
        build-wasm build-wasm-tiny serve-wasm mobile-init build-android \
        build-android-apk build-ios dist help bk