package main

import "time"

// Category represents a category of movement snacks
type Category struct {
	Category   string   `yaml:"category"`
	Code       string   `yaml:"code"`
	Weight     float64  `yaml:"weight"`
	DefaultRPE int      `yaml:"default_rpe"`
	Tags       []string `yaml:"tags"`
	Snacks     []Snack  `yaml:"snacks"`
}

// Snack represents a single movement snack
type Snack struct {
	Code        string   `yaml:"code"`
	Title       string   `yaml:"title"`
	Description string   `yaml:"description"`
	DurationMin int      `yaml:"duration_min"`
	DurationMax int      `yaml:"duration_max"`
	RPE         *int     `yaml:"rpe,omitempty"` // Pointer to distinguish between 0 and unset
	MaxPerDay   int      `yaml:"max_per_day"`
	MaxPerWeek  int      `yaml:"max_per_week,omitempty"`
	Weight      float64  `yaml:"weight"`
	EveryDay    bool     `yaml:"every_day"`
	Tags        []string `yaml:"tags"`

	// Computed fields (not in YAML)
	CategoryCode string  `yaml:"-"`
	FullCode     string  `yaml:"-"`
	AllTags      []string `yaml:"-"`
	EffectiveRPE int     `yaml:"-"`
}

// HistoryEntry represents a single log entry
type HistoryEntry struct {
	Timestamp time.Time
	Code      string
	Status    string // "done" or "skip"
	Duration  int    // actual duration in minutes
	RPE       int    // RPE value
}

// FilterOptions contains all filtering options for snack selection
type FilterOptions struct {
	Tags           []string
	Category       string
	MinDuration    int
	MaxDuration    int
	ExactDuration  int
	MinRPE         int
	MaxRPE         int
}

// DailyStats contains statistics for a given day
type DailyStats struct {
	Date          time.Time
	TotalSnacks   int
	TotalDuration int
	TotalRPE      int
	CompletedSnacks []HistoryEntry
	SkippedSnacks   []HistoryEntry
}
