package protocol

// AC protocol version (matches q2pro's AC_PROTOCOL_VERSION 0xAC03)
const ProtocolVersion uint16 = 0xAC03

// Server → Client message types (ACS_*)
type ServerByte byte

const (
	ACS_BAD              ServerByte = 0
	ACS_CLIENTACK        ServerByte = 1
	ACS_VIOLATION        ServerByte = 2
	ACS_NOACCESS         ServerByte = 3
	ACS_FILE_VIOLATION   ServerByte = 4
	ACS_READY            ServerByte = 5
	ACS_QUERYREPLY       ServerByte = 6
	ACS_PONG             ServerByte = 7 // NOTE: q2pro calls this ACS_PONG but it's byte 8 in some docs
	ACS_UPDATE_REQUIRED  ServerByte = 8
	ACS_DISCONNECT       ServerByte = 9
	ACS_ERROR            ServerByte = 10
	// New message types for screenshot support
	ACS_REQUEST_SCREENSHOT ServerByte = 11
	ACS_SCREENSHOT_ACK     ServerByte = 12
	// Cvar warning: client should fix cvars and reconnect (not a kick)
	ACS_CVARWARNING ServerByte = 13
)

// Client → Server message types (ACC_*)
type ClientByte byte

const (
	ACC_BAD               ClientByte = 0
	ACC_VERSION           ClientByte = 1
	ACC_PREF              ClientByte = 2
	ACC_REQUESTCHALLENGE  ClientByte = 3
	ACC_CLIENTDISCONNECT  ClientByte = 4
	ACC_QUERYCLIENT       ClientByte = 5
	ACC_PING              ClientByte = 6
	ACC_UPDATECHECKS      ClientByte = 7
	ACC_SETPREFERENCES    ClientByte = 8
	// New message types for screenshot support
	ACC_SCREENSHOT_DATA   ClientByte = 9
	// Client-reported file hashes and cvar values for validation
	ACC_CLIENTDATA        ClientByte = 10
	// Client-reported process snapshot (running processes + loaded modules)
	ACC_PROCESSDATA       ClientByte = 11
	// Player name change notification
	ACC_NAMEUPDATE        ClientByte = 12
	// Server hostname change notification
	ACC_HOSTNAMEUPDATE    ClientByte = 13
)

// Cvar comparison operators
type CvarOp byte

const (
	OP_INVALID    CvarOp = 0
	OP_EQUAL      CvarOp = 1
	OP_NEQUAL     CvarOp = 2
	OP_GTEQUAL    CvarOp = 3
	OP_LTEQUAL    CvarOp = 4
	OP_LT         CvarOp = 5
	OP_GT         CvarOp = 6
	OP_STREQUAL   CvarOp = 7
	OP_STRNEQUAL  CvarOp = 8
	OP_STRSTR     CvarOp = 9
)

// Client types (bitmask, used for internal tracking)
const (
	AC_CLIENT_R1Q2  = 0x01
	AC_CLIENT_EGL   = 0x02
	AC_CLIENT_APRGL = 0x04
	AC_CLIENT_APRSW = 0x08
	AC_CLIENT_Q2PRO = 0x10
)

// Client type indices for ACS_CLIENTACK / ACS_QUERYREPLY wire format.
// q2pro uses the received byte as an index into ac_clients[]:
//   0="???", 1="R1Q2", 2="EGL", 3="Apr GL", 4="Apr SW", 5="Q2PRO"
const (
	AC_TYPE_UNKNOWN byte = 0
	AC_TYPE_R1Q2    byte = 1
	AC_TYPE_EGL     byte = 2
	AC_TYPE_APRGL   byte = 3
	AC_TYPE_APRSW   byte = 4
	AC_TYPE_Q2PRO   byte = 5
)

// File hash flags
const (
	ACH_REQUIRED = 0x01
	ACH_NEGATIVE = 0x02
)

// Preferences flags
const (
	ACP_BLOCKPLAY = 0x01
)

// Client required states
type RequiredState int

const (
	AC_NORMAL   RequiredState = 0
	AC_REQUIRED RequiredState = 1
	AC_EXEMPT   RequiredState = 2
)
