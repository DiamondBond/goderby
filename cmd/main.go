package main

import (
	"fmt"
	"log"
	"os"

	"goderby/internal/data"
	"goderby/internal/models"
	"goderby/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

const GameVersion = "v1.2"

type AppModel struct {
	currentView ui.ViewState
	gameState   *models.GameState
	dataLoader  *data.DataLoader

	// View models
	mainMenu           ui.MainMenuModel
	scout              ui.ScoutModel
	supporterSelection ui.SupporterSelectionModel
	train              ui.TrainModel
	race               ui.RaceModel
	supporters         ui.SupportersModel
	summary            ui.SummaryModel
	info               ui.InfoModel

	// Data
	availableHorses     []models.Horse
	availableRaces      []models.Race
	availableSupporters []models.Supporter

	// State
	initialized bool
	quitting    bool
}

func NewAppModel() *AppModel {
	dataLoader := data.NewDataLoader("")
	gameState := models.NewGameState()

	return &AppModel{
		currentView: ui.MainMenuView,
		gameState:   gameState,
		dataLoader:  dataLoader,
		initialized: false,
		quitting:    false,
	}
}

func (m *AppModel) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg { return InitDataMsg{} },
	)
}

func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m.handleQuit()
		}

	case InitDataMsg:
		return m.initializeData()

	case ui.NavigationMsg:
		return m.handleNavigation(msg)

	case ui.MenuSelectionMsg:
		return m.handleMenuSelection(msg)

	case ui.HorseSelectedMsg:
		m.gameState.PlayerHorse = &msg.Horse
		m.mainMenu = ui.NewMainMenuModel(m.gameState, GameVersion)
		m.currentView = ui.MainMenuView
		return m, nil

	case ui.SupportersSelectedMsg:
		m.mainMenu = ui.NewMainMenuModel(m.gameState, GameVersion)
		m.currentView = ui.MainMenuView
		return m, nil

	case ui.WeekCompleteMsg:
		m.gameState.Season.NextWeek()
		m.train = ui.NewTrainModel(m.gameState)
		return m, nil

	case tea.QuitMsg:
		return m.handleQuit()
	}

	// Update current view
	switch m.currentView {
	case ui.MainMenuView:
		var model tea.Model
		model, cmd = m.mainMenu.Update(msg)
		m.mainMenu = model.(ui.MainMenuModel)
	case ui.ScoutView:
		var model tea.Model
		model, cmd = m.scout.Update(msg)
		m.scout = model.(ui.ScoutModel)
	case ui.SupporterSelectionView:
		var model tea.Model
		model, cmd = m.supporterSelection.Update(msg)
		m.supporterSelection = model.(ui.SupporterSelectionModel)
	case ui.TrainView:
		var model tea.Model
		model, cmd = m.train.Update(msg)
		m.train = model.(ui.TrainModel)
	case ui.RaceView:
		var model tea.Model
		model, cmd = m.race.Update(msg)
		m.race = model.(ui.RaceModel)
	case ui.SupportersView:
		var model tea.Model
		model, cmd = m.supporters.Update(msg)
		m.supporters = model.(ui.SupportersModel)
	case ui.SummaryView:
		var model tea.Model
		model, cmd = m.summary.Update(msg)
		m.summary = model.(ui.SummaryModel)
	case ui.InfoView:
		var model tea.Model
		model, cmd = m.info.Update(msg)
		m.info = model.(ui.InfoModel)
	}

	return m, cmd
}

func (m *AppModel) View() string {
	if !m.initialized {
		return ui.RenderTitle("Loading Go Derby "+GameVersion+"...") + "\n\n" + ui.RenderInfo("Loading game data...")
	}

	if m.quitting {
		return ui.RenderTitle("Thanks for playing Go Derby "+GameVersion+"!") + "\n\n" + ui.RenderInfo("Game saved successfully. See you next time!")
	}

	switch m.currentView {
	case ui.MainMenuView:
		return m.mainMenu.View()
	case ui.ScoutView:
		return m.scout.View()
	case ui.SupporterSelectionView:
		return m.supporterSelection.View()
	case ui.TrainView:
		return m.train.View()
	case ui.RaceView:
		return m.race.View()
	case ui.SupportersView:
		return m.supporters.View()
	case ui.SummaryView:
		return m.summary.View()
	case ui.InfoView:
		return m.info.View()
	default:
		return m.mainMenu.View()
	}
}

func (m *AppModel) initializeData() (*AppModel, tea.Cmd) {
	// Load or create game state
	gameState, err := m.dataLoader.LoadGameState()
	if err != nil {
		log.Printf("Failed to load game state: %v", err)
		gameState = models.NewGameState()
	}
	m.gameState = gameState

	// Load horses
	horses, err := m.dataLoader.LoadHorses()
	if err != nil {
		log.Printf("Failed to load horses: %v", err)
		horses = []models.Horse{}
	}
	m.availableHorses = horses
	m.gameState.AvailableHorses = horses

	// Load supporters
	supporters, err := m.dataLoader.LoadSupporters()
	if err != nil {
		log.Printf("Failed to load supporters: %v", err)
		supporters = []models.Supporter{}
	}
	m.availableSupporters = supporters
	m.gameState.Supporters = supporters

	// Load races
	races, err := m.dataLoader.LoadRaces()
	if err != nil {
		log.Printf("Failed to load races: %v", err)
		races = []models.Race{}
	}
	m.availableRaces = races
	m.gameState.AvailableRaces = races

	// Initialize view models
	m.mainMenu = ui.NewMainMenuModel(m.gameState, GameVersion)
	m.scout = ui.NewScoutModel(m.gameState, m.availableHorses)
	m.supporterSelection = ui.NewSupporterSelectionModel(m.gameState)
	m.train = ui.NewTrainModel(m.gameState)
	m.race = ui.NewRaceModel(m.gameState, m.availableRaces)
	m.summary = ui.NewSummaryModel(m.gameState)
	m.info = ui.NewInfoModel(GameVersion)

	m.initialized = true

	return m, nil
}

func (m *AppModel) handleNavigation(msg ui.NavigationMsg) (*AppModel, tea.Cmd) {
	m.currentView = msg.State

	// Refresh models when switching views
	switch msg.State {
	case ui.MainMenuView:
		m.mainMenu = ui.NewMainMenuModel(m.gameState, GameVersion)
	case ui.ScoutView:
		m.scout = ui.NewScoutModel(m.gameState, m.availableHorses)
	case ui.SupporterSelectionView:
		m.supporterSelection = ui.NewSupporterSelectionModel(m.gameState)
	case ui.TrainView:
		m.train = ui.NewTrainModel(m.gameState)
	case ui.RaceView:
		m.race = ui.NewRaceModel(m.gameState, m.availableRaces)
	case ui.SupportersView:
		m.supporters = ui.NewSupportersModel(m.gameState)
	case ui.SummaryView:
		m.summary = ui.NewSummaryModel(m.gameState)
	case ui.InfoView:
		m.info = ui.NewInfoModel(GameVersion)
	}

	return m, nil
}

func (m *AppModel) handleMenuSelection(msg ui.MenuSelectionMsg) (*AppModel, tea.Cmd) {
	switch msg.Choice {
	case "Scout Horse":
		m.currentView = ui.ScoutView
		m.scout = ui.NewScoutModel(m.gameState, m.availableHorses)
	case "Train":
		m.currentView = ui.TrainView
		m.train = ui.NewTrainModel(m.gameState)
	case "Race":
		m.currentView = ui.RaceView
		m.race = ui.NewRaceModel(m.gameState, m.availableRaces)
	case "Supporters":
		m.currentView = ui.SupportersView
		m.supporters = ui.NewSupportersModel(m.gameState)
	case "Season Summary":
		m.currentView = ui.SummaryView
		m.summary = ui.NewSummaryModel(m.gameState)
	case "Save & Quit":
		return m.handleQuit()
	}

	return m, nil
}

func (m *AppModel) handleQuit() (*AppModel, tea.Cmd) {
	// Save game state
	if err := m.dataLoader.SaveGameState(m.gameState); err != nil {
		log.Printf("Failed to save game state: %v", err)
	}

	m.quitting = true
	return m, tea.Quit
}

type InitDataMsg struct{}

func main() {
	app := NewAppModel()
	p := tea.NewProgram(app, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}
