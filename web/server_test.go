package web

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/user/anticheat_cl/database"
	"github.com/user/anticheat_cl/screenshots"
	"github.com/user/anticheat_cl/server"
)

func setupTestWeb(t *testing.T) (*WebServer, *database.DB, *server.Handler, func()) {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}

	storage, err := screenshots.NewStorage(filepath.Join(dir, "screenshots"), db)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	srv := server.New("127.0.0.1:0", storage)
	handler := srv.GetHandler()
	handler.Blacklist().SetDB(db)

	ws := New("127.0.0.1:0", db, handler)

	cleanup := func() {
		db.Close()
		os.RemoveAll(dir)
	}
	return ws, db, handler, cleanup
}

func TestWebBlacklistActions(t *testing.T) {
	ws, db, handler, cleanup := setupTestWeb(t)
	defer cleanup()

	// 1. Add pattern via web POST
	form := url.Values{}
	form.Set("action", "add")
	form.Set("type", "process")
	form.Set("pattern", "badtool.exe")
	form.Set("added_by", "tester")

	req := httptest.NewRequest("POST", "/blacklist", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	ws.handleBlacklist(w, req)
	if w.Code != http.StatusFound {
		t.Fatalf("expected redirect StatusFound (302), got %d", w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "/blacklist?msg=added" {
		t.Fatalf("expected location /blacklist?msg=added, got %s", loc)
	}

	// Verify added in memory and db
	matched, _ := handler.Blacklist().CheckProcess("badtool.exe")
	if !matched {
		t.Fatalf("expected badtool.exe to be blacklisted")
	}

	entries, _ := db.GetBlacklist()
	var badToolID int64
	for _, e := range entries {
		if e.Pattern == "badtool.exe" {
			badToolID = e.ID
			break
		}
	}
	if badToolID == 0 {
		t.Fatalf("expected badtool.exe in database")
	}

	// 2. Toggle pattern via web POST
	form = url.Values{}
	form.Set("action", "toggle")
	form.Set("id", strconv.FormatInt(badToolID, 10))

	req = httptest.NewRequest("POST", "/blacklist", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()

	ws.handleBlacklist(w, req)
	if w.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d", w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "/blacklist?msg=toggled" {
		t.Fatalf("expected location /blacklist?msg=toggled, got %s", loc)
	}

	matchedAfterToggle, _ := handler.Blacklist().CheckProcess("badtool.exe")
	if matchedAfterToggle {
		t.Fatalf("expected badtool.exe to be inactive after toggle")
	}

	// 3. Delete pattern via web POST
	form = url.Values{}
	form.Set("action", "delete")
	form.Set("id", strconv.FormatInt(badToolID, 10))

	req = httptest.NewRequest("POST", "/blacklist", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()

	ws.handleBlacklist(w, req)
	if w.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d", w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "/blacklist?msg=deleted" {
		t.Fatalf("expected location /blacklist?msg=deleted, got %s", loc)
	}

	// Verify deleted from db
	entries, _ = db.GetBlacklist()
	for _, e := range entries {
		if e.ID == badToolID {
			t.Fatalf("expected badToolID %d to be deleted from db", badToolID)
		}
	}

	// 4. Render GET /blacklist
	getReq := httptest.NewRequest("GET", "/blacklist?msg=deleted", nil)
	getW := httptest.NewRecorder()
	ws.handleBlacklist(getW, getReq)

	if getW.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for GET /blacklist, got %d", getW.Code)
	}
	body := getW.Body.String()
	if !strings.Contains(body, "Patrón eliminado correctamente de la blacklist") {
		t.Fatalf("expected success alert in response body")
	}
	if !strings.Contains(body, "name=\"action\" value=\"delete\"") {
		t.Fatalf("expected delete action form in blacklist table")
	}
}

func TestWebPlayerFiltering(t *testing.T) {
	ws, _, _, cleanup := setupTestWeb(t)
	defer cleanup()

	// 1. Screenshots filter
	req := httptest.NewRequest("GET", "/screenshots?name=Sniper", nil)
	w := httptest.NewRecorder()
	ws.handleScreenshots(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "name=\"name\" value=\"Sniper\"") {
		t.Fatalf("expected name input with 'Sniper' value in HTML")
	}

	// 2. Violations filter
	req = httptest.NewRequest("GET", "/violations?name=Killer", nil)
	w = httptest.NewRecorder()
	ws.handleViolations(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}
	body = w.Body.String()
	if !strings.Contains(body, "name=\"name\" value=\"Killer\"") {
		t.Fatalf("expected name input with 'Killer' value in HTML")
	}

	// 3. Process Snapshots filter
	req = httptest.NewRequest("GET", "/process-snapshots?name=ProGamer", nil)
	w = httptest.NewRecorder()
	ws.handleProcessSnapshots(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}
	body = w.Body.String()
	if !strings.Contains(body, "name=\"name\" value=\"ProGamer\"") {
		t.Fatalf("expected name input with 'ProGamer' value in HTML")
	}
}
