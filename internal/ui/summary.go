package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"goderby/internal/models"
)

type SummaryModel struct {
	gameState  *models.GameState
	mode       SummaryMode
	canAdvance bool
}

type SummaryMode int

const (
	ViewingSeason SummaryMode = iota
	AdvancingSeason
)

func NewSummaryModel(gameState *models.GameState) SummaryModel {
	return SummaryModel{
		gameState:  gameState,
		mode:       ViewingSeason,
		canAdvance: gameState.Season.IsComplete(),
	}
}

func (m SummaryModel) Init() tea.Cmd {
	return nil
}

func (m SummaryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, func() tea.Msg {
				return NavigationMsg{State: MainMenuView}
			}
		case "esc":
			return m, func() tea.Msg {
				return NavigationMsg{State: MainMenuView}
			}
		case "enter", " ":
			if m.canAdvance && m.mode == ViewingSeason {
				return m.advanceSeason()
			}
		case "n":
			if m.canAdvance {
				return m.advanceSeason()
			}
		}
	}

	return m, nil
}

func (m SummaryModel) View() string {
	var b strings.Builder

	if m.gameState.PlayerHorse == nil {
		b.WriteString(RenderTitle("Season Summary"))
		b.WriteString("\n\n")
		b.WriteString(RenderError("No horse selected! Please scout a horse first."))
		b.WriteString("\n\n")
		b.WriteString(RenderHelp("ESC/q to go back"))
		return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
	}

	switch m.mode {
	case AdvancingSeason:
		return m.renderSeasonAdvanceView()
	default:
		return m.renderSeasonSummaryView()
	}
}

func (m SummaryModel) renderSeasonSummaryView() string {
	var b strings.Builder

	horse := m.gameState.PlayerHorse
	season := m.gameState.Season

	b.WriteString(RenderTitle(fmt.Sprintf("Season %d Summary", season.Number)))
	b.WriteString("\n\n")

	// Horse progress
	b.WriteString(RenderHeader(fmt.Sprintf("%s's Progress", horse.Name)))
	b.WriteString("\n")

	progressInfo := fmt.Sprintf("Age: %d years old\n", horse.Age)
	progressInfo += fmt.Sprintf("Overall Rating: %d\n", horse.GetOverallRating())
	progressInfo += fmt.Sprintf("Races: %d | Wins: %d (%.1f%%)\n",
		horse.Races, horse.Wins, m.getWinPercentage(horse))
	progressInfo += fmt.Sprintf("Total Fans: %d | Money: $%d\n", horse.FanSupport, horse.Money)

	b.WriteString(cardStyle.Render(progressInfo))
	b.WriteString("\n\n")

	// Current stats
	b.WriteString(RenderHeader("Final Stats"))
	b.WriteString("\n")
	statsCard := m.renderStatsCard(horse)
	b.WriteString(cardStyle.Render(statsCard))
	b.WriteString("\n\n")

	// Season achievements
	b.WriteString(RenderHeader("Season Achievements"))
	b.WriteString("\n")
	achievements := m.getSeasonAchievements(season)
	b.WriteString(cardStyle.Render(achievements))
	b.WriteString("\n\n")

	// Training summary
	b.WriteString(RenderHeader("Training Summary"))
	b.WriteString("\n")
	trainingSum := m.getTrainingSummary(season)
	b.WriteString(cardStyle.Render(trainingSum))
	b.WriteString("\n\n")

	// Next season
	if m.canAdvance {
		if horse.Age >= 8 {
			b.WriteString(RenderWarning("Your horse is getting old. Consider retirement after this season."))
		} else {
			b.WriteString(RenderSuccess("Ready to advance to next season!"))
		}
		b.WriteString("\n\n")
		b.WriteString(RenderButton("Advance to Next Season (Enter/n)", true))
		b.WriteString("\n\n")
		b.WriteString(RenderHelp("Enter/n to advance season, ESC/q to go back"))
	} else {
		b.WriteString(RenderInfo(fmt.Sprintf("Season in progress - Week %d/%d", season.CurrentWeek, season.MaxWeeks)))
		b.WriteString("\n\n")
		b.WriteString(RenderHelp("ESC/q to go back"))
	}

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}

func (m SummaryModel) renderSeasonAdvanceView() string {
	var b strings.Builder

	horse := m.gameState.PlayerHorse
	newSeason := m.gameState.Season.Number + 1

	b.WriteString(RenderTitle("Season Advanced!"))
	b.WriteString("\n\n")

	b.WriteString(RenderSuccess(fmt.Sprintf("Welcome to Season %d!", newSeason)))
	b.WriteString("\n\n")

	// Age progression
	if horse.Age < 8 {
		b.WriteString(RenderInfo(fmt.Sprintf("%s is now %d years old", horse.Name, horse.Age)))
	} else {
		b.WriteString(RenderWarning(fmt.Sprintf("%s is now %d years old - approaching retirement age", horse.Name, horse.Age)))
	}
	b.WriteString("\n\n")

	// Season goals
	goals := m.generateSeasonGoals(newSeason)
	b.WriteString(RenderHeader("Season Goals"))
	b.WriteString("\n")
	b.WriteString(cardStyle.Render(goals))
	b.WriteString("\n\n")

	b.WriteString(RenderInfo("Returning to main menu..."))
	b.WriteString("\n\n")
	b.WriteString(RenderHelp("The new season has begun! Time to get back to training."))

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}

func (m SummaryModel) renderStatsCard(horse *models.Horse) string {
	var stats strings.Builder

	stats.WriteString(RenderStatBar("Stamina", horse.Stamina, horse.MaxStamina))
	stats.WriteString("\n")
	stats.WriteString(RenderStatBar("Speed", horse.Speed, horse.MaxSpeed))
	stats.WriteString("\n")
	stats.WriteString(RenderStatBar("Technique", horse.Technique, horse.MaxTechnique))
	stats.WriteString("\n")
	stats.WriteString(RenderStatBar("Mental", horse.Mental, horse.MaxMental))

	return stats.String()
}

func (m SummaryModel) getWinPercentage(horse *models.Horse) float64 {
	if horse.Races == 0 {
		return 0.0
	}
	return float64(horse.Wins) / float64(horse.Races) * 100.0
}

func (m SummaryModel) getSeasonAchievements(season models.Season) string {
	var achievements strings.Builder

	achievements.WriteString("🏆 Achievements This Season:\n\n")

	// Count training days and races
	trainingDays := 0
	restDays := 0
	racesCompleted := len(season.CompletedRaces)

	for _, day := range season.TrainingDays {
		if day.IsCompleted {
			if day.IsRest {
				restDays++
			} else {
				trainingDays++
			}
		}
	}

	achievements.WriteString(fmt.Sprintf("✓ Completed %d training sessions\n", trainingDays))
	achievements.WriteString(fmt.Sprintf("✓ Took %d rest days\n", restDays))
	achievements.WriteString(fmt.Sprintf("✓ Participated in %d races\n", racesCompleted))

	horse := m.gameState.PlayerHorse
	if horse.Wins > 0 {
		achievements.WriteString(fmt.Sprintf("✓ Won %d races!\n", horse.Wins))
	}

	if horse.GetOverallRating() >= 200 {
		achievements.WriteString("✓ Achieved Elite rating (200+)!\n")
	} else if horse.GetOverallRating() >= 150 {
		achievements.WriteString("✓ Achieved Expert rating (150+)!\n")
	}

	if horse.FanSupport >= 10000 {
		achievements.WriteString("✓ Gained 10,000+ fans!\n")
	} else if horse.FanSupport >= 5000 {
		achievements.WriteString("✓ Gained 5,000+ fans!\n")
	}

	return achievements.String()
}

func (m SummaryModel) getTrainingSummary(season models.Season) string {
	var summary strings.Builder

	// Count training by type
	staminaTraining := 0
	speedTraining := 0
	techniqueTraining := 0
	mentalTraining := 0

	for _, day := range season.TrainingDays {
		if day.IsCompleted && !day.IsRest {
			switch day.TrainingType {
			case models.StaminaTraining:
				staminaTraining++
			case models.SpeedTraining:
				speedTraining++
			case models.TechniqueTraining:
				techniqueTraining++
			case models.MentalTraining:
				mentalTraining++
			}
		}
	}

	summary.WriteString("Training Focus This Season:\n\n")
	summary.WriteString(fmt.Sprintf("Stamina Training: %d sessions\n", staminaTraining))
	summary.WriteString(fmt.Sprintf("Speed Training: %d sessions\n", speedTraining))
	summary.WriteString(fmt.Sprintf("Technique Training: %d sessions\n", techniqueTraining))
	summary.WriteString(fmt.Sprintf("Mental Training: %d sessions\n", mentalTraining))

	// Identify focus area
	maxTraining := staminaTraining
	focusArea := "Stamina"

	if speedTraining > maxTraining {
		maxTraining = speedTraining
		focusArea = "Speed"
	}
	if techniqueTraining > maxTraining {
		maxTraining = techniqueTraining
		focusArea = "Technique"
	}
	if mentalTraining > maxTraining {
		maxTraining = mentalTraining
		focusArea = "Mental"
	}

	if maxTraining > 0 {
		summary.WriteString(fmt.Sprintf("\nPrimary focus: %s", focusArea))
	}

	return summary.String()
}

func (m SummaryModel) generateSeasonGoals(seasonNumber int) string {
	goals := "Goals for this season:\n\n"

	horse := m.gameState.PlayerHorse
	rating := horse.GetOverallRating()

	if rating < 100 {
		goals += "• Reach 100+ overall rating\n"
		goals += "• Win your first race\n"
		goals += "• Gain 1,000 fans\n"
	} else if rating < 150 {
		goals += "• Reach 150+ overall rating\n"
		goals += "• Win 3+ races this season\n"
		goals += "• Compete in Grade 2 races\n"
	} else if rating < 200 {
		goals += "• Reach 200+ overall rating\n"
		goals += "• Win a Grade 1 race\n"
		goals += "• Gain 5,000+ fans\n"
	} else {
		goals += "• Maintain elite performance\n"
		goals += "• Win the Grand Prix\n"
		goals += "• Become a legend\n"
	}

	goals += fmt.Sprintf("• Complete Season %d successfully", seasonNumber)

	return goals
}

func (m SummaryModel) advanceSeason() (SummaryModel, tea.Cmd) {
	// Age the horse
	m.gameState.PlayerHorse.Age++

	// Preserve completed races history before creating new season
	for _, raceID := range m.gameState.Season.CompletedRaces {
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
	}

	// Create new season
	newSeason := models.NewSeason(m.gameState.Season.Number + 1)
	m.gameState.Season = newSeason

	// Reset horse condition
	horse := m.gameState.PlayerHorse
	horse.Fatigue = 0
	horse.Morale = 100

	// Update game stats
	m.gameState.GameStats.SeasonsCompleted++

	// Check for retirement
	if horse.Age >= 10 {
		horse.IsRetired = true
	}

	m.mode = AdvancingSeason
	m.canAdvance = false

	return m, tea.Batch(
		tea.Tick(time.Second*3, func(t time.Time) tea.Msg {
			return NavigationMsg{State: MainMenuView}
		}),
	)
}
