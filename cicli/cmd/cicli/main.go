package main

import (
	"fmt"
	"os"

	"cicli/internal/config"
)

func main() {
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
		fmt.Println("Generating pipeline...")
		// TODO: Call generator.Generate()
	case "docker":
		fmt.Println("Docker operations...")
		// TODO: Handle docker publish
	case "deploy":
		fmt.Println("Deploying...")
		// TODO: Handle deployment
	case "notify":
		fmt.Println("Sending notification...")
		// TODO: Handle notification
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
