package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type InfoModel struct {
	gameVersion string
	viewStart   int // For scrolling
	maxVisible  int // Maximum sections visible at once
}

func NewInfoModel(gameVersion string) InfoModel {
	return InfoModel{
		gameVersion: gameVersion,
		viewStart:   0,
		maxVisible:  3, // Show 3 sections at a time
	}
}

func (m InfoModel) Init() tea.Cmd {
	return nil
}

func (m InfoModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc", "enter", " ":
			return m, func() tea.Msg {
				return NavigationMsg{State: MainMenuView}
			}
		case "up", "k":
			if m.viewStart > 0 {
				m.viewStart--
			}
		case "down", "j":
			// Total sections: About, Features, How to Play, Training Types, Controls = 5
			totalSections := 5
			if m.viewStart < totalSections-m.maxVisible {
				m.viewStart++
			}
		}
	}

	return m, nil
}

func (m InfoModel) View() string {
	var b strings.Builder

	// Title
	b.WriteString(RenderTitle("Go! Derby " + m.gameVersion))
	b.WriteString("\n\n")

	// Define sections
	sections := []struct {
		title   string
		content string
	}{
		{
			"About",
			"🏇 Horse racing life simulation game inspired by Umamusume: Pretty Derby,\nbuilt with Go and Bubble Tea.",
		},
		{
			"Features",
			"• Horse Scouting: Choose from 28 uniquely named horses\n" +
				"• Training System: Weekly training calendar with 4 training types\n" +
				"• Racing: Live race simulation with real-time progress\n" +
				"• Season Progression: 24-week seasons with aging\n" +
				"• Supporter System: Support cards for training bonuses\n" +
				"• Save/Load: Persistent game state with JSON saves",
		},
		{
			"How to Play",
			"1. Scout a Horse: Choose your racing partner\n" +
				"2. Train Weekly: Plan training to improve stats\n" +
				"3. Enter Races: Compete in races matching your level\n" +
				"4. Progress Seasons: Advance as your horse ages\n" +
				"5. Achieve Fame: Win races, gain fans, become a legend",
		},
		{
			"Training Types",
			"• Stamina: Improves endurance for longer races\n" +
				"• Speed: Increases base racing speed\n" +
				"• Technique: Enhances consistency and skill\n" +
				"• Mental: Improves performance under pressure",
		},
		{
			"Controls",
			"↑/↓: Navigate menus | ←/→: Navigate options\n" +
				"Enter/Space: Select/Confirm | ESC/q: Go back/Quit\n" +
				"r: Rest (training) | i: Inspect (scout) | n: Next week/season",
		},
	}

	// Display sections with scrolling
	viewEnd := m.viewStart + m.maxVisible
	if viewEnd > len(sections) {
		viewEnd = len(sections)
	}

	for i := m.viewStart; i < viewEnd; i++ {
		section := sections[i]
		b.WriteString(RenderHeader(section.title))
		b.WriteString("\n")
		b.WriteString(cardStyle.Render(section.content))
		if i < viewEnd-1 {
			b.WriteString("\n\n")
		}
	}

	// Show scroll indicators
	if len(sections) > m.maxVisible {
		b.WriteString("\n\n")
		scrollInfo := fmt.Sprintf("Showing %d-%d of %d sections",
			m.viewStart+1, viewEnd, len(sections))
		if m.viewStart > 0 {
			scrollInfo += " ↑"
		}
		if viewEnd < len(sections) {
			scrollInfo += " ↓"
		}
		b.WriteString(RenderInfo(scrollInfo))
	}

	// Help
	b.WriteString("\n\n")
	b.WriteString(RenderHelp("Use ↑/↓ to scroll, any other key to return to main menu"))

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}
