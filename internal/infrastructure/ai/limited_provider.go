package ai

import "ai-werewolf-go/internal/domain"

type LimitedProvider struct {
	base    domain.DecisionProvider
	permits chan struct{}
}

func WrapWithConcurrencyLimit(base domain.DecisionProvider, concurrency int) domain.DecisionProvider {
	if base == nil {
		return nil
	}
	if concurrency <= 0 {
		concurrency = 1
	}
	return &LimitedProvider{
		base:    base,
		permits: make(chan struct{}, concurrency),
	}
}

func withPermit[T any](permits chan struct{}, fn func() (T, error)) (T, error) {
	permits <- struct{}{}
	defer func() {
		<-permits
	}()
	return fn()
}

func (p *LimitedProvider) Speak(player domain.Player, context domain.DecisionContext) (string, error) {
	return withPermit(p.permits, func() (string, error) {
		return p.base.Speak(player, context)
	})
}

func (p *LimitedProvider) VoteTarget(player domain.Player, context domain.DecisionContext) (int, error) {
	return withPermit(p.permits, func() (int, error) {
		return p.base.VoteTarget(player, context)
	})
}

func (p *LimitedProvider) WerewolfTarget(player domain.Player, context domain.DecisionContext) (int, error) {
	return withPermit(p.permits, func() (int, error) {
		return p.base.WerewolfTarget(player, context)
	})
}

func (p *LimitedProvider) SeerTarget(player domain.Player, context domain.DecisionContext) (int, error) {
	return withPermit(p.permits, func() (int, error) {
		return p.base.SeerTarget(player, context)
	})
}

func (p *LimitedProvider) WitchAction(player domain.Player, context domain.DecisionContext) (domain.WitchAction, error) {
	return withPermit(p.permits, func() (domain.WitchAction, error) {
		return p.base.WitchAction(player, context)
	})
}
