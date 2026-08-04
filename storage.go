package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
)

// StateManager 状态管理器，提供线程安全的 JSON 文件持久化
type StateManager struct {
	StatePath string
	Mu        sync.RWMutex
	State     AppState
}

// NewStateManager 创建状态管理器实例
func NewStateManager(path string) *StateManager {
	if path == "" {
		path = filepath.Join("data", "state.json")
	}
	return &StateManager{
		StatePath: path,
		State:     defaultState(),
	}
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
				Second:     0,
				MaxWorkers: 4,
			},
			Security: SecuritySettings{
				Enabled:      true,
				PasswordHash: HashPassword(GetDefaultPassword()),
			},
		},
	}
}

// Load 加载 state.json 配置文件
func (sm *StateManager) Load() error {
	sm.Mu.Lock()
	defer sm.Mu.Unlock()

	if _, err := os.Stat(sm.StatePath); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(sm.StatePath), 0755); err != nil {
			return fmt.Errorf("创建数据目录失败: %w", err)
		}
		sm.State = defaultState()
		return sm.saveUnlocked()
	}

	data, err := os.ReadFile(sm.StatePath)
	if err != nil {
		slog.Error("读取状态文件失败，恢复默认配置", "error", err)
		sm.State = defaultState()
		return sm.saveUnlocked()
	}

	var loaded AppState
	if err := json.Unmarshal(data, &loaded); err != nil {
		slog.Error("解析状态文件 JSON 失败，恢复默认配置", "error", err)
		sm.State = defaultState()
		return sm.saveUnlocked()
	}

	// 容错与修复
	if loaded.NextTaskID < 1 {
		loaded.NextTaskID = 1
	}
	if loaded.NextHistoryID < 1 {
		loaded.NextHistoryID = 1
	}
	if loaded.Tasks == nil {
		loaded.Tasks = []Task{}
	}
	if loaded.History == nil {
		loaded.History = []HistoryItem{}
	}
	if loaded.Settings.Security.PasswordHash == "" {
		loaded.Settings.Security.PasswordHash = HashPassword(GetDefaultPassword())
		loaded.Settings.Security.Enabled = true
	}

	sm.State = loaded
	return nil
}

// Save 安全保存状态到文件（自动加写锁）
func (sm *StateManager) Save() error {
	sm.Mu.Lock()
	defer sm.Mu.Unlock()
	return sm.saveUnlocked()
}

// saveUnlocked 写入临时文件后原子重命名
func (sm *StateManager) saveUnlocked() error {
	if err := os.MkdirAll(filepath.Dir(sm.StatePath), 0755); err != nil {
		return fmt.Errorf("创建数据目录失败: %w", err)
	}

	data, err := json.MarshalIndent(sm.State, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化状态数据失败: %w", err)
	}

	tmpPath := sm.StatePath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("写入临时状态文件失败: %w", err)
	}

	if err := os.Rename(tmpPath, sm.StatePath); err != nil {
		return fmt.Errorf("覆盖替换状态文件失败: %w", err)
	}
	return nil
}
