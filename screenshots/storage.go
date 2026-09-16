package screenshots

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/user/anticheat_cl/database"
)

// Storage handles screenshot file storage
type Storage struct {
	baseDir string
	db      *database.DB
}

// NewStorage creates a new screenshot storage
func NewStorage(baseDir string, db *database.DB) (*Storage, error) {
	// Ensure base directory exists
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("create screenshot dir: %w", err)
	}

	return &Storage{
		baseDir: baseDir,
		db:      db,
	}, nil
}

// DB returns the underlying database connection
func (s *Storage) DB() *database.DB {
	return s.db
}

// SaveScreenshot saves a screenshot to disk and stores metadata in DB
// format: 0=jpeg, 1=webp
func (s *Storage) SaveScreenshot(serverAddr, playerIP, playerName string,
	clientID uint32, width, height int, format byte, data []byte) (string, error) {

	// Generate filename based on date and player
	now := time.Now()
	dateDir := now.Format("2006-01-02")
	timeStr := now.Format("150405")

	// Create date directory
	dirPath := filepath.Join(s.baseDir, dateDir)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return "", fmt.Errorf("create date dir: %w", err)
	}

	// Select extension based on format
	ext := ".webp"
	if format == 0 {
		ext = ".jpg"
	}

	// Generate filename
	filename := fmt.Sprintf("%s_%s%s", playerIP, timeStr, ext)
	filePath := filepath.Join(dirPath, filename)

	// Write image data
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}

	log.Printf("[SCREENSHOT] Saved: %s (%dx%d, format=%d, %d bytes)", filePath, width, height, format, len(data))

	// Store metadata in database
	formatStr := "webp"
	if format == 0 {
		formatStr = "jpeg"
	}

	record := &database.ScreenshotRecord{
		ServerAddr: serverAddr,
		PlayerIP:   playerIP,
		PlayerName: playerName,
		ClientID:   int(clientID),
		Width:      width,
		Height:     height,
		Format:     formatStr,
		FilePath:   filePath,
		FileSize:   int64(len(data)),
		Timestamp:  now,
	}

	if _, err := s.db.InsertScreenshot(record); err != nil {
		log.Printf("[SCREENSHOT] Warning: failed to store metadata: %v", err)
	}

	return filePath, nil
}

// BaseDir returns the screenshots base directory
func (s *Storage) BaseDir() string {
	return s.baseDir
}

// GetDiskStats computes total number of screenshot files and their total size in bytes
func (s *Storage) GetDiskStats() (int, int64, error) {
	var count int
	var totalBytes int64

	err := filepath.Walk(s.baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if !info.IsDir() {
			count++
			totalBytes += info.Size()
		}
		return nil
	})

	return count, totalBytes, err
}

// PurgeOlderThan deletes screenshots and folders older than specified days.
// If days <= 0, no action is taken.
func (s *Storage) PurgeOlderThan(days int) (int, int64, error) {
	if days <= 0 {
		return 0, 0, nil
	}

	cutoff := time.Now().AddDate(0, 0, -days)
	oldRecords, err := s.db.GetScreenshotsOlderThan(cutoff)
	if err != nil {
		return 0, 0, fmt.Errorf("get old screenshots: %w", err)
	}

	var deletedCount int
	var freedBytes int64

	for _, rec := range oldRecords {
		if rec.FilePath != "" {
			if fi, err := os.Stat(rec.FilePath); err == nil {
				freedBytes += fi.Size()
				if err := os.Remove(rec.FilePath); err != nil && !os.IsNotExist(err) {
					log.Printf("[SCREENSHOT] Warning: failed to delete file %s: %v", rec.FilePath, err)
				}
			}
		}
	}

	// Delete from database
	dbDeleted, err := s.db.DeleteScreenshotsOlderThan(cutoff)
	if err != nil {
		log.Printf("[SCREENSHOT] Warning: failed to delete DB records: %v", err)
	}
	deletedCount = int(dbDeleted)
	if deletedCount == 0 && len(oldRecords) > 0 {
		deletedCount = len(oldRecords)
	}

	// Clean up empty/old date directories in baseDir
	entries, err := os.ReadDir(s.baseDir)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				dirPath := filepath.Join(s.baseDir, entry.Name())
				if folderDate, parseErr := time.Parse("2006-01-02", entry.Name()); parseErr == nil {
					if folderDate.Before(cutoff.Truncate(24 * time.Hour)) {
						os.RemoveAll(dirPath)
					}
				} else {
					subEntries, _ := os.ReadDir(dirPath)
					if len(subEntries) == 0 {
						os.Remove(dirPath)
					}
				}
			}
		}
	}

	if deletedCount > 0 {
		_ = s.db.Vacuum()
	}

	log.Printf("[SCREENSHOT] Purged %d old screenshots (%d bytes freed, older than %d days)", deletedCount, freedBytes, days)
	return deletedCount, freedBytes, nil
}

