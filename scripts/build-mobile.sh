#!/bin/bash
# Mobile Build Script for Android and iOS
# Requires: gomobile, Android SDK (for Android), Xcode (for iOS)

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
DIST_DIR="$PROJECT_ROOT/dist"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_gomobile() {
    if ! command -v gomobile &> /dev/null; then
        log_error "gomobile not found. Installing..."
        go install golang.org/x/mobile/cmd/gomobile@latest
        gomobile init
    fi
}

build_android_aar() {
    log_info "Building Android AAR..."
    mkdir -p "$DIST_DIR/android"
    
    cd "$PROJECT_ROOT"
    gomobile bind -target=android -androidapi 21 -o "$DIST_DIR/android/game.aar" ./cmd/game
    
    log_info "Android AAR built: $DIST_DIR/android/game.aar"
}

build_android_apk() {
    log_info "Building Android APK..."
    mkdir -p "$DIST_DIR/android"
    
    cd "$PROJECT_ROOT"
    gomobile build -target=android -androidapi 21 -o "$DIST_DIR/android/game.apk" ./cmd/game
    
    log_info "Android APK built: $DIST_DIR/android/game.apk"
}

build_ios() {
    log_info "Building iOS XCFramework..."
    
    if [[ "$OSTYPE" != "darwin"* ]]; then
        log_error "iOS builds require macOS"
        exit 1
    fi
    
    mkdir -p "$DIST_DIR/ios"
    
    cd "$PROJECT_ROOT"
    gomobile bind -target=ios -o "$DIST_DIR/ios/Game.xcframework" ./cmd/game
    
    log_info "iOS framework built: $DIST_DIR/ios/Game.xcframework"
}

show_usage() {
    echo "Mobile Build Script"
    echo ""
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  android-aar   Build Android AAR library"
    echo "  android-apk   Build Android APK"
    echo "  ios           Build iOS XCFramework"
    echo "  all           Build all mobile targets"
    echo "  init          Initialize gomobile"
    echo ""
}

case "$1" in
    android-aar)
        check_gomobile
        build_android_aar
        ;;
    android-apk)
        check_gomobile
        build_android_apk
        ;;
    ios)
        check_gomobile
        build_ios
        ;;
    all)
        check_gomobile
        build_android_aar
        if [[ "$OSTYPE" == "darwin"* ]]; then
            build_ios
        else
            log_warn "Skipping iOS build (requires macOS)"
        fi
        ;;
    init)
        go install golang.org/x/mobile/cmd/gomobile@latest
        gomobile init
        log_info "gomobile initialized"
        ;;
    *)
        show_usage
        ;;
esac
