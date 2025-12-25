package evolution

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"
)

// Issue represents a potential improvement or problem.
type Issue struct {
	Type        string // "todo", "fixme", "optimization", "deprecation"
	File        string
	Line        int
	Description string
	Severity    int // 1=low, 2=medium, 3=high
}

// PackageInfo contains stats about a Go package.
type PackageInfo struct {
	Name      string
	Path      string
	Files     int
	Lines     int
	TestFiles int
	TODOs     int
}

// AnalysisReport contains the full codebase analysis.
type AnalysisReport struct {
	Timestamp    time.Time
	RootPath     string
	TotalFiles   int
	TotalLines   int
	GoFiles      int
	TestFiles    int
	TODOCount    int
	FIXMECount   int
	Packages     []PackageInfo
	Issues       []Issue
	Dependencies int
}

// CodebaseAnalyzer scans and analyzes the codebase.
type CodebaseAnalyzer struct {
	rootPath       string
	ignorePatterns []string
}

// NewCodebaseAnalyzer creates a new analyzer for the given path.
func NewCodebaseAnalyzer(rootPath string) *CodebaseAnalyzer {
	return &CodebaseAnalyzer{
		rootPath: rootPath,
		ignorePatterns: []string{
			"vendor", "node_modules", ".git", ".idea",
		},
	}
}

// Analyze performs a full codebase analysis.
func (a *CodebaseAnalyzer) Analyze() (*AnalysisReport, error) {
	report := &AnalysisReport{
		Timestamp: time.Now(),
		RootPath:  a.rootPath,
		Packages:  make([]PackageInfo, 0),
		Issues:    make([]Issue, 0),
	}

	todoRegex := regexp.MustCompile(`(?i)//\s*TODO[:\s](.*)`)
	fixmeRegex := regexp.MustCompile(`(?i)//\s*FIXME[:\s](.*)`)

	packages := make(map[string]*PackageInfo)

	err := filepath.Walk(a.rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip ignored directories
		if info.IsDir() {
			if slices.Contains(a.ignorePatterns, info.Name()) {
				return filepath.SkipDir
			}

			return nil
		}

		// Only process Go files
		if !strings.HasSuffix(info.Name(), ".go") {
			return nil
		}

		report.TotalFiles++
		report.GoFiles++

		isTest := strings.HasSuffix(info.Name(), "_test.go")
		if isTest {
			report.TestFiles++
		}

		// Get package directory
		pkgDir := filepath.Dir(path)

		relPkg, _ := filepath.Rel(a.rootPath, pkgDir)
		if relPkg == "" {
			relPkg = "."
		}

		if _, ok := packages[relPkg]; !ok {
			packages[relPkg] = &PackageInfo{
				Name: filepath.Base(pkgDir),
				Path: relPkg,
			}
		}

		pkg := packages[relPkg]

		pkg.Files++
		if isTest {
			pkg.TestFiles++
		}

		// Analyze file contents
		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)

		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := scanner.Text()
			report.TotalLines++
			pkg.Lines++

			// Check for TODOs
			if matches := todoRegex.FindStringSubmatch(line); matches != nil {
				report.TODOCount++
				pkg.TODOs++

				report.Issues = append(report.Issues, Issue{
					Type:        "todo",
					File:        path,
					Line:        lineNum,
					Description: strings.TrimSpace(matches[1]),
					Severity:    1,
				})
			}

			// Check for FIXMEs
			if matches := fixmeRegex.FindStringSubmatch(line); matches != nil {
				report.FIXMECount++
				report.Issues = append(report.Issues, Issue{
					Type:        "fixme",
					File:        path,
					Line:        lineNum,
					Description: strings.TrimSpace(matches[1]),
					Severity:    2,
				})
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Convert packages map to slice
	for _, pkg := range packages {
		report.Packages = append(report.Packages, *pkg)
	}

	return report, nil
}

// ExportMarkdown generates a markdown report.
func (a *CodebaseAnalyzer) ExportMarkdown(report *AnalysisReport) string {
	var sb strings.Builder

	sb.WriteString("# Codebase Analysis Report\n\n")
	sb.WriteString("Generated: " + report.Timestamp.Format(time.RFC3339) + "\n\n")

	sb.WriteString("## Summary\n\n")
	sb.WriteString("| Metric | Value |\n")
	sb.WriteString("|--------|-------|\n")
	sb.WriteString("| Total Go Files | " + itoa(report.GoFiles) + " |\n")
	sb.WriteString("| Test Files | " + itoa(report.TestFiles) + " |\n")
	sb.WriteString("| Total Lines | " + itoa(report.TotalLines) + " |\n")
	sb.WriteString("| Packages | " + itoa(len(report.Packages)) + " |\n")
	sb.WriteString("| TODOs | " + itoa(report.TODOCount) + " |\n")
	sb.WriteString("| FIXMEs | " + itoa(report.FIXMECount) + " |\n")
	sb.WriteString("\n")

	if len(report.Packages) > 0 {
		sb.WriteString("## Packages\n\n")
		sb.WriteString("| Package | Files | Lines | Tests | TODOs |\n")
		sb.WriteString("|---------|-------|-------|-------|-------|\n")

		for _, pkg := range report.Packages {
			sb.WriteString(
				"| " + pkg.Path + " | " + itoa(
					pkg.Files,
				) + " | " + itoa(
					pkg.Lines,
				) + " | " + itoa(
					pkg.TestFiles,
				) + " | " + itoa(
					pkg.TODOs,
				) + " |\n",
			)
		}

		sb.WriteString("\n")
	}

	if len(report.Issues) > 0 && len(report.Issues) <= 20 {
		sb.WriteString("## Issues\n\n")

		for _, issue := range report.Issues {
			sb.WriteString("- **" + strings.ToUpper(issue.Type) + "**: " + issue.Description + "\n")
		}
	} else if len(report.Issues) > 20 {
		sb.WriteString("## Issues\n\n")
		sb.WriteString("Found " + itoa(len(report.Issues)) + " issues (showing first 10):\n\n")

		for i, issue := range report.Issues[:10] {
			_ = i

			sb.WriteString("- **" + strings.ToUpper(issue.Type) + "**: " + issue.Description + "\n")
		}
	}

	return sb.String()
}

func itoa(i int) string {
	return strings.TrimSpace(strings.ReplaceAll(string(rune('0'+i%10)), "\x00", ""))
}
