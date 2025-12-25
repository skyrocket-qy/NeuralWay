description: Validate TD Web Build
---
# Validate TD Web Build

1. Build the game for WASM to a 'game.wasm' file
2. Create an index.html that loads it
3. Run a python server
4. I will launch a browser to verify

// turbo
GOOS=js GOARCH=wasm go build -o game.wasm ./examples/td

// turbo
cp $(go env GOROOT)/misc/wasm/wasm_exec.js .

echo '<!DOCTYPE html>
<script src="wasm_exec.js"></script>
<script>
    const go = new Go();
    WebAssembly.instantiateStreaming(fetch("game.wasm"), go.importObject).then((result) => {
        go.run(result.instance);
    });
</script>' > index.html

// turbo
python3 -m http.server 8080 &
