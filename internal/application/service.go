package application

import (
	"context"
	"errors"
	"sync"

	"ai-werewolf-go/internal/domain"
)

var ErrNoGameState = errors.New("game state is not initialized")

type AIDecisionProvider = domain.DecisionProvider

type StateRepository interface {
	Save(context.Context, domain.GameState) error
	Load(context.Context) (domain.GameState, error)
}

type Service struct {
	mu         sync.Mutex
	repository StateRepository
	ai         AIDecisionProvider
	state      domain.GameState
	hasState   bool
}

const maxAIAttempts = 2

func NewService(repository StateRepository, ai AIDecisionProvider) *Service {
	return &Service{repository: repository, ai: ai}
}

func (s *Service) StartGame(ctx context.Context) (domain.GameState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	state := domain.NewGame()
	if err := s.repository.Save(ctx, state); err != nil {
		return domain.GameState{}, err
	}
	s.state = state
	s.hasState = true
	return domain.CloneGameState(state), nil
}

func (s *Service) NextPhase(ctx context.Context) (domain.GameState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, err := s.currentState(ctx)
	if err != nil {
		return domain.GameState{}, err
	}

	var next domain.GameState
	for attempt := 0; attempt < maxAIAttempts; attempt++ {
		next, err = domain.AdvancePhase(state, s.ai)
		if err == nil {
			break
		}
	}
	if err != nil {
		next, err = domain.AdvancePhase(state, nil)
		if err != nil {
			return domain.GameState{}, err
		}
	}

	if err := s.repository.Save(ctx, next); err != nil {
		return domain.GameState{}, err
	}
	s.state = next
	s.hasState = true
	return domain.CloneGameState(next), nil
}

func (s *Service) GetState(ctx context.Context) (domain.GameState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.currentState(ctx)
}

func (s *Service) GetMessages(ctx context.Context) ([]domain.Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, err := s.currentState(ctx)
	if err != nil {
		return nil, err
	}
	return append([]domain.Message(nil), state.Messages...), nil
}

func (s *Service) currentState(ctx context.Context) (domain.GameState, error) {
	if s.hasState {
		return domain.CloneGameState(s.state), nil
	}
	state, err := s.repository.Load(ctx)
	if err != nil {
		return domain.GameState{}, err
	}
	if state.Round == 0 {
		return domain.GameState{}, ErrNoGameState
	}
	s.state = state
	s.hasState = true
	return domain.CloneGameState(state), nil
}
