package store

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"ai-werewolf-go/internal/domain"
)

type JSONStore struct {
	path   string
	rename func(string, string) error
}

func NewJSONStore(path string) *JSONStore {
	return &JSONStore{path: path, rename: os.Rename}
}

func (s *JSONStore) Save(_ context.Context, state domain.GameState) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("create state directory: %w", err)
	}

	temp, err := os.CreateTemp(filepath.Dir(s.path), ".world_state_*.tmp")
	if err != nil {
		return fmt.Errorf("create temp state file: %w", err)
	}
	tempPath := temp.Name()
	removeTemp := true
	defer func() {
		if removeTemp {
			_ = os.Remove(tempPath)
		}
	}()

	encoder := json.NewEncoder(temp)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(domain.CloneGameState(state)); err != nil {
		_ = temp.Close()
		return fmt.Errorf("encode state: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close temp state file: %w", err)
	}
	if err := s.rename(tempPath, s.path); err != nil {
		return fmt.Errorf("replace state file: %w", err)
	}
	removeTemp = false
	return nil
}

func (s *JSONStore) Load(_ context.Context) (domain.GameState, error) {
	file, err := os.Open(s.path)
	if err != nil {
		return domain.GameState{}, fmt.Errorf("open state file: %w", err)
	}
	defer file.Close()

	var state domain.GameState
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&state); err != nil {
		return domain.GameState{}, fmt.Errorf("decode state file: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return domain.GameState{}, fmt.Errorf("decode state file: trailing JSON content")
	}
	return domain.CloneGameState(state), nil
}
