package ai

import (
	"strings"
)

// DialogueNode represents a node in a dialogue tree.
type DialogueNode struct {
	ID        string
	Speaker   string
	Text      string
	Choices   []DialogueChoice
	Condition func() bool // Optional condition to show this node
	OnEnter   func()      // Callback when entering this node
}

// DialogueChoice represents a player choice.
type DialogueChoice struct {
	Text      string
	NextID    string
	Condition func() bool // Optional condition to show this choice
	OnSelect  func()      // Callback when selected
}

// Quest represents a game quest.
type Quest struct {
	ID          string
	Title       string
	Description string
	Objectives  []QuestObjective
	Status      QuestStatus
	Rewards     []string
	OnComplete  func()
}

// QuestObjective is a single objective within a quest.
type QuestObjective struct {
	Description string
	Current     int
	Target      int
	Completed   bool
}

// QuestStatus represents quest state.
type QuestStatus int

const (
	QuestNotStarted QuestStatus = iota
	QuestActive
	QuestCompleted
	QuestFailed
)

// NarrativeController manages dynamic storytelling and dialogue.
type NarrativeController struct {
	dialogues      map[string]*DialogueNode
	quests         map[string]*Quest
	currentDialog  string
	variables      map[string]any
	onDialogChange func(node *DialogueNode)
}

// NewNarrativeController creates a new narrative controller.
func NewNarrativeController() *NarrativeController {
	return &NarrativeController{
		dialogues: make(map[string]*DialogueNode),
		quests:    make(map[string]*Quest),
		variables: make(map[string]any),
	}
}

// AddDialogue adds a dialogue node.
func (n *NarrativeController) AddDialogue(node DialogueNode) {
	n.dialogues[node.ID] = &node
}

// StartDialogue begins a dialogue sequence.
func (n *NarrativeController) StartDialogue(id string) *DialogueNode {
	node, ok := n.dialogues[id]
	if !ok {
		return nil
	}

	if node.Condition != nil && !node.Condition() {
		return nil
	}

	n.currentDialog = id

	if node.OnEnter != nil {
		node.OnEnter()
	}

	if n.onDialogChange != nil {
		n.onDialogChange(node)
	}

	return node
}

// GetCurrentDialogue returns the current dialogue node.
func (n *NarrativeController) GetCurrentDialogue() *DialogueNode {
	return n.dialogues[n.currentDialog]
}

// SelectChoice selects a dialogue choice.
func (n *NarrativeController) SelectChoice(choiceIndex int) *DialogueNode {
	current := n.dialogues[n.currentDialog]
	if current == nil || choiceIndex >= len(current.Choices) {
		return nil
	}

	choice := current.Choices[choiceIndex]
	if choice.OnSelect != nil {
		choice.OnSelect()
	}

	return n.StartDialogue(choice.NextID)
}

// GetAvailableChoices returns choices that pass their conditions.
func (n *NarrativeController) GetAvailableChoices() []DialogueChoice {
	current := n.dialogues[n.currentDialog]
	if current == nil {
		return nil
	}

	available := make([]DialogueChoice, 0)

	for _, choice := range current.Choices {
		if choice.Condition == nil || choice.Condition() {
			available = append(available, choice)
		}
	}

	return available
}

// AddQuest adds a quest.
func (n *NarrativeController) AddQuest(quest Quest) {
	n.quests[quest.ID] = &quest
}

// StartQuest activates a quest.
func (n *NarrativeController) StartQuest(id string) {
	if quest, ok := n.quests[id]; ok {
		quest.Status = QuestActive
	}
}

// UpdateObjective updates a quest objective progress.
func (n *NarrativeController) UpdateObjective(questID string, objectiveIndex, progress int) {
	quest, ok := n.quests[questID]
	if !ok || objectiveIndex >= len(quest.Objectives) {
		return
	}

	obj := &quest.Objectives[objectiveIndex]

	obj.Current = progress
	if obj.Current >= obj.Target {
		obj.Completed = true
	}

	// Check if all objectives complete
	allComplete := true

	for _, o := range quest.Objectives {
		if !o.Completed {
			allComplete = false

			break
		}
	}

	if allComplete {
		quest.Status = QuestCompleted
		if quest.OnComplete != nil {
			quest.OnComplete()
		}
	}
}

// GetActiveQuests returns all active quests.
func (n *NarrativeController) GetActiveQuests() []*Quest {
	active := make([]*Quest, 0)

	for _, quest := range n.quests {
		if quest.Status == QuestActive {
			active = append(active, quest)
		}
	}

	return active
}

// SetVariable sets a narrative variable.
func (n *NarrativeController) SetVariable(key string, value any) {
	n.variables[key] = value
}

// GetVariable gets a narrative variable.
func (n *NarrativeController) GetVariable(key string) any {
	return n.variables[key]
}

// OnDialogueChange sets a callback for dialogue changes.
func (n *NarrativeController) OnDialogueChange(callback func(node *DialogueNode)) {
	n.onDialogChange = callback
}

// GenerateDialoguePrompt creates a prompt for LLM dialogue generation.
func (n *NarrativeController) GenerateDialoguePrompt(character, context string) string {
	var sb strings.Builder
	sb.WriteString("Generate dialogue for a game character.\n\n")
	sb.WriteString("Character: " + character + "\n")
	sb.WriteString("Context: " + context + "\n\n")
	sb.WriteString("Generate a short, in-character response (1-3 sentences).\n")
	sb.WriteString("Format: Just the dialogue text, no attribution.")

	return sb.String()
}

// GenerateQuestPrompt creates a prompt for LLM quest generation.
func (n *NarrativeController) GenerateQuestPrompt(theme, difficulty string) string {
	var sb strings.Builder
	sb.WriteString("Generate a game quest.\n\n")
	sb.WriteString("Theme: " + theme + "\n")
	sb.WriteString("Difficulty: " + difficulty + "\n\n")
	sb.WriteString("Format as JSON:\n")
	sb.WriteString(`{"title": "...", "description": "...", "objectives": ["..."], "rewards": ["..."]}`)

	return sb.String()
}
