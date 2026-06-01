package application

import (
	"context"
	"errors"

	"ai-werewolf-go/internal/domain"
)

var ErrNoGameState = errors.New("game state is not initialized")

type AIDecisionProvider interface {
	Speak(player domain.Player, context domain.DecisionContext) (string, error)
	VoteTarget(player domain.Player, context domain.DecisionContext) (int, error)
	WerewolfTarget(player domain.Player, context domain.DecisionContext) (int, error)
}

type StateRepository interface {
	Save(context.Context, domain.GameState) error
	Load(context.Context) (domain.GameState, error)
}

type Service struct {
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
	state := domain.NewGame()
	if err := s.repository.Save(ctx, state); err != nil {
		return domain.GameState{}, err
	}
	s.state = state
	s.hasState = true
	return state, nil
}

func (s *Service) NextPhase(ctx context.Context) (domain.GameState, error) {
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
	return next, nil
}

func (s *Service) GetState(ctx context.Context) (domain.GameState, error) {
	return s.currentState(ctx)
}

func (s *Service) GetMessages(ctx context.Context) ([]domain.Message, error) {
	state, err := s.currentState(ctx)
	if err != nil {
		return nil, err
	}
	return append([]domain.Message(nil), state.Messages...), nil
}

func (s *Service) currentState(ctx context.Context) (domain.GameState, error) {
	if s.hasState {
		return s.state, nil
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
	return state, nil
}
