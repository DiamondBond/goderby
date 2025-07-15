📘 1. Overview

Product Name

    TermaHorse Derby – A terminal-based, immersive horse racing life-sim inspired by Umamusume: Pretty Derby, powered by Go + Bubble Tea.

Target Users

    Terminal-savvy gamers

    Lovers of sports simulation & visual novels

    Developers and TUI enthusiasts

Goals

    Deliver a compelling training, management, and racing experience through text/UI in terminal

    Recreate core mechanics (scouting, training, racing) in a concise TUI

    Showcase Go + Bubble Tea’s potential for interactive storytelling

🎯 2. Key Features
2.1 Trainee Scouting & Supporters

    Scout screen: list of available horses (trainees), each with basic stats—name, breed, stamina, speed, fan support level.

    Select supporters: up to 4 supporter cards with passive bonuses (e.g., +5% stamina gain).

2.2 Training System

    Training Dashboard: weekly calendar of training slots (Mon–Sat).

    Training Types: Stamina, Speed, Technique, Mental.

    Status tracking: stamina level, fatigue, morale.

    Events: random popups (e.g., “Your horse strained a muscle!”) with options to rest, treat, or push on.

2.3 Racing Simulation

    Race Listing: upcoming races with distance, grade, prize, and competitors.

    Race Strategy Setup: choose formation (lead, draft, mount), pace (fast, even, conserve).

    Live Race View: ANSI-based progress bar per horse, updating each turn with position changes, random events, commentary lines.

    Results & Rewards: finishing order, earned fans, money, bonuses.

2.4 Progression & UI Flow

    Season Loop: 6-month training → race → offseason summary

    Long-term progression: horse ages up, retires or advances to higher-tier races

    Save/Load: JSON save file with training history, stats, supporters, wins.

📐 3. UX / UI Interaction

Build on Bubble Tea (https://github.com/charmbracelet/bubbletea):

    Main View (menu): Scout | Train | Race | Summary | Save/Quit

    Scout View: list with keyboard nav (↑↓), enter to inspect, space to pick.

    Training View: grid UI – select day + training type.

    Event Popups: modal overlay with question, yes/no selection.

    Race View: multi-line, updating per tick; refresh live with Bubble Tea's Update loop.

    Summary View: show season stats, summary table, navigation to next season or exit.

🖥️ 4. Technical
4.1 Tech Stack

    Go (1.21+) + Bubble Tea + Lip Gloss for styling

    Data storage: embedded JSON files for horse/supporters metadata; runtime state saved to JSON

4.2 Architecture

    Model: structs for Horse, Supporter, TrainingDay, Race, Season

    TUI: separate Bubble Tea models for each state (ScoutModel, TrainModel, RaceModel, SummaryModel)

    Controller: commands transitioning model -> view

    Simulator: random events engine, race RNG, fatigue, morale, stamina updates

4.3 Assets (Terminal)

    ANSI color-coded bars (e.g., green stamina, red fatigue)

    Unicode icons (🏇, ⭐️, 🏆)

    Minimal ASCII art for headings

⚙️ 5. Requirements
ID	Requirement	Priority
FR1	List 10 trainees with stats	H
FR2	Choose up to 4 supporters	H
FR3	Weekly training that updates stats + fatigue	H
FR4	Random training events	M
FR5	Race simulator with live progress display	H
FR6	Season summary and progression	H
FR7	Save/load game state	H
NFR1	Runs smoothly in common terminals (80x24)	H
NFR2	Under 10 MB binary	M
NFR3	Test coverage ≥ 80% for core logic	M
📈 6. MVP Roadmap

    Core data structs + JSON loaders

    TUI main menu + Scout screen

    Training calendar + stat updates

    Simple race sim + output

    Add random events and supporters

    Save/load functionality

    Season loop + summary view

    Polish UI: colors, layout, responsiveness

    Add fans, money, race rewards

    Testing, documentation, release

🧪 7. Success Metrics

    MVP completeness: all core FRs implemented

    Usability: zero reported blocker bugs on terminal platforms

    Adoption: 100+ GitHub stars in 1 month post-launch

    Performance: race simulation completes < 1 s

✅ 8. Out Of Scope (v1)

    3D graphics, voiceover, VR

    Multiplayer

    Complex bonds, idol-style story arcs

    Licensing Umamusume IP (fictional horse names only)

🚀 Summary

TermaHorse Derby reimagines the thrill of Umamusume: Pretty Derby in a charming, terminal-based life sim. With Go and Bubble Tea, you'll deliver scouting, training, racing, and progression—all in expressive text and color.
