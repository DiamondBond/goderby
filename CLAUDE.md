# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Development Commands

- **Build the game**: `go build -o goderby cmd/main.go` or `./build.sh`
- **Run the game**: `./goderby`
- **Run tests**: `go test ./...` or `go test cmd/main_test.go` for main tests
- **Run a single test**: `go test -run TestGameInitialization cmd/main_test.go`
- **Go module management**: Standard Go commands (`go mod tidy`, `go mod download`)

## Architecture Overview

Go Derby is a terminal-based horse racing life simulation game built with Go and Bubble Tea TUI framework. The architecture follows a clean separation of concerns:

### Core Components

- **cmd/main.go**: Application entry point with Bubble Tea program setup
- **internal/models/**: Game data structures and business logic
  - `game.go`: GameState, Season, and Event management
  - `horse.go`: Horse entities with training, stats, and progression
  - `race.go`: Race definitions and simulation logic
  - `supporter.go`: Support card system for training bonuses
- **internal/ui/**: Terminal UI components built with Bubble Tea
  - Each view has its own model (MainMenu, Scout, Train, Race, Summary)
  - Shared styles and rendering utilities
- **internal/game/**: Game simulation logic (race simulator)
- **internal/data/**: Data persistence and loading (JSON-based)
- **assets/**: Game data files (horses.json, races.json, supporters.json, saves/)

### Key Patterns

1. **Bubble Tea MVC Pattern**: Each UI view implements Update/View methods
2. **Message-based Communication**: Custom message types for navigation and state changes
3. **JSON Persistence**: Game state and data stored as JSON files
4. **Modular UI**: Each game screen is a separate Bubble Tea model

### Data Flow

1. Main app initializes DataLoader and GameState
2. UI views communicate via message passing
3. Game state updates flow through the main AppModel
4. Save/load operations use JSON serialization to assets/saves/

### Game Mechanics

- **Training System**: Weekly calendar with 4 training types affecting horse stats
- **Racing**: Live simulation with formation/pace strategy
- **Season Progression**: 24-week seasons with aging and long-term progression
- **Supporter Cards**: Provide training bonuses

## Testing

The project includes unit tests in `cmd/main_test.go` covering:
- Game initialization and data loading
- Horse training mechanics
- Race creation
- Game state management

Tests use the standard Go testing framework and can be run with standard `go test` commands.

## Dependencies

- **Bubble Tea**: TUI framework for interactive terminal applications
- **Lip Gloss**: Terminal styling and layout
- **Standard Go libraries**: JSON, time, file I/O

All dependencies are managed through go.mod with Go 1.24.5+ required.