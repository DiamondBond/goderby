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
