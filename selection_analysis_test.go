package main

import (
	"os"
	"sort"
	"testing"
	"time"
)

// TestSelectionWeighting generates many snacks and analyzes selection distribution
func TestSelectionWeighting(t *testing.T) {
	// Load test snacks directly from testdata
	tmpDir := t.TempDir()

	// Set up test environment with testdata
	originalMovosDir := os.Getenv("MOVODORO_MOVOS_DIR")
	testMovosPath := "testdata/movos"
	os.Setenv("MOVODORO_MOVOS_DIR", testMovosPath)
	defer os.Setenv("MOVODORO_MOVOS_DIR", originalMovosDir)

	snacks, err := LoadSnacks()
	if err != nil {
		t.Fatalf("Failed to load snacks: %v", err)
	}

	// Initialize test config for history tracking
	cfg := TestConfig(tmpDir)
	cfg.MovosDir = testMovosPath

	if len(snacks) == 0 {
		t.Fatal("No snacks loaded")
	}

	// Track selections
	selectionCount := make(map[string]int)
	totalSelections := 1000

	// Generate 1000 selections
	for i := 0; i < totalSelections; i++ {
		snack, err := SelectSnack(snacks, FilterOptions{}, maxDailyRPEDefault)
		if err != nil {
			t.Fatalf("Selection %d failed: %v", i, err)
		}
		selectionCount[snack.FullCode]++
	}

	// Analyze results
	t.Logf("\n=== Selection Analysis (%d total selections) ===\n", totalSelections)

	// Sort by selection count (descending)
	type movoCount struct {
		code   string
		count  int
		weight float64
	}

	var results []movoCount
	for _, snack := range snacks {
		count := selectionCount[snack.FullCode]
		results = append(results, movoCount{
			code:   snack.FullCode,
			count:  count,
			weight: snack.Weight,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].count > results[j].count
	})

	// Display top 10
	t.Log("\nTop 10 most selected:")
	for i := 0; i < 10 && i < len(results); i++ {
		percentage := float64(results[i].count) / float64(totalSelections) * 100
		t.Logf("  %2d. %-30s | Count: %4d (%.1f%%) | Weight: %.2f",
			i+1, results[i].code, results[i].count, percentage, results[i].weight)
	}

	// Display least selected
	t.Log("\nLeast selected (bottom 5):")
	start := len(results) - 5
	if start < 0 {
		start = 0
	}
	for i := start; i < len(results); i++ {
		percentage := float64(results[i].count) / float64(totalSelections) * 100
		t.Logf("  %2d. %-30s | Count: %4d (%.1f%%) | Weight: %.2f",
			len(results)-i, results[i].code, results[i].count, percentage, results[i].weight)
	}

	// Check that every snack was selected at least once
	for _, result := range results {
		if result.count == 0 {
			t.Logf("Warning: %s was never selected", result.code)
		}
	}
}

// TestMultiDaySelectionPattern simulates selections across multiple days
func TestMultiDaySelectionPattern(t *testing.T) {
	tmpDir := t.TempDir()

	// Set up test environment
	originalMovosDir := os.Getenv("MOVODORO_MOVOS_DIR")
	testMovosPath := "testdata/movos"
	os.Setenv("MOVODORO_MOVOS_DIR", testMovosPath)
	defer os.Setenv("MOVODORO_MOVOS_DIR", originalMovosDir)

	snacks, err := LoadSnacks()
	if err != nil {
		t.Fatalf("Failed to load snacks: %v", err)
	}

	cfg := TestConfig(tmpDir)
	cfg.MovosDir = testMovosPath

	// Simulate 7 days of snacks
	days := 7
	snacksPerDay := 6
	daySelections := make(map[int]map[string]int) // day -> code -> count

	t.Logf("\n=== Multi-Day Selection Pattern (%d days, %d snacks/day) ===\n", days, snacksPerDay)

	for day := 0; day < days; day++ {
		daySelections[day] = make(map[string]int)

		// Simulate this day
		currentDate := time.Now().Add(time.Duration(day-days+1) * 24 * time.Hour)

		for snackNum := 0; snackNum < snacksPerDay; snackNum++ {
			snack, err := SelectSnack(snacks, FilterOptions{}, maxDailyRPEDefault)
			if err != nil {
				t.Fatalf("Day %d, snack %d failed: %v", day, snackNum, err)
			}

			// Log this completion
			entry := HistoryEntry{
				Timestamp: currentDate.Add(time.Duration(snackNum) * time.Hour),
				Code:      snack.FullCode,
				Status:    "done",
				Duration:  snack.GetDefaultDuration(),
				RPE:       snack.EffectiveRPE,
			}

			if err := AppendTodayLog(cfg.LogsDir, entry); err != nil {
				t.Fatalf("Failed to log entry: %v", err)
			}

			daySelections[day][snack.FullCode]++
		}

		// Show this day's selections
		t.Logf("Day %d:", day+1)
		for code, count := range daySelections[day] {
			if count > 1 {
				t.Logf("  %s (x%d)", code, count)
			} else {
				t.Logf("  %s", code)
			}
		}
	}

	// Calculate variety metrics
	totalUnique := make(map[string]bool)
	for _, dayMap := range daySelections {
		for code := range dayMap {
			totalUnique[code] = true
		}
	}

	t.Logf("\nVariety Metrics:")
	t.Logf("  Total snacks available: %d", len(snacks))
	t.Logf("  Unique snacks selected:  %d", len(totalUnique))
	t.Logf("  Coverage: %.1f%%", float64(len(totalUnique))/float64(len(snacks))*100)
}

// TestWeightBoostEffects verifies that boosts are applied correctly
func TestWeightBoostEffects(t *testing.T) {
	// Set up test environment
	originalMovosDir := os.Getenv("MOVODORO_MOVOS_DIR")
	testMovosPath := "testdata/movos"
	os.Setenv("MOVODORO_MOVOS_DIR", testMovosPath)
	defer os.Setenv("MOVODORO_MOVOS_DIR", originalMovosDir)

	snacks, err := LoadSnacks()
	if err != nil {
		t.Fatalf("Failed to load snacks: %v", err)
	}

	// Find an everyday snack
	var everydayMovo *Movo
	for _, s := range snacks {
		if s.MinPerDay > 0 {
			everydayMovo = &s
			break
		}
	}

	if everydayMovo == nil {
		t.Skip("No everyday snacks in test data")
	}

	// Calculate weights for everyday snack
	weight, err := calculateWeight(*everydayMovo)
	if err != nil {
		t.Fatalf("Failed to calculate weight: %v", err)
	}

	t.Logf("\n=== Boost Effect Analysis ===")
	t.Logf("Everyday snack: %s", everydayMovo.FullCode)
	t.Logf("Base weight: %.2f", everydayMovo.Weight)
	t.Logf("Final weight (with boosts): %.2f", weight)
	t.Logf("Boost multiplier: %.1fx", weight/everydayMovo.Weight)

	// Verify everyday boost is applied
	expectedBoost := everydayMovo.Weight * minPerDayBoost
	if weight < expectedBoost {
		t.Errorf("Expected min_per_day boost to be at least %.2f, got %.2f", expectedBoost, weight)
	}
}
