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

func (d *Deployer) Rollback(appName, env string) error {
	fmt.Printf("Initiating rollback for %s (Env: %s)...\n", appName, env)

	s, err := store.NewStore()
	if err != nil {
		return fmt.Errorf("failed to open store: %w", err)
	}

	deployments, err := s.Load()
	if err != nil {
		return fmt.Errorf("failed to load history: %w", err)
	}

	// Find the last successful deployment for this app and env
	// We need to skip the *current* deployment if it was successful (which is unlikely if we are rolling back,
	// but if we are rolling back a bad deployment, the bad one might be 'failed' or 'success' but buggy).
	// Strategy: Find the most recent 'success' that is NOT the latest deployment if the latest is also 'success'.
	// Actually, simpler strategy for v2: Find the most recent 'success' that isn't the one we just did.
	// But to be safe, let's just look for the last successful deployment that has a different image than the current state?
	// For simplicity in this CLI: Just find the last successful deployment. If the user just deployed a bad version,
	// it likely failed or they are manually rolling back.

	var targetDeployment *store.Deployment

	// Iterate backwards
	for i := len(deployments) - 1; i >= 0; i-- {
		dep := deployments[i]
		if dep.Project == appName && dep.Env == env && dep.Status == "success" {
			// Found a candidate.
			// In a real system we might check if this is the *current* running version.
			// Here we assume the user wants the *previous* one.
			// So if the last entry is success, we might want the one before it?
			// Let's assume the user runs rollback because the *current* state is bad.
			// So we want the *latest successful* deployment.
			// Wait, if the bad deployment failed, then the latest success IS the stable one.
			// If the bad deployment succeeded (but is buggy), then the latest success IS the bad one.
			// We need to find the success *before* the current head if the head is success.

			// Let's keep it simple: List the last 5 successes and ask user?
			// Or just take the last success.
			// If the latest deployment in history is 'failed', take the last 'success'.
			// If the latest deployment in history is 'success', take the *previous* 'success'.

			isLatest := (i == len(deployments)-1)
			if isLatest {
				continue // Skip the very latest if it's success (assume it's the buggy one we want to revert)
			}

			targetDeployment = &dep
			break
		}
	}

	if targetDeployment == nil {
		return fmt.Errorf("no stable deployment found to rollback to")
	}

	fmt.Printf("Rolling back to version: %s (Image: %s)\n", targetDeployment.Timestamp.Format(time.RFC3339), targetDeployment.Image)

	// Perform deployment
	return d.DeployToK8s("k8s/deployment.yaml", targetDeployment.Image, appName, env)
}
