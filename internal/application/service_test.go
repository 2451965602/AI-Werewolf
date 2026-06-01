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
	err       error
	failures  int
	speakCall int
}

func (a *fakeAI) Speak(domain.Player, domain.DecisionContext) (string, error) {
	a.speakCall++
	if a.speakCall <= a.failures {
		return "", a.err
	}
	return "测试发言", nil
}

func (a *fakeAI) VoteTarget(domain.Player, domain.DecisionContext) (int, error) {
	return 4, a.err
}

func (a *fakeAI) WerewolfTarget(domain.Player, domain.DecisionContext) (int, error) {
	return 4, a.err
}

func TestStartGameSavesInitializedState(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository, &fakeAI{})

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
	service := NewService(repository, &fakeAI{})

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
	service := NewService(repository, &fakeAI{err: errors.New("model unavailable"), failures: 99})

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
	service := NewService(repository, &fakeAI{})

	_, err := service.StartGame(context.Background())
	if err == nil {
		t.Fatal("expected repository error")
	}
}

func TestNextPhaseSaveFailureDoesNotPolluteCachedState(t *testing.T) {
	initial := domain.NewGame()
	repository := &fakeRepository{state: initial, saveErr: errors.New("disk full")}
	service := NewService(repository, &fakeAI{})
	service.state = initial
	service.hasState = true

	_, err := service.NextPhase(context.Background())
	if err == nil {
		t.Fatal("expected save failure")
	}
	state, err := service.GetState(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if state.Phase != domain.PhaseDay {
		t.Fatalf("expected cached phase to remain day after save failure, got %s", state.Phase)
	}
}

func TestGetStateReturnsCopy(t *testing.T) {
	repository := &fakeRepository{state: domain.NewGame()}
	service := NewService(repository, &fakeAI{})

	state, err := service.GetState(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	state.Players[0].Alive = false

	again, err := service.GetState(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !again.Players[0].Alive {
		t.Fatal("GetState must not expose mutable cached state")
	}
}

func TestAIFailureRetriesBeforeFallback(t *testing.T) {
	repository := &fakeRepository{state: domain.NewGame()}
	ai := &fakeAI{err: errors.New("temporary model error"), failures: 1}
	service := NewService(repository, ai)

	state, err := service.NextPhase(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if ai.speakCall < 2 {
		t.Fatalf("expected bounded retry, got %d calls", ai.speakCall)
	}
	if state.Phase != domain.PhaseNight {
		t.Fatalf("expected retry to advance phase, got %s", state.Phase)
	}
}
