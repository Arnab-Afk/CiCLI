package config

import (
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
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
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

func InitConfig() error {
	filename := "cicli.yaml"
	if _, err := os.Stat(filename); err == nil {
		return fmt.Errorf("%s already exists", filename)
	}

	var (
		projectName string
		language    string
		imageName   string
		provider    string
	)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Project Name").
				Value(&projectName),
			huh.NewSelect[string]().
				Title("Language").
				Options(
					huh.NewOption("Node.js", "node"),
					huh.NewOption("Go", "go"),
					huh.NewOption("Python", "python"),
				).
				Value(&language),
			huh.NewInput().
				Title("Docker Image Name").
				Description("e.g. ghcr.io/user/app").
				Value(&imageName),
			huh.NewSelect[string]().
				Title("Deployment Provider").
				Options(
					huh.NewOption("Kubernetes", "kubernetes"),
					huh.NewOption("AWS EKS", "aws"),
				).
				Value(&provider),
		),
	)

	if err := form.Run(); err != nil {
		return fmt.Errorf("failed to run form: %w", err)
	}

	defaultConfig := Config{
		ProjectName:  projectName,
		Language:     language,
		BuildCommand: "npm run build", // Default, could be refined based on lang
		TestCommand:  "npm test",
	}

	if language == "go" {
		defaultConfig.BuildCommand = "go build ./..."
		defaultConfig.TestCommand = "go test ./..."
	} else if language == "python" {
		defaultConfig.BuildCommand = "pip install -r requirements.txt"
		defaultConfig.TestCommand = "pytest"
	}

	defaultConfig.Docker.ImageName = imageName
	defaultConfig.Docker.Context = "."
	defaultConfig.Docker.Dockerfile = "Dockerfile"
	defaultConfig.Deploy.Provider = provider
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
