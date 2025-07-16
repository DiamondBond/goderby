package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type InfoModel struct {
	gameVersion string
}

func NewInfoModel(gameVersion string) InfoModel {
	return InfoModel{
		gameVersion: gameVersion,
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
		}
	}

	return m, nil
}

func (m InfoModel) View() string {
	var b strings.Builder

	// Title
	b.WriteString(RenderTitle("Go Derby " + m.gameVersion))
	b.WriteString("\n\n")

	// Game description
	b.WriteString(RenderHeader("About"))
	b.WriteString("\n")
	description := "🏇 Horse racing life simulation game inspired by Umamusume: Pretty Derby,\nbuilt with Go and Bubble Tea."
	b.WriteString(cardStyle.Render(description))
	b.WriteString("\n\n")

	// Features
	b.WriteString(RenderHeader("Features"))
	b.WriteString("\n")
	features := "• Horse Scouting: Choose from 28 uniquely named horses\n" +
		"• Training System: Weekly training calendar with 4 training types\n" +
		"• Racing: Live race simulation with real-time progress\n" +
		"• Season Progression: 24-week seasons with aging\n" +
		"• Supporter System: Support cards for training bonuses\n" +
		"• Save/Load: Persistent game state with JSON saves"
	b.WriteString(cardStyle.Render(features))
	b.WriteString("\n\n")

	// How to Play
	b.WriteString(RenderHeader("How to Play"))
	b.WriteString("\n")
	howToPlay := "1. Scout a Horse: Choose your racing partner\n" +
		"2. Train Weekly: Plan training to improve stats\n" +
		"3. Enter Races: Compete in races matching your level\n" +
		"4. Progress Seasons: Advance as your horse ages\n" +
		"5. Achieve Fame: Win races, gain fans, become a legend"
	b.WriteString(cardStyle.Render(howToPlay))
	b.WriteString("\n\n")

	// Training Types
	b.WriteString(RenderHeader("Training Types"))
	b.WriteString("\n")
	training := "• Stamina: Improves endurance for longer races\n" +
		"• Speed: Increases base racing speed\n" +
		"• Technique: Enhances consistency and skill\n" +
		"• Mental: Improves performance under pressure"
	b.WriteString(cardStyle.Render(training))
	b.WriteString("\n\n")

	// Controls
	b.WriteString(RenderHeader("Controls"))
	b.WriteString("\n")
	controls := "↑/↓: Navigate menus | ←/→: Navigate options\n" +
		"Enter/Space: Select/Confirm | ESC/q: Go back/Quit\n" +
		"r: Rest (training) | i: Inspect (scout) | n: Next week/season"
	b.WriteString(cardStyle.Render(controls))
	b.WriteString("\n\n")

	// Credits
	b.WriteString(RenderHeader("Credits"))
	b.WriteString("\n")
	credits := "Made by Diamond\n🎮 Enjoy racing to victory in Go Derby! 🏆"
	b.WriteString(cardStyle.Render(credits))
	b.WriteString("\n\n")

	// Help
	b.WriteString(RenderHelp("Press any key to return to main menu"))

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}
