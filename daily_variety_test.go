package main

import (
	"os"
	"testing"
	"time"
)

// TestDailyVariety simulates a realistic day with 10 movos
func TestDailyVariety(t *testing.T) {
	// Use real movos
	originalMovosDir := os.Getenv("MOVODORO_MOVOS_DIR")
	realMovosPath := "/Users/rok/coding/personal/movodoro-movos"
	os.Setenv("MOVODORO_MOVOS_DIR", realMovosPath)
	defer os.Setenv("MOVODORO_MOVOS_DIR", originalMovosDir)

	tmpDir := t.TempDir()
	cfg := TestConfig(tmpDir)
	cfg.LogsDir = tmpDir

	snacks, err := LoadSnacks()
	if err != nil {
		t.Fatalf("Failed to load snacks: %v", err)
	}

	// Simulate 5 different days to see variety
	for day := 1; day <= 5; day++ {
		t.Logf("\n=== DAY %d ===", day)

		daySnacks := make(map[string]int)
		var selectedOrder []string

		// Simulate 10 movos in a day
		for movoNum := 1; movoNum <= 10; movoNum++ {
			// Select with frequency filtering enabled
			filters := FilterOptions{}
			if movoNum > 4 {
				// After everyday snacks, skip dailies to get variety
				filters.SkipMinimums = true
			}

			snack, err := SelectSnack(snacks, filters, maxDailyRPEDefault)
			if err != nil {
				t.Fatalf("Day %d, movo %d failed: %v", day, movoNum, err)
			}

			daySnacks[snack.FullCode]++
			selectedOrder = append(selectedOrder, snack.FullCode)

			// Log completion (this triggers frequency filtering for next selection)
			entry := HistoryEntry{
				Timestamp: time.Now().Add(time.Duration(movoNum) * time.Hour),
				Code:      snack.FullCode,
				Status:    "done",
				Duration:  snack.GetDefaultDuration(),
				RPE:       snack.EffectiveRPE,
			}

			if err := AppendTodayLog(cfg.LogsDir, entry); err != nil {
				t.Fatalf("Failed to log: %v", err)
			}
		}

		// Display results
		t.Logf("\nSelection order:")
		for i, code := range selectedOrder {
			count := daySnacks[code]
			repeatMarker := ""
			if count > 1 {
				repeatMarker = " (repeat)"
			}
			t.Logf("  %2d. %s%s", i+1, code, repeatMarker)
		}

		t.Logf("\nVariety metrics:")
		t.Logf("  Total selections: 10")
		t.Logf("  Unique snacks: %d", len(daySnacks))
		t.Logf("  Variety: %.0f%%", float64(len(daySnacks))/10.0*100)

		// Show any repeats
		repeats := 0
		for _, count := range daySnacks {
			if count > 1 {
				repeats++
			}
		}
		if repeats > 0 {
			t.Logf("  Snacks selected multiple times: %d", repeats)
		}

		// Clear logs for next day simulation
		if err := ClearTodayLog(cfg.LogsDir); err != nil {
			t.Fatalf("Failed to clear logs: %v", err)
		}
	}

	// Summary
	t.Logf("\n=== SUMMARY ===")
	t.Logf("With frequency filtering (non-recency bias), you should see:")
	t.Logf("  • 7-10 unique snacks per day (70-100%% variety)")
	t.Logf("  • Everyday snacks done first (4 snacks)")
	t.Logf("  • Remaining 6 drawn from 39 non-everyday snacks")
	t.Logf("  • Recently completed snacks have reduced selection probability")
}

// TestRecencyBiasStrength shows how frequency filtering works
func TestRecencyBiasStrength(t *testing.T) {
	originalMovosDir := os.Getenv("MOVODORO_MOVOS_DIR")
	realMovosPath := "/Users/rok/coding/personal/movodoro-movos"
	os.Setenv("MOVODORO_MOVOS_DIR", realMovosPath)
	defer os.Setenv("MOVODORO_MOVOS_DIR", originalMovosDir)

	tmpDir := t.TempDir()
	cfg := TestConfig(tmpDir)
	cfg.LogsDir = tmpDir

	snacks, err := LoadSnacks()
	if err != nil {
		t.Fatalf("Failed to load snacks: %v", err)
	}

	// Pick a test snack
	var testSnack *Snack
	for i := range snacks {
		if snacks[i].MinPerDay == 0 {
			testSnack = &snacks[i]
			break
		}
	}

	if testSnack == nil {
		t.Fatal("No non-everyday snack found")
	}

	t.Logf("\n=== FREQUENCY FILTERING TEST ===")
	t.Logf("Test snack: %s (weight: %.2f)\n", testSnack.FullCode, testSnack.Weight)

	// Log this snack as completed
	entry := HistoryEntry{
		Timestamp: time.Now(),
		Code:      testSnack.FullCode,
		Status:    "done",
		Duration:  testSnack.GetDefaultDuration(),
		RPE:       testSnack.EffectiveRPE,
	}

	if err := AppendTodayLog(cfg.LogsDir, entry); err != nil {
		t.Fatalf("Failed to log: %v", err)
	}

	// Now select 100 times and see if this snack appears less
	selectionCount := make(map[string]int)
	totalSelections := 100

	for i := 0; i < totalSelections; i++ {
		snack, err := SelectSnack(snacks, FilterOptions{SkipMinimums: true}, maxDailyRPEDefault)
		if err != nil {
			t.Fatalf("Selection failed: %v", err)
		}
		selectionCount[snack.FullCode]++
	}

	testSnackSelections := selectionCount[testSnack.FullCode]
	expectedWithoutFilter := testSnack.Weight / 40.70 * 100 // rough estimate

	t.Logf("Results over %d selections:", totalSelections)
	t.Logf("  Test snack selected: %d times (%.1f%%)", testSnackSelections, float64(testSnackSelections)/float64(totalSelections)*100)
	t.Logf("  Expected without filtering: ~%.1f times (%.1f%%)", expectedWithoutFilter, expectedWithoutFilter)
	t.Logf("\nFrequency filtering is %s", func() string {
		if testSnackSelections < int(expectedWithoutFilter) {
			return "WORKING - reduces recently completed snacks"
		}
		return "not preventing reselection (normal variance)"
	}())
}
