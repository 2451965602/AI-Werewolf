package domain

import (
	"errors"
	"testing"
)

type staticDecisionProvider struct {
	speech         string
	voteTarget     int
	werewolfTarget int
	seerTarget     int
	witchAction    WitchAction
	err            error
}

func (p staticDecisionProvider) Speak(Player, DecisionContext) (string, error) {
	return p.speech, p.err
}

func (p staticDecisionProvider) VoteTarget(Player, DecisionContext) (int, error) {
	return p.voteTarget, p.err
}

func (p staticDecisionProvider) WerewolfTarget(Player, DecisionContext) (int, error) {
	return p.werewolfTarget, p.err
}

func (p staticDecisionProvider) SeerTarget(Player, DecisionContext) (int, error) {
	return p.seerTarget, p.err
}

func (p staticDecisionProvider) WitchAction(Player, DecisionContext) (WitchAction, error) {
	return p.witchAction, p.err
}

func TestStartGameEntersDayOne(t *testing.T) {
	state := NewGame()
	if state.Round != 1 || state.Phase != PhaseDay {
		t.Fatalf("expected day 1, got round=%d phase=%s", state.Round, state.Phase)
	}
	if len(state.Players) != 10 {
		t.Fatalf("expected 10 players, got %d", len(state.Players))
	}
}

func TestProviderErrorIsReturned(t *testing.T) {
	state := NewGame()
	_, err := AdvancePhase(state, staticDecisionProvider{err: errors.New("ai unavailable")})
	if err == nil {
		t.Fatal("expected provider error")
	}
}

func TestDecisionContextDoesNotMutateState(t *testing.T) {
	state := NewGame()
	context := NewDecisionContext(state)
	context.Players[0].Alive = false
	if !state.Players[0].Alive {
		t.Fatal("decision context must not mutate source state")
	}
}

func TestDecisionContextScopesPrivateRoles(t *testing.T) {
	state := NewGame()
	villager := state.Players[3]
	context := NewDecisionContextForPlayer(state, villager)
	for _, player := range context.Players {
		if player.ID == villager.ID {
			if player.Role == "" || player.Team == "" {
				t.Fatal("actor should keep own private role and team")
			}
			continue
		}
		if player.Role != "" || player.Team != "" {
			t.Fatalf("non-wolf actor should not see player %d private role/team", player.ID)
		}
	}
}

func TestWerewolfContextIncludesWolfTeammates(t *testing.T) {
	state := NewGame()
	wolf := state.Players[0]
	context := NewDecisionContextForPlayer(state, wolf)
	visibleWolves := 0
	for _, player := range context.Players {
		if player.Team == TeamWolf {
			visibleWolves++
		}
	}
	if visibleWolves != 3 {
		t.Fatalf("expected wolf actor to see 3 wolves, got %d", visibleWolves)
	}
}

func TestDecisionContextScopesInspectionsAndNightDeath(t *testing.T) {
	state := NewGame()
	killed := 4
	state.LastNightKilled = &killed
	state.WitchHealUsed = true
	state.Inspections = []InspectionResult{{SeerID: 8, TargetID: 1, Role: RoleWerewolf, Round: 1}}

	villagerContext := NewDecisionContextForPlayer(state, state.Players[3])
	if villagerContext.LastNightKilled != nil || len(villagerContext.Inspections) != 0 {
		t.Fatal("villager must not see witch night death or seer inspections")
	}

	seerContext := NewDecisionContextForPlayer(state, state.Players[7])
	if len(seerContext.Inspections) != 1 || seerContext.LastNightKilled != nil {
		t.Fatal("seer should see own inspections but not witch night death")
	}

	witchContext := NewDecisionContextForPlayer(state, state.Players[8])
	if witchContext.LastNightKilled == nil || !witchContext.WitchHealUsed {
		t.Fatal("witch should see night death and potion state")
	}
	if len(witchContext.Inspections) != 0 {
		t.Fatal("witch must not see seer inspections")
	}
}

func TestSecondDayVoteExilesTarget(t *testing.T) {
	state := NewGame()
	state.Round = 2
	state.Phase = PhaseDay
	next, err := AdvancePhase(state, staticDecisionProvider{voteTarget: 4})
	if err != nil {
		t.Fatal(err)
	}
	if next.Players[3].Alive {
		t.Fatal("expected voted target to be exiled")
	}
	foundVote := false
	for _, msg := range next.Messages {
		if msg.Type == MessageTypeVote {
			foundVote = true
		}
	}
	if !foundVote {
		t.Fatal("expected vote message")
	}
}

func TestInvalidVoteTargetDoesNotMutateState(t *testing.T) {
	state := NewGame()
	state.Round = 2
	state.Phase = PhaseDay
	next, err := AdvancePhase(state, staticDecisionProvider{voteTarget: 999})
	if err != nil {
		t.Fatal(err)
	}
	for _, player := range next.Players {
		if !player.Alive {
			t.Fatalf("invalid vote target must not exile player %d", player.ID)
		}
	}
}

func TestDayOneDoesNotVote(t *testing.T) {
	state := NewGame()
	next, err := AdvancePhase(state, staticDecisionProvider{speech: "自我介绍"})
	if err != nil {
		t.Fatal(err)
	}
	if next.Ended {
		t.Fatal("game should not end on day one")
	}
	if next.Phase != PhaseNight {
		t.Fatalf("expected next phase night, got %s", next.Phase)
	}
	for _, msg := range next.Messages {
		if msg.Type == MessageTypeVote {
			t.Fatal("day one must not vote")
		}
	}
}

func TestWolvesEliminatedEndsImmediately(t *testing.T) {
	state := NewGame()
	for i := range state.Players {
		if state.Players[i].Role == RoleWerewolf {
			state.Players[i].Alive = false
		}
	}
	ended := CheckGameEnd(&state)
	if !ended || !state.Ended || state.Winner != WinnerVillage {
		t.Fatalf("expected village win, got ended=%v winner=%q", state.Ended, state.Winner)
	}
}

func TestInvalidWerewolfTargetDoesNotMutateState(t *testing.T) {
	state := NewGame()
	state.Phase = PhaseNight
	next, err := AdvancePhase(state, staticDecisionProvider{werewolfTarget: 1})
	if err != nil {
		t.Fatal(err)
	}
	if next.LastNightKilled != nil {
		t.Fatal("invalid target must not kill fallback target")
	}
	for _, player := range next.Players {
		if !player.Alive {
			t.Fatalf("invalid AI target must not mutate player %d", player.ID)
		}
	}
}

func TestSeerInspectionRecordsRole(t *testing.T) {
	state := NewGame()
	state.Phase = PhaseNight
	next, err := AdvancePhase(state, staticDecisionProvider{werewolfTarget: 4, seerTarget: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(next.Inspections) != 1 {
		t.Fatalf("expected one inspection, got %d", len(next.Inspections))
	}
	if next.Inspections[0].Role != RoleWerewolf {
		t.Fatalf("expected inspected role werewolf, got %s", next.Inspections[0].Role)
	}
}

func TestWitchCanOnlyHealLastNightKilled(t *testing.T) {
	state := NewGame()
	state.Phase = PhaseNight
	next, err := AdvancePhase(state, staticDecisionProvider{
		werewolfTarget: 4,
		witchAction:    WitchAction{Type: WitchActionHeal, TargetID: 5},
	})
	if err != nil {
		t.Fatal(err)
	}
	if next.Players[3].Alive {
		t.Fatal("actual last-night victim should remain dead when witch heals wrong target")
	}
	if next.WitchHealUsed {
		t.Fatal("invalid heal target must not consume antidote")
	}
}

func TestWitchHealRestoresLastNightKilled(t *testing.T) {
	state := NewGame()
	state.Phase = PhaseNight
	next, err := AdvancePhase(state, staticDecisionProvider{
		werewolfTarget: 4,
		witchAction:    WitchAction{Type: WitchActionHeal, TargetID: 4},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !next.Players[3].Alive {
		t.Fatal("witch should restore last-night killed target")
	}
	if !next.WitchHealUsed {
		t.Fatal("valid heal should consume antidote")
	}
}
