package graphics

import (
	"image"
	"image/color"
	"image/draw"
)

// RemoveBackground removes the background color from an image using Chroma Keying.
// Supports traditional Magenta (#FF00FF) and Green (#00FF00) keys.
// Also performs a safety check to preserve existing transparency (PNG-32).
func RemoveBackground(src image.Image) image.Image {
	bounds := src.Bounds()

	// 1. Safety Check: If top-left pixel is already transparent, assume PNG-32 and skip
	c := src.At(bounds.Min.X, bounds.Min.Y)

	//nolint:dogsled // We only need alpha here to check transparency
	_, _, _, a := c.RGBA()
	if a < 1000 {
		return src
	}

	// 2. Use Top-Left pixel as Key Color
	keyR, keyG, keyB, _ := c.RGBA()

	const keyThreshold = 8000 // Tolerance (approx 12% or 30/255)

	// 3. Apply Chroma Key Removal
	dst := image.NewRGBA(bounds)
	draw.Draw(dst, bounds, src, bounds.Min, draw.Src)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			original := dst.At(x, y)
			r, g, b, _ := original.RGBA()

			// Check if color matches key within threshold
			if intAbs(int(r)-int(keyR)) < int(keyThreshold) &&
				intAbs(int(g)-int(keyG)) < int(keyThreshold) &&
				intAbs(int(b)-int(keyB)) < int(keyThreshold) {
				dst.Set(x, y, color.Transparent)
			}
		}
	}

	return dst
}

func intAbs(x int) int {
	if x < 0 {
		return -x
	}

	return x
}
