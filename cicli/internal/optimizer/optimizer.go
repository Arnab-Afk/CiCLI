package optimizer

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Optimization represents a suggested optimization
type Optimization struct {
	Category      string  `json:"category"`
	Title         string  `json:"title"`
	Description   string  `json:"description"`
	Impact        string  `json:"impact"` // high, medium, low
	EstimatedSave string  `json:"estimated_save,omitempty"`
	Before        string  `json:"before,omitempty"`
	After         string  `json:"after,omitempty"`
	AutoApply     bool    `json:"auto_apply"`
}

// OptimizationResult contains optimization analysis
type OptimizationResult struct {
	File          string         `json:"file"`
	Platform      string         `json:"platform"`
	Optimizations []Optimization `json:"optimizations"`
	PotentialSave string         `json:"potential_save"`
}

// Optimizer analyzes and optimizes CI/CD configurations
type Optimizer struct{}

// NewOptimizer creates a new optimizer
func NewOptimizer() *Optimizer {
	return &Optimizer{}
}

// Analyze analyzes a CI config and suggests optimizations
func (o *Optimizer) Analyze(filePath string) (*OptimizationResult, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	platform := detectPlatform(filePath)
	result := &OptimizationResult{
		File:          filePath,
		Platform:      platform,
		Optimizations: []Optimization{},
	}

	switch platform {
	case "github":
		o.analyzeGitHub(content, result)
	case "gitlab":
		o.analyzeGitLab(content, result)
	default:
		o.analyzeGeneric(content, result)
	}

	result.PotentialSave = o.estimateTotalSave(result.Optimizations)
	return result, nil
}

// analyzeGitHub analyzes GitHub Actions workflows
func (o *Optimizer) analyzeGitHub(content []byte, result *OptimizationResult) {
	var config map[string]interface{}
	if err := yaml.Unmarshal(content, &config); err != nil {
		return
	}

	// Check for caching opportunities
	o.checkCachingGitHub(content, result)

	// Check for parallel job opportunities
	o.checkParallelization(config, result)

	// Check for matrix builds
	o.checkMatrixBuilds(config, result)

	// Check for job consolidation
	o.checkJobConsolidation(config, result)

	// Check for composite actions opportunity
	o.checkCompositeActions(config, result)

	// Check for artifact optimization
	o.checkArtifactOptimization(content, result)

	// Check for runner optimization
	o.checkRunnerOptimization(config, result)

	// Check for conditional execution
	o.checkConditionalExecution(content, result)
}

// checkCachingGitHub checks for caching opportunities in GitHub Actions
func (o *Optimizer) checkCachingGitHub(content []byte, result *OptimizationResult) {
	contentStr := string(content)
	hasCache := strings.Contains(contentStr, "actions/cache") ||
		strings.Contains(contentStr, "cache:") ||
		strings.Contains(contentStr, "with:\n          cache:")

	// Check for package manager usage without caching
	packageManagers := []struct {
		pattern   string
		cacheType string
		saveTime  string
	}{
		{"npm ci", "npm", "30-60s"},
		{"npm install", "npm", "30-60s"},
		{"yarn install", "yarn", "30-60s"},
		{"pnpm install", "pnpm", "20-40s"},
		{"pip install", "pip", "20-40s"},
		{"go build", "go", "20-30s"},
		{"cargo build", "cargo", "60-120s"},
		{"bundle install", "bundler", "30-60s"},
		{"composer install", "composer", "15-30s"},
	}

	if !hasCache {
		for _, pm := range packageManagers {
			if strings.Contains(contentStr, pm.pattern) {
				result.Optimizations = append(result.Optimizations, Optimization{
					Category:      "caching",
					Title:         fmt.Sprintf("Add %s dependency caching", pm.cacheType),
					Description:   fmt.Sprintf("Dependencies are installed with '%s' but not cached between runs", pm.pattern),
					Impact:        "high",
					EstimatedSave: pm.saveTime,
					Before:        fmt.Sprintf("- run: %s", pm.pattern),
					After: fmt.Sprintf(`- uses: actions/setup-node@v4
  with:
    cache: '%s'
- run: %s`, pm.cacheType, pm.pattern),
					AutoApply: true,
				})
				break
			}
		}
	}

	// Check for build artifact caching
	if strings.Contains(contentStr, "go build") && !strings.Contains(contentStr, "go-cache") {
		result.Optimizations = append(result.Optimizations, Optimization{
			Category:      "caching",
			Title:         "Cache Go build artifacts",
			Description:   "Go build artifacts can be cached to speed up subsequent builds",
			Impact:        "medium",
			EstimatedSave: "15-30s",
			AutoApply:     true,
		})
	}

	// Check for Docker layer caching
	if strings.Contains(contentStr, "docker build") && !strings.Contains(contentStr, "cache-from") {
		result.Optimizations = append(result.Optimizations, Optimization{
			Category:      "caching",
			Title:         "Enable Docker layer caching",
			Description:   "Docker builds can use layer caching to speed up image builds",
			Impact:        "high",
			EstimatedSave: "60-300s",
			Before:        "docker build -t image .",
			After: `docker build -t image . \\
  --cache-from type=gha \\
  --cache-to type=gha,mode=max`,
			AutoApply: true,
		})
	}
}

// checkParallelization checks for jobs that could run in parallel
func (o *Optimizer) checkParallelization(config map[string]interface{}, result *OptimizationResult) {
	jobs, ok := config["jobs"].(map[string]interface{})
	if !ok {
		return
	}

	// Analyze job dependencies
	jobOrder := []string{}
	jobDeps := make(map[string][]string)

	for jobName, jobData := range jobs {
		jobOrder = append(jobOrder, jobName)
		if jd, ok := jobData.(map[string]interface{}); ok {
			if needs, ok := jd["needs"]; ok {
				switch n := needs.(type) {
				case string:
					jobDeps[jobName] = []string{n}
				case []interface{}:
					for _, dep := range n {
						jobDeps[jobName] = append(jobDeps[jobName], fmt.Sprint(dep))
					}
				}
			}
		}
	}

	// Find jobs that could be parallelized
	independentJobs := []string{}
	for job, deps := range jobDeps {
		if len(deps) == 0 {
			independentJobs = append(independentJobs, job)
		}
	}

	// Check for lint/test that could run in parallel
	hasLint := false
	hasTest := false
	lintDependsOnBuild := false
	testDependsOnBuild := false

	for jobName, deps := range jobDeps {
		if strings.Contains(strings.ToLower(jobName), "lint") {
			hasLint = true
			for _, dep := range deps {
				if strings.Contains(strings.ToLower(dep), "build") {
					lintDependsOnBuild = true
				}
			}
		}
		if strings.Contains(strings.ToLower(jobName), "test") {
			hasTest = true
			for _, dep := range deps {
				if strings.Contains(strings.ToLower(dep), "build") {
					testDependsOnBuild = true
				}
			}
		}
	}

	if hasLint && hasTest && lintDependsOnBuild && testDependsOnBuild {
		result.Optimizations = append(result.Optimizations, Optimization{
			Category:      "parallelization",
			Title:         "Run lint and test in parallel",
			Description:   "Lint and test jobs both depend on build but could run in parallel after build completes",
			Impact:        "medium",
			EstimatedSave: "30-120s",
			AutoApply:     false,
		})
	}
}

// checkMatrixBuilds checks if matrix builds could be used
func (o *Optimizer) checkMatrixBuilds(config map[string]interface{}, result *OptimizationResult) {
	jobs, ok := config["jobs"].(map[string]interface{})
	if !ok {
		return
	}

	// Look for similar jobs that could be consolidated into a matrix
	for _, jobData := range jobs {
		jd, ok := jobData.(map[string]interface{})
		if !ok {
			continue
		}

		// Check if already using matrix
		if _, hasMatrix := jd["strategy"]; hasMatrix {
			continue
		}

		// Check for multiple node/python/go versions
		steps, ok := jd["steps"].([]interface{})
		if !ok {
			continue
		}

		for _, step := range steps {
			if sd, ok := step.(map[string]interface{}); ok {
				uses := getString(sd, "uses")

				// Check if using setup actions without matrix
				if strings.Contains(uses, "setup-node") ||
					strings.Contains(uses, "setup-python") ||
					strings.Contains(uses, "setup-go") {

					result.Optimizations = append(result.Optimizations, Optimization{
						Category:    "matrix-builds",
						Title:       "Consider matrix builds for multiple versions",
						Description: "Using setup action for single version. Matrix builds can test multiple versions in parallel",
						Impact:      "low",
						Before: `jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/setup-node@v4
        with:
          node-version: '18'`,
						After: `jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        node-version: [18, 20, 22]
    steps:
      - uses: actions/setup-node@v4
        with:
          node-version: \${{ matrix.node-version }}`,
						AutoApply: false,
					})
					return
				}
			}
		}
	}
}

// checkJobConsolidation checks if multiple jobs could be consolidated
func (o *Optimizer) checkJobConsolidation(config map[string]interface{}, result *OptimizationResult) {
	jobs, ok := config["jobs"].(map[string]interface{})
	if !ok {
		return
	}

	// Check for many small sequential jobs
	if len(jobs) > 5 {
		linearChain := 0
		for _, jobData := range jobs {
			if jd, ok := jobData.(map[string]interface{}); ok {
				if needs, ok := jd["needs"]; ok {
					if _, isString := needs.(string); isString {
						linearChain++
					}
				}
			}
		}

		if linearChain > 3 {
			result.Optimizations = append(result.Optimizations, Optimization{
				Category:      "consolidation",
				Title:         "Consider consolidating sequential jobs",
				Description:   "Multiple small sequential jobs add overhead. Consider combining related steps",
				Impact:        "medium",
				EstimatedSave: "30-60s (job startup overhead)",
				AutoApply:     false,
			})
		}
	}
}

// checkCompositeActions checks for repeated steps that could be composite actions
func (o *Optimizer) checkCompositeActions(config map[string]interface{}, result *OptimizationResult) {
	jobs, ok := config["jobs"].(map[string]interface{})
	if !ok {
		return
	}

	// Count repeated step patterns
	stepPatterns := make(map[string]int)

	for _, jobData := range jobs {
		if jd, ok := jobData.(map[string]interface{}); ok {
			if steps, ok := jd["steps"].([]interface{}); ok {
				for _, step := range steps {
					if sd, ok := step.(map[string]interface{}); ok {
						// Create a signature for the step
						uses := getString(sd, "uses")
						run := getString(sd, "run")
						sig := uses + "|" + run
						if sig != "|" {
							stepPatterns[sig]++
						}
					}
				}
			}
		}
	}

	// Check for patterns repeated more than twice
	for pattern, count := range stepPatterns {
		if count > 2 && len(pattern) > 20 {
			result.Optimizations = append(result.Optimizations, Optimization{
				Category:    "reusability",
				Title:       "Extract repeated steps to composite action",
				Description: fmt.Sprintf("Same step pattern repeated %d times across jobs", count),
				Impact:      "low",
				AutoApply:   false,
			})
			break
		}
	}
}

// checkArtifactOptimization checks for artifact optimization opportunities
func (o *Optimizer) checkArtifactOptimization(content []byte, result *OptimizationResult) {
	contentStr := string(content)

	// Check for large artifact uploads without compression
	if strings.Contains(contentStr, "upload-artifact") {
		if !strings.Contains(contentStr, "compression-level") {
			result.Optimizations = append(result.Optimizations, Optimization{
				Category:    "artifacts",
				Title:       "Optimize artifact compression",
				Description: "Artifact uploads can use compression-level to reduce storage and transfer time",
				Impact:      "low",
				Before: `- uses: actions/upload-artifact@v4
  with:
    name: build
    path: dist/`,
				After: `- uses: actions/upload-artifact@v4
  with:
    name: build
    path: dist/
    compression-level: 9`,
				AutoApply: true,
			})
		}
	}

	// Check for uploading node_modules or other large directories
	if regexp.MustCompile(`upload-artifact[\s\S]*?path:[\s\S]*?node_modules`).MatchString(contentStr) {
		result.Optimizations = append(result.Optimizations, Optimization{
			Category:    "artifacts",
			Title:       "Don't upload node_modules as artifact",
			Description: "node_modules should not be uploaded as artifacts. Use caching instead",
			Impact:      "high",
			AutoApply:   false,
		})
	}
}

// checkRunnerOptimization checks for runner optimization opportunities
func (o *Optimizer) checkRunnerOptimization(config map[string]interface{}, result *OptimizationResult) {
	jobs, ok := config["jobs"].(map[string]interface{})
	if !ok {
		return
	}

	for jobName, jobData := range jobs {
		if jd, ok := jobData.(map[string]interface{}); ok {
			runsOn := getString(jd, "runs-on")

			// Check for ubuntu-latest (prefer specific version for consistency)
			if runsOn == "ubuntu-latest" {
				result.Optimizations = append(result.Optimizations, Optimization{
					Category:    "runners",
					Title:       fmt.Sprintf("Pin runner version in '%s'", jobName),
					Description: "Using 'ubuntu-latest' can cause unexpected breaks when GitHub updates the image",
					Impact:      "low",
					Before:     "runs-on: ubuntu-latest",
					After:      "runs-on: ubuntu-24.04",
					AutoApply:  true,
				})
				break
			}

			// Check if larger runners could help
			steps, ok := jd["steps"].([]interface{})
			if ok && len(steps) > 10 {
				result.Optimizations = append(result.Optimizations, Optimization{
					Category:    "runners",
					Title:       "Consider larger runners for complex jobs",
					Description: "Jobs with many steps may benefit from larger runners (GitHub Team/Enterprise)",
					Impact:      "medium",
					AutoApply:   false,
				})
				break
			}
		}
	}
}

// checkConditionalExecution checks for opportunities to skip unnecessary runs
func (o *Optimizer) checkConditionalExecution(content []byte, result *OptimizationResult) {
	contentStr := string(content)

	// Check for path filters
	if !strings.Contains(contentStr, "paths:") && !strings.Contains(contentStr, "paths-ignore:") {
		result.Optimizations = append(result.Optimizations, Optimization{
			Category:    "conditional",
			Title:       "Add path filters to triggers",
			Description: "Adding path filters can skip workflow runs when irrelevant files change (e.g., docs)",
			Impact:      "medium",
			Before: `on:
  push:
    branches: [main]`,
			After: `on:
  push:
    branches: [main]
    paths-ignore:
      - '**.md'
      - 'docs/**'`,
			AutoApply: true,
		})
	}

	// Check for skip ci pattern
	if !strings.Contains(contentStr, "[skip ci]") && !strings.Contains(contentStr, "contains") {
		result.Optimizations = append(result.Optimizations, Optimization{
			Category:    "conditional",
			Title:       "Support [skip ci] in commit messages",
			Description: "Allow skipping CI for documentation-only changes",
			Impact:      "low",
			AutoApply:   false,
		})
	}
}

// analyzeGitLab analyzes GitLab CI configs
func (o *Optimizer) analyzeGitLab(content []byte, result *OptimizationResult) {
	contentStr := string(content)

	// Check for caching
	if !strings.Contains(contentStr, "cache:") {
		result.Optimizations = append(result.Optimizations, Optimization{
			Category:    "caching",
			Title:       "Add dependency caching",
			Description: "GitLab CI supports caching dependencies between pipeline runs",
			Impact:      "high",
			AutoApply:   true,
		})
	}

	// Check for rules vs only/except
	if strings.Contains(contentStr, "only:") || strings.Contains(contentStr, "except:") {
		result.Optimizations = append(result.Optimizations, Optimization{
			Category:    "modernization",
			Title:       "Migrate from only/except to rules",
			Description: "'only' and 'except' are deprecated. Use 'rules' for more flexibility",
			Impact:      "low",
			AutoApply:   true,
		})
	}

	// Check for needs keyword for DAG
	if !strings.Contains(contentStr, "needs:") {
		result.Optimizations = append(result.Optimizations, Optimization{
			Category:    "parallelization",
			Title:       "Use 'needs' for parallel execution",
			Description: "Without 'needs', jobs run sequentially by stage. 'needs' enables DAG-based parallel execution",
			Impact:      "high",
			AutoApply:   false,
		})
	}
}

// analyzeGeneric performs generic optimization analysis
func (o *Optimizer) analyzeGeneric(content []byte, result *OptimizationResult) {
	contentStr := string(content)

	// Check for common anti-patterns
	if strings.Contains(contentStr, "npm install") && !strings.Contains(contentStr, "npm ci") {
		result.Optimizations = append(result.Optimizations, Optimization{
			Category:    "reliability",
			Title:       "Use 'npm ci' instead of 'npm install'",
			Description: "'npm ci' is faster and more reliable for CI environments",
			Impact:      "medium",
			Before:     "npm install",
			After:      "npm ci",
			AutoApply:  true,
		})
	}
}

func (o *Optimizer) estimateTotalSave(opts []Optimization) string {
	// Simple estimation based on impact levels
	highCount := 0
	mediumCount := 0

	for _, opt := range opts {
		switch opt.Impact {
		case "high":
			highCount++
		case "medium":
			mediumCount++
		}
	}

	if highCount > 2 {
		return "2-5 minutes per run"
	} else if highCount > 0 || mediumCount > 2 {
		return "1-2 minutes per run"
	} else if mediumCount > 0 {
		return "30-60 seconds per run"
	}
	return "< 30 seconds per run"
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
	default:
		return "unknown"
	}
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

// PrintReport outputs a formatted optimization report
func (r *OptimizationResult) PrintReport() {
	fmt.Printf("\n⚡ Optimization Report: %s\n", r.File)
	fmt.Printf("   Platform: %s\n", r.Platform)
	fmt.Printf("   Potential time savings: %s\n", r.PotentialSave)
	fmt.Println(strings.Repeat("─", 50))

	if len(r.Optimizations) == 0 {
		fmt.Println("   ✅ No optimizations needed - your pipeline looks great!")
		return
	}

	// Group by impact
	high := []Optimization{}
	medium := []Optimization{}
	low := []Optimization{}

	for _, opt := range r.Optimizations {
		switch opt.Impact {
		case "high":
			high = append(high, opt)
		case "medium":
			medium = append(medium, opt)
		case "low":
			low = append(low, opt)
		}
	}

	if len(high) > 0 {
		fmt.Println("\n   🔴 High Impact:")
		for _, opt := range high {
			printOptimization(opt)
		}
	}

	if len(medium) > 0 {
		fmt.Println("\n   🟡 Medium Impact:")
		for _, opt := range medium {
			printOptimization(opt)
		}
	}

	if len(low) > 0 {
		fmt.Println("\n   🟢 Low Impact:")
		for _, opt := range low {
			printOptimization(opt)
		}
	}

	fmt.Println()
}

func printOptimization(opt Optimization) {
	autoFix := ""
	if opt.AutoApply {
		autoFix = " [auto-fixable]"
	}
	fmt.Printf("      • %s%s\n", opt.Title, autoFix)
	fmt.Printf("        %s\n", opt.Description)
	if opt.EstimatedSave != "" {
		fmt.Printf("        💨 Estimated save: %s\n", opt.EstimatedSave)
	}
}

// Apply applies auto-fixable optimizations
func (o *Optimizer) Apply(filePath string, result *OptimizationResult) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	contentStr := string(content)
	modified := false

	for _, opt := range result.Optimizations {
		if opt.AutoApply && opt.Before != "" && opt.After != "" {
			if strings.Contains(contentStr, opt.Before) {
				contentStr = strings.Replace(contentStr, opt.Before, opt.After, 1)
				modified = true
				fmt.Printf("   ✅ Applied: %s\n", opt.Title)
			}
		}
	}

	if modified {
		if err := os.WriteFile(filePath, []byte(contentStr), 0644); err != nil {
			return err
		}
	}

	return nil
}
