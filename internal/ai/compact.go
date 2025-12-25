package ai

import (
	"fmt"
	"strings"

	"github.com/mlange-42/ark/ecs"
	"td/internal/components"
)

// CompactExporter exports world state in minimal-token format for LLMs.
// Format: "T{tick}|{entities}"
// Entity format: "{type}:{x},{y}[;{extras}]"
// Type codes: P=player, E=enemy, I=item, X=projectile, U=ui, F=effect.
type CompactExporter struct {
	metaFilter *ecs.Filter1[components.AIMetadata]
}

// NewCompactExporter creates a compact exporter.
func NewCompactExporter(world *ecs.World) *CompactExporter {
	return &CompactExporter{
		metaFilter: ecs.NewFilter1[components.AIMetadata](world),
	}
}

// typeCode converts entity type to single-char code.
func typeCode(entityType string) string {
	switch entityType {
	case "player":
		return "P"
	case "enemy":
		return "E"
	case "item":
		return "I"
	case "projectile":
		return "X"
	case "ui":
		return "U"
	case "effect":
		return "F"
	case "npc":
		return "N"
	default:
		return "?"
	}
}

// Export returns compact state string.
// Format: "T42|P:100,200|E:50,75;H:80/100|I:30,40".
func (e *CompactExporter) Export(world *ecs.World, tick int64) string {
	var parts []string

	parts = append(parts, fmt.Sprintf("T%d", tick))

	posMap := ecs.NewMap[components.Position](world)
	healthMap := ecs.NewMap[components.Health](world)
	velMap := ecs.NewMap[components.Velocity](world)

	query := e.metaFilter.Query()
	for query.Next() {
		entity := query.Entity()
		meta := query.Get()

		var sb strings.Builder
		sb.WriteString(typeCode(meta.EntityType))

		// Position
		if posMap.Has(entity) {
			pos := posMap.Get(entity)
			sb.WriteString(fmt.Sprintf(":%.0f,%.0f", pos.X, pos.Y))
		}

		// Optional extras
		extras := make([]string, 0)

		// Health as H:current/max
		if healthMap.Has(entity) {
			h := healthMap.Get(entity)
			extras = append(extras, fmt.Sprintf("H:%d/%d", h.Current, h.Max))
		}

		// Velocity as V:x,y (only if moving)
		if velMap.Has(entity) {
			v := velMap.Get(entity)
			if v.X != 0 || v.Y != 0 {
				extras = append(extras, fmt.Sprintf("V:%.0f,%.0f", v.X, v.Y))
			}
		}

		// State as S:state (only if not idle)
		if meta.VisualState != "" && meta.VisualState != "idle" {
			extras = append(extras, "S:"+meta.VisualState)
		}

		if len(extras) > 0 {
			sb.WriteString(";")
			sb.WriteString(strings.Join(extras, ";"))
		}

		parts = append(parts, sb.String())
	}

	return strings.Join(parts, "|")
}

// ExportDelta returns only entities that changed since last tick.
// changedEntities is a set of entity IDs that have changes.
func (e *CompactExporter) ExportDelta(
	world *ecs.World,
	tick int64,
	changedEntities map[ecs.Entity]bool,
) string {
	var parts []string

	parts = append(parts, fmt.Sprintf("D%d", tick)) // D for delta

	posMap := ecs.NewMap[components.Position](world)
	healthMap := ecs.NewMap[components.Health](world)

	query := e.metaFilter.Query()
	for query.Next() {
		entity := query.Entity()
		if !changedEntities[entity] {
			continue
		}

		meta := query.Get()

		var sb strings.Builder
		sb.WriteString(typeCode(meta.EntityType))

		if posMap.Has(entity) {
			pos := posMap.Get(entity)
			sb.WriteString(fmt.Sprintf(":%.0f,%.0f", pos.X, pos.Y))
		}

		if healthMap.Has(entity) {
			h := healthMap.Get(entity)
			sb.WriteString(fmt.Sprintf(";H:%d/%d", h.Current, h.Max))
		}

		parts = append(parts, sb.String())
	}

	return strings.Join(parts, "|")
}

// TokenEstimate estimates token count for a compact string.
// Assumes ~4 chars per token on average.
func TokenEstimate(compact string) int {
	return (len(compact) + 3) / 4
}
