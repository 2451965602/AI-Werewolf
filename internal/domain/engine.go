package domain

import "fmt"

type DecisionProvider interface {
	Speak(player Player, state GameState) (string, error)
	WerewolfTarget(player Player, state GameState) (int, error)
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
	if CheckGameEnd(&state) {
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

func CheckGameEnd(state *GameState) bool {
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

func advanceDay(state GameState, provider DecisionProvider) (GameState, error) {
	if state.Round == 1 {
		for _, player := range state.Players {
			if !player.Alive {
				continue
			}
			content := fmt.Sprintf("大家好，我是%d号%s。", player.ID, player.Name)
			if provider != nil {
				if speech, err := provider.Speak(player, state); err == nil && speech != "" {
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

	state.Messages = append(state.Messages, Message{
		SpeakerID: 0,
		Speaker:   "系统",
		Content:   "白天讨论结束，进入放逐投票。",
		Phase:     PhaseDay,
		Round:     state.Round,
		Type:      MessageTypeVote,
	})
	state.Phase = PhaseNight
	return state, nil
}

func advanceNight(state GameState, provider DecisionProvider) (GameState, error) {
	wolves := alivePlayersByTeam(state, TeamWolf)
	if len(wolves) == 0 || CheckGameEnd(&state) {
		return state, nil
	}

	targetID := 0
	if provider != nil {
		if candidate, err := provider.WerewolfTarget(wolves[0], state); err == nil {
			targetID = candidate
		}
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

	CheckGameEnd(&state)
	if !state.Ended {
		state.Round++
		state.Phase = PhaseDay
	}
	return state, nil
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
