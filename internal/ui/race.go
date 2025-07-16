package ui

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"goderby/internal/game"
	"goderby/internal/models"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type RaceModel struct {
	gameState         *models.GameState
	races             []models.Race
	selectedRace      int
	selectedStrat     models.RaceStrategy
	mode              RaceMode
	result            *models.RaceResult
	liveProgress      []models.RaceProgressUpdate
	currentTurn       int
	acquiredSupporter *models.Supporter
}

type RaceMode int

const (
	SelectingRace RaceMode = iota
	SettingStrategy
	ConfirmingEntry
	Racing
	ViewingResult
)

func NewRaceModel(gameState *models.GameState, races []models.Race) RaceModel {
	// Filter races player can enter with progression requirements
	availableRaces := make([]models.Race, 0)
	if gameState.PlayerHorse != nil {
		for _, race := range races {
			if race.CanEnterWithGameState(gameState.PlayerHorse, gameState) {
				availableRaces = append(availableRaces, race)
			}
		}
	}

	return RaceModel{
		gameState:    gameState,
		races:        availableRaces,
		selectedRace: 0,
		selectedStrat: models.RaceStrategy{
			Formation: models.Draft,
			Pace:      models.Even,
		},
		mode: SelectingRace,
	}
}

func (m RaceModel) Init() tea.Cmd {
	return nil
}

func (m RaceModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.mode == Racing {
				return m, nil // Don't allow quitting during race
			}
			return m, func() tea.Msg {
				return NavigationMsg{State: MainMenuView}
			}
		case "esc":
			switch m.mode {
			case ViewingResult:
				m.mode = SelectingRace
				return m, nil
			case SettingStrategy, ConfirmingEntry:
				m.mode = SelectingRace
				return m, nil
			default:
				return m, func() tea.Msg {
					return NavigationMsg{State: MainMenuView}
				}
			}
		case "up", "k":
			switch m.mode {
			case SelectingRace:
				if m.selectedRace > 0 {
					m.selectedRace--
				}
			case SettingStrategy:
				// Cycle through formation
				switch m.selectedStrat.Formation {
				case models.Lead:
					m.selectedStrat.Formation = models.Mount
				case models.Draft:
					m.selectedStrat.Formation = models.Lead
				case models.Mount:
					m.selectedStrat.Formation = models.Draft
				}
			}
		case "down", "j":
			switch m.mode {
			case SelectingRace:
				if m.selectedRace < len(m.races)-1 {
					m.selectedRace++
				}
			case SettingStrategy:
				// Cycle through formation
				switch m.selectedStrat.Formation {
				case models.Lead:
					m.selectedStrat.Formation = models.Draft
				case models.Draft:
					m.selectedStrat.Formation = models.Mount
				case models.Mount:
					m.selectedStrat.Formation = models.Lead
				}
			}
		case "left", "h":
			if m.mode == SettingStrategy {
				// Cycle through pace
				switch m.selectedStrat.Pace {
				case models.Fast:
					m.selectedStrat.Pace = models.Conserve
				case models.Even:
					m.selectedStrat.Pace = models.Fast
				case models.Conserve:
					m.selectedStrat.Pace = models.Even
				}
			}
		case "right", "l":
			if m.mode == SettingStrategy {
				// Cycle through pace
				switch m.selectedStrat.Pace {
				case models.Fast:
					m.selectedStrat.Pace = models.Even
				case models.Even:
					m.selectedStrat.Pace = models.Conserve
				case models.Conserve:
					m.selectedStrat.Pace = models.Fast
				}
			}
		case "enter", " ":
			switch m.mode {
			case SelectingRace:
				if len(m.races) > 0 {
					m.mode = SettingStrategy
				}
			case SettingStrategy:
				m.mode = ConfirmingEntry
			case ConfirmingEntry:
				return m.startRace()
			case ViewingResult:
				// Apply race result and return to main menu
				return m.completeRace()
			}
		}
	case RaceTickMsg:
		if m.mode == Racing {
			if m.currentTurn < len(m.liveProgress) {
				m.currentTurn++
				if m.currentTurn < len(m.liveProgress) {
					return m, tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
						return RaceTickMsg{}
					})
				} else {
					m.mode = ViewingResult
				}
			}
		}
	}

	return m, nil
}

func (m RaceModel) View() string {
	var b strings.Builder

	if m.gameState.PlayerHorse == nil {
		b.WriteString(RenderTitle("Racing"))
		b.WriteString("\n\n")
		b.WriteString(RenderError("No horse selected! Please scout a horse first."))
		b.WriteString("\n\n")
		b.WriteString(RenderHelp("ESC/q to go back"))
		return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
	}

	if len(m.races) == 0 {
		b.WriteString(RenderTitle("Racing"))
		b.WriteString("\n\n")
		b.WriteString(RenderWarning("No races available for your horse's level!"))
		b.WriteString("\n")
		b.WriteString(RenderInfo("Train your horse to unlock more races."))
		b.WriteString("\n\n")
		b.WriteString(RenderHelp("ESC/q to go back"))
		return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
	}

	switch m.mode {
	case ViewingResult:
		return m.renderResultView()
	case Racing:
		return m.renderRaceView()
	case ConfirmingEntry:
		return m.renderConfirmView()
	case SettingStrategy:
		return m.renderStrategyView()
	default:
		return m.renderRaceListView()
	}
}

func (m RaceModel) renderRaceListView() string {
	var b strings.Builder

	b.WriteString(RenderTitle("Available Races"))
	b.WriteString("\n\n")

	horse := m.gameState.PlayerHorse
	b.WriteString(RenderHeader(fmt.Sprintf("%s (Rating: %d)", horse.Name, horse.GetOverallRating())))
	b.WriteString("\n\n")

	for i, race := range m.races {
		cursor := " "
		if m.selectedRace == i {
			cursor = ">"
		}

		raceInfo := fmt.Sprintf("%s 🏁 %s (%s)", cursor, race.Name, race.Grade.String())
		raceInfo += fmt.Sprintf("\n   Distance: %dm | Prize: $%d | Min Rating: %d",
			race.Distance, race.Prize, race.MinRating)

		if m.selectedRace == i {
			b.WriteString(RenderCard(raceInfo, true))
		} else {
			b.WriteString(RenderCard(raceInfo, false))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(RenderHelp("Enter to select race, ↑/↓ to navigate, ESC/q to go back"))

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}

func (m RaceModel) renderStrategyView() string {
	var b strings.Builder

	b.WriteString(RenderTitle("Race Strategy"))
	b.WriteString("\n\n")

	race := m.races[m.selectedRace]
	b.WriteString(RenderHeader(fmt.Sprintf("Strategy for %s", race.Name)))
	b.WriteString("\n\n")

	strategyInfo := fmt.Sprintf("Formation: %s | Pace: %s\n\n", m.selectedStrat.Formation.String(), m.selectedStrat.Pace.String())

	strategyInfo += "Formation Effects:\n"
	strategyInfo += "• Lead: Fast start, maintain position\n"
	strategyInfo += "• Draft: Stay mid-pack, surge at end\n"
	strategyInfo += "• Mount: Conservative start, strong finish\n\n"

	strategyInfo += "Pace Effects:\n"
	strategyInfo += "• Fast: Quick early pace, may tire\n"
	strategyInfo += "• Even: Consistent throughout\n"
	strategyInfo += "• Conserve: Save energy for final push"

	b.WriteString(cardStyle.Render(strategyInfo))
	b.WriteString("\n\n")

	b.WriteString(RenderHelp("↑/↓ for formation, ←/→ for pace, Enter to confirm, ESC to go back"))

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}

func (m RaceModel) renderConfirmView() string {
	var b strings.Builder

	b.WriteString(RenderTitle("Confirm Entry"))
	b.WriteString("\n\n")

	race := m.races[m.selectedRace]
	horse := m.gameState.PlayerHorse

	confirmInfo := fmt.Sprintf("Race: %s (%s)\n", race.Name, race.Grade.String())
	confirmInfo += fmt.Sprintf("Distance: %dm | Prize: $%d\n\n", race.Distance, race.Prize)
	confirmInfo += fmt.Sprintf("Horse: %s (Rating: %d)\n", horse.Name, horse.GetOverallRating())
	confirmInfo += fmt.Sprintf("Formation: %s | Pace: %s\n\n",
		m.selectedStrat.Formation.String(), m.selectedStrat.Pace.String())
	confirmInfo += "Current Status:\n"
	confirmInfo += fmt.Sprintf("Fatigue: %d/100 | Morale: %d/100", horse.Fatigue, horse.Morale)

	b.WriteString(cardStyle.Render(confirmInfo))
	b.WriteString("\n\n")

	if horse.Fatigue > 60 {
		b.WriteString(RenderWarning("Warning: Your horse has high fatigue!"))
		b.WriteString("\n")
	}

	b.WriteString(RenderButton("Enter Race (Enter)", true))
	b.WriteString("\n\n")
	b.WriteString(RenderHelp("Enter to confirm, ESC to go back"))

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}

func (m RaceModel) renderRaceView() string {
	var b strings.Builder

	b.WriteString(RenderTitle("🏁 Live Race 🏁"))
	b.WriteString("\n\n")

	race := m.races[m.selectedRace]
	b.WriteString(RenderHeader(fmt.Sprintf("%s - Turn %d", race.Name, m.currentTurn)))
	b.WriteString("\n\n")

	if m.currentTurn < len(m.liveProgress) {
		progress := m.liveProgress[m.currentTurn]

		// Animated race track with horses
		if len(progress.Positions) > 0 {
			b.WriteString(m.renderAnimatedRaceTrack(progress, race))
			b.WriteString("\n")

			// // Current standings
			// b.WriteString("🏆 Current Standings:\n")
			// b.WriteString(m.renderRaceStandings(progress))
		}

		b.WriteString("\n")

		// Commentary
		if progress.Commentary != "" {
			b.WriteString(RenderInfo("📢 " + progress.Commentary))
			b.WriteString("\n")
		}

		// Events
		for _, event := range progress.Events {
			b.WriteString(RenderWarning("⚡ " + event))
			b.WriteString("\n")
		}
	}

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}

func (m RaceModel) renderAnimatedRaceTrack(progress models.RaceProgressUpdate, race models.Race) string {
	var b strings.Builder
	trackWidth := 60

	// Create a slice of horse positions for sorting
	type HorsePosition struct {
		HorseID  string
		Position int
		Name     string
		Distance int
	}

	var positions []HorsePosition
	for horseID, position := range progress.Positions {
		name := horseID
		// Try to find horse name from results
		for _, entrant := range m.result.Results {
			if entrant.HorseID == horseID {
				name = entrant.HorseName
				break
			}
		}
		// If not found in results, try to get from game state
		if name == horseID && horseID == m.gameState.PlayerHorse.ID {
			name = m.gameState.PlayerHorse.Name
		}

		positions = append(positions, HorsePosition{
			HorseID:  horseID,
			Position: position,
			Name:     name,
			Distance: progress.Distances[horseID],
		})
	}

	// Sort by current distance (race order)
	for i := 0; i < len(positions)-1; i++ {
		for j := i + 1; j < len(positions); j++ {
			if positions[i].Distance < positions[j].Distance {
				positions[i], positions[j] = positions[j], positions[i]
			}
		}
	}

	// Render the race track header
	b.WriteString("🏁 RACE TRACK 🏁\n")
	startLine := "START|"
	finishLine := "|FINISH"
	trackLine := startLine + strings.Repeat("─", trackWidth-len(startLine)-len(finishLine)) + finishLine
	b.WriteString(trackLine + "\n")

	// Render each horse on the track
	for i, pos := range positions {
		if i >= 8 { // Limit to 8 horses to fit on screen
			break
		}

		// Calculate horse position on track (0 to trackWidth-8)
		horsePos := int(float64(pos.Distance) / float64(race.Distance) * float64(trackWidth-8))
		if horsePos < 0 {
			horsePos = 0
		}
		if horsePos > trackWidth-8 {
			horsePos = trackWidth - 8
		}

		// Create the track line with the horse
		trackLine := "     |"
		spaces := strings.Repeat(" ", horsePos)

		// Choose horse animation based on position and speed
		var horseSprite string
		isPlayerHorse := pos.HorseID == m.gameState.PlayerHorse.ID

		// Different horse sprites for animation variety
		horseSprites := []string{
			"🐎", "🏇", "🐴", "🦄",
		}

		// Use different sprite based on horse index for variety
		spriteIndex := i % len(horseSprites)
		if isPlayerHorse {
			// Player horse is always the unicorn for special visibility
			horseSprite = "🦄⭐"
		} else {
			horseSprite = horseSprites[spriteIndex]
		}

		// Add some trailing dust/wind effects for leading horses
		if i <= 2 && pos.Distance > race.Distance/4 {
			horseSprite += "💨"
		}

		trackLine += spaces + horseSprite

		// Fill remaining space and close track
		// Calculate actual sprite width for proper alignment
		spriteWidth := len(horseSprite)
		remainingSpace := trackWidth - len(spaces) - spriteWidth - 6 // 6 for start marker
		if remainingSpace > 0 {
			trackLine += strings.Repeat(" ", remainingSpace)
		}
		trackLine += "|"

		// Horse name and position info
		horseName := pos.Name
		if len(horseName) > 25 {
			horseName = horseName[:22] + "..."
		}

		positionInfo := fmt.Sprintf(" %d. %s", pos.Position, horseName)

		// Highlight player horse
		if isPlayerHorse {
			trackLine = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700")).Bold(true).Render(trackLine)
			positionInfo = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700")).Bold(true).Render(positionInfo + " ⭐")
		}

		b.WriteString(trackLine + positionInfo + "\n")
	}

	// Render the track footer
	b.WriteString(trackLine + "\n")

	// Distance markers
	distanceMarkers := "      "
	quarter := trackWidth / 4
	for i := 0; i < 4; i++ {
		marker := fmt.Sprintf("%dm", (race.Distance/4)*(i+1))
		distanceMarkers += strings.Repeat(" ", quarter-len(marker)/2) + marker
	}
	b.WriteString(distanceMarkers + "\n")

	return b.String()
}

func (m RaceModel) renderRaceStandings(progress models.RaceProgressUpdate) string {
	var b strings.Builder

	// Create a slice of horse positions for sorting by position
	type HorsePosition struct {
		HorseID  string
		Position int
		Name     string
		Distance int
	}

	var positions []HorsePosition
	for horseID, position := range progress.Positions {
		name := horseID
		// Try to find horse name from results
		for _, entrant := range m.result.Results {
			if entrant.HorseID == horseID {
				name = entrant.HorseName
				break
			}
		}
		// If not found in results, try to get from game state
		if name == horseID && horseID == m.gameState.PlayerHorse.ID {
			name = m.gameState.PlayerHorse.Name
		}

		positions = append(positions, HorsePosition{
			HorseID:  horseID,
			Position: position,
			Name:     name,
			Distance: progress.Distances[horseID],
		})
	}

	// Sort by position
	for i := 0; i < len(positions)-1; i++ {
		for j := i + 1; j < len(positions); j++ {
			if positions[i].Position > positions[j].Position {
				positions[i], positions[j] = positions[j], positions[i]
			}
		}
	}

	// Show top 5 positions
	for i, pos := range positions {
		if i >= 5 {
			break
		}

		isPlayerHorse := pos.HorseID == m.gameState.PlayerHorse.ID

		// Position medal/icon
		var posIcon string
		switch pos.Position {
		case 1:
			posIcon = "🥇"
		case 2:
			posIcon = "🥈"
		case 3:
			posIcon = "🥉"
		default:
			posIcon = fmt.Sprintf("%d.", pos.Position)
		}

		// Horse name
		horseName := pos.Name
		if len(horseName) > 30 {
			horseName = horseName[:27] + "..."
		}

		standingLine := fmt.Sprintf("  %s %s", posIcon, horseName)

		// Highlight player horse
		if isPlayerHorse {
			standingLine = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700")).Bold(true).Render(standingLine + " ⭐")
		}

		b.WriteString(standingLine + "\n")
	}

	return b.String()
}

func (m RaceModel) renderResultView() string {
	var b strings.Builder

	b.WriteString(RenderTitle("Race Results"))
	b.WriteString("\n\n")

	race := m.races[m.selectedRace]
	b.WriteString(RenderHeader(race.Name))
	b.WriteString("\n")

	// Player result
	if m.result.PlayerRank <= 3 {
		b.WriteString(RenderSuccess(fmt.Sprintf("🏆 Finished %d%s place!",
			m.result.PlayerRank, getOrdinalSuffix(m.result.PlayerRank))))
	} else {
		b.WriteString(RenderInfo(fmt.Sprintf("Finished %d%s place",
			m.result.PlayerRank, getOrdinalSuffix(m.result.PlayerRank))))
	}
	b.WriteString("\n\n")

	// Rewards
	rewardsInfo := fmt.Sprintf("Prize Money: $%d\n", m.result.PrizeMoney)
	rewardsInfo += fmt.Sprintf("Fans Gained: %d", m.result.FansGained)

	// Show acquired supporter if any
	if m.acquiredSupporter != nil {
		rewardsInfo += fmt.Sprintf("\n\n🎉 New Supporter Acquired!\n")
		rewardsInfo += fmt.Sprintf("%s %s\n", m.acquiredSupporter.Rarity.String(), m.acquiredSupporter.Name)
		rewardsInfo += fmt.Sprintf("📝 %s", m.acquiredSupporter.Description)
	}

	b.WriteString(cardStyle.Render(rewardsInfo))
	b.WriteString("\n\n")

	// Final standings
	b.WriteString(RenderHeader("Final Results"))
	b.WriteString("\n")
	for i, entrant := range m.result.Results {
		if i >= 5 { // Show top 5
			break
		}

		marker := "  "
		isPlayerHorse := entrant.HorseID == m.gameState.PlayerHorse.ID
		if isPlayerHorse {
			marker = "→ "
		}

		resultLine := fmt.Sprintf("%s%d. %s (%s)",
			marker, entrant.Position, entrant.HorseName, entrant.Time)

		// Highlight player's horse line
		if isPlayerHorse {
			resultLine = lipgloss.NewStyle().Foreground(accentColor).Bold(true).Render(resultLine)
		}

		b.WriteString(resultLine + "\n")
	}

	b.WriteString("\n\n")
	b.WriteString(RenderHelp("Enter to continue"))

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}

func (m RaceModel) startRace() (RaceModel, tea.Cmd) {
	// Reset acquired supporter for new race
	m.acquiredSupporter = nil

	race := m.races[m.selectedRace]

	// Add player horse to race
	race.AddEntrant(m.gameState.PlayerHorse.ID)

	// Generate AI horses for the race
	horses := make(map[string]*models.Horse)
	horses[m.gameState.PlayerHorse.ID] = m.gameState.PlayerHorse

	// Add AI opponents (simplified)
	for len(race.Entrants) < race.MaxEntrants && len(race.Entrants) < 8 {
		aiHorse := m.generateAIHorse(race)
		horses[aiHorse.ID] = aiHorse
		race.AddEntrant(aiHorse.ID)
	}

	// Run simulation
	simulator := game.NewRaceSimulator(race, horses, m.gameState.PlayerHorse.ID, m.selectedStrat)
	result := simulator.Simulate()

	m.result = &result
	m.liveProgress = result.LiveProgress
	m.currentTurn = 0
	m.mode = Racing

	return m, tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
		return RaceTickMsg{}
	})
}

func (m RaceModel) completeRace() (RaceModel, tea.Cmd) {
	// Apply race results to player horse
	horse := m.gameState.PlayerHorse
	horse.Money += m.result.PrizeMoney
	horse.FanSupport += m.result.FansGained
	horse.Races++

	if m.result.PlayerRank == 1 {
		horse.Wins++
	}

	// Add some fatigue from racing
	horse.Fatigue += 25
	if horse.Fatigue > 100 {
		horse.Fatigue = 100
	}

	// Update game stats
	m.gameState.GameStats.TotalRaces++
	m.gameState.GameStats.TotalPrizeMoney += m.result.PrizeMoney
	m.gameState.GameStats.TotalFans += m.result.FansGained

	if m.result.PlayerRank == 1 {
		m.gameState.GameStats.TotalWins++
	}

	// Record race completion for progression tracking
	if m.gameState.Season.CompletedRaces == nil {
		m.gameState.Season.CompletedRaces = make([]string, 0)
	}
	if m.gameState.AllCompletedRaces == nil {
		m.gameState.AllCompletedRaces = make([]string, 0)
	}

	if len(m.races) > m.selectedRace {
		raceID := m.races[m.selectedRace].ID
		// Add to current season
		m.gameState.Season.CompletedRaces = append(m.gameState.Season.CompletedRaces, raceID)

		// Add to global completion tracker if not already present
		found := false
		for _, completedID := range m.gameState.AllCompletedRaces {
			if completedID == raceID {
				found = true
				break
			}
		}
		if !found {
			m.gameState.AllCompletedRaces = append(m.gameState.AllCompletedRaces, raceID)
		}

		// Try to acquire supporter based on race performance
		m.TryAcquireSupporter(m.races[m.selectedRace], m.result.PlayerRank)
	}

	return m, func() tea.Msg {
		return NavigationMsg{State: MainMenuView}
	}
}

func (m *RaceModel) TryAcquireSupporter(race models.Race, playerRank int) {
	// Base acquisition chance based on race grade
	var baseChance float64
	var targetRarity models.Rarity

	switch race.Grade {
	case models.GradeG1:
		// Ultra Rare supporters for top tier races
		baseChance = 0.25
		targetRarity = models.UltraRare
	case models.Grade1:
		// Super Rare supporters for Grade 1 races
		baseChance = 0.30
		targetRarity = models.SuperRare
	case models.Grade2:
		// Rare supporters for Grade 2 races
		baseChance = 0.40
		targetRarity = models.Rare
	case models.Grade3:
		// Common supporters for Grade 3 races
		baseChance = 0.50
		targetRarity = models.Common
	default: // MaidenRace
		// Common supporters for Maiden races
		baseChance = 0.30
		targetRarity = models.Common
	}

	// Position multiplier
	var positionMultiplier float64
	switch playerRank {
	case 1:
		positionMultiplier = 1.0
	case 2:
		positionMultiplier = 0.75
	case 3:
		positionMultiplier = 0.50
	default:
		positionMultiplier = 0.25
	}

	finalChance := baseChance * positionMultiplier

	// Roll for acquisition
	if rand.Float64() < finalChance {
		// Find unowned supporter of target rarity
		for i := range m.gameState.Supporters {
			if m.gameState.Supporters[i].Rarity == targetRarity && !m.gameState.Supporters[i].IsOwned {
				// Mark as owned
				m.gameState.Supporters[i].IsOwned = true

				// Store acquired supporter info for UI display
				m.acquiredSupporter = &m.gameState.Supporters[i]
				break
			}
		}
	}
}

func (m RaceModel) generateAIHorse(race models.Race) *models.Horse {
	// Pool of fantasy horse names
	prefixes := []string{"Velvet", "Midnight", "Golden", "Silver", "Crimson", "Sapphire", "Obsidian", "Ethereal", "Aurora", "Phoenix", "Thunder", "Lightning", "Storm", "Mystic", "Nebula", "Starfall", "Copper", "Ivory", "Prism", "Jade", "Opal", "Wildfire", "Cobalt", "Sunset", "Raven", "Glacier", "Twilight", "Amethyst"}
	suffixes := []string{"Thunder", "Mirage", "Legacy", "Grace", "Spirit", "Dreamer", "Zephyr", "Majesty", "Shadow", "Awakening", "Voyager", "Whisper", "Embrace", "Promise", "Flame", "Cascade", "Horizon", "Tempest", "Reverie", "Symphony", "Canyon", "Eclipse", "Strike", "Wind", "Runner", "Star", "Express", "Wave", "Dancer", "Bolt", "Flash", "Dust", "Dream"}

	// Generate a random name by combining prefix + suffix
	prefix := prefixes[rand.Intn(len(prefixes))]
	suffix := suffixes[rand.Intn(len(suffixes))]
	name := prefix + " " + suffix

	// Random horse breeds
	breeds := []string{"Thoroughbred", "Arabian", "Quarter Horse", "Mustang", "Friesian", "Clydesdale", "Appaloosa", "Paint Horse"}
	breed := breeds[rand.Intn(len(breeds))]

	// Generate stats based on race requirements
	baseRating := race.MinRating + (race.MinRating / 4)

	aiHorse := &models.Horse{
		ID:        fmt.Sprintf("ai_%d", len(race.Entrants)),
		Name:      name,
		Breed:     breed,
		Age:       3,
		Stamina:   baseRating + (-10 + (len(race.Entrants) * 5)),
		Speed:     baseRating + (-10 + (len(race.Entrants) * 5)),
		Technique: baseRating + (-10 + (len(race.Entrants) * 5)),
		Mental:    baseRating + (-10 + (len(race.Entrants) * 5)),
		Fatigue:   0,
		Morale:    100,
	}

	return aiHorse
}

func getOrdinalSuffix(n int) string {
	if n >= 11 && n <= 13 {
		return "th"
	}
	switch n % 10 {
	case 1:
		return "st"
	case 2:
		return "nd"
	case 3:
		return "rd"
	default:
		return "th"
	}
}

type RaceTickMsg struct{}
