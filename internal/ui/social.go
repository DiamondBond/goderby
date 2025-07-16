package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"goderby/internal/models"
)

// RenderShareableHorseProfile creates a beautiful shareable horse profile card
func RenderShareableHorseProfile(horse *models.Horse, highlights models.CareerHighlights, awards []models.Award) string {
	var b strings.Builder

	// Header with horse name and breed
	header := fmt.Sprintf("🐎 %s 🐎", horse.Name)
	b.WriteString(shareableHeaderStyle.Render(header))
	b.WriteString("\n")

	subHeader := fmt.Sprintf("「 %s • Age %d 」", horse.Breed, horse.Age)
	b.WriteString(shareableSubHeaderStyle.Render(subHeader))
	b.WriteString("\n\n")

	// Career stats section
	b.WriteString(shareableHeaderStyle.Render("🏆 CAREER RECORD 🏆"))
	b.WriteString("\n")

	// Stats in two columns
	leftStats := []string{
		fmt.Sprintf("Races: %d", highlights.TotalRaces),
		fmt.Sprintf("Wins: %d", highlights.TotalWins),
		fmt.Sprintf("Win Rate: %.1f%%", highlights.WinPercentage),
	}

	rightStats := []string{
		fmt.Sprintf("Earnings: $%d", highlights.TotalPrizeMoney),
		fmt.Sprintf("Fans: %d", highlights.TotalFanSupport),
		fmt.Sprintf("Rating: %d", highlights.HighestRating),
	}

	for i := 0; i < len(leftStats); i++ {
		leftCol := shareableLabelStyle.Render(leftStats[i])
		rightCol := shareableValueStyle.Render(rightStats[i])
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Left, leftCol, rightCol))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Current stats bars
	b.WriteString(shareableHeaderStyle.Render("📊 CURRENT STATS 📊"))
	b.WriteString("\n")

	// Compact stat bars for sharing
	b.WriteString(renderCompactStatBar("Stamina", horse.Stamina, horse.MaxStamina))
	b.WriteString("\n")
	b.WriteString(renderCompactStatBar("Speed", horse.Speed, horse.MaxSpeed))
	b.WriteString("\n")
	b.WriteString(renderCompactStatBar("Technique", horse.Technique, horse.MaxTechnique))
	b.WriteString("\n")
	b.WriteString(renderCompactStatBar("Mental", horse.Mental, horse.MaxMental))
	b.WriteString("\n\n")

	// Awards section
	if len(awards) > 0 {
		b.WriteString(shareableHeaderStyle.Render("🏅 ACHIEVEMENTS 🏅"))
		b.WriteString("\n")

		// Show top 3 awards
		maxAwards := 3
		if len(awards) < maxAwards {
			maxAwards = len(awards)
		}

		for i := 0; i < maxAwards; i++ {
			award := awards[i]
			awardLine := fmt.Sprintf("%s %s", award.Icon, award.Name)
			b.WriteString(shareableAwardStyle.Render(awardLine))
			b.WriteString("\n")
		}

		if len(awards) > 3 {
			b.WriteString(shareableFooterStyle.Render(fmt.Sprintf("+ %d more achievements", len(awards)-3)))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Footer
	footer := fmt.Sprintf("Go! Derby • %s", time.Now().Format("2006-01-02"))
	b.WriteString(shareableFooterStyle.Render(footer))

	return shareableCardStyle.Render(b.String())
}

// RenderShareableSeasonSummary creates a shareable season summary card
func RenderShareableSeasonSummary(horse *models.Horse, season models.Season, gameStats models.GameStats) string {
	var b strings.Builder

	// Header
	header := fmt.Sprintf("🏇 SEASON %d SUMMARY 🏇", season.Number)
	b.WriteString(shareableHeaderStyle.Render(header))
	b.WriteString("\n")

	subHeader := fmt.Sprintf("「 %s • Age %d 」", horse.Name, horse.Age)
	b.WriteString(shareableSubHeaderStyle.Render(subHeader))
	b.WriteString("\n\n")

	// Season performance
	b.WriteString(shareableHeaderStyle.Render("📈 SEASON PERFORMANCE 📈"))
	b.WriteString("\n")

	// Calculate season-specific stats
	seasonRaces := len(season.CompletedRaces)
	winRate := 0.0
	if horse.Races > 0 {
		winRate = float64(horse.Wins) / float64(horse.Races) * 100
	}

	perfStats := [][]string{
		{"Races This Season:", fmt.Sprintf("%d", seasonRaces)},
		{"Total Career Wins:", fmt.Sprintf("%d", horse.Wins)},
		{"Win Rate:", fmt.Sprintf("%.1f%%", winRate)},
		{"Current Rating:", fmt.Sprintf("%d", horse.GetOverallRating())},
		{"Fan Support:", fmt.Sprintf("%d", horse.FanSupport)},
		{"Career Earnings:", fmt.Sprintf("$%d", horse.Money)},
	}

	for _, stat := range perfStats {
		leftCol := shareableLabelStyle.Render(stat[0])
		rightCol := shareableValueStyle.Render(stat[1])
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Left, leftCol, rightCol))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Training focus
	b.WriteString(shareableHeaderStyle.Render("💪 TRAINING FOCUS 💪"))
	b.WriteString("\n")

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

	trainingStats := [][]string{
		{"Stamina Training:", fmt.Sprintf("%d sessions", staminaTraining)},
		{"Speed Training:", fmt.Sprintf("%d sessions", speedTraining)},
		{"Technique Training:", fmt.Sprintf("%d sessions", techniqueTraining)},
		{"Mental Training:", fmt.Sprintf("%d sessions", mentalTraining)},
	}

	for _, stat := range trainingStats {
		leftCol := shareableLabelStyle.Render(stat[0])
		rightCol := shareableValueStyle.Render(stat[1])
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Left, leftCol, rightCol))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Career milestones
	b.WriteString(shareableHeaderStyle.Render("🌟 MILESTONES 🌟"))
	b.WriteString("\n")

	milestones := []string{}
	if horse.Wins >= 10 {
		milestones = append(milestones, "✓ 10+ Race Champion")
	} else if horse.Wins >= 5 {
		milestones = append(milestones, "✓ 5+ Race Winner")
	} else if horse.Wins >= 1 {
		milestones = append(milestones, "✓ First Victory")
	}

	if horse.Money >= 100000 {
		milestones = append(milestones, "✓ $100K+ Earnings")
	} else if horse.Money >= 50000 {
		milestones = append(milestones, "✓ $50K+ Earnings")
	}

	if horse.GetOverallRating() >= 200 {
		milestones = append(milestones, "✓ Elite Rating (200+)")
	} else if horse.GetOverallRating() >= 150 {
		milestones = append(milestones, "✓ Expert Rating (150+)")
	}

	if len(milestones) == 0 {
		milestones = append(milestones, "Building toward first milestone...")
	}

	for _, milestone := range milestones {
		b.WriteString(shareableAwardStyle.Render(milestone))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Footer
	footer := fmt.Sprintf("Go! Derby • Season %d • %s", season.Number, time.Now().Format("2006-01-02"))
	b.WriteString(shareableFooterStyle.Render(footer))

	return shareableCardStyle.Render(b.String())
}

// RenderShareableRetirementCard creates a shareable retirement ceremony card
func RenderShareableRetirementCard(retired models.RetiredHorse) string {
	var b strings.Builder

	// Header
	header := fmt.Sprintf("🎉 RETIREMENT CEREMONY 🎉")
	b.WriteString(shareableHeaderStyle.Render(header))
	b.WriteString("\n")

	subHeader := fmt.Sprintf("「 %s • %s 」", retired.Horse.Name, retired.Horse.Breed)
	b.WriteString(shareableSubHeaderStyle.Render(subHeader))
	b.WriteString("\n\n")

	// Career summary
	b.WriteString(shareableHeaderStyle.Render("🏆 CAREER LEGEND 🏆"))
	b.WriteString("\n")

	careerStats := [][]string{
		{"Age at Retirement:", fmt.Sprintf("%d years", retired.Horse.Age)},
		{"Career Span:", fmt.Sprintf("%d seasons", retired.CareerHighlights.CareerLength)},
		{"Total Races:", fmt.Sprintf("%d", retired.CareerHighlights.TotalRaces)},
		{"Total Wins:", fmt.Sprintf("%d (%.1f%%)", retired.CareerHighlights.TotalWins, retired.CareerHighlights.WinPercentage)},
		{"Career Earnings:", fmt.Sprintf("$%d", retired.CareerHighlights.TotalPrizeMoney)},
		{"Peak Rating:", fmt.Sprintf("%d", retired.CareerHighlights.HighestRating)},
	}

	for _, stat := range careerStats {
		leftCol := shareableLabelStyle.Render(stat[0])
		rightCol := shareableValueStyle.Render(stat[1])
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Left, leftCol, rightCol))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Post-retirement info
	b.WriteString(shareableHeaderStyle.Render("🏠 RETIREMENT LIFE 🏠"))
	b.WriteString("\n")

	retirementInfo := [][]string{
		{"Retirement Home:", retired.RetirementHome.Name},
		{"New Role:", retired.PostRetirementRole.String()},
		{"Passive Income:", fmt.Sprintf("$%d/month", retired.PassiveIncome)},
		{"Legacy Fame:", fmt.Sprintf("%d/month", retired.PassiveFame)},
	}

	for _, info := range retirementInfo {
		leftCol := shareableLabelStyle.Render(info[0])
		rightCol := shareableValueStyle.Render(info[1])
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Left, leftCol, rightCol))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Awards
	if len(retired.Awards) > 0 {
		b.WriteString(shareableHeaderStyle.Render("🏅 HALL OF FAME 🏅"))
		b.WriteString("\n")

		// Show top 4 awards
		maxAwards := 4
		if len(retired.Awards) < maxAwards {
			maxAwards = len(retired.Awards)
		}

		for i := 0; i < maxAwards; i++ {
			award := retired.Awards[i]
			awardLine := fmt.Sprintf("%s %s", award.Icon, award.Name)
			b.WriteString(shareableAwardStyle.Render(awardLine))
			b.WriteString("\n")
		}

		if len(retired.Awards) > 4 {
			b.WriteString(shareableFooterStyle.Render(fmt.Sprintf("+ %d more awards", len(retired.Awards)-4)))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Footer
	footer := fmt.Sprintf("Go! Derby • Retired %s • A True Champion", retired.RetiredAt.Format("2006-01-02"))
	b.WriteString(shareableFooterStyle.Render(footer))

	return shareableCardStyle.Render(b.String())
}

// Helper function to render compact stat bars for sharing
func renderCompactStatBar(label string, current, max int) string {
	percentage := float64(current) / float64(max)
	barWidth := 25
	filled := int(float64(barWidth) * percentage)

	bar := ""
	for i := 0; i < barWidth; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}

	labelCol := shareableLabelStyle.Width(12).Render(label + ":")
	barCol := shareableStatStyle.Render(bar)
	valueCol := shareableValueStyle.Width(8).Align(lipgloss.Right).Render(fmt.Sprintf("%d/%d", current, max))

	return lipgloss.JoinHorizontal(lipgloss.Left, labelCol, barCol, valueCol)
}
