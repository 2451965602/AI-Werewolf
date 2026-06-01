package main

import (
	"ai-werewolf-go/internal/application"
	"ai-werewolf-go/internal/infrastructure/ai"
	"ai-werewolf-go/internal/infrastructure/store"
	transporthttp "ai-werewolf-go/internal/transport/http"
)

func main() {
	repository := store.NewJSONStore("data/world_state.json")
	aiProvider := ai.FallbackProvider{}
	service := application.NewService(repository, aiProvider)
	router := transporthttp.NewRouter(service)
	router.Spin()
}
