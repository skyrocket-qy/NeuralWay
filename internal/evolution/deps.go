package evolution

import (
	"bufio"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// Dependency represents a Go module dependency.
type Dependency struct {
	Path    string
	Version string
	Latest  string
	Direct  bool
}

// Update represents an available update.
type Update struct {
	Path       string
	Current    string
	Latest     string
	IsBreaking bool // Major version change
}

// DependencyReport contains dependency analysis.
type DependencyReport struct {
	ModulePath string
	GoVersion  string
	Direct     []Dependency
	Indirect   []Dependency
	Outdated   []Update
}

// DependencyAuditor analyzes Go module dependencies.
type DependencyAuditor struct{}

// NewDependencyAuditor creates a new dependency auditor.
func NewDependencyAuditor() *DependencyAuditor {
	return &DependencyAuditor{}
}

// Audit analyzes the go.mod file and checks for updates.
func (d *DependencyAuditor) Audit(projectPath string) (*DependencyReport, error) {
	report := &DependencyReport{
		Direct:   make([]Dependency, 0),
		Indirect: make([]Dependency, 0),
		Outdated: make([]Update, 0),
	}

	// Parse go.mod
	goModPath := projectPath + "/go.mod"

	file, err := os.Open(goModPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	inRequire := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Get module path
		if after, ok := strings.CutPrefix(line, "module "); ok {
			report.ModulePath = after
		}

		// Get Go version
		if after, ok := strings.CutPrefix(line, "go "); ok {
			report.GoVersion = after
		}

		// Parse require block
		if line == "require (" {
			inRequire = true

			continue
		}

		if line == ")" {
			inRequire = false

			continue
		}

		if inRequire && line != "" {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				dep := Dependency{
					Path:    parts[0],
					Version: parts[1],
					Direct:  !strings.Contains(line, "// indirect"),
				}
				if dep.Direct {
					report.Direct = append(report.Direct, dep)
				} else {
					report.Indirect = append(report.Indirect, dep)
				}
			}
		}
	}

	// Check for outdated deps (uses go list -m -u)
	report.Outdated = d.checkOutdated(projectPath)

	return report, nil
}

// checkOutdated runs go list -m -u to find outdated dependencies.
func (d *DependencyAuditor) checkOutdated(projectPath string) []Update {
	updates := make([]Update, 0)

	cmd := exec.Command("go", "list", "-m", "-u", "-json", "all")
	cmd.Dir = projectPath

	output, err := cmd.Output()
	if err != nil {
		return updates // Return empty on error
	}

	// Simple parsing - look for Update field
	lines := strings.Split(string(output), "\n")

	var currentPath, currentVersion string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, `"Path":`) {
			currentPath = extractJSONString(line)
		}

		if strings.HasPrefix(line, `"Version":`) {
			currentVersion = extractJSONString(line)
		}

		if strings.HasPrefix(line, `"Update":`) ||
			strings.Contains(line, `"Version":`) && strings.Contains(line, "Update") {
			// Has update available
			if currentPath != "" && currentVersion != "" {
				updates = append(updates, Update{
					Path:    currentPath,
					Current: currentVersion,
				})
			}
		}
	}

	return updates
}

// extractJSONString extracts a string value from a JSON line.
func extractJSONString(line string) string {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) < 2 {
		return ""
	}

	value := strings.TrimSpace(parts[1])
	value = strings.Trim(value, `",`)

	return value
}

// ExportMarkdown generates a markdown report.
func (d *DependencyAuditor) ExportMarkdown(report *DependencyReport) string {
	var sb strings.Builder

	sb.WriteString("# Dependency Report\n\n")
	sb.WriteString("Module: `" + report.ModulePath + "`\n")
	sb.WriteString("Go Version: `" + report.GoVersion + "`\n\n")

	sb.WriteString("## Direct Dependencies (" + strconv.Itoa(len(report.Direct)) + ")\n\n")

	if len(report.Direct) > 0 {
		sb.WriteString("| Package | Version |\n")
		sb.WriteString("|---------|----------|\n")

		for _, dep := range report.Direct {
			sb.WriteString("| " + dep.Path + " | " + dep.Version + " |\n")
		}
	}

	sb.WriteString("\n")

	if len(report.Outdated) > 0 {
		sb.WriteString("## Updates Available (" + strconv.Itoa(len(report.Outdated)) + ")\n\n")

		for _, upd := range report.Outdated {
			sb.WriteString("- `" + upd.Path + "`: " + upd.Current + " → " + upd.Latest + "\n")
		}
	}

	return sb.String()
}
