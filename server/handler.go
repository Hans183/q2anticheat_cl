package server

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
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
	blacklist *Blacklist
	// Callback for violations
	OnViolation func(serverAddr, playerIP, playerName string, clientID uint32, vType, reason string)
}

// NewHandler creates a new message handler
func NewHandler(storage *screenshots.Storage) *Handler {
	return &Handler{
		servers:   make(map[string]*GameServer),
		storage:   storage,
		blacklist: NewBlacklist(),
	}
}

// Blacklist returns the blacklist manager
func (h *Handler) Blacklist() *Blacklist {
	return h.blacklist
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

	case protocol.ACC_NAMEUPDATE:
		h.handleNameUpdate(gs, msg.NameUpdate)

	case protocol.ACC_HOSTNAMEUPDATE:
		h.handleHostnameUpdate(gs, msg.HostnameUpdate)

	case protocol.ACC_CVARCHANGE:
		h.handleCvarChange(gs, msg.CvarChange)

	case protocol.ACC_SPIKEDMODEL:
		h.handleSpikedModel(gs, msg.SpikedModel)

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

	log.Printf("[HANDLER] Screenshot received from %s: client=%d, %dx%d, format=%d, %d bytes",
		gs.RemoteAddr, ss.ClientID, ss.Width, ss.Height, ss.Format, len(ss.Data))

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
			ss.ClientID, int(ss.Width), int(ss.Height), ss.Format, ss.Data)
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

	// Check processes and modules against blacklist
	var violations []string
	for _, proc := range pd.Processes {
		if matched, pattern := h.blacklist.CheckProcess(proc.Name); matched {
			violations = append(violations, fmt.Sprintf("proceso sospechoso: %s (pid=%d, patron: %s)", proc.Name, proc.PID, pattern))
		}
	}
	for _, mod := range pd.Modules {
		modSHA1Hex := hex.EncodeToString(mod.SHA1[:])
		if matched, pattern, matchIn := h.blacklist.CheckModuleFull(mod.Name, mod.Path, modSHA1Hex); matched {
			violations = append(violations, fmt.Sprintf("modulo sospechoso: %s (%s: %s)", mod.Name, matchIn, pattern))
		}
	}

	violationStr := strings.Join(violations, "; ")

	// Serialize processes and modules to JSON
	processesJSON := "[]"
	modulesJSON := "[]"

	if len(pd.Processes) > 0 {
		type jsonProcess struct {
			PID      uint32 `json:"pid"`
			ParentPID uint32 `json:"parent_pid"`
			Name     string `json:"name"`
		}
		procs := make([]jsonProcess, len(pd.Processes))
		for i, p := range pd.Processes {
			procs[i] = jsonProcess{PID: p.PID, ParentPID: p.ParentPID, Name: p.Name}
		}
		if data, err := json.Marshal(procs); err == nil {
			processesJSON = string(data)
		}
	}

	if len(pd.Modules) > 0 {
		type jsonModule struct {
			Name string `json:"name"`
			Path string `json:"path"`
			SHA1 string `json:"sha1"`
		}
		mods := make([]jsonModule, len(pd.Modules))
		for i, m := range pd.Modules {
			mods[i] = jsonModule{Name: m.Name, Path: m.Path, SHA1: fmt.Sprintf("%x", m.SHA1)}
		}
		if data, err := json.Marshal(mods); err == nil {
			modulesJSON = string(data)
		}
	}

	// Store process snapshot in database
	if h.storage != nil && h.storage.DB() != nil {
		_, err := h.storage.DB().InsertProcessSnapshot(
			gs.Hostname, playerIP, playerName,
			int(pd.ClientID), len(pd.Processes), len(pd.Modules),
			violationStr, processesJSON, modulesJSON)
		if err != nil {
			log.Printf("[HANDLER] Error storing process snapshot: %v", err)
		}
	}

	// Log violations
	if len(violations) > 0 {
		log.Printf("[HANDLER] Process violations from %s (client %d): %s",
			gs.RemoteAddr, pd.ClientID, violationStr)
		if h.OnViolation != nil {
			for _, v := range violations {
				h.OnViolation(gs.RemoteAddr.String(), playerIP, playerName, pd.ClientID, "process", v)
			}
		}
	}
}

func (h *Handler) handleNameUpdate(gs *GameServer, nu *protocol.NameUpdateMessage) {
	if nu == nil {
		return
	}

	client := gs.GetClient(nu.ClientID)
	if client == nil {
		log.Printf("[HANDLER] Name update for unknown client %d from %s", nu.ClientID, gs.RemoteAddr)
		return
	}

	oldName := client.Name
	client.Name = nu.NewName

	log.Printf("[HANDLER] Client %d name updated: %s -> %s (from %s)",
		nu.ClientID, oldName, nu.NewName, gs.RemoteAddr)
}

func (h *Handler) handleHostnameUpdate(gs *GameServer, hu *protocol.HostnameUpdateMessage) {
	if hu == nil {
		return
	}

	gs.mu.Lock()
	oldHostname := gs.Hostname
	gs.Hostname = hu.Hostname
	gs.mu.Unlock()

	log.Printf("[HANDLER] Hostname updated: %s -> %s (from %s)",
		oldHostname, hu.Hostname, gs.RemoteAddr)
}

// Number of cvar tamper violations before a client gets kicked
const cvarTamperKickLimit = 3

// Length of the window in which violations are counted before kicking
const cvarTamperWindow = 60 * time.Second

// handleCvarChange handles ACC_CVARCHANGE (real-time cvar value changes).
// The game server has already reverted violating values server-side; here we
// count repeated tampering and kick the client after several violations.
func (h *Handler) handleCvarChange(gs *GameServer, cc *protocol.CvarChangeMessage) {
	if cc == nil {
		return
	}

	client := gs.GetClient(cc.ClientID)
	if client == nil {
		log.Printf("[HANDLER] Cvar change for unknown client %d from %s",
			cc.ClientID, gs.RemoteAddr)
		return
	}

	for _, entry := range cc.Cvars {
		check := gs.FindCvarCheck(entry.Name)
		if check == nil {
			// cvar not on the watchlist, ignore
			continue
		}

		if compareCvar(entry.Value, *check) {
			// value passes the check, nothing to do
			continue
		}

		// tampering detected
		now := time.Now()
		if client.CvarTamperLastViol.IsZero() || now.Sub(client.CvarTamperLastViol) > cvarTamperWindow {
			client.CvarTamperCount = 0
		}
		client.CvarTamperLastViol = now
		client.CvarTamperCount++

		playerIP := ""
		playerName := client.Name
		if client.IP != nil {
			playerIP = client.IP.String()
		}

		log.Printf("[HANDLER] Cvar tamper: %s=%s from client %d (violation %d/%d in window) from %s",
			entry.Name, entry.Value, cc.ClientID, client.CvarTamperCount, cvarTamperKickLimit, gs.RemoteAddr)

		if h.OnViolation != nil {
			h.OnViolation(gs.RemoteAddr.String(), playerIP, playerName, cc.ClientID,
				"cvartamper", fmt.Sprintf("cvar %s=%s (tamper)", entry.Name, entry.Value))
		}

		if client.CvarTamperCount >= cvarTamperKickLimit {
			reason := fmt.Sprintf("repeated cvar tampering: %s=%s (%d violations)", entry.Name, entry.Value, client.CvarTamperCount)
			clientMsg := "Anticheat: repeated cvar tampering detected"
			log.Printf("[HANDLER] Kicking client %d from %s: %s", cc.ClientID, gs.RemoteAddr, reason)
			gs.SendViolation(cc.ClientID, cc.Challenge, reason, clientMsg)
			client.CvarTamperCount = 0
		}
	}
}

// handleSpikedModel handles ACC_SPIKEDMODEL (client-reported rejected
// player/weapon geometry). A spiked model is a wallhack vector, so the
// player is kicked immediately and the violation is recorded.
func (h *Handler) handleSpikedModel(gs *GameServer, sm *protocol.SpikedModelMessage) {
	if sm == nil {
		return
	}

	client := gs.GetClient(sm.ClientID)
	playerIP := ""
	playerName := ""
	if client != nil {
		if client.IP != nil {
			playerIP = client.IP.String()
		}
		if client.Name != "" {
			playerName = client.Name
		}
	}

	reason := "spiked model geometry: " + sm.Path
	clientMsg := "Anticheat: geometria de modelo invalida detectada"

	log.Printf("[HANDLER] Spiked model from %s: client=%d, %s, path=%s",
		gs.RemoteAddr, sm.ClientID, playerName, sm.Path)

	if h.OnViolation != nil {
		h.OnViolation(gs.RemoteAddr.String(), playerIP, playerName, sm.ClientID,
			"spikedmodel", reason)
	}

	log.Printf("[HANDLER] Kicking client %d from %s: %s", sm.ClientID, gs.RemoteAddr, reason)
	gs.SendViolation(sm.ClientID, sm.Challenge, reason, clientMsg)
}

// compareCvar checks a client's cvar value against expected check rules.
// Returns TRUE if value PASSES (no violation), FALSE if value FAILS (violation).
//
// The operator describes the violation condition (r1ch convention):
//   - >= <= < > : numeric comparison, violation when value is out of range
//   - =  eq  ~  : denylist, violation when value equals/equals-ci/contains
//   - != ne     : must match one of the listed values, violation otherwise
func compareCvar(value string, check protocol.CvarCheck) bool {
	switch check.Op {
	case protocol.OP_EQUAL:
		// violation when value equals any listed value
		for _, v := range check.Values {
			if value == v {
				return false
			}
		}
		return true
	case protocol.OP_NEQUAL:
		// violation when value matches none of the listed values
		for _, v := range check.Values {
			if value == v {
				return true
			}
		}
		return false
	case protocol.OP_STREQUAL:
		// violation when value equals (case-insensitive) any listed value
		for _, v := range check.Values {
			if strings.EqualFold(value, v) {
				return false
			}
		}
		return true
	case protocol.OP_STRNEQUAL:
		// violation when value matches none of the listed values (case-insensitive)
		for _, v := range check.Values {
			if strings.EqualFold(value, v) {
				return true
			}
		}
		return false
	case protocol.OP_STRSTR:
		// violation when value contains any listed substring
		for _, v := range check.Values {
			if len(value) >= len(v) && strings.Contains(value, v) {
				return false
			}
		}
		return true
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
