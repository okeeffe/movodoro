package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// GetDailyLogPath returns the path for a specific date's log file
func GetDailyLogPath(logsDir string, date time.Time) string {
	filename := date.Format("20060102") + ".log"
	return filepath.Join(logsDir, filename)
}

// GetTodayLogPath returns the path for today's log file
func GetTodayLogPath(logsDir string) string {
	return GetDailyLogPath(logsDir, time.Now())
}

// ensureLogsDir creates the logs directory if it doesn't exist
func ensureLogsDir(logsDir string) error {
	return os.MkdirAll(logsDir, 0755)
}

// LoadDailyLog loads entries from a specific daily log file
func LoadDailyLog(logsDir string, date time.Time) ([]HistoryEntry, error) {
	logPath := GetDailyLogPath(logsDir, date)

	file, err := os.Open(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []HistoryEntry{}, nil
		}
		return nil, fmt.Errorf("error opening log file: %w", err)
	}
	defer file.Close()

	var entries []HistoryEntry
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		entry, err := parseHistoryLine(line)
		if err != nil {
			continue
		}

		entries = append(entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading log file: %w", err)
	}

	return entries, nil
}

// LoadHistoryRange loads entries from a date range (inclusive)
func LoadHistoryRange(logsDir string, startDate, endDate time.Time) ([]HistoryEntry, error) {
	var allEntries []HistoryEntry

	// Normalize dates to midnight
	start := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
	end := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 0, 0, 0, 0, endDate.Location())

	// Iterate through each day
	for date := start; !date.After(end); date = date.AddDate(0, 0, 1) {
		entries, err := LoadDailyLog(logsDir, date)
		if err != nil {
			return nil, err
		}
		allEntries = append(allEntries, entries...)
	}

	return allEntries, nil
}

// LoadAllHistory loads all history entries from all log files
func LoadAllHistory(logsDir string) ([]HistoryEntry, error) {
	// Ensure logs directory exists
	if err := ensureLogsDir(logsDir); err != nil {
		return nil, err
	}

	// Find all .log files
	pattern := filepath.Join(logsDir, "*.log")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("error finding log files: %w", err)
	}

	if len(files) == 0 {
		return []HistoryEntry{}, nil
	}

	// Sort files (they're named YYYYMMDD.log so alphabetical = chronological)
	sort.Strings(files)

	var allEntries []HistoryEntry

	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}

			entry, err := parseHistoryLine(line)
			if err != nil {
				continue
			}

			allEntries = append(allEntries, entry)
		}

		f.Close()
	}

	return allEntries, nil
}

// AppendTodayLog appends an entry to today's log file
func AppendTodayLog(logsDir string, entry HistoryEntry) error {
	// Ensure logs directory exists
	if err := ensureLogsDir(logsDir); err != nil {
		return err
	}

	logPath := GetTodayLogPath(logsDir)

	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening log file: %w", err)
	}
	defer file.Close()

	line := fmt.Sprintf("%s %s %s %d %d\n",
		entry.Timestamp.Format(time.RFC3339),
		entry.Code,
		entry.Status,
		entry.Duration,
		entry.RPE,
	)

	if _, err := file.WriteString(line); err != nil {
		return fmt.Errorf("error writing to log file: %w", err)
	}

	return nil
}

// GetTodayStatsDaily returns today's stats (optimized for daily files)
func GetTodayStatsDaily(logsDir string) (DailyStats, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	entries, err := LoadDailyLog(logsDir, today)
	if err != nil {
		return DailyStats{}, err
	}

	stats := DailyStats{
		Date: today,
	}

	for _, entry := range entries {
		stats.TotalSnacks++

		if entry.Status == "done" {
			stats.TotalDuration += entry.Duration
			stats.TotalRPE += entry.RPE
			stats.CompletedSnacks = append(stats.CompletedSnacks, entry)
		} else if entry.Status == "skip" {
			stats.SkippedSnacks = append(stats.SkippedSnacks, entry)
		}
	}

	return stats, nil
}

// GetCountTodayDaily returns today's counts for a specific code
func GetCountTodayDaily(logsDir string, code string) (done int, skipped int, err error) {
	entries, err := LoadDailyLog(logsDir, time.Now())
	if err != nil {
		return 0, 0, err
	}

	for _, entry := range entries {
		if entry.Code == code {
			if entry.Status == "done" {
				done++
			} else if entry.Status == "skip" {
				skipped++
			}
		}
	}

	return done, skipped, nil
}

// GetLastDoneDaily returns when a snack was last completed
func GetLastDoneDaily(logsDir string, code string) (*time.Time, error) {
	// Load all history (we need to scan everything for this)
	entries, err := LoadAllHistory(logsDir)
	if err != nil {
		return nil, err
	}

	// Iterate backwards to find most recent
	for i := len(entries) - 1; i >= 0; i-- {
		entry := entries[i]
		if entry.Code == code && entry.Status == "done" {
			return &entry.Timestamp, nil
		}
	}

	return nil, nil
}

// HasEverBeenDoneDaily checks if a snack has ever been completed
func HasEverBeenDoneDaily(logsDir string, code string) (bool, error) {
	lastDone, err := GetLastDoneDaily(logsDir, code)
	if err != nil {
		return false, err
	}
	return lastDone != nil, nil
}

// ClearTodayLog deletes today's log file
func ClearTodayLog(logsDir string) error {
	logPath := GetTodayLogPath(logsDir)

	err := os.Remove(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Already cleared
		}
		return fmt.Errorf("error removing log file: %w", err)
	}

	return nil
}

// parseHistoryLine parses a single line: TIMESTAMP CODE STATUS DURATION RPE
func parseHistoryLine(line string) (HistoryEntry, error) {
	parts := strings.Fields(line)
	if len(parts) != 5 {
		return HistoryEntry{}, fmt.Errorf("expected 5 fields, got %d", len(parts))
	}

	// Parse timestamp
	timestamp, err := time.Parse(time.RFC3339, parts[0])
	if err != nil {
		return HistoryEntry{}, fmt.Errorf("invalid timestamp: %w", err)
	}

	// Parse duration
	duration, err := strconv.Atoi(parts[3])
	if err != nil {
		return HistoryEntry{}, fmt.Errorf("invalid duration: %w", err)
	}

	// Parse RPE
	rpe, err := strconv.Atoi(parts[4])
	if err != nil {
		return HistoryEntry{}, fmt.Errorf("invalid RPE: %w", err)
	}

	return HistoryEntry{
		Timestamp: timestamp,
		Code:      parts[1],
		Status:    parts[2],
		Duration:  duration,
		RPE:       rpe,
	}, nil
}
