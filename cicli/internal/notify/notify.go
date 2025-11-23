package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Notifier struct{}

func NewNotifier() *Notifier {
	return &Notifier{}
}

type Payload struct {
	Project   string `json:"project"`
	Status    string `json:"status"`
	Env       string `json:"env"`
	Version   string `json:"version"`
	Timestamp string `json:"timestamp"`
}

func (n *Notifier) Send(webhookURL, project, status, env, version string) error {
	fmt.Printf("Sending notification to %s...\n", webhookURL)

	payload := Payload{
		Project:   project,
		Status:    status,
		Env:       env,
		Version:   version,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status: %s", resp.Status)
	}

	fmt.Println("Notification sent successfully!")
	return nil
}
