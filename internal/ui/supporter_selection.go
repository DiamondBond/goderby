package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"goderby/internal/models"
)

type SupporterSelectionModel struct {
	gameState       *models.GameState
	cursor          int
	ownedSupporters []models.Supporter
	confirmed       bool
}

func NewSupporterSelectionModel(gameState *models.GameState) SupporterSelectionModel {
	return SupporterSelectionModel{
		gameState:       gameState,
		cursor:          0,
		ownedSupporters: gameState.GetOwnedSupporters(),
		confirmed:       false,
	}
}

func (m SupporterSelectionModel) Init() tea.Cmd {
	return nil
}

func (m SupporterSelectionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			if m.confirmed {
				return m, func() tea.Msg {
					return NavigationMsg{State: MainMenuView}
				}
			}
			return m, func() tea.Msg {
				return NavigationMsg{State: ScoutView}
			}
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.ownedSupporters)-1 {
				m.cursor++
			}
		case "enter", " ":
			if m.confirmed {
				return m, func() tea.Msg {
					return NavigationMsg{State: MainMenuView}
				}
			}

			if len(m.ownedSupporters) > 0 {
				selectedSupporter := m.ownedSupporters[m.cursor]

				// Toggle selection
				if m.gameState.IsSupporterSelected(selectedSupporter.ID) {
					m.gameState.DeselectSupporter(selectedSupporter.ID)
				} else {
					m.gameState.SelectSupporter(selectedSupporter.ID)
				}
			}
		case "c":
			if !m.confirmed && len(m.gameState.ActiveSupporters) > 0 {
				m.confirmed = true
				return m, func() tea.Msg {
					return SupportersSelectedMsg{Supporters: m.gameState.GetActiveSupporters()}
				}
			}
		}
	}

	return m, nil
}

func (m SupporterSelectionModel) View() string {
	var b strings.Builder

	if m.confirmed {
		// Confirmation screen
		b.WriteString(RenderTitle("Supporters Selected!"))
		b.WriteString("\n\n")

		activeSupporters := m.gameState.GetActiveSupporters()
		b.WriteString(RenderSuccess(fmt.Sprintf("You've selected %d supporters for your team!", len(activeSupporters))))
		b.WriteString("\n\n")

		for _, supporter := range activeSupporters {
			supporterInfo := fmt.Sprintf("%s %s", supporter.Rarity.String(), supporter.Name)
			b.WriteString(RenderCard(supporterInfo, false))
			b.WriteString("\n")
		}

		b.WriteString("\n")
		b.WriteString(RenderInfo("Continuing to training..."))
		b.WriteString("\n")
		b.WriteString(RenderHelp("Press Enter to continue"))

		return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
	}

	// Main selection screen
	b.WriteString(RenderTitle("Select Your Supporters"))
	b.WriteString("\n\n")

	activeCount := len(m.gameState.ActiveSupporters)
	b.WriteString(RenderHeader(fmt.Sprintf("Choose up to 4 supporters (%d/4 selected)", activeCount)))
	b.WriteString("\n")
	b.WriteString(RenderInfo("These supporters will provide training bonuses throughout the season"))
	b.WriteString("\n\n")

	if len(m.ownedSupporters) == 0 {
		b.WriteString(RenderWarning("You don't own any supporters yet!"))
		b.WriteString("\n\n")
		b.WriteString(RenderHelp("Press ESC to go back"))
		return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
	}

	// Supporter list
	for i, supporter := range m.ownedSupporters {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		// Check if selected
		isSelected := m.gameState.IsSupporterSelected(supporter.ID)
		selectedIcon := ""
		if isSelected {
			selectedIcon = " ✓"
		}

		supporterInfo := fmt.Sprintf("%s %s %s%s", cursor, supporter.Rarity.String(), supporter.Name, selectedIcon)
		supporterInfo += "\n"
		supporterInfo += fmt.Sprintf("   %s", supporter.Description)

		// Show bonuses
		if len(supporter.TrainingBonus) > 0 {
			supporterInfo += "\n   Bonuses: "
			var bonuses []string
			for trainingType, bonus := range supporter.TrainingBonus {
				if bonus > 0 {
					bonuses = append(bonuses, fmt.Sprintf("%s +%d", trainingType.String(), bonus))
				}
			}
			supporterInfo += strings.Join(bonuses, ", ")
		}

		// Style based on selection and cursor
		style := lipgloss.NewStyle().
			Padding(1, 2).
			Margin(0, 0, 1, 0).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#DDA0DD")).
			Foreground(lipgloss.Color("#FFFFFF"))

		if isSelected {
			style = style.BorderForeground(lipgloss.Color("#7ED321")).
				Background(lipgloss.Color("#1A4A1A"))
		}

		if m.cursor == i {
			style = style.BorderForeground(lipgloss.Color("#7ED321"))
		}

		rarityColor := lipgloss.Color(supporter.Rarity.Color())
		styledContent := style.Foreground(rarityColor).Render(supporterInfo)
		b.WriteString(styledContent)
		b.WriteString("\n")
	}

	// Instructions
	b.WriteString("\n")
	b.WriteString(RenderHelp("Controls:"))
	b.WriteString("\n")
	b.WriteString(RenderHelp("↑/↓: Navigate | Enter/Space: Select/Deselect | C: Confirm selection | ESC: Back"))

	if activeCount > 0 {
		b.WriteString("\n")
		b.WriteString(RenderButton("Press C to confirm selection", true))
	}

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}

type SupportersSelectedMsg struct {
	Supporters []models.Supporter
}
