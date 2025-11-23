package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type Deployment struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Project   string    `json:"project"`
	Env       string    `json:"env"`
	Image     string    `json:"image"`
	Status    string    `json:"status"`
}

type Store struct {
	FilePath string
}

func NewStore() (*Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	dir := filepath.Join(home, ".cicli")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	return &Store{
		FilePath: filepath.Join(dir, "history.json"),
	}, nil
}

func (s *Store) Load() ([]Deployment, error) {
	if _, err := os.Stat(s.FilePath); os.IsNotExist(err) {
		return []Deployment{}, nil
	}

	data, err := os.ReadFile(s.FilePath)
	if err != nil {
		return nil, err
	}

	var deployments []Deployment
	if err := json.Unmarshal(data, &deployments); err != nil {
		return nil, err
	}

	return deployments, nil
}

func (s *Store) Add(d Deployment) error {
	deployments, err := s.Load()
	if err != nil {
		return err
	}

	deployments = append(deployments, d)

	data, err := json.MarshalIndent(deployments, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.FilePath, data, 0644)
}
