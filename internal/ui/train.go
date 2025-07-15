package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"goderby/internal/models"
)

type TrainModel struct {
	gameState     *models.GameState
	selectedDay   int
	selectedType  models.TrainingType
	trainingTypes []models.TrainingType
	mode          TrainingMode
	weekCompleted bool
	lastResult    *models.TrainingResult
}

type TrainingMode int

const (
	SelectingDay TrainingMode = iota
	SelectingType
	Confirming
	ViewingTrainingResult
)

func NewTrainModel(gameState *models.GameState) TrainModel {
	return TrainModel{
		gameState:    gameState,
		selectedDay:  0,
		selectedType: models.StaminaTraining,
		trainingTypes: []models.TrainingType{
			models.StaminaTraining,
			models.SpeedTraining,
			models.TechniqueTraining,
			models.MentalTraining,
		},
		mode:          SelectingDay,
		weekCompleted: false,
	}
}

func (m TrainModel) Init() tea.Cmd {
	return nil
}

func (m TrainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, func() tea.Msg {
				return NavigationMsg{State: MainMenuView}
			}
		case "esc":
			if m.mode == ViewingTrainingResult {
				m.mode = SelectingDay
				return m, nil
			}
			return m, func() tea.Msg {
				return NavigationMsg{State: MainMenuView}
			}
		case "up", "k":
			switch m.mode {
			case SelectingDay:
				if m.selectedDay > 0 {
					m.selectedDay--
				}
			case SelectingType:
				idx := int(m.selectedType)
				if idx > 0 {
					m.selectedType = m.trainingTypes[idx-1]
				}
			}
		case "down", "j":
			switch m.mode {
			case SelectingDay:
				if m.selectedDay < 5 { // 6 days (0-5)
					m.selectedDay++
				}
			case SelectingType:
				idx := int(m.selectedType)
				if idx < len(m.trainingTypes)-1 {
					m.selectedType = m.trainingTypes[idx+1]
				}
			}
		case "enter", " ":
			switch m.mode {
			case SelectingDay:
				if m.canTrainToday() {
					m.mode = SelectingType
				}
			case SelectingType:
				m.mode = Confirming
			case Confirming:
				return m.performTraining()
			case ViewingTrainingResult:
				m.mode = SelectingDay
				if m.isWeekComplete() {
					return m, func() tea.Msg {
						return WeekCompleteMsg{}
					}
				}
			}
		case "r":
			if m.mode == SelectingDay && m.canTrainToday() {
				return m.performRest()
			}
		case "n":
			if m.mode == SelectingDay && m.isWeekComplete() {
				return m, func() tea.Msg {
					return WeekCompleteMsg{}
				}
			}
		}
	}

	return m, nil
}

func (m TrainModel) View() string {
	var b strings.Builder

	if m.gameState.PlayerHorse == nil {
		b.WriteString(RenderTitle("Training"))
		b.WriteString("\n\n")
		b.WriteString(RenderError("No horse selected! Please scout a horse first."))
		b.WriteString("\n\n")
		b.WriteString(RenderHelp("ESC/q to go back"))
		return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
	}

	horse := m.gameState.PlayerHorse

	switch m.mode {
	case ViewingTrainingResult:
		return m.renderResultView(horse)
	case Confirming:
		return m.renderConfirmView(horse)
	case SelectingType:
		return m.renderTypeSelectionView(horse)
	default:
		return m.renderCalendarView(horse)
	}
}

func (m TrainModel) renderCalendarView(horse *models.Horse) string {
	var b strings.Builder

	b.WriteString(RenderTitle("Training Calendar"))
	b.WriteString("\n\n")

	// Horse status
	b.WriteString(RenderHeader(fmt.Sprintf("%s - Week %d", horse.Name, m.gameState.Season.CurrentWeek)))
	b.WriteString("\n")

	statusInfo := fmt.Sprintf("Overall Rating: %d | Fatigue: %d/100 | Morale: %d/100",
		horse.GetOverallRating(), horse.Fatigue, horse.Morale)
	b.WriteString(cardStyle.Render(statusInfo))
	b.WriteString("\n\n")

	// Stats
	b.WriteString(RenderHeader("Current Stats"))
	b.WriteString("\n")
	statsCard := m.renderStatsCard(horse)
	b.WriteString(cardStyle.Render(statsCard))
	b.WriteString("\n\n")

	// Training calendar
	b.WriteString(RenderHeader("This Week's Training"))
	b.WriteString("\n")

	days := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	for i, day := range days {
		cursor := " "
		if m.selectedDay == i {
			cursor = ">"
		}

		status := m.getDayStatus(i)
		dayInfo := fmt.Sprintf("%s %s - %s", cursor, day, status)

		if m.selectedDay == i {
			b.WriteString(selectedMenuItemStyle.Render(dayInfo))
		} else {
			b.WriteString(menuItemStyle.Render(dayInfo))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Help text
	if m.isWeekComplete() {
		b.WriteString(RenderSuccess("Week completed!"))
		b.WriteString("\n")
		b.WriteString(RenderHelp("Press 'n' to advance to next week, ESC/q to go back"))
	} else if m.canTrainToday() {
		b.WriteString(RenderHelp("Enter to train, 'r' to rest, ↑/↓ to navigate, ESC/q to go back"))
	} else {
		b.WriteString(RenderHelp("↑/↓ to navigate, ESC/q to go back"))
	}

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}

func (m TrainModel) renderTypeSelectionView(horse *models.Horse) string {
	var b strings.Builder

	b.WriteString(RenderTitle("Select Training Type"))
	b.WriteString("\n\n")

	day := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}[m.selectedDay]
	b.WriteString(RenderHeader(fmt.Sprintf("Training for %s", day)))
	b.WriteString("\n\n")

	for _, trainingType := range m.trainingTypes {
		cursor := " "
		if m.selectedType == trainingType {
			cursor = ">"
		}

		current, max := m.getStatValues(horse, trainingType)
		typeInfo := fmt.Sprintf("%s %s Training (Current: %d/%d)",
			cursor, trainingType.String(), current, max)

		if m.selectedType == trainingType {
			b.WriteString(selectedMenuItemStyle.Render(typeInfo))
		} else {
			b.WriteString(menuItemStyle.Render(typeInfo))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(RenderHelp("Enter to confirm, ↑/↓ to navigate, ESC to go back"))

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}

func (m TrainModel) renderConfirmView(horse *models.Horse) string {
	var b strings.Builder

	b.WriteString(RenderTitle("Confirm Training"))
	b.WriteString("\n\n")

	day := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}[m.selectedDay]
	b.WriteString(RenderHeader(fmt.Sprintf("%s - %s Training", day, m.selectedType.String())))
	b.WriteString("\n")

	confirmInfo := fmt.Sprintf("Horse: %s\n", horse.Name)
	confirmInfo += fmt.Sprintf("Current Fatigue: %d/100\n", horse.Fatigue)
	confirmInfo += fmt.Sprintf("Current Morale: %d/100\n", horse.Morale)

	current, max := m.getStatValues(horse, m.selectedType)
	confirmInfo += fmt.Sprintf("Current %s: %d/%d", m.selectedType.String(), current, max)

	b.WriteString(cardStyle.Render(confirmInfo))
	b.WriteString("\n\n")

	if horse.Fatigue >= 80 {
		b.WriteString(RenderWarning("Warning: Horse is very fatigued! Training may not be effective."))
		b.WriteString("\n")
	}

	b.WriteString(RenderButton("Confirm Training (Enter)", true))
	b.WriteString("\n\n")
	b.WriteString(RenderHelp("Enter to confirm, ESC to go back"))

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}

func (m TrainModel) renderResultView(horse *models.Horse) string {
	var b strings.Builder

	b.WriteString(RenderTitle("Training Complete"))
	b.WriteString("\n\n")

	if m.lastResult.Success {
		b.WriteString(RenderSuccess(m.lastResult.Message))
		if m.lastResult.StatGain > 0 {
			b.WriteString("\n")
			b.WriteString(RenderSuccess(fmt.Sprintf("+%d %s gained!",
				m.lastResult.StatGain, m.selectedType.String())))
		}
	} else {
		b.WriteString(RenderError(m.lastResult.Message))
	}

	b.WriteString("\n\n")

	// Updated stats
	statsCard := m.renderStatsCard(horse)
	b.WriteString(RenderCard(statsCard, false))

	b.WriteString("\n\n")
	b.WriteString(RenderHelp("Enter to continue"))

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}

func (m TrainModel) renderStatsCard(horse *models.Horse) string {
	var stats strings.Builder

	stats.WriteString(RenderStatBar("Stamina", horse.Stamina, horse.MaxStamina))
	stats.WriteString("\n")
	stats.WriteString(RenderStatBar("Speed", horse.Speed, horse.MaxSpeed))
	stats.WriteString("\n")
	stats.WriteString(RenderStatBar("Technique", horse.Technique, horse.MaxTechnique))
	stats.WriteString("\n")
	stats.WriteString(RenderStatBar("Mental", horse.Mental, horse.MaxMental))
	stats.WriteString("\n\n")
	stats.WriteString(RenderProgressBar(horse.Fatigue, 100, 20, fatigueBarStyle))
	stats.WriteString(" Fatigue\n")
	stats.WriteString(RenderProgressBar(horse.Morale, 100, 20, statBarStyle))
	stats.WriteString(" Morale")

	return stats.String()
}

func (m TrainModel) getDayStatus(day int) string {
	currentDays := m.gameState.Season.GetCurrentTrainingDays()

	for _, td := range currentDays {
		if td.Day == day {
			if td.IsCompleted {
				if td.IsRest {
					return "Rest ✓"
				}
				return fmt.Sprintf("%s Training ✓", td.TrainingType.String())
			}
		}
	}

	return "Available"
}

func (m TrainModel) canTrainToday() bool {
	currentDays := m.gameState.Season.GetCurrentTrainingDays()

	for _, td := range currentDays {
		if td.Day == m.selectedDay && td.IsCompleted {
			return false
		}
	}

	return true
}

func (m TrainModel) isWeekComplete() bool {
	currentDays := m.gameState.Season.GetCurrentTrainingDays()
	completedDays := 0

	for _, td := range currentDays {
		if td.IsCompleted {
			completedDays++
		}
	}

	return completedDays >= 6
}

func (m TrainModel) performTraining() (TrainModel, tea.Cmd) {
	horse := m.gameState.PlayerHorse
	supporters := make([]models.Supporter, 0)

	// Get owned supporters
	for _, supporter := range m.gameState.Supporters {
		if supporter.IsOwned {
			supporters = append(supporters, supporter)
		}
	}

	result := horse.Train(m.selectedType, supporters)
	m.lastResult = &result

	// Add training day to season
	trainingDay := models.TrainingDay{
		Week:         m.gameState.Season.CurrentWeek,
		Day:          m.selectedDay,
		TrainingType: m.selectedType,
		IsRest:       false,
		IsCompleted:  true,
		Result:       &result,
	}

	m.gameState.Season.AddTrainingDay(trainingDay)
	m.mode = ViewingTrainingResult

	return m, nil
}

func (m TrainModel) performRest() (TrainModel, tea.Cmd) {
	horse := m.gameState.PlayerHorse
	horse.Rest()

	// Add rest day to season
	trainingDay := models.TrainingDay{
		Week:        m.gameState.Season.CurrentWeek,
		Day:         m.selectedDay,
		IsRest:      true,
		IsCompleted: true,
	}

	m.gameState.Season.AddTrainingDay(trainingDay)

	result := &models.TrainingResult{
		Success: true,
		Message: "Horse rested well!",
	}
	m.lastResult = result
	m.mode = ViewingTrainingResult

	return m, nil
}

func (m TrainModel) getStatValues(horse *models.Horse, trainingType models.TrainingType) (int, int) {
	switch trainingType {
	case models.StaminaTraining:
		return horse.Stamina, horse.MaxStamina
	case models.SpeedTraining:
		return horse.Speed, horse.MaxSpeed
	case models.TechniqueTraining:
		return horse.Technique, horse.MaxTechnique
	case models.MentalTraining:
		return horse.Mental, horse.MaxMental
	default:
		return 0, 0
	}
}

type WeekCompleteMsg struct{}
