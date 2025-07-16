package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"goderby/internal/models"
)

type SupportersModel struct {
	gameState    *models.GameState
	cursor       int
	selectedPage int // 0 = owned, 1 = all available
	viewStart    int // First visible supporter
	maxVisible   int // Maximum visible supporters
}

func NewSupportersModel(gameState *models.GameState) SupportersModel {
	return SupportersModel{
		gameState:    gameState,
		cursor:       0,
		selectedPage: 0,
		viewStart:    0,
		maxVisible:   3,
	}
}

func (m SupportersModel) Init() tea.Cmd {
	return nil
}

func (m SupportersModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			return m, func() tea.Msg {
				return NavigationMsg{State: MainMenuView}
			}
		case "tab":
			m.selectedPage = (m.selectedPage + 1) % 2
			m.cursor = 0
			m.viewStart = 0
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				// Adjust view if cursor goes above visible area
				if m.cursor < m.viewStart {
					m.viewStart = m.cursor
				}
			}
		case "down", "j":
			maxItems := m.getMaxItems()
			if m.cursor < maxItems-1 {
				m.cursor++
				// Adjust view if cursor goes below visible area
				if m.cursor >= m.viewStart+m.maxVisible {
					m.viewStart = m.cursor - m.maxVisible + 1
				}
			}
		}
	}

	return m, nil
}

func (m SupportersModel) getMaxItems() int {
	if m.selectedPage == 0 {
		// Count owned supporters
		count := 0
		for _, supporter := range m.gameState.Supporters {
			if supporter.IsOwned {
				count++
			}
		}
		return count
	}
	// All supporters
	return len(m.gameState.Supporters)
}

func (m SupportersModel) View() string {
	var b strings.Builder

	// Title
	b.WriteString(RenderTitle("Supporter Management"))
	b.WriteString("\n\n")

	// Tab navigation
	tabStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#DDA0DD")).
		Foreground(lipgloss.Color("#FFFFFF"))

	activeTabStyle := tabStyle.
		BorderForeground(lipgloss.Color("#7ED321")).
		Background(lipgloss.Color("#7ED321")).
		Foreground(lipgloss.Color("#FFFFFF"))

	var ownedTab, allTab string
	if m.selectedPage == 0 {
		ownedTab = activeTabStyle.Render("Owned")
		allTab = tabStyle.Render("All")
	} else {
		ownedTab = tabStyle.Render("Owned")
		allTab = activeTabStyle.Render("All")
	}

	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, ownedTab, "  ", allTab))
	b.WriteString("\n\n")

	// Filter supporters based on current page
	var displaySupporter []models.Supporter
	if m.selectedPage == 0 {
		// Show only owned supporters
		for _, supporter := range m.gameState.Supporters {
			if supporter.IsOwned {
				displaySupporter = append(displaySupporter, supporter)
			}
		}
		if len(displaySupporter) == 0 {
			b.WriteString(RenderInfo("No supporters owned yet."))
			b.WriteString("\n\n")
		}
	} else {
		// Show all supporters
		displaySupporter = m.gameState.Supporters
	}

	// Display supporters with scrolling
	viewEnd := min(m.viewStart+m.maxVisible, len(displaySupporter))

	for i := m.viewStart; i < viewEnd; i++ {
		supporter := displaySupporter[i]
		style := lipgloss.NewStyle().
			Padding(1, 2).
			Margin(0, 0, 1, 0).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#DDA0DD")).
			Foreground(lipgloss.Color("#FFFFFF"))

		if i == m.cursor {
			style = style.BorderForeground(lipgloss.Color("#7ED321"))
		}

		// Supporter info
		content := fmt.Sprintf("%s %s", supporter.Rarity.String(), supporter.Name)
		if supporter.IsOwned {
			content += " ✓"
		}
		content += "\n"
		content += supporter.Description + "\n"

		// Training bonuses
		if len(supporter.TrainingBonus) > 0 {
			content += "Bonuses: "
			var bonuses []string
			for trainingType, bonus := range supporter.TrainingBonus {
				if bonus > 0 {
					bonuses = append(bonuses, fmt.Sprintf("%s +%d", trainingType.String(), bonus))
				}
			}
			content += strings.Join(bonuses, ", ")
		}

		rarityColor := lipgloss.Color(supporter.Rarity.Color())
		styledContent := style.Foreground(rarityColor).Render(content)
		b.WriteString(styledContent)
		b.WriteString("\n")
	}

	// Show scroll indicators
	if len(displaySupporter) > m.maxVisible {
		scrollInfo := fmt.Sprintf("Showing %d-%d of %d supporters",
			m.viewStart+1, viewEnd, len(displaySupporter))
		if m.viewStart > 0 {
			scrollInfo += " ↑"
		}
		if viewEnd < len(displaySupporter) {
			scrollInfo += " ↓"
		}
		b.WriteString(RenderInfo(scrollInfo))
		b.WriteString("\n")
	}

	// Instructions
	b.WriteString("\n")
	b.WriteString(RenderHelp("Controls:"))
	b.WriteString("\n")
	b.WriteString(RenderHelp("↑/↓: Navigate | Tab: Switch tabs | Esc: Back to menu"))

	return b.String()
}
