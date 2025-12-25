#!/bin/bash
# =============================================================================
# Build Example Script - Build any example game for multiple platforms
# =============================================================================
#
# Usage: ./scripts/build-example.sh <example-name> [platform]
#
# Platforms:
#   desktop     - Build for current OS (default)
#   windows     - Build for Windows (.exe)
#   linux       - Build for Linux
#   darwin      - Build for macOS
#   wasm        - Build for WebAssembly
#   android     - Build Android APK
#   ios         - Build iOS framework (macOS only)
#   all         - Build all platforms
#
# Examples:
#   ./scripts/build-example.sh survivor
#   ./scripts/build-example.sh survivor wasm
#   ./scripts/build-example.sh snake all
#
# =============================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
EXAMPLE_NAME="${1:-}"
PLATFORM="${2:-desktop}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_step() { echo -e "${CYAN}[STEP]${NC} $1"; }

show_usage() {
    echo "Build Example Script - Multi-platform build for example games"
    echo ""
    echo "Usage: $0 <example-name> [platform]"
    echo ""
    echo "Available examples:"
    for dir in "$PROJECT_ROOT/examples"/*/; do
        if [ -f "${dir}main.go" ]; then
            echo "  - $(basename "$dir")"
        fi
    done
    echo ""
    echo "Platforms:"
    echo "  desktop   - Current OS (default)"
    echo "  windows   - Windows .exe"
    echo "  linux     - Linux binary"
    echo "  darwin    - macOS binary"
    echo "  wasm      - WebAssembly (browser)"
    echo "  android   - Android APK"
    echo "  ios       - iOS framework (macOS only)"
    echo "  all       - All platforms"
    echo ""
    echo "Example:"
    echo "  $0 survivor wasm"
}

check_example() {
    local example_dir="$PROJECT_ROOT/examples/$EXAMPLE_NAME"
    if [ ! -d "$example_dir" ] || [ ! -f "$example_dir/main.go" ]; then
        log_error "Example '$EXAMPLE_NAME' not found"
        echo ""
        show_usage
        exit 1
    fi
}

build_desktop() {
    local os="${1:-$(go env GOOS)}"
    local arch="${2:-$(go env GOARCH)}"
    local ext=""
    [[ "$os" == "windows" ]] && ext=".exe"
    
    local output_dir="$PROJECT_ROOT/dist/$EXAMPLE_NAME/$os-$arch"
    mkdir -p "$output_dir"
    
    log_step "Building $EXAMPLE_NAME for $os/$arch..."
    
    if [[ "$os" == "windows" ]] && [[ "$(go env GOOS)" != "windows" ]]; then
        CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc \
            GOOS=$os GOARCH=$arch \
            go build -o "$output_dir/$EXAMPLE_NAME$ext" "./examples/$EXAMPLE_NAME"
    else
        GOOS=$os GOARCH=$arch \
            go build -o "$output_dir/$EXAMPLE_NAME$ext" "./examples/$EXAMPLE_NAME"
    fi
    
    # Copy assets if they exist
    if [ -d "$PROJECT_ROOT/examples/$EXAMPLE_NAME/assets" ]; then
        cp -r "$PROJECT_ROOT/examples/$EXAMPLE_NAME/assets" "$output_dir/"
    fi
    
    log_info "Built: $output_dir/$EXAMPLE_NAME$ext"
}

build_wasm() {
    local output_dir="$PROJECT_ROOT/dist/$EXAMPLE_NAME/wasm"
    mkdir -p "$output_dir"
    
    log_step "Building $EXAMPLE_NAME for WebAssembly..."
    
    GOOS=js GOARCH=wasm go build -o "$output_dir/$EXAMPLE_NAME.wasm" "./examples/$EXAMPLE_NAME"
    
    # Copy assets if they exist
    if [ -d "$PROJECT_ROOT/examples/$EXAMPLE_NAME/assets" ]; then
        cp -r "$PROJECT_ROOT/examples/$EXAMPLE_NAME/assets" "$output_dir/"
    fi
    
    # Copy wasm_exec.js
    if [ -f "$(go env GOROOT)/lib/wasm/wasm_exec.js" ]; then
        cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" "$output_dir/"
    elif [ -f "$(go env GOROOT)/misc/wasm/wasm_exec.js" ]; then
        cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" "$output_dir/"
    fi
    
    # Create index.html
    cat > "$output_dir/index.html" << 'HTMLEOF'
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>GAME_TITLE</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
            min-height: 100vh;
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            font-family: 'Segoe UI', system-ui, sans-serif;
            color: #e8e8e8;
        }
        h1 {
            font-size: 2rem;
            margin-bottom: 1rem;
            background: linear-gradient(90deg, #00d4ff, #7b2ff7);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
        }
        #game-container {
            border-radius: 12px;
            overflow: hidden;
            box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
        }
        #loading { padding: 2rem; text-align: center; }
        .spinner {
            width: 50px; height: 50px;
            border: 4px solid rgba(255, 255, 255, 0.2);
            border-top-color: #7b2ff7;
            border-radius: 50%;
            animation: spin 1s linear infinite;
            margin: 0 auto 1rem;
        }
        @keyframes spin { to { transform: rotate(360deg); } }
        .controls {
            margin-top: 1.5rem;
            padding: 1rem 2rem;
            background: rgba(255, 255, 255, 0.05);
            border-radius: 8px;
        }
        kbd {
            background: rgba(255, 255, 255, 0.1);
            padding: 0.2rem 0.5rem;
            border-radius: 4px;
            font-family: monospace;
        }
    </style>
</head>
<body>
    <h1>🎮 GAME_TITLE</h1>
    <div id="game-container">
        <div id="loading">
            <div class="spinner"></div>
            <p>Loading game...</p>
        </div>
    </div>
    <div class="controls">
        <p><kbd>WASD</kbd> / <kbd>Arrow Keys</kbd> - Move | <kbd>ESC</kbd> - Pause</p>
    </div>
    <script src="wasm_exec.js"></script>
    <script>
        const go = new Go();
        WebAssembly.instantiateStreaming(fetch("WASM_FILE"), go.importObject)
            .then(result => {
                document.getElementById("loading").style.display = "none";
                go.run(result.instance);
            })
            .catch(err => {
                document.getElementById("loading").innerHTML = 
                    '<p style="color: #ff6b6b;">Error: ' + err.message + '</p>';
            });
    </script>
</body>
</html>
HTMLEOF

    # Replace placeholders
    # Replace placeholders
    if [[ "$OSTYPE" == "darwin"* ]]; then
        sed -i '' "s/GAME_TITLE/$EXAMPLE_NAME/g" "$output_dir/index.html"
        sed -i '' "s/WASM_FILE/$EXAMPLE_NAME.wasm/g" "$output_dir/index.html"
    else
        sed -i "s/GAME_TITLE/$EXAMPLE_NAME/g" "$output_dir/index.html"
        sed -i "s/WASM_FILE/$EXAMPLE_NAME.wasm/g" "$output_dir/index.html"
    fi
    
    log_info "Built: $output_dir/"
    log_info "To run: cd $output_dir && python3 -m http.server 8080"
}

build_android() {
    local output_dir="$PROJECT_ROOT/dist/$EXAMPLE_NAME/android"
    mkdir -p "$output_dir"
    
    if ! command -v gomobile &> /dev/null; then
        log_warn "gomobile not found. Installing..."
        go install golang.org/x/mobile/cmd/gomobile@latest
        gomobile init
    fi
    
    log_step "Building $EXAMPLE_NAME for Android..."
    gomobile build -target=android -androidapi 21 -o "$output_dir/$EXAMPLE_NAME.apk" "./examples/$EXAMPLE_NAME"
    
    log_info "Built: $output_dir/$EXAMPLE_NAME.apk"
}

build_ios() {
    if [[ "$OSTYPE" != "darwin"* ]]; then
        log_error "iOS builds require macOS"
        return 1
    fi
    
    local output_dir="$PROJECT_ROOT/dist/$EXAMPLE_NAME/ios"
    mkdir -p "$output_dir"
    
    if ! command -v gomobile &> /dev/null; then
        log_warn "gomobile not found. Installing..."
        go install golang.org/x/mobile/cmd/gomobile@latest
        gomobile init
    fi
    
    log_step "Building $EXAMPLE_NAME for iOS..."
    gomobile bind -target=ios -o "$output_dir/$EXAMPLE_NAME.xcframework" "./examples/$EXAMPLE_NAME"
    
    log_info "Built: $output_dir/$EXAMPLE_NAME.xcframework"
}

# Main
if [ -z "$EXAMPLE_NAME" ]; then
    show_usage
    exit 1
fi

check_example
cd "$PROJECT_ROOT"

case "$PLATFORM" in
    desktop)
        build_desktop
        ;;
    windows)
        build_desktop windows amd64
        ;;
    linux)
        build_desktop linux amd64
        ;;
    darwin)
        build_desktop darwin amd64
        ;;
    wasm|web)
        build_wasm
        ;;
    android)
        build_android
        ;;
    ios)
        build_ios
        ;;
    all)
        log_info "Building $EXAMPLE_NAME for all platforms..."
        build_desktop linux amd64
        build_desktop darwin amd64
        build_desktop windows amd64 || log_warn "Windows build failed (requires MinGW)"
        build_wasm
        build_android || log_warn "Android build failed (requires gomobile + Android SDK)"
        if [[ "$OSTYPE" == "darwin"* ]]; then
            build_ios || log_warn "iOS build failed"
        fi
        log_info "All builds complete!"
        ;;
    *)
        log_error "Unknown platform: $PLATFORM"
        show_usage
        exit 1
        ;;
esac

echo ""
log_info "Build complete! Output in dist/$EXAMPLE_NAME/"
