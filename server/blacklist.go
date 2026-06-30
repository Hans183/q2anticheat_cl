package server

import (
	"log"
	"strings"
	"sync"

	"github.com/user/anticheat_cl/database"
)

// Blacklist manages hardcoded + database blacklist entries
type Blacklist struct {
	mu        sync.RWMutex
	db        *database.DB
	processes map[string]bool
	modules   map[string]bool
	allEntries []database.BlacklistEntry
}

// NewBlacklist creates a new blacklist manager
func NewBlacklist() *Blacklist {
	return &Blacklist{
		processes: make(map[string]bool),
		modules:   make(map[string]bool),
	}
}

// SetDB sets the database reference, ensures hardcoded entries exist, and loads all entries
func (bl *Blacklist) SetDB(db *database.DB) {
	bl.mu.Lock()
	defer bl.mu.Unlock()
	bl.db = db
	bl.ensureHardcoded()
	bl.loadFromDB()
}

// hardcodedProcessPatterns are the built-in cheat process patterns
var hardcodedProcessPatterns = []string{
	"aimbot", "wallhack", "cheat", "inject", "hook",
	"trainer", "hack", "speedhack", "noclip", "aimassist",
	"cheatengine", "cheat engine", "ce.exe",
	"artmoney", "gamehack", "gamemonitor",
	"memoryhack", "memhack",
	"injector", "dllinject", "processinject",
	"extremepro injector", "xenos", "gh injector",
	"sin injector", "keystone", "builtinject",
	"ollydbg", "x64dbg", "x32dbg", "ida.exe", "idag.exe",
	"windbg", "immunity", "immdebug", "dnspy", "de4dot",
	"overwolf", "gameseal",
	"aimware", "ezfrags", "onetap",
}

// hardcodedModulePatterns are the built-in cheat module patterns
var hardcodedModulePatterns = []string{
	"cheat", "hack", "inject", "hook", "overlay",
	"trainer", "speedhack", "aimbot", "wallhack",
	"sbyte", "megajump", "sentry", "hookdll",
	"d3dhook", "d3d9hook", "dxgihook",
	"cheatengine", "artmoney", "gameguard",
	"speedhack", "noclip", "aimassist",
	"steamhook", "discordhook",
	"reshade", "enb", "sweetfx",
}

// ensureHardcoded inserts hardcoded patterns into DB if they don't exist
func (bl *Blacklist) ensureHardcoded() {
	if bl.db == nil {
		return
	}

	var entries []database.BlacklistEntry
	for _, p := range hardcodedProcessPatterns {
		entries = append(entries, database.BlacklistEntry{
			Type:    "process",
			Pattern: p,
			Enabled: true,
		})
	}
	for _, p := range hardcodedModulePatterns {
		entries = append(entries, database.BlacklistEntry{
			Type:    "module",
			Pattern: p,
			Enabled: true,
		})
	}

	if err := bl.db.EnsureHardcodedEntries(entries); err != nil {
		log.Printf("[BLACKLIST] Error ensuring hardcoded entries: %v", err)
	}

	log.Printf("[BLACKLIST] Ensured %d hardcoded patterns in DB", len(entries))
}

// loadFromDB loads all entries from the database and populates active maps
func (bl *Blacklist) loadFromDB() {
	if bl.db == nil {
		return
	}

	entries, err := bl.db.GetBlacklist()
	if err != nil {
		log.Printf("[BLACKLIST] Error loading from DB: %v", err)
		return
	}

	bl.allEntries = entries
	activeCount := 0
	for _, e := range entries {
		if !e.Enabled {
			continue
		}
		pattern := strings.ToLower(e.Pattern)
		switch e.Type {
		case "process":
			bl.processes[pattern] = true
		case "module":
			bl.modules[pattern] = true
		}
		activeCount++
	}

	log.Printf("[BLACKLIST] Loaded %d entries from DB (%d active, %d disabled)",
		len(entries), activeCount, len(entries)-activeCount)
}

// Reload forces a full reload from DB
func (bl *Blacklist) Reload() {
	bl.mu.Lock()
	defer bl.mu.Unlock()

	bl.allEntries = nil
	bl.processes = make(map[string]bool)
	bl.modules = make(map[string]bool)
	bl.loadFromDB()
}

// AddEntry adds a new user-defined blacklist entry
func (bl *Blacklist) AddEntry(entryType, pattern, addedBy string) error {
	if bl.db == nil {
		return nil
	}

	if err := bl.db.AddBlacklistEntry(entryType, pattern, addedBy); err != nil {
		return err
	}

	bl.Reload()
	return nil
}

// RemoveEntry removes a user-defined blacklist entry by ID
func (bl *Blacklist) RemoveEntry(id int64) error {
	if bl.db == nil {
		return nil
	}

	if err := bl.db.RemoveBlacklistEntry(id); err != nil {
		return err
	}

	bl.Reload()
	return nil
}

// ToggleEntry toggles the enabled state of a blacklist entry
func (bl *Blacklist) ToggleEntry(id int64) error {
	if bl.db == nil {
		return nil
	}

	if err := bl.db.ToggleBlacklistEntry(id); err != nil {
		return err
	}

	bl.Reload()
	return nil
}

// GetAllEntries returns all entries (hardcoded + user) with their state
func (bl *Blacklist) GetAllEntries() []database.BlacklistEntry {
	bl.mu.RLock()
	defer bl.mu.RUnlock()
	return bl.allEntries
}

// CheckProcess checks if a process name matches any active blacklist pattern
func (bl *Blacklist) CheckProcess(name string) (bool, string) {
	bl.mu.RLock()
	defer bl.mu.RUnlock()

	lower := strings.ToLower(name)
	for pattern := range bl.processes {
		if strings.Contains(lower, pattern) {
			return true, pattern
		}
	}
	return false, ""
}

// CheckModule checks if a module name matches any active blacklist pattern
func (bl *Blacklist) CheckModule(name string) (bool, string) {
	bl.mu.RLock()
	defer bl.mu.RUnlock()

	lower := strings.ToLower(name)
	for pattern := range bl.modules {
		if strings.Contains(lower, pattern) {
			return true, pattern
		}
	}
	return false, ""
}

// CheckModuleWithPath checks both module name and path against active patterns
func (bl *Blacklist) CheckModuleWithPath(name, path string) (bool, string, string) {
	bl.mu.RLock()
	defer bl.mu.RUnlock()

	lowerName := strings.ToLower(name)
	lowerPath := strings.ToLower(path)

	for pattern := range bl.modules {
		if strings.Contains(lowerName, pattern) {
			return true, pattern, "name"
		}
		if strings.Contains(lowerPath, pattern) {
			return true, pattern, "path"
		}
	}
	return false, "", ""
}

// Stats returns blacklist statistics
func (bl *Blacklist) Stats() (activeProcesses, activeModules, totalEntries int) {
	bl.mu.RLock()
	defer bl.mu.RUnlock()
	return len(bl.processes), len(bl.modules), len(bl.allEntries)
}
