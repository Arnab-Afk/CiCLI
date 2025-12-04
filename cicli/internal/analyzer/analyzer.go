package analyzer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ProjectInfo contains analyzed project information
type ProjectInfo struct {
	Name         string            `json:"name"`
	Language     string            `json:"language"`
	Framework    string            `json:"framework"`
	PackageManager string          `json:"package_manager"`
	BuildCommand string            `json:"build_command"`
	TestCommand  string            `json:"test_command"`
	TestFramework string           `json:"test_framework"`
	HasDocker    bool              `json:"has_docker"`
	HasCI        bool              `json:"has_ci"`
	CIPlatform   string            `json:"ci_platform"`
	Dependencies []string          `json:"dependencies"`
	DevDependencies []string       `json:"dev_dependencies"`
	Ports        []int             `json:"ports"`
	EnvVars      []string          `json:"env_vars"`
	EntryPoint   string            `json:"entry_point"`
	Suggestions  []Suggestion      `json:"suggestions"`
}

// Suggestion represents an improvement suggestion
type Suggestion struct {
	Category    string `json:"category"`
	Severity    string `json:"severity"` // info, warning, critical
	Title       string `json:"title"`
	Description string `json:"description"`
	Fix         string `json:"fix,omitempty"`
}

// Analyzer analyzes project structure and generates insights
type Analyzer struct {
	rootPath string
}

// NewAnalyzer creates a new analyzer instance
func NewAnalyzer(rootPath string) *Analyzer {
	if rootPath == "" {
		rootPath = "."
	}
	return &Analyzer{rootPath: rootPath}
}

// Analyze performs full project analysis
func (a *Analyzer) Analyze() (*ProjectInfo, error) {
	info := &ProjectInfo{
		Suggestions: []Suggestion{},
	}

	// Detect project name from directory
	absPath, _ := filepath.Abs(a.rootPath)
	info.Name = filepath.Base(absPath)

	// Detect language and framework
	a.detectLanguage(info)
	a.detectFramework(info)
	a.detectPackageManager(info)
	a.detectBuildCommands(info)
	a.detectTestFramework(info)
	a.detectDocker(info)
	a.detectCI(info)
	a.detectPorts(info)
	a.detectEnvVars(info)
	a.generateSuggestions(info)

	return info, nil
}

// detectLanguage identifies the primary programming language
func (a *Analyzer) detectLanguage(info *ProjectInfo) {
	checks := []struct {
		file     string
		language string
	}{
		{"package.json", "node"},
		{"go.mod", "go"},
		{"requirements.txt", "python"},
		{"pyproject.toml", "python"},
		{"Pipfile", "python"},
		{"Cargo.toml", "rust"},
		{"pom.xml", "java"},
		{"build.gradle", "java"},
		{"Gemfile", "ruby"},
		{"composer.json", "php"},
		{"*.csproj", "dotnet"},
	}

	for _, check := range checks {
		if check.file == "*.csproj" {
			matches, _ := filepath.Glob(filepath.Join(a.rootPath, check.file))
			if len(matches) > 0 {
				info.Language = check.language
				return
			}
		} else if a.fileExists(check.file) {
			info.Language = check.language
			return
		}
	}

	info.Language = "unknown"
}

// detectFramework identifies the framework being used
func (a *Analyzer) detectFramework(info *ProjectInfo) {
	switch info.Language {
	case "node":
		a.detectNodeFramework(info)
	case "python":
		a.detectPythonFramework(info)
	case "go":
		a.detectGoFramework(info)
	case "java":
		a.detectJavaFramework(info)
	}
}

func (a *Analyzer) detectNodeFramework(info *ProjectInfo) {
	content, err := os.ReadFile(filepath.Join(a.rootPath, "package.json"))
	if err != nil {
		return
	}

	var pkg map[string]interface{}
	if err := json.Unmarshal(content, &pkg); err != nil {
		return
	}

	// Extract dependencies
	if deps, ok := pkg["dependencies"].(map[string]interface{}); ok {
		for dep := range deps {
			info.Dependencies = append(info.Dependencies, dep)
		}
	}
	if devDeps, ok := pkg["devDependencies"].(map[string]interface{}); ok {
		for dep := range devDeps {
			info.DevDependencies = append(info.DevDependencies, dep)
		}
	}

	// Detect framework
	frameworks := map[string]string{
		"next":      "nextjs",
		"react":     "react",
		"vue":       "vue",
		"@angular/core": "angular",
		"express":   "express",
		"fastify":   "fastify",
		"nestjs":    "nestjs",
		"@nestjs/core": "nestjs",
		"koa":       "koa",
		"hapi":      "hapi",
		"svelte":    "svelte",
		"nuxt":      "nuxt",
	}

	for dep, framework := range frameworks {
		for _, d := range info.Dependencies {
			if d == dep {
				info.Framework = framework
				return
			}
		}
	}
}

func (a *Analyzer) detectPythonFramework(info *ProjectInfo) {
	files := []string{"requirements.txt", "pyproject.toml", "Pipfile"}
	
	for _, file := range files {
		content, err := os.ReadFile(filepath.Join(a.rootPath, file))
		if err != nil {
			continue
		}

		contentStr := strings.ToLower(string(content))
		frameworks := map[string]string{
			"django":     "django",
			"flask":      "flask",
			"fastapi":    "fastapi",
			"tornado":    "tornado",
			"pyramid":    "pyramid",
			"starlette":  "starlette",
			"sanic":      "sanic",
		}

		for dep, framework := range frameworks {
			if strings.Contains(contentStr, dep) {
				info.Framework = framework
				return
			}
		}
	}
}

func (a *Analyzer) detectGoFramework(info *ProjectInfo) {
	content, err := os.ReadFile(filepath.Join(a.rootPath, "go.mod"))
	if err != nil {
		return
	}

	contentStr := string(content)
	frameworks := map[string]string{
		"github.com/gin-gonic/gin":   "gin",
		"github.com/labstack/echo":   "echo",
		"github.com/gofiber/fiber":   "fiber",
		"github.com/gorilla/mux":     "gorilla",
		"github.com/go-chi/chi":      "chi",
		"github.com/beego/beego":     "beego",
	}

	for dep, framework := range frameworks {
		if strings.Contains(contentStr, dep) {
			info.Framework = framework
			return
		}
	}
}

func (a *Analyzer) detectJavaFramework(info *ProjectInfo) {
	files := []string{"pom.xml", "build.gradle"}
	
	for _, file := range files {
		content, err := os.ReadFile(filepath.Join(a.rootPath, file))
		if err != nil {
			continue
		}

		contentStr := string(content)
		if strings.Contains(contentStr, "spring-boot") {
			info.Framework = "spring-boot"
			return
		}
		if strings.Contains(contentStr, "quarkus") {
			info.Framework = "quarkus"
			return
		}
		if strings.Contains(contentStr, "micronaut") {
			info.Framework = "micronaut"
			return
		}
	}
}

// detectPackageManager identifies the package manager
func (a *Analyzer) detectPackageManager(info *ProjectInfo) {
	switch info.Language {
	case "node":
		if a.fileExists("pnpm-lock.yaml") {
			info.PackageManager = "pnpm"
		} else if a.fileExists("yarn.lock") {
			info.PackageManager = "yarn"
		} else if a.fileExists("bun.lockb") {
			info.PackageManager = "bun"
		} else {
			info.PackageManager = "npm"
		}
	case "python":
		if a.fileExists("poetry.lock") {
			info.PackageManager = "poetry"
		} else if a.fileExists("Pipfile.lock") {
			info.PackageManager = "pipenv"
		} else if a.fileExists("uv.lock") {
			info.PackageManager = "uv"
		} else {
			info.PackageManager = "pip"
		}
	case "go":
		info.PackageManager = "go mod"
	case "rust":
		info.PackageManager = "cargo"
	case "java":
		if a.fileExists("pom.xml") {
			info.PackageManager = "maven"
		} else if a.fileExists("build.gradle") {
			info.PackageManager = "gradle"
		}
	}
}

// detectBuildCommands determines build and start commands
func (a *Analyzer) detectBuildCommands(info *ProjectInfo) {
	switch info.Language {
	case "node":
		content, err := os.ReadFile(filepath.Join(a.rootPath, "package.json"))
		if err != nil {
			return
		}
		var pkg map[string]interface{}
		if err := json.Unmarshal(content, &pkg); err != nil {
			return
		}

		if scripts, ok := pkg["scripts"].(map[string]interface{}); ok {
			if _, ok := scripts["build"]; ok {
				info.BuildCommand = fmt.Sprintf("%s run build", info.PackageManager)
			}
			if _, ok := scripts["test"]; ok {
				info.TestCommand = fmt.Sprintf("%s test", info.PackageManager)
			}
			if main, ok := pkg["main"].(string); ok {
				info.EntryPoint = main
			}
		}

	case "go":
		info.BuildCommand = "go build -o app ."
		info.TestCommand = "go test ./..."
		// Find main.go
		filepath.Walk(a.rootPath, func(path string, f os.FileInfo, err error) error {
			if filepath.Base(path) == "main.go" {
				info.EntryPoint = path
				return filepath.SkipAll
			}
			return nil
		})

	case "python":
		info.BuildCommand = "pip install -r requirements.txt"
		info.TestCommand = "pytest"
		// Find entry point
		for _, entry := range []string{"app.py", "main.py", "run.py", "manage.py"} {
			if a.fileExists(entry) {
				info.EntryPoint = entry
				break
			}
		}

	case "java":
		if info.PackageManager == "maven" {
			info.BuildCommand = "mvn clean package"
			info.TestCommand = "mvn test"
		} else {
			info.BuildCommand = "gradle build"
			info.TestCommand = "gradle test"
		}
	}
}

// detectTestFramework identifies the test framework
func (a *Analyzer) detectTestFramework(info *ProjectInfo) {
	switch info.Language {
	case "node":
		testFrameworks := map[string]string{
			"jest":       "jest",
			"mocha":      "mocha",
			"vitest":     "vitest",
			"ava":        "ava",
			"tap":        "tap",
			"@playwright/test": "playwright",
			"cypress":    "cypress",
		}
		for dep, framework := range testFrameworks {
			for _, d := range info.DevDependencies {
				if d == dep {
					info.TestFramework = framework
					return
				}
			}
		}

	case "python":
		if a.fileExists("pytest.ini") || a.fileExists("pyproject.toml") {
			info.TestFramework = "pytest"
		} else {
			info.TestFramework = "unittest"
		}

	case "go":
		info.TestFramework = "go test"

	case "java":
		info.TestFramework = "junit"
	}
}

// detectDocker checks for Docker configuration
func (a *Analyzer) detectDocker(info *ProjectInfo) {
	dockerfiles := []string{"Dockerfile", "dockerfile", "Dockerfile.dev", "Dockerfile.prod"}
	for _, df := range dockerfiles {
		if a.fileExists(df) {
			info.HasDocker = true
			return
		}
	}
	// Check for docker-compose
	composeFiles := []string{"docker-compose.yml", "docker-compose.yaml", "compose.yml", "compose.yaml"}
	for _, cf := range composeFiles {
		if a.fileExists(cf) {
			info.HasDocker = true
			return
		}
	}
}

// detectCI checks for existing CI configuration
func (a *Analyzer) detectCI(info *ProjectInfo) {
	ciConfigs := map[string]string{
		".github/workflows":      "github-actions",
		".gitlab-ci.yml":         "gitlab-ci",
		"Jenkinsfile":            "jenkins",
		".circleci/config.yml":   "circleci",
		"azure-pipelines.yml":    "azure-pipelines",
		".travis.yml":            "travis-ci",
		"bitbucket-pipelines.yml": "bitbucket",
		".drone.yml":             "drone",
	}

	for path, platform := range ciConfigs {
		fullPath := filepath.Join(a.rootPath, path)
		if fileInfo, err := os.Stat(fullPath); err == nil {
			if fileInfo.IsDir() {
				// Check if directory has files
				entries, _ := os.ReadDir(fullPath)
				if len(entries) > 0 {
					info.HasCI = true
					info.CIPlatform = platform
				}
			} else {
				info.HasCI = true
				info.CIPlatform = platform
			}
			return
		}
	}
}

// detectPorts scans for common port definitions
func (a *Analyzer) detectPorts(info *ProjectInfo) {
	portPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)port["\s:=]+(\d{4,5})`),
		regexp.MustCompile(`(?i)listen\s*\(\s*(\d{4,5})`),
		regexp.MustCompile(`(?i)EXPOSE\s+(\d{4,5})`),
	}

	// Common files to check
	files := []string{"Dockerfile", "docker-compose.yml", "package.json", "app.py", "main.go", "server.js", ".env", ".env.example"}

	for _, file := range files {
		content, err := os.ReadFile(filepath.Join(a.rootPath, file))
		if err != nil {
			continue
		}

		for _, pattern := range portPatterns {
			matches := pattern.FindAllStringSubmatch(string(content), -1)
			for _, match := range matches {
				if len(match) > 1 {
					var port int
					fmt.Sscanf(match[1], "%d", &port)
					if port > 0 && port < 65536 {
						// Check for duplicates
						found := false
						for _, p := range info.Ports {
							if p == port {
								found = true
								break
							}
						}
						if !found {
							info.Ports = append(info.Ports, port)
						}
					}
				}
			}
		}
	}
}

// detectEnvVars finds environment variable references
func (a *Analyzer) detectEnvVars(info *ProjectInfo) {
	// Check .env.example or .env.sample
	envFiles := []string{".env.example", ".env.sample", ".env.template"}
	
	for _, envFile := range envFiles {
		content, err := os.ReadFile(filepath.Join(a.rootPath, envFile))
		if err != nil {
			continue
		}

		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) > 0 {
				info.EnvVars = append(info.EnvVars, parts[0])
			}
		}
	}
}

// generateSuggestions creates improvement recommendations
func (a *Analyzer) generateSuggestions(info *ProjectInfo) {
	// Check for missing Dockerfile
	if !info.HasDocker {
		info.Suggestions = append(info.Suggestions, Suggestion{
			Category:    "containerization",
			Severity:    "warning",
			Title:       "Missing Dockerfile",
			Description: "No Dockerfile found. Containerizing your app improves deployment consistency.",
			Fix:         "Run 'cicli generate dockerfile' to create one",
		})
	}

	// Check for missing CI
	if !info.HasCI {
		info.Suggestions = append(info.Suggestions, Suggestion{
			Category:    "ci-cd",
			Severity:    "warning",
			Title:       "No CI/CD Pipeline",
			Description: "No CI/CD configuration detected. Automated testing and deployment improves reliability.",
			Fix:         "Run 'cicli generate pipeline' to create one",
		})
	}

	// Check for missing tests
	if info.TestCommand == "" {
		info.Suggestions = append(info.Suggestions, Suggestion{
			Category:    "testing",
			Severity:    "critical",
			Title:       "No Test Configuration",
			Description: "No test script or framework detected. Tests are essential for code quality.",
			Fix:         "Add a test framework appropriate for your language",
		})
	}

	// Check for .env in .gitignore
	if a.fileExists(".env") && !a.isInGitignore(".env") {
		info.Suggestions = append(info.Suggestions, Suggestion{
			Category:    "security",
			Severity:    "critical",
			Title:       ".env File May Be Exposed",
			Description: ".env file exists but may not be in .gitignore. This could expose secrets.",
			Fix:         "Add .env to your .gitignore file",
		})
	}

	// Check for outdated Node.js version in CI
	if info.Language == "node" && info.HasCI {
		info.Suggestions = append(info.Suggestions, Suggestion{
			Category:    "optimization",
			Severity:    "info",
			Title:       "Consider Node.js Version Matrix",
			Description: "Testing against multiple Node.js versions catches compatibility issues early.",
			Fix:         "Add a matrix strategy to test on Node 18, 20, and 22",
		})
	}
}

// isInGitignore checks if a pattern is in .gitignore
func (a *Analyzer) isInGitignore(pattern string) bool {
	content, err := os.ReadFile(filepath.Join(a.rootPath, ".gitignore"))
	if err != nil {
		return false
	}
	return strings.Contains(string(content), pattern)
}

// fileExists checks if a file exists
func (a *Analyzer) fileExists(name string) bool {
	_, err := os.Stat(filepath.Join(a.rootPath, name))
	return err == nil
}

// PrintReport outputs a formatted analysis report
func (info *ProjectInfo) PrintReport() {
	fmt.Println("\n📊 Project Analysis Report")
	fmt.Println(strings.Repeat("═", 50))
	
	fmt.Printf("\n📦 Project: %s\n", info.Name)
	fmt.Printf("🔧 Language: %s\n", info.Language)
	if info.Framework != "" {
		fmt.Printf("🏗️  Framework: %s\n", info.Framework)
	}
	if info.PackageManager != "" {
		fmt.Printf("📦 Package Manager: %s\n", info.PackageManager)
	}
	
	fmt.Println("\n📋 Commands:")
	if info.BuildCommand != "" {
		fmt.Printf("   Build: %s\n", info.BuildCommand)
	}
	if info.TestCommand != "" {
		fmt.Printf("   Test:  %s\n", info.TestCommand)
	}
	
	fmt.Println("\n🔍 Detection:")
	fmt.Printf("   Docker: %v\n", boolToEmoji(info.HasDocker))
	fmt.Printf("   CI/CD:  %v", boolToEmoji(info.HasCI))
	if info.CIPlatform != "" {
		fmt.Printf(" (%s)", info.CIPlatform)
	}
	fmt.Println()

	if len(info.Ports) > 0 {
		fmt.Printf("   Ports:  %v\n", info.Ports)
	}

	if len(info.Suggestions) > 0 {
		fmt.Println("\n💡 Suggestions:")
		for _, s := range info.Suggestions {
			icon := "ℹ️"
			if s.Severity == "warning" {
				icon = "⚠️"
			} else if s.Severity == "critical" {
				icon = "🚨"
			}
			fmt.Printf("   %s %s\n", icon, s.Title)
			fmt.Printf("      %s\n", s.Description)
			if s.Fix != "" {
				fmt.Printf("      → %s\n", s.Fix)
			}
		}
	}
	
	fmt.Println()
}

func boolToEmoji(b bool) string {
	if b {
		return "✅"
	}
	return "❌"
}
