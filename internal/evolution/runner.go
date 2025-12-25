package evolution

import (
	"fmt"
	"strings"
	"time"
)

// Proposal represents a suggested improvement.
type Proposal struct {
	Title       string
	Category    string // "feature", "refactor", "dependency", "documentation"
	Priority    int    // 1=low, 2=medium, 3=high
	Description string
	Rationale   string
}

// EvolutionReport contains the full evolution analysis.
type EvolutionReport struct {
	Timestamp   time.Time
	Codebase    *AnalysisReport
	Deps        *DependencyReport
	Proposals   []Proposal
	HealthScore int // 0-100
}

// EvolutionRunner orchestrates the self-evolution analysis.
type EvolutionRunner struct {
	rootPath string
	analyzer *CodebaseAnalyzer
	auditor  *DependencyAuditor
}

// NewEvolutionRunner creates a new evolution runner.
func NewEvolutionRunner(rootPath string) *EvolutionRunner {
	return &EvolutionRunner{
		rootPath: rootPath,
		analyzer: NewCodebaseAnalyzer(rootPath),
		auditor:  NewDependencyAuditor(),
	}
}

// Run performs a full evolution analysis.
func (r *EvolutionRunner) Run() (*EvolutionReport, error) {
	report := &EvolutionReport{
		Timestamp: time.Now(),
		Proposals: make([]Proposal, 0),
	}

	// Analyze codebase
	codebaseReport, err := r.analyzer.Analyze()
	if err != nil {
		return nil, err
	}

	report.Codebase = codebaseReport

	// Audit dependencies
	depsReport, err := r.auditor.Audit(r.rootPath)
	if err != nil {
		// Non-fatal, continue without deps
		report.Deps = &DependencyReport{}
	} else {
		report.Deps = depsReport
	}

	// Generate proposals
	report.Proposals = r.generateProposals(report)

	// Calculate health score
	report.HealthScore = r.calculateHealth(report)

	return report, nil
}

// generateProposals creates improvement suggestions based on analysis.
func (r *EvolutionRunner) generateProposals(report *EvolutionReport) []Proposal {
	proposals := make([]Proposal, 0)

	// Check for high FIXME count
	if report.Codebase != nil && report.Codebase.FIXMECount > 5 {
		proposals = append(proposals, Proposal{
			Title:       "Address FIXME Comments",
			Category:    "refactor",
			Priority:    3,
			Description: fmt.Sprintf("Found %d FIXME comments in codebase", report.Codebase.FIXMECount),
			Rationale:   "FIXMEs indicate known issues that should be resolved",
		})
	}

	// Check for packages without tests
	if report.Codebase != nil {
		for _, pkg := range report.Codebase.Packages {
			if pkg.Files > 2 && pkg.TestFiles == 0 {
				proposals = append(proposals, Proposal{
					Title:       "Add Tests for " + pkg.Name,
					Category:    "testing",
					Priority:    2,
					Description: fmt.Sprintf("Package %s has %d files but no tests", pkg.Path, pkg.Files),
					Rationale:   "Test coverage improves reliability",
				})
			}
		}
	}

	// Check for outdated dependencies
	if report.Deps != nil && len(report.Deps.Outdated) > 0 {
		proposals = append(proposals, Proposal{
			Title:       "Update Dependencies",
			Category:    "dependency",
			Priority:    2,
			Description: fmt.Sprintf("%d dependencies have updates available", len(report.Deps.Outdated)),
			Rationale:   "Keep dependencies current for security and features",
		})
	}

	// Check for large packages
	if report.Codebase != nil {
		for _, pkg := range report.Codebase.Packages {
			if pkg.Lines > 2000 {
				proposals = append(proposals, Proposal{
					Title:       "Consider Splitting " + pkg.Name,
					Category:    "refactor",
					Priority:    1,
					Description: fmt.Sprintf("Package %s has %d lines", pkg.Path, pkg.Lines),
					Rationale:   "Large packages may benefit from being split into smaller modules",
				})
			}
		}
	}

	return proposals
}

// calculateHealth computes a health score from 0-100.
func (r *EvolutionRunner) calculateHealth(report *EvolutionReport) int {
	score := 100

	if report.Codebase != nil {
		// Deduct for excessive TODOs
		todoRatio := float64(report.Codebase.TODOCount) / float64(max(report.Codebase.GoFiles, 1))
		if todoRatio > 2 {
			score -= 10
		}

		// Deduct for FIXMEs
		score -= min(report.Codebase.FIXMECount*3, 20)

		// Bonus for test coverage
		if report.Codebase.TestFiles > 0 {
			testRatio := float64(
				report.Codebase.TestFiles,
			) / float64(
				max(report.Codebase.GoFiles-report.Codebase.TestFiles, 1),
			)
			if testRatio > 0.5 {
				score += 5
			}
		}
	}

	// Deduct for outdated deps
	if report.Deps != nil {
		score -= min(len(report.Deps.Outdated)*2, 15)
	}

	return max(0, min(100, score))
}

// ExportMarkdown generates a full evolution report.
func (r *EvolutionRunner) ExportMarkdown(report *EvolutionReport) string {
	var sb strings.Builder

	sb.WriteString("# 🧬 Evolution Report\n\n")
	sb.WriteString("Generated: " + report.Timestamp.Format(time.RFC3339) + "\n\n")
	sb.WriteString(fmt.Sprintf("**Health Score: %d/100**\n\n", report.HealthScore))

	sb.WriteString("---\n\n")

	// Summary
	if report.Codebase != nil {
		sb.WriteString("## Codebase Summary\n\n")
		sb.WriteString(
			fmt.Sprintf(
				"- **Go Files**: %d (%d tests)\n",
				report.Codebase.GoFiles,
				report.Codebase.TestFiles,
			),
		)
		sb.WriteString(fmt.Sprintf("- **Total Lines**: %d\n", report.Codebase.TotalLines))
		sb.WriteString(fmt.Sprintf("- **Packages**: %d\n", len(report.Codebase.Packages)))
		sb.WriteString(
			fmt.Sprintf(
				"- **TODOs**: %d | **FIXMEs**: %d\n\n",
				report.Codebase.TODOCount,
				report.Codebase.FIXMECount,
			),
		)
	}

	// Proposals
	if len(report.Proposals) > 0 {
		sb.WriteString("## Proposals\n\n")

		for _, p := range report.Proposals {
			priority := []string{"🟢", "🟡", "🔴"}[p.Priority-1]
			sb.WriteString(fmt.Sprintf("### %s %s\n\n", priority, p.Title))
			sb.WriteString(p.Description + "\n\n")
			sb.WriteString("*" + p.Rationale + "*\n\n")
		}
	}

	return sb.String()
}
