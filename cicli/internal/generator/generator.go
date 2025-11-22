package generator

import (
	"cicli/internal/config"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

type Generator struct{}

func NewGenerator() *Generator {
	return &Generator{}
}

const githubWorkflowTemplate = `name: CI/CD Pipeline

on:
  push:
    branches: [ "main" ]
  pull_request:
    branches: [ "main" ]

jobs:
  build-test-deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      {{- if eq .Language "node" }}
      - name: Setup Node
        uses: actions/setup-node@v3
        with:
          node-version: 18
      
      - name: Install dependencies
        run: npm install
      {{- else if eq .Language "go" }}
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Install dependencies
        run: go mod download
      {{- else if eq .Language "python" }}
      - name: Setup Python
        uses: actions/setup-python@v4
        with:
          python-version: '3.10'
          
      - name: Install dependencies
        run: pip install -r requirements.txt
      {{- end }}

      - name: Run tests
        run: {{ .TestCommand }}

      # In a real scenario, we would install cicli here or use a docker action.
      # For this demo, we assume cicli is available or we mock the commands.
      # Since cicli is a CLI tool we are building, we can't easily "install" it in the runner 
      # without publishing it first. For the sake of the generated pipeline, 
      # we will assume the user wants to see where 'cicli' commands would go.
      
      # - name: Build Docker image
      #   run: cicli docker publish --use-git-sha

      # - name: Deploy to Kubernetes
      #   run: cicli deploy --env prod

      # - name: Send deployment notification
      #   if: always()
      #   run: cicli notify --status ${{ "{{" }} job.status }} --env prod
`

func (g *Generator) Generate(cfg *config.Config) error {
	fmt.Printf("Generating pipeline for project: %s\n", cfg.ProjectName)

	// Ensure .github/workflows exists
	workflowDir := filepath.Join(".github", "workflows")
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", workflowDir, err)
	}

	filePath := filepath.Join(workflowDir, "ci-cd.yml")
	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filePath, err)
	}
	defer f.Close()

	tmpl, err := template.New("workflow").Parse(githubWorkflowTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	if err := tmpl.Execute(f, cfg); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	fmt.Printf("Generated %s\n", filePath)
	return nil
}
