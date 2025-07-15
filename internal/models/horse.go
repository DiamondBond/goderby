package models

import (
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
	actualGain := int(float64(baseGain) * (float64(h.Morale) / 100.0))

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
	return (h.Stamina + h.Speed + h.Technique + h.Mental) / 4
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
	return time.Now().Format("20060102150405")
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
