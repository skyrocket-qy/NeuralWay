package game

// TutorialStep represents a single tutorial instruction.
type TutorialStep struct {
	ID           string
	Title        string
	Description  string
	Trigger      TutorialTrigger
	Condition    func() bool // Optional condition to check
	HighlightX   float64     // Screen position to highlight
	HighlightY   float64
	HighlightW   float64
	HighlightH   float64
	RequireInput string  // Input required to proceed (e.g., "click", "key:W")
	AutoAdvance  float64 // Auto-advance after N seconds (0 = manual)
	Completed    bool
}

// TutorialTrigger defines when a tutorial step should show.
type TutorialTrigger int

const (
	TriggerImmediate   TutorialTrigger = iota // Show immediately
	TriggerOnEvent                            // Show when event fires
	TriggerOnCondition                        // Show when condition met
	TriggerAfterDelay                         // Show after delay
)

// TutorialSystem manages in-game tutorials.
type TutorialSystem struct {
	Steps       []*TutorialStep
	CurrentStep int
	Active      bool
	Paused      bool
	StepTimer   float64
	SkipEnabled bool

	// Display state
	ShowingStep  bool
	DisplayTimer float64

	// Callbacks
	onStepStart        func(step *TutorialStep)
	onStepComplete     func(step *TutorialStep)
	onTutorialComplete func()

	// Progress tracking
	CompletedIDs map[string]bool
}

// NewTutorialSystem creates a tutorial system.
func NewTutorialSystem() *TutorialSystem {
	return &TutorialSystem{
		Steps:        make([]*TutorialStep, 0),
		CompletedIDs: make(map[string]bool),
		SkipEnabled:  true,
	}
}

// SetOnStepStart sets the step start callback.
func (t *TutorialSystem) SetOnStepStart(fn func(step *TutorialStep)) {
	t.onStepStart = fn
}

// SetOnStepComplete sets the step complete callback.
func (t *TutorialSystem) SetOnStepComplete(fn func(step *TutorialStep)) {
	t.onStepComplete = fn
}

// SetOnTutorialComplete sets the tutorial complete callback.
func (t *TutorialSystem) SetOnTutorialComplete(fn func()) {
	t.onTutorialComplete = fn
}

// AddStep adds a tutorial step.
func (t *TutorialSystem) AddStep(step *TutorialStep) {
	t.Steps = append(t.Steps, step)
}

// Start begins the tutorial.
func (t *TutorialSystem) Start() {
	if len(t.Steps) == 0 {
		return
	}

	t.Active = true
	t.CurrentStep = 0
	t.startCurrentStep()
}

// startCurrentStep initiates the current step.
func (t *TutorialSystem) startCurrentStep() {
	if t.CurrentStep >= len(t.Steps) {
		t.complete()

		return
	}

	step := t.Steps[t.CurrentStep]
	t.ShowingStep = true
	t.StepTimer = 0
	t.DisplayTimer = 0

	if t.onStepStart != nil {
		t.onStepStart(step)
	}
}

// Update updates the tutorial system.
func (t *TutorialSystem) Update(dt float64) {
	if !t.Active || t.Paused {
		return
	}

	if t.CurrentStep >= len(t.Steps) {
		return
	}

	step := t.Steps[t.CurrentStep]
	t.StepTimer += dt
	t.DisplayTimer += dt

	// Check auto-advance
	if step.AutoAdvance > 0 && t.StepTimer >= step.AutoAdvance {
		t.AdvanceStep()

		return
	}

	// Check condition
	if step.Condition != nil && step.Condition() {
		t.AdvanceStep()
	}
}

// AdvanceStep moves to the next step.
func (t *TutorialSystem) AdvanceStep() {
	if t.CurrentStep >= len(t.Steps) {
		return
	}

	step := t.Steps[t.CurrentStep]
	step.Completed = true
	t.CompletedIDs[step.ID] = true

	if t.onStepComplete != nil {
		t.onStepComplete(step)
	}

	t.CurrentStep++
	t.ShowingStep = false

	if t.CurrentStep >= len(t.Steps) {
		t.complete()
	} else {
		t.startCurrentStep()
	}
}

// complete finishes the tutorial.
func (t *TutorialSystem) complete() {
	t.Active = false

	t.ShowingStep = false
	if t.onTutorialComplete != nil {
		t.onTutorialComplete()
	}
}

// Skip skips the current step.
func (t *TutorialSystem) Skip() {
	if t.SkipEnabled {
		t.AdvanceStep()
	}
}

// SkipAll skips the entire tutorial.
func (t *TutorialSystem) SkipAll() {
	t.Active = false
	t.ShowingStep = false
	// Mark all as completed
	for _, step := range t.Steps {
		step.Completed = true
		t.CompletedIDs[step.ID] = true
	}
}

// Pause pauses the tutorial.
func (t *TutorialSystem) Pause() {
	t.Paused = true
}

// Resume resumes the tutorial.
func (t *TutorialSystem) Resume() {
	t.Paused = false
}

// GetCurrentStep returns the current tutorial step.
func (t *TutorialSystem) GetCurrentStep() *TutorialStep {
	if t.CurrentStep < len(t.Steps) {
		return t.Steps[t.CurrentStep]
	}

	return nil
}

// GetProgress returns tutorial progress (0.0 - 1.0).
func (t *TutorialSystem) GetProgress() float64 {
	if len(t.Steps) == 0 {
		return 1.0
	}

	return float64(t.CurrentStep) / float64(len(t.Steps))
}

// IsComplete returns true if the tutorial is finished.
func (t *TutorialSystem) IsComplete() bool {
	return t.CurrentStep >= len(t.Steps)
}

// HasCompleted returns true if a specific step was completed.
func (t *TutorialSystem) HasCompleted(stepID string) bool {
	return t.CompletedIDs[stepID]
}

// Reset resets the tutorial to the beginning.
func (t *TutorialSystem) Reset() {
	t.CurrentStep = 0
	t.Active = false
	t.ShowingStep = false
	t.StepTimer = 0

	t.CompletedIDs = make(map[string]bool)
	for _, step := range t.Steps {
		step.Completed = false
	}
}

// GetHighlight returns the highlight rectangle for the current step.
func (t *TutorialSystem) GetHighlight() (x, y, w, h float64, hasHighlight bool) {
	step := t.GetCurrentStep()
	if step == nil || (step.HighlightW == 0 && step.HighlightH == 0) {
		return 0, 0, 0, 0, false
	}

	return step.HighlightX, step.HighlightY, step.HighlightW, step.HighlightH, true
}

// CommonTutorials provides pre-built tutorial steps for common actions.
type CommonTutorials struct{}

// MovementTutorial creates steps for teaching movement.
func (c *CommonTutorials) MovementTutorial() []*TutorialStep {
	return []*TutorialStep{
		{
			ID:          "move_intro",
			Title:       "Movement",
			Description: "Use WASD or Arrow Keys to move your character.",
			Trigger:     TriggerImmediate,
			AutoAdvance: 3.0,
		},
		{
			ID:           "move_practice",
			Title:        "Try Moving",
			Description:  "Move around to continue.",
			Trigger:      TriggerOnCondition,
			RequireInput: "move",
		},
	}
}

// CombatTutorial creates steps for teaching combat.
func (c *CommonTutorials) CombatTutorial() []*TutorialStep {
	return []*TutorialStep{
		{
			ID:          "combat_intro",
			Title:       "Combat",
			Description: "Click or press Space to attack enemies.",
			Trigger:     TriggerImmediate,
			AutoAdvance: 3.0,
		},
		{
			ID:           "combat_practice",
			Title:        "Attack an Enemy",
			Description:  "Attack an enemy to continue.",
			Trigger:      TriggerOnCondition,
			RequireInput: "attack",
		},
	}
}

// InventoryTutorial creates steps for teaching inventory.
func (c *CommonTutorials) InventoryTutorial() []*TutorialStep {
	return []*TutorialStep{
		{
			ID:          "inventory_intro",
			Title:       "Inventory",
			Description: "Press I to open your inventory.",
			Trigger:     TriggerImmediate,
			AutoAdvance: 3.0,
		},
	}
}

// TooltipSystem provides contextual help tooltips.
type TooltipSystem struct {
	Tooltips   map[string]*Tooltip
	CurrentID  string
	ShowDelay  float64
	HideDelay  float64
	hoverTimer float64
	Visible    bool
}

// Tooltip represents a help tooltip.
type Tooltip struct {
	ID          string
	Title       string
	Description string
	X, Y        float64
	Width       float64
	Shown       bool
}

// NewTooltipSystem creates a tooltip system.
func NewTooltipSystem() *TooltipSystem {
	return &TooltipSystem{
		Tooltips:  make(map[string]*Tooltip),
		ShowDelay: 0.5,
		HideDelay: 0.2,
	}
}

// Register registers a tooltip.
func (ts *TooltipSystem) Register(tooltip *Tooltip) {
	ts.Tooltips[tooltip.ID] = tooltip
}

// Show shows a tooltip.
func (ts *TooltipSystem) Show(id string) {
	if tooltip, ok := ts.Tooltips[id]; ok {
		ts.CurrentID = id
		ts.Visible = true
		tooltip.Shown = true
	}
}

// Hide hides the current tooltip.
func (ts *TooltipSystem) Hide() {
	if ts.CurrentID != "" {
		if tooltip, ok := ts.Tooltips[ts.CurrentID]; ok {
			tooltip.Shown = false
		}
	}

	ts.Visible = false
	ts.CurrentID = ""
}

// Update updates the tooltip system.
func (ts *TooltipSystem) Update(dt float64) {
	// Handle hover timing logic here
}

// GetCurrent returns the current tooltip.
func (ts *TooltipSystem) GetCurrent() *Tooltip {
	if ts.CurrentID != "" {
		return ts.Tooltips[ts.CurrentID]
	}

	return nil
}
