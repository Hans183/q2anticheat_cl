package protocol

import (
	"fmt"
	"net"
)

// Message represents a parsed protocol message
type Message struct {
	Type    ClientByte
	Version *VersionMessage
	Challenge *ChallengeMessage
	Disconnect *DisconnectMessage
	QueryClient *QueryClientMessage
	Ping     bool
	Checks   *ChecksMessage
	Prefs    *PrefsMessage
	Screenshot *ScreenshotData
	ClientData *ClientDataMessage
	ProcessData *ProcessDataMessage
}

// ClientFileData represents a single file hash reported by the client
type ClientFileData struct {
	Path string
	Hash [20]byte
}

// ClientCvarData represents a single cvar value reported by the client
type ClientCvarData struct {
	Name  string
	Value string
}

// ClientDataMessage contains file hashes and cvar values reported by a client (ACC_CLIENTDATA)
type ClientDataMessage struct {
	ClientID   uint32
	Challenge  uint32
	PlayerName string
	Files      []ClientFileData
	Cvars      []ClientCvarData
}

// ProcessEntry represents a single running process
type ProcessEntry struct {
	PID      uint32
	ParentPID uint32
	Name     string
}

// ModuleEntry represents a single loaded module/library
type ModuleEntry struct {
	Name   string
	Path   string
	SHA1   [20]byte
}

// ProcessDataMessage contains a process snapshot from a client (ACC_PROCESSDATA)
type ProcessDataMessage struct {
	ClientID    uint32
	Challenge   uint32
	PlayerName  string
	Processes   []ProcessEntry
	Modules     []ModuleEntry
}

// VersionMessage is sent by q2pro server during handshake (ACC_VERSION)
type VersionMessage struct {
	ProtocolVersion uint16
	Hostname        string
	Version         string
	Port            uint32
}

// ChallengeMessage is sent to request client validation (ACC_REQUESTCHALLENGE)
type ChallengeMessage struct {
	IP        net.IP
	Port      uint16
	ClientID  uint32
	Challenge uint32
}

// DisconnectMessage is sent when a client disconnects (ACC_CLIENTDISCONNECT)
type DisconnectMessage struct {
	ClientID  uint32
	Challenge uint32
}

// QueryClientMessage is sent to query a client's status (ACC_QUERYCLIENT)
type QueryClientMessage struct {
	ClientID  uint32
	Challenge uint32
}

// ChecksMessage contains file hashes and cvar checks (ACC_UPDATECHECKS)
type ChecksMessage struct {
	NumFiles uint32
	NumCvars uint32
	Files    []FileHash
	Cvars    []CvarCheck
}

// FileHash represents a single file hash entry
type FileHash struct {
	Hash  [20]byte
	Flags byte
	Path  string
}

// CvarCheck represents a single cvar validation rule
type CvarCheck struct {
	Name      string
	Op        CvarOp
	NumValues uint8
	Values    []string
	Default   string
}

// PrefsMessage contains server preferences (ACC_SETPREFERENCES)
type PrefsMessage struct {
	Flags uint32
}

// ScreenshotData contains a forwarded screenshot from a client (ACC_SCREENSHOT_DATA)
type ScreenshotData struct {
	ClientID  uint32
	Challenge uint32
	Width     uint16
	Height    uint16
	JPEGSize  uint32
	JPEGData  []byte
}

// ParseMessage parses a raw message buffer into a Message struct
func ParseMessage(buf []byte) (*Message, error) {
	if len(buf) < 1 {
		return nil, fmt.Errorf("empty message")
	}

	msg := &Message{
		Type: ClientByte(buf[0]),
	}

	r := &Reader{buf: buf[1:], pos: 0}

	switch msg.Type {
	case ACC_VERSION:
		err := parseVersionMessage(r, msg)
		if err != nil {
			return nil, fmt.Errorf("parse version: %w", err)
		}

	case ACC_REQUESTCHALLENGE:
		err := parseChallengeMessage(r, msg)
		if err != nil {
			return nil, fmt.Errorf("parse challenge: %w", err)
		}

	case ACC_CLIENTDISCONNECT:
		err := parseDisconnectMessage(r, msg)
		if err != nil {
			return nil, fmt.Errorf("parse disconnect: %w", err)
		}

	case ACC_QUERYCLIENT:
		err := parseQueryClientMessage(r, msg)
		if err != nil {
			return nil, fmt.Errorf("parse query client: %w", err)
		}

	case ACC_PING:
		msg.Ping = true

	case ACC_UPDATECHECKS:
		err := parseChecksMessage(r, msg)
		if err != nil {
			return nil, fmt.Errorf("parse checks: %w", err)
		}

	case ACC_SETPREFERENCES:
		err := parsePrefsMessage(r, msg)
		if err != nil {
			return nil, fmt.Errorf("parse prefs: %w", err)
		}

	case ACC_SCREENSHOT_DATA:
		err := parseScreenshotData(r, msg)
		if err != nil {
			return nil, fmt.Errorf("parse screenshot: %w", err)
		}

	case ACC_CLIENTDATA:
		err := parseClientDataMessage(r, msg)
		if err != nil {
			return nil, fmt.Errorf("parse client data: %w", err)
		}

	case ACC_PROCESSDATA:
		err := parseProcessDataMessage(r, msg)
		if err != nil {
			return nil, fmt.Errorf("parse process data: %w", err)
		}

	default:
		return nil, fmt.Errorf("unknown message type: %d", msg.Type)
	}

	return msg, nil
}

func parseVersionMessage(r *Reader, msg *Message) error {
	ver, err := r.ReadUint16()
	if err != nil {
		return err
	}

	// q2pro writes hostname as: uint16 hostlen + raw data (NOT length-prefixed string)
	hostlen, err := r.ReadUint16()
	if err != nil {
		return err
	}
	hostnameBytes, err := r.ReadBytes(int(hostlen))
	if err != nil {
		return err
	}

	// q2pro writes version string as: uint16 verlen + raw data
	verlen, err := r.ReadUint16()
	if err != nil {
		return err
	}
	versionBytes, err := r.ReadBytes(int(verlen))
	if err != nil {
		return err
	}

	port, err := r.ReadUint32()
	if err != nil {
		return err
	}

	msg.Version = &VersionMessage{
		ProtocolVersion: ver,
		Hostname:        string(hostnameBytes),
		Version:         string(versionBytes),
		Port:            port,
	}
	return nil
}

func parseChallengeMessage(r *Reader, msg *Message) error {
	ipBytes, err := r.ReadBytes(4)
	if err != nil {
		return err
	}

	port, err := r.ReadUint16()
	if err != nil {
		return err
	}

	clientID, err := r.ReadUint32()
	if err != nil {
		return err
	}

	challenge, err := r.ReadUint32()
	if err != nil {
		return err
	}

	msg.Challenge = &ChallengeMessage{
		IP:        net.IP(ipBytes),
		Port:      port,
		ClientID:  clientID,
		Challenge: challenge,
	}
	return nil
}

func parseDisconnectMessage(r *Reader, msg *Message) error {
	clientID, err := r.ReadUint32()
	if err != nil {
		return err
	}

	challenge, err := r.ReadUint32()
	if err != nil {
		return err
	}

	msg.Disconnect = &DisconnectMessage{
		ClientID:  clientID,
		Challenge: challenge,
	}
	return nil
}

func parseQueryClientMessage(r *Reader, msg *Message) error {
	clientID, err := r.ReadUint32()
	if err != nil {
		return err
	}

	challenge, err := r.ReadUint32()
	if err != nil {
		return err
	}

	msg.QueryClient = &QueryClientMessage{
		ClientID:  clientID,
		Challenge: challenge,
	}
	return nil
}

func parseChecksMessage(r *Reader, msg *Message) error {
	numFiles, err := r.ReadUint32()
	if err != nil {
		return err
	}

	numCvars, err := r.ReadUint32()
	if err != nil {
		return err
	}

	checks := &ChecksMessage{
		NumFiles: numFiles,
		NumCvars: numCvars,
		Files:    make([]FileHash, 0, numFiles),
		Cvars:    make([]CvarCheck, 0, numCvars),
	}

	// Parse file hashes
	for i := uint32(0); i < numFiles; i++ {
		hashBytes, err := r.ReadBytes(20)
		if err != nil {
			return fmt.Errorf("read file hash %d: %w", i, err)
		}

		flags, err := r.ReadByte()
		if err != nil {
			return fmt.Errorf("read file flags %d: %w", i, err)
		}

		path, err := r.ReadString()
		if err != nil {
			return fmt.Errorf("read file path %d: %w", i, err)
		}

		var hash [20]byte
		copy(hash[:], hashBytes)

		checks.Files = append(checks.Files, FileHash{
			Hash:  hash,
			Flags: flags,
			Path:  path,
		})
	}

	// Parse cvar checks
	for i := uint32(0); i < numCvars; i++ {
		name, err := r.ReadString()
		if err != nil {
			return fmt.Errorf("read cvar name %d: %w", i, err)
		}

		op, err := r.ReadByte()
		if err != nil {
			return fmt.Errorf("read cvar op %d: %w", i, err)
		}

		numValues, err := r.ReadByte()
		if err != nil {
			return fmt.Errorf("read cvar num values %d: %w", i, err)
		}

		values := make([]string, numValues)
		for j := byte(0); j < numValues; j++ {
			values[j], err = r.ReadString()
			if err != nil {
				return fmt.Errorf("read cvar value %d/%d: %w", i, j, err)
			}
		}

		def, err := r.ReadString()
		if err != nil {
			return fmt.Errorf("read cvar default %d: %w", i, err)
		}

		checks.Cvars = append(checks.Cvars, CvarCheck{
			Name:      name,
			Op:        CvarOp(op),
			NumValues: numValues,
			Values:    values,
			Default:   def,
		})
	}

	msg.Checks = checks
	return nil
}

func parsePrefsMessage(r *Reader, msg *Message) error {
	flags, err := r.ReadUint32()
	if err != nil {
		return err
	}

	msg.Prefs = &PrefsMessage{Flags: flags}
	return nil
}

func parseScreenshotData(r *Reader, msg *Message) error {
	clientID, err := r.ReadUint32()
	if err != nil {
		return err
	}

	challenge, err := r.ReadUint32()
	if err != nil {
		return err
	}

	width, err := r.ReadUint16()
	if err != nil {
		return err
	}

	height, err := r.ReadUint16()
	if err != nil {
		return err
	}

	jpegSize, err := r.ReadUint32()
	if err != nil {
		return err
	}

	jpegData, err := r.ReadBytes(int(jpegSize))
	if err != nil {
		return err
	}

	msg.Screenshot = &ScreenshotData{
		ClientID:  clientID,
		Challenge: challenge,
		Width:     width,
		Height:    height,
		JPEGSize:  jpegSize,
		JPEGData:  jpegData,
	}
	return nil
}

func parseClientDataMessage(r *Reader, msg *Message) error {
	clientID, err := r.ReadUint32()
	if err != nil {
		return err
	}

	challenge, err := r.ReadUint32()
	if err != nil {
		return err
	}

	playerName, err := r.ReadString()
	if err != nil {
		return err
	}

	numFiles, err := r.ReadUint32()
	if err != nil {
		return err
	}

	numCvars, err := r.ReadUint32()
	if err != nil {
		return err
	}

	files := make([]ClientFileData, numFiles)
	for i := uint32(0); i < numFiles; i++ {
		hashBytes, err := r.ReadBytes(20)
		if err != nil {
			return fmt.Errorf("read file hash %d: %w", i, err)
		}

		path, err := r.ReadString()
		if err != nil {
			return fmt.Errorf("read file path %d: %w", i, err)
		}

		var hash [20]byte
		copy(hash[:], hashBytes)
		files[i] = ClientFileData{Path: path, Hash: hash}
	}

	cvars := make([]ClientCvarData, numCvars)
	for i := uint32(0); i < numCvars; i++ {
		name, err := r.ReadString()
		if err != nil {
			return fmt.Errorf("read cvar name %d: %w", i, err)
		}

		value, err := r.ReadString()
		if err != nil {
			return fmt.Errorf("read cvar value %d: %w", i, err)
		}

		cvars[i] = ClientCvarData{Name: name, Value: value}
	}

	msg.ClientData = &ClientDataMessage{
		ClientID:   clientID,
		Challenge:  challenge,
		PlayerName: playerName,
		Files:      files,
		Cvars:      cvars,
	}
	return nil
}

func parseProcessDataMessage(r *Reader, msg *Message) error {
	clientID, err := r.ReadUint32()
	if err != nil {
		return err
	}

	challenge, err := r.ReadUint32()
	if err != nil {
		return err
	}

	playerName, err := r.ReadString()
	if err != nil {
		return err
	}

	numProcesses, err := r.ReadUint32()
	if err != nil {
		return err
	}

	processes := make([]ProcessEntry, numProcesses)
	for i := uint32(0); i < numProcesses; i++ {
		pid, err := r.ReadUint32()
		if err != nil {
			return fmt.Errorf("read process pid %d: %w", i, err)
		}

		parentPid, err := r.ReadUint32()
		if err != nil {
			return fmt.Errorf("read process parent_pid %d: %w", i, err)
		}

		name, err := r.ReadString()
		if err != nil {
			return fmt.Errorf("read process name %d: %w", i, err)
		}

		processes[i] = ProcessEntry{
			PID:      pid,
			ParentPID: parentPid,
			Name:     name,
		}
	}

	numModules, err := r.ReadUint32()
	if err != nil {
		return err
	}

	modules := make([]ModuleEntry, numModules)
	for i := uint32(0); i < numModules; i++ {
		name, err := r.ReadString()
		if err != nil {
			return fmt.Errorf("read module name %d: %w", i, err)
		}

		path, err := r.ReadString()
		if err != nil {
			return fmt.Errorf("read module path %d: %w", i, err)
		}

		sha1Bytes, err := r.ReadBytes(20)
		if err != nil {
			return fmt.Errorf("read module sha1 %d: %w", i, err)
		}

		var sha1 [20]byte
		copy(sha1[:], sha1Bytes)

		modules[i] = ModuleEntry{
			Name: name,
			Path: path,
			SHA1: sha1,
		}
	}

	msg.ProcessData = &ProcessDataMessage{
		ClientID:   clientID,
		Challenge:  challenge,
		PlayerName: playerName,
		Processes:  processes,
		Modules:    modules,
	}
	return nil
}
