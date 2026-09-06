package server

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/user/anticheat_cl/database"
)

func TestServerBlacklistManager(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}
	defer func() {
		db.Close()
		os.RemoveAll(dir)
	}()

	bl := NewBlacklist()
	bl.SetDB(db)

	// Check default hardcoded patterns loaded
	matched, _ := bl.CheckProcess("cheatengine.exe")
	if !matched {
		t.Fatalf("expected cheatengine.exe to match blacklist")
	}

	matchedMod, _, _ := bl.CheckModuleWithPath("speedhack.dll", "C:\\games\\speedhack.dll")
	if !matchedMod {
		t.Fatalf("expected speedhack.dll to match blacklist")
	}

	// Add custom pattern
	if err := bl.AddEntry("process", "my_bot.exe", "admin"); err != nil {
		t.Fatalf("failed to add custom pattern: %v", err)
	}

	matchedCustom, pattern := bl.CheckProcess("c:\\bots\\my_bot.exe")
	if !matchedCustom || pattern != "my_bot.exe" {
		t.Fatalf("expected my_bot.exe to match custom pattern")
	}

	// Add custom SHA1 hash
	testHash := "da39a3ee5e6b4b0d3255bfef95601890afd80709"
	if err := bl.AddEntry("sha1", testHash, "admin"); err != nil {
		t.Fatalf("failed to add SHA1 hash entry: %v", err)
	}

	matchedHash, pat, _ := bl.CheckModuleFull("innocent_name.dll", "C:\\games\\innocent_name.dll", testHash)
	if !matchedHash || pat != testHash {
		t.Fatalf("expected innocent_name.dll to match due to SHA1 hash blacklist")
	}

	// Find the entry ID and remove it
	entries := bl.GetAllEntries()
	var botID int64
	for _, e := range entries {
		if e.Pattern == "my_bot.exe" {
			botID = e.ID
			break
		}
	}
	if botID == 0 {
		t.Fatalf("expected to find my_bot.exe in blacklist entries")
	}

	// Remove it
	if err := bl.RemoveEntry(botID); err != nil {
		t.Fatalf("failed to remove entry: %v", err)
	}

	// Verify no longer matches
	matchedRemoved, _ := bl.CheckProcess("my_bot.exe")
	if matchedRemoved {
		t.Fatalf("expected my_bot.exe to no longer match after deletion")
	}
}
