# Go Derby

🏇 A terminal-based horse racing life simulation game inspired by Umamusume: Pretty Derby, built with Go and Bubble Tea.

## Features

- **Horse Scouting**: Choose from 12 unique horses with different stats and breeds
- **Training System**: Weekly training calendar with 4 training types (Stamina, Speed, Technique, Mental)
- **Racing**: Live race simulation with real-time progress bars and commentary
- **Season Progression**: 24-week seasons with aging and long-term progression
- **Supporter System**: Support cards that provide training bonuses
- **Save/Load**: Persistent game state with JSON saves
- **Beautiful TUI**: Colorful terminal interface with Unicode icons and progress bars

## Installation

```bash
git clone <repository>
cd goderby
go build -o goderby cmd/main.go
./goderby
```

## How to Play

1. **Scout a Horse**: Choose your racing partner from available horses
2. **Train Weekly**: Plan training schedules to improve your horse's stats
3. **Enter Races**: Compete in races matching your horse's level
4. **Progress Seasons**: Advance through seasons as your horse ages and improves
5. **Achieve Fame**: Win races, gain fans, and become a racing legend

## Game Mechanics

### Training Types
- **Stamina**: Improves endurance for longer races
- **Speed**: Increases base racing speed
- **Technique**: Enhances consistency and skill
- **Mental**: Improves performance under pressure

### Race Strategy
- **Formation**: Lead, Draft, or Mount tactics
- **Pace**: Fast, Even, or Conservative racing approach

### Progression
- Horses age each season (2-10 years old)
- Stats can be improved through training up to maximums
- Fatigue and morale affect training and racing performance
- Win races to gain fans and prize money

## Controls

- **↑/↓**: Navigate menus
- **←/→**: Navigate strategy options
- **Enter/Space**: Select/Confirm
- **ESC/q**: Go back/Quit
- **r**: Rest (in training mode)
- **i**: Inspect (in scout mode)
- **n**: Next week/season

## Technical Details

- Built with Go 1.21+
- Uses Bubble Tea for TUI framework
- Lip Gloss for styling and colors
- JSON for data storage and saves
- Modular architecture with separate models, UI, and game logic

## File Structure

```
goderby/
├── cmd/main.go              # Main application entry point
├── internal/
│   ├── models/              # Game data structures
│   ├── ui/                  # TUI components and views
│   ├── game/                # Game logic and simulation
│   └── data/                # Data loading and persistence
├── assets/                  # Game data files (auto-generated)
│   ├── horses.json
│   ├── supporters.json
│   ├── races.json
│   └── saves/game.json
└── go.mod                   # Go module definition
```

---

🎮 **Enjoy racing to victory in Go Derby!** 🏆
