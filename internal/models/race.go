package models

import (
	"math/rand"
	"time"
)

type Race struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Distance    int       `json:"distance"` // in meters
	Grade       RaceGrade `json:"grade"`
	Prize       int       `json:"prize"`
	MinRating   int       `json:"min_rating"`
	MaxEntrants int       `json:"max_entrants"`
	Date        time.Time `json:"date"`
	Entrants    []string  `json:"entrants"` // Horse IDs
}

type RaceGrade int

const (
	MaidenRace RaceGrade = iota
	Grade3
	Grade2
	Grade1
	GradeG1 // Top tier
)

func (g RaceGrade) String() string {
	switch g {
	case MaidenRace:
		return "Maiden"
	case Grade3:
		return "G3"
	case Grade2:
		return "G2"
	case Grade1:
		return "G1"
	case GradeG1:
		return "GI"
	default:
		return "?"
	}
}

type RaceStrategy struct {
	Formation Formation `json:"formation"`
	Pace      Pace      `json:"pace"`
}

type Formation int

const (
	Lead Formation = iota
	Draft
	Mount
)

func (f Formation) String() string {
	switch f {
	case Lead:
		return "Lead"
	case Draft:
		return "Draft"
	case Mount:
		return "Mount"
	default:
		return "Unknown"
	}
}

type Pace int

const (
	Fast Pace = iota
	Even
	Conserve
)

func (p Pace) String() string {
	switch p {
	case Fast:
		return "Fast"
	case Even:
		return "Even"
	case Conserve:
		return "Conserve"
	default:
		return "Unknown"
	}
}

type RaceResult struct {
	RaceID       string               `json:"race_id"`
	Results      []RaceEntrant        `json:"results"`
	PlayerHorse  string               `json:"player_horse"`
	PlayerRank   int                  `json:"player_rank"`
	PrizeMoney   int                  `json:"prize_money"`
	FansGained   int                  `json:"fans_gained"`
	Commentary   []string             `json:"commentary"`
	LiveProgress []RaceProgressUpdate `json:"live_progress"`
}

type RaceEntrant struct {
	HorseID   string `json:"horse_id"`
	HorseName string `json:"horse_name"`
	Position  int    `json:"position"`
	Time      string `json:"time"`
	Distance  int    `json:"distance"`
}

type RaceProgressUpdate struct {
	Turn       int            `json:"turn"`
	Positions  map[string]int `json:"positions"` // HorseID -> position
	Distances  map[string]int `json:"distances"` // HorseID -> distance covered
	Commentary string         `json:"commentary"`
	Events     []string       `json:"events"`
}

// CompletedRaceResult represents a historical race result for season tracking
type CompletedRaceResult struct {
	RaceID        string    `json:"race_id"`
	RaceName      string    `json:"race_name"`
	Grade         RaceGrade `json:"grade"`
	Distance      int       `json:"distance"`
	Date          time.Time `json:"date"`
	Position      int       `json:"position"`
	TotalEntrants int       `json:"total_entrants"`
	PrizeMoney    int       `json:"prize_money"`
	FansGained    int       `json:"fans_gained"`
}

func NewRace(name string, distance int, grade RaceGrade, prize int, minRating int) *Race {
	return &Race{
		ID:          generateID(),
		Name:        name,
		Distance:    distance,
		Grade:       grade,
		Prize:       prize,
		MinRating:   minRating,
		MaxEntrants: 16,
		Date:        time.Now().AddDate(0, 0, rand.Intn(30)+1),
		Entrants:    make([]string, 0),
	}
}

// GetEntryFee returns the cost to enter this race
func (r *Race) GetEntryFee() int {
	switch r.Grade {
	case MaidenRace:
		return 100
	case Grade3:
		return 300
	case Grade2:
		return 500
	case Grade1:
		return 1000
	case GradeG1:
		return 2000
	default:
		return 100
	}
}

func (r *Race) CanEnter(horse *Horse) bool {
	if horse.IsRetired {
		return false
	}
	if horse.GetOverallRating() < r.MinRating {
		return false
	}
	if len(r.Entrants) >= r.MaxEntrants {
		return false
	}
	return true
}

// CanEnterWithGameState checks if the horse can enter the race considering game progression
func (r *Race) CanEnterWithGameState(horse *Horse, gameState *GameState) bool {
	// Basic eligibility check
	if !r.CanEnter(horse) {
		return false
	}

	// Additional progression checks
	return r.MeetsProgressionRequirements(gameState)
}

// MeetsProgressionRequirements checks if player has met requirements to access this race
func (r *Race) MeetsProgressionRequirements(gameState *GameState) bool {
	// For higher tier races, require progression
	switch r.Grade {
	case Grade2:
		// Need to have completed at least one Grade3 race
		return r.hasCompletedRaceOfGrade(gameState, Grade3)
	case Grade1:
		// Need to have completed at least one Grade2 race and won a Grade3
		return r.hasCompletedRaceOfGrade(gameState, Grade2) && r.hasWonRaceOfGrade(gameState, Grade3)
	case GradeG1:
		// Need to have won at least one Grade1 race
		return r.hasWonRaceOfGrade(gameState, Grade1)
	default:
		// Maiden and Grade3 are always accessible if rating requirements are met
		return true
	}
}

func (r *Race) hasCompletedRaceOfGrade(gameState *GameState, targetGrade RaceGrade) bool {
	// Check both current season and all-time completed races
	allCompletedRaces := make([]string, 0)

	// Add current season races
	if gameState.Season.CompletedRaces != nil {
		allCompletedRaces = append(allCompletedRaces, gameState.Season.CompletedRaces...)
	}

	// Add all-time completed races
	if gameState.AllCompletedRaces != nil {
		allCompletedRaces = append(allCompletedRaces, gameState.AllCompletedRaces...)
	}

	// Check against available races in game state
	for _, completedRaceID := range allCompletedRaces {
		for _, race := range gameState.AvailableRaces {
			if race.ID == completedRaceID && race.Grade == targetGrade {
				return true
			}
		}
	}
	return false
}

func (r *Race) hasWonRaceOfGrade(gameState *GameState, targetGrade RaceGrade) bool {
	// For now, assume completion means winning for simplicity
	// Could be enhanced to track actual win records
	return r.hasCompletedRaceOfGrade(gameState, targetGrade)
}

func (r *Race) AddEntrant(horseID string) bool {
	if len(r.Entrants) >= r.MaxEntrants {
		return false
	}

	for _, id := range r.Entrants {
		if id == horseID {
			return false // Already entered
		}
	}

	r.Entrants = append(r.Entrants, horseID)
	return true
}

func (r *Race) GetPrizeForPosition(position int) int {
	switch position {
	case 1:
		return r.Prize
	case 2:
		return r.Prize / 2
	case 3:
		return r.Prize / 4
	case 4, 5:
		return r.Prize / 8
	default:
		return 0
	}
}

func (r *Race) GetFansForPosition(position int) int {
	switch position {
	case 1:
		return 1000 + int(r.Grade)*500
	case 2:
		return 500 + int(r.Grade)*250
	case 3:
		return 250 + int(r.Grade)*100
	case 4, 5:
		return 100 + int(r.Grade)*50
	default:
		return 25
	}
}
