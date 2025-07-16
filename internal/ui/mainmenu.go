package ui

import (
	"fmt"
	"strings"

	"goderby/internal/models"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type MainMenuModel struct {
	choices     []string
	cursor      int
	selected    bool
	gameState   *models.GameState
	gameVersion string
}

func NewMainMenuModel(gameState *models.GameState, gameVersion string) MainMenuModel {
	choices := []string{"Scout Horse", "Train", "Race", "Supporters", "Season Summary", "Save & Quit"}
	if gameState.PlayerHorse == nil {
		choices = []string{"Scout Horse", "Supporters", "Save & Quit"}
	}

	return MainMenuModel{
		choices:     choices,
		cursor:      0,
		selected:    false,
		gameState:   gameState,
		gameVersion: gameVersion,
	}
}

func (m MainMenuModel) Init() tea.Cmd {
	return nil
}

func (m MainMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter", " ":
			m.selected = true
			return m, func() tea.Msg {
				return MenuSelectionMsg{Choice: m.choices[m.cursor]}
			}
		case "i":
			return m, func() tea.Msg {
				return NavigationMsg{State: InfoView}
			}
		}
	}

	return m, nil
}

func (m MainMenuModel) View() string {
	var b strings.Builder

	// Title
	b.WriteString(RenderTitle("Derby Go! " + m.gameVersion))
	b.WriteString("\n\n")

	// Player horse info if available
	if m.gameState.PlayerHorse != nil {
		horse := m.gameState.PlayerHorse
		b.WriteString(RenderHeader("Current Horse"))
		b.WriteString("\n")

		horseInfo := fmt.Sprintf("🐎 %s (%s)\n", horse.Name, horse.Breed)
		horseInfo += fmt.Sprintf("Age: %d | Rating: %d | Fans: %d\n",
			horse.Age, horse.GetOverallRating(), horse.FanSupport)
		horseInfo += fmt.Sprintf("Money: $%d | Wins: %d/%d\n",
			horse.Money, horse.Wins, horse.Races)

		b.WriteString(cardStyle.Render(horseInfo))
		b.WriteString("\n\n")

		// Season info
		season := m.gameState.Season
		b.WriteString(RenderHeader("Season Progress"))
		b.WriteString("\n")
		seasonInfo := fmt.Sprintf("Season %d - Week %d/%d",
			season.Number, season.CurrentWeek, season.MaxWeeks)
		b.WriteString(cardStyle.Render(seasonInfo))
		b.WriteString("\n\n")
	}

	// Menu
	b.WriteString(RenderHeader("Main Menu"))
	b.WriteString("\n")

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		if m.cursor == i {
			b.WriteString(selectedMenuItemStyle.Render(fmt.Sprintf("%s %s", cursor, choice)))
		} else {
			b.WriteString(menuItemStyle.Render(fmt.Sprintf("%s %s", cursor, choice)))
		}
		b.WriteString("\n")
	}

	// Help
	b.WriteString("\n")
	b.WriteString(RenderHelp("Use ↑/↓ to navigate, Enter to select, q to quit, i for info"))

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}

type MenuSelectionMsg struct {
	Choice string
}

// Navigation state constants
type ViewState int

const (
	MainMenuView ViewState = iota
	ScoutView
	SupporterSelectionView
	TrainView
	RaceView
	SupportersView
	SummaryView
	InfoView
)

type NavigationMsg struct {
	State ViewState
	Data  interface{}
}
