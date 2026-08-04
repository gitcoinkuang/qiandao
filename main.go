// 用途：QianDao V2 自动签到与定时 HTTP 请求管理系统主服务入口
// Usage: go run .

package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

var stateManager *StateManager

func main() {
	// 使用 slog 日志规范
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// 初始化状态持久化管理器
	stateManager = NewStateManager("")
	if err := stateManager.Load(); err != nil {
		slog.Error("初始化应用持久化状态失败", "error", err)
		os.Exit(1)
	}

	// 启动后台定时调度引擎
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go SchedulerLoop(ctx, stateManager)

	// 设置 Gin 运行模式
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	// 挂载静态文件与 HTML 模板
	r.Static("/static", "./static")
	r.LoadHTMLGlob("templates/*")

	// 路由注册
	registerWebRoutes(r)
	registerAPIRoutes(r)

	serverAddr := defaultAddr
	if envAddr := strings.TrimSpace(os.Getenv("ADDR")); envAddr != "" {
		serverAddr = envAddr
	}

	srv := &http.Server{
		Addr:    serverAddr,
		Handler: r,
	}

	slog.Info("QianDao V2 服务已启动", "addr", fmt.Sprintf("http://%s", serverAddr))

	// 监听系统信号实现优雅退出
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("服务启动失败", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("正在停止 QianDao V2 服务...")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("服务强制退出", "error", err)
	}
	slog.Info("服务已成功优雅退出")
}

// requireAuth 鉴权中间件
func requireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		stateManager.Mu.RLock()
		securityEnabled := stateManager.State.Settings.Security.Enabled
		stateManager.Mu.RUnlock()

		if securityEnabled && !IsAuthenticated(c, securityEnabled) {
			c.JSON(http.StatusUnauthorized, APIResponse{
				Success: false,
				Error:   "unauthorized",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// ---------------- Web 视图路由 ----------------

func registerWebRoutes(r *gin.Engine) {
	r.GET("/", func(c *gin.Context) {
		stateManager.Mu.RLock()
		securityEnabled := stateManager.State.Settings.Security.Enabled
		stateManager.Mu.RUnlock()

		if securityEnabled && !IsAuthenticated(c, securityEnabled) {
			c.Redirect(http.StatusSeeOther, "/login")
			return
		}
		c.HTML(http.StatusOK, "index.html", gin.H{
			"Title": "QianDao V2",
		})
	})

	r.GET("/login", func(c *gin.Context) {
		stateManager.Mu.RLock()
		securityEnabled := stateManager.State.Settings.Security.Enabled
		stateManager.Mu.RUnlock()

		if !securityEnabled {
			c.Redirect(http.StatusSeeOther, "/")
			return
		}
		c.HTML(http.StatusOK, "login.html", gin.H{
			"Error": "",
		})
	})

	r.POST("/login", func(c *gin.Context) {
		stateManager.Mu.RLock()
		securityEnabled := stateManager.State.Settings.Security.Enabled
		storedHash := stateManager.State.Settings.Security.PasswordHash
		stateManager.Mu.RUnlock()

		if !securityEnabled {
			c.Redirect(http.StatusSeeOther, "/")
			return
		}

		password := strings.TrimSpace(c.PostForm("password"))
		if !CheckPassword(password, storedHash, securityEnabled) {
			c.HTML(http.StatusOK, "login.html", gin.H{
				"Error": "密码错误",
			})
			return
		}

		token := GenerateToken(32)
		CreateSession(token)

		c.SetCookie("qiandao_session", token, 7*86400, "/", "", false, true)
		c.Redirect(http.StatusSeeOther, "/")
	})

	r.GET("/logout", func(c *gin.Context) {
		cookieVal, err := c.Cookie("qiandao_session")
		if err == nil && cookieVal != "" {
			DeleteSession(cookieVal)
		}
		c.SetCookie("qiandao_session", "", -1, "/", "", false, true)
		c.Redirect(http.StatusSeeOther, "/login")
	})
}

// ---------------- REST API 路由 ----------------

func registerAPIRoutes(r *gin.Engine) {
	api := r.Group("/api")
	api.Use(requireAuth())

	api.GET("/bootstrap", func(c *gin.Context) {
		stateManager.Mu.RLock()
		defer stateManager.Mu.RUnlock()

		history := stateManager.State.History
		tasks := stateManager.State.Tasks

		successCount := 0
		failedCount := 0
		var durationItems []int
		for _, item := range history {
			if item.Status == "success" {
				successCount++
			} else if item.Status == "failed" {
				failedCount++
			}
			if item.ResponseTimeMS > 0 {
				durationItems = append(durationItems, item.ResponseTimeMS)
			}
		}

		avgDuration := 0
		if len(durationItems) > 0 {
			total := 0
			for _, d := range durationItems {
				total += d
			}
			avgDuration = total / len(durationItems)
		}

		enabledCount := 0
		for _, t := range tasks {
			if t.Enabled {
				enabledCount++
			}
		}

		data := map[string]interface{}{
			"stats": map[string]interface{}{
				"total_tasks":     len(tasks),
				"enabled_tasks":   enabledCount,
				"recent_success":  successCount,
				"recent_failed":   failedCount,
				"avg_duration_ms": avgDuration,
			},
			"tasks":           tasks,
			"history":         history,
			"notify_config":   stateManager.State.Settings.Notify,
			"schedule_config": stateManager.State.Settings.Schedule,
			"security_config": map[string]interface{}{
				"enabled": stateManager.State.Settings.Security.Enabled,
			},
		}

		c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Data:    data,
		})
	})

	api.GET("/tasks", func(c *gin.Context) {
		stateManager.Mu.RLock()
		tasks := stateManager.State.Tasks
		stateManager.Mu.RUnlock()

		c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Tasks:   tasks,
		})
	})

	api.POST("/tasks", func(c *gin.Context) {
		var payload Task
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Error:   "无效请求体参数",
			})
			return
		}

		task, err := NormalizeTask(payload)
		if err != nil {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Error:   err.Error(),
			})
			return
		}

		stateManager.Mu.Lock()
		task.ID = stateManager.State.NextTaskID
		stateManager.State.NextTaskID++
		nowStr := FormatNow()
		task.CreatedAt = nowStr
		task.UpdatedAt = nowStr
		task.LastStatus = "idle"
		stateManager.State.Tasks = append([]Task{task}, stateManager.State.Tasks...)
		_ = stateManager.saveUnlocked()
		stateManager.Mu.Unlock()

		c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Task:    task,
		})
	})

	api.POST("/tasks/parse", func(c *gin.Context) {
		var payload Task
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Error:   "无效请求体参数",
			})
			return
		}

		parsedURL, parsedMethod, parsedHeaders, parsedBody, err := ParseCurl(payload.CurlCommand)
		if err != nil {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Error:   err.Error(),
			})
			return
		}

		method := parsedMethod
		if payload.Method != "" {
			method = strings.ToUpper(payload.Method)
		}

		parsed := Task{
			Name:        payload.Name,
			URL:         parsedURL,
			Method:      method,
			Headers:     parsedHeaders,
			Body:        parsedBody,
			CurlCommand: payload.CurlCommand,
		}

		c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Config:  parsed,
		})
	})

	api.POST("/tasks/run-all", func(c *gin.Context) {
		results := RunAllEnabledTasks(c.Request.Context(), "manual-batch", stateManager)
		c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Results: results,
		})
	})

	api.POST("/tasks/:id/run", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Error:   "无效任务 ID",
			})
			return
		}

		var targetTask *Task
		stateManager.Mu.RLock()
		for _, t := range stateManager.State.Tasks {
			if t.ID == id {
				taskCopy := t
				targetTask = &taskCopy
				break
			}
		}
		stateManager.Mu.RUnlock()

		if targetTask == nil {
			c.JSON(http.StatusNotFound, APIResponse{
				Success: false,
				Error:   "task not found",
			})
			return
		}

		result := ExecuteTask(c.Request.Context(), *targetTask, "manual", stateManager)
		c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Result:  result,
		})
	})

	api.PUT("/tasks/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Error:   "无效任务 ID",
			})
			return
		}

		var payload Task
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Error:   "无效请求体参数",
			})
			return
		}

		updated, err := NormalizeTask(payload)
		if err != nil {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Error:   err.Error(),
			})
			return
		}

		found := false
		stateManager.Mu.Lock()
		for i, t := range stateManager.State.Tasks {
			if t.ID == id {
				updated.ID = id
				updated.CreatedAt = t.CreatedAt
				updated.LastStatus = t.LastStatus
				updated.LastRunAt = t.LastRunAt
				updated.LastDurationMS = t.LastDurationMS
				updated.UpdatedAt = FormatNow()
				stateManager.State.Tasks[i] = updated
				_ = stateManager.saveUnlocked()
				found = true
				break
			}
		}
		stateManager.Mu.Unlock()

		if !found {
			c.JSON(http.StatusNotFound, APIResponse{
				Success: false,
				Error:   "task not found",
			})
			return
		}

		c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Task:    updated,
		})
	})

	api.DELETE("/tasks/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Error:   "无效任务 ID",
			})
			return
		}

		found := false
		stateManager.Mu.Lock()
		for i, t := range stateManager.State.Tasks {
			if t.ID == id {
				stateManager.State.Tasks = append(stateManager.State.Tasks[:i], stateManager.State.Tasks[i+1:]...)
				var newHistory []HistoryItem
				for _, h := range stateManager.State.History {
					if h.TaskID != id {
						newHistory = append(newHistory, h)
					}
				}
				stateManager.State.History = newHistory
				_ = stateManager.saveUnlocked()
				found = true
				break
			}
		}
		stateManager.Mu.Unlock()

		if !found {
			c.JSON(http.StatusNotFound, APIResponse{
				Success: false,
				Error:   "task not found",
			})
			return
		}

		c.JSON(http.StatusOK, APIResponse{
			Success: true,
		})
	})

	api.GET("/history", func(c *gin.Context) {
		stateManager.Mu.RLock()
		history := stateManager.State.History
		stateManager.Mu.RUnlock()

		c.JSON(http.StatusOK, APIResponse{
			Success: true,
			History: history,
		})
	})

	api.DELETE("/history", func(c *gin.Context) {
		stateManager.Mu.Lock()
		stateManager.State.History = []HistoryItem{}
		_ = stateManager.saveUnlocked()
		stateManager.Mu.Unlock()

		c.JSON(http.StatusOK, APIResponse{
			Success: true,
		})
	})

	api.PUT("/settings/notify", func(c *gin.Context) {
		var notifyConfig NotifySettings
		if err := c.ShouldBindJSON(&notifyConfig); err != nil {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Error:   "无效的配置参数",
			})
			return
		}

		stateManager.Mu.Lock()
		stateManager.State.Settings.Notify = notifyConfig
		_ = stateManager.saveUnlocked()
		stateManager.Mu.Unlock()

		c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Config:  notifyConfig,
		})
	})

	api.POST("/settings/notify/test", func(c *gin.Context) {
		stateManager.Mu.RLock()
		notifyConfig := stateManager.State.Settings.Notify
		stateManager.Mu.RUnlock()

		success, errMsg := SendTestTelegramNotification(notifyConfig)
		if !success {
			c.JSON(http.StatusOK, APIResponse{
				Success: false,
				Error:   errMsg,
			})
			return
		}

		c.JSON(http.StatusOK, APIResponse{
			Success: true,
		})
	})

	api.PUT("/settings/schedule", func(c *gin.Context) {
		var scheduleConfig ScheduleSettings
		if err := c.ShouldBindJSON(&scheduleConfig); err != nil {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Error:   "无效的配置参数",
			})
			return
		}

		if scheduleConfig.Hour < 0 || scheduleConfig.Hour > 23 ||
			scheduleConfig.Minute < 0 || scheduleConfig.Minute > 59 ||
			scheduleConfig.Second < 0 || scheduleConfig.Second > 59 {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Error:   "invalid schedule time",
			})
			return
		}

		if scheduleConfig.MaxWorkers < 1 {
			scheduleConfig.MaxWorkers = 1
		} else if scheduleConfig.MaxWorkers > 8 {
			scheduleConfig.MaxWorkers = 8
		}

		stateManager.Mu.Lock()
		stateManager.State.Settings.Schedule = scheduleConfig
		_ = stateManager.saveUnlocked()
		stateManager.Mu.Unlock()

		c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Config:  scheduleConfig,
		})
	})

	api.POST("/settings/schedule/check", func(c *gin.Context) {
		now := GetAppNow()
		results := RunScheduledTasks(c.Request.Context(), now, "manual-schedule-check", stateManager)
		c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Results: results,
		})
	})

	api.PUT("/settings/security", func(c *gin.Context) {
		stateManager.Mu.RLock()
		securityEnabled := stateManager.State.Settings.Security.Enabled
		stateManager.Mu.RUnlock()

		if securityEnabled && !IsAuthenticated(c, securityEnabled) {
			c.JSON(http.StatusUnauthorized, APIResponse{
				Success: false,
				Error:   "unauthorized",
			})
			return
		}

		var payload struct {
			Enabled  *bool  `json:"enabled"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Error:   "无效请求体参数",
			})
			return
		}

		enabled := true
		if payload.Enabled != nil {
			enabled = *payload.Enabled
		}
		password := strings.TrimSpace(payload.Password)

		if password != "" && len(password) < 6 {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Error:   "密码长度至少为 6 位",
			})
			return
		}

		stateManager.Mu.Lock()
		if password != "" {
			stateManager.State.Settings.Security.PasswordHash = HashPassword(password)
		}
		stateManager.State.Settings.Security.Enabled = enabled
		_ = stateManager.saveUnlocked()
		stateManager.Mu.Unlock()

		c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Config: map[string]interface{}{
				"enabled": enabled,
			},
		})
	})
}
