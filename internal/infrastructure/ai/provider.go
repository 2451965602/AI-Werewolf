package ai

import "ai-werewolf-go/internal/domain"

type FallbackProvider struct{}

func (FallbackProvider) Speak(player domain.Player, _ domain.DecisionContext) (string, error) {
	return "大家好，我是" + player.Name + "。", nil
}

func (FallbackProvider) VoteTarget(_ domain.Player, context domain.DecisionContext) (int, error) {
	for _, player := range context.Players {
		if player.Alive {
			return player.ID, nil
		}
	}
	return 0, nil
}

func (FallbackProvider) WerewolfTarget(_ domain.Player, context domain.DecisionContext) (int, error) {
	for _, player := range context.Players {
		if player.Alive && player.Team != domain.TeamWolf {
			return player.ID, nil
		}
	}
	return 0, nil
}
