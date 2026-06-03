package ai

import "ai-werewolf-go/internal/domain"

type FailoverProvider struct {
	primary   domain.DecisionProvider
	fallback  domain.DecisionProvider
}

func NewFailoverProvider(primary domain.DecisionProvider, fallback domain.DecisionProvider) *FailoverProvider {
	return &FailoverProvider{primary: primary, fallback: fallback}
}

func (p *FailoverProvider) Speak(player domain.Player, context domain.DecisionContext) (string, error) {
	result, err := p.primary.Speak(player, context)
	if err == nil && result != "" {
		return result, nil
	}
	return p.fallback.Speak(player, context)
}

func (p *FailoverProvider) VoteTarget(player domain.Player, context domain.DecisionContext) (int, error) {
	result, err := p.primary.VoteTarget(player, context)
	if err == nil && result > 0 {
		return result, nil
	}
	return p.fallback.VoteTarget(player, context)
}

func (p *FailoverProvider) WerewolfTarget(player domain.Player, context domain.DecisionContext) (int, error) {
	result, err := p.primary.WerewolfTarget(player, context)
	if err == nil && result > 0 {
		return result, nil
	}
	return p.fallback.WerewolfTarget(player, context)
}

func (p *FailoverProvider) SeerTarget(player domain.Player, context domain.DecisionContext) (int, error) {
	result, err := p.primary.SeerTarget(player, context)
	if err == nil && result > 0 {
		return result, nil
	}
	return p.fallback.SeerTarget(player, context)
}

func (p *FailoverProvider) WitchAction(player domain.Player, context domain.DecisionContext) (domain.WitchAction, error) {
	result, err := p.primary.WitchAction(player, context)
	if err == nil {
		return result, nil
	}
	return p.fallback.WitchAction(player, context)
}
