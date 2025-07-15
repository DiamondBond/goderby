package data

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"

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
	horsesPath := filepath.Join(dl.AssetsPath, "horses.json")

	// Create default horses if file doesn't exist
	if _, err := os.Stat(horsesPath); os.IsNotExist(err) {
		horses := dl.generateDefaultHorses()
		if err := dl.saveHorses(horses); err != nil {
			return nil, fmt.Errorf("failed to save default horses: %w", err)
		}
		return horses, nil
	}

	data, err := os.ReadFile(horsesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read horses file: %w", err)
	}

	var horses []models.Horse
	if err := json.Unmarshal(data, &horses); err != nil {
		return nil, fmt.Errorf("failed to unmarshal horses: %w", err)
	}

	return horses, nil
}

func (dl *DataLoader) LoadSupporters() ([]models.Supporter, error) {
	supportersPath := filepath.Join(dl.AssetsPath, "supporters.json")

	// Create default supporters if file doesn't exist
	if _, err := os.Stat(supportersPath); os.IsNotExist(err) {
		supporters := dl.generateDefaultSupporters()
		if err := dl.saveSupporters(supporters); err != nil {
			return nil, fmt.Errorf("failed to save default supporters: %w", err)
		}
		return supporters, nil
	}

	data, err := os.ReadFile(supportersPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read supporters file: %w", err)
	}

	var supporters []models.Supporter
	if err := json.Unmarshal(data, &supporters); err != nil {
		return nil, fmt.Errorf("failed to unmarshal supporters: %w", err)
	}

	return supporters, nil
}

func (dl *DataLoader) LoadRaces() ([]models.Race, error) {
	racesPath := filepath.Join(dl.AssetsPath, "races.json")

	// Create default races if file doesn't exist
	if _, err := os.Stat(racesPath); os.IsNotExist(err) {
		races := dl.generateDefaultRaces()
		if err := dl.saveRaces(races); err != nil {
			return nil, fmt.Errorf("failed to save default races: %w", err)
		}
		return races, nil
	}

	data, err := os.ReadFile(racesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read races file: %w", err)
	}

	var races []models.Race
	if err := json.Unmarshal(data, &races); err != nil {
		return nil, fmt.Errorf("failed to unmarshal races: %w", err)
	}

	return races, nil
}

func (dl *DataLoader) SaveGameState(gameState *models.GameState) error {
	saveDir := filepath.Join(dl.AssetsPath, "saves")
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return fmt.Errorf("failed to create save directory: %w", err)
	}

	savePath := filepath.Join(saveDir, "game.json")
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
	savePath := filepath.Join(dl.AssetsPath, "saves", "game.json")

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
		"Thunder Strike", "Golden Wind", "Storm Runner", "Silver Star",
		"Midnight Express", "Fire Storm", "Ocean Wave", "Sky Dancer",
		"Lightning Bolt", "Crimson Flash", "Diamond Dust", "Emerald Dream",
	}

	breeds := []string{
		"Thoroughbred", "Arabian", "Quarter Horse", "Mustang",
		"Friesian", "Clydesdale", "Appaloosa", "Paint Horse",
	}

	horses := make([]models.Horse, 0, 12)

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
		{
			ID:          "sup_001",
			Name:        "Speed Coach",
			Rarity:      models.Common,
			Description: "Improves speed training effectiveness",
			TrainingBonus: map[models.TrainingType]int{
				models.SpeedTraining: 5,
			},
			IsOwned: true,
		},
		{
			ID:          "sup_002",
			Name:        "Stamina Trainer",
			Rarity:      models.Rare,
			Description: "Boosts stamina and technique training",
			TrainingBonus: map[models.TrainingType]int{
				models.StaminaTraining:   7,
				models.TechniqueTraining: 3,
			},
			IsOwned: true,
		},
		{
			ID:          "sup_003",
			Name:        "Mental Coach",
			Rarity:      models.SuperRare,
			Description: "Expert in mental training and morale",
			TrainingBonus: map[models.TrainingType]int{
				models.MentalTraining: 10,
			},
			IsOwned: false,
		},
		{
			ID:          "sup_004",
			Name:        "Elite Trainer",
			Rarity:      models.UltraRare,
			Description: "Master trainer with balanced bonuses",
			TrainingBonus: map[models.TrainingType]int{
				models.StaminaTraining:   5,
				models.SpeedTraining:     5,
				models.TechniqueTraining: 5,
				models.MentalTraining:    5,
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

func (dl *DataLoader) saveHorses(horses []models.Horse) error {
	data, err := json.MarshalIndent(horses, "", "  ")
	if err != nil {
		return err
	}

	horsesPath := filepath.Join(dl.AssetsPath, "horses.json")
	return os.WriteFile(horsesPath, data, 0644)
}

func (dl *DataLoader) saveSupporters(supporters []models.Supporter) error {
	data, err := json.MarshalIndent(supporters, "", "  ")
	if err != nil {
		return err
	}

	supportersPath := filepath.Join(dl.AssetsPath, "supporters.json")
	return os.WriteFile(supportersPath, data, 0644)
}

func (dl *DataLoader) saveRaces(races []models.Race) error {
	data, err := json.MarshalIndent(races, "", "  ")
	if err != nil {
		return err
	}

	racesPath := filepath.Join(dl.AssetsPath, "races.json")
	return os.WriteFile(racesPath, data, 0644)
}
