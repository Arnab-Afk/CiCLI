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

const workflowTemplate = `name: CI/CD Pipeline

on:
  push:
    branches: [ "main" ]
  pull_request:
    branches: [ "main" ]

jobs:
  build-and-test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3

    {{- if eq .Language "node" }}
    - name: Set up Node.js
      uses: actions/setup-node@v3
      with:
        node-version: '18'
        cache: 'npm'
    - name: Install dependencies
      run: npm ci
    {{- else if eq .Language "go" }}
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'
        cache: true
    {{- else if eq .Language "python" }}
    - name: Set up Python
      uses: actions/setup-python@v4
      with:
        python-version: '3.9'
        cache: 'pip'
    {{- end }}

    - name: Build
      run: {{ .BuildCommand }}

    - name: Test
      run: {{ .TestCommand }}

  deploy:
    needs: build-and-test
    if: github.ref == 'refs/heads/main' && github.event_name == 'push'
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Log in to Docker Hub
      uses: docker/login-action@v2
      with:
        username: ${{ "{{" }} secrets.DOCKER_USERNAME }}
        password: ${{ "{{" }} secrets.DOCKER_PASSWORD }}
    
    - name: Build and Push Docker image
      uses: docker/build-push-action@v4
      with:
        context: {{ .Docker.Context }}
        file: {{ .Docker.Dockerfile }}
        push: true
        tags: {{ .Docker.ImageName }}:${{ "{{" }} github.sha }},{{ .Docker.ImageName }}:latest

    - name: Deploy to Kubernetes
      uses: azure/k8s-set-context@v3
      with:
        method: kubeconfig
        kubeconfig: ${{ "{{" }} secrets.KUBECONFIG }}
        
    - name: Update Deployment
      run: |
        kubectl set image deployment/{{ .ProjectName }} {{ .ProjectName }}={{ .Docker.ImageName }}:${{ "{{" }} github.sha }}
        kubectl rollout status deployment/{{ .ProjectName }}
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

	tmpl, err := template.New("workflow").Parse(workflowTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	if err := tmpl.Execute(f, cfg); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	fmt.Printf("Generated %s\n", filePath)
	return nil
}
