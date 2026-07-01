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
