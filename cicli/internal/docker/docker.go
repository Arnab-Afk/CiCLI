package docker

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Client struct{}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Build(imageName, context, dockerfile string) error {
	fmt.Printf("Building Docker image: %s\n", imageName)
	cmd := exec.Command("docker", "build", "-t", imageName, "-f", dockerfile, context)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (c *Client) Push(imageName string) error {
	fmt.Printf("Pushing Docker image: %s\n", imageName)
	cmd := exec.Command("docker", "push", imageName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (c *Client) GetGitSHA() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
