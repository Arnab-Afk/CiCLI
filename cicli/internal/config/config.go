package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ProjectName  string `yaml:"project_name"`
	Language     string `yaml:"language"`
	BuildCommand string `yaml:"build_command"`
	TestCommand  string `yaml:"test_command"`
	Docker       struct {
		ImageName  string `yaml:"image_name"`
		Context    string `yaml:"context"`
		Dockerfile string `yaml:"dockerfile"`
	} `yaml:"docker"`
	Deploy struct {
		Provider     string `yaml:"provider"`
		ManifestPath string `yaml:"manifest_path"`
	} `yaml:"deploy"`
	Notifications struct {
		WebhookURL string `yaml:"webhook_url"`
	} `yaml:"notifications"`
}

func LoadConfig(path string) (*Config, error) {
	// TODO: Implement YAML parsing
	fmt.Printf("Loading config from %s\n", path)
	return &Config{}, nil
}

func InitConfig() error {
	filename := "cicli.yaml"
	if _, err := os.Stat(filename); err == nil {
		return fmt.Errorf("%s already exists", filename)
	}

	defaultConfig := Config{
		ProjectName:  "my-app",
		Language:     "node",
		BuildCommand: "npm run build",
		TestCommand:  "npm test",
	}
	defaultConfig.Docker.ImageName = "ghcr.io/user/my-app"
	defaultConfig.Docker.Context = "."
	defaultConfig.Docker.Dockerfile = "Dockerfile"
	defaultConfig.Deploy.Provider = "kubernetes"
	defaultConfig.Deploy.ManifestPath = "k8s/deployment.yaml"
	defaultConfig.Notifications.WebhookURL = "https://example.com/hook"

	data, err := yaml.Marshal(&defaultConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal default config: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", filename, err)
	}

	fmt.Printf("Initialized %s\n", filename)
	return nil
}
