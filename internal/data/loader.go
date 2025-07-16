package data

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"

	"goderby/internal/models"
)

type DataLoader struct {
	AssetsPath string
}

func NewDataLoader(assetsPath string) *DataLoader {
	return &DataLoader{
		AssetsPath: assetsPath,
	}
}

func (dl *DataLoader) LoadHorses() ([]models.Horse, error) {
	return dl.generateDefaultHorses(), nil
}

func (dl *DataLoader) LoadSupporters() ([]models.Supporter, error) {
	return dl.generateDefaultSupporters(), nil
}

func (dl *DataLoader) LoadRaces() ([]models.Race, error) {
	return dl.generateDefaultRaces(), nil
}

func (dl *DataLoader) SaveGameState(gameState *models.GameState) error {
	savePath := "game.json"
	data, err := json.MarshalIndent(gameState, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal game state: %w", err)
	}

	if err := os.WriteFile(savePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write save file: %w", err)
	}

	return nil
}

func (dl *DataLoader) LoadGameState() (*models.GameState, error) {
	// Load from exe directory as game.json
	savePath := "game.json"

	if _, err := os.Stat(savePath); os.IsNotExist(err) {
		return models.NewGameState(), nil
	}

	data, err := os.ReadFile(savePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read save file: %w", err)
	}

	var gameState models.GameState
	if err := json.Unmarshal(data, &gameState); err != nil {
		return nil, fmt.Errorf("failed to unmarshal game state: %w", err)
	}

	return &gameState, nil
}

func (dl *DataLoader) generateDefaultHorses() []models.Horse {
	horseNames := []string{
		"Velvet Thunder", "Midnight Mirage", "Golden Legacy", "Silver Grace",
		"Crimson Spirit", "Sapphire Dreamer", "Obsidian Zephyr", "Ethereal Majesty",
		"Aurora Shadow", "Phoenix Awakening", "Thunder Voyager", "Lightning Whisper",
		"Storm Embrace", "Mystic Promise", "Nebula Flame", "Starfall Cascade",
		"Copper Horizon", "Ivory Tempest", "Prism Reverie", "Jade Symphony",
		"Opal Canyon", "Wildfire Eclipse", "Cobalt Strike", "Sunset Wind",
		"Raven Runner", "Glacier Star", "Twilight Express", "Amethyst Wave",
	}

	breeds := []string{
		"Thoroughbred", "Arabian", "Quarter Horse", "Mustang",
		"Friesian", "Clydesdale", "Appaloosa", "Paint Horse",
	}

	horses := make([]models.Horse, 0, 28)

	for i, name := range horseNames {
		baseStats := models.Stats{
			Stamina:   50 + rand.Intn(30),
			Speed:     50 + rand.Intn(30),
			Technique: 50 + rand.Intn(30),
			Mental:    50 + rand.Intn(30),
		}

		horse := models.NewHorse(name, breeds[i%len(breeds)], baseStats)
		horses = append(horses, *horse)
	}

	return horses
}

func (dl *DataLoader) generateDefaultSupporters() []models.Supporter {
	supporters := []models.Supporter{
		// Common supporters (basic single-stat bonuses)
		{
			ID:          "sup_001",
			Name:        "Speed Coach",
			Rarity:      models.Common,
			Description: "Improves speed training effectiveness",
			TrainingBonus: map[models.TrainingType]int{
				models.SpeedTraining: 5,
			},
			IsOwned: false,
		},
		{
			ID:          "sup_002",
			Name:        "Endurance Trainer",
			Rarity:      models.Common,
			Description: "Focuses on stamina building",
			TrainingBonus: map[models.TrainingType]int{
				models.StaminaTraining: 5,
			},
			IsOwned: false,
		},
		{
			ID:          "sup_003",
			Name:        "Technique Specialist",
			Rarity:      models.Common,
			Description: "Helps perfect racing technique",
			TrainingBonus: map[models.TrainingType]int{
				models.TechniqueTraining: 5,
			},
			IsOwned: false,
		},
		{
			ID:          "sup_004",
			Name:        "Mental Coach",
			Rarity:      models.Common,
			Description: "Basic mental training support",
			TrainingBonus: map[models.TrainingType]int{
				models.MentalTraining: 5,
			},
			IsOwned: false,
		},
		// Rare supporters (dual-stat bonuses)
		{
			ID:          "sup_005",
			Name:        "Sprint Master",
			Rarity:      models.Rare,
			Description: "Combines speed and technique training",
			TrainingBonus: map[models.TrainingType]int{
				models.SpeedTraining:     7,
				models.TechniqueTraining: 3,
			},
			IsOwned: false,
		},
		{
			ID:          "sup_006",
			Name:        "Stamina Expert",
			Rarity:      models.Rare,
			Description: "Boosts stamina and mental resilience",
			TrainingBonus: map[models.TrainingType]int{
				models.StaminaTraining: 7,
				models.MentalTraining:  3,
			},
			IsOwned: false,
		},
		{
			ID:          "sup_007",
			Name:        "Racing Tactician",
			Rarity:      models.Rare,
			Description: "Enhances technique and mental focus",
			TrainingBonus: map[models.TrainingType]int{
				models.TechniqueTraining: 7,
				models.MentalTraining:    3,
			},
			IsOwned: false,
		},
		{
			ID:          "sup_008",
			Name:        "Power Trainer",
			Rarity:      models.Rare,
			Description: "Builds speed and stamina together",
			TrainingBonus: map[models.TrainingType]int{
				models.SpeedTraining:   6,
				models.StaminaTraining: 4,
			},
			IsOwned: false,
		},
		// Super Rare supporters (tri-stat bonuses)
		{
			ID:          "sup_009",
			Name:        "Derby Champion",
			Rarity:      models.SuperRare,
			Description: "Former champion with vast experience",
			TrainingBonus: map[models.TrainingType]int{
				models.StaminaTraining:   8,
				models.SpeedTraining:     5,
				models.TechniqueTraining: 2,
			},
			IsOwned: false,
		},
		{
			ID:          "sup_010",
			Name:        "Mind & Body Coach",
			Rarity:      models.SuperRare,
			Description: "Holistic training approach",
			TrainingBonus: map[models.TrainingType]int{
				models.MentalTraining:    8,
				models.TechniqueTraining: 5,
				models.StaminaTraining:   2,
			},
			IsOwned: false,
		},
		{
			ID:          "sup_011",
			Name:        "Speed Virtuoso",
			Rarity:      models.SuperRare,
			Description: "Master of speed and finesse",
			TrainingBonus: map[models.TrainingType]int{
				models.SpeedTraining:     8,
				models.TechniqueTraining: 4,
				models.MentalTraining:    3,
			},
			IsOwned: false,
		},
		// Ultra Rare supporters (quad-stat bonuses)
		{
			ID:          "sup_012",
			Name:        "Legendary Trainer",
			Rarity:      models.UltraRare,
			Description: "Master trainer with balanced expertise",
			TrainingBonus: map[models.TrainingType]int{
				models.StaminaTraining:   5,
				models.SpeedTraining:     5,
				models.TechniqueTraining: 5,
				models.MentalTraining:    5,
			},
			IsOwned: false,
		},
		{
			ID:          "sup_013",
			Name:        "Triple Crown Winner",
			Rarity:      models.UltraRare,
			Description: "Elite champion with winning mentality",
			TrainingBonus: map[models.TrainingType]int{
				models.SpeedTraining:     6,
				models.TechniqueTraining: 6,
				models.MentalTraining:    6,
				models.StaminaTraining:   2,
			},
			IsOwned: false,
		},
		{
			ID:          "sup_014",
			Name:        "Miracle Worker",
			Rarity:      models.UltraRare,
			Description: "Transforms any horse into a champion",
			TrainingBonus: map[models.TrainingType]int{
				models.StaminaTraining:   8,
				models.SpeedTraining:     4,
				models.TechniqueTraining: 4,
				models.MentalTraining:    4,
			},
			IsOwned: false,
		},
	}

	return supporters
}

func (dl *DataLoader) generateDefaultRaces() []models.Race {
	races := []*models.Race{
		models.NewRace("Maiden Stakes", 1600, models.MaidenRace, 5000, 0),
		models.NewRace("Spring Classic", 2000, models.Grade3, 15000, 120),
		models.NewRace("Summer Derby", 2400, models.Grade2, 30000, 150),
		models.NewRace("Autumn Championship", 2000, models.Grade1, 50000, 180),
		models.NewRace("Winter Cup", 1800, models.Grade1, 75000, 200),
		models.NewRace("Grand Prix", 2500, models.GradeG1, 100000, 220),
	}

	result := make([]models.Race, len(races))
	for i, race := range races {
		result[i] = *race
	}

	return result
}
