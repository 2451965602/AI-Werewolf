package domain

type Phase string

const (
	PhaseDay   Phase = "day"
	PhaseNight Phase = "night"
	PhaseEnded Phase = "ended"
)

type Role string

const (
	RoleWerewolf Role = "werewolf"
	RoleVillager Role = "villager"
	RoleSeer     Role = "seer"
	RoleWitch    Role = "witch"
	RoleHunter   Role = "hunter"
)

type Player struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Role  Role   `json:"role"`
	Alive bool   `json:"alive"`
	Team  string `json:"team"`
}

type Message struct {
	SpeakerID int    `json:"speakerId"`
	Speaker   string `json:"speaker"`
	Content   string `json:"content"`
	Phase     Phase  `json:"phase"`
	Round     int    `json:"round"`
	Type      string `json:"type"`
}

type GameState struct {
	Round           int       `json:"round"`
	Phase           Phase     `json:"phase"`
	Ended           bool      `json:"ended"`
	Winner          string    `json:"winner,omitempty"`
	Players         []Player  `json:"players"`
	Messages        []Message `json:"messages"`
	LastNightKilled int       `json:"lastNightKilled,omitempty"`
}
