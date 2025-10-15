package main

import (
	"fmt"
	"os"
)

const version = "0.1.0"

func main() {
	// If no command provided (or starts with --), enter interactive mode
	if len(os.Args) < 2 || (len(os.Args) >= 2 && os.Args[1][:1] == "-") {
		handleInteractive(os.Args[1:])
		return
	}

	command := os.Args[1]

	switch command {
	case "get":
		handleGet(os.Args[2:])
	case "done":
		handleDone(os.Args[2:])
	case "skip":
		handleSkip(os.Args[2:])
	case "report":
		handleReport(os.Args[2:])
	case "clear":
		handleClear(os.Args[2:])
	case "config":
		handleConfig(os.Args[2:])
	case "everyday":
		handleEveryday(os.Args[2:])
	case "subsets":
		handleSubsets(os.Args[2:])
	case "version", "--version", "-v":
		fmt.Printf("movodoro version %s\n", version)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Print(`movodoro - Movement snack generator

USAGE:
    movodoro [options]         # Interactive mode (default)
    movodoro <command> [options]

COMMANDS:
    get                 Get a random movement snack
    done [CODE]         Mark the current/specified snack as completed
    skip [CODE]         Skip the current/specified snack
    report [period]     Show report (day, week, month)
    clear               Clear today's history (requires confirmation)
    config              Show current configuration
    everyday            Show "every day" snacks and completion status
    subsets             List available subsets from subsets.yaml
    version             Show version information
    help                Show this help message

INTERACTIVE MODE OPTIONS:
    --subset NAME       Use a named subset from subsets.yaml

REPORT OPTIONS:
    --markdown, --md    Output report in markdown format
    -v, --verbose       Show titles and tags

GET OPTIONS:
    -c, --category CODE       Filter by category code (e.g., RB, CF, TS)
    -t, --tags TAGS           Filter by tags (comma-separated)
    -d, --duration MINS       Exact duration in minutes
    -m, --min-duration MINS   Minimum duration
    -M, --max-duration MINS   Maximum duration
    -r, --min-rpe RPE         Minimum RPE (for intense work)
    -R, --max-rpe RPE         Maximum RPE (for recovery)
    --subset NAME             Use a named subset from subsets.yaml

SUBSETS:
    Subsets allow you to restrict movement selection to a specific collection
    of movos. Perfect for injury recovery, travel, or equipment constraints.

    Define subsets in: $MOVODORO_MOVOS_DIR/subsets.yaml

    Activation:
      movodoro get --subset NAME           # One-time use
      movodoro --subset NAME               # Interactive mode
      export MOVODORO_ACTIVE_SUBSET=NAME   # Persistent (env var)

EXAMPLES:
    movodoro                              # Interactive mode
    movodoro --subset back-safe           # Interactive with subset
    movodoro get                          # Get any snack
    movodoro get --subset back-safe       # Get from subset
    movodoro get -c RB                    # Get from Reset & Breath category
    movodoro get -t kbx,swingx            # Kettlebell swings
    movodoro get -R 2                     # Very light recovery snacks
    movodoro done                         # Mark current snack completed
    movodoro report --md -v               # Verbose markdown report
    movodoro subsets                      # List available subsets
`)
}

// Command handlers are implemented in commands.go
