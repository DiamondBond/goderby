# Go Derby - Quick Start Guide

## Installation & Setup

1. **Build the game:**
   ```bash
   go build -o goderby cmd/main.go
   ```

2. **Run the game:**
   ```bash
   ./goderby
   ```

## First Time Playing

1. **Scout a Horse**: Select "Scout Horse" from main menu
2. **Browse Available Horses**: Use ↑/↓ to navigate, Enter to inspect
3. **Select Your Horse**: Press Enter when viewing a horse's details

## Training Your Horse

1. **Access Training**: Select "Train" from main menu
2. **Select Training Day**: Choose Monday through Saturday
3. **Choose Training Type**: 
   - Stamina: For endurance in longer races
   - Speed: For basic racing speed
   - Technique: For consistency and skill
   - Mental: For performance under pressure
4. **Rest When Needed**: Press 'r' on any day to rest and reduce fatigue

## Racing

1. **Select Race**: Choose from available races matching your horse's level
2. **Set Strategy**:
   - Formation: Lead (fast start), Draft (mid-pack surge), Mount (strong finish)
   - Pace: Fast (early speed), Even (consistent), Conserve (save energy)
3. **Watch Live Race**: Real-time progress bars and commentary
4. **Collect Rewards**: Prize money and fans based on finishing position

## Season Progression

- **Complete 24 weeks** to finish a season
- **Age your horse** each season (2-10 years old)
- **Track achievements** in Season Summary
- **Plan for retirement** around age 8-10

## Tips for Success

- **Balance training types** to improve all stats
- **Monitor fatigue** - rest when needed
- **Start with easier races** and work up to Grade 1
- **Save frequently** using "Save & Quit"

## Game Features

✅ **12 unique horses** with different stats and breeds  
✅ **Weekly training system** with 4 training types  
✅ **Live race simulation** with real-time updates  
✅ **24-week seasons** with aging progression  
✅ **6 race grades** from Maiden to Grand Prix  
✅ **Supporter system** for training bonuses  
✅ **Save/load functionality** with JSON persistence  
✅ **Beautiful TUI** with colors and progress bars  

## Controls Reference

| Key | Action |
|-----|--------|
| ↑/↓ | Navigate menus |
| ←/→ | Navigate strategy options |
| Enter/Space | Select/Confirm |
| ESC/q | Go back/Quit |
| r | Rest (training mode) |
| i | Inspect (scout mode) |
| n | Next week/season |

## File Structure

```
./
├── goderby              # Game executable
├── assets/              # Game data (auto-generated)
│   ├── horses.json      # Available horses
│   ├── supporters.json  # Support cards
│   ├── races.json       # Race definitions
│   └── saves/game.json  # Your save file
└── README.md            # Documentation
```

Enjoy your journey to racing greatness! 🏇🏆