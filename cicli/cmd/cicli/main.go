package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"cicli/internal/analyzer"
	"cicli/internal/config"
	"cicli/internal/converter"
	"cicli/internal/deploy"
	"cicli/internal/docker"
	"cicli/internal/generator"
	"cicli/internal/linter"
	"cicli/internal/notify"
	"cicli/internal/optimizer"
	"cicli/internal/store"
	"cicli/internal/validator"
)

const version = "2.0.0"

func main() {
	printBanner()

	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "version", "-v", "--version":
		fmt.Printf("cicli version %s\n", version)

	case "help", "-h", "--help":
		printHelp()

	case "init":
		handleInit()

	case "analyze":
		handleAnalyze()

	case "generate":
		handleGenerate()

	case "convert":
		handleConvert()

	case "lint":
		handleLint()

	case "optimize":
		handleOptimize()

	case "docker":
		handleDocker()

	case "deploy":
		handleDeploy()

	case "rollback":
		handleRollback()

	case "history":
		handleHistory()

	case "notify":
		handleNotify()

	default:
		fmt.Printf("Unknown command: %s\n", command)
		printHelp()
		os.Exit(1)
	}
}

func printBanner() {
	fmt.Println(`
   ______  _   ______  __     ____
  / ____/ (_) / ____/ / /    /  _/
 / /     / / / /     / /     / /  
/ /___  / / / /___  / /___ _/ /   
\____/ /_/  \____/ /_____//___/   
                                  `)
	fmt.Println("CiCLI - Universal CI/CD Toolkit v" + version)
	fmt.Println()
}

func printHelp() {
	fmt.Println(`Usage: cicli <command> [options]

Core Commands:
  init                    Initialize project configuration
  analyze                 Analyze project and detect technologies
  generate                Generate CI/CD pipelines and configs

Pipeline Tools:
  convert                 Convert between CI/CD platforms
  lint                    Lint and validate CI/CD configurations
  optimize                Analyze and optimize pipelines

Deployment:
  docker publish          Build & push Docker images
  deploy                  Deploy to Kubernetes/AWS
  rollback                Rollback to previous version
  history                 View deployment history
  notify                  Send deployment notifications

Examples:
  cicli analyze                              Analyze current project
  cicli generate --platform github           Generate GitHub Actions workflow
  cicli convert --from gitlab --to github    Convert GitLab CI to GitHub Actions
  cicli lint .github/workflows/ci.yml        Lint a workflow file
  cicli optimize .github/workflows/ci.yml    Get optimization suggestions

Options:
  -h, --help      Show this help message
  -v, --version   Show version information`)
}

// handleInit initializes project configuration
func handleInit() {
	if err := config.InitConfig(); err != nil {
		fmt.Printf("Error initializing config: %v\n", err)
		os.Exit(1)
	}
}

// handleAnalyze analyzes the project
func handleAnalyze() {
	path := "."
	if len(os.Args) > 2 {
		path = os.Args[2]
	}

	fmt.Println("🔍 Analyzing project...")

	a := analyzer.NewAnalyzer(path)
	info, err := a.Analyze()
	if err != nil {
		fmt.Printf("Error analyzing project: %v\n", err)
		os.Exit(1)
	}

	info.PrintReport()
}

// handleGenerate generates CI/CD configurations
func handleGenerate() {
	if len(os.Args) < 3 {
		// Smart generate based on analysis
		fmt.Println("🔍 Analyzing project for smart generation...")

		a := analyzer.NewAnalyzer(".")
		info, err := a.Analyze()
		if err != nil {
			fmt.Printf("Error analyzing project: %v\n", err)
			os.Exit(1)
		}

		// Generate based on detected stack
		generateSmartPipeline(info)
		return
	}

	subCmd := os.Args[2]
	switch subCmd {
	case "pipeline", "workflow":
		platform := "github"
		for _, arg := range os.Args[3:] {
			if strings.HasPrefix(arg, "--platform=") {
				platform = strings.TrimPrefix(arg, "--platform=")
			}
		}
		generatePipeline(platform)

	case "dockerfile":
		generateDockerfile()

	case "k8s", "kubernetes":
		generateKubernetes()

	default:
		// Try loading cicli.yaml for traditional generate
		cfg, err := config.LoadConfig("cicli.yaml")
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		gen := generator.NewGenerator()
		if err := gen.Generate(cfg); err != nil {
			fmt.Printf("Error generating pipeline: %v\n", err)
			os.Exit(1)
		}
	}
}

// generateSmartPipeline creates a pipeline based on project analysis
func generateSmartPipeline(info *analyzer.ProjectInfo) {
	fmt.Printf("\n📦 Detected: %s", info.Language)
	if info.Framework != "" {
		fmt.Printf(" (%s)", info.Framework)
	}
	fmt.Println()

	// Create .github/workflows directory
	workflowDir := filepath.Join(".github", "workflows")
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		os.Exit(1)
	}

	// Generate workflow based on detected stack
	workflow := generateWorkflowForStack(info)
	
	outputPath := filepath.Join(workflowDir, "ci.yml")
	if err := os.WriteFile(outputPath, []byte(workflow), 0644); err != nil {
		fmt.Printf("Error writing workflow: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Generated: %s\n", outputPath)
	fmt.Println("\n💡 Tip: Run 'cicli lint' to validate your new workflow")
}

func generateWorkflowForStack(info *analyzer.ProjectInfo) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(`name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

`))

	switch info.Language {
	case "node":
		pm := info.PackageManager
		if pm == "" {
			pm = "npm"
		}
		sb.WriteString(fmt.Sprintf(`      - uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: '%s'

      - name: Install dependencies
        run: %s install

`, pm, pm))
		if info.BuildCommand != "" {
			sb.WriteString(fmt.Sprintf(`      - name: Build
        run: %s

`, info.BuildCommand))
		}
		if info.TestCommand != "" {
			sb.WriteString(fmt.Sprintf(`      - name: Test
        run: %s
`, info.TestCommand))
		}

	case "go":
		sb.WriteString(`      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Build
        run: go build -v ./...

      - name: Test
        run: go test -v ./...
`)

	case "python":
		sb.WriteString(`      - uses: actions/setup-python@v5
        with:
          python-version: '3.12'
          cache: 'pip'

      - name: Install dependencies
        run: |
          python -m pip install --upgrade pip
          pip install -r requirements.txt

      - name: Lint
        run: |
          pip install flake8
          flake8 . --count --select=E9,F63,F7,F82 --show-source --statistics

      - name: Test
        run: pytest
`)

	case "java":
		if info.PackageManager == "maven" {
			sb.WriteString(`      - uses: actions/setup-java@v4
        with:
          java-version: '17'
          distribution: 'temurin'
          cache: 'maven'

      - name: Build
        run: mvn -B package --file pom.xml

      - name: Test
        run: mvn test
`)
		} else {
			sb.WriteString(`      - uses: actions/setup-java@v4
        with:
          java-version: '17'
          distribution: 'temurin'
          cache: 'gradle'

      - name: Build
        run: ./gradlew build

      - name: Test
        run: ./gradlew test
`)
		}

	default:
		sb.WriteString(`      - name: Build
        run: echo "Add your build command here"

      - name: Test
        run: echo "Add your test command here"
`)
	}

	return sb.String()
}

func generatePipeline(platform string) {
	fmt.Printf("Generating %s pipeline...\n", platform)
	
	// First analyze the project
	a := analyzer.NewAnalyzer(".")
	info, _ := a.Analyze()
	
	generateSmartPipeline(info)
}

func generateDockerfile() {
	a := analyzer.NewAnalyzer(".")
	info, _ := a.Analyze()

	dockerfile := generateDockerfileForStack(info)
	
	if err := os.WriteFile("Dockerfile", []byte(dockerfile), 0644); err != nil {
		fmt.Printf("Error writing Dockerfile: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Generated: Dockerfile")
}

func generateDockerfileForStack(info *analyzer.ProjectInfo) string {
	switch info.Language {
	case "node":
		return `# Build stage
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production

# Production stage
FROM node:20-alpine
WORKDIR /app
COPY --from=builder /app/node_modules ./node_modules
COPY . .
EXPOSE 3000
CMD ["node", "index.js"]
`

	case "go":
		return `# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main .

# Production stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]
`

	case "python":
		return `FROM python:3.12-slim
WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY . .
EXPOSE 8000
CMD ["python", "app.py"]
`

	default:
		return `FROM alpine:latest
WORKDIR /app
COPY . .
# Add your build and run commands
CMD ["echo", "Configure your Dockerfile"]
`
	}
}

func generateKubernetes() {
	a := analyzer.NewAnalyzer(".")
	info, _ := a.Analyze()

	// Create k8s directory
	if err := os.MkdirAll("k8s", 0755); err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		os.Exit(1)
	}

	deployment := fmt.Sprintf(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: %s
  labels:
    app: %s
spec:
  replicas: 2
  selector:
    matchLabels:
      app: %s
  template:
    metadata:
      labels:
        app: %s
    spec:
      containers:
        - name: %s
          image: %s:latest
          ports:
            - containerPort: 3000
          resources:
            requests:
              memory: "128Mi"
              cpu: "100m"
            limits:
              memory: "256Mi"
              cpu: "500m"
---
apiVersion: v1
kind: Service
metadata:
  name: %s
spec:
  selector:
    app: %s
  ports:
    - protocol: TCP
      port: 80
      targetPort: 3000
  type: LoadBalancer
`, info.Name, info.Name, info.Name, info.Name, info.Name, info.Name, info.Name, info.Name)

	if err := os.WriteFile("k8s/deployment.yaml", []byte(deployment), 0644); err != nil {
		fmt.Printf("Error writing deployment: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Generated: k8s/deployment.yaml")
}

// handleConvert converts between CI/CD platforms
func handleConvert() {
	var from, to, input, output string

	for i := 2; i < len(os.Args); i++ {
		arg := os.Args[i]
		if strings.HasPrefix(arg, "--from=") {
			from = strings.TrimPrefix(arg, "--from=")
		} else if strings.HasPrefix(arg, "--to=") {
			to = strings.TrimPrefix(arg, "--to=")
		} else if strings.HasPrefix(arg, "--input=") {
			input = strings.TrimPrefix(arg, "--input=")
		} else if strings.HasPrefix(arg, "--output=") {
			output = strings.TrimPrefix(arg, "--output=")
		}
	}

	if from == "" || to == "" {
		fmt.Println("Usage: cicli convert --from=<platform> --to=<platform> [--input=<file>] [--output=<file>]")
		fmt.Println("\nSupported platforms: github, gitlab, circleci, azure, jenkins, bitbucket")
		fmt.Println("\nExamples:")
		fmt.Println("  cicli convert --from=gitlab --to=github")
		fmt.Println("  cicli convert --from=jenkins --to=github --input=Jenkinsfile")
		os.Exit(1)
	}

	// Auto-detect input file if not specified
	if input == "" {
		input = detectCIFile(converter.Platform(from))
		if input == "" {
			fmt.Printf("Could not find %s CI configuration file\n", from)
			os.Exit(1)
		}
	}

	// Generate output path if not specified
	if output == "" {
		output = getDefaultOutputPath(converter.Platform(to))
	}

	fmt.Printf("🔄 Converting %s → %s\n", from, to)
	fmt.Printf("   Input:  %s\n", input)
	fmt.Printf("   Output: %s\n", output)

	c := converter.NewConverter()
	if err := c.Convert(converter.Platform(from), converter.Platform(to), input, output); err != nil {
		fmt.Printf("Error converting: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n💡 Tip: Run 'cicli lint' to validate the converted workflow")
}

func detectCIFile(platform converter.Platform) string {
	paths := map[converter.Platform][]string{
		converter.GitHub:   {".github/workflows/ci.yml", ".github/workflows/main.yml"},
		converter.GitLab:   {".gitlab-ci.yml"},
		converter.CircleCI: {".circleci/config.yml"},
		converter.Jenkins:  {"Jenkinsfile"},
		converter.Azure:    {"azure-pipelines.yml"},
	}

	if candidates, ok := paths[platform]; ok {
		for _, path := range candidates {
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	}
	return ""
}

func getDefaultOutputPath(platform converter.Platform) string {
	switch platform {
	case converter.GitHub:
		return ".github/workflows/ci.yml"
	case converter.GitLab:
		return ".gitlab-ci.yml"
	case converter.CircleCI:
		return ".circleci/config.yml"
	case converter.Jenkins:
		return "Jenkinsfile"
	case converter.Azure:
		return "azure-pipelines.yml"
	default:
		return "pipeline.yml"
	}
}

// handleLint lints CI/CD configurations
func handleLint() {
	path := "."
	if len(os.Args) > 2 {
		path = os.Args[2]
	}

	l := linter.NewLinter()

	info, err := os.Stat(path)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if info.IsDir() {
		results, err := l.LintDirectory(path)
		if err != nil {
			fmt.Printf("Error linting directory: %v\n", err)
			os.Exit(1)
		}

		if len(results) == 0 {
			fmt.Println("No CI/CD configuration files found")
			os.Exit(0)
		}

		totalIssues := 0
		for _, result := range results {
			result.PrintReport()
			totalIssues += len(result.Issues)
		}

		if totalIssues > 0 {
			os.Exit(1)
		}
	} else {
		result, err := l.Lint(path)
		if err != nil {
			fmt.Printf("Error linting file: %v\n", err)
			os.Exit(1)
		}

		result.PrintReport()

		if len(result.Issues) > 0 {
			os.Exit(1)
		}
	}
}

// handleOptimize analyzes and suggests optimizations
func handleOptimize() {
	path := "."
	apply := false

	for _, arg := range os.Args[2:] {
		if arg == "--apply" {
			apply = true
		} else if !strings.HasPrefix(arg, "-") {
			path = arg
		}
	}

	o := optimizer.NewOptimizer()

	// Check if path is a file or directory
	info, err := os.Stat(path)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if info.IsDir() {
		// Find CI files in directory
		patterns := []string{
			".github/workflows/*.yml",
			".github/workflows/*.yaml",
			".gitlab-ci.yml",
		}

		found := false
		for _, pattern := range patterns {
			matches, _ := filepath.Glob(filepath.Join(path, pattern))
			for _, match := range matches {
				found = true
				analyzeAndOptimize(o, match, apply)
			}
		}

		if !found {
			fmt.Println("No CI/CD configuration files found")
		}
	} else {
		analyzeAndOptimize(o, path, apply)
	}
}

func analyzeAndOptimize(o *optimizer.Optimizer, path string, apply bool) {
	result, err := o.Analyze(path)
	if err != nil {
		fmt.Printf("Error analyzing %s: %v\n", path, err)
		return
	}

	result.PrintReport()

	if apply && len(result.Optimizations) > 0 {
		fmt.Println("\n🔧 Applying auto-fixable optimizations...")
		if err := o.Apply(path, result); err != nil {
			fmt.Printf("Error applying optimizations: %v\n", err)
		}
	}
}

// handleDocker handles docker commands
func handleDocker() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: cicli docker publish [flags]")
		os.Exit(1)
	}

	subCmd := os.Args[2]
	if subCmd != "publish" {
		fmt.Printf("Unknown docker command: %s\n", subCmd)
		os.Exit(1)
	}

	cfg, err := config.LoadConfig("cicli.yaml")
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	tag := "latest"
	useGitSha := false
	for _, arg := range os.Args[3:] {
		if arg == "--use-git-sha" {
			useGitSha = true
		} else if strings.HasPrefix(arg, "--tag=") {
			tag = strings.TrimPrefix(arg, "--tag=")
		}
	}

	if err := validator.CheckDocker(); err != nil {
		fmt.Printf("Pre-flight check failed: %v\n", err)
		os.Exit(1)
	}

	d := docker.NewClient()

	if useGitSha {
		sha, err := d.GetGitSHA()
		if err != nil {
			fmt.Printf("Error getting git SHA: %v\n", err)
			os.Exit(1)
		}
		tag = sha
	}

	fullImageName := fmt.Sprintf("%s:%s", cfg.Docker.ImageName, tag)

	if err := d.Build(fullImageName, cfg.Docker.Context, cfg.Docker.Dockerfile); err != nil {
		fmt.Printf("Error building image: %v\n", err)
		os.Exit(1)
	}

	if err := d.Push(fullImageName); err != nil {
		fmt.Printf("Error pushing image: %v\n", err)
		os.Exit(1)
	}
}

// handleDeploy handles deployment
func handleDeploy() {
	if err := validator.CheckKubectl(); err != nil {
		fmt.Printf("Pre-flight check failed: %v\n", err)
		os.Exit(1)
	}

	cfg, err := config.LoadConfig("cicli.yaml")
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	env := "dev"
	tag := "latest"
	for _, arg := range os.Args[2:] {
		if strings.HasPrefix(arg, "--env=") {
			env = strings.TrimPrefix(arg, "--env=")
		} else if strings.HasPrefix(arg, "--tag=") {
			tag = strings.TrimPrefix(arg, "--tag=")
		}
	}

	dep := deploy.NewDeployer()

	if cfg.Deploy.Provider == "aws" {
		if err := dep.ConfigureEKS(cfg.Deploy.Region, cfg.Deploy.ClusterName); err != nil {
			fmt.Printf("Error configuring EKS: %v\n", err)
			os.Exit(1)
		}
	}

	fullImageName := fmt.Sprintf("%s:%s", cfg.Docker.ImageName, tag)
	appName := cfg.ProjectName

	if err := dep.DeployToK8s(cfg.Deploy.ManifestPath, fullImageName, appName, env); err != nil {
		fmt.Printf("Error deploying: %v\n", err)
		os.Exit(1)
	}
}

// handleRollback handles rollback
func handleRollback() {
	if err := validator.CheckKubectl(); err != nil {
		fmt.Printf("Pre-flight check failed: %v\n", err)
		os.Exit(1)
	}

	cfg, err := config.LoadConfig("cicli.yaml")
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	env := "dev"
	for _, arg := range os.Args[2:] {
		if strings.HasPrefix(arg, "--env=") {
			env = strings.TrimPrefix(arg, "--env=")
		}
	}

	dep := deploy.NewDeployer()
	appName := cfg.ProjectName

	if err := dep.Rollback(appName, env); err != nil {
		fmt.Printf("Error rolling back: %v\n", err)
		os.Exit(1)
	}
}

// handleHistory shows deployment history
func handleHistory() {
	s, err := store.NewStore()
	if err != nil {
		fmt.Printf("Error opening store: %v\n", err)
		os.Exit(1)
	}

	deployments, err := s.Load()
	if err != nil {
		fmt.Printf("Error loading history: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Deployment History:")
	fmt.Printf("%-20s %-15s %-10s %-20s %s\n", "TIMESTAMP", "PROJECT", "ENV", "STATUS", "IMAGE")
	for _, d := range deployments {
		fmt.Printf("%-20s %-15s %-10s %-20s %s\n",
			d.Timestamp.Format("2006-01-02 15:04"),
			d.Project,
			d.Env,
			d.Status,
			d.Image)
	}
}

// handleNotify sends notifications
func handleNotify() {
	cfg, err := config.LoadConfig("cicli.yaml")
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	status := "success"
	env := "dev"
	version := "latest"

	for _, arg := range os.Args[2:] {
		if strings.HasPrefix(arg, "--status=") {
			status = strings.TrimPrefix(arg, "--status=")
		} else if strings.HasPrefix(arg, "--env=") {
			env = strings.TrimPrefix(arg, "--env=")
		} else if strings.HasPrefix(arg, "--version=") {
			version = strings.TrimPrefix(arg, "--version=")
		}
	}

	n := notify.NewNotifier()
	if err := n.Send(cfg.Notifications.WebhookURL, cfg.ProjectName, status, env, version); err != nil {
		fmt.Printf("Error sending notification: %v\n", err)
		os.Exit(1)
	}
}
