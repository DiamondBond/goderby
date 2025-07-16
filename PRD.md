# Derby Go! Expansion: Endgame & Social Features

**Version:** 1.0  
**Date:** July 16, 2025

---

## 1. Overview

This PRD defines the expansion of Derby Go!’s gameplay loop to include a richer endgame experience and enhanced social sharing mechanics. The goal is to improve player retention by adding depth to horse progression, retirement mechanics, and social sharing of achievements.

---

## 2. Goals & Objectives

- **Extend endgame progression** by allowing players to take their horses beyond racing.
- **Introduce horse retirement systems** with meaningful post-retirement activities.
- **Enable social sharing** of achievements such as high scores and season summaries to drive community engagement.

---

## 3. Features

### 3.1. Endgame Loop Expansion

- **How far can you take a horse?**
  - Track a horse’s career milestones (wins, earnings, records).
  - Introduce diminishing returns (age, fatigue) to encourage retirement decisions. (Already completed)
- **Season Summary Enhancement**
  - Expand the current summary screen to include:
    - Lifetime stats for each horse.
    - Career highlights (e.g., most prestigious race won, longest winning streak).
    - Fanbase size and awards earned.

### 3.2. Horse Retirement Mechanics

- **Retirement Homes**
  - Players can retire a horse into different tiers of retirement homes:
    - **Basic:** Free, no bonuses.
    - **Premium:** Costs in-game currency, unlocks passive benefits.
- **Post-Retirement Roles**
  - Retired horses can generate passive income or fan engagement as:
    - **Show Horses:** Earn periodic fame points or cosmetic rewards.
    - **Stud/Breeding:** Potential future mechanic for lineage-based gameplay.
- **Awards & Honors**
  - Retired horses can earn awards (e.g., “Hall of Fame,” “Crowd Favorite”) displayed in a Horse Summary Gallery.

### 3.3. Social Sharing Features

- **High Scores & Season Summaries**
  - Allow players to share season summaries and career highlights with friends:
    - In-game friend system or external share (image export).
    - Leaderboards for “Longest Career,” “Highest Earnings,” etc.
- **Horse Profiles**
  - Generate shareable horse profiles with:
    - Name, stats, image, awards, and retirement status.

---

## 4. User Stories

- As a player, I want to retire my horse into a prestigious home so I can earn passive bonuses and see its legacy continue.
- As a player, I want to share my season highlights with friends to showcase my achievements.
- As a player, I want to view a gallery of all my retired horses and their awards to feel a sense of progression.

---

## 5. Success Metrics

- Increase in average session length by 15%.
- 20% of players use retirement mechanics within 1 week of release.
- 30% of active players share a season summary or horse profile.

---

## 6. Q/A

- Should post-retirement roles (e.g., stud/breeding) affect gameplay or remain cosmetic for now?
  - it should be a passive source of fame/income.
- Will sharing be limited to in-game friends, or include social media integration (e.g., Instagram, X/Twitter)?
  - there is no ingame friends or social media integration, i just want these pretty TUI ui cards with the info like the existing summary screen and etc, then users can just screenshot this and share it.
- Should there be a cap on the number of retired horses stored in the gallery?
  - yes you can only have 8 retired horses, maybe purchasable slots, start with only 1-2 available by default and the slots cost alot for the good retirement homes, the cheaper ones dont foster the horse properly for it to become a proper stud/breeding horse, so less passive income/fame, whereas the more expensive homes foster a good horse into producing a lot of passive income/fame.

---

## 7. Dependencies

- UI/UX redesign for season summary and retirement screens.
- Support for persistent horse profiles and leaderboards.
