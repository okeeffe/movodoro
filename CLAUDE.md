# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What is Movodoro?

Movodoro is a CLI tool for generating and tracking movement snacks (called "movos") during work breaks. It helps users incorporate 3-10 minute movement patterns throughout their day with intelligent selection based on recency, RPE (Rate of Perceived Exertion), and daily minimums.

## Build and Test Commands

```bash
# Build the project
go build -o movodoro

# Run tests
go test -v

# Run specific test
go test -v -run TestName

# Run the binary directly
./movodoro help
```

## Environment Configuration

The tool requires a movos directory containing YAML files that define movement patterns:

```bash
# Set custom movos directory (recommended)
export MOVODORO_MOVOS_DIR=/path/to/your/movos

# Set active subset (optional - for injury recovery, travel, etc.)
export MOVODORO_ACTIVE_SUBSET=back-safe

# Or use default location
mkdir -p ~/.movodoro/movos

# Verify configuration
./movodoro config
```

## Architecture Overview

### Core Selection Algorithm (selector.go)

The selection system uses a **priority-based weighted random** approach with a multi-stage filtering pipeline:

1. **Basic Filters** (category, tags, duration, RPE) - applied first
2. **Subset Filter** (if active) - restricts to specific movo codes
3. **Daily Minimums Priority**: Snacks with `min_per_day` > 0 that haven't been completed the required number of times are prioritized exclusively until met
4. **Frequency Filtering**: Snacks at their `max_per_day` or `max_per_week` limit are excluded
5. **Auto-recovery Mode**: When daily cumulative RPE reaches 30, automatically limits to RPE ≤ 2
6. **Weight Boosts**:
   - 10x boost for incomplete `min_per_day` snacks
   - 3x boost for never-completed snacks
   - 2x boost for snacks not done in 7+ days

**Important**: Subset filtering happens BEFORE min_per_day priority, so dailies are still prioritized but only those within the active subset.

### Data Storage (history.go)

Uses **daily log files** instead of a single monolithic file:
- Each day gets its own file: `~/.movodoro/logs/YYYYMMDD.log`
- Format: `TIMESTAMP CODE STATUS DURATION RPE` (space-separated)
- Enables fast today-focused operations and easy cleanup
- All "today" operations (`GetTodayStatsDaily`, `GetCountTodayDaily`) only read current day's file

### YAML Loading (loader.go)

Snacks are organized by category in separate YAML files:
- Each file defines one category with multiple snacks
- Category fields apply defaults to all snacks (weight, default_rpe, tags)
- Full snack code format: `{CATEGORY_CODE}-{snack-code}` (e.g., `RB-box-breathing`)
- Tags are combined: category tags + snack-specific tags → `AllTags`

**Subsets Configuration** (optional):
- Subsets are defined in `subsets.yaml` in the same directory as movo YAML files
- `LoadSubsets()` returns empty config (not error) if file doesn't exist
- Each subset contains a description and array of full movo codes
- Activated via `MOVODORO_ACTIVE_SUBSET` env var or `--subset` flag

### Interactive vs Command Mode (commands.go)

Two operational modes:
1. **Interactive Mode** (default when running `movodoro`): Single-key input loop with resume capability
2. **Command Mode**: Direct commands for scripting (`movodoro get`, `movodoro done`, etc.)

The interactive mode saves current snack to `~/.movodoro/current` for session persistence.

## Key Concepts

### Subsets for Situational Filtering

Subsets restrict movo selection to a predefined list of codes. Key behaviors:
- **Intersection with filters**: Subset + tag filter = only movos in both
- **Respects min_per_day**: Dailies within subset still prioritized
- **Precedence**: Command flag (`--subset`) > env var (`MOVODORO_ACTIVE_SUBSET`)
- **Commands aware**: `everyday` and `config` commands show subset status
- **File location**: Must be in movos directory (same as YAML files)

Use cases: injury recovery (`back-safe`), travel (`bodyweight-only`), equipment constraints.

### "Movos" not "Snacks"

Recent refactor renamed "snacks" to "movos" throughout:
- Type name: `Movo` (not `Snack`)
- Variable names: `movo`, `movos` (not `snack`, `snacks`)
- Display functions: `displayMovo*` (not `displaySnack*`)
- User-facing output already uses "movos" terminology

### Min Per Day vs Max Per Day

- `min_per_day`: Creates **priority requirement** (snack is prioritized until completed N times)
- `max_per_day`: Creates **frequency limit** (snack is filtered out after N completions)
- Example: `min_per_day: 1, max_per_day: 2` means "do at least once (priority), at most twice (limit)"

### RPE-Based Load Management

RPE (Rate of Perceived Exertion) 1-10 scale:
- Used for tracking cumulative daily load
- Auto-recovery at threshold (default: 30 total RPE)
- Users can override actual RPE when marking done

## File Structure

```
main.go         - Entry point, command routing
commands.go     - All command handlers (get, done, skip, report, interactive)
types.go        - Core data structures (Movo, Category, HistoryEntry, etc.)
selector.go     - Selection algorithm with priority/weighting logic
loader.go       - YAML parsing and snack loading
history.go      - Daily log file management
config.go       - Configuration (paths, defaults)
*_test.go       - Tests use testdata/movos/ fixtures
```

## Common Tasks

### Adding a New Command

1. Add case to switch in `main.go`
2. Implement `handle{Command}()` function in `commands.go`
3. Update `printUsage()` in `main.go`

### Modifying Selection Logic

Selection happens in `selector.go:SelectSnack()`:
1. Check for auto-recovery mode (may override max RPE)
2. Apply basic filters (category, tags, duration, RPE) via `filterSnacks()`
3. Apply subset filter (if active) via `filterBySubset()`
4. Priority filtering for incomplete minimums (unless `SkipMinimums` flag set)
5. Frequency filtering (max_per_day) via `filterByFrequency()`
6. Weight calculation with boosts via `calculateWeight()`
7. Weighted random selection via `weightedRandomSelect()`

**Adding a new filter**: Insert between steps 2-3 (after basic filters, before subset) or step 3-4 (after subset, before min_per_day priority) depending on desired interaction with subsets.

### Working with History

All history operations should use the daily log functions:
- `LoadDailyLog()` for single day
- `LoadHistoryRange()` for date ranges
- `LoadAllHistory()` for full history (use sparingly)
- `AppendTodayLog()` for new entries

## Testing Notes

- Tests use `testdata/movos/` directory with fixture YAML files
- **Important**: `testdata/movos/subsets.yaml` exists for subset tests - don't confuse with production subsets
- Use `TestConfig()` helper for isolated test environments, but note it uses `testdata/test-movos` which doesn't exist - set `MOVODORO_MOVOS_DIR=testdata/movos` manually in tests instead
- Daily log tests should clean up created files
- Selection tests may need multiple runs due to randomness (see `*_analysis_test.go`)

### Subset Testing Pattern

When testing features that interact with subsets:
1. Set `MOVODORO_MOVOS_DIR` env var to `testdata/movos`
2. Call `LoadSubsets("testdata/movos")` to get test subsets config
3. Test subsets available: `recovery`, `breath-only`, `high-intensity`, `single-item`, `empty`
4. Test codes: `TB-box-breath`, `TB-deep-breath`, `TS-pushups`, `TS-heavy-lift`, `TS-light-move`
