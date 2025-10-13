package main

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	minPerDayBoost     = 10.0 // Boost for snacks with incomplete min_per_day
	neverDoneBoost     = 3.0  // Boost for snacks never completed
	recencyBoost       = 2.0  // Boost for snacks not done in 7+ days
	recencyDays        = 7    // Days threshold for recency boost
	autoRecoveryMaxRPE = 2    // What the max RPE ends up as if we hit the daily threshold
)

// SelectSnack selects a random snack based on weights and constraints
func SelectSnack(snacks []Movo, filters FilterOptions, maxDailyRPE int) (*Movo, error) {
	cfg := DefaultConfig()

	// Get today's stats
	todayStats, err := GetTodayStatsDaily(cfg.LogsDir)
	if err != nil {
		return nil, fmt.Errorf("error loading today's stats: %w", err)
	}

	// Check if we're in auto-recovery mode
	inRecoveryMode := todayStats.TotalRPE >= maxDailyRPE
	if inRecoveryMode {
		// Override max RPE to 2 for recovery
		filters.MaxRPE = autoRecoveryMaxRPE
		fmt.Println("🔋 Auto-recovery mode: limiting to RPE ≤ 2")
	}

	// Filter snacks
	candidates := filterSnacks(snacks, filters)
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no snacks match the specified filters")
	}

	// Apply min_per_day priority (unless explicitly skipped)
	if !filters.SkipMinimums {
		minimumCandidates, err := filterToIncompleteMinimums(candidates, cfg.LogsDir)
		if err != nil {
			return nil, err
		}
		// If there are incomplete minimum snacks, use only those
		if len(minimumCandidates) > 0 {
			candidates = minimumCandidates
		}
	}

	// Remove snacks that have hit their max_per_day limit
	candidates, err = filterByFrequency(candidates)
	if err != nil {
		return nil, err
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("all matching snacks have reached their daily limit")
	}

	// Calculate weights
	weighted := make([]weightedSnack, len(candidates))
	for i, snack := range candidates {
		weight, err := calculateWeight(snack)
		if err != nil {
			return nil, err
		}
		weighted[i] = weightedSnack{snack: snack, weight: weight}
	}

	// Select using weighted random
	selected := weightedRandomSelect(weighted)
	return &selected, nil
}

type weightedSnack struct {
	snack  Movo
	weight float64
}

// filterSnacks applies all filters to the snack list
func filterSnacks(snacks []Movo, filters FilterOptions) []Movo {
	var filtered []Movo

	for _, snack := range snacks {
		// Category filter
		if filters.Category != "" && snack.CategoryCode != filters.Category {
			continue
		}

		// Tag filter
		if !snack.HasAllTags(filters.Tags) {
			continue
		}

		// RPE filters
		if filters.MinRPE > 0 && snack.EffectiveRPE < filters.MinRPE {
			continue
		}
		if filters.MaxRPE > 0 && snack.EffectiveRPE > filters.MaxRPE {
			continue
		}

		// Duration filters
		if filters.ExactDuration > 0 {
			// For exact duration, check if the duration falls in the range
			if filters.ExactDuration < snack.DurationMin || filters.ExactDuration > snack.DurationMax {
				continue
			}
		} else {
			// Range-based filtering
			minDur := filters.MinDuration
			maxDur := filters.MaxDuration

			// Set defaults if not specified
			if minDur == 0 {
				minDur = 0
			}
			if maxDur == 0 {
				maxDur = 999
			}

			if !snack.MatchesDuration(minDur, maxDur) {
				continue
			}
		}

		filtered = append(filtered, snack)
	}

	return filtered
}

// filterToIncompleteMinimums returns only snacks that haven't met their min_per_day requirement
func filterToIncompleteMinimums(snacks []Movo, logsDir string) ([]Movo, error) {
	var incomplete []Movo

	for _, snack := range snacks {
		// Only consider snacks with a minimum requirement
		if snack.MinPerDay == 0 {
			continue
		}

		// Check how many times done today
		doneToday, _, err := GetCountTodayDaily(logsDir, snack.FullCode)
		if err != nil {
			return nil, err
		}

		// Include if haven't met minimum yet
		if doneToday < snack.MinPerDay {
			incomplete = append(incomplete, snack)
		}
	}

	return incomplete, nil
}

// filterByFrequency removes snacks that have hit their daily/weekly limits
func filterByFrequency(snacks []Movo) ([]Movo, error) {
	cfg := DefaultConfig()
	var filtered []Movo

	for _, snack := range snacks {
		doneToday, _, err := GetCountTodayDaily(cfg.LogsDir, snack.FullCode)
		if err != nil {
			return nil, err
		}

		// Check max_per_day
		if snack.MaxPerDay > 0 && doneToday >= snack.MaxPerDay {
			continue
		}

		// TODO: Implement max_per_week check if needed

		filtered = append(filtered, snack)
	}

	return filtered, nil
}

// calculateWeight calculates the final weight for a snack with all boosts
func calculateWeight(snack Movo) (float64, error) {
	cfg := DefaultConfig()
	weight := snack.Weight

	// Min per day boost - applies when snack has minimum and hasn't met it yet
	if snack.MinPerDay > 0 {
		doneToday, _, err := GetCountTodayDaily(cfg.LogsDir, snack.FullCode)
		if err != nil {
			return 0, err
		}
		if doneToday < snack.MinPerDay {
			weight *= minPerDayBoost
		}
	}

	// Never done boost
	everDone, err := HasEverBeenDoneDaily(cfg.LogsDir, snack.FullCode)
	if err != nil {
		return 0, err
	}
	if !everDone {
		weight *= neverDoneBoost
	}

	// Recency boost
	lastDone, err := GetLastDoneDaily(cfg.LogsDir, snack.FullCode)
	if err != nil {
		return 0, err
	}
	if lastDone != nil {
		daysSince := time.Since(*lastDone).Hours() / 24
		if daysSince >= float64(recencyDays) {
			weight *= recencyBoost
		}
	}

	return weight, nil
}

// weightedRandomSelect selects a snack using weighted random selection
func weightedRandomSelect(weighted []weightedSnack) Movo {
	// Calculate total weight
	totalWeight := 0.0
	for _, w := range weighted {
		totalWeight += w.weight
	}

	// Random selection
	r := rand.Float64() * totalWeight
	cumulative := 0.0

	for _, w := range weighted {
		cumulative += w.weight
		if r <= cumulative {
			return w.snack
		}
	}

	// Fallback (shouldn't happen)
	return weighted[len(weighted)-1].snack
}

func init() {
	// Seed random number generator
	rand.Seed(time.Now().UnixNano())
}
