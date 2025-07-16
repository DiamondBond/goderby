package models

import (
	"fmt"
	"time"
)

type GameState struct {
	PlayerHorse       *Horse           `json:"player_horse"`
	Supporters        []Supporter      `json:"supporters"`
	ActiveSupporters  []string         `json:"active_supporters"` // IDs of selected supporters (max 4)
	AvailableHorses   []Horse          `json:"available_horses"`
	AvailableRaces    []Race           `json:"available_races"`
	Season            Season           `json:"season"`
	GameStats         GameStats        `json:"game_stats"`
	AllCompletedRaces []string         `json:"all_completed_races"` // Track all races ever completed across seasons
	RetiredHorses     []RetiredHorse   `json:"retired_horses"`      // Gallery of retired horses
	RetirementHomes   []RetirementHome `json:"retirement_homes"`    // Available retirement homes
	SavedAt           time.Time        `json:"saved_at"`
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
		RetiredHorses:     make([]RetiredHorse, 0),
		RetirementHomes:   initializeRetirementHomes(),
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

	// Check if supporter exists
	for _, supporter := range gs.Supporters {
		if supporter.ID == supporterID {
			// If supporter is owned, they can be selected
			if supporter.IsOwned {
				return true
			}
			// If supporter is not owned, only allow Common rarity for initial selection
			return supporter.Rarity == Common
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

// Retirement system structures and functions

type RetiredHorse struct {
	Horse              Horse              `json:"horse"`
	RetiredAt          time.Time          `json:"retired_at"`
	RetirementHome     RetirementHome     `json:"retirement_home"`
	PostRetirementRole PostRetirementRole `json:"post_retirement_role"`
	CareerHighlights   CareerHighlights   `json:"career_highlights"`
	Awards             []Award            `json:"awards"`
	PassiveIncome      int                `json:"passive_income"`    // Monthly passive income
	PassiveFame        int                `json:"passive_fame"`      // Monthly passive fame
	LastPassiveGain    time.Time          `json:"last_passive_gain"` // Track when passive gains were last calculated
}

type RetirementHome struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	Description      string  `json:"description"`
	Cost             int     `json:"cost"`
	Tier             int     `json:"tier"`              // 1=Basic, 2=Premium, 3=Luxury
	IsUnlocked       bool    `json:"is_unlocked"`       // Whether player can access this home
	IsOwned          bool    `json:"is_owned"`          // Whether player has purchased this home
	Capacity         int     `json:"capacity"`          // How many horses can retire here
	IncomeMultiplier float64 `json:"income_multiplier"` // Multiplier for passive income generation
	FameMultiplier   float64 `json:"fame_multiplier"`   // Multiplier for passive fame generation
}

type PostRetirementRole int

const (
	ShowHorse PostRetirementRole = iota
	StudHorse
	BreedingMare
	TrainingMentor
)

func (r PostRetirementRole) String() string {
	switch r {
	case ShowHorse:
		return "Show Horse"
	case StudHorse:
		return "Stud Horse"
	case BreedingMare:
		return "Breeding Mare"
	case TrainingMentor:
		return "Training Mentor"
	default:
		return "Unknown"
	}
}

type CareerHighlights struct {
	TotalWins            int     `json:"total_wins"`
	TotalRaces           int     `json:"total_races"`
	TotalPrizeMoney      int     `json:"total_prize_money"`
	TotalFanSupport      int     `json:"total_fan_support"`
	HighestRating        int     `json:"highest_rating"`
	LongestWinStreak     int     `json:"longest_win_streak"`
	MostPrestigiousRace  string  `json:"most_prestigious_race"`
	WinPercentage        float64 `json:"win_percentage"`
	CareerLength         int     `json:"career_length"` // Number of seasons
	PeakAge              int     `json:"peak_age"`
	FavoriteTrainingType string  `json:"favorite_training_type"`
}

type Award struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Icon        string      `json:"icon"`
	EarnedAt    time.Time   `json:"earned_at"`
	Rarity      AwardRarity `json:"rarity"`
}

type AwardRarity int

const (
	AwardCommon AwardRarity = iota
	AwardUncommon
	AwardRare
	AwardEpic
	AwardLegendary
)

func (r AwardRarity) String() string {
	switch r {
	case AwardCommon:
		return "Common"
	case AwardUncommon:
		return "Uncommon"
	case AwardRare:
		return "Rare"
	case AwardEpic:
		return "Epic"
	case AwardLegendary:
		return "Legendary"
	default:
		return "Common"
	}
}

// Initialize retirement homes with default options
func initializeRetirementHomes() []RetirementHome {
	return []RetirementHome{
		{
			ID:               "basic_paddock",
			Name:             "Basic Paddock",
			Description:      "A simple retirement home with basic care",
			Cost:             0,
			Tier:             1,
			IsUnlocked:       true,
			IsOwned:          true,
			Capacity:         2,
			IncomeMultiplier: 0.5,
			FameMultiplier:   0.3,
		},
		{
			ID:               "premium_ranch",
			Name:             "Premium Ranch",
			Description:      "Well-maintained facilities with professional care",
			Cost:             50000,
			Tier:             2,
			IsUnlocked:       true,
			IsOwned:          false,
			Capacity:         4,
			IncomeMultiplier: 1.0,
			FameMultiplier:   0.8,
		},
		{
			ID:               "luxury_estate",
			Name:             "Luxury Estate",
			Description:      "Top-tier facilities with world-class breeding programs",
			Cost:             150000,
			Tier:             3,
			IsUnlocked:       false,
			IsOwned:          false,
			Capacity:         6,
			IncomeMultiplier: 2.0,
			FameMultiplier:   1.5,
		},
		{
			ID:               "champions_hall",
			Name:             "Champion's Hall",
			Description:      "Elite retirement home for legendary horses",
			Cost:             500000,
			Tier:             4,
			IsUnlocked:       false,
			IsOwned:          false,
			Capacity:         8,
			IncomeMultiplier: 3.0,
			FameMultiplier:   2.5,
		},
	}
}

// CalculateCareerHighlights calculates career highlights for a horse
func (gs *GameState) CalculateCareerHighlights(horse *Horse) CareerHighlights {
	winPercentage := 0.0
	if horse.Races > 0 {
		winPercentage = float64(horse.Wins) / float64(horse.Races) * 100
	}

	return CareerHighlights{
		TotalWins:            horse.Wins,
		TotalRaces:           horse.Races,
		TotalPrizeMoney:      horse.Money,
		TotalFanSupport:      horse.FanSupport,
		HighestRating:        horse.GetOverallRating(),
		LongestWinStreak:     0,          // TODO: Track this in future
		MostPrestigiousRace:  "G1 Derby", // TODO: Track this from race history
		WinPercentage:        winPercentage,
		CareerLength:         gs.Season.Number,
		PeakAge:              5,       // TODO: Track actual peak performance age
		FavoriteTrainingType: "Speed", // TODO: Track from training history
	}
}

// CalculateAwards calculates awards earned by a horse based on career performance
func (gs *GameState) CalculateAwards(horse *Horse, highlights CareerHighlights) []Award {
	var awards []Award
	currentTime := time.Now()

	// Win-based awards
	if highlights.TotalWins >= 1 {
		awards = append(awards, Award{
			ID:          "first_win",
			Name:        "First Victory",
			Description: "Won your first race",
			Icon:        "🏆",
			EarnedAt:    currentTime,
			Rarity:      AwardCommon,
		})
	}

	if highlights.TotalWins >= 5 {
		awards = append(awards, Award{
			ID:          "winner",
			Name:        "Champion",
			Description: "Won 5 races",
			Icon:        "🏆",
			EarnedAt:    currentTime,
			Rarity:      AwardUncommon,
		})
	}

	if highlights.TotalWins >= 10 {
		awards = append(awards, Award{
			ID:          "superstar",
			Name:        "Superstar",
			Description: "Won 10 races",
			Icon:        "⭐",
			EarnedAt:    currentTime,
			Rarity:      AwardRare,
		})
	}

	// Win percentage awards
	if highlights.WinPercentage >= 70 && highlights.TotalRaces >= 5 {
		awards = append(awards, Award{
			ID:          "consistent_winner",
			Name:        "Consistent Winner",
			Description: "Maintained 70%+ win rate",
			Icon:        "💯",
			EarnedAt:    currentTime,
			Rarity:      AwardEpic,
		})
	}

	// Longevity awards
	if highlights.CareerLength >= 5 {
		awards = append(awards, Award{
			ID:          "veteran",
			Name:        "Veteran",
			Description: "Competed for 5+ seasons",
			Icon:        "🎖️",
			EarnedAt:    currentTime,
			Rarity:      AwardRare,
		})
	}

	// Prize money awards
	if highlights.TotalPrizeMoney >= 100000 {
		awards = append(awards, Award{
			ID:          "money_maker",
			Name:        "Money Maker",
			Description: "Earned $100,000+ in prize money",
			Icon:        "💰",
			EarnedAt:    currentTime,
			Rarity:      AwardUncommon,
		})
	}

	// Fan support awards
	if highlights.TotalFanSupport >= 1000 {
		awards = append(awards, Award{
			ID:          "crowd_favorite",
			Name:        "Crowd Favorite",
			Description: "Gained 1000+ fan support",
			Icon:        "❤️",
			EarnedAt:    currentTime,
			Rarity:      AwardUncommon,
		})
	}

	// Age-based awards
	if horse.Age >= 8 {
		awards = append(awards, Award{
			ID:          "iron_horse",
			Name:        "Iron Horse",
			Description: "Competed until age 8+",
			Icon:        "🐎",
			EarnedAt:    currentTime,
			Rarity:      AwardRare,
		})
	}

	// Performance awards
	if highlights.HighestRating >= 400 {
		awards = append(awards, Award{
			ID:          "elite_performer",
			Name:        "Elite Performer",
			Description: "Achieved 400+ rating",
			Icon:        "⚡",
			EarnedAt:    currentTime,
			Rarity:      AwardEpic,
		})
	}

	if highlights.HighestRating >= 500 {
		awards = append(awards, Award{
			ID:          "legend",
			Name:        "Legend",
			Description: "Achieved 500+ rating",
			Icon:        "👑",
			EarnedAt:    currentTime,
			Rarity:      AwardLegendary,
		})
	}

	return awards
}

// Retire a horse to a specific retirement home
func (gs *GameState) RetireHorse(homeID string, role PostRetirementRole) error {
	if gs.PlayerHorse == nil {
		return fmt.Errorf("no horse to retire")
	}

	// Find the retirement home
	var selectedHome *RetirementHome
	for i, home := range gs.RetirementHomes {
		if home.ID == homeID {
			selectedHome = &gs.RetirementHomes[i]
			break
		}
	}

	if selectedHome == nil {
		return fmt.Errorf("retirement home not found")
	}

	if !selectedHome.IsOwned {
		return fmt.Errorf("retirement home not owned")
	}

	// Check capacity
	currentResidents := 0
	for _, retired := range gs.RetiredHorses {
		if retired.RetirementHome.ID == homeID {
			currentResidents++
		}
	}

	if currentResidents >= selectedHome.Capacity {
		return fmt.Errorf("retirement home at capacity")
	}

	// Calculate career highlights and awards
	highlights := gs.CalculateCareerHighlights(gs.PlayerHorse)
	awards := gs.CalculateAwards(gs.PlayerHorse, highlights)

	// Calculate passive income and fame based on performance and home quality
	baseIncome := highlights.TotalPrizeMoney / 100 // 1% of career earnings per month
	baseFame := highlights.TotalFanSupport / 50    // 2% of career fan support per month

	passiveIncome := int(float64(baseIncome) * selectedHome.IncomeMultiplier)
	passiveFame := int(float64(baseFame) * selectedHome.FameMultiplier)

	// Create retired horse record
	retiredHorse := RetiredHorse{
		Horse:              *gs.PlayerHorse,
		RetiredAt:          time.Now(),
		RetirementHome:     *selectedHome,
		PostRetirementRole: role,
		CareerHighlights:   highlights,
		Awards:             awards,
		PassiveIncome:      passiveIncome,
		PassiveFame:        passiveFame,
		LastPassiveGain:    time.Now(),
	}

	// Mark horse as retired
	gs.PlayerHorse.IsRetired = true

	// Add to retired horses gallery
	gs.RetiredHorses = append(gs.RetiredHorses, retiredHorse)

	// Clear player horse
	gs.PlayerHorse = nil

	return nil
}

// Get available retirement homes that the player can afford
func (gs *GameState) GetAvailableRetirementHomes() []RetirementHome {
	var available []RetirementHome
	for _, home := range gs.RetirementHomes {
		if home.IsUnlocked && (home.IsOwned || gs.PlayerHorse.Money >= home.Cost) {
			available = append(available, home)
		}
	}
	return available
}

// Purchase a retirement home
func (gs *GameState) PurchaseRetirementHome(homeID string) error {
	if gs.PlayerHorse == nil {
		return fmt.Errorf("no horse to purchase retirement home")
	}

	for i, home := range gs.RetirementHomes {
		if home.ID == homeID {
			if !home.IsUnlocked {
				return fmt.Errorf("retirement home not unlocked")
			}
			if home.IsOwned {
				return fmt.Errorf("retirement home already owned")
			}
			if gs.PlayerHorse.Money < home.Cost {
				return fmt.Errorf("insufficient funds")
			}

			gs.PlayerHorse.Money -= home.Cost
			gs.RetirementHomes[i].IsOwned = true
			return nil
		}
	}

	return fmt.Errorf("retirement home not found")
}

// Update passive income and fame from retired horses
func (gs *GameState) UpdatePassiveGains() {
	currentTime := time.Now()

	for i, retired := range gs.RetiredHorses {
		// Calculate months since last passive gain
		monthsSince := int(currentTime.Sub(retired.LastPassiveGain).Hours() / (24 * 30))

		if monthsSince >= 1 {
			// Add passive income and fame
			gs.GameStats.TotalPrizeMoney += retired.PassiveIncome * monthsSince
			gs.GameStats.TotalFans += retired.PassiveFame * monthsSince

			// Update last passive gain time
			gs.RetiredHorses[i].LastPassiveGain = currentTime
		}
	}
}
