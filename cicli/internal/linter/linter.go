package linter

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Severity represents the severity of a lint issue
type Severity string

const (
	Error   Severity = "error"
	Warning Severity = "warning"
	Info    Severity = "info"
)

// Issue represents a linting issue
type Issue struct {
	Severity    Severity `json:"severity"`
	Rule        string   `json:"rule"`
	Message     string   `json:"message"`
	File        string   `json:"file"`
	Line        int      `json:"line,omitempty"`
	Suggestion  string   `json:"suggestion,omitempty"`
	AutoFixable bool     `json:"auto_fixable"`
}

// LintResult contains all linting results
type LintResult struct {
	Platform string  `json:"platform"`
	File     string  `json:"file"`
	Issues   []Issue `json:"issues"`
	Score    int     `json:"score"` // 0-100
}

// Linter validates CI/CD configuration files
type Linter struct {
	rules []Rule
}

// Rule defines a linting rule
type Rule struct {
	ID          string
	Name        string
	Description string
	Severity    Severity
	Platforms   []string // Which platforms this rule applies to
	Check       func(content []byte, file string) []Issue
}

// NewLinter creates a new linter with all rules
func NewLinter() *Linter {
	l := &Linter{}
	l.registerRules()
	return l
}

// registerRules registers all linting rules
func (l *Linter) registerRules() {
	l.rules = []Rule{
		// Security rules
		{
			ID:          "SEC001",
			Name:        "hardcoded-secrets",
			Description: "Detect hardcoded secrets and credentials",
			Severity:    Error,
			Platforms:   []string{"github", "gitlab", "circleci", "azure", "jenkins"},
			Check:       checkHardcodedSecrets,
		},
		{
			ID:          "SEC002",
			Name:        "insecure-commands",
			Description: "Detect potentially insecure commands",
			Severity:    Warning,
			Platforms:   []string{"github", "gitlab", "circleci", "azure", "jenkins"},
			Check:       checkInsecureCommands,
		},
		{
			ID:          "SEC003",
			Name:        "unpinned-actions",
			Description: "Actions should use SHA pinning for security",
			Severity:    Warning,
			Platforms:   []string{"github"},
			Check:       checkUnpinnedActions,
		},

		// Best practices
		{
			ID:          "BP001",
			Name:        "missing-timeout",
			Description: "Jobs should have timeout limits",
			Severity:    Warning,
			Platforms:   []string{"github", "gitlab"},
			Check:       checkMissingTimeout,
		},
		{
			ID:          "BP002",
			Name:        "missing-concurrency",
			Description: "Workflows should define concurrency to prevent duplicate runs",
			Severity:    Info,
			Platforms:   []string{"github"},
			Check:       checkMissingConcurrency,
		},
		{
			ID:          "BP003",
			Name:        "outdated-actions",
			Description: "Using outdated action versions",
			Severity:    Warning,
			Platforms:   []string{"github"},
			Check:       checkOutdatedActions,
		},

		// Performance
		{
			ID:          "PERF001",
			Name:        "missing-cache",
			Description: "Dependencies should be cached for faster builds",
			Severity:    Info,
			Platforms:   []string{"github", "gitlab", "circleci"},
			Check:       checkMissingCache,
		},
		{
			ID:          "PERF002",
			Name:        "sequential-jobs",
			Description: "Jobs that could run in parallel are sequential",
			Severity:    Info,
			Platforms:   []string{"github", "gitlab"},
			Check:       checkSequentialJobs,
		},

		// Reliability
		{
			ID:          "REL001",
			Name:        "missing-retry",
			Description: "Flaky steps should have retry logic",
			Severity:    Info,
			Platforms:   []string{"github", "gitlab"},
			Check:       checkMissingRetry,
		},
		{
			ID:          "REL002",
			Name:        "missing-error-handling",
			Description: "Commands should handle errors appropriately",
			Severity:    Warning,
			Platforms:   []string{"github", "gitlab", "jenkins"},
			Check:       checkErrorHandling,
		},
	}
}

// Lint performs linting on a CI config file
func (l *Linter) Lint(filePath string) (*LintResult, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	platform := detectPlatform(filePath)
	result := &LintResult{
		Platform: platform,
		File:     filePath,
		Issues:   []Issue{},
	}

	for _, rule := range l.rules {
		if !l.ruleApplies(rule, platform) {
			continue
		}

		issues := rule.Check(content, filePath)
		for i := range issues {
			issues[i].Rule = rule.ID
		}
		result.Issues = append(result.Issues, issues...)
	}

	result.Score = l.calculateScore(result.Issues)
	return result, nil
}

// LintDirectory lints all CI config files in a directory
func (l *Linter) LintDirectory(dir string) ([]*LintResult, error) {
	var results []*LintResult

	ciPaths := []string{
		".github/workflows/*.yml",
		".github/workflows/*.yaml",
		".gitlab-ci.yml",
		".circleci/config.yml",
		"azure-pipelines.yml",
		"Jenkinsfile",
		"bitbucket-pipelines.yml",
	}

	for _, pattern := range ciPaths {
		matches, err := filepath.Glob(filepath.Join(dir, pattern))
		if err != nil {
			continue
		}

		for _, match := range matches {
			result, err := l.Lint(match)
			if err != nil {
				continue
			}
			results = append(results, result)
		}
	}

	return results, nil
}

func (l *Linter) ruleApplies(rule Rule, platform string) bool {
	for _, p := range rule.Platforms {
		if p == platform {
			return true
		}
	}
	return false
}

func (l *Linter) calculateScore(issues []Issue) int {
	score := 100
	for _, issue := range issues {
		switch issue.Severity {
		case Error:
			score -= 20
		case Warning:
			score -= 10
		case Info:
			score -= 2
		}
	}
	if score < 0 {
		score = 0
	}
	return score
}

// Rule implementations

func checkHardcodedSecrets(content []byte, file string) []Issue {
	var issues []Issue

	patterns := []struct {
		name    string
		pattern *regexp.Regexp
	}{
		{"AWS Access Key", regexp.MustCompile(`AKIA[0-9A-Z]{16}`)},
		{"AWS Secret Key", regexp.MustCompile(`(?i)(aws_secret_access_key|aws_secret_key)\s*[:=]\s*['"]?[A-Za-z0-9/+=]{40}`)},
		{"Generic API Key", regexp.MustCompile(`(?i)(api[_-]?key|apikey)\s*[:=]\s*['"]?[A-Za-z0-9]{20,}`)},
		{"Generic Secret", regexp.MustCompile(`(?i)(secret|password|passwd|pwd)\s*[:=]\s*['"][^'"]{8,}['"]`)},
		{"Private Key", regexp.MustCompile(`-----BEGIN (RSA |EC |DSA )?PRIVATE KEY-----`)},
		{"GitHub Token", regexp.MustCompile(`ghp_[A-Za-z0-9]{36}`)},
		{"Slack Token", regexp.MustCompile(`xox[baprs]-[0-9]{10,13}-[0-9]{10,13}[a-zA-Z0-9-]*`)},
	}

	lines := strings.Split(string(content), "\n")
	for lineNum, line := range lines {
		for _, p := range patterns {
			if p.pattern.MatchString(line) {
				issues = append(issues, Issue{
					Severity:   Error,
					Message:    fmt.Sprintf("Potential %s detected", p.name),
					File:       file,
					Line:       lineNum + 1,
					Suggestion: "Use secrets/environment variables instead of hardcoded values",
				})
			}
		}
	}

	return issues
}

func checkInsecureCommands(content []byte, file string) []Issue {
	var issues []Issue

	dangerousPatterns := []struct {
		pattern *regexp.Regexp
		message string
		fix     string
	}{
		{
			regexp.MustCompile(`curl\s+[^|]*\|\s*(ba)?sh`),
			"Piping curl to shell is dangerous",
			"Download and verify scripts before execution",
		},
		{
			regexp.MustCompile(`wget\s+[^|]*\|\s*(ba)?sh`),
			"Piping wget to shell is dangerous",
			"Download and verify scripts before execution",
		},
		{
			regexp.MustCompile(`chmod\s+777`),
			"Setting 777 permissions is insecure",
			"Use more restrictive permissions (e.g., 755 or 644)",
		},
		{
			regexp.MustCompile(`--insecure|--no-check-certificate|-k\s`),
			"Disabling SSL verification is insecure",
			"Fix SSL certificates instead of disabling verification",
		},
		{
			regexp.MustCompile(`eval\s*\(`),
			"Using eval can be dangerous",
			"Avoid eval when possible, use safer alternatives",
		},
	}

	lines := strings.Split(string(content), "\n")
	for lineNum, line := range lines {
		for _, p := range dangerousPatterns {
			if p.pattern.MatchString(line) {
				issues = append(issues, Issue{
					Severity:   Warning,
					Message:    p.message,
					File:       file,
					Line:       lineNum + 1,
					Suggestion: p.fix,
				})
			}
		}
	}

	return issues
}

func checkUnpinnedActions(content []byte, file string) []Issue {
	var issues []Issue

	// Match actions that use branch refs instead of SHA
	pattern := regexp.MustCompile(`uses:\s*([^@]+)@(v\d+|main|master|latest)`)
	lines := strings.Split(string(content), "\n")

	for lineNum, line := range lines {
		matches := pattern.FindStringSubmatch(line)
		if len(matches) > 2 {
			// Skip first-party actions from actions/ org
			if !strings.HasPrefix(matches[1], "actions/") {
				issues = append(issues, Issue{
					Severity:    Warning,
					Message:     fmt.Sprintf("Action '%s' uses tag '%s' instead of SHA pin", matches[1], matches[2]),
					File:        file,
					Line:        lineNum + 1,
					Suggestion:  "Pin to a specific commit SHA for security (e.g., @a1b2c3d4...)",
					AutoFixable: false,
				})
			}
		}
	}

	return issues
}

func checkMissingTimeout(content []byte, file string) []Issue {
	var issues []Issue

	var config map[string]interface{}
	if err := yaml.Unmarshal(content, &config); err != nil {
		return issues
	}

	// Check GitHub Actions
	if jobs, ok := config["jobs"].(map[string]interface{}); ok {
		for jobName, jobData := range jobs {
			if jd, ok := jobData.(map[string]interface{}); ok {
				if _, hasTimeout := jd["timeout-minutes"]; !hasTimeout {
					issues = append(issues, Issue{
						Severity:    Warning,
						Message:     fmt.Sprintf("Job '%s' has no timeout defined", jobName),
						File:        file,
						Suggestion:  "Add 'timeout-minutes' to prevent hung jobs",
						AutoFixable: true,
					})
				}
			}
		}
	}

	return issues
}

func checkMissingConcurrency(content []byte, file string) []Issue {
	var issues []Issue

	if !strings.Contains(string(content), "concurrency:") {
		issues = append(issues, Issue{
			Severity:   Info,
			Message:    "Workflow has no concurrency control",
			File:       file,
			Suggestion: "Add 'concurrency' to cancel outdated runs on the same branch",
		})
	}

	return issues
}

func checkOutdatedActions(content []byte, file string) []Issue {
	var issues []Issue

	// Map of actions to their latest major versions
	latestVersions := map[string]int{
		"actions/checkout":       4,
		"actions/setup-node":     4,
		"actions/setup-python":   5,
		"actions/setup-go":       5,
		"actions/cache":          4,
		"actions/upload-artifact": 4,
		"actions/download-artifact": 4,
		"docker/build-push-action": 6,
		"docker/login-action":    3,
	}

	pattern := regexp.MustCompile(`uses:\s*([^@]+)@v(\d+)`)
	lines := strings.Split(string(content), "\n")

	for lineNum, line := range lines {
		matches := pattern.FindStringSubmatch(line)
		if len(matches) > 2 {
			action := matches[1]
			var version int
			fmt.Sscanf(matches[2], "%d", &version)

			if latest, ok := latestVersions[action]; ok && version < latest {
				issues = append(issues, Issue{
					Severity:    Warning,
					Message:     fmt.Sprintf("Action '%s@v%d' is outdated (latest: v%d)", action, version, latest),
					File:        file,
					Line:        lineNum + 1,
					Suggestion:  fmt.Sprintf("Update to %s@v%d", action, latest),
					AutoFixable: true,
				})
			}
		}
	}

	return issues
}

func checkMissingCache(content []byte, file string) []Issue {
	var issues []Issue

	contentStr := string(content)

	// Check if package managers are used but cache is missing
	packageManagers := []struct {
		indicator string
		cacheKey  string
		name      string
	}{
		{"npm install", "npm", "npm"},
		{"npm ci", "npm", "npm"},
		{"yarn install", "yarn", "yarn"},
		{"pnpm install", "pnpm", "pnpm"},
		{"pip install", "pip", "pip"},
		{"go build", "go", "go modules"},
		{"cargo build", "cargo", "cargo"},
	}

	for _, pm := range packageManagers {
		if strings.Contains(contentStr, pm.indicator) {
			if !strings.Contains(contentStr, "cache") && !strings.Contains(contentStr, pm.cacheKey+"-cache") {
				issues = append(issues, Issue{
					Severity:   Info,
					Message:    fmt.Sprintf("Using %s but no cache configured", pm.name),
					File:       file,
					Suggestion: fmt.Sprintf("Add caching for %s dependencies to speed up builds", pm.name),
				})
				break // Only report once
			}
		}
	}

	return issues
}

func checkSequentialJobs(content []byte, file string) []Issue {
	var issues []Issue

	var config map[string]interface{}
	if err := yaml.Unmarshal(content, &config); err != nil {
		return issues
	}

	// Check for jobs that don't depend on each other but run sequentially
	if jobs, ok := config["jobs"].(map[string]interface{}); ok {
		jobsWithNeeds := 0
		totalJobs := len(jobs)

		for _, jobData := range jobs {
			if jd, ok := jobData.(map[string]interface{}); ok {
				if _, hasNeeds := jd["needs"]; hasNeeds {
					jobsWithNeeds++
				}
			}
		}

		// If many jobs have dependencies, they might be overly sequential
		if totalJobs > 2 && jobsWithNeeds == totalJobs-1 {
			issues = append(issues, Issue{
				Severity:   Info,
				Message:    "Jobs appear to be running sequentially",
				File:       file,
				Suggestion: "Consider if some jobs can run in parallel to reduce build time",
			})
		}
	}

	return issues
}

func checkMissingRetry(content []byte, file string) []Issue {
	var issues []Issue

	// Check for network-related commands without retry
	flakyPatterns := []string{
		"npm install",
		"npm ci",
		"yarn install",
		"pip install",
		"apt-get install",
		"docker pull",
		"docker push",
	}

	contentStr := string(content)
	hasRetry := strings.Contains(contentStr, "retry") ||
		strings.Contains(contentStr, "continue-on-error") ||
		strings.Contains(contentStr, "attempts")

	if !hasRetry {
		for _, pattern := range flakyPatterns {
			if strings.Contains(contentStr, pattern) {
				issues = append(issues, Issue{
					Severity:   Info,
					Message:    fmt.Sprintf("'%s' may fail due to network issues", pattern),
					File:       file,
					Suggestion: "Consider adding retry logic for network-dependent operations",
				})
				break
			}
		}
	}

	return issues
}

func checkErrorHandling(content []byte, file string) []Issue {
	var issues []Issue

	// Check for multi-line run commands without error handling
	pattern := regexp.MustCompile(`run:\s*\|`)
	lines := strings.Split(string(content), "\n")

	for lineNum, line := range lines {
		if pattern.MatchString(line) {
			// Check next few lines for 'set -e' or error handling
			hasErrorHandling := false
			for i := lineNum + 1; i < len(lines) && i < lineNum+5; i++ {
				if strings.Contains(lines[i], "set -e") ||
					strings.Contains(lines[i], "set -o errexit") ||
					strings.Contains(lines[i], "|| exit") ||
					strings.Contains(lines[i], "|| true") {
					hasErrorHandling = true
					break
				}
				// Stop if we hit another key
				if strings.Contains(lines[i], ":") && !strings.HasPrefix(strings.TrimSpace(lines[i]), "-") {
					break
				}
			}

			if !hasErrorHandling {
				issues = append(issues, Issue{
					Severity:   Warning,
					Message:    "Multi-line script without explicit error handling",
					File:       file,
					Line:       lineNum + 1,
					Suggestion: "Add 'set -e' at the start of multi-line scripts",
				})
			}
		}
	}

	return issues
}

func detectPlatform(path string) string {
	switch {
	case strings.Contains(path, ".github"):
		return "github"
	case strings.Contains(path, ".gitlab-ci"):
		return "gitlab"
	case strings.Contains(path, "circleci"):
		return "circleci"
	case strings.Contains(path, "azure-pipelines"):
		return "azure"
	case strings.Contains(path, "Jenkinsfile"):
		return "jenkins"
	case strings.Contains(path, "bitbucket"):
		return "bitbucket"
	default:
		return "unknown"
	}
}

// PrintReport outputs a formatted lint report
func (r *LintResult) PrintReport() {
	fmt.Printf("\n🔍 Lint Report: %s\n", r.File)
	fmt.Printf("   Platform: %s\n", r.Platform)
	fmt.Printf("   Score: %d/100\n", r.Score)
	fmt.Println(strings.Repeat("─", 50))

	if len(r.Issues) == 0 {
		fmt.Println("   ✅ No issues found!")
		return
	}

	// Group by severity
	errors := []Issue{}
	warnings := []Issue{}
	infos := []Issue{}

	for _, issue := range r.Issues {
		switch issue.Severity {
		case Error:
			errors = append(errors, issue)
		case Warning:
			warnings = append(warnings, issue)
		case Info:
			infos = append(infos, issue)
		}
	}

	if len(errors) > 0 {
		fmt.Println("\n   🚨 Errors:")
		for _, issue := range errors {
			printIssue(issue)
		}
	}

	if len(warnings) > 0 {
		fmt.Println("\n   ⚠️  Warnings:")
		for _, issue := range warnings {
			printIssue(issue)
		}
	}

	if len(infos) > 0 {
		fmt.Println("\n   ℹ️  Info:")
		for _, issue := range infos {
			printIssue(issue)
		}
	}

	fmt.Println()
}

func printIssue(issue Issue) {
	loc := ""
	if issue.Line > 0 {
		loc = fmt.Sprintf(" (line %d)", issue.Line)
	}
	fmt.Printf("      [%s]%s %s\n", issue.Rule, loc, issue.Message)
	if issue.Suggestion != "" {
		fmt.Printf("         → %s\n", issue.Suggestion)
	}
}
