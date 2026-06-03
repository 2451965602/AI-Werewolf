package ai

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"ai-werewolf-go/internal/domain"
)

type blockingProvider struct {
	mu      sync.Mutex
	active  int
	maxSeen int
}

func (p *blockingProvider) Speak(player domain.Player, _ domain.DecisionContext) (string, error) {
	p.mu.Lock()
	p.active++
	if p.active > p.maxSeen {
		p.maxSeen = p.active
	}
	p.mu.Unlock()

	time.Sleep(50 * time.Millisecond)

	p.mu.Lock()
	p.active--
	p.mu.Unlock()
	return fmt.Sprintf("%d号发言", player.ID), nil
}

func (p *blockingProvider) VoteTarget(domain.Player, domain.DecisionContext) (int, error) {
	return 0, nil
}

func (p *blockingProvider) WerewolfTarget(domain.Player, domain.DecisionContext) (int, error) {
	return 0, nil
}

func (p *blockingProvider) SeerTarget(domain.Player, domain.DecisionContext) (int, error) {
	return 0, nil
}

func (p *blockingProvider) WitchAction(domain.Player, domain.DecisionContext) (domain.WitchAction, error) {
	return domain.WitchAction{}, nil
}

func TestLimitedProviderSerializesCallsWhenConcurrencyIsOne(t *testing.T) {
	base := &blockingProvider{}
	provider := WrapWithConcurrencyLimit(base, 1)

	var wg sync.WaitGroup
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_, err := provider.Speak(domain.Player{ID: id, Name: fmt.Sprintf("玩家%d", id)}, domain.DecisionContext{})
			if err != nil {
				t.Errorf("Speak() error = %v", err)
			}
		}(i)
	}
	wg.Wait()

	if base.maxSeen != 1 {
		t.Fatalf("max concurrent calls = %d, want 1", base.maxSeen)
	}
}
