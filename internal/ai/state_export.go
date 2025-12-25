// Package ai provides AI-native infrastructure for LLM integration.
package ai

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mlange-42/ark/ecs"
	"td/internal/components"
)

// WorldSnapshot represents a serializable snapshot of the world state.
type WorldSnapshot struct {
	Tick     int64            `json:"tick"`
	Entities []EntitySnapshot `json:"entities"`
	Summary  string           `json:"summary,omitempty"`
}

// EntitySnapshot represents a serialized entity with its key components.
type EntitySnapshot struct {
	ID          uint32      `json:"id"`
	EntityType  string      `json:"type"`
	Description string      `json:"description,omitempty"`
	Position    *[2]float64 `json:"position,omitempty"`
	Velocity    *[2]float64 `json:"velocity,omitempty"`
	State       string      `json:"state,omitempty"`
	Tags        []string    `json:"tags,omitempty"`
	Effects     []string    `json:"effects,omitempty"`
	Health      *[2]int     `json:"health,omitempty"` // [current, max]
}

// StateExporter exports ECS world state for LLM consumption.
type StateExporter struct {
	metaFilter   *ecs.Filter1[components.AIMetadata]
	posFilter    *ecs.Filter1[components.Position]
	healthFilter *ecs.Filter1[components.Health]
}

// NewStateExporter creates a new state exporter for the given world.
func NewStateExporter(world *ecs.World) *StateExporter {
	return &StateExporter{
		metaFilter:   ecs.NewFilter1[components.AIMetadata](world),
		posFilter:    ecs.NewFilter1[components.Position](world),
		healthFilter: ecs.NewFilter1[components.Health](world),
	}
}

// ExportWorld creates a snapshot of the current world state.
func (e *StateExporter) ExportWorld(world *ecs.World, tick int64) WorldSnapshot {
	snapshot := WorldSnapshot{
		Tick:     tick,
		Entities: make([]EntitySnapshot, 0),
	}

	// Get component maps for random access
	posMap := ecs.NewMap[components.Position](world)
	velMap := ecs.NewMap[components.Velocity](world)
	healthMap := ecs.NewMap[components.Health](world)

	// Export entities with AIMetadata
	query := e.metaFilter.Query()
	for query.Next() {
		entity := query.Entity()
		meta := query.Get()

		snap := EntitySnapshot{
			ID:          entity.ID(),
			EntityType:  meta.EntityType,
			Description: meta.Description,
			State:       meta.VisualState,
			Tags:        meta.Tags,
			Effects:     meta.ActiveEffects,
		}

		// Add position if available
		if posMap.Has(entity) {
			pos := posMap.Get(entity)
			snap.Position = &[2]float64{pos.X, pos.Y}
		}

		// Add velocity if available
		if velMap.Has(entity) {
			vel := velMap.Get(entity)
			snap.Velocity = &[2]float64{vel.X, vel.Y}
		}

		// Add health if available
		if healthMap.Has(entity) {
			health := healthMap.Get(entity)
			snap.Health = &[2]int{health.Current, health.Max}
		}

		snapshot.Entities = append(snapshot.Entities, snap)
	}

	// Generate summary
	snapshot.Summary = e.generateSummary(snapshot)

	return snapshot
}

// generateSummary creates a brief text summary of the world state.
func (e *StateExporter) generateSummary(snapshot WorldSnapshot) string {
	counts := make(map[string]int)
	for _, entity := range snapshot.Entities {
		counts[entity.EntityType]++
	}

	parts := make([]string, 0, len(counts))
	for entityType, count := range counts {
		parts = append(parts, fmt.Sprintf("%d %s(s)", count, entityType))
	}

	if len(parts) == 0 {
		return "Empty world"
	}

	return fmt.Sprintf("Tick %d: %s", snapshot.Tick, strings.Join(parts, ", "))
}

// ExportJSON returns the world state as JSON string.
func (e *StateExporter) ExportJSON(world *ecs.World, tick int64) string {
	snapshot := e.ExportWorld(world, tick)

	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return "{}"
	}

	return string(data)
}

// ExportMarkdown returns the world state as Markdown string.
func (e *StateExporter) ExportMarkdown(world *ecs.World, tick int64) string {
	snapshot := e.ExportWorld(world, tick)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# World State (Tick %d)\n\n", snapshot.Tick))
	sb.WriteString(fmt.Sprintf("**Summary**: %s\n\n", snapshot.Summary))

	if len(snapshot.Entities) == 0 {
		sb.WriteString("No entities with AI metadata.\n")

		return sb.String()
	}

	sb.WriteString("## Entities\n\n")
	sb.WriteString("| ID | Type | Description | Position | State |\n")
	sb.WriteString("|---|---|---|---|---|\n")

	for _, e := range snapshot.Entities {
		posStr := "N/A"
		if e.Position != nil {
			posStr = fmt.Sprintf("(%.1f, %.1f)", e.Position[0], e.Position[1])
		}

		sb.WriteString(fmt.Sprintf("| %d | %s | %s | %s | %s |\n",
			e.ID, e.EntityType, e.Description, posStr, e.State))
	}

	return sb.String()
}
