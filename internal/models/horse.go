package models

import (
	cryptorand "crypto/rand"
	"encoding/hex"
	"math"
	"math/rand"
	"time"
)

type Horse struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Breed        string    `json:"breed"`
	Age          int       `json:"age"`
	Stamina      int       `json:"stamina"`
	Speed        int       `json:"speed"`
	Technique    int       `json:"technique"`
	Mental       int       `json:"mental"`
	MaxStamina   int       `json:"max_stamina"`
	MaxSpeed     int       `json:"max_speed"`
	MaxTechnique int       `json:"max_technique"`
	MaxMental    int       `json:"max_mental"`
	Fatigue      int       `json:"fatigue"`
	Morale       int       `json:"morale"`
	FanSupport   int       `json:"fan_support"`
	Money        int       `json:"money"`
	Wins         int       `json:"wins"`
	Races        int       `json:"races"`
	IsRetired    bool      `json:"is_retired"`
	CreatedAt    time.Time `json:"created_at"`
}

func NewHorse(name, breed string, baseStats Stats) *Horse {
	return &Horse{
		ID:           generateID(),
		Name:         name,
		Breed:        breed,
		Age:          2,
		Stamina:      baseStats.Stamina,
		Speed:        baseStats.Speed,
		Technique:    baseStats.Technique,
		Mental:       baseStats.Mental,
		MaxStamina:   baseStats.Stamina + 200,
		MaxSpeed:     baseStats.Speed + 200,
		MaxTechnique: baseStats.Technique + 200,
		MaxMental:    baseStats.Mental + 200,
		Fatigue:      0,
		Morale:       100,
		FanSupport:   0,
		Money:        10000,
		Wins:         0,
		Races:        0,
		IsRetired:    false,
		CreatedAt:    time.Now(),
	}
}

func (h *Horse) Train(trainingType TrainingType, supporters []Supporter) TrainingResult {
	if h.Fatigue >= 80 {
		return TrainingResult{
			Success: false,
			Message: "Horse is too fatigued to train effectively!",
		}
	}

	bonus := calculateSupporterBonus(supporters, trainingType)
	baseGain := 10 + bonus

	// Enhanced morale multiplier system
	var moraleMultiplier float64
	if h.Morale >= 100 {
		moraleMultiplier = 1.20 // 20% bonus for excellent morale
	} else if h.Morale >= 80 {
		moraleMultiplier = 1.10 // 10% bonus for good morale
	} else if h.Morale >= 60 {
		moraleMultiplier = 1.00 // Normal training
	} else if h.Morale >= 40 {
		moraleMultiplier = 0.90 // 10% penalty for low morale
	} else {
		moraleMultiplier = 0.80 // 20% penalty for very low morale
	}

	actualGain := int(math.Round(float64(baseGain) * moraleMultiplier))

	// Track if stat was at max before training
	wasAtMax := false
	switch trainingType {
	case StaminaTraining:
		wasAtMax = h.Stamina >= h.MaxStamina
		if h.Stamina < h.MaxStamina {
			h.Stamina = min(h.Stamina+actualGain, h.MaxStamina)
		}
	case SpeedTraining:
		wasAtMax = h.Speed >= h.MaxSpeed
		if h.Speed < h.MaxSpeed {
			h.Speed = min(h.Speed+actualGain, h.MaxSpeed)
		}
	case TechniqueTraining:
		wasAtMax = h.Technique >= h.MaxTechnique
		if h.Technique < h.MaxTechnique {
			h.Technique = min(h.Technique+actualGain, h.MaxTechnique)
		}
	case MentalTraining:
		wasAtMax = h.Mental >= h.MaxMental
		if h.Mental < h.MaxMental {
			h.Mental = min(h.Mental+actualGain, h.MaxMental)
		}
	}

	// Apply fatigue
	h.Fatigue += 15
	if h.Fatigue > 100 {
		h.Fatigue = 100
	}

	// Daily morale decay (-2 per training session)
	h.Morale = max(h.Morale-2, 20)

	// Bonus morale for maxing out a stat
	if !wasAtMax && h.IsStatMaxed(trainingType) {
		h.Morale = min(h.Morale+5, 100)
	}

	// Generate random training event
	event := h.generateTrainingEvent()
	result := TrainingResult{
		Success:     true,
		Message:     "Training completed successfully!",
		StatGain:    actualGain,
		FatigueGain: 15,
		Event:       event,
	}

	// Apply event effects if any
	if event != nil {
		result.Message = event.Description
		h.applyEventEffects(event.Effects)
	}

	return result
}

func (h *Horse) Rest() {
	h.Fatigue = max(h.Fatigue-30, 0)
	h.Morale = min(h.Morale+10, 100)
}

func (h *Horse) GetOverallRating() int {
	baseRating := (h.Stamina + h.Speed + h.Technique + h.Mental) / 4
	return h.GetAgeAdjustedRating(baseRating)
}

// GetAgeAdjustedRating applies age-based performance modifiers
func (h *Horse) GetAgeAdjustedRating(baseRating int) int {
	ageFactor := h.GetAgePerformanceFactor()
	adjustedRating := int(float64(baseRating) * ageFactor)

	// Ensure rating doesn't go below 10% of base
	minRating := baseRating / 10
	if adjustedRating < minRating {
		adjustedRating = minRating
	}

	return adjustedRating
}

// GetAgePerformanceFactor returns age-based performance multiplier
func (h *Horse) GetAgePerformanceFactor() float64 {
	switch h.Age {
	case 2:
		return 0.85 // Young horse, not fully developed
	case 3:
		return 0.95 // Developing
	case 4:
		return 1.00 // Prime starts
	case 5:
		return 1.02 // Peak performance
	case 6:
		return 1.00 // Still prime
	case 7:
		return 0.98 // Slight decline begins
	case 8:
		return 0.94 // Noticeable decline
	case 9:
		return 0.88 // Clear aging effects
	case 10:
		return 0.80 // Major decline before retirement
	default:
		if h.Age > 10 {
			return 0.70 // Severe decline for older horses
		}
		return 1.00 // Fallback for unusual ages
	}
}

type Stats struct {
	Stamina   int `json:"stamina"`
	Speed     int `json:"speed"`
	Technique int `json:"technique"`
	Mental    int `json:"mental"`
}

type TrainingType int

const (
	StaminaTraining TrainingType = iota
	SpeedTraining
	TechniqueTraining
	MentalTraining
)

func (t TrainingType) String() string {
	switch t {
	case StaminaTraining:
		return "Stamina"
	case SpeedTraining:
		return "Speed"
	case TechniqueTraining:
		return "Technique"
	case MentalTraining:
		return "Mental"
	default:
		return "Unknown"
	}
}

type TrainingResult struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	StatGain    int    `json:"stat_gain"`
	FatigueGain int    `json:"fatigue_gain"`
	Event       *Event `json:"event,omitempty"`
}

func generateID() string {
	// Generate a UUID-like random ID to prevent collisions
	bytes := make([]byte, 8)
	if _, err := cryptorand.Read(bytes); err != nil {
		// Fallback to timestamp with nanosecond precision if crypto/rand fails
		return time.Now().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(bytes)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// IsStatMaxed checks if the given training type's stat is at maximum
func (h *Horse) IsStatMaxed(trainingType TrainingType) bool {
	switch trainingType {
	case StaminaTraining:
		return h.Stamina >= h.MaxStamina
	case SpeedTraining:
		return h.Speed >= h.MaxSpeed
	case TechniqueTraining:
		return h.Technique >= h.MaxTechnique
	case MentalTraining:
		return h.Mental >= h.MaxMental
	default:
		return false
	}
}

// generateTrainingEvent creates random events during training
func (h *Horse) generateTrainingEvent() *Event {
	// 15% chance for an event to occur
	if rand.Float64() > 0.15 {
		return nil
	}

	events := []Event{
		{
			ID:          "good_weather",
			Name:        "Good Weather Day",
			Description: "☀️ Beautiful weather boosted your horse's spirits!",
			Effects:     map[string]int{"morale": 10},
		},
		{
			ID:          "friendly_visitor",
			Name:        "Friendly Visitor",
			Description: "👋 A friendly fan visited and cheered your horse on!",
			Effects:     map[string]int{"morale": 15, "fan_support": 50},
		},
		{
			ID:          "bad_weather",
			Name:        "Bad Weather",
			Description: "🌧️ Rain made training unpleasant...",
			Effects:     map[string]int{"morale": -5},
		},
		{
			ID:          "injury_scare",
			Name:        "Minor Injury Scare",
			Description: "😰 A stumble scared your horse, but no real injury!",
			Effects:     map[string]int{"morale": -10, "fatigue": 10},
		},
		{
			ID:          "great_workout",
			Name:        "Excellent Training Session",
			Description: "💪 Your horse felt amazing during training!",
			Effects:     map[string]int{"morale": 8, "fatigue": -5},
		},
		{
			ID:          "distracted",
			Name:        "Distracted Training",
			Description: "😵‍💫 Your horse seemed unfocused today...",
			Effects:     map[string]int{"morale": -3},
		},
	}

	// Weight events based on current horse condition
	var weightedEvents []Event
	for _, event := range events {
		// Good events more likely with higher morale
		if event.Effects["morale"] > 0 && h.Morale >= 70 {
			weightedEvents = append(weightedEvents, event, event) // Double chance
		} else if event.Effects["morale"] < 0 && h.Morale <= 40 {
			weightedEvents = append(weightedEvents, event, event) // Double chance for bad events when morale is low
		} else {
			weightedEvents = append(weightedEvents, event)
		}
	}

	if len(weightedEvents) == 0 {
		return nil
	}

	selectedEvent := weightedEvents[rand.Intn(len(weightedEvents))]
	return &selectedEvent
}

// applyEventEffects applies the effects of an event to the horse
func (h *Horse) applyEventEffects(effects map[string]int) {
	for effect, value := range effects {
		switch effect {
		case "morale":
			h.Morale = min(max(h.Morale+value, 20), 100)
		case "fatigue":
			h.Fatigue = min(max(h.Fatigue+value, 0), 100)
		case "fan_support":
			h.FanSupport = max(h.FanSupport+value, 0)
		case "stamina":
			h.Stamina = min(max(h.Stamina+value, 0), h.MaxStamina)
		case "speed":
			h.Speed = min(max(h.Speed+value, 0), h.MaxSpeed)
		case "technique":
			h.Technique = min(max(h.Technique+value, 0), h.MaxTechnique)
		case "mental":
			h.Mental = min(max(h.Mental+value, 0), h.MaxMental)
		}
	}
}

// CalculateDisobedienceChance calculates the chance of disobedience based on horse's condition
func (h *Horse) CalculateDisobedienceChance(whipUsageCount int) float64 {
	baseChance := 0.05 // 5% base chance

	// Fatigue modifier (0-30% additional chance)
	fatigueModifier := float64(h.Fatigue) * 0.003 // 0.3% per fatigue point

	// Morale modifier (-15% to +15% chance)
	moraleModifier := (50.0 - float64(h.Morale)) * 0.003 // 0.3% per point below 50

	// Mental modifier (-10% to +10% chance)
	mentalModifier := (50.0 - float64(h.Mental)) * 0.002 // 0.2% per point below 50

	// Whip abuse modifier (escalating penalty)
	whipAbuseModifier := float64(whipUsageCount) * 0.05 // 5% per recent whip use

	totalChance := baseChance + fatigueModifier + moraleModifier + mentalModifier + whipAbuseModifier

	// Cap between 5% and 80%
	if totalChance < 0.05 {
		return 0.05
	}
	if totalChance > 0.80 {
		return 0.80
	}
	return totalChance
}

// ApplyRaceResults applies morale changes based on race performance
func (h *Horse) ApplyRaceResults(position int, totalEntrants int) {
	// Morale changes based on race position
	switch {
	case position == 1:
		h.Morale = min(h.Morale+20, 100) // Big morale boost for winning
	case position <= 3:
		h.Morale = min(h.Morale+10, 100) // Good morale for podium finish
	case position <= totalEntrants/2:
		h.Morale = min(h.Morale+5, 100) // Small boost for decent finish
	case position <= totalEntrants*3/4:
		// No change for middle finish
	default:
		h.Morale = max(h.Morale-10, 20) // Morale penalty for poor finish
	}

	// Fatigue from racing
	h.Fatigue = min(h.Fatigue+25, 100)
}

// Dope increases all max stats by 50 and costs 5000
func (h *Horse) Dope() bool {
	const dopeCost = 5000
	const maxStatIncrease = 50

	if h.Money < dopeCost {
		return false
	}

	h.Money -= dopeCost
	h.MaxStamina += maxStatIncrease
	h.MaxSpeed += maxStatIncrease
	h.MaxTechnique += maxStatIncrease
	h.MaxMental += maxStatIncrease

	return true
}

// AreAllStatsMaxed checks if all current stats are at their maximum
func (h *Horse) AreAllStatsMaxed() bool {
	return h.Stamina >= h.MaxStamina &&
		h.Speed >= h.MaxSpeed &&
		h.Technique >= h.MaxTechnique &&
		h.Mental >= h.MaxMental
}

func calculateSupporterBonus(supporters []Supporter, trainingType TrainingType) int {
	bonus := 0
	for _, supporter := range supporters {
		if supporter.TrainingBonus[trainingType] > 0 {
			bonus += supporter.TrainingBonus[trainingType]
		}
	}
	return bonus
}
