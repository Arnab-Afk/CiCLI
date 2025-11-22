package config

import (
	"fmt"
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
	// TODO: Create default cicli.yaml
	fmt.Println("Creating cicli.yaml...")
	return nil
}
