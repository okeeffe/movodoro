package main

import (
	"os"
	"sort"
	"testing"
)

// TestRealMovosWeighting analyzes selection distribution using actual movos
func TestRealMovosWeighting(t *testing.T) {
	// Use the real movos directory
	originalMovosDir := os.Getenv("MOVODORO_MOVOS_DIR")
	realMovosPath := "/Users/rok/coding/personal/movodoro-movos"
	os.Setenv("MOVODORO_MOVOS_DIR", realMovosPath)
	defer os.Setenv("MOVODORO_MOVOS_DIR", originalMovosDir)

	tmpDir := t.TempDir()
	cfg := TestConfig(tmpDir)
	cfg.MovosDir = realMovosPath

	snacks, err := LoadSnacks()
	if err != nil {
		t.Fatalf("Failed to load snacks: %v", err)
	}

	if len(snacks) == 0 {
		t.Fatal("No snacks loaded from real movos")
	}

	t.Logf("\n=== Loaded %d snacks from real movos ===\n", len(snacks))

	// Test 1: With everyday priority (default behavior)
	t.Run("with_everyday_priority", func(t *testing.T) {
		selectionCount := make(map[string]int)
		totalSelections := 1000

		for i := 0; i < totalSelections; i++ {
			snack, err := SelectSnack(snacks, FilterOptions{}, maxDailyRPEDefault)
			if err != nil {
				t.Fatalf("Selection %d failed: %v", i, err)
			}
			selectionCount[snack.FullCode]++
		}

		displayResults(t, snacks, selectionCount, totalSelections, "WITH EVERYDAY PRIORITY")
	})

	// Test 2: Skipping everyday snacks to see weight distribution
	t.Run("skip_everyday_snacks", func(t *testing.T) {
		selectionCount := make(map[string]int)
		totalSelections := 1000

		for i := 0; i < totalSelections; i++ {
			snack, err := SelectSnack(snacks, FilterOptions{SkipMinimums: true}, maxDailyRPEDefault)
			if err != nil {
				t.Fatalf("Selection %d failed: %v", i, err)
			}
			selectionCount[snack.FullCode]++
		}

		displayResults(t, snacks, selectionCount, totalSelections, "SKIP EVERYDAY SNACKS (weight distribution)")
	})
}

func displayResults(t *testing.T, snacks []Movo, selectionCount map[string]int, totalSelections int, title string) {
	t.Logf("\n=== %s (%d selections) ===\n", title, totalSelections)

	type movoCount struct {
		code     string
		count    int
		weight   float64
		everyday bool
	}

	var results []movoCount
	movoMap := make(map[string]Movo)
	for _, snack := range snacks {
		movoMap[snack.FullCode] = snack
		count := selectionCount[snack.FullCode]
		results = append(results, movoCount{
			code:     snack.FullCode,
			count:    count,
			weight:   snack.Weight,
			everyday: snack.MinPerDay > 0,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].count > results[j].count
	})

	// Display top 20
	t.Log("\nTop 20 most selected:")
	for i := 0; i < 20 && i < len(results); i++ {
		if results[i].count == 0 {
			break
		}
		percentage := float64(results[i].count) / float64(totalSelections) * 100
		everyDayMarker := ""
		if results[i].everyday {
			everyDayMarker = " [EVERYDAY]"
		}
		t.Logf("  %2d. %-35s | %4d (%.1f%%) | Weight: %.2f%s",
			i+1, results[i].code, results[i].count, percentage, results[i].weight, everyDayMarker)
	}

	// Count never selected
	neverSelected := 0
	for _, result := range results {
		if result.count == 0 {
			neverSelected++
		}
	}

	if neverSelected > 0 {
		t.Logf("\nWarning: %d snacks were never selected out of %d total", neverSelected, len(results))
	}

	// Show variety metrics
	selectedCount := len(results) - neverSelected
	t.Logf("\nVariety: %d/%d snacks selected (%.1f%% coverage)",
		selectedCount, len(results), float64(selectedCount)/float64(len(results))*100)
}
