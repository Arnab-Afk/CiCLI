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
  cicli notify            Send monitoring alerts`)
}
