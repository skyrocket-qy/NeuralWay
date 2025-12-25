# Alpha Blending Standard

Guidelines for using alpha channel transparency in assets.

---

## Format
**PNG-32**: 24-bit Color + 8-bit Alpha

## Usage
- Use for modern, high-resolution assets.
- Essential for:
  - Anti-aliased edges
  - Drop shadows
  - Semi-transparent effects (magic, UI)
  - Smoothed particle textures

## Storage
- Store textures with **Straight Alpha** (unassociated alpha).
- Avoid pre-multiplied alpha in source files to prevent double-multiplication artifacts.

## Rendering
- Engine should handle Alpha Blending `(SrcAlpha, OneMinusSrcAlpha)`
- Or Pre-multiplied Alpha `(One, OneMinusSrcAlpha)` if textures are converted at load time.

## Legacy / Pixel Art
For pixel-perfect retro style where partial transparency is undesirable, use **Chroma Key** instead (see `chroma-key.md`).
