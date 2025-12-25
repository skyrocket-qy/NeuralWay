package ai

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/mlange-42/ark/ecs"
	"td/internal/components"
	"td/internal/engine"
)

func TestStateExporterExportWorld(t *testing.T) {
	world := ecs.NewWorld()
	exporter := NewStateExporter(&world)

	// Create entities with AIMetadata
	mapper := ecs.NewMap2[components.Position, components.AIMetadata](&world)

	mapper.NewEntity(
		&components.Position{X: 100, Y: 200},
		&components.AIMetadata{
			EntityType:  "player",
			Description: "Hero character",
			VisualState: "idle",
			Tags:        []string{"controllable"},
		},
	)

	mapper.NewEntity(
		&components.Position{X: 50, Y: 75},
		&components.AIMetadata{
			EntityType:  "enemy",
			Description: "Goblin",
			VisualState: "patrol",
			Tags:        []string{"hostile"},
		},
	)

	snapshot := exporter.ExportWorld(&world, 42)

	if snapshot.Tick != 42 {
		t.Errorf("Tick should be 42, got %d", snapshot.Tick)
	}

	if len(snapshot.Entities) != 2 {
		t.Errorf("Should have 2 entities, got %d", len(snapshot.Entities))
	}
}

func TestStateExporterExportJSON(t *testing.T) {
	world := ecs.NewWorld()
	exporter := NewStateExporter(&world)

	mapper := ecs.NewMap2[components.Position, components.AIMetadata](&world)
	mapper.NewEntity(
		&components.Position{X: 10, Y: 20},
		&components.AIMetadata{EntityType: "test", Description: "Test entity"},
	)

	jsonStr := exporter.ExportJSON(&world, 1)

	// Verify it's valid JSON
	var data map[string]any

	err := json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		t.Errorf("Invalid JSON: %v", err)
	}

	if data["tick"].(float64) != 1 {
		t.Error("JSON tick should be 1")
	}
}

func TestStateExporterExportMarkdown(t *testing.T) {
	world := ecs.NewWorld()
	exporter := NewStateExporter(&world)

	mapper := ecs.NewMap2[components.Position, components.AIMetadata](&world)
	mapper.NewEntity(
		&components.Position{X: 10, Y: 20},
		&components.AIMetadata{EntityType: "player", Description: "Hero"},
	)

	md := exporter.ExportMarkdown(&world, 5)

	if !strings.Contains(md, "# World State (Tick 5)") {
		t.Error("Markdown should contain header with tick")
	}

	if !strings.Contains(md, "player") {
		t.Error("Markdown should contain entity type")
	}
}

func TestActionExecutorMoveAction(t *testing.T) {
	world := ecs.NewWorld()
	executor := NewActionExecutor(&world)

	mapper := ecs.NewMap1[components.Position](&world)
	entity := mapper.NewEntity(&components.Position{X: 0, Y: 0})

	action := MoveAction{Entity: entity, DX: 10, DY: 5}
	executor.Execute(action)

	posMap := ecs.NewMap[components.Position](&world)

	pos := posMap.Get(entity)
	if pos.X != 10 || pos.Y != 5 {
		t.Errorf("Position should be (10, 5), got (%.1f, %.1f)", pos.X, pos.Y)
	}
}

func TestActionExecutorSetVelocityAction(t *testing.T) {
	world := ecs.NewWorld()
	executor := NewActionExecutor(&world)

	mapper := ecs.NewMap1[components.Velocity](&world)
	entity := mapper.NewEntity(&components.Velocity{X: 0, Y: 0})

	action := SetVelocityAction{Entity: entity, VX: 5, VY: -3}
	executor.Execute(action)

	velMap := ecs.NewMap[components.Velocity](&world)

	vel := velMap.Get(entity)
	if vel.X != 5 || vel.Y != -3 {
		t.Errorf("Velocity should be (5, -3), got (%.1f, %.1f)", vel.X, vel.Y)
	}
}

func TestActionExecutorRemoveEntityAction(t *testing.T) {
	world := ecs.NewWorld()
	executor := NewActionExecutor(&world)

	mapper := ecs.NewMap1[components.Position](&world)
	entity := mapper.NewEntity(&components.Position{X: 0, Y: 0})

	if !world.Alive(entity) {
		t.Error("Entity should be alive before remove")
	}

	action := RemoveEntityAction{Entity: entity}
	executor.Execute(action)

	if world.Alive(entity) {
		t.Error("Entity should be removed")
	}
}

func TestActionExecutorBatch(t *testing.T) {
	world := ecs.NewWorld()
	executor := NewActionExecutor(&world)

	mapper := ecs.NewMap1[components.Position](&world)
	entity := mapper.NewEntity(&components.Position{X: 0, Y: 0})

	actions := []Action{
		MoveAction{Entity: entity, DX: 10, DY: 0},
		MoveAction{Entity: entity, DX: 5, DY: 3},
	}

	errors := executor.ExecuteBatch(actions)
	if len(errors) != 0 {
		t.Errorf("Should have no errors, got %d", len(errors))
	}

	posMap := ecs.NewMap[components.Position](&world)

	pos := posMap.Get(entity)
	if pos.X != 15 || pos.Y != 3 {
		t.Errorf("Position should be (15, 3), got (%.1f, %.1f)", pos.X, pos.Y)
	}
}

func TestCompactExporterExport(t *testing.T) {
	world := ecs.NewWorld()
	exporter := NewCompactExporter(&world)

	mapper := ecs.NewMap2[components.Position, components.AIMetadata](&world)
	mapper.NewEntity(
		&components.Position{X: 100, Y: 200},
		&components.AIMetadata{EntityType: "player", VisualState: "idle"},
	)
	mapper.NewEntity(
		&components.Position{X: 50, Y: 75},
		&components.AIMetadata{EntityType: "enemy", VisualState: "patrol"},
	)

	compact := exporter.Export(&world, 42)

	if !strings.HasPrefix(compact, "T42|") {
		t.Errorf("Should start with T42|, got %s", compact)
	}

	if !strings.Contains(compact, "P:100,200") {
		t.Errorf("Should contain player position, got %s", compact)
	}

	if !strings.Contains(compact, "E:50,75") {
		t.Errorf("Should contain enemy position, got %s", compact)
	}
}

func TestCompactExporterWithHealth(t *testing.T) {
	world := ecs.NewWorld()
	exporter := NewCompactExporter(&world)

	mapper := ecs.NewMap3[components.Position, components.AIMetadata, components.Health](&world)
	mapper.NewEntity(
		&components.Position{X: 10, Y: 20},
		&components.AIMetadata{EntityType: "player"},
		&components.Health{Current: 80, Max: 100},
	)

	compact := exporter.Export(&world, 1)

	if !strings.Contains(compact, "H:80/100") {
		t.Errorf("Should contain health, got %s", compact)
	}
}

func TestTokenEstimate(t *testing.T) {
	// Short string
	est := TokenEstimate("T1|P:0,0")
	if est != 2 {
		t.Errorf("Token estimate for 8 chars should be 2, got %d", est)
	}

	// Longer string
	est2 := TokenEstimate("T100|P:100,200|E:50,75;H:80/100|I:30,40")
	if est2 > 15 {
		t.Errorf("Compact format should be ~10 tokens, got %d", est2)
	}
}

func TestQAAgentStep(t *testing.T) {
	game := engine.NewHeadlessGame()
	agent := NewQAAgent(game)

	entry := agent.Step(nil)
	if entry.Tick != 1 {
		t.Errorf("Tick should be 1, got %d", entry.Tick)
	}

	if len(agent.Log()) != 1 {
		t.Error("Log should have 1 entry")
	}
}

func TestQAAgentRun(t *testing.T) {
	game := engine.NewHeadlessGame()
	agent := NewQAAgent(game)

	entries := agent.Run(10, nil)
	if len(entries) != 10 {
		t.Errorf("Should have 10 entries, got %d", len(entries))
	}
}

func TestQAAgentExportLog(t *testing.T) {
	game := engine.NewHeadlessGame()
	agent := NewQAAgent(game)

	agent.Run(3, nil)
	log := agent.ExportLog()

	if !strings.Contains(log, "# QA Agent Log") {
		t.Error("Log should have header")
	}

	if !strings.Contains(log, "Total steps: 3") {
		t.Error("Log should show 3 steps")
	}
}

func TestBehaviorTreeSequence(t *testing.T) {
	ctx := NewBTContext(nil, ecs.Entity{})

	count := 0
	seq := NewSequence(
		NewActionNode(func(ctx *BTContext) NodeStatus {
			count++

			return Success
		}),
		NewActionNode(func(ctx *BTContext) NodeStatus {
			count++

			return Success
		}),
	)

	status := seq.Tick(ctx)
	if status != Success {
		t.Error("Sequence should succeed")
	}

	if count != 2 {
		t.Errorf("Both actions should run, count=%d", count)
	}
}

func TestBehaviorTreeSelector(t *testing.T) {
	ctx := NewBTContext(nil, ecs.Entity{})

	sel := NewSelector(
		NewActionNode(func(ctx *BTContext) NodeStatus { return Failure }),
		NewActionNode(func(ctx *BTContext) NodeStatus { return Success }),
		NewActionNode(func(ctx *BTContext) NodeStatus { return Failure }),
	)

	status := sel.Tick(ctx)
	if status != Success {
		t.Error("Selector should succeed on second child")
	}
}

func TestBehaviorTreeCondition(t *testing.T) {
	ctx := NewBTContext(nil, ecs.Entity{})
	ctx.Data["health"] = 50

	cond := NewCondition(func(ctx *BTContext) bool {
		return ctx.Data["health"].(int) > 25
	})

	if cond.Tick(ctx) != Success {
		t.Error("Condition should succeed when health > 25")
	}
}

func TestBehaviorTreeInverter(t *testing.T) {
	ctx := NewBTContext(nil, ecs.Entity{})

	inv := NewInverter(&SucceedNode{})
	if inv.Tick(ctx) != Failure {
		t.Error("Inverter should turn Success into Failure")
	}
}

func TestDifficultyBalancerRegisterParameter(t *testing.T) {
	db := NewDifficultyBalancer()

	db.RegisterParameter("enemy_speed", 5.0, 2.0, 10.0, false)

	val := db.GetParameter("enemy_speed")
	if val != 5.0 {
		t.Errorf("Parameter should be 5.0, got %.1f", val)
	}
}

func TestDifficultyBalancerRecordMetric(t *testing.T) {
	db := NewDifficultyBalancer()

	db.RegisterMetric("win_rate", 0.5, 0.1, 10)
	db.RecordMetric("win_rate", 0.8)
	db.RecordMetric("win_rate", 0.7)

	avg := db.GetMetricAverage("win_rate")
	if avg != 0.75 {
		t.Errorf("Metric average should be 0.75, got %.2f", avg)
	}
}

func TestDifficultyBalancerUpdate(t *testing.T) {
	db := NewDifficultyBalancer()

	db.RegisterMetric("win_rate", 0.5, 0.1, 10)
	db.RegisterParameter("enemy_speed", 5.0, 2.0, 10.0, false)

	// Player winning too much
	for range 5 {
		db.RecordMetric("win_rate", 0.9)
	}

	initialLevel := db.GetDifficultyLevel()
	db.Update()

	// Difficulty should increase
	if db.GetDifficultyLevel() <= initialLevel {
		t.Error("Difficulty should increase when player is winning too much")
	}
}

func TestDifficultyBalancerReset(t *testing.T) {
	db := NewDifficultyBalancer()

	db.RegisterParameter("test", 5.0, 1.0, 10.0, false)
	db.SetDifficultyLevel(1.5)
	db.Reset()

	if db.GetDifficultyLevel() != 1.0 {
		t.Error("Difficulty level should reset to 1.0")
	}
}
