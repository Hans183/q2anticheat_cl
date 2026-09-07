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
	pushMgr    *PushManager
}

// New creates a new web server
func New(listenAddr string, db *database.DB, handler *server.Handler) *WebServer {
	pushMgr, err := NewPushManager(db)
	if err != nil {
		log.Printf("[WEB] Warning initializing push manager: %v", err)
	}

	ws := &WebServer{
		listenAddr: listenAddr,
		db:         db,
		handler:    handler,
		auth:       NewAuth(db),
		templates:  NewTemplates(),
		mux:        http.NewServeMux(),
		pushMgr:    pushMgr,
	}
	ws.routes()
	return ws
}

func (ws *WebServer) routes() {
	ws.mux.HandleFunc("/login", ws.handleLogin)
	ws.mux.HandleFunc("/logout", ws.handleLogout)
	ws.mux.HandleFunc("/sw.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		w.Header().Set("Service-Worker-Allowed", "/")
		w.Header().Set("Cache-Control", "no-cache")
		data, err := staticFiles.ReadFile("static/sw.js")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Write(data)
	})
	ws.mux.HandleFunc("/manifest.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/manifest+json")
		data, err := staticFiles.ReadFile("static/manifest.json")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Write(data)
	})
	staticFS, _ := fs.Sub(staticFiles, "static")
	ws.mux.Handle("/static/", http.StripPrefix("/static/",
		http.FileServer(http.FS(staticFS))))
	ws.mux.Handle("/", ws.authMiddleware(http.HandlerFunc(ws.handleDashboard)))
	ws.mux.Handle("/player", ws.authMiddleware(http.HandlerFunc(ws.handlePlayerProfile)))
	ws.mux.Handle("/screenshots", ws.authMiddleware(http.HandlerFunc(ws.handleScreenshots)))
	ws.mux.Handle("/screenshots/review", ws.authMiddleware(http.HandlerFunc(ws.handleReviewScreenshot)))
	ws.mux.Handle("/screenshots/bulk-review", ws.authMiddleware(http.HandlerFunc(ws.handleBulkReviewScreenshots)))
	ws.mux.Handle("/screenshots/export.csv", ws.authMiddleware(http.HandlerFunc(ws.handleExportScreenshotsCSV)))
	ws.mux.Handle("/screenshots/image/", ws.authMiddleware(http.HandlerFunc(ws.handleScreenshotImage)))
	ws.mux.Handle("/violations", ws.authMiddleware(http.HandlerFunc(ws.handleViolations)))
	ws.mux.Handle("/violations/export.csv", ws.authMiddleware(http.HandlerFunc(ws.handleExportViolationsCSV)))
	ws.mux.Handle("/process-snapshots", ws.authMiddleware(http.HandlerFunc(ws.handleProcessSnapshots)))
	ws.mux.Handle("/process-snapshots/", ws.authMiddleware(http.HandlerFunc(ws.handleProcessSnapshotDetail)))
	ws.mux.Handle("/blacklist", ws.authMiddleware(http.HandlerFunc(ws.handleBlacklist)))
	ws.mux.Handle("/servers", ws.authMiddleware(http.HandlerFunc(ws.handleServers)))
	ws.mux.Handle("/settings", ws.authMiddleware(http.HandlerFunc(ws.handleSettings)))
	ws.mux.Handle("/change-password", ws.authMiddleware(http.HandlerFunc(ws.handleChangePassword)))
	ws.mux.Handle("/maintenance/purge-screenshots", ws.authMiddleware(http.HandlerFunc(ws.handlePurgeScreenshots)))
	ws.mux.Handle("/api/stats", ws.authMiddleware(http.HandlerFunc(ws.handleAPIStats)))
	ws.mux.Handle("/api/push/vapid-key", ws.authMiddleware(http.HandlerFunc(ws.handleAPIPushVapidKey)))
	ws.mux.Handle("/api/push/subscribe", ws.authMiddleware(http.HandlerFunc(ws.handleAPIPushSubscribe)))
	ws.mux.Handle("/api/push/unsubscribe", ws.authMiddleware(http.HandlerFunc(ws.handleAPIPushUnsubscribe)))
	ws.mux.Handle("/api/push/test", ws.authMiddleware(http.HandlerFunc(ws.handleAPIPushTest)))
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

	recentViolations, _ := ws.db.GetRecentViolations(5)
	recentScreenshots, _ := ws.db.GetRecentScreenshots(6)

	data := map[string]interface{}{
		"Stats":             stats,
		"ServerCount":       serverCount,
		"ClientCount":       clientCount,
		"RecentViolations":  recentViolations,
		"RecentScreenshots": recentScreenshots,
		"DailyViolations":   stats["daily_violations"],
		"CurrentPage":       "dashboard",
	}
	ws.templates.Execute(w, "dashboard", data)
}

func (ws *WebServer) handlePlayerProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		q = strings.TrimSpace(r.URL.Query().Get("name"))
	}
	if q == "" {
		q = strings.TrimSpace(r.URL.Query().Get("player"))
	}

	if q == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	profile, err := ws.db.GetPlayerProfile(q)
	if err != nil {
		log.Printf("[WEB] Error getting player profile: %v", err)
	}

	data := map[string]interface{}{
		"Profile":     profile,
		"Query":       q,
		"CurrentPage": "players",
	}
	ws.templates.Execute(w, "player", data)
}

func (ws *WebServer) handleScreenshots(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	playerIP := strings.TrimSpace(r.URL.Query().Get("player"))
	playerName := strings.TrimSpace(r.URL.Query().Get("name"))
	serverAddr := strings.TrimSpace(r.URL.Query().Get("server"))
	dateFrom := strings.TrimSpace(r.URL.Query().Get("from"))
	dateTo := strings.TrimSpace(r.URL.Query().Get("to"))
	unreviewed := r.URL.Query().Get("unreviewed") == "1"
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if perPage != 20 && perPage != 50 && perPage != 100 {
		perPage = 20
	}

	screenshots, total, err := ws.db.GetScreenshots(playerIP, playerName, serverAddr, dateFrom, dateTo, unreviewed, page, perPage)
	if err != nil {
		log.Printf("[WEB] Error getting screenshots: %v", err)
	}
	totalPages := (total + perPage - 1) / perPage
	distinctServers, _ := ws.db.GetDistinctServers()

	data := map[string]interface{}{
		"Screenshots":  screenshots,
		"Total":        total,
		"Page":         page,
		"PerPage":      perPage,
		"TotalPages":   totalPages,
		"PlayerIP":     playerIP,
		"PlayerName":   playerName,
		"ServerAddr":   serverAddr,
		"Servers":      distinctServers,
		"DateFrom":     dateFrom,
		"DateTo":       dateTo,
		"Unreviewed":   unreviewed,
		"CurrentPage":  "screenshots",
		"Msg":          r.URL.Query().Get("msg"),
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

func (ws *WebServer) handleBulkReviewScreenshots(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/screenshots", http.StatusFound)
		return
	}
	idsRaw := r.FormValue("ids")
	var ids []int64
	for _, idStr := range strings.Split(idsRaw, ",") {
		idStr = strings.TrimSpace(idStr)
		if id, err := strconv.ParseInt(idStr, 10, 64); err == nil && id > 0 {
			ids = append(ids, id)
		}
	}
	if len(ids) > 0 {
		ws.db.MarkAllReviewed(ids)
	}
	http.Redirect(w, r, "/screenshots?msg=bulk_reviewed", http.StatusFound)
}

func (ws *WebServer) handleExportScreenshotsCSV(w http.ResponseWriter, r *http.Request) {
	playerIP := strings.TrimSpace(r.URL.Query().Get("player"))
	playerName := strings.TrimSpace(r.URL.Query().Get("name"))
	serverAddr := strings.TrimSpace(r.URL.Query().Get("server"))
	dateFrom := strings.TrimSpace(r.URL.Query().Get("from"))
	dateTo := strings.TrimSpace(r.URL.Query().Get("to"))
	unreviewed := r.URL.Query().Get("unreviewed") == "1"

	screenshots, _, err := ws.db.GetScreenshots(playerIP, playerName, serverAddr, dateFrom, dateTo, unreviewed, 1, 5000)
	if err != nil {
		http.Error(w, "Error retrieving screenshots", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"screenshots.csv\"")

	fmt.Fprintf(w, "ID,Fecha,Servidor,Jugador,IP,Revisado,Notas\n")
	for _, s := range screenshots {
		fmt.Fprintf(w, "%d,\"%s\",\"%s\",\"%s\",\"%s\",%t,\"%s\"\n",
			s.ID,
			s.Timestamp.Format("2006-01-02 15:04:05"),
			s.ServerAddr,
			strings.ReplaceAll(s.PlayerName, "\"", "\"\""),
			s.PlayerIP,
			s.Reviewed,
			strings.ReplaceAll(s.Notes, "\"", "\"\""))
	}
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
	playerIP := strings.TrimSpace(r.URL.Query().Get("player"))
	playerName := strings.TrimSpace(r.URL.Query().Get("name"))
	serverAddr := strings.TrimSpace(r.URL.Query().Get("server"))
	vType := strings.TrimSpace(r.URL.Query().Get("type"))
	dateFrom := strings.TrimSpace(r.URL.Query().Get("from"))
	dateTo := strings.TrimSpace(r.URL.Query().Get("to"))
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if perPage != 20 && perPage != 50 && perPage != 100 {
		perPage = 50
	}

	violations, total, err := ws.db.GetViolations(playerIP, playerName, serverAddr, vType, dateFrom, dateTo, page, perPage)
	if err != nil {
		log.Printf("[WEB] Error getting violations: %v", err)
	}
	totalPages := (total + perPage - 1) / perPage
	distinctServers, _ := ws.db.GetDistinctServers()

	data := map[string]interface{}{
		"Violations":  violations,
		"Total":       total,
		"Page":        page,
		"PerPage":     perPage,
		"TotalPages":  totalPages,
		"PlayerIP":    playerIP,
		"PlayerName":  playerName,
		"ServerAddr":  serverAddr,
		"Servers":     distinctServers,
		"Type":        vType,
		"DateFrom":    dateFrom,
		"DateTo":      dateTo,
		"CurrentPage": "violations",
	}
	ws.templates.Execute(w, "violations", data)
}

func (ws *WebServer) handleExportViolationsCSV(w http.ResponseWriter, r *http.Request) {
	playerIP := strings.TrimSpace(r.URL.Query().Get("player"))
	playerName := strings.TrimSpace(r.URL.Query().Get("name"))
	serverAddr := strings.TrimSpace(r.URL.Query().Get("server"))
	vType := strings.TrimSpace(r.URL.Query().Get("type"))
	dateFrom := strings.TrimSpace(r.URL.Query().Get("from"))
	dateTo := strings.TrimSpace(r.URL.Query().Get("to"))

	violations, _, err := ws.db.GetViolations(playerIP, playerName, serverAddr, vType, dateFrom, dateTo, 1, 5000)
	if err != nil {
		http.Error(w, "Error retrieving violations", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"violations.csv\"")

	fmt.Fprintf(w, "ID,Fecha,Servidor,Jugador,IP,Tipo,Razon,Detalles\n")
	for _, v := range violations {
		fmt.Fprintf(w, "%d,\"%s\",\"%s\",\"%s\",\"%s\",\"%s\",\"%s\",\"%s\"\n",
			v.ID,
			v.Timestamp.Format("2006-01-02 15:04:05"),
			v.ServerAddr,
			strings.ReplaceAll(v.PlayerName, "\"", "\"\""),
			v.PlayerIP,
			v.Type,
			strings.ReplaceAll(v.Reason, "\"", "\"\""),
			strings.ReplaceAll(v.Details, "\"", "\"\""))
	}
}

func (ws *WebServer) handleProcessSnapshots(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	playerIP := strings.TrimSpace(r.URL.Query().Get("player"))
	playerName := strings.TrimSpace(r.URL.Query().Get("name"))
	serverAddr := strings.TrimSpace(r.URL.Query().Get("server"))
	dateFrom := strings.TrimSpace(r.URL.Query().Get("from"))
	dateTo := strings.TrimSpace(r.URL.Query().Get("to"))
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if perPage != 20 && perPage != 50 && perPage != 100 {
		perPage = 20
	}

	snapshots, total, err := ws.db.GetProcessSnapshots(playerIP, playerName, serverAddr, dateFrom, dateTo, page, perPage)
	if err != nil {
		log.Printf("[WEB] Error getting process snapshots: %v", err)
	}
	totalPages := (total + perPage - 1) / perPage
	distinctServers, _ := ws.db.GetDistinctServers()

	data := map[string]interface{}{
		"Snapshots":   snapshots,
		"Total":       total,
		"TotalPages":  totalPages,
		"HasNext":     page < totalPages,
		"CurrentPage": "process-snapshots",
		"PlayerIP":    playerIP,
		"PlayerName":  playerName,
		"ServerAddr":  serverAddr,
		"Servers":     distinctServers,
		"DateFrom":    dateFrom,
		"DateTo":      dateTo,
		"Page":        page,
		"PerPage":     perPage,
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
		if matched, pattern, _ := bl.CheckModuleFull(modules[i].Name, modules[i].Path, modules[i].SHA1); matched {
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
		redirectURL := "/blacklist"

		if action == "add" || (action == "" && r.FormValue("pattern") != "") {
			entryType := strings.TrimSpace(r.FormValue("type"))
			pattern := strings.TrimSpace(r.FormValue("pattern"))
			addedBy := strings.TrimSpace(r.FormValue("added_by"))
			if addedBy == "" {
				addedBy = "admin"
			}
			if entryType == "" {
				entryType = "process"
			}
			if pattern != "" {
				if err := ws.handler.Blacklist().AddEntry(entryType, pattern, addedBy); err != nil {
					log.Printf("[WEB] Error adding blacklist entry: %v", err)
					redirectURL = "/blacklist?error=add_failed"
				} else {
					redirectURL = "/blacklist?msg=added"
				}
			} else {
				redirectURL = "/blacklist?error=empty_pattern"
			}
		} else if action == "delete" {
			idStr := r.FormValue("id")
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err == nil && id > 0 {
				if err := ws.handler.Blacklist().RemoveEntry(id); err != nil {
					log.Printf("[WEB] Error removing blacklist entry: %v", err)
					redirectURL = "/blacklist?error=delete_failed"
				} else {
					redirectURL = "/blacklist?msg=deleted"
				}
			} else {
				redirectURL = "/blacklist?error=invalid_id"
			}
		} else if action == "toggle" {
			idStr := r.FormValue("id")
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err == nil && id > 0 {
				if err := ws.handler.Blacklist().ToggleEntry(id); err != nil {
					log.Printf("[WEB] Error toggling blacklist entry: %v", err)
					redirectURL = "/blacklist?error=toggle_failed"
				} else {
					redirectURL = "/blacklist?msg=toggled"
				}
			} else {
				redirectURL = "/blacklist?error=invalid_id"
			}
		}

		http.Redirect(w, r, redirectURL, http.StatusFound)
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
		"Msg":          r.URL.Query().Get("msg"),
		"Error":        r.URL.Query().Get("error"),
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

func (ws *WebServer) handleSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	cookie, _ := r.Cookie("session")
	var adminUser string
	if cookie != nil {
		sess, err := ws.db.GetSession(cookie.Value)
		if err == nil {
			admin, err := ws.db.GetAdminByID(sess.AdminID)
			if err == nil {
				adminUser = admin.Username
			}
		}
	}
	stats, _ := ws.db.GetStats()

	data := map[string]interface{}{
		"AdminUser":   adminUser,
		"Stats":       stats,
		"CurrentPage": "settings",
		"Msg":         r.URL.Query().Get("msg"),
		"Error":       r.URL.Query().Get("error"),
		"Count":       r.URL.Query().Get("count"),
	}
	ws.templates.Execute(w, "settings", data)
}

func (ws *WebServer) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/settings", http.StatusFound)
		return
	}

	cookie, _ := r.Cookie("session")
	if cookie == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	sess, err := ws.db.GetSession(cookie.Value)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	admin, err := ws.db.GetAdminByID(sess.AdminID)
	if err != nil {
		http.Redirect(w, r, "/settings?error=admin_not_found", http.StatusFound)
		return
	}

	oldPass := r.FormValue("old_password")
	newPass := r.FormValue("new_password")
	confirmPass := r.FormValue("confirm_password")

	if !CheckPassword(admin.Password, oldPass) {
		http.Redirect(w, r, "/settings?error=wrong_old_password", http.StatusFound)
		return
	}

	if len(newPass) < 4 {
		http.Redirect(w, r, "/settings?error=password_too_short", http.StatusFound)
		return
	}

	if newPass != confirmPass {
		http.Redirect(w, r, "/settings?error=passwords_do_not_match", http.StatusFound)
		return
	}

	newHash := HashPassword(newPass)
	if err := ws.db.ChangeAdminPassword(admin.ID, newHash); err != nil {
		log.Printf("[WEB] Error changing password: %v", err)
		http.Redirect(w, r, "/settings?error=db_error", http.StatusFound)
		return
	}

	http.Redirect(w, r, "/settings?msg=password_updated", http.StatusFound)
}

func (ws *WebServer) handlePurgeScreenshots(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/settings", http.StatusFound)
		return
	}

	days, _ := strconv.Atoi(r.FormValue("days"))
	if days <= 0 {
		days = 30
	}

	count, err := ws.db.DeleteOldScreenshots(days)
	if err != nil {
		log.Printf("[WEB] Error purging screenshots: %v", err)
		http.Redirect(w, r, "/settings?error=purge_failed", http.StatusFound)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/settings?msg=purged&count=%d", count), http.StatusFound)
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

// SendViolationPush triggers a Web Push notification for a player violation
func (ws *WebServer) SendViolationPush(serverAddr, playerIP, playerName, vType, reason string) {
	if ws.pushMgr == nil {
		return
	}

	title := "⚠️ Violación Anticheat"
	if playerName != "" {
		title = fmt.Sprintf("⚠️ Violación: %s", playerName)
	}

	body := fmt.Sprintf("Tipo: %s | Razón: %s | IP: %s", vType, reason, playerIP)
	if serverAddr != "" {
		body += fmt.Sprintf(" (%s)", serverAddr)
	}

	url := "/violations"
	if playerName != "" {
		url = fmt.Sprintf("/violations?name=%s", playerName)
	}

	ws.pushMgr.SendToAll(PushNotificationPayload{
		Title: title,
		Body:  body,
		Icon:  "/static/icon.svg",
		URL:   url,
		Tag:   "violation-" + vType,
	})
}

func (ws *WebServer) handleAPIPushVapidKey(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if ws.pushMgr == nil {
		http.Error(w, `{"error":"push manager not initialized"}`, http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{
		"publicKey": ws.pushMgr.PublicKey(),
	})
}

func (ws *WebServer) handleAPIPushSubscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Endpoint string `json:"endpoint"`
		Keys     struct {
			P256dh string `json:"p256dh"`
			Auth   string `json:"auth"`
		} `json:"keys"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	if req.Endpoint == "" || req.Keys.P256dh == "" || req.Keys.Auth == "" {
		http.Error(w, `{"error":"missing endpoint or keys"}`, http.StatusBadRequest)
		return
	}

	userAgent := r.UserAgent()
	if err := ws.db.InsertPushSubscription(req.Endpoint, req.Keys.P256dh, req.Keys.Auth, userAgent); err != nil {
		log.Printf("[WEB] Error inserting push subscription: %v", err)
		http.Error(w, `{"error":"db error"}`, http.StatusInternalServerError)
		return
	}

	log.Printf("[WEB] New Web Push subscription registered (%s)", userAgent)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"ok":true}`)
}

func (ws *WebServer) handleAPIPushUnsubscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Endpoint string `json:"endpoint"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	if err := ws.db.DeletePushSubscription(req.Endpoint); err != nil {
		log.Printf("[WEB] Error deleting push subscription: %v", err)
		http.Error(w, `{"error":"db error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"ok":true}`)
}

func (ws *WebServer) handleAPIPushTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if ws.pushMgr == nil {
		http.Error(w, `{"error":"push manager not initialized"}`, http.StatusInternalServerError)
		return
	}

	ws.pushMgr.SendToAll(PushNotificationPayload{
		Title: "🔔 Notificación de Prueba",
		Body:  "Las notificaciones Push de Anticheat Q2PRO están funcionando correctamente.",
		Icon:  "/static/icon.svg",
		URL:   "/violations",
		Tag:   "test-notification",
	})

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"ok":true}`)
}

