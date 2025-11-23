package validator

import (
	"fmt"
	"os/exec"
)

func CheckDocker() error {
	cmd := exec.Command("docker", "info")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker is not running or not installed: %w", err)
	}
	return nil
}

func CheckKubectl() error {
	cmd := exec.Command("kubectl", "cluster-info")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("kubectl is not configured or cluster is unreachable: %w", err)
	}
	return nil
}
