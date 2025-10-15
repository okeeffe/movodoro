package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/term"
)

var appConfig = DefaultConfig()

const (
	maxDailyRPEDefault = 30
)

// handleGet implements the 'get' command
func handleGet(args []string) {
	fs := flag.NewFlagSet("get", flag.ExitOnError)

	var (
		tags         string
		category     string
		duration     int
		minDuration  int
		maxDuration  int
		minRPE       int
		maxRPE       int
		skipMinimums bool
		subset       string
	)

	fs.StringVar(&tags, "tags", "", "Filter by tags (comma-separated)")
	fs.StringVar(&tags, "t", "", "Filter by tags (comma-separated)")
	fs.StringVar(&category, "category", "", "Filter by category code")
	fs.StringVar(&category, "c", "", "Filter by category code")
	fs.IntVar(&duration, "duration", 0, "Exact duration in minutes")
	fs.IntVar(&duration, "d", 0, "Exact duration in minutes")
	fs.IntVar(&minDuration, "min-duration", 0, "Minimum duration")
	fs.IntVar(&minDuration, "m", 0, "Minimum duration")
	fs.IntVar(&maxDuration, "max-duration", 0, "Maximum duration")
	fs.IntVar(&maxDuration, "M", 0, "Maximum duration")
	fs.IntVar(&minRPE, "min-rpe", 0, "Minimum RPE")
	fs.IntVar(&minRPE, "r", 0, "Minimum RPE")
	fs.IntVar(&maxRPE, "max-rpe", 0, "Maximum RPE")
	fs.IntVar(&maxRPE, "R", 0, "Maximum RPE")
	fs.BoolVar(&skipMinimums, "skip-minimums", false, "Skip min_per_day priority")
	fs.StringVar(&subset, "subset", "", "Use a named subset from subsets.yaml")

	fs.Parse(args)

	// Load snacks
	snacks, err := LoadSnacks()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading snacks: %v\n", err)
		os.Exit(1)
	}

	// Determine active subset: command flag takes precedence over env var
	activeSubset := subset
	if activeSubset == "" {
		activeSubset = appConfig.ActiveSubset
	}

	// Parse filters
	filters := FilterOptions{
		Category:      strings.TrimSpace(strings.ToUpper(category)),
		MinDuration:   minDuration,
		MaxDuration:   maxDuration,
		ExactDuration: duration,
		MinRPE:        minRPE,
		MaxRPE:        maxRPE,
		SkipMinimums:  skipMinimums,
		Subset:        activeSubset,
	}

	if tags != "" {
		filters.Tags = strings.Split(tags, ",")
		// Trim whitespace
		for i := range filters.Tags {
			filters.Tags[i] = strings.TrimSpace(filters.Tags[i])
		}
	}

	// Select a snack
	snack, err := SelectSnack(snacks, filters, maxDailyRPEDefault)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error selecting snack: %v\n", err)
		os.Exit(1)
	}

	// Save as current snack
	if err := saveCurrentSnack(snack.FullCode); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not save current snack: %v\n", err)
	}

	// Display the movo
	displayMovo(snack)
}

// handleDone implements the 'done' command
func handleDone(args []string) {
	var code string

	// Check if code was provided as argument
	if len(args) > 0 {
		code = args[0]
	} else {
		// Use current snack
		var err error
		code, err = loadCurrentSnack()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: no current snack. Use 'movodoro get' first or specify a code.\n")
			os.Exit(1)
		}
	}

	// Load snacks to get RPE
	snacks, err := LoadSnacks()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading snacks: %v\n", err)
		os.Exit(1)
	}

	// Find the snack
	var snack *Movo
	for _, s := range snacks {
		if s.FullCode == code {
			snack = &s
			break
		}
	}

	if snack == nil {
		fmt.Fprintf(os.Stderr, "Error: snack code '%s' not found\n", code)
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)

	// Prompt for actual duration
	defaultDuration := snack.GetDefaultDuration()
	fmt.Printf("How many minutes did you spend? (default: %d): ", defaultDuration)

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	duration := defaultDuration
	if input != "" {
		parsed, err := strconv.Atoi(input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid duration, using default: %d\n", defaultDuration)
		} else {
			duration = parsed
		}
	}

	// Prompt for RPE
	defaultRPE := snack.EffectiveRPE
	fmt.Printf("How hard was it? RPE (default: %d): ", defaultRPE)

	input, _ = reader.ReadString('\n')
	input = strings.TrimSpace(input)

	rpe := defaultRPE
	if input != "" {
		parsed, err := strconv.Atoi(input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid RPE, using default: %d\n", defaultRPE)
		} else {
			rpe = parsed
		}
	}

	// Create history entry
	entry := HistoryEntry{
		Timestamp: time.Now(),
		Code:      code,
		Status:    "done",
		Duration:  duration,
		RPE:       rpe,
		Subset:    appConfig.ActiveSubset,
	}

	// Save to history
	if err := AppendTodayLog(appConfig.LogsDir, entry); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving to history: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Marked '%s' as completed (%d minutes, RPE %d)\n", snack.Title, duration, rpe)

	// Show updated daily stats
	stats, _ := GetTodayStatsDaily(appConfig.LogsDir)
	fmt.Printf("📊 Today: %d movos, %d minutes, %d RPE\n", stats.TotalMovos, stats.TotalDuration, stats.TotalRPE)
}

// handleSkip implements the 'skip' command
func handleSkip(args []string) {
	var code string

	// Check if code was provided as argument
	if len(args) > 0 {
		code = args[0]
	} else {
		// Use current snack
		var err error
		code, err = loadCurrentSnack()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: no current snack. Use 'movodoro get' first or specify a code.\n")
			os.Exit(1)
		}
	}

	// Load snacks to verify code exists
	snacks, err := LoadSnacks()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading snacks: %v\n", err)
		os.Exit(1)
	}

	// Find the snack
	var snack *Movo
	for _, s := range snacks {
		if s.FullCode == code {
			snack = &s
			break
		}
	}

	if snack == nil {
		fmt.Fprintf(os.Stderr, "Error: snack code '%s' not found\n", code)
		os.Exit(1)
	}

	// Create history entry with 0 duration and RPE
	entry := HistoryEntry{
		Timestamp: time.Now(),
		Code:      code,
		Status:    "skip",
		Duration:  0,
		RPE:       0,
		Subset:    appConfig.ActiveSubset,
	}

	// Save to history
	if err := AppendTodayLog(appConfig.LogsDir, entry); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving to history: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("⏭️  Skipped '%s'\n", snack.Title)
}

// handleReport implements the 'report' command
func handleReport(args []string) {
	fs := flag.NewFlagSet("report", flag.ExitOnError)
	var markdown bool
	var verbose bool
	fs.BoolVar(&markdown, "markdown", false, "Output in markdown format")
	fs.BoolVar(&markdown, "md", false, "Output in markdown format")
	fs.BoolVar(&verbose, "verbose", false, "Show titles and tags (great for workout logs)")
	fs.BoolVar(&verbose, "v", false, "Show titles and tags (great for workout logs)")

	fs.Parse(args)

	remaining := fs.Args()
	period := "day"
	if len(remaining) > 0 {
		period = remaining[0]
	}

	switch period {
	case "day", "today":
		if markdown {
			showDayReportMarkdown(verbose)
		} else {
			showDayReport(verbose)
		}
	case "week":
		fmt.Println("Week report - not yet implemented")
	case "month":
		fmt.Println("Month report - not yet implemented")
	default:
		fmt.Fprintf(os.Stderr, "Unknown report period: %s (use: day, week, month)\n", period)
		os.Exit(1)
	}
}

func showDayReport(verbose bool) {
	stats, err := GetTodayStatsDaily(appConfig.LogsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading stats: %v\n", err)
		os.Exit(1)
	}

	// Load snacks for verbose mode
	var movoMap map[string]*Movo
	if verbose {
		snacks, err := LoadSnacks()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading snacks: %v\n", err)
			os.Exit(1)
		}
		movoMap = make(map[string]*Movo)
		for i := range snacks {
			movoMap[snacks[i].FullCode] = &snacks[i]
		}
	}

	// Detect active subset(s) from entries
	subsetMap := make(map[string]bool)
	for _, entry := range stats.CompletedSnacks {
		if entry.Subset != "" {
			subsetMap[entry.Subset] = true
		}
	}
	for _, entry := range stats.SkippedSnacks {
		if entry.Subset != "" {
			subsetMap[entry.Subset] = true
		}
	}

	var subsets []string
	for subset := range subsetMap {
		subsets = append(subsets, subset)
	}

	fmt.Println("═══════════════════════════════════════")
	fmt.Printf("  TODAY'S MOVODORO REPORT\n")
	fmt.Printf("  %s\n", stats.Date.Format("Monday, January 2, 2006"))
	fmt.Println("═══════════════════════════════════════")
	fmt.Println()

	if len(subsets) > 0 {
		fmt.Println("Active subset(s):")
		for _, subset := range subsets {
			fmt.Printf("  - %s\n", subset)
		}
		fmt.Println()
	}

	fmt.Printf("📊 Summary:\n")
	fmt.Printf("   Total movos:     %d\n", len(stats.CompletedSnacks))
	fmt.Printf("   Total duration:  %d minutes\n", stats.TotalDuration)
	fmt.Printf("   Total RPE:       %d / %d\n", stats.TotalRPE, maxDailyRPEDefault)
	fmt.Println()

	if len(stats.CompletedSnacks) > 0 {
		fmt.Printf("✅ Completed:\n")
		for _, entry := range stats.CompletedSnacks {
			subsetStr := ""
			if entry.Subset != "" {
				subsetStr = ", " + entry.Subset
			}

			if verbose {
				movo := movoMap[entry.Code]
				if movo != nil {
					tagsStr := formatMovoTags(movo)
					fmt.Printf("   %s - %s [%s] (%dm, RPE %d%s)%s\n",
						entry.Timestamp.Format("15:04"),
						movo.Title,
						entry.Code,
						entry.Duration,
						entry.RPE,
						subsetStr,
						tagsStr)
				} else {
					// Fallback if snack not found
					fmt.Printf("   %s - %s (%dm, RPE %d%s)\n",
						entry.Timestamp.Format("15:04"),
						entry.Code,
						entry.Duration,
						entry.RPE,
						subsetStr)
				}
			} else {
				fmt.Printf("   %s - %s (%dm, RPE %d%s)\n",
					entry.Timestamp.Format("15:04"),
					entry.Code,
					entry.Duration,
					entry.RPE,
					subsetStr)
			}
		}
		fmt.Println()
	}

	if len(stats.SkippedSnacks) > 0 {
		fmt.Printf("⏭️  Skipped:\n")
		for _, entry := range stats.SkippedSnacks {
			if verbose {
				movo := movoMap[entry.Code]
				if movo != nil {
					fmt.Printf("   %s - %s [%s]\n",
						entry.Timestamp.Format("15:04"),
						movo.Title,
						entry.Code)
				} else {
					fmt.Printf("   %s - %s\n",
						entry.Timestamp.Format("15:04"),
						entry.Code)
				}
			} else {
				fmt.Printf("   %s - %s\n",
					entry.Timestamp.Format("15:04"),
					entry.Code)
			}
		}
		fmt.Println()
	}

	if stats.TotalRPE >= maxDailyRPEDefault {
		fmt.Println("🔋 Auto-recovery mode active (RPE limit reached)")
	}
}

func showDayReportMarkdown(verbose bool) {
	stats, err := GetTodayStatsDaily(appConfig.LogsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading stats: %v\n", err)
		os.Exit(1)
	}

	// Load snacks for verbose mode
	var movoMap map[string]*Movo
	if verbose {
		snacks, err := LoadSnacks()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading snacks: %v\n", err)
			os.Exit(1)
		}
		movoMap = make(map[string]*Movo)
		for i := range snacks {
			movoMap[snacks[i].FullCode] = &snacks[i]
		}
	}

	// Detect active subset(s) from entries
	subsetMap := make(map[string]bool)
	for _, entry := range stats.CompletedSnacks {
		if entry.Subset != "" {
			subsetMap[entry.Subset] = true
		}
	}
	for _, entry := range stats.SkippedSnacks {
		if entry.Subset != "" {
			subsetMap[entry.Subset] = true
		}
	}

	var subsets []string
	for subset := range subsetMap {
		subsets = append(subsets, subset)
	}

	fmt.Printf("# Movodoro Report - %s\n\n", stats.Date.Format("Monday, January 2, 2006"))

	if len(subsets) > 0 {
		fmt.Println("**Active subset(s):**")
		fmt.Println()
		for _, subset := range subsets {
			fmt.Printf("- %s\n", subset)
		}
		fmt.Println()
	}

	fmt.Println("## Summary")
	fmt.Println()
	fmt.Printf("- **Total movos:** %d\n", len(stats.CompletedSnacks))
	fmt.Printf("- **Total duration:** %d minutes\n", stats.TotalDuration)
	fmt.Printf("- **Total RPE:** %d / %d\n", stats.TotalRPE, maxDailyRPEDefault)
	fmt.Println()

	if len(stats.CompletedSnacks) > 0 {
		fmt.Println("## Completed")
		fmt.Println()
		for _, entry := range stats.CompletedSnacks {
			subsetStr := ""
			if entry.Subset != "" {
				subsetStr = ", " + entry.Subset
			}

			if verbose {
				movo := movoMap[entry.Code]
				if movo != nil {
					tagsStr := formatMovoTags(movo)
					fmt.Printf("- **%s** - %s [`%s`] (%d min, RPE %d%s)%s\n",
						entry.Timestamp.Format("15:04"),
						movo.Title,
						entry.Code,
						entry.Duration,
						entry.RPE,
						subsetStr,
						tagsStr)
				} else {
					// Fallback if snack not found
					fmt.Printf("- **%s** - `%s` (%d min, RPE %d%s)\n",
						entry.Timestamp.Format("15:04"),
						entry.Code,
						entry.Duration,
						entry.RPE,
						subsetStr)
				}
			} else {
				fmt.Printf("- **%s** - `%s` (%d min, RPE %d%s)\n",
					entry.Timestamp.Format("15:04"),
					entry.Code,
					entry.Duration,
					entry.RPE,
					subsetStr)
			}
		}
		fmt.Println()
	}

	if len(stats.SkippedSnacks) > 0 {
		fmt.Println("## Skipped")
		fmt.Println()
		for _, entry := range stats.SkippedSnacks {
			if verbose {
				movo := movoMap[entry.Code]
				if movo != nil {
					fmt.Printf("- **%s** - %s [`%s`]\n",
						entry.Timestamp.Format("15:04"),
						movo.Title,
						entry.Code)
				} else {
					fmt.Printf("- **%s** - `%s`\n",
						entry.Timestamp.Format("15:04"),
						entry.Code)
				}
			} else {
				fmt.Printf("- **%s** - `%s`\n",
					entry.Timestamp.Format("15:04"),
					entry.Code)
			}
		}
		fmt.Println()
	}

	if stats.TotalRPE >= maxDailyRPEDefault {
		fmt.Println("*Auto-recovery mode active (RPE limit reached)*")
	}
}

// formatMovoTags formats tags for verbose report output
func formatMovoTags(movo *Movo) string {
	if len(movo.AllTags) == 0 && movo.MinPerDay == 0 {
		return ""
	}

	// Format tags with # prefix
	tagList := make([]string, len(movo.AllTags))
	for i, tag := range movo.AllTags {
		tagList[i] = "#" + tag
	}

	// Add #daily tag if this is an everyday movo
	if movo.MinPerDay > 0 {
		tagList = append(tagList, "#daily")
	}

	return fmt.Sprintf(" | %s", strings.Join(tagList, ", "))
}

func displayMovo(movo *Movo) {
	fmt.Println()
	fmt.Println("═══════════════════════════════════════")
	fmt.Printf("  %s\n", movo.Title)
	fmt.Println("═══════════════════════════════════════")
	fmt.Println()

	fmt.Println(movo.Description)
	fmt.Println()

	fmt.Printf("⏱️  Duration: %d-%d minutes\n", movo.DurationMin, movo.DurationMax)
	fmt.Printf("💪 RPE: %d/10\n", movo.EffectiveRPE)
	fmt.Printf("🏷️  Code: %s\n", movo.FullCode)

	if len(movo.AllTags) > 0 {
		fmt.Printf("🔖 Tags: %s\n", strings.Join(movo.AllTags, ", "))
	}

	fmt.Println()
	fmt.Println("When done, run:")
	fmt.Printf("  movodoro done\n")
	fmt.Println("Or skip with:")
	fmt.Printf("  movodoro skip\n")
	fmt.Println()
}

// saveCurrentSnack saves the current snack code to a file
func saveCurrentSnack(code string) error {
	return os.WriteFile(appConfig.CurrentPath, []byte(code), 0644)
}

// loadCurrentSnack loads the current snack code from file
func loadCurrentSnack() (string, error) {
	data, err := os.ReadFile(appConfig.CurrentPath)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(data)), nil
}

// handleClear implements the 'clear' command
func handleClear(args []string) {
	// Get today's stats first
	stats, err := GetTodayStatsDaily(appConfig.LogsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading stats: %v\n", err)
		os.Exit(1)
	}

	// Show what will be cleared
	fmt.Println("═══════════════════════════════════════")
	fmt.Println("  CLEAR TODAY'S HISTORY")
	fmt.Println("═══════════════════════════════════════")
	fmt.Println()

	if stats.TotalMovos == 0 {
		fmt.Println("No entries for today to clear.")
		return
	}

	fmt.Printf("This will delete today's log file with %d entries:\n", stats.TotalMovos)
	fmt.Printf("  - %d completed (%d minutes, %d RPE)\n",
		len(stats.CompletedSnacks), stats.TotalDuration, stats.TotalRPE)
	fmt.Printf("  - %d skipped\n", len(stats.SkippedSnacks))
	fmt.Println()

	// Prompt for confirmation
	fmt.Print("Are you sure you want to clear today's history? (yes/no): ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input != "yes" && input != "y" {
		fmt.Println("Cancelled.")
		return
	}

	// Delete today's log file
	if err := ClearTodayLog(appConfig.LogsDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error clearing today's log: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Cleared %d entries from today's history\n", stats.TotalMovos)
}

// handleConfig implements the 'config' command
func handleConfig(args []string) {
	cfg := appConfig

	fmt.Println("═══════════════════════════════════════")
	fmt.Println("  MOVODORO CONFIGURATION")
	fmt.Println("═══════════════════════════════════════")
	fmt.Println()
	fmt.Printf("Movos directory:  %s\n", cfg.MovosDir)
	fmt.Printf("Logs directory:   %s\n", cfg.LogsDir)
	fmt.Printf("Current file:     %s\n", cfg.CurrentPath)
	fmt.Printf("Max daily RPE:    %d\n", cfg.MaxDailyRPE)
	if cfg.ActiveSubset != "" {
		fmt.Printf("Active subset:    %s\n", cfg.ActiveSubset)
	}
	fmt.Println()

	// Check if movos directory exists
	if _, err := os.Stat(cfg.MovosDir); os.IsNotExist(err) {
		fmt.Printf("⚠️  Movos directory does not exist: %s\n", cfg.MovosDir)
		fmt.Println()
		fmt.Println("To set a custom movos directory, use:")
		fmt.Println("  export MOVODORO_MOVOS_DIR=/path/to/your/movos")
	} else {
		// Count snacks
		snacks, err := LoadSnacks()
		if err != nil {
			fmt.Printf("⚠️  Error loading snacks: %v\n", err)
		} else {
			fmt.Printf("✅ Found %d movement snacks\n", len(snacks))
		}
	}
	fmt.Println()
}

// handleEveryday implements the 'everyday' command
func handleEveryday(args []string) {
	cfg := appConfig

	// Load snacks
	snacks, err := LoadSnacks()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading snacks: %v\n", err)
		os.Exit(1)
	}

	// Filter to only snacks with min_per_day requirement
	var everydayMovos []Movo
	for _, snack := range snacks {
		if snack.MinPerDay > 0 {
			everydayMovos = append(everydayMovos, snack)
		}
	}

	if len(everydayMovos) == 0 {
		fmt.Println("No movos with min_per_day requirement")
		return
	}

	// Check for active subset
	activeSubset := cfg.ActiveSubset
	var subsetCodes map[string]bool
	if activeSubset != "" {
		subsetsConfig, err := LoadSubsets(cfg.MovosDir)
		if err == nil {
			if subset, exists := subsetsConfig.Subsets[activeSubset]; exists {
				subsetCodes = make(map[string]bool)
				for _, code := range subset.Codes {
					subsetCodes[code] = true
				}
			}
		}
	}

	fmt.Println("═══════════════════════════════════════")
	fmt.Println("  EVERY DAY MOVOS")
	if activeSubset != "" {
		fmt.Printf("  (Subset: %s)\n", activeSubset)
	}
	fmt.Println("═══════════════════════════════════════")
	fmt.Println()

	// Get today's stats
	stats, err := GetTodayStatsDaily(appConfig.LogsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading today's stats: %v\n", err)
		os.Exit(1)
	}

	// Create map of completed snacks today
	completedToday := make(map[string]int)
	for _, entry := range stats.CompletedSnacks {
		completedToday[entry.Code]++
	}

	// Display each everyday snack
	inSubsetCount := 0
	excludedCount := 0
	for _, snack := range everydayMovos {
		// Check if excluded by subset
		inSubset := subsetCodes == nil || subsetCodes[snack.FullCode]
		if !inSubset {
			excludedCount++
			continue
		}
		inSubsetCount++

		count := completedToday[snack.FullCode]
		status := "❌"
		if count >= snack.MinPerDay {
			status = "✅"
		}

		fmt.Printf("%s %s\n", status, snack.Title)
		fmt.Printf("   Code: %s | RPE: %d | Duration: %d-%d min\n",
			snack.FullCode, snack.EffectiveRPE, snack.DurationMin, snack.DurationMax)

		if count > 0 {
			fmt.Printf("   Completed %d of %d today\n", count, snack.MinPerDay)
		} else {
			fmt.Printf("   Not yet done (0 of %d today)\n", snack.MinPerDay)
		}
		fmt.Println()
	}

	if excludedCount > 0 {
		fmt.Printf("⚠️  %d everyday movos excluded by active subset\n", excludedCount)
		fmt.Println()
	}

	completed := 0
	for _, snack := range everydayMovos {
		// Only count if in subset
		inSubset := subsetCodes == nil || subsetCodes[snack.FullCode]
		if inSubset && completedToday[snack.FullCode] >= snack.MinPerDay {
			completed++
		}
	}

	fmt.Printf("Summary: %d/%d everyday movos completed", completed, inSubsetCount)
	if activeSubset != "" {
		fmt.Printf(" (in subset)")
	}
	fmt.Println()
}

// handleInteractive implements the interactive mode (default when running `movodoro`)
func handleInteractive(args []string) {
	// Parse flags for interactive mode
	fs := flag.NewFlagSet("interactive", flag.ExitOnError)
	var subset string
	fs.StringVar(&subset, "subset", "", "Use a named subset from subsets.yaml")
	fs.Parse(args)

	// Load snacks
	snacks, err := LoadSnacks()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading snacks: %v\n", err)
		os.Exit(1)
	}

	// Determine active subset: command flag takes precedence over env var
	activeSubset := subset
	if activeSubset == "" {
		activeSubset = appConfig.ActiveSubset
	}

	// Display subset info if active
	if activeSubset != "" {
		fmt.Printf("🎯 Using subset: %s\n\n", activeSubset)
	}

	// Start with default filters
	filters := FilterOptions{
		Subset: activeSubset,
	}

	for {
		var snack *Movo

		// Try to load saved snack from previous session
		savedCode, err := loadCurrentSnack()
		if err == nil && savedCode != "" {
			// Find the saved snack
			for i := range snacks {
				if snacks[i].FullCode == savedCode {
					snack = &snacks[i]
					fmt.Println("📥 Resuming saved snack...")
					fmt.Println()
					break
				}
			}
		}

		// If no saved snack or couldn't find it, select a new one
		if snack == nil {
			selected, err := SelectSnack(snacks, filters, maxDailyRPEDefault)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error selecting snack: %v\n", err)
				os.Exit(1)
			}
			snack = selected
		}

		// Save as current snack (overwrites existing or saves new)
		if err := saveCurrentSnack(snack.FullCode); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not save current snack: %v\n", err)
		}

		// Display the movo
		displayMovoInteractive(snack)

		// Get user choice
		hasMinimum := snack.MinPerDay > 0
		choice := getInteractiveChoice(hasMinimum)

		switch choice {
		case "d": // Done
			handleDoneInteractive(snack)
			os.Remove(appConfig.CurrentPath) // Clear saved snack
			return                           // Exit after marking done

		case "s": // Skip
			handleSkipInteractive(snack)
			os.Remove(appConfig.CurrentPath) // Clear saved snack
			filters.SkipMinimums = false     // Reset skip minimums flag
			// Continue loop to get next snack

		case "x": // Skip dailies (only if snack has min_per_day)
			if snack.MinPerDay > 0 {
				fmt.Printf("\n⏭️  Skipping dailies for now...\n")
				os.Remove(appConfig.CurrentPath) // Clear saved snack
				filters.SkipMinimums = true
				// Continue loop to get next snack (will reset flag after)
			}

		case "q": // Quit
			fmt.Println("\n👋 Saved for later. Run 'movodoro' to resume.")
			return

		default:
			fmt.Println("Invalid choice, please try again.")
		}
	}
}

// displayMovoInteractive displays a movo in interactive mode
func displayMovoInteractive(movo *Movo) {
	fmt.Println()
	fmt.Println("═══════════════════════════════════════")
	fmt.Printf("  %s\n", movo.Title)
	fmt.Println("═══════════════════════════════════════")
	fmt.Println()

	fmt.Println(movo.Description)
	fmt.Println()

	fmt.Printf("⏱️  Duration: %d-%d minutes\n", movo.DurationMin, movo.DurationMax)
	fmt.Printf("💪 RPE: %d/10\n", movo.EffectiveRPE)
	fmt.Printf("🏷️  Code: %s\n", movo.FullCode)

	if len(movo.AllTags) > 0 {
		fmt.Printf("🔖 Tags: %s\n", strings.Join(movo.AllTags, ", "))
	}
	fmt.Println()
}

// getInteractiveChoice prompts user for action choice
func getInteractiveChoice(hasMinimum bool) string {
	fmt.Println("What would you like to do?")
	fmt.Println("  [d] Done (log completion)")
	fmt.Println("  [s] Skip (try another movo)")
	if hasMinimum {
		fmt.Println("  [x] Skip dailies (ignore min_per_day > 0 movos)")
	}
	fmt.Println("  [q] Quit (save for later)")
	fmt.Println("\n  (Press 'h' for help: movodoro --help)")
	fmt.Print("\nChoice: ")

	// Put terminal in raw mode for single-key input
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		// Fallback to regular input if terminal doesn't support raw mode
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		return strings.TrimSpace(strings.ToLower(input))
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	// Read single character
	buf := make([]byte, 1)
	for {
		_, err := os.Stdin.Read(buf)
		if err != nil {
			fmt.Println()
			return "q"
		}

		char := strings.ToLower(string(buf[0]))

		// Validate input
		validChars := []string{"d", "s", "q"}
		if hasMinimum {
			validChars = append(validChars, "x")
		}

		valid := false
		for _, v := range validChars {
			if char == v {
				valid = true
				break
			}
		}

		if valid {
			fmt.Println(char) // Echo the character
			return char
		}

		// Handle Ctrl+C (ASCII 3)
		if buf[0] == 3 {
			fmt.Println("^C")
			return "q"
		}

		// Invalid key - show error but keep prompt open
		fmt.Print("\r\033[KInvalid choice. Choice: ")
	}
}

// handleDoneInteractive handles completing a movo in interactive mode
func handleDoneInteractive(movo *Movo) {
	reader := bufio.NewReader(os.Stdin)

	// Prompt for actual duration
	defaultDuration := movo.GetDefaultDuration()
	fmt.Printf("\nHow many minutes did you spend? (default: %d): ", defaultDuration)

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	duration := defaultDuration
	if input != "" {
		parsed, err := strconv.Atoi(input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid duration, using default: %d\n", defaultDuration)
		} else {
			duration = parsed
		}
	}

	// Prompt for RPE
	defaultRPE := movo.EffectiveRPE
	fmt.Printf("How hard was it? RPE (default: %d): ", defaultRPE)

	input, _ = reader.ReadString('\n')
	input = strings.TrimSpace(input)

	rpe := defaultRPE
	if input != "" {
		parsed, err := strconv.Atoi(input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid RPE, using default: %d\n", defaultRPE)
		} else {
			rpe = parsed
		}
	}

	// Create history entry
	entry := HistoryEntry{
		Timestamp: time.Now(),
		Code:      movo.FullCode,
		Status:    "done",
		Duration:  duration,
		RPE:       rpe,
		Subset:    appConfig.ActiveSubset,
	}

	// Save to history
	if err := AppendTodayLog(appConfig.LogsDir, entry); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving to history: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✅ Marked '%s' as completed (%d minutes, RPE %d)\n", movo.Title, duration, rpe)

	// Show updated daily stats
	stats, _ := GetTodayStatsDaily(appConfig.LogsDir)
	fmt.Printf("📊 Today: %d movos, %d minutes, %d RPE\n\n", stats.TotalMovos, stats.TotalDuration, stats.TotalRPE)
}

// handleSkipInteractive handles skipping a movo in interactive mode
func handleSkipInteractive(movo *Movo) {
	// Create history entry with 0 duration and RPE
	entry := HistoryEntry{
		Timestamp: time.Now(),
		Code:      movo.FullCode,
		Status:    "skip",
		Duration:  0,
		RPE:       0,
		Subset:    appConfig.ActiveSubset,
	}

	// Save to history
	if err := AppendTodayLog(appConfig.LogsDir, entry); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving to history: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n⏭️  Skipped '%s'\n", movo.Title)
}

// handleSubsets implements the 'subsets' command
func handleSubsets(args []string) {
	cfg := appConfig

	// Load subsets configuration
	subsetsConfig, err := LoadSubsets(cfg.MovosDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading subsets: %v\n", err)
		os.Exit(1)
	}

	if len(subsetsConfig.Subsets) == 0 {
		fmt.Println("No subsets configured.")
		fmt.Println()
		fmt.Printf("Create a subsets.yaml file in your movos directory:\n")
		fmt.Printf("  %s/subsets.yaml\n", cfg.MovosDir)
		return
	}

	fmt.Println("═══════════════════════════════════════")
	fmt.Println("  AVAILABLE SUBSETS")
	fmt.Println("═══════════════════════════════════════")
	fmt.Println()

	// Display each subset
	for name, subset := range subsetsConfig.Subsets {
		fmt.Printf("📦 %s\n", name)
		if subset.Description != "" {
			fmt.Printf("   %s\n", subset.Description)
		}
		fmt.Printf("   %d movos\n", len(subset.Codes))
		fmt.Println()
	}

	fmt.Println("Usage:")
	fmt.Printf("  movodoro get --subset SUBSET_NAME\n")
	fmt.Printf("  movodoro --subset SUBSET_NAME          # Interactive mode\n")
	fmt.Printf("  export MOVODORO_ACTIVE_SUBSET=SUBSET_NAME\n")
}

// handleMigrateLogsToCsv implements the 'migrate-logs-to-csv' command
func handleMigrateLogsToCsv(args []string) {
	cfg := appConfig

	fmt.Println("═══════════════════════════════════════")
	fmt.Println("  MIGRATE LOGS TO CSV FORMAT (v1.0.0)")
	fmt.Println("═══════════════════════════════════════")
	fmt.Println()

	// Find all .log files
	pattern := filepath.Join(cfg.LogsDir, "*.log")
	files, err := filepath.Glob(pattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding log files: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Println("No log files found.")
		return
	}

	fmt.Printf("Found %d log file(s) to check\n\n", len(files))

	converted := 0
	skipped := 0
	failed := 0

	for _, filePath := range files {
		filename := filepath.Base(filePath)
		
		// Try to detect if it's already CSV format
		file, err := os.Open(filePath)
		if err != nil {
			fmt.Printf("⚠️  %s: Could not open (%v)\n", filename, err)
			failed++
			continue
		}

		scanner := bufio.NewScanner(file)
		var firstLine string
		if scanner.Scan() {
			firstLine = scanner.Text()
		}
		file.Close()

		// Check if already CSV (has header)
		if strings.HasPrefix(firstLine, "timestamp,") {
			fmt.Printf("✓  %s: Already in CSV format\n", filename)
			skipped++
			continue
		}

		// Old format detected - convert it
		fmt.Printf("→  %s: Converting to CSV...\n", filename)

		// Read all old format lines
		file, err = os.Open(filePath)
		if err != nil {
			fmt.Printf("⚠️  %s: Could not read (%v)\n", filename, err)
			failed++
			continue
		}

		var entries []HistoryEntry
		scanner = bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}

			// Parse old space-separated format: TIMESTAMP CODE STATUS DURATION RPE
			parts := strings.Fields(line)
			if len(parts) != 5 {
				continue
			}

			timestamp, err := time.Parse(time.RFC3339, parts[0])
			if err != nil {
				continue
			}

			duration, err := strconv.Atoi(parts[3])
			if err != nil {
				continue
			}

			rpe, err := strconv.Atoi(parts[4])
			if err != nil {
				continue
			}

			entries = append(entries, HistoryEntry{
				Timestamp: timestamp,
				Code:      parts[1],
				Status:    parts[2],
				Duration:  duration,
				RPE:       rpe,
				Subset:    "", // Old logs don't have subset info
			})
		}
		file.Close()

		if len(entries) == 0 {
			fmt.Printf("⚠️  %s: No valid entries found\n", filename)
			failed++
			continue
		}

		// Create backup
		backupPath := filePath + ".bak"
		if err := os.Rename(filePath, backupPath); err != nil {
			fmt.Printf("⚠️  %s: Could not create backup (%v)\n", filename, err)
			failed++
			continue
		}

		// Write new CSV format (with .csv extension)
		// Change from .log to .csv extension
		newFilePath := strings.TrimSuffix(filePath, ".log") + ".csv"
		newFile, err := os.Create(newFilePath)
		if err != nil {
			// Restore backup
			os.Rename(backupPath, filePath)
			fmt.Printf("⚠️  %s: Could not create new file (%v)\n", filename, err)
			failed++
			continue
		}

		writer := csv.NewWriter(newFile)
		
		// Write header
		if err := writer.Write([]string{"timestamp", "code", "status", "duration", "rpe", "subset"}); err != nil {
			newFile.Close()
			os.Rename(backupPath, filePath)
			fmt.Printf("⚠️  %s: Could not write header (%v)\n", filename, err)
			failed++
			continue
		}

		// Write entries
		for _, entry := range entries {
			record := []string{
				entry.Timestamp.Format(time.RFC3339),
				entry.Code,
				entry.Status,
				strconv.Itoa(entry.Duration),
				strconv.Itoa(entry.RPE),
				entry.Subset,
			}
			if err := writer.Write(record); err != nil {
				newFile.Close()
				os.Rename(backupPath, filePath)
				fmt.Printf("⚠️  %s: Could not write entries (%v)\n", filename, err)
				failed++
				continue
			}
		}

		writer.Flush()
		newFile.Close()

		newFilename := strings.TrimSuffix(filename, ".log") + ".csv"
		fmt.Printf("✅ %s → %s: Converted %d entries (backup: %s.bak)\n", filename, newFilename, len(entries), filename)
		converted++
	}

	fmt.Println()
	fmt.Println("═══════════════════════════════════════")
	fmt.Printf("Migration complete:\n")
	fmt.Printf("  Converted: %d\n", converted)
	fmt.Printf("  Skipped:   %d (already CSV)\n", skipped)
	fmt.Printf("  Failed:    %d\n", failed)
	fmt.Println("═══════════════════════════════════════")
	
	if converted > 0 {
		fmt.Println()
		fmt.Println("Backup files (.bak) have been created.")
		fmt.Println("After verifying the migration, you can delete them:")
		fmt.Printf("  rm %s/*.bak\n", cfg.LogsDir)
	}
}
