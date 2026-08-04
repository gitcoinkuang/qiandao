package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	scheduledBurstRetries  = 3
	scheduledBurstInterval = 250 * time.Millisecond
)

var defaultTransport = &http.Transport{
	MaxIdleConns:        128,
	MaxIdleConnsPerHost: 64,
	IdleConnTimeout:     90 * time.Second,
}

var runnerHTTPClient = &http.Client{
	Transport: defaultTransport,
}

// EvaluateResponse 评估 HTTP 状态码及关键字规则
func EvaluateResponse(task Task, statusCode int, body string) string {
	if statusCode < 200 || statusCode >= 300 {
		return fmt.Sprintf("http %d", statusCode)
	}

	bodyLower := strings.ToLower(body)

	// 校验失败关键字
	if task.FailureKeywords != "" {
		failWords := strings.Split(task.FailureKeywords, ",")
		for _, w := range failWords {
			word := strings.ToLower(strings.TrimSpace(w))
			if word != "" && strings.Contains(bodyLower, word) {
				return "failure keyword matched"
			}
		}
	}

	// 校验成功关键字
	if task.SuccessKeywords != "" {
		succWords := strings.Split(task.SuccessKeywords, ",")
		var matched bool
		hasValidWord := false
		for _, w := range succWords {
			word := strings.ToLower(strings.TrimSpace(w))
			if word != "" {
				hasValidWord = true
				if strings.Contains(bodyLower, word) {
					matched = true
					break
				}
			}
		}
		if hasValidWord && !matched {
			return "success keyword missing"
		}
	}

	return "success"
}

// ShouldUseBurstMode 判断是否开启抢零点 Burst 重试模式
func ShouldUseBurstMode(task Task, triggeredBy string) bool {
	if !task.AggressiveMode {
		return false
	}
	return triggeredBy == "schedule" || triggeredBy == "manual-schedule-check"
}

// ExecuteTask 执行单个签到任务
func ExecuteTask(ctx context.Context, task Task, triggeredBy string, stateMgr *StateManager) HistoryItem {
	maxAttempts := task.RetryCount + 1
	burstMode := ShouldUseBurstMode(task, triggeredBy)

	if burstMode && maxAttempts < scheduledBurstRetries+1 {
		maxAttempts = scheduledBurstRetries + 1
	}

	var lastMessage string
	var lastStatusCode int
	var preview string
	var durationMS int
	var lastRequestStartedAt string

	for attempt := 0; attempt < maxAttempts; attempt++ {
		startTime := time.Now()
		requestStartedAt := FormatNowMS()
		lastRequestStartedAt = requestStartedAt

		reqTimeout := time.Duration(task.TimeoutSeconds) * time.Second
		if reqTimeout <= 0 {
			reqTimeout = 30 * time.Second
		}
		reqCtx, cancel := context.WithTimeout(ctx, reqTimeout)

		var reqBody io.Reader
		if task.Body != "" {
			reqBody = bytes.NewBufferString(task.Body)
		}

		req, err := http.NewRequestWithContext(reqCtx, task.Method, task.URL, reqBody)
		if err == nil {
			for k, v := range task.Headers {
				req.Header.Set(k, v)
			}

			resp, err := runnerHTTPClient.Do(req)
			if err == nil {
				durationMS = int(time.Since(startTime).Milliseconds())
				lastStatusCode = resp.StatusCode

				// 最多读取前 4096 字节
				bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
				_ = resp.Body.Close()
				preview = string(bodyBytes)

				lastMessage = EvaluateResponse(task, resp.StatusCode, preview)
				cancel()

				if lastMessage == "success" {
					result := HistoryItem{
						TaskID:           task.ID,
						TaskName:         task.Name,
						Status:           "success",
						StatusCode:       resp.StatusCode,
						Message:          "request completed",
						ResponsePreview:  preview,
						ResponseTimeMS:   durationMS,
						RequestStartedAt: requestStartedAt,
						TriggeredBy:      triggeredBy,
						CreatedAt:        FormatNow(),
					}
					finishRun(task.ID, result, stateMgr)
					return result
				}
			} else {
				durationMS = int(time.Since(startTime).Milliseconds())
				lastMessage = err.Error()
				cancel()
			}
		} else {
			lastMessage = err.Error()
			cancel()
		}

		// 失败重试等待
		if attempt < maxAttempts-1 {
			if burstMode && attempt < scheduledBurstRetries {
				time.Sleep(scheduledBurstInterval)
			} else {
				time.Sleep(time.Duration(attempt+1) * time.Second)
			}
		}
	}

	result := HistoryItem{
		TaskID:           task.ID,
		TaskName:         task.Name,
		Status:           "failed",
		StatusCode:       lastStatusCode,
		Message:          lastMessage,
		ResponsePreview:  preview,
		ResponseTimeMS:   durationMS,
		RequestStartedAt: lastRequestStartedAt,
		TriggeredBy:      triggeredBy,
		CreatedAt:        FormatNow(),
	}
	finishRun(task.ID, result, stateMgr)
	return result
}

// finishRun 保存记录、更新任务运行状态并异步推送通知
func finishRun(taskID int, item HistoryItem, stateMgr *StateManager) {
	stateMgr.Mu.Lock()

	item.ID = stateMgr.State.NextHistoryID
	stateMgr.State.NextHistoryID++

	// 插入到最前面
	stateMgr.State.History = append([]HistoryItem{item}, stateMgr.State.History...)

	// 限制历史记录最多 200 条
	if len(stateMgr.State.History) > 200 {
		stateMgr.State.History = stateMgr.State.History[:200]
	}

	for i := range stateMgr.State.Tasks {
		if stateMgr.State.Tasks[i].ID == taskID {
			stateMgr.State.Tasks[i].LastStatus = item.Status
			stateMgr.State.Tasks[i].LastRunAt = item.CreatedAt
			stateMgr.State.Tasks[i].LastDurationMS = item.ResponseTimeMS
			stateMgr.State.Tasks[i].UpdatedAt = FormatNow()
			break
		}
	}

	notifySettings := stateMgr.State.Settings.Notify
	_ = stateMgr.saveUnlocked()
	stateMgr.Mu.Unlock()

	// 异步推送通知
	go SendNotifications(item, notifySettings)
}

// RunAllEnabledTasks 并发运行所有已启用的任务
func RunAllEnabledTasks(ctx context.Context, triggeredBy string, stateMgr *StateManager) []HistoryItem {
	stateMgr.Mu.RLock()
	var enabledTasks []Task
	for _, t := range stateMgr.State.Tasks {
		if t.Enabled {
			enabledTasks = append(enabledTasks, t)
		}
	}
	maxWorkers := stateMgr.State.Settings.Schedule.MaxWorkers
	stateMgr.Mu.RUnlock()

	if len(enabledTasks) == 0 {
		return []HistoryItem{}
	}

	if maxWorkers < 1 {
		maxWorkers = 4
	} else if maxWorkers > 8 {
		maxWorkers = 8
	}

	sem := make(chan struct{}, maxWorkers)
	resultsChan := make(chan HistoryItem, len(enabledTasks))
	var wg sync.WaitGroup

	for _, t := range enabledTasks {
		wg.Add(1)
		go func(task Task) {
			defer wg.Done()
			sem <- struct{}{}
			res := ExecuteTask(ctx, task, triggeredBy, stateMgr)
			<-sem
			resultsChan <- res
		}(t)
	}

	wg.Wait()
	close(resultsChan)

	var results []HistoryItem
	for r := range resultsChan {
		results = append(results, r)
	}

	// 按 TaskID 倒序排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].TaskID > results[j].TaskID
	})

	return results
}
