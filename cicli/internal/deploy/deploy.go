package deploy

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"cicli/internal/store"
)

type Deployer struct{}

func NewDeployer() *Deployer {
	return &Deployer{}
}

func (d *Deployer) DeployToK8s(manifestPath, imageName, appName, env string) error {
	fmt.Printf("Deploying to Kubernetes (Env: %s)...\n", env)

	status := "success"
	var deployErr error

	defer func() {
		// Record history
		s, err := store.NewStore()
		if err == nil {
			if deployErr != nil {
				status = "failed"
			}
			_ = s.Add(store.Deployment{
				ID:        fmt.Sprintf("%d", time.Now().Unix()),
				Timestamp: time.Now(),
				Project:   appName,
				Env:       env,
				Image:     imageName,
				Status:    status,
			})
			fmt.Println("Deployment recorded in history.")
		}
	}()

	// 1. Apply manifest
	fmt.Printf("Applying manifest: %s\n", manifestPath)
	applyCmd := exec.Command("kubectl", "apply", "-f", manifestPath)
	applyCmd.Stdout = os.Stdout
	applyCmd.Stderr = os.Stderr
	if err := applyCmd.Run(); err != nil {
		deployErr = fmt.Errorf("failed to apply manifest: %w", err)
		return deployErr
	}

	// 2. Set image
	fmt.Printf("Updating image for deployment/%s to %s\n", appName, imageName)
	setImageCmd := exec.Command("kubectl", "set", "image", fmt.Sprintf("deployment/%s", appName), fmt.Sprintf("%s=%s", appName, imageName))
	setImageCmd.Stdout = os.Stdout
	setImageCmd.Stderr = os.Stderr
	if err := setImageCmd.Run(); err != nil {
		deployErr = fmt.Errorf("failed to set image: %w", err)
		return deployErr
	}

	// 3. Rollout status
	fmt.Printf("Waiting for rollout status...\n")
	rolloutCmd := exec.Command("kubectl", "rollout", "status", fmt.Sprintf("deployment/%s", appName))
	rolloutCmd.Stdout = os.Stdout
	rolloutCmd.Stderr = os.Stderr
	if err := rolloutCmd.Run(); err != nil {
		deployErr = err
		return deployErr
	}

	return nil
}
