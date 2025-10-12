package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadSnacks loads all snack definitions from YAML files in the movos directory
func LoadSnacks() ([]Snack, error) {
	cfg := DefaultConfig()
	movosDir := cfg.MovosDir

	// Check if movos directory exists
	if _, err := os.Stat(movosDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("movos directory not found: %s", movosDir)
	}

	// Find all .yaml files
	files, err := filepath.Glob(filepath.Join(movosDir, "*.yaml"))
	if err != nil {
		return nil, fmt.Errorf("error finding YAML files: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no YAML files found in movos directory")
	}

	var allSnacks []Snack

	// Load each file
	for _, file := range files {
		category, err := loadCategory(file)
		if err != nil {
			return nil, fmt.Errorf("error loading %s: %w", file, err)
		}

		// Process snacks in this category
		for i := range category.Snacks {
			snack := &category.Snacks[i]

			// Set category code
			snack.CategoryCode = category.Code

			// Set full code
			snack.FullCode = fmt.Sprintf("%s-%s", category.Code, snack.Code)

			// Combine tags (category tags + snack tags)
			snack.AllTags = append([]string{}, category.Tags...)
			snack.AllTags = append(snack.AllTags, snack.Tags...)

			// Set effective RPE (use snack RPE if set, otherwise use category default)
			if snack.RPE != nil {
				snack.EffectiveRPE = *snack.RPE
			} else {
				snack.EffectiveRPE = category.DefaultRPE
			}

			// Apply category weight if snack weight is 1.0 (i.e., not customized)
			if snack.Weight == 1.0 && category.Weight != 1.0 {
				snack.Weight = category.Weight
			}

			allSnacks = append(allSnacks, *snack)
		}
	}

	return allSnacks, nil
}

func loadCategory(filepath string) (*Category, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var category Category
	if err := yaml.Unmarshal(data, &category); err != nil {
		return nil, err
	}

	return &category, nil
}

// HasAllTags checks if snack has all specified tags
func (s *Snack) HasAllTags(tags []string) bool {
	if len(tags) == 0 {
		return true
	}

	snackTagSet := make(map[string]bool)
	for _, tag := range s.AllTags {
		snackTagSet[strings.ToLower(tag)] = true
	}

	for _, requiredTag := range tags {
		if !snackTagSet[strings.ToLower(requiredTag)] {
			return false
		}
	}

	return true
}

// MatchesDuration checks if snack duration overlaps with filter
func (s *Snack) MatchesDuration(minDur, maxDur int) bool {
	if minDur == 0 && maxDur == 0 {
		return true
	}

	// Check for overlap: snack range [s.Min, s.Max] overlaps with filter [minDur, maxDur]
	return s.DurationMax >= minDur && s.DurationMin <= maxDur
}

// GetDefaultDuration returns the middle of the duration range, rounded up
func (s *Snack) GetDefaultDuration() int {
	total := s.DurationMin + s.DurationMax
	// Round up: (total + 1) / 2
	return (total + 1) / 2
}
