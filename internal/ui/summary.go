package ui

import (
	"fmt"
	"strings"
	"time"

	"goderby/internal/models"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SummaryModel struct {
	gameState         *models.GameState
	mode              SummaryMode
	canAdvance        bool
	cursor            int
	viewStart         int
	maxVisible        int
	sections          []SummarySection
	raceHistoryCursor int
	raceHistoryStart  int
	maxRacesVisible   int
}

type SummarySection struct {
	Title   string
	Content string
	Icon    string
}

type SummaryMode int

const (
	ViewingSeason SummaryMode = iota
	AdvancingSeason
	RetirementCeremony
	RetirementHomes
	RetiredHorsesGallery
	ShareableProfile
	ShareableSeasonSummary
	ShareableRetirementCard
)

func NewSummaryModel(gameState *models.GameState) *SummaryModel {
	model := &SummaryModel{
		gameState:         gameState,
		mode:              ViewingSeason,
		canAdvance:        gameState.Season.IsComplete(),
		cursor:            0,
		viewStart:         0,
		maxVisible:        3, // Show 3 sections at a time like horse selection
		sections:          []SummarySection{},
		raceHistoryCursor: 0,
		raceHistoryStart:  0,
		maxRacesVisible:   1, // Show 1 race at a time in race history
	}
	model.buildSections()
	return model
}

func (m *SummaryModel) Init() tea.Cmd {
	return nil
}

func (m *SummaryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, func() tea.Msg {
				return NavigationMsg{State: MainMenuView}
			}
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				if m.cursor < m.viewStart {
					m.viewStart = m.cursor
				}
				// Reset race history scroll when changing sections
				m.raceHistoryCursor = 0
				m.raceHistoryStart = 0
			}
			return m, nil
		case "down", "j":
			if m.cursor < len(m.sections)-1 {
				m.cursor++
				if m.cursor >= m.viewStart+m.maxVisible {
					m.viewStart = m.cursor - m.maxVisible + 1
				}
				// Reset race history scroll when changing sections
				m.raceHistoryCursor = 0
				m.raceHistoryStart = 0
			}
			return m, nil
		case "left", "h":
			// Handle horizontal scrolling in race history section
			if m.isRaceHistorySelected() {
				if m.raceHistoryCursor > 0 {
					m.raceHistoryCursor--
					if m.raceHistoryCursor < m.raceHistoryStart {
						m.raceHistoryStart = m.raceHistoryCursor
					}
				}
			}
			return m, nil
		case "right", "l":
			// Handle horizontal scrolling in race history section
			if m.isRaceHistorySelected() {
				if m.raceHistoryCursor < len(m.gameState.Season.CompletedRaces)-1 {
					m.raceHistoryCursor++
					if m.raceHistoryCursor >= m.raceHistoryStart+m.maxRacesVisible {
						m.raceHistoryStart = m.raceHistoryCursor - m.maxRacesVisible + 1
					}
				}
			}
			return m, nil
		case "home":
			m.cursor = 0
			m.viewStart = 0
			return m, nil
		case "end":
			m.cursor = len(m.sections) - 1
			if m.cursor >= m.maxVisible {
				m.viewStart = m.cursor - m.maxVisible + 1
			} else {
				m.viewStart = 0
			}
			return m, nil
		case "esc":
			if m.mode == ViewingSeason {
				return m, func() tea.Msg {
					return NavigationMsg{State: MainMenuView}
				}
			} else {
				m.mode = ViewingSeason
				return m, nil
			}
		case "enter", " ":
			if m.canAdvance && m.mode == ViewingSeason {
				return m.advanceSeason()
			}
		case "n":
			if m.canAdvance {
				return m.advanceSeason()
			}
		case "r":
			if m.gameState.PlayerHorse != nil && m.gameState.PlayerHorse.Age >= 8 {
				m.mode = RetirementCeremony
				return m, nil
			}
		case "g":
			if len(m.gameState.RetiredHorses) > 0 {
				m.mode = RetiredHorsesGallery
				return m, nil
			}
		case "b": // Changed from "h" to avoid conflict with left navigation
			if m.mode == RetirementCeremony {
				m.mode = RetirementHomes
				return m, nil
			}
		case "s":
			if m.gameState.PlayerHorse != nil {
				if m.mode == ViewingSeason {
					m.mode = ShareableSeasonSummary
				} else if m.mode == RetirementCeremony {
					m.mode = ShareableRetirementCard
				} else {
					m.mode = ShareableProfile
				}
				return m, nil
			}
		case "p":
			if m.gameState.PlayerHorse != nil {
				m.mode = ShareableProfile
				return m, nil
			}
		}
	}

	return m, nil
}

func (m *SummaryModel) View() string {
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
	case RetirementCeremony:
		return m.renderRetirementCeremonyView()
	case RetirementHomes:
		return m.renderRetirementHomesView()
	case RetiredHorsesGallery:
		return m.renderRetiredHorsesGalleryView()
	case ShareableProfile:
		return m.renderShareableProfileView()
	case ShareableSeasonSummary:
		return m.renderShareableSeasonSummaryView()
	case ShareableRetirementCard:
		return m.renderShareableRetirementCardView()
	case ViewingSeason:
		return m.renderSeasonSummaryView()
	default:
		return m.renderSeasonSummaryView()
	}
}

func (m *SummaryModel) renderSeasonSummaryView() string {
	var b strings.Builder

	season := m.gameState.Season

	b.WriteString(RenderTitle(fmt.Sprintf("Season %d Summary", season.Number)))
	b.WriteString("\n\n")

	// Calculate visible window
	viewEnd := m.viewStart + m.maxVisible
	if viewEnd > len(m.sections) {
		viewEnd = len(m.sections)
	}

	// Render only visible sections
	for i := m.viewStart; i < viewEnd; i++ {
		section := m.sections[i]
		cursor := "  "
		if m.cursor == i {
			cursor = ">"
		}

		// Section header with cursor
		sectionHeader := fmt.Sprintf("%s %s %s", cursor, section.Icon, section.Title)
		if m.cursor == i {
			sectionHeader = RenderSuccess(sectionHeader)
		}

		b.WriteString(sectionHeader)
		b.WriteString("\n")

		// Section content - rebuild race history dynamically
		content := section.Content
		if m.cursor == i && m.isRaceHistorySelected() {
			content = m.getRaceHistoryScrollable(m.gameState.Season)
		}

		if m.cursor == i {
			b.WriteString(cardStyle.Render(content))
		} else {
			// Show abbreviated content for non-selected sections
			lines := strings.Split(section.Content, "\n")
			if len(lines) > 0 {
				preview := lines[0]
				if len(preview) > 60 {
					preview = preview[:57] + "..."
				}
				b.WriteString(RenderCard(preview, false))
			}
		}
		b.WriteString("\n\n")
	}

	// Scroll indicators
	if len(m.sections) > m.maxVisible {
		scrollInfo := fmt.Sprintf("Section %d of %d", m.cursor+1, len(m.sections))
		if m.viewStart > 0 {
			scrollInfo += " ↑"
		}
		if viewEnd < len(m.sections) {
			scrollInfo += " ↓"
		}
		b.WriteString(RenderInfo(scrollInfo))
		b.WriteString("\n")
	}

	// Action buttons and help text
	b.WriteString(m.renderActionButtons())

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

func (m *SummaryModel) getRaceHistory(season models.Season) string {
	var history strings.Builder

	if len(season.CompletedRaces) == 0 {
		history.WriteString("No races completed this season yet.\n")
		history.WriteString("Visit the Race menu to enter your first race!")
		return history.String()
	}

	history.WriteString(fmt.Sprintf("🏁 Season %d Race History (%d races):\n\n", season.Number, len(season.CompletedRaces)))

	// Group races by unique race ID to avoid duplicates in display
	uniqueRaces := make(map[string]int) // raceID -> count
	for _, raceID := range season.CompletedRaces {
		uniqueRaces[raceID]++
	}

	raceCount := 1
	for raceID, count := range uniqueRaces {
		// Find race details from available races
		var raceDetails *models.Race
		isFallback := false
		for _, race := range m.gameState.AvailableRaces {
			if race.ID == raceID {
				raceDetails = &race
				break
			}
		}

		if raceDetails == nil {
			// Create a fallback race entry with estimated details
			raceDetails = m.createFallbackRace(raceID, count)
			isFallback = true
		}

		// Display race information
		gradeIcon := m.getGradeIcon(raceDetails.Grade)
		nameDisplay := raceDetails.Name
		if isFallback {
			nameDisplay += " (Estimated)"
		}
		history.WriteString(fmt.Sprintf("%d. %s %s (%s)\n", raceCount, gradeIcon, nameDisplay, raceDetails.Grade.String()))
		history.WriteString(fmt.Sprintf("   📏 Distance: %dm | 💰 Prize Pool: $%d\n", raceDetails.Distance, raceDetails.Prize))

		// Show race date if available
		if !raceDetails.Date.IsZero() {
			history.WriteString(fmt.Sprintf("   📅 Date: %s\n", raceDetails.Date.Format("Jan 2, 2006")))
		}

		if count > 1 {
			history.WriteString(fmt.Sprintf("   🔄 Entered %d times this season\n", count))
		}

		// Estimate performance based on current horse stats and race requirements
		performance := m.estimateRacePerformance(raceDetails)
		performanceIcon := m.getPerformanceIcon(performance)
		history.WriteString(fmt.Sprintf("   %s Performance: %s\n", performanceIcon, performance))

		// Add earnings estimate
		earnings := m.estimateEarnings(raceDetails, count)
		if earnings > 0 {
			history.WriteString(fmt.Sprintf("   💵 Est. Earnings: $%d\n", earnings))
		}

		// Add difficulty indicator
		difficulty := m.getRaceDifficulty(raceDetails)
		history.WriteString(fmt.Sprintf("   🎯 Difficulty: %s\n", difficulty))

		history.WriteString("\n")
		raceCount++
	}

	// Add season race statistics summary
	horse := m.gameState.PlayerHorse
	winRate := 0.0
	if horse.Races > 0 {
		winRate = float64(horse.Wins) / float64(horse.Races) * 100
	}

	history.WriteString("📊 Season Racing Summary:\n")
	history.WriteString(fmt.Sprintf("• Total Race Entries: %d\n", len(season.CompletedRaces)))
	history.WriteString(fmt.Sprintf("• Unique Races Competed: %d\n", len(uniqueRaces)))
	history.WriteString(fmt.Sprintf("• Career Record: %d wins in %d races (%.1f%%)\n", horse.Wins, horse.Races, winRate))
	history.WriteString(fmt.Sprintf("• Total Career Earnings: $%d\n", horse.Money))
	history.WriteString(fmt.Sprintf("• Fan Support Gained: %d fans\n", horse.FanSupport))

	// Calculate grade distribution
	gradeStats := make(map[models.RaceGrade]int)
	for raceID := range uniqueRaces {
		for _, race := range m.gameState.AvailableRaces {
			if race.ID == raceID {
				gradeStats[race.Grade]++
				break
			}
		}
	}

	if len(gradeStats) > 0 {
		history.WriteString("• Grade Distribution: ")
		first := true
		for grade, count := range gradeStats {
			if !first {
				history.WriteString(", ")
			}
			history.WriteString(fmt.Sprintf("%s×%d", grade.String(), count))
			first = false
		}
		history.WriteString("\n")
	}

	return history.String()
}

func (m *SummaryModel) estimateRacePerformance(race *models.Race) string {
	horse := m.gameState.PlayerHorse
	horseRating := horse.GetOverallRating()

	// Basic performance estimation based on rating vs race requirements
	if horseRating >= race.MinRating+100 {
		return "Excellent (likely podium finish)"
	} else if horseRating >= race.MinRating+50 {
		return "Good (competitive performance)"
	} else if horseRating >= race.MinRating+25 {
		return "Fair (middle of pack)"
	} else if horseRating >= race.MinRating {
		return "Challenging (met minimum requirements)"
	} else {
		return "Very challenging (below minimum rating)"
	}
}

func (m *SummaryModel) estimateEarnings(race *models.Race, raceCount int) int {
	horse := m.gameState.PlayerHorse

	// Estimate earnings based on likely finishing position
	horseRating := horse.GetOverallRating()
	estimatedPosition := 4 // Default to mid-pack

	if horseRating >= race.MinRating+100 {
		estimatedPosition = 1 // Likely winner
	} else if horseRating >= race.MinRating+50 {
		estimatedPosition = 2 // Likely runner-up
	} else if horseRating >= race.MinRating+25 {
		estimatedPosition = 3 // Likely third
	}

	prizePerRace := race.GetPrizeForPosition(estimatedPosition)
	return prizePerRace * raceCount
}

func (m *SummaryModel) getGradeIcon(grade models.RaceGrade) string {
	switch grade {
	case models.MaidenRace:
		return "🌱" // Beginner
	case models.Grade3:
		return "🥉" // Bronze
	case models.Grade2:
		return "🥈" // Silver
	case models.Grade1:
		return "🥇" // Gold
	case models.GradeG1:
		return "👑" // Crown for top tier
	default:
		return "🏁"
	}
}

func (m *SummaryModel) getPerformanceIcon(performance string) string {
	if strings.Contains(performance, "Excellent") {
		return "🌟"
	} else if strings.Contains(performance, "Good") {
		return "✅"
	} else if strings.Contains(performance, "Fair") {
		return "⚖️"
	} else if strings.Contains(performance, "Challenging") {
		return "⚠️"
	}
	return "📊"
}

func (m *SummaryModel) getRaceDifficulty(race *models.Race) string {
	horse := m.gameState.PlayerHorse
	horseRating := horse.GetOverallRating()
	ratingDiff := horseRating - race.MinRating

	if ratingDiff >= 100 {
		return "Easy ⭐"
	} else if ratingDiff >= 50 {
		return "Moderate ⭐⭐"
	} else if ratingDiff >= 25 {
		return "Hard ⭐⭐⭐"
	} else if ratingDiff >= 0 {
		return "Very Hard ⭐⭐⭐⭐"
	} else {
		return "Extreme ⭐⭐⭐⭐⭐"
	}
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

func (m *SummaryModel) advanceSeason() (*SummaryModel, tea.Cmd) {
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

// renderLifetimeStats renders lifetime career statistics for the horse
func (m SummaryModel) renderLifetimeStats(horse *models.Horse) string {
	var stats strings.Builder

	stats.WriteString("🏆 Career Overview:\n\n")

	// Basic career stats
	winRate := 0.0
	if horse.Races > 0 {
		winRate = float64(horse.Wins) / float64(horse.Races) * 100
	}

	stats.WriteString(fmt.Sprintf("Total Races: %d\n", horse.Races))
	stats.WriteString(fmt.Sprintf("Total Wins: %d (%.1f%%)\n", horse.Wins, winRate))
	stats.WriteString(fmt.Sprintf("Career Earnings: $%d\n", horse.Money))
	stats.WriteString(fmt.Sprintf("Fan Support: %d\n", horse.FanSupport))
	stats.WriteString(fmt.Sprintf("Seasons Competed: %d\n", m.gameState.Season.Number))
	stats.WriteString(fmt.Sprintf("Current Age: %d years\n", horse.Age))
	stats.WriteString(fmt.Sprintf("Peak Rating: %d\n", horse.GetOverallRating()))

	// Career milestones
	stats.WriteString("\n🌟 Career Milestones:\n")
	if horse.Wins >= 10 {
		stats.WriteString("✓ 10+ Race Wins\n")
	} else if horse.Wins >= 5 {
		stats.WriteString("✓ 5+ Race Wins\n")
	} else if horse.Wins >= 1 {
		stats.WriteString("✓ First Race Win\n")
	}

	if horse.Money >= 100000 {
		stats.WriteString("✓ $100,000+ Earnings\n")
	} else if horse.Money >= 50000 {
		stats.WriteString("✓ $50,000+ Earnings\n")
	}

	if horse.FanSupport >= 5000 {
		stats.WriteString("✓ 5,000+ Fans\n")
	} else if horse.FanSupport >= 1000 {
		stats.WriteString("✓ 1,000+ Fans\n")
	}

	if horse.GetOverallRating() >= 200 {
		stats.WriteString("✓ Elite Performance (200+ rating)\n")
	} else if horse.GetOverallRating() >= 150 {
		stats.WriteString("✓ Expert Performance (150+ rating)\n")
	}

	return stats.String()
}

// renderRetirementCeremonyView displays the retirement ceremony with career highlights
func (m SummaryModel) renderRetirementCeremonyView() string {
	var b strings.Builder

	horse := m.gameState.PlayerHorse

	b.WriteString(RenderTitle(fmt.Sprintf("🎉 Retirement Ceremony - %s", horse.Name)))
	b.WriteString("\n\n")

	b.WriteString(RenderSuccess("Congratulations on an amazing career!"))
	b.WriteString("\n\n")

	// Career highlights
	highlights := m.gameState.CalculateCareerHighlights(horse)
	awards := m.gameState.CalculateAwards(horse, highlights)

	b.WriteString(RenderHeader("🏆 Career Highlights"))
	b.WriteString("\n")

	var highlightText strings.Builder
	highlightText.WriteString(fmt.Sprintf("Career Span: %d seasons (%d years old)\n", highlights.CareerLength, horse.Age))
	highlightText.WriteString(fmt.Sprintf("Total Races: %d\n", highlights.TotalRaces))
	highlightText.WriteString(fmt.Sprintf("Total Wins: %d (%.1f%%)\n", highlights.TotalWins, highlights.WinPercentage))
	highlightText.WriteString(fmt.Sprintf("Career Earnings: $%d\n", highlights.TotalPrizeMoney))
	highlightText.WriteString(fmt.Sprintf("Fan Support: %d\n", highlights.TotalFanSupport))
	highlightText.WriteString(fmt.Sprintf("Peak Rating: %d\n", highlights.HighestRating))
	highlightText.WriteString(fmt.Sprintf("Most Prestigious Race: %s\n", highlights.MostPrestigiousRace))

	b.WriteString(cardStyle.Render(highlightText.String()))
	b.WriteString("\n\n")

	// Awards earned
	if len(awards) > 0 {
		b.WriteString(RenderHeader("🏅 Awards Earned"))
		b.WriteString("\n")

		var awardText strings.Builder
		for _, award := range awards {
			awardText.WriteString(fmt.Sprintf("%s %s - %s\n", award.Icon, award.Name, award.Description))
		}

		b.WriteString(cardStyle.Render(awardText.String()))
		b.WriteString("\n\n")
	}

	// Retirement home selection
	b.WriteString(RenderHeader("🏠 Choose Retirement Home"))
	b.WriteString("\n")
	b.WriteString(RenderInfo("Select where your horse will spend their retirement years."))
	b.WriteString("\n\n")

	b.WriteString(RenderButton("Browse Retirement Homes (b)", true))
	b.WriteString("  ")
	b.WriteString(RenderButton("Create Retirement Card (s)", false))
	b.WriteString("\n\n")

	b.WriteString(RenderHelp("b to browse retirement homes, s for shareable card, ESC to go back"))

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}

// renderRetirementHomesView displays available retirement homes
func (m SummaryModel) renderRetirementHomesView() string {
	var b strings.Builder

	b.WriteString(RenderTitle("🏠 Retirement Homes"))
	b.WriteString("\n\n")

	homes := m.gameState.GetAvailableRetirementHomes()

	if len(homes) == 0 {
		b.WriteString(RenderError("No retirement homes available. You need to earn more money to unlock better homes."))
		b.WriteString("\n\n")
		b.WriteString(RenderHelp("ESC to go back"))
		return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
	}

	b.WriteString(RenderHeader("Available Retirement Homes"))
	b.WriteString("\n")

	for i, home := range homes {
		var homeInfo strings.Builder
		homeInfo.WriteString(fmt.Sprintf("🏠 %s\n", home.Name))
		homeInfo.WriteString(fmt.Sprintf("Description: %s\n", home.Description))
		homeInfo.WriteString(fmt.Sprintf("Capacity: %d horses\n", home.Capacity))
		homeInfo.WriteString(fmt.Sprintf("Income Multiplier: %.1fx\n", home.IncomeMultiplier))
		homeInfo.WriteString(fmt.Sprintf("Fame Multiplier: %.1fx\n", home.FameMultiplier))

		if home.Cost > 0 {
			if home.IsOwned {
				homeInfo.WriteString("Status: Owned ✓\n")
			} else {
				homeInfo.WriteString(fmt.Sprintf("Cost: $%d\n", home.Cost))
				if m.gameState.PlayerHorse.Money >= home.Cost {
					homeInfo.WriteString("Status: Available for purchase\n")
				} else {
					homeInfo.WriteString("Status: Cannot afford\n")
				}
			}
		} else {
			homeInfo.WriteString("Status: Free\n")
		}

		b.WriteString(cardStyle.Render(homeInfo.String()))
		if i < len(homes)-1 {
			b.WriteString("\n")
		}
	}

	b.WriteString("\n\n")
	b.WriteString(RenderHelp("ESC to go back to retirement ceremony"))

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}

// renderRetiredHorsesGalleryView displays all retired horses
func (m SummaryModel) renderRetiredHorsesGalleryView() string {
	var b strings.Builder

	b.WriteString(RenderTitle("🖼️ Retired Horses Gallery"))
	b.WriteString("\n\n")

	if len(m.gameState.RetiredHorses) == 0 {
		b.WriteString(RenderInfo("No retired horses yet. Continue playing to build your legacy!"))
		b.WriteString("\n\n")
		b.WriteString(RenderHelp("ESC to go back"))
		return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
	}

	for i, retired := range m.gameState.RetiredHorses {
		var horseCard strings.Builder
		horseCard.WriteString(fmt.Sprintf("🐎 %s (%s)\n", retired.Horse.Name, retired.Horse.Breed))
		horseCard.WriteString(fmt.Sprintf("Retired: %s (Age %d)\n", retired.RetiredAt.Format("2006-01-02"), retired.Horse.Age))
		horseCard.WriteString(fmt.Sprintf("Retirement Home: %s\n", retired.RetirementHome.Name))
		horseCard.WriteString(fmt.Sprintf("Post-Retirement Role: %s\n", retired.PostRetirementRole.String()))
		horseCard.WriteString(fmt.Sprintf("Career Record: %d wins in %d races (%.1f%%)\n",
			retired.CareerHighlights.TotalWins,
			retired.CareerHighlights.TotalRaces,
			retired.CareerHighlights.WinPercentage))
		horseCard.WriteString(fmt.Sprintf("Career Earnings: $%d\n", retired.CareerHighlights.TotalPrizeMoney))
		horseCard.WriteString(fmt.Sprintf("Passive Income: $%d/month\n", retired.PassiveIncome))
		horseCard.WriteString(fmt.Sprintf("Awards: %d earned\n", len(retired.Awards)))

		b.WriteString(cardStyle.Render(horseCard.String()))
		if i < len(m.gameState.RetiredHorses)-1 {
			b.WriteString("\n")
		}
	}

	b.WriteString("\n\n")
	b.WriteString(RenderHelp("ESC to go back"))

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}

// renderShareableProfileView displays a shareable horse profile card
func (m SummaryModel) renderShareableProfileView() string {
	var b strings.Builder

	horse := m.gameState.PlayerHorse

	b.WriteString(RenderTitle("📱 Shareable Horse Profile"))
	b.WriteString("\n\n")

	b.WriteString(RenderInfo("Screenshot this card to share your horse's profile!"))
	b.WriteString("\n\n")

	// Generate shareable profile card
	highlights := m.gameState.CalculateCareerHighlights(horse)
	awards := m.gameState.CalculateAwards(horse, highlights)

	shareableCard := RenderShareableHorseProfile(horse, highlights, awards)
	b.WriteString(shareableCard)

	b.WriteString("\n\n")
	b.WriteString(RenderHelp("ESC to go back"))

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}

// renderShareableSeasonSummaryView displays a shareable season summary card
func (m SummaryModel) renderShareableSeasonSummaryView() string {
	var b strings.Builder

	horse := m.gameState.PlayerHorse
	season := m.gameState.Season
	gameStats := m.gameState.GameStats

	b.WriteString(RenderTitle("📱 Shareable Season Summary"))
	b.WriteString("\n\n")

	b.WriteString(RenderInfo("Screenshot this card to share your season performance!"))
	b.WriteString("\n\n")

	// Generate shareable season summary card
	shareableCard := RenderShareableSeasonSummary(horse, season, gameStats)
	b.WriteString(shareableCard)

	b.WriteString("\n\n")
	b.WriteString(RenderHelp("ESC to go back"))

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}

// renderShareableRetirementCardView displays a shareable retirement card
func (m SummaryModel) renderShareableRetirementCardView() string {
	var b strings.Builder

	horse := m.gameState.PlayerHorse

	b.WriteString(RenderTitle("📱 Shareable Retirement Card"))
	b.WriteString("\n\n")

	b.WriteString(RenderInfo("Screenshot this card to share your horse's retirement ceremony!"))
	b.WriteString("\n\n")

	// Create a temporary retired horse for the card
	highlights := m.gameState.CalculateCareerHighlights(horse)
	awards := m.gameState.CalculateAwards(horse, highlights)

	// Use the basic paddock as default retirement home for preview
	basicHome := m.gameState.RetirementHomes[0]

	tempRetired := models.RetiredHorse{
		Horse:              *horse,
		RetiredAt:          time.Now(),
		RetirementHome:     basicHome,
		PostRetirementRole: models.ShowHorse,
		CareerHighlights:   highlights,
		Awards:             awards,
		PassiveIncome:      highlights.TotalPrizeMoney / 100,
		PassiveFame:        highlights.TotalFanSupport / 50,
		LastPassiveGain:    time.Now(),
	}

	shareableCard := RenderShareableRetirementCard(tempRetired)
	b.WriteString(shareableCard)

	b.WriteString("\n\n")
	b.WriteString(RenderHelp("ESC to go back"))

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}

func (m *SummaryModel) buildSections() {
	horse := m.gameState.PlayerHorse
	season := m.gameState.Season

	// Clear existing sections
	m.sections = []SummarySection{}

	// If no horse selected, return early with empty sections
	if horse == nil {
		return
	}

	// Horse Progress Section
	progressInfo := fmt.Sprintf("Age: %d years old\n", horse.Age)
	progressInfo += fmt.Sprintf("Overall Rating: %d\n", horse.GetOverallRating())
	progressInfo += fmt.Sprintf("Races: %d | Wins: %d (%.1f%%)\n",
		horse.Races, horse.Wins, m.getWinPercentage(horse))
	progressInfo += fmt.Sprintf("Total Fans: %d | Money: $%d", horse.FanSupport, horse.Money)

	m.sections = append(m.sections, SummarySection{
		Title:   fmt.Sprintf("%s's Progress", horse.Name),
		Content: progressInfo,
		Icon:    "🐎",
	})

	// Current Stats Section
	statsCard := m.renderStatsCard(horse)
	m.sections = append(m.sections, SummarySection{
		Title:   "Final Stats",
		Content: statsCard,
		Icon:    "📊",
	})

	// Season Achievements Section
	achievements := m.getSeasonAchievements(season)
	m.sections = append(m.sections, SummarySection{
		Title:   "Season Achievements",
		Content: achievements,
		Icon:    "🏆",
	})

	// Training Summary Section
	trainingSum := m.getTrainingSummary(season)
	m.sections = append(m.sections, SummarySection{
		Title:   "Training Summary",
		Content: trainingSum,
		Icon:    "💪",
	})

	// Race History Section
	raceHistory := m.getRaceHistoryScrollable(season)
	m.sections = append(m.sections, SummarySection{
		Title:   "Race History This Season",
		Content: raceHistory,
		Icon:    "🏁",
	})

	// Lifetime Career Stats Section
	lifetimeStats := m.renderLifetimeStats(horse)
	m.sections = append(m.sections, SummarySection{
		Title:   "Lifetime Career Stats",
		Content: lifetimeStats,
		Icon:    "📈",
	})

	// Social Sharing Section
	sharingContent := "Create shareable cards to show off your horse's achievements!\n\n"
	sharingContent += "• Create Shareable Profile Card (p)\n"
	sharingContent += "• Create Season Summary Card (s)"
	m.sections = append(m.sections, SummarySection{
		Title:   "Social Sharing",
		Content: sharingContent,
		Icon:    "📤",
	})

	// Retired Horses Gallery (if applicable)
	if len(m.gameState.RetiredHorses) > 0 {
		galleryContent := fmt.Sprintf("View your retired horses gallery with %d horses.\n\n", len(m.gameState.RetiredHorses))
		galleryContent += "Browse through your legendary horses and their achievements!"
		m.sections = append(m.sections, SummarySection{
			Title:   "Retired Horses Gallery",
			Content: galleryContent,
			Icon:    "🖼️",
		})
	}
}

func (m *SummaryModel) renderActionButtons() string {
	var b strings.Builder
	horse := m.gameState.PlayerHorse

	// Next season or retirement actions
	if m.canAdvance {
		if horse.Age >= 10 {
			b.WriteString(RenderWarning("Your horse has reached retirement age (10+ years). Time to retire!"))
			b.WriteString("\n\n")
			b.WriteString(RenderButton("Begin Retirement Ceremony (r)", true))
			b.WriteString("\n\n")
		} else if horse.Age >= 8 {
			b.WriteString(RenderWarning("Your horse is getting old. Consider retirement or continue for one more season."))
			b.WriteString("\n\n")
			b.WriteString(RenderButton("Advance to Next Season (Enter/n)", true))
			b.WriteString("  ")
			b.WriteString(RenderButton("Retire Horse (r)", false))
			b.WriteString("\n\n")
		} else {
			b.WriteString(RenderSuccess("Ready to advance to next season!"))
			b.WriteString("\n\n")
			b.WriteString(RenderButton("Advance to Next Season (Enter/n)", true))
			b.WriteString("\n\n")
		}

		var helpText string
		if horse.Age >= 8 {
			helpText = "Enter/n to advance season, r to retire"
		} else {
			helpText = "Enter/n to advance season"
		}
		if len(m.gameState.RetiredHorses) > 0 {
			helpText += ", g for gallery"
		}
		helpText += ", p/s for shareable cards, ↑↓/j/k to navigate, ←→/h/l for race history, ESC/q to go back"
		b.WriteString(RenderHelp(helpText))
	} else {
		season := m.gameState.Season
		b.WriteString(RenderInfo(fmt.Sprintf("Season in progress - Week %d/%d", season.CurrentWeek, season.MaxWeeks)))
		b.WriteString("\n\n")

		var helpText string
		if horse.Age >= 8 {
			helpText = "r to retire"
		}
		if len(m.gameState.RetiredHorses) > 0 {
			if helpText != "" {
				helpText += ", "
			}
			helpText += "g for gallery"
		}
		if helpText != "" {
			helpText += ", "
		}
		helpText += "p/s for shareable cards, ↑↓/j/k to navigate, ←→/h/l for race history, ESC/q to go back"
		b.WriteString(RenderHelp(helpText))
	}

	return b.String()
}

func (m *SummaryModel) isRaceHistorySelected() bool {
	return m.cursor < len(m.sections) && m.sections[m.cursor].Title == "Race History This Season"
}

func (m *SummaryModel) getUniqueRaces() map[string]int {
	season := m.gameState.Season
	uniqueRaces := make(map[string]int)
	for _, raceID := range season.CompletedRaces {
		uniqueRaces[raceID]++
	}
	return uniqueRaces
}

func (m *SummaryModel) getRaceHistoryScrollable(season models.Season) string {
	var history strings.Builder

	if len(season.CompletedRaces) == 0 {
		history.WriteString("No races completed this season yet.\n")
		history.WriteString("Visit the Race menu to enter your first race!")
		return history.String()
	}

	history.WriteString(fmt.Sprintf("🏁 Season %d Race History (%d races):\n\n", season.Number, len(season.CompletedRaces)))

	// Show only one race at a time based on cursor position
	if m.raceHistoryCursor < len(season.CompletedRaces) {
		raceID := season.CompletedRaces[m.raceHistoryCursor]

		// Find race details
		var raceDetails *models.Race
		isFallback := false
		for _, race := range m.gameState.AvailableRaces {
			if race.ID == raceID {
				raceDetails = &race
				break
			}
		}

		if raceDetails == nil {
			// For fallback, just pass 1 as count since we're showing individual entries
			raceDetails = m.createFallbackRace(raceID, 1)
			isFallback = true
		}

		// Display race information
		gradeIcon := m.getGradeIcon(raceDetails.Grade)
		nameDisplay := raceDetails.Name
		if isFallback {
			nameDisplay += " (Estimated)"
		}

		history.WriteString(fmt.Sprintf("%d. %s %s (%s)\n", m.raceHistoryCursor+1, gradeIcon, nameDisplay, raceDetails.Grade.String()))
		history.WriteString(fmt.Sprintf("   📏 Distance: %dm | 💰 Prize Pool: $%d\n", raceDetails.Distance, raceDetails.Prize))

		// Show race date if available
		if !raceDetails.Date.IsZero() {
			history.WriteString(fmt.Sprintf("   📅 Date: %s\n", raceDetails.Date.Format("Jan 2, 2006")))
		}

		// Estimate performance based on current horse stats and race requirements
		performance := m.estimateRacePerformance(raceDetails)
		performanceIcon := m.getPerformanceIcon(performance)
		history.WriteString(fmt.Sprintf("   %s Performance: %s\n", performanceIcon, performance))

		// Add earnings estimate (for single race entry)
		earnings := m.estimateEarnings(raceDetails, 1)
		if earnings > 0 {
			history.WriteString(fmt.Sprintf("   💵 Est. Earnings: $%d\n", earnings))
		}

		// Add difficulty indicator
		difficulty := m.getRaceDifficulty(raceDetails)
		history.WriteString(fmt.Sprintf("   🎯 Difficulty: %s\n", difficulty))
	}

	// Show scroll indicators
	if len(season.CompletedRaces) > 1 {
		history.WriteString("\n\n")
		scrollInfo := fmt.Sprintf("Race %d of %d", m.raceHistoryCursor+1, len(season.CompletedRaces))
		if m.raceHistoryCursor > 0 {
			scrollInfo += " ←"
		}
		if m.raceHistoryCursor < len(season.CompletedRaces)-1 {
			scrollInfo += " →"
		}
		scrollInfo += " (use ←→/h/l to navigate races)"
		history.WriteString(scrollInfo)
	}

	// Add season summary at the bottom
	if m.raceHistoryCursor == len(season.CompletedRaces)-1 || len(season.CompletedRaces) <= 1 {
		history.WriteString("\n\n")
		horse := m.gameState.PlayerHorse
		winRate := 0.0
		if horse.Races > 0 {
			winRate = float64(horse.Wins) / float64(horse.Races) * 100
		}

		// Calculate unique races for summary
		uniqueRaces := m.getUniqueRaces()

		history.WriteString("📊 Season Racing Summary:\n")
		history.WriteString(fmt.Sprintf("• Total Race Entries: %d\n", len(season.CompletedRaces)))
		history.WriteString(fmt.Sprintf("• Unique Races Competed: %d\n", len(uniqueRaces)))
		history.WriteString(fmt.Sprintf("• Career Record: %d wins in %d races (%.1f%%)", horse.Wins, horse.Races, winRate))
	}

	return history.String()
}

func (m *SummaryModel) createFallbackRace(raceID string, entryCount int) *models.Race {
	// Create a fallback race with estimated details based on game progression
	horse := m.gameState.PlayerHorse
	season := m.gameState.Season

	// Estimate race grade based on horse's progression and season
	var grade models.RaceGrade
	var raceName string
	var distance int
	var prize int
	var minRating int

	// Use horse's current rating and season to estimate what races they likely entered
	horseRating := horse.GetOverallRating()

	if horseRating < 50 {
		grade = models.MaidenRace
		raceName = "Maiden Stakes"
		distance = 1600
		prize = 5000
		minRating = 0
	} else if horseRating < 120 {
		grade = models.Grade3
		raceName = "Provincial Stakes"
		distance = 1800
		prize = 15000
		minRating = 80
	} else if horseRating < 160 {
		grade = models.Grade2
		raceName = "Regional Championship"
		distance = 2000
		prize = 30000
		minRating = 120
	} else if horseRating < 200 {
		grade = models.Grade1
		raceName = "Classic Stakes"
		distance = 2200
		prize = 50000
		minRating = 160
	} else {
		grade = models.GradeG1
		raceName = "Premier Championship"
		distance = 2400
		prize = 75000
		minRating = 200
	}

	// Add variety based on entry count and season number
	if entryCount > 2 {
		raceName += " (Recurring)"
	}
	if season.Number > 1 {
		raceName = fmt.Sprintf("Season %d %s", season.Number, raceName)
	}

	// Create the fallback race
	return &models.Race{
		ID:          raceID,
		Name:        raceName,
		Distance:    distance,
		Grade:       grade,
		Prize:       prize,
		MinRating:   minRating,
		MaxEntrants: 16,
		Date:        season.SeasonStartDate.AddDate(0, 0, (season.CurrentWeek-1)*7),
		Entrants:    []string{},
	}
}
