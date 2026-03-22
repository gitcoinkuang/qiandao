package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
	_ "time/tzdata"
)

var appLocation = mustLoadLocation(getenv("APP_TIMEZONE", "Asia/Shanghai"))

type Task struct {
	ID              int               `json:"id"`
	Name            string            `json:"name"`
	URL             string            `json:"url"`
	Method          string            `json:"method"`
	Headers         map[string]string `json:"headers"`
	Body            string            `json:"body"`
	CurlCommand     string            `json:"curl_command"`
	Enabled         bool              `json:"enabled"`
	ScheduleEnabled bool              `json:"schedule_enabled"`
	ScheduleHour    int               `json:"schedule_hour"`
	ScheduleMinute  int               `json:"schedule_minute"`
	TimeoutSeconds  int               `json:"timeout_seconds"`
	RetryCount      int               `json:"retry_count"`
	SuccessKeywords string            `json:"success_keywords"`
	FailureKeywords string            `json:"failure_keywords"`
	LastStatus      string            `json:"last_status"`
	LastRunAt       string            `json:"last_run_at"`
	LastDurationMS  int               `json:"last_duration_ms"`
	CreatedAt       string            `json:"created_at"`
	UpdatedAt       string            `json:"updated_at"`
}

type HistoryItem struct {
	ID              int    `json:"id"`
	TaskID          int    `json:"task_id"`
	TaskName        string `json:"task_name"`
	Status          string `json:"status"`
	StatusCode      int    `json:"status_code"`
	Message         string `json:"message"`
	ResponsePreview string `json:"response_preview"`
	ResponseTimeMS  int    `json:"response_time_ms"`
	TriggeredBy     string `json:"triggered_by"`
	CreatedAt       string `json:"created_at"`
}

type NotifySettings struct {
	TelegramEnabled  bool   `json:"telegram_enabled"`
	TelegramBotToken string `json:"telegram_bot_token"`
	TelegramChatID   string `json:"telegram_chat_id"`
	WebhookEnabled   bool   `json:"webhook_enabled"`
	WebhookURL       string `json:"webhook_url"`
	NotifyOnSuccess  bool   `json:"notify_on_success"`
	NotifyOnFailure  bool   `json:"notify_on_failure"`
}

type ScheduleSettings struct {
	Enabled    bool `json:"enabled"`
	Hour       int  `json:"hour"`
	Minute     int  `json:"minute"`
	MaxWorkers int  `json:"max_workers"`
}

type SecuritySettings struct {
	Enabled      bool   `json:"enabled"`
	PasswordHash string `json:"password_hash"`
}

type Settings struct {
	Notify   NotifySettings   `json:"notify"`
	Schedule ScheduleSettings `json:"schedule"`
	Security SecuritySettings `json:"security"`
}

type AppState struct {
	NextTaskID    int           `json:"next_task_id"`
	NextHistoryID int           `json:"next_history_id"`
	Tasks         []Task        `json:"tasks"`
	History       []HistoryItem `json:"history"`
	Settings      Settings      `json:"settings"`
}

type Server struct {
	mu          sync.RWMutex
	state       AppState
	statePath   string
	templates   *template.Template
	httpClient  *http.Client
	sessionMu   sync.Mutex
	sessions    map[string]time.Time
	scheduleMu  sync.Mutex
	scheduleRan map[string]struct{}
}

type APIResponse struct {
	Success bool        `json:"success"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Task    interface{} `json:"task,omitempty"`
	Tasks   interface{} `json:"tasks,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Results interface{} `json:"results,omitempty"`
	Config  interface{} `json:"config,omitempty"`
	History interface{} `json:"history,omitempty"`
}

func defaultState() AppState {
	return AppState{
		NextTaskID:    1,
		NextHistoryID: 1,
		Tasks:         []Task{},
		History:       []HistoryItem{},
		Settings: Settings{
			Notify: NotifySettings{
				NotifyOnFailure: true,
			},
			Schedule: ScheduleSettings{
				Hour:       8,
				Minute:     0,
				MaxWorkers: 4,
			},
			Security: SecuritySettings{
				Enabled:      true,
				PasswordHash: hashPassword(defaultAdminPassword()),
			},
		},
	}
}

func nowString() string {
	return appNow().Format("2006-01-02 15:04:05")
}

func appNow() time.Time {
	return time.Now().In(appLocation)
}

func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(strings.TrimSpace(name))
	if err != nil {
		log.Printf("invalid APP_TIMEZONE %q, fallback to Asia/Shanghai", name)
		loc, _ = time.LoadLocation("Asia/Shanghai")
		return loc
	}
	return loc
}

func main() {
	server, err := newServer()
	if err != nil {
		log.Fatal(err)
	}

	go server.schedulerLoop()

	addr := getenv("APP_ADDR", ":8080")
	log.Printf("qiandao v2 listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, server.routes()))
}

func newServer() (*Server, error) {
	tmpl, err := template.ParseGlob(filepath.Join("templates", "*.html"))
	if err != nil {
		return nil, err
	}

	statePath := filepath.Join("data", "state.json")
	s := &Server{
		statePath:   statePath,
		templates:   tmpl,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		sessions:    map[string]time.Time{},
		scheduleRan: map[string]struct{}{},
	}

	if err := os.MkdirAll(filepath.Dir(statePath), 0o755); err != nil {
		return nil, err
	}
	if err := s.loadState(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Server) loadState() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.statePath)
	if errors.Is(err, os.ErrNotExist) {
		s.state = defaultState()
		return s.saveStateLocked()
	}
	if err != nil {
		return err
	}

	state := defaultState()
	if err := json.Unmarshal(data, &state); err != nil {
		return err
	}
	if state.NextTaskID < 1 {
		state.NextTaskID = 1
	}
	if state.NextHistoryID < 1 {
		state.NextHistoryID = 1
	}
	if state.Settings.Security.PasswordHash == "" {
		state.Settings.Security.PasswordHash = hashPassword(defaultAdminPassword())
		state.Settings.Security.Enabled = true
	}
	s.state = state
	return nil
}

func (s *Server) saveStateLocked() error {
	payload, err := json.MarshalIndent(s.state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.statePath, payload, 0o644)
}

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/login", s.handleLogin)
	mux.HandleFunc("/logout", s.handleLogout)
	mux.HandleFunc("/api/bootstrap", s.requireAPIAuth(s.handleBootstrap))
	mux.HandleFunc("/api/tasks/parse", s.requireAPIAuth(s.handleParseTask))
	mux.HandleFunc("/api/tasks/run-all", s.requireAPIAuth(s.handleRunAllTasks))
	mux.HandleFunc("/api/tasks/", s.requireAPIAuth(s.handleTaskByID))
	mux.HandleFunc("/api/tasks", s.requireAPIAuth(s.handleTasks))
	mux.HandleFunc("/api/history", s.requireAPIAuth(s.handleHistory))
	mux.HandleFunc("/api/settings/notify", s.requireAPIAuth(s.handleNotifySettings))
	mux.HandleFunc("/api/settings/notify/test", s.requireAPIAuth(s.handleNotifyTest))
	mux.HandleFunc("/api/settings/schedule", s.requireAPIAuth(s.handleScheduleSettings))
	mux.HandleFunc("/api/settings/schedule/check", s.requireAPIAuth(s.handleScheduleCheck))
	mux.HandleFunc("/api/settings/security", s.handleSecuritySettings)
	return mux
}

func getenv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func defaultAdminPassword() string {
	return getenv("QIANGDAO_DEFAULT_PASSWORD", "admin123456")
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if s.securityEnabled() && !s.isAuthenticated(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	s.render(w, "index.html", map[string]any{"Title": "QianDao V2"})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if !s.securityEnabled() {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	if r.Method == http.MethodGet {
		s.render(w, "login.html", map[string]any{"Error": ""})
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	password := strings.TrimSpace(r.FormValue("password"))
	if !s.checkPassword(password) {
		s.render(w, "login.html", map[string]any{"Error": "密码错误"})
		return
	}
	token := randomToken(32)
	s.sessionMu.Lock()
	s.sessions[token] = appNow().Add(7 * 24 * time.Hour)
	s.sessionMu.Unlock()
	http.SetCookie(w, &http.Cookie{
		Name:     "qiandao_session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("qiandao_session"); err == nil {
		s.sessionMu.Lock()
		delete(s.sessions, cookie.Value)
		s.sessionMu.Unlock()
	}
	http.SetCookie(w, &http.Cookie{Name: "qiandao_session", Value: "", Path: "/", MaxAge: -1, HttpOnly: true})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (s *Server) requireAPIAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.securityEnabled() && !s.isAuthenticated(r) {
			writeJSON(w, http.StatusUnauthorized, APIResponse{Success: false, Error: "unauthorized"})
			return
		}
		next(w, r)
	}
}

func (s *Server) securityEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Settings.Security.Enabled
}

func (s *Server) isAuthenticated(r *http.Request) bool {
	cookie, err := r.Cookie("qiandao_session")
	if err != nil || cookie.Value == "" {
		return false
	}
	s.sessionMu.Lock()
	defer s.sessionMu.Unlock()
	expiresAt, ok := s.sessions[cookie.Value]
	if !ok || appNow().After(expiresAt) {
		delete(s.sessions, cookie.Value)
		return false
	}
	return true
}

func (s *Server) checkPassword(password string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.state.Settings.Security.Enabled {
		return true
	}
	return s.state.Settings.Security.PasswordHash == hashPassword(password)
}

func randomToken(length int) string {
	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%d", appNow().UnixNano())
	}
	return hex.EncodeToString(buf)
}

func hashPassword(password string) string {
	sum := sha256.Sum256([]byte(password))
	return hex.EncodeToString(sum[:])
}

func writeJSON(w http.ResponseWriter, status int, payload APIResponse) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func decodeJSON(r *http.Request, target any) error {
	defer r.Body.Close()
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return err
	}
	if len(strings.TrimSpace(string(body))) == 0 {
		return nil
	}
	return json.Unmarshal(body, target)
}

func (s *Server) render(w http.ResponseWriter, name string, data any) {
	if err := s.templates.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleBootstrap(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{Success: false, Error: "method not allowed"})
		return
	}
	writeJSON(w, http.StatusOK, APIResponse{Success: true, Data: s.bootstrapPayload()})
}

func (s *Server) bootstrapPayload() map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()

	successCount := 0
	failedCount := 0
	durationTotal := 0
	durationCount := 0
	history := slices.Clone(s.state.History)
	tasks := slices.Clone(s.state.Tasks)
	for _, item := range history {
		if item.Status == "success" {
			successCount++
		}
		if item.Status == "failed" {
			failedCount++
		}
		if item.ResponseTimeMS > 0 {
			durationTotal += item.ResponseTimeMS
			durationCount++
		}
	}
	avgDuration := 0
	if durationCount > 0 {
		avgDuration = durationTotal / durationCount
	}

	return map[string]any{
		"stats": map[string]any{
			"total_tasks":     len(tasks),
			"enabled_tasks":   countEnabledTasks(tasks),
			"recent_success":  successCount,
			"recent_failed":   failedCount,
			"avg_duration_ms": avgDuration,
		},
		"tasks":           tasks,
		"history":         history,
		"notify_config":   s.state.Settings.Notify,
		"schedule_config": s.state.Settings.Schedule,
		"security_config": map[string]any{"enabled": s.state.Settings.Security.Enabled},
	}
}

func countEnabledTasks(tasks []Task) int {
	count := 0
	for _, task := range tasks {
		if task.Enabled {
			count++
		}
	}
	return count
}

func (s *Server) handleTasks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.mu.RLock()
		tasks := slices.Clone(s.state.Tasks)
		s.mu.RUnlock()
		writeJSON(w, http.StatusOK, APIResponse{Success: true, Tasks: tasks})
	case http.MethodPost:
		var payload Task
		if err := decodeJSON(r, &payload); err != nil {
			writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: err.Error()})
			return
		}
		task, err := s.createTask(payload)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, APIResponse{Success: true, Task: task})
	default:
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{Success: false, Error: "method not allowed"})
	}
}

func (s *Server) handleTaskByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	if strings.HasSuffix(path, "/run") {
		idText := strings.TrimSuffix(path, "/run")
		id, err := strconv.Atoi(strings.Trim(idText, "/"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: "invalid task id"})
			return
		}
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, APIResponse{Success: false, Error: "method not allowed"})
			return
		}
		result, err := s.runTaskByID(id, "manual")
		if err != nil {
			writeJSON(w, http.StatusNotFound, APIResponse{Success: false, Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, APIResponse{Success: true, Result: result})
		return
	}

	id, err := strconv.Atoi(strings.Trim(path, "/"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: "invalid task id"})
		return
	}

	switch r.Method {
	case http.MethodPut:
		var payload Task
		if err := decodeJSON(r, &payload); err != nil {
			writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: err.Error()})
			return
		}
		task, err := s.updateTask(id, payload)
		if err != nil {
			status := http.StatusBadRequest
			if err.Error() == "task not found" {
				status = http.StatusNotFound
			}
			writeJSON(w, status, APIResponse{Success: false, Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, APIResponse{Success: true, Task: task})
	case http.MethodDelete:
		if err := s.deleteTask(id); err != nil {
			writeJSON(w, http.StatusNotFound, APIResponse{Success: false, Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, APIResponse{Success: true})
	default:
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{Success: false, Error: "method not allowed"})
	}
}

func (s *Server) handleRunAllTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{Success: false, Error: "method not allowed"})
		return
	}
	results := s.runAllEnabledTasks("manual-batch")
	writeJSON(w, http.StatusOK, APIResponse{Success: true, Results: results})
}

func (s *Server) handleHistory(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.mu.RLock()
		history := slices.Clone(s.state.History)
		s.mu.RUnlock()
		writeJSON(w, http.StatusOK, APIResponse{Success: true, History: history})
	case http.MethodDelete:
		s.mu.Lock()
		s.state.History = []HistoryItem{}
		_ = s.saveStateLocked()
		s.mu.Unlock()
		writeJSON(w, http.StatusOK, APIResponse{Success: true})
	default:
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{Success: false, Error: "method not allowed"})
	}
}

func (s *Server) handleParseTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{Success: false, Error: "method not allowed"})
		return
	}
	var payload Task
	if err := decodeJSON(r, &payload); err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: err.Error()})
		return
	}
	parsed, err := parseCurl(payload.CurlCommand)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: err.Error()})
		return
	}
	if payload.Name != "" {
		parsed.Name = payload.Name
	}
	if payload.Method != "" {
		parsed.Method = strings.ToUpper(payload.Method)
	}
	writeJSON(w, http.StatusOK, APIResponse{Success: true, Config: parsed})
}

func (s *Server) handleNotifySettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{Success: false, Error: "method not allowed"})
		return
	}
	var config NotifySettings
	if err := decodeJSON(r, &config); err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: err.Error()})
		return
	}
	s.mu.Lock()
	s.state.Settings.Notify = config
	err := s.saveStateLocked()
	s.mu.Unlock()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, APIResponse{Success: false, Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, APIResponse{Success: true, Config: config})
}

func (s *Server) handleNotifyTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{Success: false, Error: "method not allowed"})
		return
	}
	result := HistoryItem{
		TaskName:        "Notification test",
		Status:          "success",
		StatusCode:      200,
		Message:         "This is a test notification from QianDao V2.",
		ResponsePreview: "Notification pipeline is working.",
		ResponseTimeMS:  42,
		TriggeredBy:     "manual",
		CreatedAt:       nowString(),
	}
	s.sendNotifications(result)
	writeJSON(w, http.StatusOK, APIResponse{Success: true})
}

func (s *Server) handleScheduleSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{Success: false, Error: "method not allowed"})
		return
	}
	var config ScheduleSettings
	if err := decodeJSON(r, &config); err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: err.Error()})
		return
	}
	if config.Hour < 0 || config.Hour > 23 || config.Minute < 0 || config.Minute > 59 {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: "invalid schedule time"})
		return
	}
	if config.MaxWorkers < 1 {
		config.MaxWorkers = 1
	}
	if config.MaxWorkers > 8 {
		config.MaxWorkers = 8
	}
	s.mu.Lock()
	s.state.Settings.Schedule = config
	err := s.saveStateLocked()
	s.mu.Unlock()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, APIResponse{Success: false, Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, APIResponse{Success: true, Config: config})
}

func (s *Server) handleScheduleCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{Success: false, Error: "method not allowed"})
		return
	}
	results := s.runScheduledTasks(appNow(), "manual-schedule-check")
	writeJSON(w, http.StatusOK, APIResponse{Success: true, Results: results})
}

func (s *Server) handleSecuritySettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{Success: false, Error: "method not allowed"})
		return
	}
	if s.securityEnabled() && !s.isAuthenticated(r) {
		writeJSON(w, http.StatusUnauthorized, APIResponse{Success: false, Error: "unauthorized"})
		return
	}
	var payload struct {
		Enabled  bool   `json:"enabled"`
		Password string `json:"password"`
	}
	if err := decodeJSON(r, &payload); err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: err.Error()})
		return
	}
	if payload.Password != "" && len(payload.Password) < 6 {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: "密码长度至少为 6 位"})
		return
	}

	s.mu.Lock()
	if payload.Password != "" {
		s.state.Settings.Security.PasswordHash = hashPassword(payload.Password)
	}
	s.state.Settings.Security.Enabled = payload.Enabled
	err := s.saveStateLocked()
	s.mu.Unlock()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, APIResponse{Success: false, Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Config:  map[string]any{"enabled": payload.Enabled},
	})
}

func (s *Server) createTask(input Task) (Task, error) {
	task, err := normalizeTask(input)
	if err != nil {
		return Task{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	task.ID = s.state.NextTaskID
	s.state.NextTaskID++
	task.CreatedAt = nowString()
	task.UpdatedAt = task.CreatedAt
	task.LastStatus = "idle"
	s.state.Tasks = append([]Task{task}, s.state.Tasks...)
	return task, s.saveStateLocked()
}

func (s *Server) updateTask(id int, input Task) (Task, error) {
	task, err := normalizeTask(input)
	if err != nil {
		return Task{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.state.Tasks {
		if s.state.Tasks[i].ID == id {
			task.ID = id
			task.CreatedAt = s.state.Tasks[i].CreatedAt
			task.LastStatus = s.state.Tasks[i].LastStatus
			task.LastRunAt = s.state.Tasks[i].LastRunAt
			task.LastDurationMS = s.state.Tasks[i].LastDurationMS
			task.UpdatedAt = nowString()
			s.state.Tasks[i] = task
			return task, s.saveStateLocked()
		}
	}
	return Task{}, errors.New("task not found")
}

func (s *Server) deleteTask(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.state.Tasks {
		if s.state.Tasks[i].ID == id {
			s.state.Tasks = append(s.state.Tasks[:i], s.state.Tasks[i+1:]...)
			filtered := s.state.History[:0]
			for _, item := range s.state.History {
				if item.TaskID != id {
					filtered = append(filtered, item)
				}
			}
			s.state.History = filtered
			return s.saveStateLocked()
		}
	}
	return errors.New("task not found")
}

func normalizeTask(input Task) (Task, error) {
	task := input
	if strings.TrimSpace(task.CurlCommand) != "" {
		parsed, err := parseCurl(task.CurlCommand)
		if err != nil {
			return Task{}, err
		}
		if task.URL == "" {
			task.URL = parsed.URL
		}
		if len(task.Headers) == 0 {
			task.Headers = parsed.Headers
		}
		if task.Body == "" {
			task.Body = parsed.Body
		}
		if task.Method == "" {
			task.Method = parsed.Method
		}
	}

	task.Name = strings.TrimSpace(task.Name)
	task.URL = strings.TrimSpace(task.URL)
	task.Method = strings.ToUpper(strings.TrimSpace(task.Method))
	if task.Name == "" {
		return Task{}, errors.New("task name is required")
	}
	if !strings.HasPrefix(task.URL, "http://") && !strings.HasPrefix(task.URL, "https://") {
		return Task{}, errors.New("task url must start with http:// or https://")
	}
	if task.Method == "" {
		task.Method = http.MethodGet
	}
	if !slices.Contains([]string{"GET", "POST", "PUT", "PATCH", "DELETE"}, task.Method) {
		return Task{}, errors.New("unsupported method")
	}
	if task.Headers == nil {
		task.Headers = map[string]string{}
	}
	if task.TimeoutSeconds < 1 {
		task.TimeoutSeconds = 30
	}
	if task.TimeoutSeconds > 120 {
		task.TimeoutSeconds = 120
	}
	if task.RetryCount < 0 {
		task.RetryCount = 0
	}
	if task.RetryCount > 5 {
		task.RetryCount = 5
	}
	if task.ScheduleHour < 0 || task.ScheduleHour > 23 {
		task.ScheduleHour = 8
	}
	if task.ScheduleMinute < 0 || task.ScheduleMinute > 59 {
		task.ScheduleMinute = 0
	}
	if task.CurlCommand == "" {
		task.CurlCommand = buildCurl(task)
	}
	return task, nil
}

func buildCurl(task Task) string {
	parts := []string{fmt.Sprintf("curl '%s'", task.URL)}
	if task.Method != http.MethodGet {
		parts = append(parts, fmt.Sprintf("-X %s", task.Method))
	}
	for key, value := range task.Headers {
		parts = append(parts, fmt.Sprintf("-H '%s: %s'", key, value))
	}
	if task.Body != "" {
		parts = append(parts, fmt.Sprintf("-d '%s'", strings.ReplaceAll(task.Body, "'", "'\"'\"'")))
	}
	return strings.Join(parts, " \\\n  ")
}

func parseCurl(command string) (Task, error) {
	command = strings.TrimSpace(command)
	if command == "" {
		return Task{}, errors.New("curl command is required")
	}
	task := Task{Method: http.MethodGet, Headers: map[string]string{}}
	urlPattern := regexp.MustCompile(`curl\s+(?:'([^']+)'|"([^"]+)"|([^\s]+))`)
	if match := urlPattern.FindStringSubmatch(command); len(match) > 0 {
		task.URL = firstNonEmpty(match[1], match[2], match[3])
	}
	methodPattern := regexp.MustCompile(`(?:-X|--request)\s+([A-Za-z]+)`)
	if match := methodPattern.FindStringSubmatch(command); len(match) > 1 {
		task.Method = strings.ToUpper(match[1])
	}
	headerPattern := regexp.MustCompile(`(?:-H|--header)\s+(?:'([^']+)'|"([^"]+)")`)
	for _, match := range headerPattern.FindAllStringSubmatch(command, -1) {
		header := firstNonEmpty(match[1], match[2])
		parts := strings.SplitN(header, ":", 2)
		if len(parts) == 2 {
			task.Headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	cookiePattern := regexp.MustCompile(`(?:-b|--cookie)\s+(?:'([^']+)'|"([^"]+)")`)
	if match := cookiePattern.FindStringSubmatch(command); len(match) > 0 {
		task.Headers["Cookie"] = firstNonEmpty(match[1], match[2])
	}
	bodyPattern := regexp.MustCompile(`(?:-d|--data|--data-raw|--data-binary)\s+(?:'([^']*)'|"([^"]*)")`)
	if match := bodyPattern.FindStringSubmatch(command); len(match) > 0 {
		task.Body = firstNonEmpty(match[1], match[2])
		if task.Method == http.MethodGet {
			task.Method = http.MethodPost
		}
	}
	if task.URL == "" {
		return Task{}, errors.New("failed to extract url from curl command")
	}
	return task, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func (s *Server) runTaskByID(id int, triggeredBy string) (HistoryItem, error) {
	s.mu.RLock()
	var task Task
	found := false
	for _, candidate := range s.state.Tasks {
		if candidate.ID == id {
			task = candidate
			found = true
			break
		}
	}
	s.mu.RUnlock()
	if !found {
		return HistoryItem{}, errors.New("task not found")
	}
	return s.executeTask(task, triggeredBy), nil
}

func (s *Server) runAllEnabledTasks(triggeredBy string) []HistoryItem {
	s.mu.RLock()
	tasks := []Task{}
	maxWorkers := s.state.Settings.Schedule.MaxWorkers
	for _, task := range s.state.Tasks {
		if task.Enabled {
			tasks = append(tasks, task)
		}
	}
	s.mu.RUnlock()

	if maxWorkers < 1 {
		maxWorkers = 4
	}
	if maxWorkers > 8 {
		maxWorkers = 8
	}
	if len(tasks) == 0 {
		return []HistoryItem{}
	}

	results := make([]HistoryItem, 0, len(tasks))
	var resultsMu sync.Mutex
	sem := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup
	for _, task := range tasks {
		wg.Add(1)
		go func(task Task) {
			defer wg.Done()
			sem <- struct{}{}
			result := s.executeTask(task, triggeredBy)
			<-sem
			resultsMu.Lock()
			results = append(results, result)
			resultsMu.Unlock()
		}(task)
	}
	wg.Wait()
	slices.SortFunc(results, func(a, b HistoryItem) int { return b.TaskID - a.TaskID })
	return results
}

func (s *Server) executeTask(task Task, triggeredBy string) HistoryItem {
	client := &http.Client{Timeout: time.Duration(task.TimeoutSeconds) * time.Second}
	var result HistoryItem
	var lastMessage string
	var lastStatusCode int
	var preview string
	var durationMS int

	for attempt := 0; attempt <= task.RetryCount; attempt++ {
		started := appNow()
		req, err := http.NewRequest(task.Method, task.URL, strings.NewReader(task.Body))
		if err != nil {
			lastMessage = err.Error()
			break
		}
		for key, value := range task.Headers {
			req.Header.Set(key, value)
		}
		resp, err := client.Do(req)
		durationMS = int(time.Since(started).Milliseconds())
		if err != nil {
			lastMessage = err.Error()
			if attempt < task.RetryCount {
				time.Sleep(time.Duration(attempt+1) * time.Second)
				continue
			}
			break
		}
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		_ = resp.Body.Close()
		lastStatusCode = resp.StatusCode
		preview = string(bodyBytes)
		lastMessage = evaluateResponse(task, resp.StatusCode, preview)
		if lastMessage == "success" {
			result = HistoryItem{
				TaskID:          task.ID,
				TaskName:        task.Name,
				Status:          "success",
				StatusCode:      resp.StatusCode,
				Message:         "request completed",
				ResponsePreview: preview,
				ResponseTimeMS:  durationMS,
				TriggeredBy:     triggeredBy,
				CreatedAt:       nowString(),
			}
			s.finishRun(task.ID, result)
			return result
		}
		if attempt < task.RetryCount {
			time.Sleep(time.Duration(attempt+1) * time.Second)
		}
	}

	result = HistoryItem{
		TaskID:          task.ID,
		TaskName:        task.Name,
		Status:          "failed",
		StatusCode:      lastStatusCode,
		Message:         lastMessage,
		ResponsePreview: preview,
		ResponseTimeMS:  durationMS,
		TriggeredBy:     triggeredBy,
		CreatedAt:       nowString(),
	}
	s.finishRun(task.ID, result)
	return result
}

func evaluateResponse(task Task, statusCode int, body string) string {
	if statusCode < 200 || statusCode >= 300 {
		return fmt.Sprintf("http %d", statusCode)
	}
	bodyLower := strings.ToLower(body)
	for _, word := range splitKeywords(task.FailureKeywords) {
		if strings.Contains(bodyLower, word) {
			return "failure keyword matched"
		}
	}
	successWords := splitKeywords(task.SuccessKeywords)
	if len(successWords) > 0 {
		matched := false
		for _, word := range successWords {
			if strings.Contains(bodyLower, word) {
				matched = true
				break
			}
		}
		if !matched {
			return "success keyword missing"
		}
	}
	return "success"
}

func splitKeywords(raw string) []string {
	parts := strings.Split(raw, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(strings.ToLower(part))
		if part != "" {
			items = append(items, part)
		}
	}
	return items
}

func (s *Server) finishRun(taskID int, item HistoryItem) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item.ID = s.state.NextHistoryID
	s.state.NextHistoryID++
	s.state.History = append([]HistoryItem{item}, s.state.History...)
	if len(s.state.History) > 200 {
		s.state.History = s.state.History[:200]
	}
	for i := range s.state.Tasks {
		if s.state.Tasks[i].ID == taskID {
			s.state.Tasks[i].LastStatus = item.Status
			s.state.Tasks[i].LastRunAt = item.CreatedAt
			s.state.Tasks[i].LastDurationMS = item.ResponseTimeMS
			s.state.Tasks[i].UpdatedAt = nowString()
			break
		}
	}
	_ = s.saveStateLocked()
	go s.sendNotifications(item)
}

func (s *Server) schedulerLoop() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		s.runScheduledTasks(appNow(), "schedule")
	}
}

func (s *Server) runScheduledTasks(now time.Time, triggeredBy string) []HistoryItem {
	s.mu.RLock()
	tasks := slices.Clone(s.state.Tasks)
	schedule := s.state.Settings.Schedule
	s.mu.RUnlock()
	due := make([]Task, 0)
	currentKey := now.Format("2006-01-02 15:04")
	for _, task := range tasks {
		if !task.Enabled {
			continue
		}
		hour, minute := schedule.Hour, schedule.Minute
		shouldRun := schedule.Enabled
		if task.ScheduleEnabled {
			hour, minute = task.ScheduleHour, task.ScheduleMinute
			shouldRun = true
		}
		if shouldRun && now.Hour() == hour && now.Minute() == minute {
			runKey := fmt.Sprintf("%d:%s", task.ID, currentKey)
			s.scheduleMu.Lock()
			if _, ok := s.scheduleRan[runKey]; !ok {
				s.scheduleRan[runKey] = struct{}{}
				due = append(due, task)
			}
			s.scheduleMu.Unlock()
		}
	}
	if len(due) == 0 {
		return nil
	}
	results := make([]HistoryItem, 0, len(due))
	for _, task := range due {
		results = append(results, s.executeTask(task, triggeredBy))
	}
	return results
}

func (s *Server) sendNotifications(item HistoryItem) {
	s.mu.RLock()
	notify := s.state.Settings.Notify
	s.mu.RUnlock()
	if item.Status == "success" && !notify.NotifyOnSuccess {
		return
	}
	if item.Status == "failed" && !notify.NotifyOnFailure {
		return
	}
	message := fmt.Sprintf(
		"%s\nTask: %s\nBy: %s\nStatus code: %d\nDuration: %d ms\nMessage: %s\nPreview:\n%s",
		strings.ToUpper(item.Status),
		item.TaskName,
		item.TriggeredBy,
		item.StatusCode,
		item.ResponseTimeMS,
		item.Message,
		item.ResponsePreview,
	)
	if notify.TelegramEnabled && notify.TelegramBotToken != "" && notify.TelegramChatID != "" {
		form := url.Values{}
		form.Set("chat_id", notify.TelegramChatID)
		form.Set("text", message)
		req, _ := http.NewRequest(http.MethodPost, "https://api.telegram.org/bot"+notify.TelegramBotToken+"/sendMessage", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		_, _ = s.httpClient.Do(req)
	}
	if notify.WebhookEnabled && notify.WebhookURL != "" {
		payload, _ := json.Marshal(map[string]any{"title": item.Status, "message": message, "item": item})
		req, _ := http.NewRequest(http.MethodPost, notify.WebhookURL, strings.NewReader(string(payload)))
		req.Header.Set("Content-Type", "application/json")
		_, _ = s.httpClient.Do(req)
	}
}
