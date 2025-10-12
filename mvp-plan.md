# Movodoro MVP Implementation Plan

## Overview
A CLI tool that generates and records short movement snacks for 5-10 minute work breaks.

## Tech Stack
- **Language**: Go (single binary, no venv hassle)
- **Snack definitions**: `movos/` directory, one YAML file per category
- **History storage**: `~/.movodoro/history.log` (lightweight append-only log)
- **Config**: `~/.movodoro/config.yaml` (optional user preferences)

## Data Schema

### Category File
Location: `movos/<category-slug>.yaml`

Example (`movos/club-flow.yaml`):
```yaml
category: Club Flow
code: CF
weight: 1.0
default_rpe: 4  # Default RPE for this category
tags: [clubsx, flowx]

snacks:
  - code: shield-cast
    title: Light shield casts & pendulums
    description: |
      3-5 minutes of shield casting and pendulum work.
      Focus on smooth transitions.
    duration_min: 3
    duration_max: 5
    # rpe: inherits default_rpe (4) from category
    max_per_day: 1
    weight: 1.0
    every_day: false
    tags: [1hcx]

  - code: mills
    title: Single-hand mills → reverse mills
    description: Flow through single-hand mill patterns, transitioning to reverse mills.
    duration_min: 3
    duration_max: 5
    max_per_day: 1
    max_per_week: 3
    weight: 1.0
    every_day: false
    tags: [1hcx]

  - code: double-circles
    title: Light double-club side circles
    description: Flow through double-club patterns with side circles.
    duration_min: 3
    duration_max: 5
    max_per_day: 1
    max_per_week: 3
    weight: 1.0
    every_day: false
    tags: [2hcx]
    # Note: requires two light clubs
```

Example (`movos/micro-strength.yaml`):
```yaml
category: Micro-Strength
code: MS
weight: 1.0
tags: [strengthx]

snacks:
  - code: swings-flow
    title: 10 one-arm swings per side + mobility flow
    description: Explosive swings followed by gentle mobility work.
    duration_min: 3
    duration_max: 5
    rpe: 7  # Hard effort
    max_per_day: 1
    weight: 1.0
    every_day: false
    tags: [kbx, swingx]

  - code: getups
    title: 5 get-ups (light and slow)
    description: Turkish get-ups with focus on control and form.
    duration_min: 5
    duration_max: 5  # Fixed duration
    rpe: 6  # Moderate-hard
    max_per_day: 1
    weight: 1.0
    every_day: false
    tags: [kbx, getupx]
    # Use 8-12kg for slow, controlled movement
```

**Field Definitions**:
- `category.code`: 2-3 char category code (e.g., "CF", "OSR")
- `category.default_rpe`: Default RPE for all snacks in this category (optional, integer 1-10)
  - Snacks inherit this value unless they specify their own `rpe`
- `category.tags`: Array of tags for the category (e.g., ["clubsx", "flowx"])
- `snack.code`: Short slug derived from snack title (manually settable, e.g., "shield-cast", "mills")
  - Auto-generated from title if not provided (lowercase, hyphenated)
  - Full snack ID in logs: `{category.code}-{snack.code}` (e.g., `CF-shield-cast`)
- `snack.tags`: Array of tags for the snack (e.g., ["1hcx", "kbx", "getupx"])
  - Snack inherits category tags, can add additional specific tags
  - **All tags must end with 'x'** for easy grepping (e.g., "kbx" not "kb", "flowx" not "flow")
  - Common tag patterns:
    - Equipment: kbx, clubsx, ropex, bandsx
    - Method (where hand style matters): 1hcx, 2hcx
    - Movement type: flowx, strengthx, breathx, mobilityx, getupx, swingx
- `duration_min`: Minimum duration in minutes (integer)
- `duration_max`: Maximum duration in minutes (integer, can equal min for fixed duration)
  - Display string auto-generated: "5min" or "3-5min"
- `rpe`: Rate of Perceived Exertion (1-10 scale, integer, optional if category has default_rpe)
  - Inherits from `category.default_rpe` if not specified
  - Override at snack level when RPE differs from category default
  - 1-2: Very light (breathing, gentle mobility)
  - 3-4: Light (active mobility, light flows)
  - 5-6: Moderate (most strength work, moderate intensity)
  - 7-8: Hard (heavy carries, intense circuits)
  - 9-10: Very hard (max effort, rarely used for snacks)
- `weight`: Relative probability multiplier (default: 1.0)
- `every_day`: If true, heavily bias toward this snack (10x weight)
- `max_per_day`: Maximum times per day (default: 1)
- `max_per_week`: Maximum times per week (optional)

### History Log
Location: `~/.movodoro/history.log`

Format: `TIMESTAMP CODE STATUS DURATION RPE`

Example:
```
2025-10-12T10:30:00Z CF-shield-cast complete 4 4
2025-10-12T14:15:00Z OSR-head-nods complete 6 2
2025-10-12T16:45:00Z TS-cossack skipped 4 5
2025-10-12T19:00:00Z RB-box-breathing complete 4 1
# Daily totals - Time: 14min, RPE: 7 (only counts completed snacks)
```

**Duration Calculation**:
- Uses middle of `duration_min` and `duration_max`, rounded up
- Examples: 3-5min → 4, 3-6min → 5, 5-5min → 5
- Enables tracking of actual time spent on movement

**Benefits**:
- Lightweight (~50-70 bytes per entry vs 100+ for JSON)
- Fast append operations
- Easy to parse and grep
- Human-readable codes can be JOINed to snack definitions for reports
- Self-documenting (codes hint at what the snack was)
- RPE tracking enables load management
- Duration tracking shows time invested in movement

### Config File
Location: `~/.movodoro/config.yaml`

```yaml
# Movodoro configuration
movos_dir: ./movos
history_log: ~/.movodoro/history.log

# Weighting
everyday_weight_multiplier: 10.0
max_everyday_snacks: 4

# Recency boost - resurface neglected snacks
recency_boost_days: 7          # Boost snacks not done in N days
recency_boost_multiplier: 2.0  # Multiplier for stale snacks

# RPE (Rate of Perceived Exertion) management
max_daily_rpe: 30              # Max cumulative RPE per day
recovery_rpe_threshold: 2      # After max_daily_rpe hit, only show snacks <= this RPE
```

**Recency Settings**:
- `recency_boost_days`: If a snack hasn't been done in N days, boost its weight
- `recency_boost_multiplier`: Multiplier for stale snacks (e.g., 2x weight if not done in a week)

**RPE Settings**:
- `max_daily_rpe`: Cumulative RPE threshold for the day (default: 30)
  - Example: 5 moderate snacks (RPE 6) = 30 total
  - After threshold: auto-filter to recovery snacks only
- `recovery_rpe_threshold`: Max RPE for recovery snacks (default: 2)
  - Light breathing, gentle mobility only after hitting daily limit

## Selection Algorithm

1. **Load Categories**: Read all `movos/*.yaml` files
2. **Validate**: Warn if >4 snacks have `every_day: true`
3. **Parse History**: Build lookup of last completion time per snack and calculate daily RPE total
4. **Check RPE Limit**: If daily RPE >= `max_daily_rpe`, auto-filter to snacks with `rpe <= recovery_rpe_threshold`
5. **Filter Available Snacks**:
   - Apply tag filter if `--tags` provided (match ANY tag via OR logic)
   - Apply duration filter if `--min-duration` / `--max-duration` provided
     - Include if snack's duration range overlaps with requested range
     - Example: snack is 3-5min, request is 4-10min → include (overlap at 4-5min)
   - Apply RPE filter if `--max-rpe` provided or if daily limit reached
   - Parse history log for today/this week
   - Exclude snacks that hit `max_per_day` or `max_per_week`
6. **Handle Exhaustion**: If all filtered out, inform user and allow repeats
7. **Apply Dynamic Weighting**:
   - Start with base `snack.weight`
   - **Every-day boost**: If `every_day: true`, multiply by config value (default: 10x)
   - **Recency boost**: If snack not done in `recency_boost_days` (default: 7), multiply by `recency_boost_multiplier` (default: 2x)
   - **Never-done boost**: If snack never completed, multiply by 3x
8. **Weighted Random Selection**:
   - Weight categories by `category.weight`
   - Randomly select category
   - Within category, apply all dynamic weights to snacks
   - Randomly select snack using weighted probabilities
9. **Present Snack**: Show with interactive prompt and RPE rating

**Example Weight Calculation**:
```
Snack: "CF-shield-cast"
Base weight: 1.0
Every-day: false (no boost)
Last done: 10 days ago (recency boost: 2x)
Final weight: 1.0 × 2.0 = 2.0

Snack: "OSR-head-nods"
Base weight: 1.0
Every-day: true (10x boost)
Last done: yesterday (no recency boost)
Final weight: 1.0 × 10.0 = 10.0

Snack: "TS-cossack"
Base weight: 1.0
Every-day: false
Last done: never (3x boost)
Final weight: 1.0 × 3.0 = 3.0
```

## CLI Commands

### `movodoro [options]`
Main command - get a movement snack.

**Options**:
- `--tags` / `-t`: Filter snacks by tags (comma-separated, e.g., `--tags kbx,flowx`)
  - Matches if snack has ANY of the specified tags (OR logic)
  - Tags are method/equipment focused (e.g., kbx for all kettlebell work, 1hcx vs 2hcx for club hand-style)
- `--min-duration` / `-m`: Minimum duration in minutes (e.g., `--min-duration 7`)
- `--max-duration` / `-M`: Maximum duration in minutes (e.g., `--max-duration 15`)
- `--duration` / `-d`: Shorthand for duration range (e.g., `-d 7-15` sets min=7, max=15)
- `--min-rpe` / `-r`: Filter to snacks with RPE at or above this value (e.g., `--min-rpe 6` for intense work)
- `--max-rpe` / `-R`: Filter to snacks with RPE at or below this value (e.g., `--max-rpe 3` for recovery)

**Flow**:
1. Select random weighted snack (filtered by options)
2. Display with clean, visually appealing formatting:
   ```
   ╭──────────────────────────────────────────────────────────────╮
   │                        CLUB FLOW                             │
   ╰──────────────────────────────────────────────────────────────╯

   Light shield casts & pendulums

   3-5 minutes of shield casting and pendulum work.
   Focus on smooth transitions.

   Duration: 3-5min  •  RPE: 4/10  •  Tags: clubsx, flowx, 1hcx
   Daily RPE: 12/30

   ──────────────────────────────────────────────────────────────

   [c]omplete  │  [d]ifferent  │  [q]uit  ›
   ```
3. Handle input:
   - `c`: Mark as complete
     - Prompt: "Minutes spent (default: 4): "
     - User can enter actual duration or press Enter for calculated default
     - Log with actual or default duration
   - `d`: Generate different snack
   - `q`: Quit without logging

**Examples**:
```bash
movodoro                      # Random snack from all available
movodoro --tags kbx           # Only kettlebell snacks
movodoro -t 1hcx,2hcx         # Only club work (1 or 2 hand)
movodoro -t breathx,flowx     # Only breath work or flow movements
movodoro -d 5-10              # Snacks between 5-10 minutes
movodoro --max-duration 5     # Quick snacks (≤5 min)
movodoro -t kbx -d 3-7        # Kettlebell work, 3-7 minutes
movodoro --max-rpe 3          # Recovery day: light snacks only
movodoro -R 2                 # Very light snacks (breathing, gentle mobility)
movodoro --min-rpe 6          # Intense work: moderate-hard to hard snacks
movodoro -r 7 -t kbx          # Hard kettlebell work only
```

**RPE Management**:
- System automatically tracks daily RPE total
- After hitting `max_daily_rpe` (default: 30), only shows snacks with RPE ≤ 2
- User is informed: "Daily RPE limit reached (30/30). Showing recovery snacks only."
- Can manually request low-RPE days with `--max-rpe` flag

### `movodoro report [period]`
Generate usage reports.

**Periods**:
- `today` (default): Detailed list with timestamps
- `week`: Current week tally
- `month`: Current month tally
- `YYYY-MM-DD..YYYY-MM-DD`: Custom date range

**Output Examples**:

*Today*:
```
=== Movodoro Report: 2025-10-12 ===
10:30 AM - Club Flow: Light shield casts & pendulums
02:15 PM - OS Resets: Head nods → rocking → crawling
04:45 PM - Tribal Squats: Alternating Cossack squats

Total: 3 snacks
```

*Week/Month*:
```
=== Movodoro Report: Week of 2025-10-06 ===
Club Flow (4 total)
  - Light shield casts & pendulums: 2
  - Single-hand mills: 2

OS Resets (5 total)
  - Head nods → rocking → crawling: 3
  - Shoulder rolls → segmental rolling: 2

Total: 9 snacks across 2 categories
```

### `movodoro list [--tags tag1,tag2]`
Show all available snacks organized by category.

**Options**:
- `--tags` / `-t`: Filter by tags (comma-separated)

```
=== Available Movement Snacks ===

Club Flow (CF) [clubsx, flowx] - 6 snacks
  - Light shield casts & pendulums [1hcx]
  - Single-hand mills [1hcx]
  - Light double-club side circles [2hcx]
  ...

OS Resets (OSR) [bodyx, mobilityx] - 6 snacks
  - Head nods → rocking → crawling [flowx]
  - Shoulder rolls → segmental rolling [mobilityx]
  ...

Total: 45 snacks across 10 categories
```

### `movodoro tags`
List all available tags with usage counts.

```
=== Available Tags ===

Equipment:
  kbx (12 snacks)
  clubsx (8 snacks)
  ropex (3 snacks)

Method:
  1hcx (5 snacks)
  2hcx (3 snacks)
  getupx (4 snacks)
  swingx (5 snacks)

Movement Type:
  flowx (15 snacks)
  strengthx (8 snacks)
  breathx (6 snacks)
  mobilityx (10 snacks)

Total: 12 unique tags across 45 snacks
```

### `movodoro init`
Bootstrap the `movos/` directory with example snacks from `snack-ideas.md`.

Creates JSON files for each category with sensible defaults.

### `movodoro stats`
Show usage statistics and streaks.

```
=== Movodoro Statistics ===
Current streak: 5 days
Longest streak: 12 days
Total snacks completed: 127
Most frequent: OS Resets (23 times)
Favorite: Head nods → rocking → crawling (8 times)
```

## Implementation Steps

1. **Project Setup**
   - Initialize Go module
   - Set up project structure
   - Add dependencies:
     - CLI framework (cobra/urfave-cli)
     - YAML parsing (gopkg.in/yaml.v3)
     - Terminal UI/formatting (e.g., lipgloss, bubbletea for nice formatting, or just colored/boxed output)

2. **Core Data Models**
   - Define structs: Category, Snack, HistoryEntry
   - YAML marshaling/unmarshaling (using gopkg.in/yaml.v3)
   - Log file parser
   - Auto-generate snack codes from titles (if not provided)

3. **File Operations**
   - Load all `movos/*.yaml` files
   - Parse history log
   - Append to history log
   - Config file handling

4. **Selection Logic**
   - Tag filter (match ANY tag via OR logic)
   - Duration filter (range overlap logic)
   - RPE filter and daily RPE tracker
   - Frequency checker (max per day/week)
   - Recency tracker (last completion time per snack)
   - Dynamic weight calculator (every-day, recency, never-done boosts)
   - Weighted random selection

5. **CLI Interface**
   - Main command with interactive prompts, tag filtering, and duration filtering
   - Visually appealing output with boxes, separators, and formatting
   - Report command with period parsing
   - List command with tag filtering
   - Tags command (list all tags with counts)
   - Stats command

6. **Init Command**
   - Parse `snack-ideas.md`
   - Auto-generate snack codes from titles (lowercase, hyphenated, first 2-3 words)
   - Generate YAML files with reasonable defaults and helpful comments
   - Create `movos/` directory

7. **Validation & Warnings**
   - Check for >4 every-day snacks
   - Validate YAML structure and required fields
   - Check for duplicate snack codes (must be unique across all categories)
   - Validate all tags end with 'x' suffix
   - Validate duration fields (min/max must be positive integers, max >= min)
   - Handle missing files gracefully

8. **Build & Package**
   - Cross-compile for macOS/Linux/Windows
   - Single binary output
   - Installation instructions

## Code Organization

```
movodoro/
├── cmd/
│   └── movodoro/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── models/
│   │   ├── category.go
│   │   ├── snack.go
│   │   └── history.go
│   ├── loader/
│   │   └── loader.go
│   ├── selector/
│   │   └── selector.go
│   ├── history/
│   │   └── history.go
│   ├── reporter/
│   │   └── reporter.go
│   └── cli/
│       ├── main_cmd.go
│       ├── report_cmd.go
│       ├── list_cmd.go
│       ├── init_cmd.go
│       └── stats_cmd.go
├── movos/           # User's snack definitions
├── go.mod
├── go.sum
└── README.md
```

## Future Considerations (Post-MVP)

- Web interface with FastAPI
- Mobile app integration
- Visual timers during snacks
- Video demonstrations
- Social sharing/challenges
- Integration with Pomodoro timers
- Snack difficulty ratings
- Equipment requirements filter
- Custom snack creation via CLI
- Export/import snack libraries
