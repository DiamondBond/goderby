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
	// Interactive racing controls
	playerLane       int  // Current lane (0-4, left to right)
	raceStamina      int  // Current race stamina (separate from training fatigue)
	whipUses         int  // Times whip has been used this race
	obedienceCounter int  // Counter for disobedience effect
	isDisobedient    bool // Whether horse is currently disobedient
	lastWhipTurn     int  // Last turn whip was used
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
		mode:             SelectingRace,
		playerLane:       2,   // Start in middle lane
		raceStamina:      100, // Start with full race stamina
		whipUses:         0,
		obedienceCounter: 0,
		isDisobedient:    false,
		lastWhipTurn:     0,
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
			case SelectingRace:
				return m, func() tea.Msg {
					return NavigationMsg{State: MainMenuView}
				}
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
			case ViewingResult, ConfirmingEntry, Racing:
				// No action for these modes
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
			case ViewingResult, ConfirmingEntry, Racing:
				// No action for these modes
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
			} else if m.mode == Racing {
				// Move left during race
				if m.playerLane > 0 && !m.isDisobedient {
					m.playerLane--
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
			} else if m.mode == Racing {
				// Move right during race
				if m.playerLane < 4 && !m.isDisobedient {
					m.playerLane++
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
				race := m.races[m.selectedRace]
				entryFee := race.GetEntryFee()
				if m.gameState.PlayerHorse.Money >= entryFee {
					return m.startRace()
				}
				// If can't afford, do nothing (stay in confirm view)
			case ViewingResult:
				// Apply race result and return to main menu
				return m.completeRace()
			case Racing:
				// Whip during race
				return m.useWhip()
			}
		case "w":
			if m.mode == Racing {
				// Alternative whip key
				return m.useWhip()
			}
		}
	case RaceTickMsg:
		if m.mode == Racing {
			if m.currentTurn < len(m.liveProgress) {
				m.currentTurn++

				// Apply interactive modifiers to current turn if applicable
				m.applyInteractiveModifiers()

				// Update obedience state
				m.updateObedience()

				// Regenerate stamina slowly
				if m.raceStamina < 100 {
					m.raceStamina += 2
					if m.raceStamina > 100 {
						m.raceStamina = 100
					}
				}

				if m.currentTurn < len(m.liveProgress) {
					return m, tea.Tick(time.Millisecond*1500, func(t time.Time) tea.Msg {
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
	case SelectingRace:
		return m.renderRaceListView()
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
		raceInfo += fmt.Sprintf("\n   Distance: %dm | Prize: $%d | Entry Fee: $%d",
			race.Distance, race.Prize, race.GetEntryFee())
		raceInfo += fmt.Sprintf("\n   Min Rating: %d", race.MinRating)

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
	entryFee := race.GetEntryFee()

	confirmInfo := fmt.Sprintf("Race: %s (%s)\n", race.Name, race.Grade.String())
	confirmInfo += fmt.Sprintf("Distance: %dm | Prize: $%d\n", race.Distance, race.Prize)
	confirmInfo += fmt.Sprintf("Entry Fee: $%d\n\n", entryFee)
	confirmInfo += fmt.Sprintf("Horse: %s (Rating: %d)\n", horse.Name, horse.GetOverallRating())
	confirmInfo += fmt.Sprintf("Money: $%d\n", horse.Money)
	confirmInfo += fmt.Sprintf("Formation: %s | Pace: %s\n\n",
		m.selectedStrat.Formation.String(), m.selectedStrat.Pace.String())
	confirmInfo += "Current Status:\n"
	confirmInfo += fmt.Sprintf("Fatigue: %d/100 | Morale: %d/100", horse.Fatigue, horse.Morale)

	b.WriteString(cardStyle.Render(confirmInfo))
	b.WriteString("\n\n")

	// Check for warnings
	if horse.Fatigue > 60 {
		b.WriteString(RenderWarning("Warning: Your horse has high fatigue!"))
		b.WriteString("\n")
	}
	if horse.Morale < 50 {
		b.WriteString(RenderWarning("Warning: Your horse has low morale!"))
		b.WriteString("\n")
	}
	if horse.Money < entryFee {
		b.WriteString(RenderError("Error: Not enough money for entry fee!"))
		b.WriteString("\n")
	}

	canAfford := horse.Money >= entryFee
	b.WriteString(RenderButton("Enter Race (Enter)", canAfford))
	b.WriteString("\n\n")
	if canAfford {
		b.WriteString(RenderHelp("Enter to confirm, ESC to go back"))
	} else {
		b.WriteString(RenderHelp("Need more money to enter! ESC to go back"))
	}

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}

func (m RaceModel) renderRaceView() string {
	var b strings.Builder

	b.WriteString(RenderTitle("🏁 Live Race 🏁"))
	b.WriteString("\n\n")

	race := m.races[m.selectedRace]
	b.WriteString(RenderHeader(fmt.Sprintf("%s - Turn %d", race.Name, m.currentTurn)))
	b.WriteString("\n\n")

	// Player controls and status
	b.WriteString(m.renderPlayerStatus())
	b.WriteString("\n")

	if m.currentTurn < len(m.liveProgress) {
		progress := m.liveProgress[m.currentTurn]

		// Animated race track with horses
		if len(progress.Positions) > 0 {
			b.WriteString(m.renderAnimatedRaceTrack(progress, race))
			b.WriteString("\n")

			// Current standings
			b.WriteString("🏆 Current Standings:\n")
			b.WriteString(m.renderRaceStandings(progress))
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

	// Controls help
	b.WriteString(m.renderControlsHelp())

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
			// Player horse shows lane position
			laneMarker := fmt.Sprintf("L%d", m.playerLane+1)
			horseSprite = "🦄" + laneMarker
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
		rewardsInfo += "\n\n🎉 New Supporter Acquired!\n"
		rewardsInfo += fmt.Sprintf("%s %s\n", m.acquiredSupporter.Rarity.String(), m.acquiredSupporter.Name)
		rewardsInfo += "📝 " + m.acquiredSupporter.Description
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
	race := m.races[m.selectedRace]
	entryFee := race.GetEntryFee()

	// Check if player can afford entry fee
	if m.gameState.PlayerHorse.Money < entryFee {
		// Stay in confirm view, the UI will show the error
		return m, nil
	}

	// Charge entry fee
	m.gameState.PlayerHorse.Money -= entryFee

	// Reset acquired supporter for new race
	m.acquiredSupporter = nil

	// Reset interactive racing controls
	m.playerLane = 2    // Start in middle lane
	m.raceStamina = 100 // Full stamina at start
	m.whipUses = 0
	m.obedienceCounter = 0
	m.isDisobedient = false
	m.lastWhipTurn = 0

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

	return m, tea.Tick(time.Millisecond*1500, func(t time.Time) tea.Msg {
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

	// Apply race results including morale and fatigue changes
	totalEntrants := len(m.result.Results)
	horse.ApplyRaceResults(m.result.PlayerRank, totalEntrants)

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

func (m *RaceModel) useWhip() (RaceModel, tea.Cmd) {
	// Can't whip if disobedient or too soon after last whip
	if m.isDisobedient || m.currentTurn-m.lastWhipTurn < 3 {
		return *m, nil
	}

	// Check if enough stamina to whip
	staminaCost := 25
	if m.raceStamina < staminaCost {
		return *m, nil
	}

	// Use whip - consume stamina and track usage
	m.raceStamina -= staminaCost
	m.whipUses++
	m.lastWhipTurn = m.currentTurn

	// Enhanced disobedience calculation using horse stats
	horse := m.gameState.PlayerHorse
	disobedienceChance := horse.CalculateDisobedienceChance(m.whipUses)

	if rand.Float64() < disobedienceChance {
		m.isDisobedient = true
		// Duration of disobedience varies based on mental stat
		baseDuration := 5
		mentalModifier := (horse.Mental - 50) / 20               // Better mental = shorter disobedience
		m.obedienceCounter = max(baseDuration-mentalModifier, 2) // Min 2 turns, max varies
	}

	return *m, nil
}

func (m *RaceModel) updateObedience() {
	if m.isDisobedient {
		m.obedienceCounter--
		if m.obedienceCounter <= 0 {
			m.isDisobedient = false
		}
	}
}

func (m *RaceModel) applyInteractiveModifiers() {
	if m.currentTurn >= len(m.liveProgress) || m.result == nil {
		return
	}

	// Get current progress
	progress := &m.liveProgress[m.currentTurn]
	playerHorseID := m.gameState.PlayerHorse.ID

	// Apply whip boost to distance
	if m.lastWhipTurn > 0 && m.currentTurn-m.lastWhipTurn <= 2 {
		// Whip boost: increase distance by a percentage
		currentDistance := progress.Distances[playerHorseID]
		whipBoost := int(float64(currentDistance) * 0.1) // 10% boost to current distance
		progress.Distances[playerHorseID] = currentDistance + whipBoost

		// Add whip boost event
		if m.currentTurn-m.lastWhipTurn == 1 {
			progress.Events = append(progress.Events, "💨 Your horse surges forward from the whip!")
		}
	}

	// Apply lane effects to distance based on race position (turns vs straights)
	lane := m.playerLane
	currentDistance := progress.Distances[playerHorseID]

	// Calculate if we're in a turn section
	numTurns := len(m.liveProgress)
	raceProgress := float64(m.currentTurn) / float64(numTurns)
	isInTurn := (raceProgress >= 0.0 && raceProgress <= 0.25) || (raceProgress >= 0.75 && raceProgress <= 1.0)

	var laneBonus int
	if isInTurn {
		// During turns, inner lanes get advantage
		switch lane {
		case 0: // Inner rail - best advantage in turns
			laneBonus = 5
		case 1: // Second lane - good advantage in turns
			laneBonus = 3
		case 2: // Middle lane - neutral in turns
			laneBonus = 0
		case 3: // Outer middle - slight disadvantage in turns
			laneBonus = -2
		case 4: // Outside lane - significant disadvantage in turns
			laneBonus = -4
		}
	} else {
		// On straights, middle lanes are optimal
		switch lane {
		case 0, 4: // Outside lanes - slight disadvantage on straights
			laneBonus = -2
		case 1, 3: // Good lanes on straights
			laneBonus = 0
		case 2: // Perfect middle lane on straights
			laneBonus = 3
		}
	}

	progress.Distances[playerHorseID] = currentDistance + laneBonus

	// Add turn advantage events for visual feedback
	if isInTurn && laneBonus > 0 {
		progress.Events = append(progress.Events, fmt.Sprintf("🏃 Inner lane advantage! +%d speed in the turn!", laneBonus))
	} else if isInTurn && laneBonus < 0 {
		progress.Events = append(progress.Events, fmt.Sprintf("🐌 Outside lane disadvantage! %d speed in the turn", laneBonus))
	}

	// Apply disobedience penalty
	if m.isDisobedient {
		currentDistance = progress.Distances[playerHorseID]
		penalty := int(float64(currentDistance) * 0.15) // 15% penalty
		progress.Distances[playerHorseID] = currentDistance - penalty

		// Add disobedience event
		if m.obedienceCounter == 5 { // First turn of disobedience
			progress.Events = append(progress.Events, "🚫 Your horse is fighting your commands!")
		}
	}

	// Recalculate positions after distance changes
	m.recalculatePositions(progress)
}

func (m *RaceModel) recalculatePositions(progress *models.RaceProgressUpdate) {
	// Create slice of horse IDs sorted by distance (descending)
	type HorseDistance struct {
		HorseID  string
		Distance int
	}

	var horses []HorseDistance
	for horseID, distance := range progress.Distances {
		horses = append(horses, HorseDistance{
			HorseID:  horseID,
			Distance: distance,
		})
	}

	// Sort by distance (highest first)
	for i := 0; i < len(horses)-1; i++ {
		for j := i + 1; j < len(horses); j++ {
			if horses[i].Distance < horses[j].Distance {
				horses[i], horses[j] = horses[j], horses[i]
			}
		}
	}

	// Update positions
	for i, horse := range horses {
		progress.Positions[horse.HorseID] = i + 1
	}
}

func (m RaceModel) renderPlayerStatus() string {
	// Lane indicators
	laneDisplay := ""
	for i := 0; i < 5; i++ {
		if i == m.playerLane {
			laneDisplay += "[🦄]"
		} else {
			laneDisplay += "[ ]"
		}
		if i < 4 {
			laneDisplay += " "
		}
	}

	// Stamina bar
	staminaBar := ""
	staminaBars := m.raceStamina / 5 // 20 bars max
	for i := 0; i < 20; i++ {
		if i < staminaBars {
			staminaBar += "█"
		} else {
			staminaBar += "▁"
		}
	}

	statusInfo := fmt.Sprintf("Lane Position: %s\n", laneDisplay)
	statusInfo += fmt.Sprintf("Stamina: %s %d/100\n", staminaBar, m.raceStamina)
	statusInfo += fmt.Sprintf("Whip Uses: %d", m.whipUses)

	if m.isDisobedient {
		statusInfo += fmt.Sprintf("\n🚫 DISOBEDIENT (%d turns)", m.obedienceCounter)
	}

	// Style the status box
	statusStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FFD700")).
		Padding(0, 1).
		Width(60)

	return statusStyle.Render(statusInfo)
}

func (m RaceModel) renderControlsHelp() string {
	// Calculate if we're in a turn section for strategic guidance
	numTurns := len(m.liveProgress)
	raceProgress := float64(m.currentTurn) / float64(numTurns)
	isInTurn := (raceProgress >= 0.0 && raceProgress <= 0.25) || (raceProgress >= 0.75 && raceProgress <= 1.0)

	var controlsText string
	if isInTurn {
		controlsText = "🎮 IN TURN: Inner lanes (←) give speed advantage! | Enter/W Whip horse | Too much whipping = disobedience!"
	} else {
		controlsText = "🎮 Controls: ←/→ Switch lanes | Enter/W Whip horse (+speed, -stamina) | Middle lanes best on straights!"
	}

	if m.isDisobedient {
		controlsText = "🚫 Horse is disobedient! Controls disabled temporarily."
	} else if m.raceStamina < 25 {
		controlsText = "⚠️  Low stamina! Can't use whip until stamina recovers."
	} else if m.currentTurn-m.lastWhipTurn < 3 && m.lastWhipTurn > 0 {
		cooldown := 3 - (m.currentTurn - m.lastWhipTurn)
		controlsText = fmt.Sprintf("⏳ Whip cooldown: %d turns", cooldown)
	}

	return RenderHelp(controlsText)
}

// Interface methods for InteractiveRaceModel
func (m RaceModel) GetPlayerLane() int {
	return m.playerLane
}

func (m RaceModel) GetRaceStamina() int {
	return m.raceStamina
}

func (m RaceModel) GetWhipUses() int {
	return m.whipUses
}

func (m RaceModel) IsDisobedient() bool {
	return m.isDisobedient
}

func (m RaceModel) GetLastWhipTurn() int {
	return m.lastWhipTurn
}

func (m RaceModel) GetCurrentTurn() int {
	return m.currentTurn
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
