package database

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestDB(t *testing.T) (*DB, func()) {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}
	cleanup := func() {
		db.Close()
		os.RemoveAll(dir)
	}
	return db, cleanup
}

func TestBlacklistCRUD(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Initial seed
	defaults := []BlacklistEntry{
		{Type: "process", Pattern: "cheatengine", Enabled: true},
		{Type: "module", Pattern: "hook.dll", Enabled: true},
	}
	if err := db.EnsureHardcodedEntries(defaults); err != nil {
		t.Fatalf("failed to ensure entries: %v", err)
	}

	entries, err := db.GetBlacklist()
	if err != nil {
		t.Fatalf("failed to get blacklist: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	// Add custom process entry
	if err := db.AddBlacklistEntry("process", " customhack.exe ", "admin"); err != nil {
		t.Fatalf("failed to add entry: %v", err)
	}

	entries, err = db.GetBlacklist()
	if err != nil {
		t.Fatalf("failed to get blacklist: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}

	// Verify pattern normalization
	var customID int64
	var found bool
	for _, e := range entries {
		if e.Pattern == "customhack.exe" {
			found = true
			customID = e.ID
			break
		}
	}
	if !found {
		t.Fatalf("expected normalized pattern 'customhack.exe' to exist")
	}

	// Toggle entry
	if err := db.ToggleBlacklistEntry(customID); err != nil {
		t.Fatalf("failed to toggle entry: %v", err)
	}
	entries, _ = db.GetBlacklist()
	for _, e := range entries {
		if e.ID == customID && e.Enabled != false {
			t.Fatalf("expected entry %d to be disabled", customID)
		}
	}

	// Delete custom entry
	if err := db.RemoveBlacklistEntry(customID); err != nil {
		t.Fatalf("failed to remove custom entry: %v", err)
	}

	entries, _ = db.GetBlacklist()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries left after deleting custom entry, got %d", len(entries))
	}

	// Delete system/hardcoded entry
	var sysID int64
	for _, e := range entries {
		if e.Source == "hardcoded" {
			sysID = e.ID
			break
		}
	}
	if sysID == 0 {
		t.Fatalf("expected to find hardcoded system entry")
	}

	if err := db.RemoveBlacklistEntry(sysID); err != nil {
		t.Fatalf("failed to remove system entry: %v", err)
	}

	entries, _ = db.GetBlacklist()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry left, got %d", len(entries))
	}

	// Ensure re-running EnsureHardcodedEntries does not re-add deleted entries when DB is not empty
	if err := db.EnsureHardcodedEntries(defaults); err != nil {
		t.Fatalf("failed to call ensure entries: %v", err)
	}
	entries, _ = db.GetBlacklist()
	if len(entries) != 1 {
		t.Fatalf("expected still 1 entry (deleted entries should not resurrect), got %d", len(entries))
	}
}
