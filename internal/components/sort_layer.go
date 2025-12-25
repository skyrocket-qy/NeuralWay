package components

// SortLayer provides z-ordering for rendering.
// Lower Layer values are drawn first (background), higher values drawn last (foreground).
type SortLayer struct {
	Layer int
}
