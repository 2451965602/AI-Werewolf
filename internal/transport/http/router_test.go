package http

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"ai-werewolf-go/internal/application"
	"ai-werewolf-go/internal/domain"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/ut"
	"github.com/cloudwego/hertz/pkg/protocol"
)

type fakeGameService struct {
	state    domain.GameState
	messages []domain.Message
	err      error
}

func (s *fakeGameService) StartGame(context.Context) (domain.GameState, error) {
	return s.state, s.err
}

func (s *fakeGameService) NextPhase(context.Context) (domain.GameState, error) {
	state := s.state
	state.Phase = domain.PhaseNight
	return state, s.err
}

func (s *fakeGameService) GetState(context.Context) (domain.GameState, error) {
	return s.state, s.err
}

func (s *fakeGameService) GetMessages(context.Context) ([]domain.Message, error) {
	return s.messages, s.err
}

func TestStartGameEndpoint(t *testing.T) {
	router := NewRouter(&fakeGameService{state: domain.NewGame()}, ":0")
	response := perform(router, "POST", "/api/game/start")
	if response.StatusCode() != 200 {
		t.Fatalf("expected 200, got %d body=%s", response.StatusCode(), response.Body())
	}
}

func TestNextPhaseEndpoint(t *testing.T) {
	router := NewRouter(&fakeGameService{state: domain.NewGame()}, ":0")
	response := perform(router, "POST", "/api/game/next")
	if response.StatusCode() != 200 {
		t.Fatalf("expected 200, got %d body=%s", response.StatusCode(), response.Body())
	}
	if !bytes.Contains(response.Body(), []byte(`"phase":"night"`)) {
		t.Fatalf("expected night phase response, body=%s", response.Body())
	}
}

func TestGetStateEndpoint(t *testing.T) {
	router := NewRouter(&fakeGameService{state: domain.NewGame()}, ":0")
	response := perform(router, "GET", "/api/game/state")
	if response.StatusCode() != 200 {
		t.Fatalf("expected 200, got %d body=%s", response.StatusCode(), response.Body())
	}
}

func TestGetMessagesEndpoint(t *testing.T) {
	router := NewRouter(&fakeGameService{messages: []domain.Message{{Content: "hello"}}}, ":0")
	response := perform(router, "GET", "/api/game/messages")
	if response.StatusCode() != 200 {
		t.Fatalf("expected 200, got %d body=%s", response.StatusCode(), response.Body())
	}
	if !bytes.Contains(response.Body(), []byte("hello")) {
		t.Fatalf("expected messages response, body=%s", response.Body())
	}
}

func TestHealthEndpoint(t *testing.T) {
	router := NewRouter(&fakeGameService{}, ":0")
	response := perform(router, "GET", "/api/game/health")
	if response.StatusCode() != 200 {
		t.Fatalf("expected 200, got %d body=%s", response.StatusCode(), response.Body())
	}
}

func TestErrorResponse(t *testing.T) {
	router := NewRouter(&fakeGameService{err: errors.New("boom")}, ":0")
	response := perform(router, "POST", "/api/game/start")
	if response.StatusCode() != 500 {
		t.Fatalf("expected 500, got %d body=%s", response.StatusCode(), response.Body())
	}
	if !bytes.Contains(response.Body(), []byte("boom")) {
		t.Fatalf("expected error body, got %s", response.Body())
	}
}

func TestNoStateReturnsNotFound(t *testing.T) {
	router := NewRouter(&fakeGameService{err: application.ErrNoGameState}, ":0")
	response := perform(router, "GET", "/api/game/state")
	if response.StatusCode() != 404 {
		t.Fatalf("expected 404, got %d body=%s", response.StatusCode(), response.Body())
	}
}

func TestRouterUsesConfiguredAddress(t *testing.T) {
	router := NewRouter(&fakeGameService{}, "127.0.0.1:18080")
	if got := router.GetOptions().Addr; got != "127.0.0.1:18080" {
		t.Fatalf("expected configured addr, got %q", got)
	}
}

func perform(router *server.Hertz, method string, path string) *protocol.Response {
	body := &ut.Body{Body: bytes.NewBuffer(nil), Len: 0}
	return ut.PerformRequest(router.Engine, method, path, body).Result()
}
