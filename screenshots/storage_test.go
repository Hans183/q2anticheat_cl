package screenshots

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/user/anticheat_cl/database"
)

func TestStorageAndPurge(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "ac-storage-test-*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("create db: %v", err)
	}
	defer db.Close()

	screenshotDir := filepath.Join(tempDir, "screenshots")
	storage, err := NewStorage(screenshotDir, db)
	if err != nil {
		t.Fatalf("create storage: %v", err)
	}

	// Save a screenshot
	data := []byte("dummy-webp-image-data-here")
	savedPath, err := storage.SaveScreenshot("127.0.0.1:27910", "192.168.1.100", "Player1", 1, 800, 600, 1, data)
	if err != nil {
		t.Fatalf("save screenshot: %v", err)
	}

	if _, err := os.Stat(savedPath); err != nil {
		t.Fatalf("saved file does not exist: %v", err)
	}

	count, totalBytes, err := storage.GetDiskStats()
	if err != nil {
		t.Fatalf("disk stats: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 file, got %d", count)
	}
	if totalBytes != int64(len(data)) {
		t.Errorf("expected %d bytes, got %d", len(data), totalBytes)
	}

	// Purge with 0 days (disabled)
	deleted, freed, err := storage.PurgeOlderThan(0)
	if err != nil {
		t.Fatalf("purge with 0 days: %v", err)
	}
	if deleted != 0 || freed != 0 {
		t.Errorf("expected 0 deleted with days=0, got %d, %d", deleted, freed)
	}

	// Update DB timestamp to 40 days ago
	oldTime := time.Now().AddDate(0, 0, -40)
	_, err = db.GetScreenshotsOlderThan(time.Now())
	if err != nil {
		t.Fatalf("get old: %v", err)
	}

	// Manually backdate screenshot record in DB
	_, err = db.DeleteScreenshotsOlderThan(time.Now().Add(1 * time.Hour))
	if err != nil {
		t.Fatalf("delete: %v", err)
	}

	// Re-insert with old timestamp
	rec := &database.ScreenshotRecord{
		ServerAddr: "127.0.0.1:27910",
		PlayerIP:   "192.168.1.100",
		PlayerName: "Player1",
		Width:      800,
		Height:     600,
		Format:     "webp",
		FilePath:   savedPath,
		FileSize:   int64(len(data)),
		Timestamp:  oldTime,
	}
	if _, err := db.InsertScreenshot(rec); err != nil {
		t.Fatalf("re-insert: %v", err)
	}

	// Now purge older than 30 days
	deleted, freed, err = storage.PurgeOlderThan(30)
	if err != nil {
		t.Fatalf("purge: %v", err)
	}
	if deleted != 1 {
		t.Errorf("expected 1 deleted record, got %d", deleted)
	}
	if freed != int64(len(data)) {
		t.Errorf("expected %d bytes freed, got %d", len(data), freed)
	}

	// File should be gone
	if _, err := os.Stat(savedPath); !os.IsNotExist(err) {
		t.Errorf("expected file to be deleted, but it still exists")
	}
}
