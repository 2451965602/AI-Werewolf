package domain

import "fmt"

type DecisionProvider interface {
	Speak(player Player, context DecisionContext) (string, error)
	VoteTarget(player Player, context DecisionContext) (int, error)
	WerewolfTarget(player Player, context DecisionContext) (int, error)
}

type DecisionContext struct {
	Round           int
	Phase           Phase
	Players         []Player
	Messages        []Message
	LastNightKilled *int
}

func NewGame() GameState {
	roles := []Role{
		RoleWerewolf, RoleWerewolf, RoleWerewolf,
		RoleVillager, RoleVillager, RoleVillager, RoleVillager,
		RoleSeer, RoleWitch, RoleHunter,
	}
	names := []string{"李明", "王芳", "张伟", "刘洋", "陈静", "赵强", "孙悦", "周涛", "吴磊", "郑洁"}

	players := make([]Player, 0, len(roles))
	for i, role := range roles {
		team := TeamVillage
		if role == RoleWerewolf {
			team = TeamWolf
		}
		players = append(players, Player{
			ID:    i + 1,
			Name:  names[i],
			Role:  role,
			Alive: true,
			Team:  team,
		})
	}

	return GameState{
		Round:   1,
		Phase:   PhaseDay,
		Players: players,
		Messages: []Message{{
			SpeakerID: 0,
			Speaker:   "系统",
			Content:   "游戏开始，进入第1天白天",
			Phase:     PhaseDay,
			Round:     1,
			Type:      MessageTypeSystem,
		}},
	}
}

func AdvancePhase(state GameState, provider DecisionProvider) (GameState, error) {
	if ApplyGameEndIfNeeded(&state) {
		return state, nil
	}

	switch state.Phase {
	case PhaseDay:
		return advanceDay(state, provider)
	case PhaseNight:
		return advanceNight(state, provider)
	case PhaseEnded:
		state.Ended = true
		return state, nil
	default:
		return state, fmt.Errorf("unknown phase %q", state.Phase)
	}
}

func ApplyGameEndIfNeeded(state *GameState) bool {
	aliveWolves := 0
	aliveVillage := 0
	for _, player := range state.Players {
		if !player.Alive {
			continue
		}
		if player.Team == TeamWolf {
			aliveWolves++
		} else {
			aliveVillage++
		}
	}

	if aliveWolves == 0 {
		state.Ended = true
		state.Phase = PhaseEnded
		state.Winner = WinnerVillage
		return true
	}
	if aliveWolves >= aliveVillage {
		state.Ended = true
		state.Phase = PhaseEnded
		state.Winner = WinnerWolf
		return true
	}
	return false
}

func CheckGameEnd(state *GameState) bool {
	return ApplyGameEndIfNeeded(state)
}

func advanceDay(state GameState, provider DecisionProvider) (GameState, error) {
	if state.Round == 1 {
		for _, player := range state.Players {
			if !player.Alive {
				continue
			}
			content := fmt.Sprintf("大家好，我是%d号%s。", player.ID, player.Name)
			if provider != nil {
				speech, err := provider.Speak(player, NewDecisionContext(state))
				if err != nil {
					return state, fmt.Errorf("generate speech for player %d: %w", player.ID, err)
				}
				if speech != "" {
					content = speech
				}
			}
			state.Messages = append(state.Messages, Message{
				SpeakerID: player.ID,
				Speaker:   player.Name,
				Content:   content,
				Phase:     PhaseDay,
				Round:     state.Round,
				Type:      MessageTypePlayer,
			})
		}
		state.Phase = PhaseNight
		return state, nil
	}

	targetID := firstAlivePlayerID(state)
	if provider != nil {
		voters := alivePlayers(state)
		if len(voters) > 0 {
			candidate, err := provider.VoteTarget(voters[0], NewDecisionContext(state))
			if err != nil {
				return state, fmt.Errorf("vote decision for player %d: %w", voters[0].ID, err)
			}
			if isAlivePlayer(state, candidate) {
				targetID = candidate
			}
		}
	}
	if targetID != 0 {
		for i := range state.Players {
			if state.Players[i].ID == targetID {
				state.Players[i].Alive = false
				state.Messages = append(state.Messages, Message{
					SpeakerID: 0,
					Speaker:   "系统",
					Content:   fmt.Sprintf("%d号%s被放逐。", state.Players[i].ID, state.Players[i].Name),
					Phase:     PhaseDay,
					Round:     state.Round,
					Type:      MessageTypeVote,
				})
				break
			}
		}
	}
	ApplyGameEndIfNeeded(&state)
	if state.Ended {
		return state, nil
	}
	state.Phase = PhaseNight
	return state, nil
}

func advanceNight(state GameState, provider DecisionProvider) (GameState, error) {
	wolves := alivePlayersByTeam(state, TeamWolf)
	if len(wolves) == 0 || ApplyGameEndIfNeeded(&state) {
		return state, nil
	}

	targetID := 0
	if provider != nil {
		candidate, err := provider.WerewolfTarget(wolves[0], NewDecisionContext(state))
		if err != nil {
			return state, fmt.Errorf("werewolf target decision for player %d: %w", wolves[0].ID, err)
		}
		targetID = candidate
	}
	if !isLegalWerewolfTarget(state, targetID) {
		targetID = firstLegalWerewolfTarget(state)
	}
	if targetID != 0 {
		for i := range state.Players {
			if state.Players[i].ID == targetID {
				state.Players[i].Alive = false
				state.LastNightKilled = &targetID
				break
			}
		}
	}

	ApplyGameEndIfNeeded(&state)
	if !state.Ended {
		state.Round++
		state.Phase = PhaseDay
	}
	return state, nil
}

func NewDecisionContext(state GameState) DecisionContext {
	players := append([]Player(nil), state.Players...)
	messages := append([]Message(nil), state.Messages...)
	var killed *int
	if state.LastNightKilled != nil {
		value := *state.LastNightKilled
		killed = &value
	}
	return DecisionContext{
		Round:           state.Round,
		Phase:           state.Phase,
		Players:         players,
		Messages:        messages,
		LastNightKilled: killed,
	}
}

func alivePlayers(state GameState) []Player {
	players := make([]Player, 0)
	for _, player := range state.Players {
		if player.Alive {
			players = append(players, player)
		}
	}
	return players
}

func firstAlivePlayerID(state GameState) int {
	for _, player := range state.Players {
		if player.Alive {
			return player.ID
		}
	}
	return 0
}

func isAlivePlayer(state GameState, id int) bool {
	for _, player := range state.Players {
		if player.ID == id {
			return player.Alive
		}
	}
	return false
}

func alivePlayersByTeam(state GameState, team Team) []Player {
	players := make([]Player, 0)
	for _, player := range state.Players {
		if player.Alive && player.Team == team {
			players = append(players, player)
		}
	}
	return players
}

func isLegalWerewolfTarget(state GameState, targetID int) bool {
	for _, player := range state.Players {
		if player.ID == targetID {
			return player.Alive && player.Team != TeamWolf
		}
	}
	return false
}

func firstLegalWerewolfTarget(state GameState) int {
	for _, player := range state.Players {
		if player.Alive && player.Team != TeamWolf {
			return player.ID
		}
	}
	return 0
}
