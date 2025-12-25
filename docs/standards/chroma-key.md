# Chroma Key Standard

Transparency key colors for sprite assets.

---

## Primary Key
**Pure Magenta**: `#FF00FF` / `rgb(255, 0, 255)`

Use for all sprite transparency.

## Secondary Key
**Bright Green**: `#00FF00` / `rgb(0, 255, 0)`

Use when magenta conflicts with sprite colors.

---

## Guidelines

1. Use **exactly** these RGB values (no anti-aliasing on edges)
2. Fill transparent areas completely
3. Avoid using key colors in actual sprite art
4. For animated sprites, use same key across all frames
