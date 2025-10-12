package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
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
		tags        string
		category    string
		duration    int
		minDuration int
		maxDuration int
		minRPE      int
		maxRPE      int
		skipDailies bool
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
	fs.BoolVar(&skipDailies, "skip-dailies", false, "Skip everyday snacks priority")

	fs.Parse(args)

	// Load snacks
	snacks, err := LoadSnacks()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading snacks: %v\n", err)
		os.Exit(1)
	}

	// Parse filters
	filters := FilterOptions{
		Category:      strings.TrimSpace(strings.ToUpper(category)),
		MinDuration:   minDuration,
		MaxDuration:   maxDuration,
		ExactDuration: duration,
		MinRPE:        minRPE,
		MaxRPE:        maxRPE,
		SkipDailies:   skipDailies,
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

	// Display the snack
	displaySnack(snack)
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
	var snack *Snack
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

	// Prompt for actual duration
	defaultDuration := snack.GetDefaultDuration()
	fmt.Printf("How many minutes did you spend? (default: %d): ", defaultDuration)

	reader := bufio.NewReader(os.Stdin)
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

	// Create history entry
	entry := HistoryEntry{
		Timestamp: time.Now(),
		Code:      code,
		Status:    "done",
		Duration:  duration,
		RPE:       snack.EffectiveRPE,
	}

	// Save to history
	if err := AppendTodayLog(appConfig.LogsDir, entry); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving to history: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Marked '%s' as completed (%d minutes, RPE %d)\n", snack.Title, duration, snack.EffectiveRPE)

	// Show updated daily stats
	stats, _ := GetTodayStatsDaily(appConfig.LogsDir)
	fmt.Printf("📊 Today: %d snacks, %d minutes, %d RPE\n", stats.TotalSnacks, stats.TotalDuration, stats.TotalRPE)
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
	var snack *Snack
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
	fs.BoolVar(&markdown, "markdown", false, "Output in markdown format")
	fs.BoolVar(&markdown, "md", false, "Output in markdown format")

	fs.Parse(args)

	remaining := fs.Args()
	period := "day"
	if len(remaining) > 0 {
		period = remaining[0]
	}

	switch period {
	case "day", "today":
		if markdown {
			showDayReportMarkdown()
		} else {
			showDayReport()
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

func showDayReport() {
	stats, err := GetTodayStatsDaily(appConfig.LogsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading stats: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("═══════════════════════════════════════")
	fmt.Printf("  TODAY'S MOVEMENT REPORT\n")
	fmt.Printf("  %s\n", stats.Date.Format("Monday, January 2, 2006"))
	fmt.Println("═══════════════════════════════════════")
	fmt.Println()

	fmt.Printf("📊 Summary:\n")
	fmt.Printf("   Total snacks:    %d\n", len(stats.CompletedSnacks))
	fmt.Printf("   Total duration:  %d minutes\n", stats.TotalDuration)
	fmt.Printf("   Total RPE:       %d / %d\n", stats.TotalRPE, maxDailyRPEDefault)
	fmt.Println()

	if len(stats.CompletedSnacks) > 0 {
		fmt.Printf("✅ Completed:\n")
		for _, entry := range stats.CompletedSnacks {
			fmt.Printf("   %s - %s (%dm, RPE %d)\n",
				entry.Timestamp.Format("15:04"),
				entry.Code,
				entry.Duration,
				entry.RPE)
		}
		fmt.Println()
	}

	if len(stats.SkippedSnacks) > 0 {
		fmt.Printf("⏭️  Skipped:\n")
		for _, entry := range stats.SkippedSnacks {
			fmt.Printf("   %s - %s\n",
				entry.Timestamp.Format("15:04"),
				entry.Code)
		}
		fmt.Println()
	}

	if stats.TotalRPE >= maxDailyRPEDefault {
		fmt.Println("🔋 Auto-recovery mode active (RPE limit reached)")
	}
}

func showDayReportMarkdown() {
	stats, err := GetTodayStatsDaily(appConfig.LogsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading stats: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("# Movement Report - %s\n\n", stats.Date.Format("Monday, January 2, 2006"))

	fmt.Println("## Summary")
	fmt.Println()
	fmt.Printf("- **Total snacks:** %d\n", len(stats.CompletedSnacks))
	fmt.Printf("- **Total duration:** %d minutes\n", stats.TotalDuration)
	fmt.Printf("- **Total RPE:** %d / %d\n", stats.TotalRPE, maxDailyRPEDefault)
	fmt.Println()

	if len(stats.CompletedSnacks) > 0 {
		fmt.Println("## Completed")
		fmt.Println()
		for _, entry := range stats.CompletedSnacks {
			fmt.Printf("- **%s** - `%s` (%d min, RPE %d)\n",
				entry.Timestamp.Format("15:04"),
				entry.Code,
				entry.Duration,
				entry.RPE)
		}
		fmt.Println()
	}

	if len(stats.SkippedSnacks) > 0 {
		fmt.Println("## Skipped")
		fmt.Println()
		for _, entry := range stats.SkippedSnacks {
			fmt.Printf("- **%s** - `%s`\n",
				entry.Timestamp.Format("15:04"),
				entry.Code)
		}
		fmt.Println()
	}

	if stats.TotalRPE >= maxDailyRPEDefault {
		fmt.Println("*Auto-recovery mode active (RPE limit reached)*")
	}
}

func displaySnack(snack *Snack) {
	fmt.Println()
	fmt.Println("═══════════════════════════════════════")
	fmt.Printf("  %s\n", snack.Title)
	fmt.Println("═══════════════════════════════════════")
	fmt.Println()

	fmt.Println(snack.Description)
	fmt.Println()

	fmt.Printf("⏱️  Duration: %d-%d minutes\n", snack.DurationMin, snack.DurationMax)
	fmt.Printf("💪 RPE: %d/10\n", snack.EffectiveRPE)
	fmt.Printf("🏷️  Code: %s\n", snack.FullCode)

	if len(snack.AllTags) > 0 {
		fmt.Printf("🔖 Tags: %s\n", strings.Join(snack.AllTags, ", "))
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

	if stats.TotalSnacks == 0 {
		fmt.Println("No entries for today to clear.")
		return
	}

	fmt.Printf("This will delete today's log file with %d entries:\n", stats.TotalSnacks)
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

	fmt.Printf("✅ Cleared %d entries from today's history\n", stats.TotalSnacks)
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
	// Load snacks
	snacks, err := LoadSnacks()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading snacks: %v\n", err)
		os.Exit(1)
	}

	// Filter to only every_day snacks
	var everydaySnacks []Snack
	for _, snack := range snacks {
		if snack.EveryDay {
			everydaySnacks = append(everydaySnacks, snack)
		}
	}

	if len(everydaySnacks) == 0 {
		fmt.Println("No snacks marked as 'every_day'")
		return
	}

	fmt.Println("═══════════════════════════════════════")
	fmt.Println("  EVERY DAY SNACKS")
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
	for _, snack := range everydaySnacks {
		count := completedToday[snack.FullCode]
		status := "❌"
		if count > 0 {
			status = "✅"
		}

		fmt.Printf("%s %s\n", status, snack.Title)
		fmt.Printf("   Code: %s | RPE: %d | Duration: %d-%d min\n",
			snack.FullCode, snack.EffectiveRPE, snack.DurationMin, snack.DurationMax)

		if count > 0 {
			fmt.Printf("   Completed %d time(s) today\n", count)
		} else {
			fmt.Printf("   Not yet done today\n")
		}
		fmt.Println()
	}

	completed := 0
	for _, snack := range everydaySnacks {
		if completedToday[snack.FullCode] > 0 {
			completed++
		}
	}

	fmt.Printf("Summary: %d/%d everyday snacks completed\n", completed, len(everydaySnacks))
}

// handleInteractive implements the interactive mode (default when running `movodoro`)
func handleInteractive(args []string) {
	// Load snacks
	snacks, err := LoadSnacks()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading snacks: %v\n", err)
		os.Exit(1)
	}

	// Start with default filters
	filters := FilterOptions{}

	for {
		var snack *Snack

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

		// Display the snack
		displaySnackInteractive(snack)

		// Get user choice
		choice := getInteractiveChoice(snack.EveryDay)

		switch choice {
		case "d": // Done
			handleDoneInteractive(snack)
			os.Remove(appConfig.CurrentPath) // Clear saved snack
			return                           // Exit after marking done

		case "s": // Skip
			handleSkipInteractive(snack)
			os.Remove(appConfig.CurrentPath) // Clear saved snack
			filters.SkipDailies = false      // Reset skip dailies flag
			// Continue loop to get next snack

		case "x": // Skip dailies (only if everyday snack)
			if snack.EveryDay {
				fmt.Printf("\n⏭️  Skipping dailies for now...\n")
				os.Remove(appConfig.CurrentPath) // Clear saved snack
				filters.SkipDailies = true
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

// displaySnackInteractive displays a snack in interactive mode
func displaySnackInteractive(snack *Snack) {
	fmt.Println()
	fmt.Println("═══════════════════════════════════════")
	fmt.Printf("  %s\n", snack.Title)
	fmt.Println("═══════════════════════════════════════")
	fmt.Println()

	fmt.Println(snack.Description)
	fmt.Println()

	fmt.Printf("⏱️  Duration: %d-%d minutes\n", snack.DurationMin, snack.DurationMax)
	fmt.Printf("💪 RPE: %d/10\n", snack.EffectiveRPE)
	fmt.Printf("🏷️  Code: %s\n", snack.FullCode)

	if len(snack.AllTags) > 0 {
		fmt.Printf("🔖 Tags: %s\n", strings.Join(snack.AllTags, ", "))
	}
	fmt.Println()
}

// getInteractiveChoice prompts user for action choice
func getInteractiveChoice(isEveryday bool) string {
	fmt.Println("What would you like to do?")
	fmt.Println("  [d] Done (log completion)")
	fmt.Println("  [s] Skip (try another snack)")
	if isEveryday {
		fmt.Println("  [x] Skip dailies (get non-daily snack)")
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
		if isEveryday {
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

// handleDoneInteractive handles completing a snack in interactive mode
func handleDoneInteractive(snack *Snack) {
	// Prompt for actual duration
	defaultDuration := snack.GetDefaultDuration()
	fmt.Printf("\nHow many minutes did you spend? (default: %d): ", defaultDuration)

	reader := bufio.NewReader(os.Stdin)
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

	// Create history entry
	entry := HistoryEntry{
		Timestamp: time.Now(),
		Code:      snack.FullCode,
		Status:    "done",
		Duration:  duration,
		RPE:       snack.EffectiveRPE,
	}

	// Save to history
	if err := AppendTodayLog(appConfig.LogsDir, entry); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving to history: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✅ Marked '%s' as completed (%d minutes, RPE %d)\n", snack.Title, duration, snack.EffectiveRPE)

	// Show updated daily stats
	stats, _ := GetTodayStatsDaily(appConfig.LogsDir)
	fmt.Printf("📊 Today: %d snacks, %d minutes, %d RPE\n\n", stats.TotalSnacks, stats.TotalDuration, stats.TotalRPE)
}

// handleSkipInteractive handles skipping a snack in interactive mode
func handleSkipInteractive(snack *Snack) {
	// Create history entry with 0 duration and RPE
	entry := HistoryEntry{
		Timestamp: time.Now(),
		Code:      snack.FullCode,
		Status:    "skip",
		Duration:  0,
		RPE:       0,
	}

	// Save to history
	if err := AppendTodayLog(appConfig.LogsDir, entry); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving to history: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n⏭️  Skipped '%s'\n", snack.Title)
}
