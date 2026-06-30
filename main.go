package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/user/anticheat_cl/database"
	"github.com/user/anticheat_cl/screenshots"
	"github.com/user/anticheat_cl/server"
	"github.com/user/anticheat_cl/web"
)

type Config struct {
	ListenAddr    string
	DBPath        string
	ScreenshotDir string
	LogFile       string
	WebAddr       string
	AdminUser     string
	AdminPass     string
}

// parseAdminUsers parses "user1:pass1,user2:pass2" into AdminInput list
func parseAdminUsers(raw string) []database.AdminInput {
	var admins []database.AdminInput
	for _, pair := range strings.Split(raw, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) != 2 {
			log.Printf("Warning: invalid admin format %q (expected user:pass), skipping", pair)
			continue
		}
		username := strings.TrimSpace(parts[0])
		password := strings.TrimSpace(parts[1])
		if username == "" || password == "" {
			log.Printf("Warning: empty username or password in %q, skipping", pair)
			continue
		}
		admins = append(admins, database.AdminInput{
			Username:     username,
			PasswordHash: web.HashPassword(password),
		})
	}
	return admins
}

func main() {
	config := &Config{}

	flag.StringVar(&config.ListenAddr, "listen", "0.0.0.0:27915", "Listen address")
	flag.StringVar(&config.DBPath, "db", "./data/anticheat.db", "SQLite database path")
	flag.StringVar(&config.ScreenshotDir, "screenshots", "./data/screenshots", "Screenshot directory")
	flag.StringVar(&config.LogFile, "log", "", "Log file (empty for stdout)")
	flag.StringVar(&config.WebAddr, "web", "0.0.0.0:27916", "Dashboard web listen address")
	flag.StringVar(&config.AdminUser, "admin-user", "admin", "Default admin username")
	flag.StringVar(&config.AdminPass, "admin-pass", "admin123", "Default admin password")
	flag.Parse()

	// Environment variables override CLI flags (useful for Coolify/Docker)
	if envUser := os.Getenv("ADMIN_USER"); envUser != "" {
		config.AdminUser = envUser
	}
	if envPass := os.Getenv("ADMIN_PASS"); envPass != "" {
		config.AdminPass = envPass
	}

	if config.LogFile != "" {
		f, err := os.OpenFile(config.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			log.Fatalf("Failed to open log file: %v", err)
		}
		defer f.Close()
		log.SetOutput(f)
	}

	fmt.Println("========================================")
	fmt.Println("  Q2PRO Anticheat Server")
	fmt.Println("  Version: 1.1.0")
	fmt.Println("========================================")
	fmt.Printf("Listen:      %s\n", config.ListenAddr)
	fmt.Printf("Dashboard:   %s\n", config.WebAddr)
	fmt.Printf("Database:    %s\n", config.DBPath)
	fmt.Printf("Screenshots: %s\n", config.ScreenshotDir)
	fmt.Println()

	db, err := database.New(config.DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize admin users
	adminsEnv := os.Getenv("ADMIN_USERS")
	if adminsEnv != "" {
		// Multiple admins: format "user1:pass1,user2:pass2"
		admins := parseAdminUsers(adminsEnv)
		if len(admins) > 0 {
			if err := web.InitAdmins(db, admins); err != nil {
				log.Printf("Warning: could not create admin users: %v", err)
			}
		}
	} else {
		// Single admin fallback via ADMIN_USER/ADMIN_PASS or CLI flags
		if err := web.InitDefaultAdmin(db, config.AdminUser, config.AdminPass); err != nil {
			log.Printf("Warning: could not create default admin: %v", err)
		}
	}

	storage, err := screenshots.NewStorage(config.ScreenshotDir, db)
	if err != nil {
		log.Fatalf("Failed to initialize screenshot storage: %v", err)
	}

	tcpServer := server.New(config.ListenAddr, storage)

	// Initialize blacklist with database
	tcpServer.GetHandler().Blacklist().SetDB(db)

	// Wire up violation callback to log to database
	tcpServer.GetHandler().OnViolation = func(serverAddr, playerIP, playerName string, clientID uint32, vType, reason string) {
		db.InsertViolation(&database.ViolationRecord{
			ServerAddr: serverAddr,
			PlayerIP:   playerIP,
			PlayerName: playerName,
			ClientID:   int(clientID),
			Type:       vType,
			Reason:     reason,
			Timestamp:  time.Now(),
		})
	}

	if err := tcpServer.Start(); err != nil {
		log.Fatalf("Failed to start TCP server: %v", err)
	}

	// Start web dashboard
	go func() {
		ws := web.New(config.WebAddr, db, tcpServer.GetHandler())
		if err := ws.Start(); err != nil {
			log.Printf("[WEB] Dashboard error: %v", err)
		}
	}()

	// Cleanup sessions periodically
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			db.CleanupSessions()
		}
	}()

	fmt.Println("Server is ready to accept connections")
	fmt.Println("Press Ctrl+C to stop")
	fmt.Println()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	fmt.Println("\nShutting down...")
	tcpServer.Stop()

	time.Sleep(100 * time.Millisecond)
	fmt.Println("Server stopped")
}
