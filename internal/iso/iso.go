package iso

import (
	"math"
)

// CartesianToIso converts grid coordinates to screen coordinates.
// Returns the CENTER of the tile on screen.
// tileW, tileH: Dimensions of the tile in pixels.
func CartesianToIso(gridX, gridY, tileW, tileH int) (screenX, screenY float64) {
	// For isometric projection:
	// screenX = (gridX - gridY) * (tileWidth / 2)
	// screenY = (gridX + gridY) * (tileHeight / 2)
	screenX = float64(gridX-gridY) * float64(tileW/2)
	screenY = float64(gridX+gridY) * float64(tileH/2)

	return screenX, screenY
}

// CartesianToIsoWithOffset converts grid coordinates to screen coordinates with map offset.
func CartesianToIsoWithOffset(
	gridX, gridY, tileW, tileH int,
	offsetX, offsetY float64,
) (screenX, screenY float64) {
	sx, sy := CartesianToIso(gridX, gridY, tileW, tileH)

	return sx + offsetX, sy + offsetY
}

// IsoToCartesian converts screen coordinates to grid coordinates.
func IsoToCartesian(screenX, screenY float64, tileW, tileH int, offsetX, offsetY float64) (gridX, gridY int) {
	// Remove offset first
	sx := screenX - offsetX
	sy := screenY - offsetY

	// Inverse of the isometric transformation:
	// gridX = (screenX / (tileWidth/2) + screenY / (tileHeight/2)) / 2
	// gridY = (screenY / (tileHeight/2) - screenX / (tileWidth/2)) / 2

	halfWidth := float64(tileW / 2)
	halfHeight := float64(tileH / 2)

	gridXf := (sx/halfWidth + sy/halfHeight) / 2
	gridYf := (sy/halfHeight - sx/halfWidth) / 2

	// Round to nearest integer
	gridX = int(gridXf)
	gridY = int(gridYf)

	// Adjust for negative coordinates
	if gridXf < 0 {
		gridX--
	}

	if gridYf < 0 {
		gridY--
	}

	return gridX, gridY
}

// GetTileDiamondPoints returns the 4 corner points of an isometric tile diamond.
func GetTileDiamondPoints(
	gridX, gridY, tileW, tileH int,
	offsetX, offsetY float64,
) (top, right, bottom, left [2]float64) {
	centerX, centerY := CartesianToIsoWithOffset(gridX, gridY, tileW, tileH, offsetX, offsetY)

	halfW := float64(tileW / 2)
	halfH := float64(tileH / 2)

	top = [2]float64{centerX, centerY - halfH}
	right = [2]float64{centerX + halfW, centerY}
	bottom = [2]float64{centerX, centerY + halfH}
	left = [2]float64{centerX - halfW, centerY}

	return top, right, bottom, left
}

// IsPointInIsoDiamond checks if a screen point is inside an isometric tile.
func IsPointInIsoDiamond(px, py float64, gridX, gridY, tileW, tileH int, offsetX, offsetY float64) bool {
	centerX, centerY := CartesianToIsoWithOffset(gridX, gridY, tileW, tileH, offsetX, offsetY)

	halfW := float64(tileW / 2)
	halfH := float64(tileH / 2)

	// Check using the diamond inequality: |dx/halfW| + |dy/halfH| <= 1
	dx := math.Abs(px - centerX)
	dy := math.Abs(py - centerY)

	return (dx/halfW + dy/halfH) <= 1.0
}
