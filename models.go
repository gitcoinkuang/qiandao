package main

// Task 签到任务数据结构
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
	ScheduleSecond  int               `json:"schedule_second"`
	TimeoutSeconds  int               `json:"timeout_seconds"`
	RetryCount      int               `json:"retry_count"`
	AggressiveMode  bool              `json:"aggressive_mode"`
	SuccessKeywords string            `json:"success_keywords"`
	FailureKeywords string            `json:"failure_keywords"`
	LastStatus      string            `json:"last_status"`
	LastRunAt       string            `json:"last_run_at"`
	LastDurationMS  int               `json:"last_duration_ms"`
	CreatedAt       string            `json:"created_at"`
	UpdatedAt       string            `json:"updated_at"`
}

// HistoryItem 历史记录数据结构
type HistoryItem struct {
	ID               int    `json:"id"`
	TaskID           int    `json:"task_id"`
	TaskName         string `json:"task_name"`
	Status           string `json:"status"`
	StatusCode       int    `json:"status_code"`
	Message          string `json:"message"`
	ResponsePreview  string `json:"response_preview"`
	ResponseTimeMS   int    `json:"response_time_ms"`
	RequestStartedAt string `json:"request_started_at"`
	TriggeredBy      string `json:"triggered_by"`
	CreatedAt        string `json:"created_at"`
}

// NotifySettings 消息推送设置
type NotifySettings struct {
	TelegramEnabled  bool   `json:"telegram_enabled"`
	TelegramBotToken string `json:"telegram_bot_token"`
	TelegramChatID   string `json:"telegram_chat_id"`
	WebhookEnabled   bool   `json:"webhook_enabled"`
	WebhookURL       string `json:"webhook_url"`
	NotifyOnSuccess  bool   `json:"notify_on_success"`
	NotifyOnFailure  bool   `json:"notify_on_failure"`
}

// ScheduleSettings 全局定时调度设置
type ScheduleSettings struct {
	Enabled    bool `json:"enabled"`
	Hour       int  `json:"hour"`
	Minute     int  `json:"minute"`
	Second     int  `json:"second"`
	MaxWorkers int  `json:"max_workers"`
}

// SecuritySettings 访问安全配置
type SecuritySettings struct {
	Enabled      bool   `json:"enabled"`
	PasswordHash string `json:"password_hash"`
}

// Settings 全局系统配置集合
type Settings struct {
	Notify   NotifySettings   `json:"notify"`
	Schedule ScheduleSettings `json:"schedule"`
	Security SecuritySettings `json:"security"`
}

// AppState 系统持久化状态
type AppState struct {
	NextTaskID    int           `json:"next_task_id"`
	NextHistoryID int           `json:"next_history_id"`
	Tasks         []Task        `json:"tasks"`
	History       []HistoryItem `json:"history"`
	Settings      Settings      `json:"settings"`
}

// APIResponse 统一 API 响应格式
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
