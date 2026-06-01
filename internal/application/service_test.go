package application

import (
	"context"
	"errors"
	"testing"

	"ai-werewolf-go/internal/domain"
)

type fakeRepository struct {
	state     domain.GameState
	saveErr   error
	loadErr   error
	saveCount int
}

func (r *fakeRepository) Save(_ context.Context, state domain.GameState) error {
	if r.saveErr != nil {
		return r.saveErr
	}
	r.state = state
	r.saveCount++
	return nil
}

func (r *fakeRepository) Load(context.Context) (domain.GameState, error) {
	if r.loadErr != nil {
		return domain.GameState{}, r.loadErr
	}
	return r.state, nil
}

type fakeAI struct {
	err error
}

func (a fakeAI) Speak(domain.Player, domain.DecisionContext) (string, error) {
	return "测试发言", a.err
}

func (a fakeAI) VoteTarget(domain.Player, domain.DecisionContext) (int, error) {
	return 4, a.err
}

func (a fakeAI) WerewolfTarget(domain.Player, domain.DecisionContext) (int, error) {
	return 4, a.err
}

func TestStartGameSavesInitializedState(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository, fakeAI{})

	state, err := service.StartGame(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if repository.saveCount != 1 {
		t.Fatalf("expected one save, got %d", repository.saveCount)
	}
	if state.Round != 1 || state.Phase != domain.PhaseDay {
		t.Fatalf("expected day one state, got round=%d phase=%s", state.Round, state.Phase)
	}
}

func TestNextPhaseSavesUpdatedState(t *testing.T) {
	repository := &fakeRepository{state: domain.NewGame()}
	service := NewService(repository, fakeAI{})

	state, err := service.NextPhase(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if repository.saveCount != 1 {
		t.Fatalf("expected one save, got %d", repository.saveCount)
	}
	if state.Phase != domain.PhaseNight {
		t.Fatalf("expected night phase, got %s", state.Phase)
	}
}

func TestAIFailureFallsBackToLegalAction(t *testing.T) {
	repository := &fakeRepository{state: domain.NewGame()}
	service := NewService(repository, fakeAI{err: errors.New("model unavailable")})

	state, err := service.NextPhase(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if state.Phase != domain.PhaseNight {
		t.Fatalf("expected fallback to advance phase, got %s", state.Phase)
	}
}

func TestRepositoryFailurePropagates(t *testing.T) {
	repository := &fakeRepository{saveErr: errors.New("disk full")}
	service := NewService(repository, fakeAI{})

	_, err := service.StartGame(context.Background())
	if err == nil {
		t.Fatal("expected repository error")
	}
}
