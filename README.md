# CiCLI

**The Universal CI/CD Toolkit** — Analyze projects, convert between platforms, lint pipelines, and optimize builds.

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## Why CiCLI?

Most CI/CD tools just copy templates. **CiCLI actually understands your project:**

- 🔍 **Analyzes** your codebase to detect language, framework, and dependencies
- 🔄 **Converts** between GitHub Actions, GitLab CI, Jenkins, CircleCI, and Azure Pipelines
- 🔎 **Lints** your pipelines for security issues, best practices, and errors
- ⚡ **Optimizes** build times with caching, parallelization, and smart suggestions

## Installation

```bash
git clone https://github.com/Arnab-Afk/CiCLI.git
cd CiCLI/cicli
go build -o cicli ./cmd/cicli
```

## Features

### 🔍 Smart Project Analysis

```bash
cicli analyze
```

```
📊 Project Analysis Report
══════════════════════════════════════════════════

📦 Project: my-app
🔧 Language: node
🏗️  Framework: nextjs
📦 Package Manager: pnpm

📋 Commands:
   Build: pnpm run build
   Test:  pnpm test

🔍 Detection:
   Docker: ✅
   CI/CD:  ✅ (github-actions)
   Ports:  [3000]

💡 Suggestions:
   ⚠️ Using outdated action versions
      → Update actions/checkout to v4
```

### 🔄 Platform Conversion

Convert between CI/CD platforms instantly:

```bash
# GitLab → GitHub
cicli convert --from=gitlab --to=github

# Jenkins → GitHub Actions
cicli convert --from=jenkins --to=github --input=Jenkinsfile

# CircleCI → Azure Pipelines
cicli convert --from=circleci --to=azure
```

**Supported platforms:** GitHub Actions, GitLab CI, Jenkins, CircleCI, Azure Pipelines, Bitbucket

### 🔎 Pipeline Linting

Catch security issues and anti-patterns:

```bash
cicli lint .github/workflows/ci.yml
```

```
🔍 Lint Report: .github/workflows/ci.yml
   Platform: github
   Score: 70/100
──────────────────────────────────────────────────

   🚨 Errors:
      [SEC001] (line 23) Potential AWS Access Key detected
         → Use secrets/environment variables instead

   ⚠️  Warnings:
      [SEC003] Action 'some/action@v1' uses tag instead of SHA pin
         → Pin to a specific commit SHA for security
      [BP001] Job 'build' has no timeout defined
         → Add 'timeout-minutes' to prevent hung jobs
```

### ⚡ Pipeline Optimization

Get actionable suggestions to speed up your builds:

```bash
cicli optimize .github/workflows/ci.yml
```

```
⚡ Optimization Report
   Potential time savings: 2-5 minutes per run
──────────────────────────────────────────────────

   🔴 High Impact:
      • Add npm dependency caching [auto-fixable]
        Dependencies installed but not cached between runs
        💨 Estimated save: 30-60s

      • Enable Docker layer caching [auto-fixable]
        Docker builds can use layer caching
        💨 Estimated save: 60-300s

   🟡 Medium Impact:
      • Run lint and test in parallel
        Jobs could run concurrently after build
        💨 Estimated save: 30-120s
```

Apply auto-fixable optimizations:

```bash
cicli optimize --apply
```

### 🚀 Smart Generation

Generate optimized CI/CD based on your actual project:

```bash
# Auto-detect and generate
cicli generate

# Generate specific files
cicli generate dockerfile
cicli generate kubernetes
cicli generate pipeline --platform=github
```

### 📦 Deployment Commands

```bash
# Build & push Docker
cicli docker publish --tag=v1.0.0

# Deploy to Kubernetes
cicli deploy --env=prod --tag=v1.0.0

# Rollback
cicli rollback --env=prod

# View history
cicli history
```

## Full Command Reference

| Command | Description |
|---------|-------------|
| `cicli analyze` | Analyze project structure and technologies |
| `cicli generate` | Smart-generate CI/CD configs based on project |
| `cicli convert` | Convert between CI/CD platforms |
| `cicli lint` | Lint and validate CI/CD configurations |
| `cicli optimize` | Suggest and apply pipeline optimizations |
| `cicli docker publish` | Build and push Docker images |
| `cicli deploy` | Deploy to Kubernetes/AWS |
| `cicli rollback` | Rollback to previous version |
| `cicli history` | View deployment history |
| `cicli notify` | Send deployment notifications |

## Project Structure

```
cicli/
├── cmd/cicli/           # CLI entry point
├── internal/
│   ├── analyzer/        # Project analysis engine
│   ├── converter/       # Platform conversion
│   ├── linter/          # Pipeline linting rules
│   ├── optimizer/       # Build optimization
│   ├── generator/       # Smart config generation
│   ├── docker/          # Docker operations
│   ├── deploy/          # Deployment logic
│   ├── config/          # Configuration handling
│   ├── notify/          # Notifications
│   ├── store/           # Data persistence
│   └── validator/       # Pre-flight checks
└── pkg/                 # Shared utilities
```

## What Makes This Different

| Feature | Other Tools | CiCLI |
|---------|-------------|-------|
| Template generation | ✅ | ✅ |
| Project analysis | ❌ | ✅ |
| Platform conversion | ❌ | ✅ |
| Security linting | ❌ | ✅ |
| Performance optimization | ❌ | ✅ |
| Auto-fix suggestions | ❌ | ✅ |

## Development

```bash
cd cicli

# Run
go run ./cmd/cicli analyze

# Build
go build -o cicli ./cmd/cicli

# Test
go test ./...
```

## License

MIT License - see [LICENSE](LICENSE) for details.

## Author

**Arnab Bhowmik** — [GitHub](https://github.com/Arnab-Afk)
