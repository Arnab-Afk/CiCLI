package main

import (
	"fmt"
	"os"
	"strings"

	"cicli/internal/config"
	"cicli/internal/deploy"
	"cicli/internal/docker"
	"cicli/internal/generator"
	"cicli/internal/notify"
	"cicli/internal/store"
	"cicli/internal/validator"
)

func main() {
	fmt.Println(`
   ______  _   ______  __     ____
  / ____/ (_) / ____/ / /    /  _/
 / /     / / / /     / /     / /  
/ /___  / / / /___  / /___ _/ /   
\____/ /_/  \____/ /_____//___/   
                                  `)
	fmt.Println("CiCLI - CI/CD Pipeline Generator")

	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "init":
		if err := config.InitConfig(); err != nil {
			fmt.Printf("Error initializing config: %v\n", err)
			os.Exit(1)
		}
	case "generate":
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
	case "docker":
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
	case "deploy":
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
		fullImageName := fmt.Sprintf("%s:%s", cfg.Docker.ImageName, tag)

		appName := cfg.ProjectName

		if err := dep.DeployToK8s(cfg.Deploy.ManifestPath, fullImageName, appName, env); err != nil {
			fmt.Printf("Error deploying: %v\n", err)
			os.Exit(1)
		}
	case "history":
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
	case "rollback":
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
		// Assuming app name matches project name
		appName := cfg.ProjectName

		if err := dep.Rollback(appName, env); err != nil {
			fmt.Printf("Error rolling back: %v\n", err)
			os.Exit(1)
		}
	case "notify":
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
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println(`Usage:
  cicli init              Initialize project config
  cicli generate          Generate CI/CD pipeline
  cicli docker publish    Build & push Docker image
  cicli deploy            Deploy to Kubernetes/AWS
  cicli rollback          Rollback to previous stable version
  cicli history           View deployment history
  cicli notify            Send monitoring alerts`)
}
