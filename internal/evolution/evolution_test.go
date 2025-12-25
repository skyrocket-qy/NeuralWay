package evolution

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCodebaseAnalyzer(t *testing.T) {
	// Get current directory (should be in the project)
	cwd, _ := os.Getwd()
	projectRoot := filepath.Join(cwd, "..", "..")

	analyzer := NewCodebaseAnalyzer(projectRoot)

	report, err := analyzer.Analyze()
	if err != nil {
		t.Fatalf("Analyze error: %v", err)
	}

	if report.GoFiles == 0 {
		t.Error("Should find Go files")
	}

	if len(report.Packages) == 0 {
		t.Error("Should find packages")
	}
}

func TestCodebaseAnalyzerMarkdown(t *testing.T) {
	report := &AnalysisReport{
		GoFiles:    50,
		TestFiles:  10,
		TotalLines: 5000,
		TODOCount:  15,
	}

	analyzer := NewCodebaseAnalyzer(".")
	md := analyzer.ExportMarkdown(report)

	if md == "" {
		t.Error("Should generate markdown")
	}
}

func TestDependencyAuditor(t *testing.T) {
	cwd, _ := os.Getwd()
	projectRoot := filepath.Join(cwd, "..", "..")

	auditor := NewDependencyAuditor()

	report, err := auditor.Audit(projectRoot)
	if err != nil {
		t.Skipf("Skipping dep audit: %v", err)
	}

	if report.ModulePath == "" {
		t.Error("Should find module path")
	}
}

func TestEvolutionRunner(t *testing.T) {
	cwd, _ := os.Getwd()
	projectRoot := filepath.Join(cwd, "..", "..")

	runner := NewEvolutionRunner(projectRoot)

	report, err := runner.Run()
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}

	if report.HealthScore < 0 || report.HealthScore > 100 {
		t.Errorf("Health score should be 0-100, got %d", report.HealthScore)
	}
}

func TestEvolutionMarkdown(t *testing.T) {
	report := &EvolutionReport{
		HealthScore: 85,
		Proposals: []Proposal{
			{Title: "Test", Category: "test", Priority: 2, Description: "Test proposal"},
		},
	}

	runner := NewEvolutionRunner(".")
	md := runner.ExportMarkdown(report)

	if md == "" {
		t.Error("Should generate markdown")
	}
}
