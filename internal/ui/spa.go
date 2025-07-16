package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"goderby/internal/models"
)

type SpaModel struct {
	gameState      *models.GameState
	selectedOption int
	spaServices    []SpaService
	mode           SpaMode
	animation      SpaAnimation
	lastResult     *SpaResult
}

type SpaMode int

const (
	SelectingService SpaMode = iota
	ViewingAnimation
	ViewingSpaResult
)

type SpaService struct {
	Name         string
	Description  string
	Cost         int
	FatigueRedux int
	MoraleBoost  int
	Icon         string
	Animation    []string
}

type SpaAnimation struct {
	frames       []string
	currentFrame int
	duration     time.Duration
	startTime    time.Time
	isPlaying    bool
}

type SpaResult struct {
	Success        bool
	Message        string
	FatigueReduced int
	MoraleGained   int
	CostPaid       int
}

func NewSpaModel(gameState *models.GameState) SpaModel {
	services := []SpaService{
		{
			Name:         "Relaxing Massage",
			Description:  "A gentle massage to ease muscle tension",
			Cost:         500,
			FatigueRedux: 20,
			MoraleBoost:  5,
			Icon:         "💆",
			Animation: []string{
				"🐎   💆‍♀️",
				"🐎 ～ 💆‍♀️",
				"🐎 ✨ 💆‍♀️",
				"🐎 😌 💆‍♀️",
			},
		},
		{
			Name:         "Hot Spring Bath",
			Description:  "Soothing mineral bath for deep relaxation",
			Cost:         800,
			FatigueRedux: 30,
			MoraleBoost:  10,
			Icon:         "♨️",
			Animation: []string{
				"🐎   ♨️",
				"🐎 💦 ♨️",
				"🐎 ～ ♨️",
				"🐎 😊 ♨️",
				"🐎 ✨ ♨️",
			},
		},
		{
			Name:         "Aromatherapy Session",
			Description:  "Calming scents to restore mental balance",
			Cost:         600,
			FatigueRedux: 15,
			MoraleBoost:  15,
			Icon:         "🌸",
			Animation: []string{
				"🐎   🌸",
				"🐎 ～ 🌸",
				"🐎 💫 🌸",
				"🐎 😌 🌸",
			},
		},
		{
			Name:         "Luxury Spa Package",
			Description:  "The ultimate wellness experience",
			Cost:         1500,
			FatigueRedux: 50,
			MoraleBoost:  20,
			Icon:         "👑",
			Animation: []string{
				"🐎     👑",
				"🐎 ✨  👑",
				"🐎 💆‍♀️ 👑",
				"🐎 ♨️  👑",
				"🐎 🌸  👑",
				"🐎 😍  👑",
				"🐎 💖  👑",
			},
		},
		{
			Name:         "Quick Grooming",
			Description:  "Basic grooming and cleanup",
			Cost:         200,
			FatigueRedux: 10,
			MoraleBoost:  5,
			Icon:         "🧽",
			Animation: []string{
				"🐎   🧽",
				"🐎 ～ 🧽",
				"🐎 ✨ 🧽",
			},
		},
	}

	return SpaModel{
		gameState:      gameState,
		selectedOption: 0,
		spaServices:    services,
		mode:           SelectingService,
		animation:      SpaAnimation{},
	}
}

func (m SpaModel) Init() tea.Cmd {
	return nil
}

func (m SpaModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.mode {
		case SelectingService:
			switch msg.String() {
			case "ctrl+c", "q", "esc":
				return m, func() tea.Msg {
					return NavigationMsg{State: MainMenuView}
				}
			case "up", "k":
				if m.selectedOption > 0 {
					m.selectedOption--
				}
			case "down", "j":
				if m.selectedOption < len(m.spaServices)-1 {
					m.selectedOption++
				}
			case "enter", " ":
				return m.purchaseSpaService()
			}
		case ViewingAnimation:
			// Animation plays automatically, just wait for completion
		case ViewingSpaResult:
			switch msg.String() {
			case "enter", " ", "esc":
				m.mode = SelectingService
				m.lastResult = nil
			}
		}
	case AnimationTickMsg:
		if m.mode == ViewingAnimation && m.animation.isPlaying {
			m.animation.currentFrame++
			if m.animation.currentFrame >= len(m.animation.frames) {
				// Animation complete, show result
				m.mode = ViewingSpaResult
				m.animation.isPlaying = false
				return m, nil
			}
			return m, tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
				return AnimationTickMsg{}
			})
		}
	}

	return m, nil
}

func (m SpaModel) purchaseSpaService() (SpaModel, tea.Cmd) {
	service := m.spaServices[m.selectedOption]
	horse := m.gameState.PlayerHorse

	// Check if player has enough money
	if horse.Money < service.Cost {
		m.lastResult = &SpaResult{
			Success: false,
			Message: "Not enough money for this service!",
		}
		m.mode = ViewingSpaResult
		return m, nil
	}

	// Check if horse needs this service
	if horse.Fatigue == 0 && service.FatigueRedux > 0 {
		m.lastResult = &SpaResult{
			Success: false,
			Message: "Your horse is already well-rested!",
		}
		m.mode = ViewingSpaResult
		return m, nil
	}

	// Apply spa treatment
	horse.Money -= service.Cost
	oldFatigue := horse.Fatigue
	oldMorale := horse.Morale

	horse.Fatigue = max(horse.Fatigue-service.FatigueRedux, 0)
	horse.Morale = min(horse.Morale+service.MoraleBoost, 100)

	fatigueReduced := oldFatigue - horse.Fatigue
	moraleGained := horse.Morale - oldMorale

	m.lastResult = &SpaResult{
		Success:        true,
		Message:        fmt.Sprintf("Your horse enjoyed the %s!", service.Name),
		FatigueReduced: fatigueReduced,
		MoraleGained:   moraleGained,
		CostPaid:       service.Cost,
	}

	// Start animation
	m.mode = ViewingAnimation
	m.animation = SpaAnimation{
		frames:       service.Animation,
		currentFrame: 0,
		isPlaying:    true,
		startTime:    time.Now(),
	}

	return m, tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
		return AnimationTickMsg{}
	})
}

func (m SpaModel) View() string {
	var b strings.Builder

	// Title
	b.WriteString(RenderTitle("🧖‍♀️ Horse Wellness Spa 🧖‍♀️"))
	b.WriteString("\n\n")

	// Horse info
	if m.gameState.PlayerHorse != nil {
		horse := m.gameState.PlayerHorse
		b.WriteString(RenderHeader("Current Condition"))
		b.WriteString("\n")

		horseInfo := fmt.Sprintf("🐎 %s\n", horse.Name)
		horseInfo += fmt.Sprintf("💰 Money: $%d\n", horse.Money)
		horseInfo += fmt.Sprintf("😴 Fatigue: %d/100\n", horse.Fatigue)
		horseInfo += fmt.Sprintf("😊 Morale: %d/100\n", horse.Morale)

		// Add fatigue bar
		fatigueBar := RenderProgressBar(horse.Fatigue, 100, 20, fatigueBarStyle)
		horseInfo += fmt.Sprintf("Fatigue: %s\n", fatigueBar)

		// Add morale bar
		moraleBar := RenderProgressBar(horse.Morale, 100, 20, statBarStyle)
		horseInfo += fmt.Sprintf("Morale:  %s", moraleBar)

		b.WriteString(cardStyle.Render(horseInfo))
		b.WriteString("\n\n")
	}

	switch m.mode {
	case SelectingService:
		b.WriteString(RenderHeader("Available Services"))
		b.WriteString("\n")

		for i, service := range m.spaServices {
			cursor := " "
			if m.selectedOption == i {
				cursor = ">"
			}

			serviceInfo := fmt.Sprintf("%s %s %s", cursor, service.Icon, service.Name)
			serviceInfo += fmt.Sprintf("\n   %s", service.Description)
			serviceInfo += fmt.Sprintf("\n   💰 Cost: $%d", service.Cost)
			serviceInfo += fmt.Sprintf("\n   😴 Fatigue -%d | 😊 Morale +%d", service.FatigueRedux, service.MoraleBoost)

			if m.selectedOption == i {
				b.WriteString(selectedCardStyle.Render(serviceInfo))
			} else {
				// Check affordability
				if m.gameState.PlayerHorse.Money < service.Cost {
					dimmedStyle := lipgloss.NewStyle().
						Border(lipgloss.RoundedBorder()).
						BorderForeground(lipgloss.Color("#555555")).
						Padding(1, 2).
						Margin(0, 1).
						Foreground(lipgloss.Color("#888888"))
					b.WriteString(dimmedStyle.Render(serviceInfo))
				} else {
					b.WriteString(cardStyle.Render(serviceInfo))
				}
			}
			b.WriteString("\n")
		}

		b.WriteString("\n")
		b.WriteString(RenderHelp("Use ↑/↓ to navigate, Enter to purchase, q/esc to return"))

	case ViewingAnimation:
		b.WriteString(RenderHeader("Treatment in Progress"))
		b.WriteString("\n")

		service := m.spaServices[m.selectedOption]
		animationDisplay := fmt.Sprintf("✨ %s ✨\n\n", service.Name)

		if m.animation.currentFrame < len(m.animation.frames) {
			animationDisplay += fmt.Sprintf("      %s\n\n", m.animation.frames[m.animation.currentFrame])
		}

		animationDisplay += "Please wait while your horse enjoys the treatment..."

		b.WriteString(cardStyle.Render(animationDisplay))

	case ViewingSpaResult:
		b.WriteString(RenderHeader("Treatment Complete"))
		b.WriteString("\n")

		if m.lastResult != nil {
			if m.lastResult.Success {
				resultInfo := fmt.Sprintf("✅ %s\n\n", m.lastResult.Message)
				resultInfo += fmt.Sprintf("💰 Cost: $%d\n", m.lastResult.CostPaid)
				if m.lastResult.FatigueReduced > 0 {
					resultInfo += fmt.Sprintf("😴 Fatigue reduced by %d\n", m.lastResult.FatigueReduced)
				}
				if m.lastResult.MoraleGained > 0 {
					resultInfo += fmt.Sprintf("😊 Morale increased by %d\n", m.lastResult.MoraleGained)
				}
				b.WriteString(cardStyle.Render(resultInfo))
			} else {
				resultInfo := fmt.Sprintf("❌ %s", m.lastResult.Message)
				b.WriteString(cardStyle.Render(resultInfo))
			}
		}

		b.WriteString("\n")
		b.WriteString(RenderHelp("Press Enter or Esc to continue"))
	}

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}

type AnimationTickMsg struct{}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
