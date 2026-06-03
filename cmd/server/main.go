package main

import (
	"log"

	"ai-werewolf-go/internal/application"
	"ai-werewolf-go/internal/config"
	"ai-werewolf-go/internal/infrastructure/ai"
	"ai-werewolf-go/internal/infrastructure/store"
	transporthttp "ai-werewolf-go/internal/transport/http"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	repository := store.NewJSONStore(cfg.Storage.StatePath)
	aiProvider, err := ai.BuildProvider(cfg.AI)
	if err != nil {
		log.Fatalf("build ai provider: %v", err)
	}
	service := application.NewService(repository, aiProvider)
	router := transporthttp.NewRouter(service, cfg.Server.Addr)
	router.Spin()
}
