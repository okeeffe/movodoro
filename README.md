# Movodoro

A CLI tool for generating and tracking movement snacks (a.k.a movos) during work breaks. Stay active, prevent burnout, and build movement variety into your day with intelligent snack selection and RPE-based load management.

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

### Interactive Mode (Default)

Simply run `movodoro` to enter the interactive flow:

```bash
$ movodoro

═══════════════════════════════════════
  Box breathing
═══════════════════════════════════════

Standing or kneeling breath work...

⏱️  Duration: 3-5 minutes
💪 RPE: 1/10
🏷️  Code: RB-box-breathing

What would you like to do?
  [d] Done (log completion)
  [s] Skip (try another movo)
  [q] Quit (save for later)

  (Press 'h' for help: movodoro --help)

Choice: d
How many minutes did you spend? (default: 4): 5

✅ Marked 'Box breathing' as completed (5 minutes, RPE 1)
📊 Today: 1 snacks, 5 minutes, 1 RPE
```

**The Flow:**
- 🎯 **[d] Done** - Log completion, prompted for duration, then exit
- ⏭️ **[s] Skip** - Log skip, get another snack (stays in interactive mode)
- 🚪 **[q] Quit** - Save current snack, exit (can run `movodoro done` later)
- ❌ **[x] Skip dailies** - Only shown for everyday snacks, gets non-daily snack

**Ctrl+C** works as expected (same as quit).

### Command Line Mode

All individual commands still work for scripting/automation:

```bash
movodoro get                # Get a snack (non-interactive)
movodoro done               # Mark current snack as done
movodoro skip               # Skip current snack
movodoro report             # View today's report
movodoro everyday           # Check daily essential snacks
movodoro subsets            # List available subsets
movodoro config             # Show configuration
```

## Creating Your Movement Snacks (Movos)

If you'd like to see mine, for reference, you can see them [here](https://github.com/okeeffe/movodoro-movos).

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

movos:
  - code: box-breathing
    title: 4-7-8 or box breathing
    description: |
      Standing or kneeling breath work. 4-7-8: inhale 4, hold 7, exhale 8.
      Box: 4-4-4-4 pattern.
    duration_min: 3
    duration_max: 5
    rpe: 1              # Override default_rpe if needed
    max_per_day: 2      # Can be done twice per day
    min_per_day: 1      # Prioritized daily (shown first until complete)
    weight: 1.5         # Snack-specific weight multiplier
    tags: []            # Additional tags beyond category tags

  - code: deep-breath
    title: Deep belly breathing
    description: Focus on diaphragmatic breathing.
    duration_min: 2
    duration_max: 4
    # rpe: inherits default_rpe (1)
    max_per_day: 1
    weight: 1.0
    # min_per_day: omit for non-daily snacks (defaults to 0)
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
- **min_per_day**: Minimum times per day (e.g., 1, 2), **prioritized daily** until completed this many times
- **weight**: Snack-specific weight multiplier
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

## Subsets: Restricting Movement Selection

Subsets allow you to restrict movement selection to a specific collection of movos. Perfect for injury recovery, travel, equipment constraints, or seasonal training phases.

### Creating Subsets

Create a `subsets.yaml` file in your movos directory:

```yaml
# $MOVODORO_MOVOS_DIR/subsets.yaml
subsets:
  back-safe:
    description: "Safe movements during back injury recovery"
    codes:
      - BS-meditation
      - BS-box-breathing
      - OS-resets
      - OS-shoulder-rolls
      - BS-singleleg-reach

  travel:
    description: "Bodyweight-only movements requiring no equipment"
    codes:
      - BS-meditation
      - OS-resets
      - BS-singleleg-reach
      - MI-shadowbox

  recovery:
    description: "Very light movements (RPE ≤ 2) for active recovery"
    codes:
      - BS-meditation
      - BS-box-breathing
      - OS-resets
```

### Using Subsets

**Three ways to activate a subset:**

1. **Per-command** (one-time use):
   ```bash
   movodoro get --subset back-safe
   ```

2. **Interactive mode**:
   ```bash
   movodoro --subset back-safe
   ```

3. **Environment variable** (persistent):
   ```bash
   export MOVODORO_ACTIVE_SUBSET=back-safe
   ```

Command flags take precedence over environment variables.

### Subset Behavior

- **Filters are intersections**: Subset + other filters (tags, RPE) = only movos matching ALL criteria
- **Respects dailies**: Daily minimums (`min_per_day`) still prioritized, but only those within the subset
- **View available subsets**: Run `movodoro subsets` to see all configured subsets
- **Check affected dailies**: Run `movodoro everyday` to see which daily movos are excluded by active subset

**Example:**
```bash
# Set back-safe subset
export MOVODORO_ACTIVE_SUBSET=back-safe

# Check which everyday movos are available
movodoro everyday
# Output:
#   ✅ Meditation (in subset)
#   ✅ OS Resets (in subset)
#   ⚠️  2 everyday movos excluded by active subset

# Interactive mode will only select from back-safe movos
movodoro
```

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
- `--subset NAME` - Use a named subset from subsets.yaml

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
- `-v, --verbose` - Show titles and tags (perfect for workout journals)

**Examples:**
```bash
movodoro report                  # Today's report (codes only)
movodoro report day              # Same as above
movodoro report --md             # Markdown format
movodoro report -v               # Verbose with titles and tags
movodoro report --md -v          # Verbose markdown (best for logs)
movodoro report --md -v >> log.md  # Append to workout journal
```

**Verbose Output Example:**
```markdown
- **12:19** - OS Resets [`OS-resets`] (7 min, RPE 3) | #mobilityx, #flowx, #crawlx
- **13:10** - Freeform squats [`MS-freeform-squats`] (6 min, RPE 6) | #strengthx, #squatx
- **13:59** - Meditation [`BS-meditation`] (10 min, RPE 1) | #mindfulnessx
```

Verbose mode adds:
- Human-readable titles for context
- Codes in square brackets for reference
- Hashtag tags (#kbx, #strengthx) for searchability

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

Shows all snacks with `min_per_day` requirements and their completion status for today. Great for tracking your daily movement essentials.

**Example output:**
```
═══════════════════════════════════════
  EVERY DAY MOVOS
═══════════════════════════════════════

✅ Box breathing
   Code: BR-box-breathing | RPE: 1 | Duration: 3-5 min
   Completed 2 of 2 today

❌ Hip circles and leg swings
   Code: MOB-hip-circles | RPE: 3 | Duration: 5-7 min
   Not yet done (0 of 1 today)

Summary: 1/2 everyday snacks completed
```

### Version

```bash
movodoro version
```

### List Available Subsets

```bash
movodoro subsets
```

Shows all subsets configured in `subsets.yaml` with their descriptions and movo counts.

## How Selection Works

Movodoro uses a priority-based selection system focused on daily minimums first:

### Daily Minimums Priority

**Key Concept:** Snacks with `min_per_day` set (e.g., `min_per_day: 1`) are prioritized until you complete them the required number of times each day.

**How it works:**
1. Check for incomplete minimums today (where times_done < min_per_day)
2. If any exist: Only select from those (weighted among themselves)
3. If all complete: Select from the full snack pool

**Flexible Minimums:**
- `min_per_day: 1` - Once daily (most common)
- `min_per_day: 2` - Twice daily (for critical patterns)
- `min_per_day: 0` or omitted - No daily requirement

**Example daily flow:**
```bash
# Morning - 2 snacks with min_per_day incomplete
movodoro          # → Meditation (min_per_day: 1, done 0 times)
# [s] Skip
movodoro          # → Meditation again (still incomplete)
# [x] Skip dailies
movodoro          # → Kettlebell swings (escaped to non-daily)
# Exit and come back later
movodoro          # → Meditation (back to incomplete minimums)
# [d] Done - mark complete
movodoro          # → OS Resets (min_per_day: 1, done 0 times)
# [d] Done - mark complete
movodoro          # → Now random from full pool
```

**After Meeting Minimum:**
Even after completing the minimum (e.g., 1/1), snacks with `max_per_day` > `min_per_day` can still be selected from the full pool.

Example with `min_per_day: 1, max_per_day: 2`:
- After 1st completion: No longer prioritized, but can appear in general pool
- After 2nd completion: Filtered out entirely for the day

### Escape Hatch: Skip Dailies

When viewing a snack with minimum requirements, press **[x] Skip dailies** to temporarily get a non-minimum snack. This:
- Does NOT log a skip
- Gets you one snack from the full pool (skipping min_per_day priority)
- Next request returns to minimum priority (if still incomplete)

Or use the `--skip-minimums` flag:
```bash
movodoro get --skip-minimums    # Bypass min_per_day priority
```

### Within-Category Weight System

Once the candidate pool is determined, weighted random selection applies:

**Base Weight:**
Each snack starts with its `weight` value (default: 1.0), multiplied by the category's `weight`.

**Boosts Applied:**
1. **Minimum boost (10x)**: Snacks with incomplete `min_per_day` (e.g., done 0 times when min_per_day: 1)
2. **Never-done boost (3x)**: Snacks you've never completed
3. **Recency boost (2x)**: Snacks not done in 7+ days

**Filters:**
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

1. **Set 2-3 movements with `min_per_day: 1`**: These become your daily non-negotiables (breath work, fundamental patterns, etc.)
2. **Use flexible minimums**: For critical patterns, use `min_per_day: 2` to ensure you do them twice daily
3. **Use interactive mode**: Just type `movodoro` and let it guide you through your daily minimums
4. **Skip dailies when pressed for time**: Use `[x]` to grab a quick 3-min snack when you can't do your 10-min meditation
5. **Check your progress**: Run `movodoro everyday` to see what you've completed
6. **Recovery in afternoon**: Let auto-recovery kick in when you hit RPE 30, or manually use `movodoro get -R 2`
7. **Create subsets for life situations**: Set up `back-safe`, `travel`, `low-energy` subsets for different contexts
8. **Track in markdown**: Use `movodoro report --md -v >> workout-log.md` to append detailed logs to your training journal
9. **Search by tags**: Use hashtags in verbose reports to quickly find workouts (search #kbx for kettlebell, #strengthx for strength training)
10. **Non-interactive for scripts**: Use `movodoro get -t kbx` for automation or specific filters

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

MIT License - see [LICENSE](LICENSE) for details.
