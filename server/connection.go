package server

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/user/anticheat_cl/protocol"
)

// ServerState represents the state of a connected game server
type ServerState int

const (
	StateDisconnected ServerState = iota
	StateConnected
	StateReady
)

// GameServer represents a connected q2pro game server
type GameServer struct {
	conn       net.Conn
	State      ServerState
	Hostname   string
	Version    string
	Port       uint32
	RemoteAddr net.Addr

	// Client tracking
	Clients   map[uint32]*ClientInfo
	ClientsMu sync.RWMutex

	// File checks received from this server
	FileChecks []protocol.FileHash
	CvarChecks []protocol.CvarCheck

	// Preferences
	Prefs uint32

	// Timing
	ConnectedAt time.Time
	LastPingAt  time.Time
	PingPending bool

	mu sync.Mutex
}

// ClientInfo tracks a player connected to this game server
type ClientInfo struct {
	ClientID     uint32
	Challenge    uint32
	Name         string
	IP           net.IP
	Valid        bool
	QuerySent    bool
	Required     int
	FileFailures int
	ClientType   byte
	BadFiles     []string
	Token        string
}

// NewGameServer creates a new GameServer from an accepted connection
func NewGameServer(conn net.Conn) *GameServer {
	return &GameServer{
		conn:        conn,
		State:       StateDisconnected,
		Clients:     make(map[uint32]*ClientInfo),
		RemoteAddr:  conn.RemoteAddr(),
		ConnectedAt: time.Now(),
	}
}

// Handle processes messages from this game server
func (gs *GameServer) Handle(msgHandler func(*GameServer, []byte), onDisconnect func(string)) {
	defer func() {
		gs.Close()
		if onDisconnect != nil {
			onDisconnect(gs.RemoteAddr.String())
		}
	}()

	gs.mu.Lock()
	gs.State = StateConnected
	gs.mu.Unlock()

	log.Printf("[SERVER] New connection from %s", gs.RemoteAddr)

	// q2pro sends a 0x02 framing byte before the first message (ACC_VERSION hello).
	// Read and discard it if present.
	gs.conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	var magicByte [1]byte
	if _, err := io.ReadFull(gs.conn, magicByte[:]); err != nil {
		log.Printf("[SERVER] Error reading initial byte from %s: %v", gs.RemoteAddr, err)
		return
	}
	if magicByte[0] != 0x02 {
		log.Printf("[SERVER] Warning: unexpected initial byte 0x%02X from %s, treating as length hi-byte",
			magicByte[0], gs.RemoteAddr)
	}

	for {
		// Set read timeout (120s gives plenty of margin over q2pro's 60s ping interval)
		gs.conn.SetReadDeadline(time.Now().Add(120 * time.Second))

		// Read 2-byte length prefix (little-endian, matching q2pro's WL16)
		var msgLen uint16
		if err := binary.Read(gs.conn, binary.LittleEndian, &msgLen); err != nil {
			if err != io.EOF {
				log.Printf("[SERVER] Error reading from %s: %v", gs.RemoteAddr, err)
			}
			return
		}

		if msgLen == 0 {
			continue
		}

		// Read payload
		buf := make([]byte, msgLen)
		if _, err := io.ReadFull(gs.conn, buf); err != nil {
			log.Printf("[SERVER] Error reading payload from %s: %v", gs.RemoteAddr, err)
			return
		}

		// q2pro's AC_SendChecks() sends file/cvar entries via AC_Flush()
		// WITHOUT length prefixes. The initial header (9 bytes) is framed,
		// but the per-entry data is raw on the wire. Detect this and read
		// the entries directly from the connection, appending to buf.
		if len(buf) >= 5 && buf[0] == byte(protocol.ACC_UPDATECHECKS) {
			numFiles := binary.LittleEndian.Uint32(buf[1:5])
			numCvars := uint32(0)
			if len(buf) >= 9 {
				numCvars = binary.LittleEndian.Uint32(buf[5:9])
			}
			entries, err := gs.readChecksEntries(numFiles, numCvars)
			if err != nil {
				log.Printf("[SERVER] Error reading checks entries from %s: %v", gs.RemoteAddr, err)
				return
			}
			buf = append(buf, entries...)
		}

		// Call handler
		msgHandler(gs, buf)
	}
}

// SendMessage sends a message to this game server
func (gs *GameServer) SendMessage(msgType byte, payload []byte) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	totalLen := len(payload) + 1 // +1 for message type byte
	if totalLen > 65535 {
		return fmt.Errorf("message too large: %d bytes", totalLen)
	}

	// Write length prefix (little-endian, matching q2pro's LittleShort)
	var header [2]byte
	binary.LittleEndian.PutUint16(header[:], uint16(totalLen))
	if _, err := gs.conn.Write(header[:]); err != nil {
		return err
	}

	// Write message type
	if _, err := gs.conn.Write([]byte{msgType}); err != nil {
		return err
	}

	// Write payload
	if _, err := gs.conn.Write(payload); err != nil {
		return err
	}

	return nil
}

// SendClientAck sends ACS_CLIENTACK to acknowledge a client
func (gs *GameServer) SendClientAck(clientID, challenge uint32, clientType byte) error {
	var payload [7]byte
	binary.LittleEndian.PutUint16(payload[0:], uint16(clientID))
	binary.LittleEndian.PutUint32(payload[2:], challenge)
	payload[6] = clientType

	return gs.SendMessage(byte(protocol.ACS_CLIENTACK), payload[:])
}

// SendViolation sends ACS_VIOLATION to kick a cheating client
func (gs *GameServer) SendViolation(clientID, challenge uint32, reason, clientMsg string) error {
	payload := make([]byte, 0, 8+len(reason)+1+len(clientMsg)+1)
	payload = binary.LittleEndian.AppendUint16(payload, uint16(clientID))
	payload = binary.LittleEndian.AppendUint32(payload, challenge)
	payload = append(payload, []byte(reason)...)
	payload = append(payload, 0x00)
	payload = append(payload, []byte(clientMsg)...)
	payload = append(payload, 0x00)

	return gs.SendMessage(byte(protocol.ACS_VIOLATION), payload)
}

// SendFileViolation sends ACS_FILE_VIOLATION for modified files
func (gs *GameServer) SendFileViolation(clientID, challenge uint32, path, hash string) error {
	payload := make([]byte, 0, 8+len(path)+1+len(hash)+1)
	payload = binary.LittleEndian.AppendUint16(payload, uint16(clientID))
	payload = binary.LittleEndian.AppendUint32(payload, challenge)
	payload = append(payload, []byte(path)...)
	payload = append(payload, 0x00)
	payload = append(payload, []byte(hash)...)
	payload = append(payload, 0x00)

	return gs.SendMessage(byte(protocol.ACS_FILE_VIOLATION), payload)
}

// SendCvarWarning sends ACS_CVARWARNING for cvar-only violations (no kick, client reconnects with corrected values)
func (gs *GameServer) SendCvarWarning(clientID, challenge uint32, reason string) error {
	payload := make([]byte, 0, 8+len(reason)+1)
	payload = binary.LittleEndian.AppendUint16(payload, uint16(clientID))
	payload = binary.LittleEndian.AppendUint32(payload, challenge)
	payload = append(payload, []byte(reason)...)
	payload = append(payload, 0x00)

	return gs.SendMessage(byte(protocol.ACS_CVARWARNING), payload)
}

// SendReady sends ACS_READY to indicate the server is ready
func (gs *GameServer) SendReady() error {
	gs.mu.Lock()
	gs.State = StateReady
	gs.mu.Unlock()

	return gs.SendMessage(byte(protocol.ACS_READY), nil)
}

// SendQueryReply sends ACS_QUERYREPLY with client status
func (gs *GameServer) SendQueryReply(clientID, challenge uint32, valid bool, clientType byte) error {
	var validByte byte
	if valid {
		validByte = 1
	}

	var payload [8]byte
	binary.LittleEndian.PutUint16(payload[0:], uint16(clientID))
	binary.LittleEndian.PutUint32(payload[2:], challenge)
	payload[6] = validByte
	payload[7] = clientType

	return gs.SendMessage(byte(protocol.ACS_QUERYREPLY), payload[:])
}

// SendPong sends ACS_PONG in response to ping
func (gs *GameServer) SendPong() error {
	gs.mu.Lock()
	gs.PingPending = false
	gs.LastPingAt = time.Now()
	gs.mu.Unlock()

	return gs.SendMessage(byte(protocol.ACS_PONG), nil)
}

// SendRequestScreenshot asks a game server to capture a screenshot
func (gs *GameServer) SendRequestScreenshot(clientID, challenge uint32) error {
	var payload [6]byte
	binary.LittleEndian.PutUint16(payload[0:], uint16(clientID))
	binary.LittleEndian.PutUint32(payload[2:], challenge)

	return gs.SendMessage(byte(protocol.ACS_REQUEST_SCREENSHOT), payload[:])
}

// SendScreenshotAck acknowledges receipt of a screenshot
func (gs *GameServer) SendScreenshotAck(clientID, challenge uint32, status byte) error {
	var payload [7]byte
	binary.LittleEndian.PutUint16(payload[0:], uint16(clientID))
	binary.LittleEndian.PutUint32(payload[2:], challenge)
	payload[6] = status

	return gs.SendMessage(byte(protocol.ACS_SCREENSHOT_ACK), payload[:])
}

// Close closes the connection
func (gs *GameServer) Close() {
	gs.conn.Close()
	gs.mu.Lock()
	gs.State = StateDisconnected
	gs.mu.Unlock()
}

// GetState returns the current server state
func (gs *GameServer) GetState() ServerState {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	return gs.State
}

// GetClient returns a client by ID
func (gs *GameServer) GetClient(clientID uint32) *ClientInfo {
	gs.ClientsMu.RLock()
	defer gs.ClientsMu.RUnlock()
	return gs.Clients[clientID]
}

// SetClient adds or updates a client
func (gs *GameServer) SetClient(client *ClientInfo) {
	gs.ClientsMu.Lock()
	defer gs.ClientsMu.Unlock()
	gs.Clients[client.ClientID] = client
}

// RemoveClient removes a client
func (gs *GameServer) RemoveClient(clientID uint32) {
	gs.ClientsMu.Lock()
	defer gs.ClientsMu.Unlock()
	delete(gs.Clients, clientID)
}

// FindFileCheck returns the file check rule for a given path
func (gs *GameServer) FindFileCheck(path string) *protocol.FileHash {
	for i := range gs.FileChecks {
		if gs.FileChecks[i].Path == path {
			return &gs.FileChecks[i]
		}
	}
	return nil
}

// FindCvarCheck returns the cvar check rule for a given name
func (gs *GameServer) FindCvarCheck(name string) *protocol.CvarCheck {
	for i := range gs.CvarChecks {
		if gs.CvarChecks[i].Name == name {
			return &gs.CvarChecks[i]
		}
	}
	return nil
}

// ClientTypeString returns a human-readable client type name
func (ci *ClientInfo) ClientTypeString() string {
	types := []string{"???", "R1Q2", "EGL", "Apr GL", "Apr SW", "Q2PRO"}
	if int(ci.ClientType) < len(types) {
		return types[ci.ClientType]
	}
	return "???"
}

// Uptime returns how long this server has been connected
func (gs *GameServer) Uptime() time.Duration {
	return time.Since(gs.ConnectedAt)
}

// readChecksEntries reads file/cvar check entries directly from the TCP connection.
func (gs *GameServer) readChecksEntries(numFiles, numCvars uint32) ([]byte, error) {
	var buf bytes.Buffer

	// Read file entries
	for i := uint32(0); i < numFiles; i++ {
		hash := make([]byte, 20)
		if _, err := io.ReadFull(gs.conn, hash); err != nil {
			return nil, fmt.Errorf("read file hash %d: %w", i, err)
		}
		buf.Write(hash)

		var flags [1]byte
		if _, err := io.ReadFull(gs.conn, flags[:]); err != nil {
			return nil, fmt.Errorf("read file flags %d: %w", i, err)
		}
		buf.Write(flags[:])

		var pathLen [1]byte
		if _, err := io.ReadFull(gs.conn, pathLen[:]); err != nil {
			return nil, fmt.Errorf("read file path len %d: %w", i, err)
		}
		buf.Write(pathLen[:])

		if pathLen[0] > 0 {
			path := make([]byte, pathLen[0])
			if _, err := io.ReadFull(gs.conn, path); err != nil {
				return nil, fmt.Errorf("read file path %d: %w", i, err)
			}
			buf.Write(path)
		}
	}

	// Read cvar entries
	for i := uint32(0); i < numCvars; i++ {
		if err := gs.readACString(&buf); err != nil {
			return nil, fmt.Errorf("read cvar name %d: %w", i, err)
		}

		var op [1]byte
		if _, err := io.ReadFull(gs.conn, op[:]); err != nil {
			return nil, fmt.Errorf("read cvar op %d: %w", i, err)
		}
		buf.Write(op[:])

		var numVals [1]byte
		if _, err := io.ReadFull(gs.conn, numVals[:]); err != nil {
			return nil, fmt.Errorf("read cvar numValues %d: %w", i, err)
		}
		buf.Write(numVals[:])

		for j := byte(0); j < numVals[0]; j++ {
			if err := gs.readACString(&buf); err != nil {
				return nil, fmt.Errorf("read cvar value %d/%d: %w", i, j, err)
			}
		}

		if err := gs.readACString(&buf); err != nil {
			return nil, fmt.Errorf("read cvar default %d: %w", i, err)
		}
	}

	return buf.Bytes(), nil
}

func (gs *GameServer) readACString(buf *bytes.Buffer) error {
	var lenByte [1]byte
	if _, err := io.ReadFull(gs.conn, lenByte[:]); err != nil {
		return err
	}
	buf.Write(lenByte[:])

	if lenByte[0] > 0 {
		data := make([]byte, lenByte[0])
		if _, err := io.ReadFull(gs.conn, data); err != nil {
			return err
		}
		buf.Write(data)
	}
	return nil
}
