package converter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Platform represents a CI/CD platform
type Platform string

const (
	GitHub    Platform = "github"
	GitLab    Platform = "gitlab"
	Jenkins   Platform = "jenkins"
	CircleCI  Platform = "circleci"
	Azure     Platform = "azure"
	Bitbucket Platform = "bitbucket"
)

// PipelineConfig represents a normalized pipeline configuration
type PipelineConfig struct {
	Name        string            `yaml:"name"`
	Triggers    []Trigger         `yaml:"triggers"`
	Environment map[string]string `yaml:"environment,omitempty"`
	Jobs        []Job             `yaml:"jobs"`
}

// Trigger represents what triggers the pipeline
type Trigger struct {
	Type     string   `yaml:"type"` // push, pull_request, schedule, manual
	Branches []string `yaml:"branches,omitempty"`
	Paths    []string `yaml:"paths,omitempty"`
	Cron     string   `yaml:"cron,omitempty"`
}

// Job represents a pipeline job
type Job struct {
	Name        string            `yaml:"name"`
	RunsOn      string            `yaml:"runs_on"`
	DependsOn   []string          `yaml:"depends_on,omitempty"`
	Environment map[string]string `yaml:"environment,omitempty"`
	Services    []Service         `yaml:"services,omitempty"`
	Steps       []Step            `yaml:"steps"`
	Artifacts   []Artifact        `yaml:"artifacts,omitempty"`
	Cache       []Cache           `yaml:"cache,omitempty"`
	Condition   string            `yaml:"condition,omitempty"`
}

// Service represents a service container
type Service struct {
	Name  string            `yaml:"name"`
	Image string            `yaml:"image"`
	Ports []string          `yaml:"ports,omitempty"`
	Env   map[string]string `yaml:"env,omitempty"`
}

// Step represents a pipeline step
type Step struct {
	Name    string            `yaml:"name"`
	Uses    string            `yaml:"uses,omitempty"`    // For actions/plugins
	Run     string            `yaml:"run,omitempty"`     // For shell commands
	With    map[string]string `yaml:"with,omitempty"`    // Action inputs
	Env     map[string]string `yaml:"env,omitempty"`
	If      string            `yaml:"if,omitempty"`
	WorkDir string            `yaml:"working_directory,omitempty"`
}

// Artifact represents build artifacts
type Artifact struct {
	Name  string   `yaml:"name"`
	Paths []string `yaml:"paths"`
}

// Cache represents cached directories
type Cache struct {
	Key   string   `yaml:"key"`
	Paths []string `yaml:"paths"`
}

// Converter handles pipeline conversions
type Converter struct{}

// NewConverter creates a new converter instance
func NewConverter() *Converter {
	return &Converter{}
}

// Convert converts between CI/CD platforms
func (c *Converter) Convert(from, to Platform, inputPath, outputPath string) error {
	// Parse input file
	config, err := c.Parse(from, inputPath)
	if err != nil {
		return fmt.Errorf("failed to parse %s config: %w", from, err)
	}

	// Generate output
	output, err := c.Generate(to, config)
	if err != nil {
		return fmt.Errorf("failed to generate %s config: %w", to, err)
	}

	// Write output
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(outputPath, []byte(output), 0644); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	fmt.Printf("✅ Converted %s → %s\n", from, to)
	fmt.Printf("   Output: %s\n", outputPath)
	return nil
}

// Parse parses a CI config file into normalized format
func (c *Converter) Parse(platform Platform, inputPath string) (*PipelineConfig, error) {
	content, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, err
	}

	switch platform {
	case GitHub:
		return c.parseGitHub(content)
	case GitLab:
		return c.parseGitLab(content)
	case CircleCI:
		return c.parseCircleCI(content)
	case Jenkins:
		return c.parseJenkins(content)
	default:
		return nil, fmt.Errorf("unsupported source platform: %s", platform)
	}
}

// Generate generates a CI config from normalized format
func (c *Converter) Generate(platform Platform, config *PipelineConfig) (string, error) {
	switch platform {
	case GitHub:
		return c.generateGitHub(config)
	case GitLab:
		return c.generateGitLab(config)
	case CircleCI:
		return c.generateCircleCI(config)
	case Azure:
		return c.generateAzure(config)
	case Jenkins:
		return c.generateJenkins(config)
	default:
		return "", fmt.Errorf("unsupported target platform: %s", platform)
	}
}

// parseGitHub parses GitHub Actions workflow
func (c *Converter) parseGitHub(content []byte) (*PipelineConfig, error) {
	var gh map[string]interface{}
	if err := yaml.Unmarshal(content, &gh); err != nil {
		return nil, err
	}

	config := &PipelineConfig{
		Name:     getString(gh, "name"),
		Triggers: []Trigger{},
		Jobs:     []Job{},
	}

	// Parse triggers
	if on, ok := gh["on"].(map[string]interface{}); ok {
		if push, ok := on["push"].(map[string]interface{}); ok {
			trigger := Trigger{Type: "push"}
			if branches, ok := push["branches"].([]interface{}); ok {
				for _, b := range branches {
					trigger.Branches = append(trigger.Branches, fmt.Sprint(b))
				}
			}
			config.Triggers = append(config.Triggers, trigger)
		}
		if pr, ok := on["pull_request"].(map[string]interface{}); ok {
			trigger := Trigger{Type: "pull_request"}
			if branches, ok := pr["branches"].([]interface{}); ok {
				for _, b := range branches {
					trigger.Branches = append(trigger.Branches, fmt.Sprint(b))
				}
			}
			config.Triggers = append(config.Triggers, trigger)
		}
	}

	// Parse jobs
	if jobs, ok := gh["jobs"].(map[string]interface{}); ok {
		for jobName, jobData := range jobs {
			if jd, ok := jobData.(map[string]interface{}); ok {
				job := Job{
					Name:   jobName,
					RunsOn: getString(jd, "runs-on"),
					Steps:  []Step{},
				}

				if needs, ok := jd["needs"].([]interface{}); ok {
					for _, n := range needs {
						job.DependsOn = append(job.DependsOn, fmt.Sprint(n))
					}
				}

				if steps, ok := jd["steps"].([]interface{}); ok {
					for _, s := range steps {
						if sd, ok := s.(map[string]interface{}); ok {
							step := Step{
								Name: getString(sd, "name"),
								Uses: getString(sd, "uses"),
								Run:  getString(sd, "run"),
								If:   getString(sd, "if"),
							}
							if with, ok := sd["with"].(map[string]interface{}); ok {
								step.With = make(map[string]string)
								for k, v := range with {
									step.With[k] = fmt.Sprint(v)
								}
							}
							job.Steps = append(job.Steps, step)
						}
					}
				}

				config.Jobs = append(config.Jobs, job)
			}
		}
	}

	return config, nil
}

// parseGitLab parses GitLab CI config
func (c *Converter) parseGitLab(content []byte) (*PipelineConfig, error) {
	var gl map[string]interface{}
	if err := yaml.Unmarshal(content, &gl); err != nil {
		return nil, err
	}

	config := &PipelineConfig{
		Name:     "Pipeline",
		Triggers: []Trigger{{Type: "push", Branches: []string{"main"}}},
		Jobs:     []Job{},
	}

	// Parse stages and jobs
	for key, value := range gl {
		// Skip reserved keywords
		if key == "stages" || key == "variables" || key == "image" || key == "default" {
			continue
		}

		if jd, ok := value.(map[string]interface{}); ok {
			job := Job{
				Name:   key,
				RunsOn: "ubuntu-latest",
				Steps:  []Step{},
			}

			// Parse script
			if script, ok := jd["script"].([]interface{}); ok {
				for _, s := range script {
					job.Steps = append(job.Steps, Step{
						Run: fmt.Sprint(s),
					})
				}
			}

			// Parse dependencies
			if needs, ok := jd["needs"].([]interface{}); ok {
				for _, n := range needs {
					job.DependsOn = append(job.DependsOn, fmt.Sprint(n))
				}
			}

			// Parse rules/conditions
			if rules, ok := jd["rules"].([]interface{}); ok {
				for _, r := range rules {
					if rd, ok := r.(map[string]interface{}); ok {
						if ifCond, ok := rd["if"].(string); ok {
							job.Condition = ifCond
							break
						}
					}
				}
			}

			config.Jobs = append(config.Jobs, job)
		}
	}

	return config, nil
}

// parseCircleCI parses CircleCI config
func (c *Converter) parseCircleCI(content []byte) (*PipelineConfig, error) {
	var ci map[string]interface{}
	if err := yaml.Unmarshal(content, &ci); err != nil {
		return nil, err
	}

	config := &PipelineConfig{
		Name:     "Pipeline",
		Triggers: []Trigger{{Type: "push"}},
		Jobs:     []Job{},
	}

	// Parse jobs
	if jobs, ok := ci["jobs"].(map[string]interface{}); ok {
		for jobName, jobData := range jobs {
			if jd, ok := jobData.(map[string]interface{}); ok {
				job := Job{
					Name:   jobName,
					RunsOn: "ubuntu-latest",
					Steps:  []Step{},
				}

				// Parse docker executor
				if docker, ok := jd["docker"].([]interface{}); ok {
					if len(docker) > 0 {
						if d, ok := docker[0].(map[string]interface{}); ok {
							if image, ok := d["image"].(string); ok {
								job.RunsOn = image
							}
						}
					}
				}

				// Parse steps
				if steps, ok := jd["steps"].([]interface{}); ok {
					for _, s := range steps {
						switch st := s.(type) {
						case string:
							if st == "checkout" {
								job.Steps = append(job.Steps, Step{
									Name: "Checkout",
									Uses: "actions/checkout@v4",
								})
							}
						case map[string]interface{}:
							if run, ok := st["run"].(map[string]interface{}); ok {
								job.Steps = append(job.Steps, Step{
									Name: getString(run, "name"),
									Run:  getString(run, "command"),
								})
							}
						}
					}
				}

				config.Jobs = append(config.Jobs, job)
			}
		}
	}

	return config, nil
}

// parseJenkins parses Jenkinsfile (basic support)
func (c *Converter) parseJenkins(content []byte) (*PipelineConfig, error) {
	// Jenkins uses Groovy DSL, so we do basic pattern matching
	contentStr := string(content)

	config := &PipelineConfig{
		Name:     "Pipeline",
		Triggers: []Trigger{{Type: "push"}},
		Jobs:     []Job{},
	}

	// Extract stages
	stagePattern := `stage\s*\(['"]([^'"]+)['"]\)`
	// This is simplified - real Jenkins parsing would need a proper parser

	job := Job{
		Name:   "build",
		RunsOn: "ubuntu-latest",
		Steps:  []Step{},
	}

	// Look for sh commands
	lines := strings.Split(contentStr, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "sh ") || strings.HasPrefix(line, "sh(") {
			cmd := strings.TrimPrefix(line, "sh ")
			cmd = strings.Trim(cmd, "'\"()")
			job.Steps = append(job.Steps, Step{Run: cmd})
		}
	}

	if len(job.Steps) == 0 {
		job.Steps = append(job.Steps, Step{
			Name: "Build",
			Run:  "echo 'Converted from Jenkins - please review'",
		})
	}

	_ = stagePattern // Suppress unused variable warning
	config.Jobs = append(config.Jobs, job)
	return config, nil
}

// generateGitHub generates GitHub Actions workflow
func (c *Converter) generateGitHub(config *PipelineConfig) (string, error) {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("name: %s\n\n", config.Name))

	// Generate triggers
	sb.WriteString("on:\n")
	for _, trigger := range config.Triggers {
		switch trigger.Type {
		case "push":
			sb.WriteString("  push:\n")
			if len(trigger.Branches) > 0 {
				sb.WriteString("    branches:\n")
				for _, b := range trigger.Branches {
					sb.WriteString(fmt.Sprintf("      - %s\n", b))
				}
			}
		case "pull_request":
			sb.WriteString("  pull_request:\n")
			if len(trigger.Branches) > 0 {
				sb.WriteString("    branches:\n")
				for _, b := range trigger.Branches {
					sb.WriteString(fmt.Sprintf("      - %s\n", b))
				}
			}
		case "schedule":
			sb.WriteString("  schedule:\n")
			sb.WriteString(fmt.Sprintf("    - cron: '%s'\n", trigger.Cron))
		}
	}

	sb.WriteString("\njobs:\n")

	// Generate jobs
	for _, job := range config.Jobs {
		sb.WriteString(fmt.Sprintf("  %s:\n", sanitizeName(job.Name)))
		sb.WriteString(fmt.Sprintf("    runs-on: %s\n", job.RunsOn))

		if len(job.DependsOn) > 0 {
			sb.WriteString("    needs:\n")
			for _, dep := range job.DependsOn {
				sb.WriteString(fmt.Sprintf("      - %s\n", dep))
			}
		}

		if job.Condition != "" {
			sb.WriteString(fmt.Sprintf("    if: %s\n", convertCondition(job.Condition, GitHub)))
		}

		sb.WriteString("    steps:\n")
		
		// Always add checkout first if not present
		hasCheckout := false
		for _, step := range job.Steps {
			if strings.Contains(step.Uses, "checkout") {
				hasCheckout = true
				break
			}
		}
		if !hasCheckout {
			sb.WriteString("      - uses: actions/checkout@v4\n")
		}

		for _, step := range job.Steps {
			if step.Name != "" {
				sb.WriteString(fmt.Sprintf("      - name: %s\n", step.Name))
			} else {
				sb.WriteString("      -")
			}

			if step.Uses != "" {
				if step.Name != "" {
					sb.WriteString(fmt.Sprintf("        uses: %s\n", step.Uses))
				} else {
					sb.WriteString(fmt.Sprintf(" uses: %s\n", step.Uses))
				}
				if len(step.With) > 0 {
					sb.WriteString("        with:\n")
					for k, v := range step.With {
						sb.WriteString(fmt.Sprintf("          %s: %s\n", k, v))
					}
				}
			} else if step.Run != "" {
				if step.Name != "" {
					sb.WriteString(fmt.Sprintf("        run: %s\n", step.Run))
				} else {
					sb.WriteString(fmt.Sprintf(" run: %s\n", step.Run))
				}
			}

			if step.If != "" {
				sb.WriteString(fmt.Sprintf("        if: %s\n", step.If))
			}
		}
	}

	return sb.String(), nil
}

// generateGitLab generates GitLab CI config
func (c *Converter) generateGitLab(config *PipelineConfig) (string, error) {
	var sb strings.Builder

	// Generate stages
	sb.WriteString("stages:\n")
	for _, job := range config.Jobs {
		sb.WriteString(fmt.Sprintf("  - %s\n", sanitizeName(job.Name)))
	}
	sb.WriteString("\n")

	// Generate jobs
	for _, job := range config.Jobs {
		sb.WriteString(fmt.Sprintf("%s:\n", sanitizeName(job.Name)))
		sb.WriteString(fmt.Sprintf("  stage: %s\n", sanitizeName(job.Name)))

		if len(job.DependsOn) > 0 {
			sb.WriteString("  needs:\n")
			for _, dep := range job.DependsOn {
				sb.WriteString(fmt.Sprintf("    - %s\n", dep))
			}
		}

		if job.Condition != "" {
			sb.WriteString("  rules:\n")
			sb.WriteString(fmt.Sprintf("    - if: %s\n", convertCondition(job.Condition, GitLab)))
		}

		sb.WriteString("  script:\n")
		for _, step := range job.Steps {
			if step.Run != "" {
				sb.WriteString(fmt.Sprintf("    - %s\n", step.Run))
			} else if step.Uses != "" {
				// Convert common actions to commands
				cmd := convertActionToCommand(step)
				if cmd != "" {
					sb.WriteString(fmt.Sprintf("    - %s\n", cmd))
				}
			}
		}
		sb.WriteString("\n")
	}

	return sb.String(), nil
}

// generateCircleCI generates CircleCI config
func (c *Converter) generateCircleCI(config *PipelineConfig) (string, error) {
	var sb strings.Builder

	sb.WriteString("version: 2.1\n\n")
	sb.WriteString("jobs:\n")

	for _, job := range config.Jobs {
		sb.WriteString(fmt.Sprintf("  %s:\n", sanitizeName(job.Name)))
		sb.WriteString("    docker:\n")
		sb.WriteString("      - image: cimg/base:stable\n")
		sb.WriteString("    steps:\n")
		sb.WriteString("      - checkout\n")

		for _, step := range job.Steps {
			if step.Run != "" {
				sb.WriteString("      - run:\n")
				if step.Name != "" {
					sb.WriteString(fmt.Sprintf("          name: %s\n", step.Name))
				}
				sb.WriteString(fmt.Sprintf("          command: %s\n", step.Run))
			}
		}
	}

	sb.WriteString("\nworkflows:\n")
	sb.WriteString(fmt.Sprintf("  %s:\n", sanitizeName(config.Name)))
	sb.WriteString("    jobs:\n")
	for _, job := range config.Jobs {
		if len(job.DependsOn) > 0 {
			sb.WriteString(fmt.Sprintf("      - %s:\n", sanitizeName(job.Name)))
			sb.WriteString("          requires:\n")
			for _, dep := range job.DependsOn {
				sb.WriteString(fmt.Sprintf("            - %s\n", dep))
			}
		} else {
			sb.WriteString(fmt.Sprintf("      - %s\n", sanitizeName(job.Name)))
		}
	}

	return sb.String(), nil
}

// generateAzure generates Azure Pipelines config
func (c *Converter) generateAzure(config *PipelineConfig) (string, error) {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("name: %s\n\n", config.Name))

	// Triggers
	sb.WriteString("trigger:\n")
	for _, trigger := range config.Triggers {
		if trigger.Type == "push" && len(trigger.Branches) > 0 {
			sb.WriteString("  branches:\n")
			sb.WriteString("    include:\n")
			for _, b := range trigger.Branches {
				sb.WriteString(fmt.Sprintf("      - %s\n", b))
			}
		}
	}

	sb.WriteString("\npool:\n")
	sb.WriteString("  vmImage: 'ubuntu-latest'\n\n")

	sb.WriteString("stages:\n")
	for _, job := range config.Jobs {
		sb.WriteString(fmt.Sprintf("  - stage: %s\n", sanitizeName(job.Name)))
		sb.WriteString("    jobs:\n")
		sb.WriteString(fmt.Sprintf("      - job: %s\n", sanitizeName(job.Name)))

		if len(job.DependsOn) > 0 {
			sb.WriteString("        dependsOn:\n")
			for _, dep := range job.DependsOn {
				sb.WriteString(fmt.Sprintf("          - %s\n", dep))
			}
		}

		sb.WriteString("        steps:\n")
		sb.WriteString("          - checkout: self\n")

		for _, step := range job.Steps {
			if step.Run != "" {
				sb.WriteString("          - script: |\n")
				sb.WriteString(fmt.Sprintf("              %s\n", step.Run))
				if step.Name != "" {
					sb.WriteString(fmt.Sprintf("            displayName: '%s'\n", step.Name))
				}
			}
		}
	}

	return sb.String(), nil
}

// generateJenkins generates Jenkinsfile
func (c *Converter) generateJenkins(config *PipelineConfig) (string, error) {
	var sb strings.Builder

	sb.WriteString("pipeline {\n")
	sb.WriteString("    agent any\n\n")
	sb.WriteString("    stages {\n")

	for _, job := range config.Jobs {
		sb.WriteString(fmt.Sprintf("        stage('%s') {\n", job.Name))
		sb.WriteString("            steps {\n")

		for _, step := range job.Steps {
			if step.Run != "" {
				sb.WriteString(fmt.Sprintf("                sh '%s'\n", escapeJenkinsString(step.Run)))
			}
		}

		sb.WriteString("            }\n")
		sb.WriteString("        }\n")
	}

	sb.WriteString("    }\n")
	sb.WriteString("}\n")

	return sb.String(), nil
}

// Helper functions

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func sanitizeName(name string) string {
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "_", "-")
	return strings.ToLower(name)
}

func convertCondition(condition string, target Platform) string {
	// Basic condition conversion
	switch target {
	case GitHub:
		return condition
	case GitLab:
		return strings.ReplaceAll(condition, "github.", "$CI_")
	default:
		return condition
	}
}

func convertActionToCommand(step Step) string {
	// Convert common GitHub Actions to shell commands
	if strings.Contains(step.Uses, "checkout") {
		return "" // GitLab handles checkout automatically
	}
	if strings.Contains(step.Uses, "setup-node") {
		version := step.With["node-version"]
		if version == "" {
			version = "18"
		}
		return fmt.Sprintf("nvm install %s && nvm use %s", version, version)
	}
	if strings.Contains(step.Uses, "setup-go") {
		return "# Go setup - configure in image"
	}
	if strings.Contains(step.Uses, "setup-python") {
		version := step.With["python-version"]
		if version == "" {
			version = "3.11"
		}
		return fmt.Sprintf("pyenv install %s && pyenv global %s", version, version)
	}
	return fmt.Sprintf("# Action: %s (manual conversion needed)", step.Uses)
}

func escapeJenkinsString(s string) string {
	s = strings.ReplaceAll(s, "'", "\\'")
	return s
}

// GetSupportedPlatforms returns list of supported platforms
func GetSupportedPlatforms() []Platform {
	return []Platform{GitHub, GitLab, Jenkins, CircleCI, Azure, Bitbucket}
}

// DetectPlatform detects CI platform from file path
func DetectPlatform(path string) Platform {
	switch {
	case strings.Contains(path, ".github/workflows"):
		return GitHub
	case strings.Contains(path, ".gitlab-ci"):
		return GitLab
	case strings.Contains(path, "Jenkinsfile"):
		return Jenkins
	case strings.Contains(path, ".circleci"):
		return CircleCI
	case strings.Contains(path, "azure-pipelines"):
		return Azure
	case strings.Contains(path, "bitbucket-pipelines"):
		return Bitbucket
	default:
		return ""
	}
}
