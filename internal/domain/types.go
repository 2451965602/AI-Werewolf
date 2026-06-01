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

type Team string

const (
	TeamWolf    Team = "wolf"
	TeamVillage Team = "village"
)

type MessageType string

const (
	MessageTypeSystem   MessageType = "system"
	MessageTypeNarrator MessageType = "narrator"
	MessageTypePlayer   MessageType = "player"
	MessageTypeVote     MessageType = "vote"
)

type Winner string

const (
	WinnerVillage Winner = "village"
	WinnerWolf    Winner = "wolf"
)

type Player struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Role  Role   `json:"role"`
	Alive bool   `json:"alive"`
	Team  Team   `json:"team"`
}

type Message struct {
	SpeakerID int         `json:"speakerId"`
	Speaker   string      `json:"speaker"`
	Content   string      `json:"content"`
	Phase     Phase       `json:"phase"`
	Round     int         `json:"round"`
	Type      MessageType `json:"type"`
}

type Vote struct {
	VoterID  int `json:"voterId"`
	TargetID int `json:"targetId"`
	Round    int `json:"round"`
}

type GameState struct {
	Round           int       `json:"round"`
	Phase           Phase     `json:"phase"`
	Ended           bool      `json:"ended"`
	Winner          Winner    `json:"winner,omitempty"`
	Players         []Player  `json:"players"`
	Messages        []Message `json:"messages"`
	LastNightKilled *int      `json:"lastNightKilled,omitempty"`
}
