package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"
)

// GetDailyLogPath returns the path for a specific date's log file
func GetDailyLogPath(logsDir string, date time.Time) string {
	filename := date.Format("20060102") + ".csv"
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

// LoadDailyLog loads entries from a specific daily log file (CSV format)
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

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		// If CSV parsing fails, check if it's old format and provide helpful error
		return nil, fmt.Errorf("⚠️  Error reading log file. If this is an old format log, run 'movodoro migrate' to convert to v1.0.0 CSV format: %w", err)
	}

	var entries []HistoryEntry

	for i, record := range records {
		// Skip header row
		if i == 0 && record[0] == "timestamp" {
			continue
		}

		entry, err := parseCSVRecord(record)
		if err != nil {
			// Skip invalid entries but continue processing
			continue
		}

		entries = append(entries, entry)
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

	// Find all .csv files
	pattern := filepath.Join(logsDir, "*.csv")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("error finding log files: %w", err)
	}

	if len(files) == 0 {
		return []HistoryEntry{}, nil
	}

	// Sort files (they're named YYYYMMDD.csv so alphabetical = chronological)
	sort.Strings(files)

	var allEntries []HistoryEntry

	for _, filePath := range files {
		f, err := os.Open(filePath)
		if err != nil {
			continue
		}

		reader := csv.NewReader(f)
		records, err := reader.ReadAll()
		f.Close()

		if err != nil {
			// Skip files that can't be parsed as CSV
			continue
		}

		for i, record := range records {
			// Skip header row
			if i == 0 && len(record) > 0 && record[0] == "timestamp" {
				continue
			}

			entry, err := parseCSVRecord(record)
			if err != nil {
				continue
			}

			allEntries = append(allEntries, entry)
		}
	}

	return allEntries, nil
}

// AppendTodayLog appends an entry to today's log file in CSV format
func AppendTodayLog(logsDir string, entry HistoryEntry) error {
	// Ensure logs directory exists
	if err := ensureLogsDir(logsDir); err != nil {
		return err
	}

	logPath := GetTodayLogPath(logsDir)

	// Check if file exists and is empty (need to write header)
	fileInfo, err := os.Stat(logPath)
	writeHeader := err != nil || fileInfo.Size() == 0

	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening log file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header if this is a new/empty file
	if writeHeader {
		if err := writer.Write([]string{"timestamp", "code", "status", "duration", "rpe", "subset"}); err != nil {
			return fmt.Errorf("error writing CSV header: %w", err)
		}
	}

	// Write the entry
	record := []string{
		entry.Timestamp.Format(time.RFC3339),
		entry.Code,
		entry.Status,
		strconv.Itoa(entry.Duration),
		strconv.Itoa(entry.RPE),
		entry.Subset,
	}

	if err := writer.Write(record); err != nil {
		return fmt.Errorf("error writing CSV record: %w", err)
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
		stats.TotalMovos++

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

// parseCSVRecord parses a CSV record: timestamp,code,status,duration,rpe,subset
func parseCSVRecord(record []string) (HistoryEntry, error) {
	if len(record) != 6 {
		return HistoryEntry{}, fmt.Errorf("expected 6 fields, got %d", len(record))
	}

	// Parse timestamp
	timestamp, err := time.Parse(time.RFC3339, record[0])
	if err != nil {
		return HistoryEntry{}, fmt.Errorf("invalid timestamp: %w", err)
	}

	// Parse duration
	duration, err := strconv.Atoi(record[3])
	if err != nil {
		return HistoryEntry{}, fmt.Errorf("invalid duration: %w", err)
	}

	// Parse RPE
	rpe, err := strconv.Atoi(record[4])
	if err != nil {
		return HistoryEntry{}, fmt.Errorf("invalid RPE: %w", err)
	}

	return HistoryEntry{
		Timestamp: timestamp,
		Code:      record[1],
		Status:    record[2],
		Duration:  duration,
		RPE:       rpe,
		Subset:    record[5],
	}, nil
}
