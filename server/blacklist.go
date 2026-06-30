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
	dbEntries []database.BlacklistEntry
}

// NewBlacklist creates a new blacklist manager
func NewBlacklist() *Blacklist {
	bl := &Blacklist{
		processes: make(map[string]bool),
		modules:   make(map[string]bool),
	}
	bl.loadHardcoded()
	return bl
}

// SetDB sets the database reference and loads DB entries
func (bl *Blacklist) SetDB(db *database.DB) {
	bl.mu.Lock()
	defer bl.mu.Unlock()
	bl.db = db
	bl.loadFromDB()
}

// loadHardcoded populates the built-in cheat patterns
func (bl *Blacklist) loadHardcoded() {
	cheatProcesses := []string{
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

	cheatModules := []string{
		"cheat", "hack", "inject", "hook", "overlay",
		"trainer", "speedhack", "aimbot", "wallhack",
		"sbyte", "megajump", "sentry", "hookdll",
		"d3dhook", "d3d9hook", "dxgihook",
		"cheatengine", "artmoney", "gameguard",
		"speedhack", "noclip", "aimassist",
		"steamhook", "discordhook",
		"reshade", "enb", "sweetfx",
	}

	for _, p := range cheatProcesses {
		bl.processes[strings.ToLower(p)] = true
	}
	for _, p := range cheatModules {
		bl.modules[strings.ToLower(p)] = true
	}

	log.Printf("[BLACKLIST] Loaded hardcoded: %d process patterns, %d module patterns",
		len(bl.processes), len(bl.modules))
}

// loadFromDB loads custom entries from the database
func (bl *Blacklist) loadFromDB() {
	if bl.db == nil {
		return
	}

	entries, err := bl.db.GetBlacklist()
	if err != nil {
		log.Printf("[BLACKLIST] Error loading from DB: %v", err)
		return
	}

	bl.dbEntries = entries
	for _, e := range entries {
		pattern := strings.ToLower(e.Pattern)
		switch e.Type {
		case "process":
			bl.processes[pattern] = true
		case "module":
			bl.modules[pattern] = true
		}
	}

	log.Printf("[BLACKLIST] Loaded %d custom entries from DB", len(entries))
}

// Reload forces a reload of DB entries
func (bl *Blacklist) Reload() {
	bl.mu.Lock()
	defer bl.mu.Unlock()

	bl.dbEntries = nil
	bl.processes = make(map[string]bool)
	bl.modules = make(map[string]bool)
	bl.loadHardcoded()
	bl.loadFromDB()
}

// AddEntry adds a new blacklist entry
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

// RemoveEntry removes a blacklist entry by ID
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

// GetDBEntries returns the custom database entries
func (bl *Blacklist) GetDBEntries() []database.BlacklistEntry {
	bl.mu.RLock()
	defer bl.mu.RUnlock()
	return bl.dbEntries
}

// CheckProcess checks if a process name matches any blacklist pattern
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

// CheckModule checks if a module name matches any blacklist pattern
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

// CheckModuleWithPath checks both module name and path
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
func (bl *Blacklist) Stats() (processCount, moduleCount, dbCount int) {
	bl.mu.RLock()
	defer bl.mu.RUnlock()
	return len(bl.processes), len(bl.modules), len(bl.dbEntries)
}
