package ai

import (
	"errors"
	"testing"

	"ai-werewolf-go/internal/domain"
)

type failoverDecisionProvider struct {
	speech         string
	voteTarget     int
	werewolfTarget int
	seerTarget     int
	witchAction    domain.WitchAction
	err            error
}

func (p failoverDecisionProvider) Speak(domain.Player, domain.DecisionContext) (string, error) {
	return p.speech, p.err
}

func (p failoverDecisionProvider) VoteTarget(domain.Player, domain.DecisionContext) (int, error) {
	return p.voteTarget, p.err
}

func (p failoverDecisionProvider) WerewolfTarget(domain.Player, domain.DecisionContext) (int, error) {
	return p.werewolfTarget, p.err
}

func (p failoverDecisionProvider) SeerTarget(domain.Player, domain.DecisionContext) (int, error) {
	return p.seerTarget, p.err
}

func (p failoverDecisionProvider) WitchAction(domain.Player, domain.DecisionContext) (domain.WitchAction, error) {
	return p.witchAction, p.err
}

func TestFailoverProviderFallsBackOnSpeakError(t *testing.T) {
	provider := NewFailoverProvider(
		failoverDecisionProvider{err: errors.New("timeout")},
		failoverDecisionProvider{speech: "后备发言"},
	)

	got, err := provider.Speak(domain.Player{ID: 1, Name: "李明"}, domain.DecisionContext{})
	if err != nil {
		t.Fatal(err)
	}
	if got != "后备发言" {
		t.Fatalf("Speak() = %q, want fallback speech", got)
	}
}

func TestFailoverProviderFallsBackOnVoteTargetError(t *testing.T) {
	provider := NewFailoverProvider(
		failoverDecisionProvider{err: errors.New("bad response")},
		failoverDecisionProvider{voteTarget: 4},
	)

	got, err := provider.VoteTarget(domain.Player{ID: 1, Name: "李明"}, domain.DecisionContext{})
	if err != nil {
		t.Fatal(err)
	}
	if got != 4 {
		t.Fatalf("VoteTarget() = %d, want 4", got)
	}
}
