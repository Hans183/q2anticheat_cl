package database

import (
	"testing"
	"time"
)

func TestPlayerNameFiltering(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// 1. Insert screenshots
	ss1 := &ScreenshotRecord{
		ServerAddr: "127.0.0.1:27910",
		PlayerIP:   "192.168.1.50",
		PlayerName: "SniperWolf",
		ClientID:   1,
		Width:      800,
		Height:     600,
		Format:     "webp",
		FilePath:   "/tmp/ss1.webp",
		FileSize:   1024,
		Timestamp:  time.Now(),
	}
	ss2 := &ScreenshotRecord{
		ServerAddr: "127.0.0.1:27910",
		PlayerIP:   "192.168.1.51",
		PlayerName: "Shadow[TAG]",
		ClientID:   2,
		Width:      800,
		Height:     600,
		Format:     "webp",
		FilePath:   "/tmp/ss2.webp",
		FileSize:   1024,
		Timestamp:  time.Now(),
	}
	db.InsertScreenshot(ss1)
	db.InsertScreenshot(ss2)

	// Partial match on "Sniper"
	results, count, err := db.GetScreenshots("", "Sniper", "", "", false, 1, 20)
	if err != nil {
		t.Fatalf("GetScreenshots error: %v", err)
	}
	if count != 1 || len(results) != 1 || results[0].PlayerName != "SniperWolf" {
		t.Fatalf("expected 1 result for 'Sniper', got %d", count)
	}

	// Partial case-insensitive match on "shadow"
	results, count, err = db.GetScreenshots("", "shadow", "", "", false, 1, 20)
	if err != nil {
		t.Fatalf("GetScreenshots error: %v", err)
	}
	if count != 1 || len(results) != 1 || results[0].PlayerName != "Shadow[TAG]" {
		t.Fatalf("expected 1 result for 'shadow', got %d", count)
	}

	// 2. Insert violations
	v1 := &ViolationRecord{
		ServerAddr: "127.0.0.1:27910",
		PlayerIP:   "192.168.1.50",
		PlayerName: "SniperWolf",
		ClientID:   1,
		Type:       "cvar",
		Reason:     "invalid r_drawentities",
		Timestamp:  time.Now(),
	}
	v2 := &ViolationRecord{
		ServerAddr: "127.0.0.1:27910",
		PlayerIP:   "192.168.1.52",
		PlayerName: "MegaKiller99",
		ClientID:   3,
		Type:       "file",
		Reason:     "modified pak0.pak",
		Timestamp:  time.Now(),
	}
	db.InsertViolation(v1)
	db.InsertViolation(v2)

	// Partial match on "Killer"
	vResults, vCount, err := db.GetViolations("", "Killer", "", "", "", 1, 20)
	if err != nil {
		t.Fatalf("GetViolations error: %v", err)
	}
	if vCount != 1 || len(vResults) != 1 || vResults[0].PlayerName != "MegaKiller99" {
		t.Fatalf("expected 1 result for 'Killer', got %d", vCount)
	}

	// 3. Insert process snapshots
	db.InsertProcessSnapshot("127.0.0.1:27910", "192.168.1.50", "SniperWolf", 1, 10, 5, "", "[]", "[]")
	db.InsertProcessSnapshot("127.0.0.1:27910", "192.168.1.53", "ProGamer", 4, 15, 8, "", "[]", "[]")

	// Exact/Partial match on "ProGamer"
	psResults, psCount, err := db.GetProcessSnapshots("", "ProGamer", "", "", 1, 20)
	if err != nil {
		t.Fatalf("GetProcessSnapshots error: %v", err)
	}
	if psCount != 1 || len(psResults) != 1 || psResults[0].PlayerName != "ProGamer" {
		t.Fatalf("expected 1 result for 'ProGamer', got %d", psCount)
	}
}
