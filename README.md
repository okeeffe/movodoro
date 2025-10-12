# Movodoro

A CLI tool for generating and tracking movement snacks during work breaks. Stay active, prevent burnout, and build movement variety into your day with intelligent snack selection and RPE-based load management.

## What is Movodoro?

Movodoro helps you take 3-10 minute movement breaks throughout your workday by:
- **Randomly selecting** from your library of movement patterns
- **Intelligently weighting** snacks you haven't done recently
- **Managing load** with RPE (Rate of Perceived Exertion) tracking
- **Auto-recovery** mode when you've accumulated too much daily load
- **Tracking your practice** in a simple log format

Perfect for knowledge workers who want to incorporate more movement without the decision fatigue.

## Installation

### Prerequisites
- Go 1.21 or higher

### Build from source

```bash
git clone <your-repo-url>
cd movodoro
go build -o movodoro
```

Move the binary to your PATH:
```bash
mv movodoro /usr/local/bin/
```

Or use it directly:
```bash
./movodoro help
```

## Configuration

Movodoro looks for movement snacks in a configurable directory. You have two options:

### Option 1: Environment Variable (Recommended)

Set `MOVODORO_MOVOS_DIR` to point to your movos directory:

```bash
# Add to your ~/.zshrc or ~/.bashrc
export MOVODORO_MOVOS_DIR=~/my-movement-snacks
```

### Option 2: Default Location

If no environment variable is set, movodoro looks in `~/.movodoro/movos/`

```bash
mkdir -p ~/.movodoro/movos
# Copy example movos
cp movos-examples/* ~/.movodoro/movos/
```

### Check Your Configuration

```bash
movodoro config
```

This will show:
- Where movodoro is looking for your movos
- Where logs are stored
- How many snacks were found
- Warnings if the movos directory doesn't exist

### File Locations

Movodoro stores data in `~/.movodoro/`:
- `~/.movodoro/logs/YYYYMMDD.log` - Daily history logs
- `~/.movodoro/current` - Currently selected snack code

## Quick Start

1. **Get a random movement snack:**
   ```bash
   movodoro get
   ```

2. **Complete the snack:**
   ```bash
   movodoro done
   # You'll be prompted for actual duration
   ```

3. **View your daily report:**
   ```bash
   movodoro report
   ```

4. **Get a recovery snack (low intensity):**
   ```bash
   movodoro get -R 2
   ```

## Creating Your Movement Snacks (Movos)

### Directory Structure

Movement snacks are defined in YAML files in the `movos/` directory:

```
movos/
  ├── breath-work.yaml
  ├── strength.yaml
  ├── mobility.yaml
  └── ...
```

### YAML Format

Each file defines a **category** of movement snacks:

```yaml
# Reset & Breath - Breathing practices
category: Reset & Breath
code: RB
weight: 1.5           # Category weight (higher = more likely to be selected)
default_rpe: 1        # Default RPE for all snacks in this category
tags: [breathx, resetx]

snacks:
  - code: box-breathing
    title: 4-7-8 or box breathing
    description: |
      Standing or kneeling breath work. 4-7-8: inhale 4, hold 7, exhale 8.
      Box: 4-4-4-4 pattern.
    duration_min: 3
    duration_max: 5
    rpe: 1              # Override default_rpe if needed
    max_per_day: 2      # Can be done twice per day
    weight: 1.5         # Snack-specific weight multiplier
    every_day: true     # Gets 10x boost in selection
    tags: []            # Additional tags beyond category tags

  - code: deep-breath
    title: Deep belly breathing
    description: Focus on diaphragmatic breathing.
    duration_min: 2
    duration_max: 4
    # rpe: inherits default_rpe (1)
    max_per_day: 1
    weight: 1.0
    every_day: false
    tags: []
```

### Field Reference

#### Category Fields
- **category**: Human-readable name
- **code**: 2-4 letter code (e.g., "RB", "CF", "TS")
- **weight**: Base probability multiplier (default: 1.0)
- **default_rpe**: Default RPE for all snacks (1-10 scale)
- **tags**: Array of category-level tags (all must end with 'x')

#### Snack Fields
- **code**: Unique identifier within category (slug format)
- **title**: Display name
- **description**: Instructions (supports multi-line with `|`)
- **duration_min**: Minimum duration in minutes
- **duration_max**: Maximum duration in minutes
- **rpe**: Rate of Perceived Exertion (1-10), inherits `default_rpe` if not set
- **max_per_day**: Maximum times per day (0 = unlimited)
- **max_per_week**: Maximum times per week (optional)
- **weight**: Snack-specific weight multiplier
- **every_day**: Boolean, gets 10x selection boost
- **tags**: Additional tags specific to this snack

### Tag Conventions

All tags must end with 'x' for easy grepping:
- **Equipment**: `kbx` (kettlebell), `clubsx` (steel clubs), `bandsx` (resistance bands)
- **Body type**: `bodyx` (bodyweight only)
- **Method**: `1hcx` (one-hand club), `2hcx` (two-hand club), `swingx` (swinging movements)
- **Category**: `breathx`, `strengthx`, `mobilityx`, `flowx`, `hangx`, `crawlx`, `balancex`

Snacks inherit category tags and can add their own.

### Full Code Format

Each snack gets a full code: `{CATEGORY_CODE}-{snack-code}`

Example: `RB-box-breathing` (Reset & Breath - Box Breathing)

## Command Reference

### Get a Snack

```bash
movodoro get [options]
```

**Options:**
- `-t, --tags TAGS` - Filter by tags (comma-separated)
- `-d, --duration MINS` - Exact duration
- `-m, --min-duration MINS` - Minimum duration
- `-M, --max-duration MINS` - Maximum duration
- `-r, --min-rpe RPE` - Minimum RPE (for intense work)
- `-R, --max-rpe RPE` - Maximum RPE (for recovery)

**Examples:**
```bash
movodoro get -t kbx,swingx          # Kettlebell swings
movodoro get -R 2                   # Recovery snacks only
movodoro get -r 7 -t kbx            # Hard kettlebell work
movodoro get -d 5                   # Exactly 5 minutes
movodoro get -m 3 -M 7 -t breathx   # 3-7 min breath work
```

### Complete a Snack

```bash
movodoro done [CODE]
```

If no code is provided, marks the most recently selected snack as done. You'll be prompted to enter the actual duration (defaults to the midpoint of the snack's range).

**Example:**
```bash
movodoro done                    # Mark current snack done
movodoro done RB-box-breathing   # Mark specific snack done
```

### Skip a Snack

```bash
movodoro skip [CODE]
```

Records that you skipped a snack (doesn't count toward RPE or duration).

### View Reports

```bash
movodoro report [period] [options]
```

**Periods:** `day`, `week`, `month` (week and month not yet implemented)

**Options:**
- `--md, --markdown` - Output in markdown format (great for copy-pasting to logs)

**Examples:**
```bash
movodoro report                  # Today's report
movodoro report day              # Same as above
movodoro report --md             # Markdown format
```

### Clear Today's History

```bash
movodoro clear
```

Removes all of today's entries from history (requires confirmation). Useful for testing.

### Show Configuration

```bash
movodoro config
```

Displays current configuration including movos directory, logs directory, and diagnostic information. Useful for troubleshooting setup issues.

### Check Everyday Snacks

```bash
movodoro everyday
```

Shows all snacks marked with `every_day: true` and their completion status for today. Great for tracking your daily movement essentials.

**Example output:**
```
═══════════════════════════════════════
  EVERY DAY SNACKS
═══════════════════════════════════════

✅ Box breathing
   Code: BR-box-breathing | RPE: 1 | Duration: 3-5 min
   Completed 2 time(s) today

❌ Hip circles and leg swings
   Code: MOB-hip-circles | RPE: 3 | Duration: 5-7 min
   Not yet done today

Summary: 1/2 everyday snacks completed
```

### Version

```bash
movodoro version
```

## How Selection Works

Movodoro uses a weighted random selection algorithm with multiple boost factors:

### Base Weight
Each snack starts with its `weight` value (default: 1.0), multiplied by the category's `weight`.

### Boosts Applied
1. **Every-day boost (10x)**: Snacks marked `every_day: true`
2. **Never-done boost (3x)**: Snacks you've never completed
3. **Recency boost (2x)**: Snacks not done in 7+ days

### Filters
- **Tags**: Only snacks matching ALL specified tags
- **Duration**: Range overlap (snack's [min, max] overlaps with filter)
- **RPE**: Min/max thresholds
- **Frequency**: Snacks at `max_per_day` limit are excluded

### Auto-Recovery Mode

When your daily cumulative RPE reaches 30 (configurable), Movodoro automatically limits selections to RPE ≤ 2, ensuring you don't overtrain.

## File Formats

### Daily Log Files

Location: `~/.movodoro/logs/YYYYMMDD.log`

Each day gets its own log file (e.g., `20251012.log` for October 12, 2025).

Format: Append-only log, one line per entry
```
TIMESTAMP CODE STATUS DURATION RPE
```

Example `~/.movodoro/logs/20251012.log`:
```
2025-10-12T14:09:37+01:00 GUP-naked-getups done 4 3
2025-10-12T14:15:22+01:00 RB-box-breathing done 5 1
2025-10-12T14:20:18+01:00 CF-shield-cast skip 0 0
```

**Benefits of daily files:**
- Easy archival and backup
- Simple cleanup (just delete a file)
- Fast today-focused operations
- Natural organization by date

### Current Snack

Location: `~/.movodoro/current`

Contains the code of the most recently selected snack for quick `done`/`skip` commands.

## RPE Scale

Rate of Perceived Exertion (1-10):
- **1-2**: Very light (breath work, light mobility)
- **3-4**: Light-moderate (easy movement, gentle stretching)
- **5-6**: Moderate (active movement, light strength work)
- **7-8**: Hard (strength training, intense movement)
- **9-10**: Very hard (maximum effort, heavy loads)

## Tips

1. **Start your day with breath work**: Use `-t breathx -R 2` for morning reset
2. **Heavy work in the morning**: Use `-r 6` when fresh
3. **Recovery in afternoon**: Let auto-recovery kick in, or manually use `-R 2`
4. **Track in markdown**: Use `movodoro report --md >> workout-log.md` to append to your training log
5. **Tag your equipment**: Filter by what you have available (`-t kbx` when at home gym)

## Development

### Running Tests

```bash
go test -v
```

Tests use isolated fixtures in `testdata/movos/` and don't touch your live history.

### Project Structure

```
movodoro/
├── main.go              # CLI entry point
├── commands.go          # Command handlers
├── types.go             # Data structures
├── loader.go            # YAML loading
├── history.go           # Daily log management
├── selector.go          # Selection algorithm
├── config.go            # Configuration
├── movodoro_test.go     # Tests
├── testdata/            # Test fixtures
│   └── movos/
└── movos-examples/      # Example movement snacks
    ├── breathing.yaml
    ├── mobility.yaml
    └── bodyweight-strength.yaml
```

## Contributing

This is a personal project, but feel free to fork and adapt to your needs!

## License

[Your license here]
