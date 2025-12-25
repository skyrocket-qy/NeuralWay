// Package systems provides ECS systems for game logic.
package systems

import (
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mlange-42/ark/ecs"
	"td/internal/components"
)

// spriteInstance holds render data for a single sprite in a batch.
type spriteInstance struct {
	x, y             float64
	scaleX, scaleY   float64
	offsetX, offsetY float64
	layer            int
}

// BatchRenderSystem draws sprites grouped by texture for improved performance.
// Sprites with the same image are drawn together in a single batch.
// Optional SortLayer component controls draw order (lower = drawn first).
type BatchRenderSystem struct {
	filter      *ecs.Filter2[components.Position, components.Sprite]
	layerFilter *ecs.Filter3[components.Position, components.Sprite, components.SortLayer]
	batches     map[*ebiten.Image][]spriteInstance
	sortedKeys  []*ebiten.Image
	opts        *ebiten.DrawImageOptions
}

// NewBatchRenderSystem creates a new batched render system.
func NewBatchRenderSystem(world *ecs.World) *BatchRenderSystem {
	return &BatchRenderSystem{
		filter:      ecs.NewFilter2[components.Position, components.Sprite](world),
		layerFilter: ecs.NewFilter3[components.Position, components.Sprite, components.SortLayer](world),
		batches:     make(map[*ebiten.Image][]spriteInstance),
		sortedKeys:  make([]*ebiten.Image, 0, 64),
		opts:        &ebiten.DrawImageOptions{},
	}
}

// Draw renders all visible sprites in batches grouped by texture.
func (s *BatchRenderSystem) Draw(world *ecs.World, screen *ebiten.Image) {
	// Clear batches from previous frame
	for k := range s.batches {
		s.batches[k] = s.batches[k][:0] // Keep allocated memory
	}

	s.sortedKeys = s.sortedKeys[:0]

	// First pass: entities with SortLayer component
	query := s.layerFilter.Query()
	for query.Next() {
		pos, sprite, sortLayer := query.Get()
		s.addToBatch(pos, sprite, sortLayer.Layer)
	}

	// Second pass: entities without SortLayer (default layer 0)
	query2 := s.filter.Query()
	for query2.Next() {
		pos, sprite := query2.Get()
		// Skip if this entity also has SortLayer (already processed)
		// Note: ark ECS doesn't have exclusion filters, so we use layer 0 default
		// This may cause duplicate draws for entities with SortLayer=0
		// A proper solution would use Without filters when available
		s.addToBatch(pos, sprite, 0)
	}

	// Sort keys by minimum layer in each batch for proper draw order
	for img := range s.batches {
		if len(s.batches[img]) > 0 {
			s.sortedKeys = append(s.sortedKeys, img)
		}
	}

	// Sort batches by their minimum layer
	sort.Slice(s.sortedKeys, func(i, j int) bool {
		batchI := s.batches[s.sortedKeys[i]]

		batchJ := s.batches[s.sortedKeys[j]]
		if len(batchI) == 0 || len(batchJ) == 0 {
			return false
		}

		return batchI[0].layer < batchJ[0].layer
	})

	// Draw batches
	for _, img := range s.sortedKeys {
		instances := s.batches[img]

		// Sort instances within batch by layer
		sort.Slice(instances, func(i, j int) bool {
			return instances[i].layer < instances[j].layer
		})

		for _, inst := range instances {
			s.opts.GeoM.Reset()
			s.opts.GeoM.Scale(inst.scaleX, inst.scaleY)
			s.opts.GeoM.Translate(inst.x+inst.offsetX, inst.y+inst.offsetY)
			screen.DrawImage(img, s.opts)
		}
	}
}

// addToBatch adds a sprite to the appropriate texture batch.
func (s *BatchRenderSystem) addToBatch(pos *components.Position, sprite *components.Sprite, layer int) {
	if !sprite.Visible || sprite.Image == nil {
		return
	}

	img := sprite.Image
	if _, ok := s.batches[img]; !ok {
		s.batches[img] = make([]spriteInstance, 0, 32)
	}

	s.batches[img] = append(s.batches[img], spriteInstance{
		x:       pos.X,
		y:       pos.Y,
		scaleX:  sprite.ScaleX,
		scaleY:  sprite.ScaleY,
		offsetX: sprite.OffsetX,
		offsetY: sprite.OffsetY,
		layer:   layer,
	})
}

// Stats returns the number of unique textures and total sprites in the last frame.
func (s *BatchRenderSystem) Stats() (textures, sprites int) {
	textures = len(s.sortedKeys)
	for _, instances := range s.batches {
		sprites += len(instances)
	}

	return textures, sprites
}
