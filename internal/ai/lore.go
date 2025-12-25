package ai

import (
	"sort"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// LoreEntry represents a piece of game lore.
type LoreEntry struct {
	ID        string
	Title     string
	Content   string
	Category  string // "character", "location", "event", "item"
	Tags      []string
	Embedding []float32 // Optional: pre-computed embedding for similarity search
	Relevance float32   // Computed similarity score
}

// LoreDatabase is a simple RAG-style lore database.
type LoreDatabase struct {
	entries map[string]*LoreEntry
	index   map[string][]*LoreEntry // Tag -> entries index
}

// NewLoreDatabase creates a new lore database.
func NewLoreDatabase() *LoreDatabase {
	return &LoreDatabase{
		entries: make(map[string]*LoreEntry),
		index:   make(map[string][]*LoreEntry),
	}
}

// AddEntry adds a lore entry to the database.
func (db *LoreDatabase) AddEntry(entry LoreEntry) {
	db.entries[entry.ID] = &entry

	// Index by category
	db.index[entry.Category] = append(db.index[entry.Category], &entry)

	// Index by tags
	for _, tag := range entry.Tags {
		tag = strings.ToLower(tag)
		db.index[tag] = append(db.index[tag], &entry)
	}
}

// Query searches for relevant lore entries.
func (db *LoreDatabase) Query(query string, limit int) []*LoreEntry {
	queryLower := strings.ToLower(query)
	words := strings.Fields(queryLower)

	// Score all entries based on keyword matches
	scored := make([]*LoreEntry, 0)

	for _, entry := range db.entries {
		score := db.computeRelevance(entry, words)
		if score > 0 {
			entry.Relevance = score
			scored = append(scored, entry)
		}
	}

	// Sort by relevance
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Relevance > scored[j].Relevance
	})

	// Return top results
	if len(scored) > limit {
		scored = scored[:limit]
	}

	return scored
}

// computeRelevance calculates relevance score based on keyword matching.
func (db *LoreDatabase) computeRelevance(entry *LoreEntry, queryWords []string) float32 {
	var score float32

	contentLower := strings.ToLower(entry.Content)
	titleLower := strings.ToLower(entry.Title)

	for _, word := range queryWords {
		// Title match is worth more
		if strings.Contains(titleLower, word) {
			score += 2.0
		}
		// Content match
		if strings.Contains(contentLower, word) {
			score += 1.0
		}
		// Tag match is worth even more
		for _, tag := range entry.Tags {
			if strings.Contains(strings.ToLower(tag), word) {
				score += 3.0
			}
		}
	}

	return score
}

// QueryByCategory returns all entries in a category.
func (db *LoreDatabase) QueryByCategory(category string) []*LoreEntry {
	return db.index[category]
}

// QueryByTag returns all entries with a specific tag.
func (db *LoreDatabase) QueryByTag(tag string) []*LoreEntry {
	return db.index[strings.ToLower(tag)]
}

// GetEntry returns a specific entry by ID.
func (db *LoreDatabase) GetEntry(id string) *LoreEntry {
	return db.entries[id]
}

// GetAllEntries returns all entries.
func (db *LoreDatabase) GetAllEntries() []*LoreEntry {
	all := make([]*LoreEntry, 0, len(db.entries))
	for _, entry := range db.entries {
		all = append(all, entry)
	}

	return all
}

// GenerateContextPrompt creates a prompt with relevant lore for LLM.
func (db *LoreDatabase) GenerateContextPrompt(query, additionalContext string) string {
	entries := db.Query(query, 3)

	var sb strings.Builder
	sb.WriteString("## Relevant Lore\n\n")

	for _, entry := range entries {
		sb.WriteString("### " + entry.Title + "\n")
		sb.WriteString(entry.Content + "\n\n")
	}

	if additionalContext != "" {
		sb.WriteString("## Current Context\n\n")
		sb.WriteString(additionalContext + "\n\n")
	}

	sb.WriteString("## Query\n\n")
	sb.WriteString(query + "\n")

	return sb.String()
}

// ExportMarkdown exports all lore as a markdown document.
func (db *LoreDatabase) ExportMarkdown() string {
	var sb strings.Builder
	sb.WriteString("# Game Lore\n\n")

	// Group by category
	categories := make(map[string][]*LoreEntry)
	for _, entry := range db.entries {
		categories[entry.Category] = append(categories[entry.Category], entry)
	}

	for category, entries := range categories {
		sb.WriteString("## " + cases.Title(language.English).String(category) + "s\n\n")

		for _, entry := range entries {
			sb.WriteString("### " + entry.Title + "\n\n")
			sb.WriteString(entry.Content + "\n\n")

			if len(entry.Tags) > 0 {
				sb.WriteString("*Tags: " + strings.Join(entry.Tags, ", ") + "*\n\n")
			}
		}
	}

	return sb.String()
}

// Count returns the number of entries.
func (db *LoreDatabase) Count() int {
	return len(db.entries)
}
