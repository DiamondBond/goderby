package ui

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Color palette
	primaryColor   = lipgloss.Color("#FF6B6B")
	secondaryColor = lipgloss.Color("#A8E6CF")
	accentColor    = lipgloss.Color("#45B7D1")
	successColor   = lipgloss.Color("#96CEB4")
	warningColor   = lipgloss.Color("#FFEAA7")
	errorColor     = lipgloss.Color("#FD79A8")
	textColor      = lipgloss.Color("#FFFFFF")
	subtleColor    = lipgloss.Color("#A0A0A0")

	// Base styles
	titleStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor)

	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true).
			Padding(0, 1)

	menuItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 2)

	selectedMenuItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(accentColor).
				Bold(true).
				Padding(0, 2)

	cardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(subtleColor).
			Padding(1, 2).
			Margin(0, 1).
			Foreground(lipgloss.Color("#FFFFFF"))

	selectedCardStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(accentColor).
				Padding(1, 2).
				Margin(0, 1).
				Background(lipgloss.Color("#2D3436")).
				Foreground(lipgloss.Color("#FFFFFF"))

	statBarStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Background(lipgloss.Color("#DDD"))

	fatigueBarStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Background(lipgloss.Color("#DDD"))

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CCCCCC")).
			Italic(true)

	successStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	warningStyle = lipgloss.NewStyle().
			Foreground(warningColor).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CCCCCC")).
			Padding(1, 0)

	buttonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(accentColor).
			Padding(0, 2).
			Margin(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(accentColor)

	disabledButtonStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#888888")).
				Background(lipgloss.Color("#444444")).
				Padding(0, 2).
				Margin(0, 1).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#888888"))
)

func RenderProgressBar(current, max int, width int, style lipgloss.Style) string {
	if max == 0 {
		return style.Render(lipgloss.PlaceHorizontal(width, lipgloss.Left, ""))
	}

	percentage := float64(current) / float64(max)
	filled := int(float64(width) * percentage)

	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}

	return style.Render(bar)
}

func RenderStatBar(label string, current, max int) string {
	bar := RenderProgressBar(current, max, 20, statBarStyle)
	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		lipgloss.NewStyle().Width(12).Render(label+":"),
		bar,
		lipgloss.NewStyle().Width(10).Align(lipgloss.Right).Render(
			lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Render(fmt.Sprintf(" %d/%d", current, max)),
		),
	)
}

func RenderTitle(text string) string {
	return titleStyle.Render("🏇 " + text + " 🏇")
}

func RenderHeader(text string) string {
	return headerStyle.Render(text)
}

func RenderCard(content string, selected bool) string {
	if selected {
		return selectedCardStyle.Render(content)
	}
	return cardStyle.Render(content)
}

func RenderButton(text string, enabled bool) string {
	if enabled {
		return buttonStyle.Render(text)
	}
	return disabledButtonStyle.Render(text)
}

func RenderSuccess(text string) string {
	return successStyle.Render("✓ " + text)
}

func RenderError(text string) string {
	return errorStyle.Render("✗ " + text)
}

func RenderWarning(text string) string {
	return warningStyle.Render("⚠ " + text)
}

func RenderInfo(text string) string {
	return infoStyle.Render("ℹ " + text)
}

func RenderHelp(text string) string {
	return helpStyle.Render(text)
}
