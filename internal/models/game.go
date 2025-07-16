package models

import "time"

type GameState struct {
	PlayerHorse       *Horse      `json:"player_horse"`
	Supporters        []Supporter `json:"supporters"`
	ActiveSupporters  []string    `json:"active_supporters"` // IDs of selected supporters (max 4)
	AvailableHorses   []Horse     `json:"available_horses"`
	AvailableRaces    []Race      `json:"available_races"`
	Season            Season      `json:"season"`
	GameStats         GameStats   `json:"game_stats"`
	AllCompletedRaces []string    `json:"all_completed_races"` // Track all races ever completed across seasons
	SavedAt           time.Time   `json:"saved_at"`
}

type Season struct {
	Number          int           `json:"number"`
	CurrentWeek     int           `json:"current_week"`
	MaxWeeks        int           `json:"max_weeks"`
	TrainingDays    []TrainingDay `json:"training_days"`
	CompletedRaces  []string      `json:"completed_races"` // Race IDs
	SeasonStartDate time.Time     `json:"season_start_date"`
}

type TrainingDay struct {
	Week         int             `json:"week"`
	Day          int             `json:"day"`
	TrainingType TrainingType    `json:"training_type"`
	IsRest       bool            `json:"is_rest"`
	IsCompleted  bool            `json:"is_completed"`
	Result       *TrainingResult `json:"result,omitempty"`
}

func NewSeason(number int) Season {
	return Season{
		Number:          number,
		CurrentWeek:     1,
		MaxWeeks:        24, // 6 months
		TrainingDays:    make([]TrainingDay, 0),
		CompletedRaces:  make([]string, 0),
		SeasonStartDate: time.Now(),
	}
}

func (s *Season) GetCurrentTrainingDays() []TrainingDay {
	var currentDays []TrainingDay
	for _, day := range s.TrainingDays {
		if day.Week == s.CurrentWeek {
			currentDays = append(currentDays, day)
		}
	}
	return currentDays
}

func (s *Season) AddTrainingDay(day TrainingDay) {
	s.TrainingDays = append(s.TrainingDays, day)
}

func (s *Season) NextWeek() {
	if s.CurrentWeek < s.MaxWeeks {
		s.CurrentWeek++
	}
}

func (s *Season) IsComplete() bool {
	return s.CurrentWeek >= s.MaxWeeks
}

type GameStats struct {
	TotalRaces       int `json:"total_races"`
	TotalWins        int `json:"total_wins"`
	TotalPrizeMoney  int `json:"total_prize_money"`
	TotalFans        int `json:"total_fans"`
	SeasonsCompleted int `json:"seasons_completed"`
	PlayTime         int `json:"play_time"` // in minutes
}

type Event struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Type        EventType      `json:"type"`
	Choices     []EventChoice  `json:"choices"`
	Effects     map[string]int `json:"effects"` // stat name -> change
	Probability float64        `json:"probability"`
}

type EventType int

const (
	TrainingEvent EventType = iota
	RaceEvent
	SeasonEvent
)

type EventChoice struct {
	Text    string         `json:"text"`
	Effects map[string]int `json:"effects"` // stat name -> change
}

func NewGameState() *GameState {
	return &GameState{
		PlayerHorse:       nil,
		Supporters:        make([]Supporter, 0),
		ActiveSupporters:  make([]string, 0),
		AvailableHorses:   make([]Horse, 0),
		AvailableRaces:    make([]Race, 0),
		Season:            NewSeason(1),
		GameStats:         GameStats{},
		AllCompletedRaces: make([]string, 0),
		SavedAt:           time.Now(),
	}
}

// GetActiveSupporters returns the supporters that are currently selected/active
func (gs *GameState) GetActiveSupporters() []Supporter {
	var activeSupporters []Supporter
	for _, supporter := range gs.Supporters {
		for _, activeID := range gs.ActiveSupporters {
			if supporter.ID == activeID {
				activeSupporters = append(activeSupporters, supporter)
				break
			}
		}
	}
	return activeSupporters
}

// GetOwnedSupporters returns all supporters that the player owns
func (gs *GameState) GetOwnedSupporters() []Supporter {
	var ownedSupporters []Supporter
	for _, supporter := range gs.Supporters {
		if supporter.IsOwned {
			ownedSupporters = append(ownedSupporters, supporter)
		}
	}
	return ownedSupporters
}

// CanSelectSupporter checks if a supporter can be selected (owned and not at 4 limit)
func (gs *GameState) CanSelectSupporter(supporterID string) bool {
	// Check if we already have 4 active supporters
	if len(gs.ActiveSupporters) >= 4 {
		return false
	}

	// Check if supporter is owned
	for _, supporter := range gs.Supporters {
		if supporter.ID == supporterID {
			return supporter.IsOwned
		}
	}

	return false
}

// SelectSupporter adds a supporter to the active list
func (gs *GameState) SelectSupporter(supporterID string) bool {
	if !gs.CanSelectSupporter(supporterID) {
		return false
	}

	// Check if already selected
	for _, activeID := range gs.ActiveSupporters {
		if activeID == supporterID {
			return false
		}
	}

	gs.ActiveSupporters = append(gs.ActiveSupporters, supporterID)
	return true
}

// DeselectSupporter removes a supporter from the active list
func (gs *GameState) DeselectSupporter(supporterID string) bool {
	for i, activeID := range gs.ActiveSupporters {
		if activeID == supporterID {
			gs.ActiveSupporters = append(gs.ActiveSupporters[:i], gs.ActiveSupporters[i+1:]...)
			return true
		}
	}
	return false
}

// IsSupporter selected checks if a supporter is currently active
func (gs *GameState) IsSupporterSelected(supporterID string) bool {
	for _, activeID := range gs.ActiveSupporters {
		if activeID == supporterID {
			return true
		}
	}
	return false
}
