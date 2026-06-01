package store

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"ai-werewolf-go/internal/domain"
)

func TestJSONStoreSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "world_state.json")
	store := NewJSONStore(path)
	want := domain.NewGame()

	if err := store.Save(context.Background(), want); err != nil {
		t.Fatal(err)
	}
	got, err := store.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got.Round != want.Round || got.Phase != want.Phase || len(got.Players) != len(want.Players) {
		t.Fatalf("loaded state mismatch: got round=%d phase=%s players=%d", got.Round, got.Phase, len(got.Players))
	}
}

func TestJSONStoreInvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "world_state.json")
	if err := os.WriteFile(path, []byte("not-json"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := NewJSONStore(path).Load(context.Background())
	if err == nil {
		t.Fatal("expected invalid JSON error")
	}
}

func TestJSONStoreFailedWritePreservesPreviousState(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "world_state.json")
	store := NewJSONStore(path)
	initial := domain.NewGame()
	if err := store.Save(context.Background(), initial); err != nil {
		t.Fatal(err)
	}

	blockedPath := filepath.Join(dir, "blocked", "world_state.json")
	if err := os.WriteFile(filepath.Join(dir, "blocked"), []byte("file blocks directory"), 0o644); err != nil {
		t.Fatal(err)
	}
	failedStore := NewJSONStore(blockedPath)
	if err := failedStore.Save(context.Background(), initial); err == nil {
		t.Fatal("expected write failure")
	}

	loaded, err := store.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Round != initial.Round || loaded.Phase != initial.Phase {
		t.Fatal("previous state should remain readable")
	}
}
