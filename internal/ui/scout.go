package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"goderby/internal/models"
)

type ScoutModel struct {
	horses     []models.Horse
	cursor     int
	selected   bool
	gameState  *models.GameState
	inspecting bool
	confirmed  bool
	viewStart  int // For scrolling
	maxVisible int // Maximum horses visible at once
}

func NewScoutModel(gameState *models.GameState, horses []models.Horse) ScoutModel {
	return ScoutModel{
		horses:     horses,
		cursor:     0,
		selected:   false,
		gameState:  gameState,
		inspecting: false,
		confirmed:  false,
		viewStart:  0,
		maxVisible: 3, // Show 3 horses at a time
	}
}

func (m ScoutModel) Init() tea.Cmd {
	return nil
}

func (m ScoutModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.inspecting {
				m.inspecting = false
				return m, nil
			}
			return m, func() tea.Msg {
				return NavigationMsg{State: MainMenuView}
			}
		case "esc":
			if m.inspecting {
				m.inspecting = false
				return m, nil
			}
			return m, func() tea.Msg {
				return NavigationMsg{State: MainMenuView}
			}
		case "up", "k":
			if !m.inspecting && m.cursor > 0 {
				m.cursor--
				// Adjust view if cursor goes above visible area
				if m.cursor < m.viewStart {
					m.viewStart = m.cursor
				}
			}
		case "down", "j":
			if !m.inspecting && m.cursor < len(m.horses)-1 {
				m.cursor++
				// Adjust view if cursor goes below visible area
				if m.cursor >= m.viewStart+m.maxVisible {
					m.viewStart = m.cursor - m.maxVisible + 1
				}
			}
		case "enter", " ":
			if m.inspecting && !m.confirmed {
				m.confirmed = true
				selectedHorse := m.horses[m.cursor]
				m.gameState.PlayerHorse = &selectedHorse
				return m, func() tea.Msg {
					return HorseSelectedMsg{Horse: selectedHorse}
				}
			} else if !m.inspecting {
				m.inspecting = true
			}
		case "i":
			if !m.inspecting {
				m.inspecting = true
			}
		}
	}

	return m, nil
}

func (m ScoutModel) View() string {
	var b strings.Builder

	if m.confirmed {
		// Confirmation screen
		b.WriteString(RenderTitle("Horse Selected!"))
		b.WriteString("\n\n")

		horse := m.horses[m.cursor]
		b.WriteString(RenderSuccess(fmt.Sprintf("You've selected %s!", horse.Name)))
		b.WriteString("\n\n")

		horseInfo := m.renderHorseDetails(horse)
		b.WriteString(RenderCard(horseInfo, true))
		b.WriteString("\n\n")

		b.WriteString(RenderInfo("Returning to main menu..."))

		return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
	}

	if m.inspecting {
		// Inspection screen
		b.WriteString(RenderTitle("Horse Details"))
		b.WriteString("\n\n")

		horse := m.horses[m.cursor]
		horseInfo := m.renderHorseDetails(horse)
		b.WriteString(RenderCard(horseInfo, true))
		b.WriteString("\n\n")

		if m.gameState.PlayerHorse == nil {
			b.WriteString(RenderButton("Select This Horse (Enter)", true))
			b.WriteString("  ")
		}
		b.WriteString(RenderButton("Back (ESC)", true))
		b.WriteString("\n\n")

		if m.gameState.PlayerHorse == nil {
			b.WriteString(RenderHelp("Enter to select horse, ESC to go back"))
		} else {
			b.WriteString(RenderHelp("You already have a horse selected"))
		}

		return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
	}

	// Main scout screen
	b.WriteString(RenderTitle("Scout Horses"))
	b.WriteString("\n\n")

	if m.gameState.PlayerHorse != nil {
		b.WriteString(RenderWarning("You already have a horse selected"))
		b.WriteString("\n\n")
	} else {
		b.WriteString(RenderHeader("Available Horses"))
		b.WriteString("\n")
		b.WriteString(RenderInfo("Select a horse to begin your racing career!"))
		b.WriteString("\n\n")
	}

	// Horse list
	viewEnd := m.viewStart + m.maxVisible
	if viewEnd > len(m.horses) {
		viewEnd = len(m.horses)
	}

	for i := m.viewStart; i < viewEnd; i++ {
		horse := m.horses[i]
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		horsePreview := fmt.Sprintf("%s 🐎 %s (%s)", cursor, horse.Name, horse.Breed)
		horsePreview += fmt.Sprintf("\n   Rating: %d | Stamina: %d | Speed: %d | Technique: %d | Mental: %d",
			horse.GetOverallRating(), horse.Stamina, horse.Speed, horse.Technique, horse.Mental)

		if m.cursor == i {
			b.WriteString(RenderCard(horsePreview, true))
		} else {
			b.WriteString(RenderCard(horsePreview, false))
		}
		if i < viewEnd-1 {
			b.WriteString("\n")
		}
	}

	// Show scroll indicators
	if len(m.horses) > m.maxVisible {
		b.WriteString("\n\n")
		scrollInfo := fmt.Sprintf("Showing %d-%d of %d horses",
			m.viewStart+1, viewEnd, len(m.horses))
		if m.viewStart > 0 {
			scrollInfo += " ↑"
		}
		if viewEnd < len(m.horses) {
			scrollInfo += " ↓"
		}
		b.WriteString(RenderInfo(scrollInfo))
	}

	// Help
	b.WriteString("\n\n")
	if m.gameState.PlayerHorse == nil {
		b.WriteString(RenderHelp("Use ↑/↓ to navigate, Enter/i to inspect, ESC/q to go back"))
	} else {
		b.WriteString(RenderHelp("Use ↑/↓ to navigate, Enter/i to inspect, ESC/q to go back"))
	}

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}

func (m ScoutModel) renderHorseDetails(horse models.Horse) string {
	var details strings.Builder

	details.WriteString(fmt.Sprintf("🐎 %s (%s, Age %d)\n", horse.Name, horse.Breed, horse.Age))
	details.WriteString(fmt.Sprintf("Overall Rating: %d\n\n", horse.GetOverallRating()))

	// Stats in compact format
	details.WriteString("Stats:\n")
	details.WriteString(fmt.Sprintf("  Stamina: %d/%d | Speed: %d/%d\n",
		horse.Stamina, horse.MaxStamina, horse.Speed, horse.MaxSpeed))
	details.WriteString(fmt.Sprintf("  Technique: %d/%d | Mental: %d/%d\n\n",
		horse.Technique, horse.MaxTechnique, horse.Mental, horse.MaxMental))

	// Status
	details.WriteString(fmt.Sprintf("Status: Fatigue %d/100 | Morale %d/100\n",
		horse.Fatigue, horse.Morale))

	// Record
	details.WriteString(fmt.Sprintf("Record: %d wins | %d races | %d fans | $%d",
		horse.Wins, horse.Races, horse.FanSupport, horse.Money))

	return details.String()
}

type HorseSelectedMsg struct {
	Horse models.Horse
}
