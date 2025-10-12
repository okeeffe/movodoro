package main

import (
	"fmt"
	"os"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
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
    movodoro <command> [options]

COMMANDS:
    get                 Get a random movement snack
    done [CODE]         Mark the current/specified snack as completed
    skip [CODE]         Skip the current/specified snack
    report [period]     Show report (day, week, month)
    clear               Clear today's history (requires confirmation)
    config              Show current configuration
    everyday            Show "every day" snacks and completion status
    version             Show version information
    help                Show this help message

REPORT OPTIONS:
    --markdown, --md    Output report in markdown format

GET OPTIONS:
    -c, --category CODE       Filter by category code (e.g., RB, CF, TS)
    -t, --tags TAGS           Filter by tags (comma-separated)
    -d, --duration MINS       Exact duration in minutes
    -m, --min-duration MINS   Minimum duration
    -M, --max-duration MINS   Maximum duration
    -r, --min-rpe RPE         Minimum RPE (for intense work)
    -R, --max-rpe RPE         Maximum RPE (for recovery)

EXAMPLES:
    movodoro get                      # Get any snack
    movodoro get -c RB                # Get from Reset & Breath category
    movodoro get -c CF                # Get from Club Flow category
    movodoro get -t kbx,swingx        # Kettlebell swings
    movodoro get -R 2                 # Very light recovery snacks
    movodoro get -r 7 -t kbx          # Hard kettlebell work
    movodoro get -d 5                 # Exactly 5 minutes
    movodoro done                     # Mark current snack completed
    movodoro done CF-shield-cast      # Mark specific snack completed
    movodoro skip                     # Skip current snack
    movodoro report day               # Show today's report
    movodoro report --md              # Show today's report in markdown
`)
}

// Command handlers are implemented in commands.go
