package main

import (
	"os"
	"testing"
)

// TestLoadSubsets tests loading subsets.yaml
func TestLoadSubsets(t *testing.T) {
	t.Run("load valid subsets", func(t *testing.T) {
		subsetsConfig, err := LoadSubsets("testdata/movos")
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if len(subsetsConfig.Subsets) == 0 {
			t.Fatal("expected subsets to be loaded")
		}

		// Check for expected test subsets
		if _, exists := subsetsConfig.Subsets["recovery"]; !exists {
			t.Error("expected 'recovery' subset to exist")
		}
		if _, exists := subsetsConfig.Subsets["breath-only"]; !exists {
			t.Error("expected 'breath-only' subset to exist")
		}
	})

	t.Run("missing subsets.yaml returns empty config", func(t *testing.T) {
		// Use a directory that doesn't have subsets.yaml
		subsetsConfig, err := LoadSubsets("testdata")
		if err != nil {
			t.Fatalf("expected no error for missing file, got: %v", err)
		}

		if len(subsetsConfig.Subsets) != 0 {
			t.Errorf("expected empty subsets for missing file, got %d subsets", len(subsetsConfig.Subsets))
		}
	})

	t.Run("subset has correct structure", func(t *testing.T) {
		subsetsConfig, err := LoadSubsets("testdata/movos")
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		recovery, exists := subsetsConfig.Subsets["recovery"]
		if !exists {
			t.Fatal("expected 'recovery' subset to exist")
		}

		if recovery.Description == "" {
			t.Error("expected description to be set")
		}

		expectedCodes := []string{"TB-box-breath", "TB-deep-breath", "TS-light-move"}
		if len(recovery.Codes) != len(expectedCodes) {
			t.Errorf("expected %d codes, got %d", len(expectedCodes), len(recovery.Codes))
		}
	})
}

// TestFilterBySubset tests subset filtering logic
func TestFilterBySubset(t *testing.T) {
	// Set MovosDir for loading
	originalDir := os.Getenv("MOVODORO_MOVOS_DIR")
	os.Setenv("MOVODORO_MOVOS_DIR", "testdata/movos")
	defer os.Setenv("MOVODORO_MOVOS_DIR", originalDir)

	// Load test snacks
	snacks, err := LoadSnacks()
	if err != nil {
		t.Fatalf("failed to load test snacks: %v", err)
	}

	t.Run("filter to recovery subset", func(t *testing.T) {
		filtered, err := filterBySubset(snacks, "recovery", "testdata/movos")
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if len(filtered) != 3 {
			t.Errorf("expected 3 snacks in recovery subset, got %d", len(filtered))
		}

		// Check all filtered snacks are in expected codes
		expectedCodes := map[string]bool{
			"TB-box-breath":  true,
			"TB-deep-breath": true,
			"TS-light-move":  true,
		}

		for _, snack := range filtered {
			if !expectedCodes[snack.FullCode] {
				t.Errorf("unexpected snack in filtered results: %s", snack.FullCode)
			}
		}
	})

	t.Run("filter to breath-only subset", func(t *testing.T) {
		filtered, err := filterBySubset(snacks, "breath-only", "testdata/movos")
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if len(filtered) != 2 {
			t.Errorf("expected 2 snacks in breath-only subset, got %d", len(filtered))
		}

		for _, snack := range filtered {
			if snack.CategoryCode != "TB" {
				t.Errorf("expected only TB category snacks, got %s", snack.CategoryCode)
			}
		}
	})

	t.Run("non-existent subset returns error", func(t *testing.T) {
		_, err := filterBySubset(snacks, "non-existent", "testdata/movos")
		if err == nil {
			t.Error("expected error for non-existent subset")
		}
	})

	t.Run("empty subset returns empty list", func(t *testing.T) {
		filtered, err := filterBySubset(snacks, "empty", "testdata/movos")
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if len(filtered) != 0 {
			t.Errorf("expected 0 snacks in empty subset, got %d", len(filtered))
		}
	})

	t.Run("single item subset", func(t *testing.T) {
		filtered, err := filterBySubset(snacks, "single-item", "testdata/movos")
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if len(filtered) != 1 {
			t.Errorf("expected 1 snack in single-item subset, got %d", len(filtered))
		}

		if len(filtered) > 0 && filtered[0].FullCode != "TB-box-breath" {
			t.Errorf("expected TB-box-breath, got %s", filtered[0].FullCode)
		}
	})
}

// TestSubsetWithOtherFilters tests subset filtering combined with other filters
func TestSubsetWithOtherFilters(t *testing.T) {
	// Set MovosDir for loading
	originalDir := os.Getenv("MOVODORO_MOVOS_DIR")
	os.Setenv("MOVODORO_MOVOS_DIR", "testdata/movos")
	defer os.Setenv("MOVODORO_MOVOS_DIR", originalDir)

	snacks, err := LoadSnacks()
	if err != nil {
		t.Fatalf("failed to load test snacks: %v", err)
	}

	t.Run("subset + tag filter intersection", func(t *testing.T) {
		// First apply basic filters (tag filter for breathx)
		filters := FilterOptions{
			Tags: []string{"breathx"},
		}
		candidates := filterSnacks(snacks, filters)

		// Then apply subset filter (recovery)
		filtered, err := filterBySubset(candidates, "recovery", "testdata/movos")
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		// Should only get TB-box-breath and TB-deep-breath (both have breathx tag and in recovery subset)
		if len(filtered) != 2 {
			t.Errorf("expected 2 snacks (intersection of breathx + recovery), got %d", len(filtered))
		}

		for _, snack := range filtered {
			if snack.CategoryCode != "TB" {
				t.Errorf("expected only breath snacks, got %s", snack.FullCode)
			}
		}
	})

	t.Run("subset + RPE filter intersection", func(t *testing.T) {
		// Filter for RPE <= 2
		filters := FilterOptions{
			MaxRPE: 2,
		}
		candidates := filterSnacks(snacks, filters)

		// Apply recovery subset
		filtered, err := filterBySubset(candidates, "recovery", "testdata/movos")
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		// Should get TB-box-breath (RPE 1) and TB-deep-breath (RPE 2)
		// TS-light-move has RPE 3, so should be excluded by RPE filter
		if len(filtered) != 2 {
			t.Errorf("expected 2 snacks (RPE ≤ 2 in recovery), got %d", len(filtered))
		}
	})

	t.Run("subset excludes all candidates returns empty", func(t *testing.T) {
		// Filter for high RPE (≥ 7)
		filters := FilterOptions{
			MinRPE: 7,
		}
		candidates := filterSnacks(snacks, filters)

		// Apply recovery subset (which only has low RPE movos)
		filtered, err := filterBySubset(candidates, "recovery", "testdata/movos")
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		// Should be empty - no high RPE movos in recovery subset
		if len(filtered) != 0 {
			t.Errorf("expected 0 snacks (no high RPE in recovery), got %d", len(filtered))
		}
	})
}

// TestSelectSnackWithSubset tests end-to-end selection with subsets
func TestSelectSnackWithSubset(t *testing.T) {
	// Set MovosDir for loading
	originalDir := os.Getenv("MOVODORO_MOVOS_DIR")
	os.Setenv("MOVODORO_MOVOS_DIR", "testdata/movos")
	defer os.Setenv("MOVODORO_MOVOS_DIR", originalDir)

	// Create a test config
	testConfig := TestConfig("testdata")
	testConfig.MovosDir = "testdata/movos"

	// Clean up any test logs
	defer os.RemoveAll(testConfig.LogsDir)

	snacks, err := LoadSnacks()
	if err != nil {
		t.Fatalf("failed to load test snacks: %v", err)
	}

	t.Run("select from subset", func(t *testing.T) {
		filters := FilterOptions{
			Subset: "breath-only",
		}

		// Temporarily override config for selection
		originalConfig := appConfig
		appConfig = testConfig
		defer func() { appConfig = originalConfig }()

		selected, err := SelectSnack(snacks, filters, 30)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if selected.CategoryCode != "TB" {
			t.Errorf("expected breath category, got %s", selected.CategoryCode)
		}
	})

	t.Run("subset respects min_per_day priority", func(t *testing.T) {
		// TB-box-breath has min_per_day: 1 and is in recovery subset
		// Should be prioritized even within subset
		filters := FilterOptions{
			Subset: "recovery",
		}

		originalConfig := appConfig
		appConfig = testConfig
		defer func() { appConfig = originalConfig }()

		// Run multiple times to check priority
		breathCount := 0
		iterations := 10
		for i := 0; i < iterations; i++ {
			selected, err := SelectSnack(snacks, filters, 30)
			if err != nil {
				continue
			}
			if selected.FullCode == "TB-box-breath" {
				breathCount++
			}
		}

		// Should get box-breath most of the time due to min_per_day boost
		if breathCount < iterations/2 {
			t.Errorf("expected TB-box-breath to be selected frequently due to min_per_day, got %d/%d", breathCount, iterations)
		}
	})

	t.Run("empty subset after filters returns error", func(t *testing.T) {
		filters := FilterOptions{
			Subset: "breath-only",
			MinRPE: 7, // No breath movos have RPE 7
		}

		originalConfig := appConfig
		appConfig = testConfig
		defer func() { appConfig = originalConfig }()

		_, err := SelectSnack(snacks, filters, 30)
		if err == nil {
			t.Error("expected error when no snacks match subset + filters")
		}
	})
}

// TestConfigActiveSubset tests that config properly loads MOVODORO_ACTIVE_SUBSET
func TestConfigActiveSubset(t *testing.T) {
	t.Run("env var sets active subset", func(t *testing.T) {
		originalSubset := os.Getenv("MOVODORO_ACTIVE_SUBSET")
		defer os.Setenv("MOVODORO_ACTIVE_SUBSET", originalSubset)

		os.Setenv("MOVODORO_ACTIVE_SUBSET", "test-subset")

		cfg := DefaultConfig()

		if cfg.ActiveSubset != "test-subset" {
			t.Errorf("expected ActiveSubset to be 'test-subset', got '%s'", cfg.ActiveSubset)
		}
	})

	t.Run("no env var means empty active subset", func(t *testing.T) {
		originalSubset := os.Getenv("MOVODORO_ACTIVE_SUBSET")
		defer os.Setenv("MOVODORO_ACTIVE_SUBSET", originalSubset)

		os.Unsetenv("MOVODORO_ACTIVE_SUBSET")

		cfg := DefaultConfig()

		if cfg.ActiveSubset != "" {
			t.Errorf("expected ActiveSubset to be empty, got '%s'", cfg.ActiveSubset)
		}
	})
}
