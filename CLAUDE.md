# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands

**Build the game:**

```bash
go build -o goderby cmd/main.go
```

**Or use the build script:**

```bash
cd utils && ./build.sh
```

**Format code:**

```bash
cd utils && ./format.sh
# Or run directly: go fmt ./...
```

**Run tests:**

```bash
go test ./cmd -v
```

**Run the game:**

```bash
./goderby
```

## Architecture Overview

Go! Derby is a terminal-based horse racing simulation game built with Go and Bubble Tea TUI framework. The architecture follows clean separation of concerns:

### Core Structure

- **cmd/**: Application entry point and main game loop
- **internal/models/**: Core game data structures and business logic
- **internal/ui/**: Bubble Tea TUI components and views
- **internal/game/**: Race simulation engine and game mechanics
- **internal/data/**: JSON data loading and persistence

### Key Models

- `GameState`: Central game state with player horse, supporters, seasons
- `Horse`: Player's horse with stats (Stamina, Speed, Technique, Mental), age, and progression
- `Season`: 24-week training cycles with weekly progression
- `Race`: Race definitions with distance, grade, and prize money
- `Supporter`: Support cards providing training bonuses

### UI Architecture

The game uses Bubble Tea's Model-View-Update pattern with separate models for each screen:

- `MainMenuModel`: Central navigation hub
- `ScoutModel`: Horse selection interface
- `TrainModel`: Weekly training calendar
- `RaceModel`: Race simulation with live progress bars
- `SummaryModel`: Season statistics and progression

### Game Flow

1. **Scout Phase**: Select a horse from available options
2. **Training Phase**: 24-week seasons with daily training choices
3. **Racing Phase**: Enter races with strategic formation/pace choices
4. **Progression**: Horse ages and advances through race grades

### Data Management

- Game state persisted as JSON in `assets/saves/game.json`
- Static data (horses, supporters, races) loaded from `assets/*.json`
- Auto-generated assets directory structure on first run

### Styling

- Custom purple/pink theme using Lip Gloss
- Green selection highlights
- Unicode icons and progress bars
- Responsive terminal layout (works in 80x24 terminals)

## Development Notes

- Go 1.24.5+ required
- Uses Bubble Tea for TUI framework
- Charmbracelet Lip Gloss for styling
- JSON for all data persistence
- MVC pattern with message-based communication
- Single test file in `cmd/main_test.go` covering core functionality
- Build scripts handle cross-platform compilation
