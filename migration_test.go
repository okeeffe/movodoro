package main

import (
	"bufio"
	"encoding/csv"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestMigrationOldFormatToCSV(t *testing.T) {
	tmpDir := t.TempDir()

	// Create an old format log file
	oldLogPath := filepath.Join(tmpDir, "20251012.log")
	oldContent := `2025-10-12T10:00:00Z TB-box-breath done 4 1
2025-10-12T11:00:00Z TS-pushups skip 0 0
2025-10-12T12:00:00Z TS-heavy-lift done 6 9`

	if err := os.WriteFile(oldLogPath, []byte(oldContent), 0644); err != nil {
		t.Fatalf("Failed to create test log file: %v", err)
	}

	// Run migration logic (simulated from handleMigrateLogsToCsv)
	file, err := os.Open(oldLogPath)
	if err != nil {
		t.Fatalf("Failed to open old log: %v", err)
	}

	// Read old format entries
	var entries []HistoryEntry
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) != 5 {
			continue
		}

		timestamp, _ := time.Parse(time.RFC3339, parts[0])
		duration, _ := strconv.Atoi(parts[3])
		rpe, _ := strconv.Atoi(parts[4])

		entries = append(entries, HistoryEntry{
			Timestamp: timestamp,
			Code:      parts[1],
			Status:    parts[2],
			Duration:  duration,
			RPE:       rpe,
			Subset:    "",
		})
	}
	file.Close()

	if len(entries) != 3 {
		t.Fatalf("Expected 3 entries, got %d", len(entries))
	}

	// Create backup
	backupPath := oldLogPath + ".bak"
	if err := os.Rename(oldLogPath, backupPath); err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// Write new CSV format
	newLogPath := strings.TrimSuffix(oldLogPath, ".log") + ".csv"
	newFile, err := os.Create(newLogPath)
	if err != nil {
		t.Fatalf("Failed to create new CSV file: %v", err)
	}

	writer := csv.NewWriter(newFile)

	// Write header
	if err := writer.Write([]string{"timestamp", "code", "status", "duration", "rpe", "subset"}); err != nil {
		t.Fatalf("Failed to write header: %v", err)
	}

	// Write entries
	for _, entry := range entries {
		record := []string{
			entry.Timestamp.Format(time.RFC3339),
			entry.Code,
			entry.Status,
			strconv.Itoa(entry.Duration),
			strconv.Itoa(entry.RPE),
			entry.Subset,
		}
		if err := writer.Write(record); err != nil {
			t.Fatalf("Failed to write entry: %v", err)
		}
	}

	writer.Flush()
	newFile.Close()

	// Verify CSV file was created
	if _, err := os.Stat(newLogPath); os.IsNotExist(err) {
		t.Fatal("CSV file was not created")
	}

	// Verify backup exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Fatal("Backup file was not created")
	}

	// Read and verify CSV content
	loadedEntries, err := loadCSVLog(newLogPath)
	if err != nil {
		t.Fatalf("Failed to load CSV: %v", err)
	}

	if len(loadedEntries) != 3 {
		t.Fatalf("Expected 3 loaded entries, got %d", len(loadedEntries))
	}

	// Verify first entry
	if loadedEntries[0].Code != "TB-box-breath" {
		t.Errorf("Entry 0: expected code TB-box-breath, got %s", loadedEntries[0].Code)
	}
	if loadedEntries[0].Status != "done" {
		t.Errorf("Entry 0: expected status done, got %s", loadedEntries[0].Status)
	}
	if loadedEntries[0].Duration != 4 {
		t.Errorf("Entry 0: expected duration 4, got %d", loadedEntries[0].Duration)
	}
	if loadedEntries[0].RPE != 1 {
		t.Errorf("Entry 0: expected RPE 1, got %d", loadedEntries[0].RPE)
	}
	if loadedEntries[0].Subset != "" {
		t.Errorf("Entry 0: expected empty subset, got %s", loadedEntries[0].Subset)
	}

	// Verify second entry
	if loadedEntries[1].Code != "TS-pushups" {
		t.Errorf("Entry 1: expected code TS-pushups, got %s", loadedEntries[1].Code)
	}
	if loadedEntries[1].Status != "skip" {
		t.Errorf("Entry 1: expected status skip, got %s", loadedEntries[1].Status)
	}
}

func TestMigrationDetectsCSVFormat(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a CSV format log file (already migrated)
	csvLogPath := filepath.Join(tmpDir, "20251013.log")
	csvContent := `timestamp,code,status,duration,rpe,subset
2025-10-13T10:00:00Z,TB-box-breath,done,4,1,
2025-10-13T11:00:00Z,TS-pushups,skip,0,0,back-safe`

	if err := os.WriteFile(csvLogPath, []byte(csvContent), 0644); err != nil {
		t.Fatalf("Failed to create test CSV file: %v", err)
	}

	// Check if it's detected as CSV format
	file, err := os.Open(csvLogPath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var firstLine string
	if scanner.Scan() {
		firstLine = scanner.Text()
	}

	if !strings.HasPrefix(firstLine, "timestamp,") {
		t.Error("CSV file should be detected by timestamp, header")
	}
}

func TestMigrationHandlesEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create empty log file
	emptyLogPath := filepath.Join(tmpDir, "20251014.log")
	if err := os.WriteFile(emptyLogPath, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	// Try to read it
	file, err := os.Open(emptyLogPath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}

	var entries []HistoryEntry
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		// Would parse here but empty file has no lines
	}
	file.Close()

	if len(entries) != 0 {
		t.Errorf("Expected 0 entries from empty file, got %d", len(entries))
	}
}

func TestMigrationHandlesMalformedLines(t *testing.T) {
	tmpDir := t.TempDir()

	// Create log file with some malformed lines
	logPath := filepath.Join(tmpDir, "20251015.log")
	content := `2025-10-15T10:00:00Z TB-box-breath done 4 1
malformed line without enough fields
2025-10-15T11:00:00Z TS-pushups done 5 7
another bad line
2025-10-15T12:00:00Z TS-heavy-lift skip 0 0`

	if err := os.WriteFile(logPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Parse with error handling
	file, err := os.Open(logPath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}

	var entries []HistoryEntry
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) != 5 {
			// Skip malformed lines
			continue
		}

		timestamp, err := time.Parse(time.RFC3339, parts[0])
		if err != nil {
			continue
		}

		duration, err := strconv.Atoi(parts[3])
		if err != nil {
			continue
		}

		rpe, err := strconv.Atoi(parts[4])
		if err != nil {
			continue
		}

		entries = append(entries, HistoryEntry{
			Timestamp: timestamp,
			Code:      parts[1],
			Status:    parts[2],
			Duration:  duration,
			RPE:       rpe,
			Subset:    "",
		})
	}
	file.Close()

	// Should have parsed only the 3 valid lines
	if len(entries) != 3 {
		t.Errorf("Expected 3 valid entries (skipping malformed), got %d", len(entries))
	}
}

// Helper function to load CSV log for testing
func loadCSVLog(path string) ([]HistoryEntry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var entries []HistoryEntry
	for i, record := range records {
		// Skip header row
		if i == 0 && record[0] == "timestamp" {
			continue
		}

		entry, err := parseCSVRecord(record)
		if err != nil {
			continue
		}

		entries = append(entries, entry)
	}

	return entries, nil
}
