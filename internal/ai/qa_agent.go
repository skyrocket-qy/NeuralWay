package ai

import (
	"fmt"
	"strings"
	"time"

	"github.com/mlange-42/ark/ecs"
	"td/internal/engine"
)

// QAEntry represents a single step in the QA log.
type QAEntry struct {
	Tick      int64
	State     string // Compact state before action
	Action    string // Action taken
	Result    string // Outcome description
	Timestamp time.Time
}

// QAAgent is an automated game tester that logs state and actions.
type QAAgent struct {
	game     *engine.HeadlessGame
	executor *ActionExecutor
	exporter *CompactExporter
	log      []QAEntry
}

// NewQAAgent creates a QA agent for the given headless game.
func NewQAAgent(game *engine.HeadlessGame) *QAAgent {
	return &QAAgent{
		game:     game,
		executor: NewActionExecutor(&game.World),
		exporter: NewCompactExporter(&game.World),
		log:      make([]QAEntry, 0),
	}
}

// DecisionFn is a function that decides what action to take based on state.
type DecisionFn func(state string, world *ecs.World) Action

// Step runs one game tick with a decision function.
func (a *QAAgent) Step(decisionFn DecisionFn) QAEntry {
	// Capture state before
	state := a.exporter.Export(&a.game.World, a.game.CurrentTick())

	// Get action from decision function
	var (
		action     Action
		actionName string
	)

	if decisionFn != nil {
		action = decisionFn(state, &a.game.World)
		if action != nil {
			actionName = action.Name()
			a.executor.Execute(action)
		}
	}

	// Run game step
	a.game.Step()

	// Log entry
	entry := QAEntry{
		Tick:      a.game.CurrentTick(),
		State:     state,
		Action:    actionName,
		Timestamp: time.Now(),
	}
	a.log = append(a.log, entry)

	return entry
}

// Run executes multiple steps with the decision function.
func (a *QAAgent) Run(steps int, decisionFn DecisionFn) []QAEntry {
	entries := make([]QAEntry, 0, steps)
	for range steps {
		entry := a.Step(decisionFn)
		entries = append(entries, entry)
	}

	return entries
}

// RunUntil runs until condition is met or max steps reached.
func (a *QAAgent) RunUntil(cond func(*ecs.World) bool, maxSteps int, decisionFn DecisionFn) []QAEntry {
	entries := make([]QAEntry, 0)

	for range maxSteps {
		if cond(&a.game.World) {
			break
		}

		entry := a.Step(decisionFn)
		entries = append(entries, entry)
	}

	return entries
}

// Log returns all logged entries.
func (a *QAAgent) Log() []QAEntry {
	return a.log
}

// ClearLog clears the log.
func (a *QAAgent) ClearLog() {
	a.log = make([]QAEntry, 0)
}

// ExportLog exports the log as a markdown report.
func (a *QAAgent) ExportLog() string {
	var sb strings.Builder
	sb.WriteString("# QA Agent Log\n\n")
	sb.WriteString(fmt.Sprintf("Total steps: %d\n\n", len(a.log)))

	if len(a.log) == 0 {
		sb.WriteString("No entries.\n")

		return sb.String()
	}

	sb.WriteString("| Tick | State | Action |\n")
	sb.WriteString("|------|-------|--------|\n")

	for _, entry := range a.log {
		// Truncate state for table
		state := entry.State
		if len(state) > 40 {
			state = state[:40] + "..."
		}

		sb.WriteString(fmt.Sprintf("| %d | %s | %s |\n", entry.Tick, state, entry.Action))
	}

	return sb.String()
}

// ExportLogCompact exports the log as a compact string for LLM analysis.
func (a *QAAgent) ExportLogCompact() string {
	var parts []string
	for _, entry := range a.log {
		parts = append(parts, fmt.Sprintf("%d:%s→%s", entry.Tick, entry.State, entry.Action))
	}

	return strings.Join(parts, "\n")
}
