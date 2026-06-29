package server

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/user/anticheat_cl/screenshots"
)

// TCPServer listens for connections from game servers
type TCPServer struct {
	listenAddr string
	listener   net.Listener
	handler    *Handler
	running    bool
	mu         sync.Mutex
}

// New creates a new TCP server
func New(listenAddr string, storage *screenshots.Storage) *TCPServer {
	return &TCPServer{
		listenAddr: listenAddr,
		handler:    NewHandler(storage),
	}
}

// Start begins listening for connections
func (s *TCPServer) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var err error
	s.listener, err = net.Listen("tcp", s.listenAddr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	s.running = true
	log.Printf("[TCP] Listening on %s", s.listenAddr)

	// Accept loop
	go s.acceptLoop()

	// Timeout checker
	go s.timeoutLoop()

	return nil
}

// Stop stops the server
func (s *TCPServer) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.listener != nil {
		s.listener.Close()
	}
	s.running = false
	log.Printf("[TCP] Server stopped")
}

// acceptLoop accepts new connections
func (s *TCPServer) acceptLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.running {
				log.Printf("[TCP] Accept error: %v", err)
			}
			return
		}

		log.Printf("[TCP] New connection from %s", conn.RemoteAddr())

		// Create game server and handle in goroutine
		gs := NewGameServer(conn)
		go gs.Handle(s.handler.HandleMessage, func(addr string) {
			s.handler.RemoveServer(addr)
		})
	}
}

// timeoutLoop periodically checks for timeouts
func (s *TCPServer) timeoutLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if !s.running {
			return
		}
		s.handler.CheckTimeouts()
	}
}

// GetHandler returns the message handler
func (s *TCPServer) GetHandler() *Handler {
	return s.handler
}
