package models

import (
	"crypto/rand"
	"encoding/hex"
	"math"
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
	// Use proper rounding instead of truncation for morale calculation
	moraleMultiplier := float64(h.Morale) / 100.0
	actualGain := int(math.Round(float64(baseGain) * moraleMultiplier))

	switch trainingType {
	case StaminaTraining:
		if h.Stamina < h.MaxStamina {
			h.Stamina = min(h.Stamina+actualGain, h.MaxStamina)
		}
	case SpeedTraining:
		if h.Speed < h.MaxSpeed {
			h.Speed = min(h.Speed+actualGain, h.MaxSpeed)
		}
	case TechniqueTraining:
		if h.Technique < h.MaxTechnique {
			h.Technique = min(h.Technique+actualGain, h.MaxTechnique)
		}
	case MentalTraining:
		if h.Mental < h.MaxMental {
			h.Mental = min(h.Mental+actualGain, h.MaxMental)
		}
	}

	h.Fatigue += 15
	if h.Fatigue > 100 {
		h.Fatigue = 100
	}

	return TrainingResult{
		Success:     true,
		Message:     "Training completed successfully!",
		StatGain:    actualGain,
		FatigueGain: 15,
	}
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
	if _, err := rand.Read(bytes); err != nil {
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

func calculateSupporterBonus(supporters []Supporter, trainingType TrainingType) int {
	bonus := 0
	for _, supporter := range supporters {
		if supporter.TrainingBonus[trainingType] > 0 {
			bonus += supporter.TrainingBonus[trainingType]
		}
	}
	return bonus
}
