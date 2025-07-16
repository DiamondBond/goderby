package game

import (
	"fmt"
	"math/rand"

	"goderby/internal/models"
)

type RaceSimulator struct {
	race        models.Race
	horses      map[string]*models.Horse
	strategy    models.RaceStrategy
	playerHorse string
}

func NewRaceSimulator(race models.Race, horses map[string]*models.Horse, playerHorse string, strategy models.RaceStrategy) *RaceSimulator {
	return &RaceSimulator{
		race:        race,
		horses:      horses,
		strategy:    strategy,
		playerHorse: playerHorse,
	}
}

func (rs *RaceSimulator) Simulate() models.RaceResult {
	numTurns := rs.race.Distance / 100 // 100m per turn
	if numTurns < 10 {
		numTurns = 10
	}

	entrants := make([]models.RaceEntrant, 0, len(rs.race.Entrants))
	positions := make(map[string]int)
	distances := make(map[string]int)
	stamina := make(map[string]int)

	// Initialize race state
	for i, horseID := range rs.race.Entrants {
		horse := rs.horses[horseID]
		entrant := models.RaceEntrant{
			HorseID:   horseID,
			HorseName: horse.Name,
			Position:  i + 1,
			Distance:  0,
		}
		entrants = append(entrants, entrant)
		positions[horseID] = i + 1
		distances[horseID] = 0
		stamina[horseID] = horse.Stamina
	}

	var liveProgress []models.RaceProgressUpdate
	var commentary []string

	commentary = append(commentary, "🏁 The race is about to begin!")
	commentary = append(commentary, fmt.Sprintf("🏇 %d horses are lined up at the starting gate", len(rs.race.Entrants)))

	// Race simulation
	for turn := 1; turn <= numTurns; turn++ {
		turnUpdate := models.RaceProgressUpdate{
			Turn:       turn,
			Positions:  make(map[string]int),
			Distances:  make(map[string]int),
			Commentary: "",
			Events:     make([]string, 0),
		}

		// Calculate movement for each horse
		for _, horseID := range rs.race.Entrants {
			horse := rs.horses[horseID]

			// Base movement calculation
			baseSpeed := rs.calculateHorseSpeed(horse, turn, numTurns)

			// Apply strategy modifier if it's the player's horse
			if horseID == rs.playerHorse {
				baseSpeed = rs.applyStrategyModifier(baseSpeed, turn, numTurns)
			}

			// Random factor
			randomFactor := 0.8 + rand.Float64()*0.4 // 0.8 to 1.2
			movement := int(float64(baseSpeed) * randomFactor)

			// Stamina check
			staminaCost := movement / 2
			if stamina[horseID] >= staminaCost {
				distances[horseID] += movement
				stamina[horseID] -= staminaCost
			} else {
				// Reduced movement due to fatigue
				distances[horseID] += movement / 2
				stamina[horseID] = 0
			}

			turnUpdate.Distances[horseID] = distances[horseID]
		}

		// Update positions based on distance
		rs.updatePositions(positions, distances)
		for horseID, pos := range positions {
			turnUpdate.Positions[horseID] = pos
		}

		// Generate commentary
		switch turn {
		case 1:
			turnUpdate.Commentary = "🏁 And they're off!"
		case numTurns / 4:
			leader := rs.getLeader(positions)
			turnUpdate.Commentary = fmt.Sprintf("🎯 At the first quarter: %s takes the lead!", rs.horses[leader].Name)
		case numTurns / 2:
			leader := rs.getLeader(positions)
			turnUpdate.Commentary = fmt.Sprintf("⚡ Halfway point: %s is still in front!", rs.horses[leader].Name)
		case (numTurns * 3) / 4:
			leader := rs.getLeader(positions)
			turnUpdate.Commentary = fmt.Sprintf("🔥 Final quarter: %s leading into the home stretch!", rs.horses[leader].Name)
		case numTurns:
			leader := rs.getLeader(positions)
			turnUpdate.Commentary = fmt.Sprintf("🏆 %s crosses the finish line first!", rs.horses[leader].Name)
		}

		// Random events
		if rand.Float64() < 0.1 { // 10% chance of event
			event := rs.generateRandomEvent()
			if event != "" {
				turnUpdate.Events = append(turnUpdate.Events, event)
			}
		}

		liveProgress = append(liveProgress, turnUpdate)

		if turnUpdate.Commentary != "" {
			commentary = append(commentary, turnUpdate.Commentary)
		}
	}

	// Final results
	finalEntrants := make([]models.RaceEntrant, 0, len(entrants))
	for _, entrant := range entrants {
		entrant.Position = positions[entrant.HorseID]
		entrant.Distance = distances[entrant.HorseID]
		entrant.Time = rs.calculateFinishTime(entrant.Distance, rs.race.Distance)
		finalEntrants = append(finalEntrants, entrant)
	}

	// Sort by position
	for i := 0; i < len(finalEntrants)-1; i++ {
		for j := i + 1; j < len(finalEntrants); j++ {
			if finalEntrants[i].Position > finalEntrants[j].Position {
				finalEntrants[i], finalEntrants[j] = finalEntrants[j], finalEntrants[i]
			}
		}
	}

	// Calculate rewards for player horse
	playerRank := positions[rs.playerHorse]
	prizeMoney := rs.race.GetPrizeForPosition(playerRank)
	fansGained := rs.race.GetFansForPosition(playerRank)

	return models.RaceResult{
		RaceID:       rs.race.ID,
		Results:      finalEntrants,
		PlayerHorse:  rs.playerHorse,
		PlayerRank:   playerRank,
		PrizeMoney:   prizeMoney,
		FansGained:   fansGained,
		Commentary:   commentary,
		LiveProgress: liveProgress,
	}
}

func (rs *RaceSimulator) calculateHorseSpeed(horse *models.Horse, turn, totalTurns int) int {
	// Base speed calculation
	baseSpeed := horse.Speed / 5

	// Technique affects consistency
	techniqueBonus := horse.Technique / 20

	// Mental affects performance under pressure
	mentalBonus := horse.Mental / 25

	// Fatigue penalty
	fatiguePenalty := horse.Fatigue / 10

	// Age affects race performance
	ageFactor := horse.GetAgePerformanceFactor()

	// Stamina affects endurance throughout race
	raceProgress := float64(turn) / float64(totalTurns)
	staminaFactor := 1.0
	if raceProgress > 0.5 {
		staminaFactor = float64(horse.Stamina) / 100.0
	}

	speed := baseSpeed + techniqueBonus + mentalBonus - fatiguePenalty
	speed = int(float64(speed) * staminaFactor * ageFactor)

	if speed < 1 {
		speed = 1
	}

	return speed
}

func (rs *RaceSimulator) applyStrategyModifier(baseSpeed int, turn, totalTurns int) int {
	raceProgress := float64(turn) / float64(totalTurns)

	switch rs.strategy.Formation {
	case models.Lead:
		// Start fast, maintain lead
		if raceProgress < 0.3 {
			return int(float64(baseSpeed) * 1.2)
		}
		return baseSpeed
	case models.Draft:
		// Stay mid-pack, surge in final stretch
		if raceProgress > 0.7 {
			return int(float64(baseSpeed) * 1.3)
		}
		return int(float64(baseSpeed) * 0.9)
	case models.Mount:
		// Conservative start, strong finish
		if raceProgress > 0.8 {
			return int(float64(baseSpeed) * 1.4)
		}
		return int(float64(baseSpeed) * 0.8)
	}

	switch rs.strategy.Pace {
	case models.Fast:
		if raceProgress < 0.5 {
			return int(float64(baseSpeed) * 1.2)
		}
		return int(float64(baseSpeed) * 0.8)
	case models.Even:
		return baseSpeed
	case models.Conserve:
		if raceProgress > 0.6 {
			return int(float64(baseSpeed) * 1.1)
		}
		return int(float64(baseSpeed) * 0.9)
	}

	return baseSpeed
}

func (rs *RaceSimulator) updatePositions(positions map[string]int, distances map[string]int) {
	// Create slice of horse IDs sorted by distance (descending)
	horseIDs := make([]string, 0, len(distances))
	for horseID := range distances {
		horseIDs = append(horseIDs, horseID)
	}

	// Sort by distance (highest first)
	for i := 0; i < len(horseIDs)-1; i++ {
		for j := i + 1; j < len(horseIDs); j++ {
			if distances[horseIDs[i]] < distances[horseIDs[j]] {
				horseIDs[i], horseIDs[j] = horseIDs[j], horseIDs[i]
			}
		}
	}

	// Assign positions
	for i, horseID := range horseIDs {
		positions[horseID] = i + 1
	}
}

func (rs *RaceSimulator) getLeader(positions map[string]int) string {
	for horseID, pos := range positions {
		if pos == 1 {
			return horseID
		}
	}
	return ""
}

func (rs *RaceSimulator) generateRandomEvent() string {
	events := []string{
		"A gust of wind affects the field!",
		"The crowd cheers loudly!",
		"Some horses are bunching up!",
		"The pace is picking up!",
		"A horse stumbles but recovers!",
	}

	if rand.Float64() < 0.5 {
		return events[rand.Intn(len(events))]
	}

	return ""
}

func (rs *RaceSimulator) calculateFinishTime(distance, raceDistance int) string {
	// Simple time calculation based on distance covered
	baseTime := 120.0 // 2 minutes base
	efficiency := float64(distance) / float64(raceDistance)
	if efficiency > 1.0 {
		efficiency = 1.0
	}

	finalTime := baseTime / efficiency
	minutes := int(finalTime) / 60
	seconds := int(finalTime) % 60

	return fmt.Sprintf("%d:%02d", minutes, seconds)
}
