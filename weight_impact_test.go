package main

import (
	"fmt"
	"os"
	"sort"
	"testing"
)

// TestWeightImpact calculates the theoretical selection probability for different weights
func TestWeightImpact(t *testing.T) {
	// Use real movos
	originalMovosDir := os.Getenv("MOVODORO_MOVOS_DIR")
	realMovosPath := "/Users/rok/coding/personal/movodoro-movos"
	os.Setenv("MOVODORO_MOVOS_DIR", realMovosPath)
	defer os.Setenv("MOVODORO_MOVOS_DIR", originalMovosDir)

	snacks, err := LoadSnacks()
	if err != nil {
		t.Fatalf("Failed to load snacks: %v", err)
	}

	// Filter to non-everyday snacks for fair comparison
	var nonEveryday []Snack
	totalWeight := 0.0

	for _, s := range snacks {
		if s.MinPerDay == 0 {
			nonEveryday = append(nonEveryday, s)
			totalWeight += s.Weight
		}
	}

	t.Logf("\n=== WEIGHT IMPACT ANALYSIS ===\n")
	t.Logf("Non-everyday snacks: %d", len(nonEveryday))
	t.Logf("Total weight: %.2f\n", totalWeight)

	// Group by weight
	type WeightGroup struct {
		weight   float64
		count    int
		examples []string
	}

	groups := make(map[float64]*WeightGroup)

	for _, s := range nonEveryday {
		if groups[s.Weight] == nil {
			groups[s.Weight] = &WeightGroup{weight: s.Weight}
		}
		g := groups[s.Weight]
		g.count++
		if len(g.examples) < 3 {
			g.examples = append(g.examples, s.FullCode)
		}
	}

	// Sort weights
	var weights []float64
	for w := range groups {
		weights = append(weights, w)
	}
	sort.Float64s(weights)

	// Calculate probabilities
	t.Logf("Weight | Count | Probability/snack | vs 1.0 | Examples")
	t.Logf("-------|-------|-------------------|--------|-------------------")

	baseline := 1.0 / totalWeight * 100

	for _, weight := range weights {
		g := groups[weight]
		probPerSnack := (weight / totalWeight) * 100
		multiplier := probPerSnack / baseline

		exampleStr := ""
		if len(g.examples) > 0 {
			exampleStr = g.examples[0]
			if len(g.examples) > 1 {
				exampleStr += fmt.Sprintf(" (+%d more)", len(g.examples)-1)
			}
		}

		t.Logf("%.2f   | %2d    | %.3f%%           | %.2fx   | %s",
			weight, g.count, probPerSnack, multiplier, exampleStr)
	}

	t.Logf("\n=== SIMPLE COMPARISON ===")
	t.Logf("If you had 2 snacks:")
	t.Logf("  • 1.0 weight → 40%% chance (1.0/2.5)")
	t.Logf("  • 1.5 weight → 60%% chance (1.5/2.5)")
	t.Logf("\nWith %d non-everyday snacks:", len(nonEveryday))
	t.Logf("  • 1.0 weight → %.3f%% chance", baseline)
	t.Logf("  • 1.2 weight → %.3f%% chance (%.0f%% increase)", baseline*1.2, 20.0)
	t.Logf("  • 1.5 weight → %.3f%% chance (%.0f%% increase)", baseline*1.5, 50.0)
}
