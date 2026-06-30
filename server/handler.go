package server

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/user/anticheat_cl/protocol"
	"github.com/user/anticheat_cl/screenshots"
)

// Handler processes protocol messages from game servers
type Handler struct {
	servers   map[string]*GameServer // keyed by remote address
	serversMu sync.RWMutex
	storage   *screenshots.Storage
	// Callback for violations
	OnViolation func(serverAddr, playerIP, playerName string, clientID uint32, vType, reason string)
}

// NewHandler creates a new message handler
func NewHandler(storage *screenshots.Storage) *Handler {
	return &Handler{
		servers: make(map[string]*GameServer),
		storage: storage,
	}
}

// HandleMessage processes a single message from a game server
func (h *Handler) HandleMessage(gs *GameServer, buf []byte) {
	msg, err := protocol.ParseMessage(buf)
	if err != nil {
		log.Printf("[HANDLER] Error parsing message from %s: %v", gs.RemoteAddr, err)
		return
	}

	log.Printf("[HANDLER] Received message type %d from %s", msg.Type, gs.RemoteAddr)

	switch msg.Type {
	case protocol.ACC_VERSION:
		h.handleVersion(gs, msg.Version)

	case protocol.ACC_REQUESTCHALLENGE:
		h.handleChallenge(gs, msg.Challenge)

	case protocol.ACC_CLIENTDISCONNECT:
		h.handleDisconnect(gs, msg.Disconnect)

	case protocol.ACC_QUERYCLIENT:
		h.handleQueryClient(gs, msg.QueryClient)

	case protocol.ACC_PING:
		h.handlePing(gs)

	case protocol.ACC_UPDATECHECKS:
		h.handleUpdateChecks(gs, msg.Checks)

	case protocol.ACC_SETPREFERENCES:
		h.handleSetPreferences(gs, msg.Prefs)

	case protocol.ACC_SCREENSHOT_DATA:
		h.handleScreenshot(gs, msg.Screenshot)

	case protocol.ACC_CLIENTDATA:
		h.handleClientData(gs, msg.ClientData)

	case protocol.ACC_PROCESSDATA:
		h.handleProcessData(gs, msg.ProcessData)

	default:
		log.Printf("[HANDLER] Unknown message type %d from %s", msg.Type, gs.RemoteAddr)
	}
}

func (h *Handler) handleVersion(gs *GameServer, ver *protocol.VersionMessage) {
	if ver == nil {
		return
	}

	gs.mu.Lock()
	gs.Hostname = ver.Hostname
	gs.Version = ver.Version
	gs.Port = ver.Port
	gs.mu.Unlock()

	log.Printf("[HANDLER] Server connected: %s (v%s, port %d) from %s",
		ver.Hostname, ver.Version, ver.Port, gs.RemoteAddr)

	// Register server
	h.serversMu.Lock()
	h.servers[gs.RemoteAddr.String()] = gs
	h.serversMu.Unlock()

	// Send READY
	if err := gs.SendReady(); err != nil {
		log.Printf("[HANDLER] Error sending READY to %s: %v", gs.RemoteAddr, err)
	}
}

func (h *Handler) handleChallenge(gs *GameServer, ch *protocol.ChallengeMessage) {
	if ch == nil {
		return
	}

	log.Printf("[HANDLER] Challenge request from %s: client=%d, ip=%s:%d",
		gs.RemoteAddr, ch.ClientID, ch.IP, ch.Port)

	// Create client info
	client := &ClientInfo{
		ClientID:   ch.ClientID,
		Challenge:  ch.Challenge,
		IP:         ch.IP,
		Valid:      false,
		QuerySent:  false,
		Required:   int(protocol.AC_NORMAL),
		ClientType: protocol.AC_TYPE_Q2PRO,
	}

	gs.SetClient(client)

	// TODO: Validate client against checks
	// For now, acknowledge the client
	if err := gs.SendClientAck(ch.ClientID, ch.Challenge, protocol.AC_TYPE_Q2PRO); err != nil {
		log.Printf("[HANDLER] Error sending CLIENTACK to %s: %v", gs.RemoteAddr, err)
	}
}

func (h *Handler) handleDisconnect(gs *GameServer, dc *protocol.DisconnectMessage) {
	if dc == nil {
		return
	}

	log.Printf("[HANDLER] Client disconnect from %s: client=%d",
		gs.RemoteAddr, dc.ClientID)

	gs.RemoveClient(dc.ClientID)
}

func (h *Handler) handleQueryClient(gs *GameServer, qc *protocol.QueryClientMessage) {
	if qc == nil {
		return
	}

	client := gs.GetClient(qc.ClientID)
	if client == nil {
		log.Printf("[HANDLER] Query for unknown client %d from %s", qc.ClientID, gs.RemoteAddr)
		return
	}

	// Reply with current status
	if err := gs.SendQueryReply(qc.ClientID, qc.Challenge,
		client.Valid, client.ClientType); err != nil {
		log.Printf("[HANDLER] Error sending QUERYREPLY to %s: %v", gs.RemoteAddr, err)
	}
}

func (h *Handler) handlePing(gs *GameServer) {
	if err := gs.SendPong(); err != nil {
		log.Printf("[HANDLER] Error sending PONG to %s: %v", gs.RemoteAddr, err)
	}
}

func (h *Handler) handleUpdateChecks(gs *GameServer, checks *protocol.ChecksMessage) {
	if checks == nil {
		return
	}

	gs.mu.Lock()
	gs.FileChecks = checks.Files
	gs.CvarChecks = checks.Cvars
	gs.mu.Unlock()

	log.Printf("[HANDLER] Received %d file checks, %d cvar checks from %s",
		len(checks.Files), len(checks.Cvars), gs.RemoteAddr)
}

func (h *Handler) handleSetPreferences(gs *GameServer, prefs *protocol.PrefsMessage) {
	if prefs == nil {
		return
	}

	gs.mu.Lock()
	gs.Prefs = prefs.Flags
	gs.mu.Unlock()

	log.Printf("[HANDLER] Preferences set: %d from %s", prefs.Flags, gs.RemoteAddr)
}

func (h *Handler) handleScreenshot(gs *GameServer, ss *protocol.ScreenshotData) {
	if ss == nil {
		return
	}

	log.Printf("[HANDLER] Screenshot received from %s: client=%d, %dx%d, %d bytes",
		gs.RemoteAddr, ss.ClientID, ss.Width, ss.Height, len(ss.JPEGData))

	// Look up client info for IP address
	client := gs.GetClient(ss.ClientID)
	playerIP := ""
	playerName := ""
	if client != nil {
		if client.IP != nil {
			playerIP = client.IP.String()
		}
		playerName = client.Name
	}

	// Save to disk and database
	if h.storage != nil {
		filePath, err := h.storage.SaveScreenshot(
			gs.Hostname, playerIP, playerName,
			ss.ClientID, int(ss.Width), int(ss.Height), ss.JPEGData)
		if err != nil {
			log.Printf("[HANDLER] Error saving screenshot: %v", err)
		} else {
			log.Printf("[HANDLER] Screenshot saved: %s", filePath)
		}
	}
}

func (h *Handler) handleClientData(gs *GameServer, cd *protocol.ClientDataMessage) {
	if cd == nil {
		return
	}

	log.Printf("[HANDLER] Client data from %s: client=%d, %d files, %d cvars",
		gs.RemoteAddr, cd.ClientID, len(cd.Files), len(cd.Cvars))

	client := gs.GetClient(cd.ClientID)
	if client == nil {
		// Client not yet registered, create from client data
		client = &ClientInfo{
			ClientID:   cd.ClientID,
			Challenge:  cd.Challenge,
			Name:       cd.PlayerName,
			Valid:      false,
			ClientType: protocol.AC_TYPE_Q2PRO,
		}
		gs.SetClient(client)
	}

	// Update name from client data
	if cd.PlayerName != "" {
		client.Name = cd.PlayerName
	}

	var fileViolations []string
	var cvarViolations []string

	// Validate file hashes against stored checks
	for _, fileData := range cd.Files {
		for _, expected := range gs.FileChecks {
			if expected.Path != fileData.Path {
				continue
			}

			if expected.Flags&protocol.ACH_NEGATIVE != 0 {
				if bytes.Equal(expected.Hash[:], fileData.Hash[:]) {
					fileViolations = append(fileViolations, fmt.Sprintf("file %s: negative hash match", fileData.Path))
					log.Printf("[HANDLER] File violation (negative match): %s from client %d",
						fileData.Path, cd.ClientID)
				}
			} else {
				if !bytes.Equal(expected.Hash[:], fileData.Hash[:]) {
					fileViolations = append(fileViolations, fmt.Sprintf("file %s: hash mismatch", fileData.Path))
					log.Printf("[HANDLER] File violation (hash mismatch): %s from client %d: got %x, expected %x",
						fileData.Path, cd.ClientID, fileData.Hash, expected.Hash)
				}
			}
			break
		}
	}

	// Check for cvars not reported by client (missing cvars = violation)
	for _, expected := range gs.CvarChecks {
		found := false
		for _, cvarData := range cd.Cvars {
			if expected.Name == cvarData.Name {
				found = true
				break
			}
		}
		if !found {
			cvarViolations = append(cvarViolations, fmt.Sprintf("cvar %s: not reported", expected.Name))
			log.Printf("[HANDLER] Cvar missing: %s not reported by client %d from %s",
				expected.Name, cd.ClientID, gs.RemoteAddr)
		}
	}

	// Validate cvar values against stored checks
	for _, cvarData := range cd.Cvars {
		for _, expected := range gs.CvarChecks {
			if expected.Name != cvarData.Name {
				continue
			}

			if !compareCvar(cvarData.Value, expected) {
				cvarViolations = append(cvarViolations, fmt.Sprintf("cvar %s=%s", cvarData.Name, cvarData.Value))
				log.Printf("[HANDLER] Cvar violation: %s=%s from client %d (expected default %s, op %d)",
					cvarData.Name, cvarData.Value, cd.ClientID, expected.Default, expected.Op)
			}
			break
		}
	}

	// Merge all violations for logging
	allViolations := append(fileViolations, cvarViolations...)

	if len(allViolations) > 0 {
		// Log violations to database
		if h.OnViolation != nil {
			playerIP := ""
			playerName := client.Name
			if client.IP != nil {
				playerIP = client.IP.String()
			}
			for _, v := range allViolations {
				vType := "cvar"
				if strings.HasPrefix(v, "file ") {
					vType = "file"
				}
				h.OnViolation(gs.RemoteAddr.String(), playerIP, playerName, cd.ClientID, vType, v)
			}
		}

		// File violations = kick (can't auto-fix modified files)
		if len(fileViolations) > 0 {
			reason := fmt.Sprintf("anticheat violation: %s", fileViolations[0])
			clientMsg := fmt.Sprintf("Anticheat violation detected: %s", strings.Join(fileViolations, "; "))
			log.Printf("[HANDLER] Sending FILE violation (kick) for client %d from %s: %s",
				cd.ClientID, gs.RemoteAddr, reason)
			gs.SendViolation(cd.ClientID, cd.Challenge, reason, clientMsg)
			return
		}

		// Cvar-only violations = warning + reconnect (server already stuffed correct values via AC_EnforceClientCvars)
		if len(cvarViolations) > 0 {
			warning := fmt.Sprintf("cvar fixes applied: %s", strings.Join(cvarViolations, "; "))
			log.Printf("[HANDLER] Sending CVARWARNING for client %d from %s: %s",
				cd.ClientID, gs.RemoteAddr, warning)
			gs.SendCvarWarning(cd.ClientID, cd.Challenge, warning)
			return
		}
	}

	// All checks passed — client is valid
	client.Valid = true
	log.Printf("[HANDLER] Client %d passed all AC checks from %s", cd.ClientID, gs.RemoteAddr)

	if err := gs.SendClientAck(cd.ClientID, cd.Challenge, protocol.AC_TYPE_Q2PRO); err != nil {
		log.Printf("[HANDLER] Error sending CLIENTACK to %s: %v", gs.RemoteAddr, err)
	}
}

func (h *Handler) handleProcessData(gs *GameServer, pd *protocol.ProcessDataMessage) {
	if pd == nil {
		return
	}

	log.Printf("[HANDLER] Process data from %s: client=%d, %d processes, %d modules",
		gs.RemoteAddr, pd.ClientID, len(pd.Processes), len(pd.Modules))

	// Look up client info
	client := gs.GetClient(pd.ClientID)
	playerIP := ""
	playerName := pd.PlayerName
	if client != nil {
		if client.IP != nil {
			playerIP = client.IP.String()
		}
		if client.Name != "" {
			playerName = client.Name
		}
	}

	// Store process snapshot in database
	if h.storage != nil && h.storage.DB() != nil {
		_, err := h.storage.DB().InsertProcessSnapshot(
			gs.Hostname, playerIP, playerName,
			int(pd.ClientID), len(pd.Processes), len(pd.Modules),
			formatProcessViolations(pd), nil)
		if err != nil {
			log.Printf("[HANDLER] Error storing process snapshot: %v", err)
		}
	}
}

// formatProcessViolations creates a summary string of any suspicious processes/modules
func formatProcessViolations(pd *protocol.ProcessDataMessage) string {
	var violations []string

	// Check for known cheat process names
	suspiciousProcs := []string{
		"aimbot", "wallhack", "cheat", "inject", "hook",
		"trainer", "hack", "speedhack", "noclip",
	}

	for _, proc := range pd.Processes {
		name := strings.ToLower(proc.Name)
		for _, suspicious := range suspiciousProcs {
			if strings.Contains(name, suspicious) {
				violations = append(violations, fmt.Sprintf("suspicious process: %s (pid=%d)", proc.Name, proc.PID))
				break
			}
		}
	}

	// Check for known cheat module names
	suspiciousMods := []string{
		"cheat", "hack", "inject", "hook", "overlay",
		"trainer", "speedhack", "aimbot", "wallhack",
	}

	for _, mod := range pd.Modules {
		name := strings.ToLower(mod.Name)
		for _, suspicious := range suspiciousMods {
			if strings.Contains(name, suspicious) {
				violations = append(violations, fmt.Sprintf("suspicious module: %s", mod.Name))
				break
			}
		}
	}

	if len(violations) == 0 {
		return ""
	}
	return strings.Join(violations, "; ")
}

// compareCvar checks a client's cvar value against expected check rules.
// Returns TRUE if value PASSES (no violation), FALSE if value FAILS (violation).
func compareCvar(value string, check protocol.CvarCheck) bool {
	switch check.Op {
	case protocol.OP_EQUAL:
		for _, v := range check.Values {
			if value == v {
				return true
			}
		}
		return false
	case protocol.OP_NEQUAL:
		for _, v := range check.Values {
			if value == v {
				return true
			}
		}
		return false
	case protocol.OP_STREQUAL:
		for _, v := range check.Values {
			if strings.EqualFold(value, v) {
				return true
			}
		}
		return false
	case protocol.OP_STRNEQUAL:
		for _, v := range check.Values {
			if strings.EqualFold(value, v) {
				return true
			}
		}
		return false
	case protocol.OP_STRSTR:
		for _, v := range check.Values {
			if len(value) >= len(v) && strings.Contains(value, v) {
				return true
			}
		}
		return false
	default:
		if len(check.Values) == 0 {
			return true
		}
		fv, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return false
		}
		cv, err := strconv.ParseFloat(check.Values[0], 64)
		if err != nil {
			return true
		}
		switch check.Op {
		case protocol.OP_GT:
			return !(fv > cv)
		case protocol.OP_LT:
			return !(fv < cv)
		case protocol.OP_GTEQUAL:
			return !(fv >= cv)
		case protocol.OP_LTEQUAL:
			return !(fv <= cv)
		default:
			return true
		}
	}
}

// GetServer returns a game server by address
func (h *Handler) GetServer(addr string) *GameServer {
	h.serversMu.RLock()
	defer h.serversMu.RUnlock()
	return h.servers[addr]
}

// GetServers returns all connected game servers
func (h *Handler) GetServers() []*GameServer {
	h.serversMu.RLock()
	defer h.serversMu.RUnlock()

	servers := make([]*GameServer, 0, len(h.servers))
	for _, gs := range h.servers {
		if gs.GetState() != StateDisconnected {
			servers = append(servers, gs)
		}
	}
	return servers
}

// RemoveServer removes a disconnected server from the map
func (h *Handler) RemoveServer(addr string) {
	h.serversMu.Lock()
	defer h.serversMu.Unlock()
	delete(h.servers, addr)
	log.Printf("[HANDLER] Removed disconnected server: %s", addr)
}

// CheckTimeouts checks for timed out connections
func (h *Handler) CheckTimeouts() {
	h.serversMu.Lock()
	defer h.serversMu.Unlock()

	now := time.Now()
	for addr, gs := range h.servers {
		gs.mu.Lock()
		if gs.PingPending && now.Sub(gs.LastPingAt) > 15*time.Second {
			log.Printf("[HANDLER] Ping timeout from %s", addr)
			gs.Close()
			delete(h.servers, addr)
		}
		gs.mu.Unlock()
	}
}
