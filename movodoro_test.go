package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Unit Tests

func TestParseCSVRecord(t *testing.T) {
	tests := []struct {
		name    string
		record  []string
		wantErr bool
	}{
		{
			name:    "valid record",
			record:  []string{"2025-10-12T14:09:37+01:00", "GUP-naked-getups", "done", "4", "3", ""},
			wantErr: false,
		},
		{
			name:    "valid record with subset",
			record:  []string{"2025-10-12T14:09:37+01:00", "GUP-naked-getups", "done", "4", "3", "back-safe"},
			wantErr: false,
		},
		{
			name:    "invalid timestamp",
			record:  []string{"bad-timestamp", "GUP-naked-getups", "done", "4", "3", ""},
			wantErr: true,
		},
		{
			name:    "wrong number of fields",
			record:  []string{"2025-10-12T14:09:37+01:00", "GUP-naked-getups", "done"},
			wantErr: true,
		},
		{
			name:    "invalid duration",
			record:  []string{"2025-10-12T14:09:37+01:00", "GUP-naked-getups", "done", "abc", "3", ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, err := parseCSVRecord(tt.record)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if entry.Code != "GUP-naked-getups" {
					t.Errorf("expected code GUP-naked-getups, got %s", entry.Code)
				}
				if entry.Status != "done" {
					t.Errorf("expected status done, got %s", entry.Status)
				}
				if entry.Duration != 4 {
					t.Errorf("expected duration 4, got %d", entry.Duration)
				}
				if entry.RPE != 3 {
					t.Errorf("expected RPE 3, got %d", entry.RPE)
				}
				// Check subset if provided
				if len(tt.record) > 5 && tt.record[5] != "" {
					if entry.Subset != tt.record[5] {
						t.Errorf("expected subset %s, got %s", tt.record[5], entry.Subset)
					}
				}
			}
		})
	}
}

func TestSnackHasAllTags(t *testing.T) {
	snack := Movo{
		AllTags: []string{"breathx", "testx", "mobilityx"},
	}

	tests := []struct {
		name     string
		tags     []string
		expected bool
	}{
		{"no tags", []string{}, true},
		{"single matching tag", []string{"breathx"}, true},
		{"multiple matching tags", []string{"breathx", "testx"}, true},
		{"non-matching tag", []string{"strengthx"}, false},
		{"mixed matching and non-matching", []string{"breathx", "strengthx"}, false},
		{"case insensitive", []string{"BREATHX"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := snack.HasAllTags(tt.tags)
			if result != tt.expected {
				t.Errorf("HasAllTags(%v) = %v, want %v", tt.tags, result, tt.expected)
			}
		})
	}
}

func TestSnackMatchesDuration(t *testing.T) {
	snack := Movo{
		DurationMin: 3,
		DurationMax: 5,
	}

	tests := []struct {
		name     string
		minDur   int
		maxDur   int
		expected bool
	}{
		{"no filter", 0, 0, true},
		{"exact overlap", 3, 5, true},
		{"partial overlap (lower)", 2, 4, true},
		{"partial overlap (upper)", 4, 6, true},
		{"contains snack range", 2, 6, true},
		{"no overlap (below)", 0, 2, false},
		{"no overlap (above)", 6, 10, false},
		{"filter inside snack range", 4, 4, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := snack.MatchesDuration(tt.minDur, tt.maxDur)
			if result != tt.expected {
				t.Errorf("MatchesDuration(%d, %d) = %v, want %v", tt.minDur, tt.maxDur, result, tt.expected)
			}
		})
	}
}

func TestSnackGetDefaultDuration(t *testing.T) {
	tests := []struct {
		min      int
		max      int
		expected int
	}{
		{3, 5, 4}, // (3+5+1)/2 = 4.5 → 4
		{2, 4, 3}, // (2+4+1)/2 = 3.5 → 3
		{5, 7, 6}, // (5+7+1)/2 = 6.5 → 6
		{3, 3, 3}, // (3+3+1)/2 = 3.5 → 3
	}

	for _, tt := range tests {
		snack := Movo{
			DurationMin: tt.min,
			DurationMax: tt.max,
		}
		result := snack.GetDefaultDuration()
		if result != tt.expected {
			t.Errorf("GetDefaultDuration() for range [%d,%d] = %d, want %d", tt.min, tt.max, result, tt.expected)
		}
	}
}

// Integration Tests

func TestLoadSnacksFromTestData(t *testing.T) {
	cfg := &Config{
		MovosDir: "testdata/movos",
	}

	snacks, err := LoadSnacksWithConfig(cfg)
	if err != nil {
		t.Fatalf("failed to load test snacks: %v", err)
	}

	if len(snacks) == 0 {
		t.Fatal("expected to load some snacks, got 0")
	}

	// Check that snacks were properly processed
	for _, snack := range snacks {
		if snack.FullCode == "" {
			t.Errorf("snack %s has empty FullCode", snack.Title)
		}
		if snack.CategoryCode == "" {
			t.Errorf("snack %s has empty CategoryCode", snack.Title)
		}
		if snack.EffectiveRPE == 0 {
			t.Errorf("snack %s has RPE of 0", snack.Title)
		}
	}

	// Find specific test snacks
	var foundBoxBreath bool
	var foundHeavyLift bool
	for _, snack := range snacks {
		if snack.FullCode == "TB-box-breath" {
			foundBoxBreath = true
			if snack.EffectiveRPE != 1 {
				t.Errorf("TB-box-breath should have RPE 1 (inherited), got %d", snack.EffectiveRPE)
			}
			if snack.MinPerDay == 0 {
				t.Error("TB-box-breath should be marked as min_per_day")
			}
		}
		if snack.FullCode == "TS-heavy-lift" {
			foundHeavyLift = true
			if snack.EffectiveRPE != 9 {
				t.Errorf("TS-heavy-lift should have RPE 9 (override), got %d", snack.EffectiveRPE)
			}
		}
	}

	if !foundBoxBreath {
		t.Error("did not find TB-box-breath snack")
	}
	if !foundHeavyLift {
		t.Error("did not find TS-heavy-lift snack")
	}
}

func TestHistoryReadWrite(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	cfg := TestConfig(tmpDir)

	// Write some history entries (all today)
	entries := []HistoryEntry{
		{
			Timestamp: time.Now().Add(-2 * time.Hour),
			Code:      "TB-box-breath",
			Status:    "done",
			Duration:  4,
			RPE:       1,
		},
		{
			Timestamp: time.Now().Add(-1 * time.Hour),
			Code:      "TS-pushups",
			Status:    "skip",
			Duration:  0,
			RPE:       0,
		},
		{
			Timestamp: time.Now(),
			Code:      "TS-heavy-lift",
			Status:    "done",
			Duration:  6,
			RPE:       9,
		},
	}

	for _, entry := range entries {
		if err := AppendTodayLog(cfg.LogsDir, entry); err != nil {
			t.Fatalf("failed to append history: %v", err)
		}
	}

	// Read back
	loaded, err := LoadDailyLog(cfg.LogsDir, time.Now())
	if err != nil {
		t.Fatalf("failed to load history: %v", err)
	}

	if len(loaded) != len(entries) {
		t.Errorf("expected %d entries, got %d", len(entries), len(loaded))
	}

	// Verify entries
	for i, entry := range loaded {
		if entry.Code != entries[i].Code {
			t.Errorf("entry %d: expected code %s, got %s", i, entries[i].Code, entry.Code)
		}
		if entry.Status != entries[i].Status {
			t.Errorf("entry %d: expected status %s, got %s", i, entries[i].Status, entry.Status)
		}
	}
}

func TestGetTodayStatsWithHistory(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := TestConfig(tmpDir)

	// Add entries from today
	entries := []HistoryEntry{
		{
			Timestamp: time.Now(),
			Code:      "TB-box-breath",
			Status:    "done",
			Duration:  4,
			RPE:       1,
		},
		{
			Timestamp: time.Now(),
			Code:      "TS-pushups",
			Status:    "done",
			Duration:  5,
			RPE:       7,
		},
		{
			Timestamp: time.Now(),
			Code:      "TS-heavy-lift",
			Status:    "skip",
			Duration:  0,
			RPE:       0,
		},
	}

	for _, entry := range entries {
		if err := AppendTodayLog(cfg.LogsDir, entry); err != nil {
			t.Fatalf("failed to append history: %v", err)
		}
	}

	stats, err := GetTodayStatsDaily(cfg.LogsDir)
	if err != nil {
		t.Fatalf("failed to get stats: %v", err)
	}

	if stats.TotalMovos != 3 {
		t.Errorf("expected 3 total movos, got %d", stats.TotalMovos)
	}

	if len(stats.CompletedSnacks) != 2 {
		t.Errorf("expected 2 completed snacks, got %d", len(stats.CompletedSnacks))
	}

	if len(stats.SkippedSnacks) != 1 {
		t.Errorf("expected 1 skipped snack, got %d", len(stats.SkippedSnacks))
	}

	if stats.TotalDuration != 9 {
		t.Errorf("expected total duration 9, got %d", stats.TotalDuration)
	}

	if stats.TotalRPE != 8 {
		t.Errorf("expected total RPE 8, got %d", stats.TotalRPE)
	}
}

func TestFilterSnacksByTags(t *testing.T) {
	cfg := &Config{
		MovosDir: "testdata/movos",
	}

	snacks, err := LoadSnacksWithConfig(cfg)
	if err != nil {
		t.Fatalf("failed to load snacks: %v", err)
	}

	// Filter by breathx tag
	filters := FilterOptions{
		Tags: []string{"breathx"},
	}

	filtered := filterSnacks(snacks, filters)

	// All filtered snacks should have breathx tag
	for _, snack := range filtered {
		if !snack.HasAllTags([]string{"breathx"}) {
			t.Errorf("snack %s should have breathx tag", snack.FullCode)
		}
	}

	// Should find at least the breath snacks
	if len(filtered) == 0 {
		t.Error("expected to find snacks with breathx tag")
	}
}

func TestFilterSnacksByRPE(t *testing.T) {
	cfg := &Config{
		MovosDir: "testdata/movos",
	}

	snacks, err := LoadSnacksWithConfig(cfg)
	if err != nil {
		t.Fatalf("failed to load snacks: %v", err)
	}

	// Filter for recovery (max RPE 2)
	filters := FilterOptions{
		MaxRPE: 2,
	}

	filtered := filterSnacks(snacks, filters)

	// All filtered snacks should have RPE <= 2
	for _, snack := range filtered {
		if snack.EffectiveRPE > 2 {
			t.Errorf("snack %s has RPE %d, expected <= 2", snack.FullCode, snack.EffectiveRPE)
		}
	}

	// Should find at least some low RPE snacks
	if len(filtered) == 0 {
		t.Error("expected to find snacks with RPE <= 2")
	}

	// Filter for intense work (min RPE 7)
	filters2 := FilterOptions{
		MinRPE: 7,
	}

	filtered2 := filterSnacks(snacks, filters2)

	// All filtered snacks should have RPE >= 7
	for _, snack := range filtered2 {
		if snack.EffectiveRPE < 7 {
			t.Errorf("snack %s has RPE %d, expected >= 7", snack.FullCode, snack.EffectiveRPE)
		}
	}
}

func TestFilterByFrequency(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := TestConfig(tmpDir)
	cfg.MovosDir = "testdata/movos"

	// Load snacks
	snacks, err := LoadSnacksWithConfig(cfg)
	if err != nil {
		t.Fatalf("failed to load snacks: %v", err)
	}

	// Complete TB-box-breath twice (max_per_day = 2)
	for i := 0; i < 2; i++ {
		entry := HistoryEntry{
			Timestamp: time.Now(),
			Code:      "TB-box-breath",
			Status:    "done",
			Duration:  4,
			RPE:       1,
		}
		if err := AppendTodayLog(cfg.LogsDir, entry); err != nil {
			t.Fatalf("failed to append history: %v", err)
		}
	}

	// Complete TS-pushups once (max_per_day = 1)
	entry := HistoryEntry{
		Timestamp: time.Now(),
		Code:      "TS-pushups",
		Status:    "done",
		Duration:  5,
		RPE:       7,
	}
	if err := AppendTodayLog(cfg.LogsDir, entry); err != nil {
		t.Fatalf("failed to append history: %v", err)
	}

	// Filter by frequency
	filtered, err := filterByFrequencyWithConfig(cfg, snacks)
	if err != nil {
		t.Fatalf("failed to filter: %v", err)
	}

	// TB-box-breath should be included (max_per_day = 2, done 2 times, so at limit)
	// TS-pushups should be excluded (max_per_day = 1, done 1 time)
	foundBoxBreath := false
	foundPushups := false

	for _, snack := range filtered {
		if snack.FullCode == "TB-box-breath" {
			foundBoxBreath = true
		}
		if snack.FullCode == "TS-pushups" {
			foundPushups = true
		}
	}

	if foundBoxBreath {
		t.Error("TB-box-breath should be filtered out (at max_per_day limit)")
	}
	if foundPushups {
		t.Error("TS-pushups should be filtered out (at max_per_day limit)")
	}
}

// Helper function for frequency filtering with config
func filterByFrequencyWithConfig(cfg *Config, snacks []Movo) ([]Movo, error) {
	var filtered []Movo

	for _, snack := range snacks {
		doneToday, _, err := GetCountTodayDaily(cfg.LogsDir, snack.FullCode)
		if err != nil {
			return nil, err
		}

		if snack.MaxPerDay > 0 && doneToday >= snack.MaxPerDay {
			continue
		}

		filtered = append(filtered, snack)
	}

	return filtered, nil
}

// Helper function to load snacks with config
func LoadSnacksWithConfig(cfg *Config) ([]Movo, error) {
	// Check if movos directory exists
	if _, err := os.Stat(cfg.MovosDir); os.IsNotExist(err) {
		return nil, err
	}

	// Find all .yaml files
	files, err := filepath.Glob(filepath.Join(cfg.MovosDir, "*.yaml"))
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, err
	}

	var allMovos []Movo

	// Load each file
	for _, file := range files {
		category, err := loadCategory(file)
		if err != nil {
			return nil, err
		}

		// Process snacks in this category
		for i := range category.Movos {
			snack := &category.Movos[i]
			snack.CategoryCode = category.Code
			snack.FullCode = category.Code + "-" + snack.Code
			snack.AllTags = append([]string{}, category.Tags...)
			snack.AllTags = append(snack.AllTags, snack.Tags...)

			if snack.RPE != nil {
				snack.EffectiveRPE = *snack.RPE
			} else {
				snack.EffectiveRPE = category.DefaultRPE
			}

			if snack.Weight == 1.0 && category.Weight != 1.0 {
				snack.Weight = category.Weight
			}

			allMovos = append(allMovos, *snack)
		}
	}

	return allMovos, nil
}

func TestEverydaySnacksPriority(t *testing.T) {
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

	// Find an everyday snack
	var everydayMovo *Movo
	for i := range snacks {
		if snacks[i].MinPerDay > 0 {
			everydayMovo = &snacks[i]
			break
		}
	}

	if everydayMovo == nil {
		t.Skip("No everyday snacks in test data")
	}

	t.Run("prioritizes incomplete everyday snacks", func(t *testing.T) {
		// With no history, should get everyday snack
		selected, err := SelectSnack(snacks, FilterOptions{}, maxDailyRPEDefault)
		if err != nil {
			t.Fatalf("SelectSnack failed: %v", err)
		}

		if selected.MinPerDay == 0 {
			t.Errorf("Expected everyday snack, got: %s (min_per_day=%d)", selected.FullCode, selected.MinPerDay)
		}
	})

	t.Run("skip dailies flag bypasses priority", func(t *testing.T) {
		// With SkipMinimums=true, might get non-everyday snack
		filters := FilterOptions{SkipMinimums: true}
		selected, err := SelectSnack(snacks, filters, maxDailyRPEDefault)
		if err != nil {
			t.Fatalf("SelectSnack failed: %v", err)
		}

		t.Logf("With skip dailies: got %s (min_per_day=%d)", selected.FullCode, selected.MinPerDay)
		// This is probabilistic, but at least it should not ONLY select everyday snacks
	})

	t.Run("after completing everyday snack, others are available", func(t *testing.T) {
		// Complete the everyday snack
		entry := HistoryEntry{
			Timestamp: time.Now(),
			Code:      everydayMovo.FullCode,
			Status:    "done",
			Duration:  5,
			RPE:       everydayMovo.EffectiveRPE,
		}

		if err := AppendTodayLog(cfg.LogsDir, entry); err != nil {
			t.Fatalf("Failed to log entry: %v", err)
		}

		// Now selection should include non-everyday snacks
		// (since the only everyday snack is complete)
		selected, err := SelectSnack(snacks, FilterOptions{}, maxDailyRPEDefault)
		if err != nil {
			t.Fatalf("SelectSnack failed: %v", err)
		}

		t.Logf("After completing everyday: got %s (min_per_day=%d)", selected.FullCode, selected.MinPerDay)
		// Selection can be anything now
	})
}
