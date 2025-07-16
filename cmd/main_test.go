package main

import (
	"fmt"
	"testing"

	"goderby/internal/data"
	"goderby/internal/models"
)

func TestGameInitialization(t *testing.T) {
	// Test data loader
	loader := data.NewDataLoader("../assets")

	// Create a new game state for testing
	gameState := models.NewGameState()

	// Test loading horses
	horses, err := loader.LoadHorses(gameState)
	if err != nil {
		t.Errorf("Failed to load horses: %v", err)
	}

	if len(horses) == 0 {
		t.Error("No horses loaded")
	}

	fmt.Printf("Loaded %d horses\n", len(horses))

	// Test loading supporters
	supporters, err := loader.LoadSupporters(gameState)
	if err != nil {
		t.Errorf("Failed to load supporters: %v", err)
	}

	if len(supporters) == 0 {
		t.Error("No supporters loaded")
	}

	fmt.Printf("Loaded %d supporters\n", len(supporters))

	// Test loading races
	races, err := loader.LoadRaces(gameState)
	if err != nil {
		t.Errorf("Failed to load races: %v", err)
	}

	if len(races) == 0 {
		t.Error("No races loaded")
	}

	fmt.Printf("Loaded %d races\n", len(races))
}

func TestHorseTraining(t *testing.T) {
	// Create a test horse
	baseStats := models.Stats{
		Stamina:   50,
		Speed:     50,
		Technique: 50,
		Mental:    50,
	}

	horse := models.NewHorse("Test Horse", "Test Breed", baseStats)

	// Test training
	supporters := []models.Supporter{}
	result := horse.Train(models.StaminaTraining, supporters)

	if !result.Success {
		t.Errorf("Training failed: %s", result.Message)
	}

	if result.StatGain <= 0 {
		t.Error("Expected stat gain from training")
	}

	fmt.Printf("Training result: %+v\n", result)
}

func TestRaceCreation(t *testing.T) {
	race := models.NewRace("Test Race", 1600, models.Grade1, 50000, 100)

	if race.Name != "Test Race" {
		t.Error("Race name not set correctly")
	}

	if race.Distance != 1600 {
		t.Error("Race distance not set correctly")
	}

	fmt.Printf("Created race: %+v\n", race)
}

func TestGameState(t *testing.T) {
	gameState := models.NewGameState()

	if gameState.Season.Number != 1 {
		t.Error("Season should start at 1")
	}

	if gameState.Season.CurrentWeek != 1 {
		t.Error("Week should start at 1")
	}

	fmt.Printf("Game state initialized: Season %d, Week %d\n",
		gameState.Season.Number, gameState.Season.CurrentWeek)
}
