package http

import (
	"context"
	"errors"

	"ai-werewolf-go/internal/application"
	"ai-werewolf-go/internal/domain"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

type GameService interface {
	StartGame(context.Context) (domain.GameState, error)
	NextPhase(context.Context) (domain.GameState, error)
	GetState(context.Context) (domain.GameState, error)
	GetMessages(context.Context) ([]domain.Message, error)
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewRouter(service GameService) *server.Hertz {
	h := server.Default()
	h.POST("/api/game/start", func(ctx context.Context, c *app.RequestContext) {
		state, err := service.StartGame(ctx)
		writeState(c, state, err)
	})
	h.POST("/api/game/next", func(ctx context.Context, c *app.RequestContext) {
		state, err := service.NextPhase(ctx)
		writeState(c, state, err)
	})
	h.GET("/api/game/state", func(ctx context.Context, c *app.RequestContext) {
		state, err := service.GetState(ctx)
		writeState(c, state, err)
	})
	h.GET("/api/game/messages", func(ctx context.Context, c *app.RequestContext) {
		messages, err := service.GetMessages(ctx)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(consts.StatusOK, messages)
	})
	h.GET("/api/game/health", func(_ context.Context, c *app.RequestContext) {
		c.JSON(consts.StatusOK, map[string]string{"status": "ok"})
	})
	return h
}

func writeState(c *app.RequestContext, state domain.GameState, err error) {
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(consts.StatusOK, state)
}

func writeError(c *app.RequestContext, err error) {
	status := consts.StatusInternalServerError
	if errors.Is(err, application.ErrNoGameState) {
		status = consts.StatusNotFound
	}
	c.JSON(status, ErrorResponse{Error: err.Error()})
}
