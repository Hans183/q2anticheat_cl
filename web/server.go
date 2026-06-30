package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/user/anticheat_cl/database"
	"github.com/user/anticheat_cl/server"
)

//go:embed static/*
var staticFiles embed.FS

// WebServer serves the dashboard
type WebServer struct {
	listenAddr string
	db         *database.DB
	handler    *server.Handler
	auth       *Auth
	templates  *Templates
	mux        *http.ServeMux
}

// New creates a new web server
func New(listenAddr string, db *database.DB, handler *server.Handler) *WebServer {
	ws := &WebServer{
		listenAddr: listenAddr,
		db:         db,
		handler:    handler,
		auth:       NewAuth(db),
		templates:  NewTemplates(),
		mux:        http.NewServeMux(),
	}
	ws.routes()
	return ws
}

func (ws *WebServer) routes() {
	ws.mux.HandleFunc("/login", ws.handleLogin)
	ws.mux.HandleFunc("/logout", ws.handleLogout)
	staticFS, _ := fs.Sub(staticFiles, "static")
	ws.mux.Handle("/static/", http.StripPrefix("/static/",
		http.FileServer(http.FS(staticFS))))
	ws.mux.Handle("/", ws.authMiddleware(http.HandlerFunc(ws.handleDashboard)))
	ws.mux.Handle("/screenshots", ws.authMiddleware(http.HandlerFunc(ws.handleScreenshots)))
	ws.mux.Handle("/screenshots/review", ws.authMiddleware(http.HandlerFunc(ws.handleReviewScreenshot)))
	ws.mux.Handle("/screenshots/image/", ws.authMiddleware(http.HandlerFunc(ws.handleScreenshotImage)))
	ws.mux.Handle("/violations", ws.authMiddleware(http.HandlerFunc(ws.handleViolations)))
	ws.mux.Handle("/process-snapshots", ws.authMiddleware(http.HandlerFunc(ws.handleProcessSnapshots)))
	ws.mux.Handle("/process-snapshots/", ws.authMiddleware(http.HandlerFunc(ws.handleProcessSnapshotDetail)))
	ws.mux.Handle("/blacklist", ws.authMiddleware(http.HandlerFunc(ws.handleBlacklist)))
	ws.mux.Handle("/servers", ws.authMiddleware(http.HandlerFunc(ws.handleServers)))
	ws.mux.Handle("/api/stats", ws.authMiddleware(http.HandlerFunc(ws.handleAPIStats)))
}

// Start begins listening
func (ws *WebServer) Start() error {
	log.Printf("[WEB] Dashboard listening on http://%s", ws.listenAddr)
	return http.ListenAndServe(ws.listenAddr, ws.mux)
}

func (ws *WebServer) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err != nil || cookie.Value == "" {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		_, err = ws.db.GetSession(cookie.Value)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (ws *WebServer) handleLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if r.Method == "GET" {
		ws.templates.Execute(w, "login", nil)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	admin, err := ws.db.GetAdminByUsername(username)
	if err != nil {
		ws.templates.Execute(w, "login", map[string]interface{}{
			"Error": "Usuario o contrasena invalidos",
		})
		return
	}

	if !CheckPassword(admin.Password, password) {
		ws.templates.Execute(w, "login", map[string]interface{}{
			"Error": "Usuario o contrasena invalidos",
		})
		return
	}

	token, err := ws.db.CreateSession(admin.ID)
	if err != nil {
		ws.templates.Execute(w, "login", map[string]interface{}{
			"Error": "Error creando sesion",
		})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400,
	})
	http.Redirect(w, r, "/", http.StatusFound)
}

func (ws *WebServer) handleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err == nil {
		ws.db.DeleteSession(cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/login", http.StatusFound)
}

func (ws *WebServer) handleDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	stats, _ := ws.db.GetStats()

	servers := ws.handler.GetServers()
	serverCount := len(servers)
	clientCount := 0
	for _, s := range servers {
		s.ClientsMu.RLock()
		clientCount += len(s.Clients)
		s.ClientsMu.RUnlock()
	}

	data := map[string]interface{}{
		"Stats":       stats,
		"ServerCount": serverCount,
		"ClientCount": clientCount,
		"CurrentPage": "dashboard",
	}
	ws.templates.Execute(w, "dashboard", data)
}

func (ws *WebServer) handleScreenshots(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	playerIP := r.URL.Query().Get("player")
	dateFrom := r.URL.Query().Get("from")
	dateTo := r.URL.Query().Get("to")
	unreviewed := r.URL.Query().Get("unreviewed") == "1"
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	screenshots, total, err := ws.db.GetScreenshots(playerIP, dateFrom, dateTo, unreviewed, page, 20)
	if err != nil {
		log.Printf("[WEB] Error getting screenshots: %v", err)
	}
	totalPages := (total + 19) / 20

	data := map[string]interface{}{
		"Screenshots":  screenshots,
		"Total":        total,
		"Page":         page,
		"TotalPages":   totalPages,
		"PlayerIP":     playerIP,
		"DateFrom":     dateFrom,
		"DateTo":       dateTo,
		"Unreviewed":   unreviewed,
		"CurrentPage":  "screenshots",
	}
	ws.templates.Execute(w, "screenshots", data)
}

func (ws *WebServer) handleReviewScreenshot(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/screenshots", http.StatusFound)
		return
	}
	idStr := r.FormValue("id")
	notes := r.FormValue("notes")
	if idStr != "" {
		id, _ := strconv.ParseInt(idStr, 10, 64)
		if id > 0 {
			ws.db.MarkReviewed(id, notes)
		}
	}
	http.Redirect(w, r, "/screenshots", http.StatusFound)
}

func (ws *WebServer) handleScreenshotImage(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/screenshots/image/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id == 0 {
		http.NotFound(w, r)
		return
	}

	rec, err := ws.db.GetScreenshot(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	http.ServeFile(w, r, rec.FilePath)
}

func (ws *WebServer) handleViolations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	playerIP := r.URL.Query().Get("player")
	vType := r.URL.Query().Get("type")
	dateFrom := r.URL.Query().Get("from")
	dateTo := r.URL.Query().Get("to")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	violations, total, err := ws.db.GetViolations(playerIP, vType, dateFrom, dateTo, page, 50)
	if err != nil {
		log.Printf("[WEB] Error getting violations: %v", err)
	}
	totalPages := (total + 49) / 50

	data := map[string]interface{}{
		"Violations":  violations,
		"Total":       total,
		"Page":        page,
		"TotalPages":  totalPages,
		"PlayerIP":    playerIP,
		"Type":        vType,
		"DateFrom":    dateFrom,
		"DateTo":      dateTo,
		"CurrentPage": "violations",
	}
	ws.templates.Execute(w, "violations", data)
}

func (ws *WebServer) handleProcessSnapshots(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	playerIP := r.URL.Query().Get("player")
	dateFrom := r.URL.Query().Get("from")
	dateTo := r.URL.Query().Get("to")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	snapshots, total, err := ws.db.GetProcessSnapshots(playerIP, dateFrom, dateTo, page, 20)
	if err != nil {
		log.Printf("[WEB] Error getting process snapshots: %v", err)
	}
	totalPages := (total + 19) / 20

	data := map[string]interface{}{
		"Snapshots":   snapshots,
		"Total":       total,
		"TotalPages":  totalPages,
		"HasNext":     page < totalPages,
		"CurrentPage": "process-snapshots",
		"PlayerIP":    playerIP,
		"DateFrom":    dateFrom,
		"DateTo":      dateTo,
		"Page":        page,
	}
	ws.templates.Execute(w, "process-snapshots", data)
}

func (ws *WebServer) handleProcessSnapshotDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	idStr := strings.TrimPrefix(r.URL.Path, "/process-snapshots/")
	if idStr == "" || idStr == "process-snapshots" {
		http.Redirect(w, r, "/process-snapshots", http.StatusFound)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id == 0 {
		http.NotFound(w, r)
		return
	}

	snapshot, err := ws.db.GetProcessSnapshotByID(id)
	if err != nil {
		log.Printf("[WEB] Error getting process snapshot: %v", err)
		http.NotFound(w, r)
		return
	}

	type processView struct {
		PID          uint32 `json:"pid"`
		ParentPID    uint32 `json:"parent_pid"`
		Name         string `json:"name"`
		Suspicious   bool
		MatchPattern string
	}
	type moduleView struct {
		Name         string `json:"name"`
		Path         string `json:"path"`
		SHA1         string `json:"sha1"`
		Suspicious   bool
		MatchPattern string
	}

	var processes []processView
	var modules []moduleView

	if snapshot.ProcessesJSON != "" && snapshot.ProcessesJSON != "[]" {
		json.Unmarshal([]byte(snapshot.ProcessesJSON), &processes)
	}
	if snapshot.ModulesJSON != "" && snapshot.ModulesJSON != "[]" {
		json.Unmarshal([]byte(snapshot.ModulesJSON), &modules)
	}

	bl := ws.handler.Blacklist()
	for i := range processes {
		if matched, pattern := bl.CheckProcess(processes[i].Name); matched {
			processes[i].Suspicious = true
			processes[i].MatchPattern = pattern
		}
	}
	for i := range modules {
		if matched, pattern, _ := bl.CheckModuleWithPath(modules[i].Name, modules[i].Path); matched {
			modules[i].Suspicious = true
			modules[i].MatchPattern = pattern
		}
	}

	data := map[string]interface{}{
		"Snapshot":    snapshot,
		"Processes":   processes,
		"Modules":     modules,
		"CurrentPage": "process-snapshots",
	}
	ws.templates.Execute(w, "process-snapshot-detail", data)
}

func (ws *WebServer) handleBlacklist(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if r.Method == "POST" {
		action := r.FormValue("action")
		switch action {
		case "add":
			entryType := r.FormValue("type")
			pattern := r.FormValue("pattern")
			addedBy := r.FormValue("added_by")
			if addedBy == "" {
				addedBy = "admin"
			}
			if entryType != "" && pattern != "" {
				if err := ws.handler.Blacklist().AddEntry(entryType, pattern, addedBy); err != nil {
					log.Printf("[WEB] Error adding blacklist entry: %v", err)
				}
			}
		case "delete":
			idStr := r.FormValue("id")
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err == nil && id > 0 {
				if err := ws.handler.Blacklist().RemoveEntry(id); err != nil {
					log.Printf("[WEB] Error removing blacklist entry: %v", err)
				}
			}
		case "toggle":
			idStr := r.FormValue("id")
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err == nil && id > 0 {
				if err := ws.handler.Blacklist().ToggleEntry(id); err != nil {
					log.Printf("[WEB] Error toggling blacklist entry: %v", err)
				}
			}
		}
		http.Redirect(w, r, "/blacklist", http.StatusFound)
		return
	}

	entries := ws.handler.Blacklist().GetAllEntries()
	processCount, moduleCount, totalEntries := ws.handler.Blacklist().Stats()

	data := map[string]interface{}{
		"Entries":      entries,
		"ProcessCount": processCount,
		"ModuleCount":  moduleCount,
		"TotalEntries": totalEntries,
		"CurrentPage":  "blacklist",
	}
	ws.templates.Execute(w, "blacklist", data)
}

func (ws *WebServer) handleServers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	servers := ws.handler.GetServers()
	data := map[string]interface{}{
		"Servers":     servers,
		"CurrentPage": "servers",
	}
	ws.templates.Execute(w, "servers", data)
}

func (ws *WebServer) handleAPIStats(w http.ResponseWriter, r *http.Request) {
	stats, _ := ws.db.GetStats()
	servers := ws.handler.GetServers()
	stats["servers_online"] = len(servers)
	clientCount := 0
	for _, s := range servers {
		s.ClientsMu.RLock()
		clientCount += len(s.Clients)
		s.ClientsMu.RUnlock()
	}
	stats["clients_online"] = clientCount

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (ws *WebServer) handleAPIScreenshotReview(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ID    int64  `json:"id"`
		Notes string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := ws.db.MarkReviewed(req.ID, req.Notes); err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"ok":true}`)
}
