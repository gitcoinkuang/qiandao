package main

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

var (
	scheduleRanMu  sync.Mutex
	scheduleRanMap = make(map[string]time.Time)
)

func shouldTriggerAt(now time.Time, hour, minute, second int) bool {
	if now.Hour() != hour || now.Minute() != minute {
		return false
	}
	curSec := now.Second()
	return curSec >= second && curSec < (second+3)
}

func scheduledRunKey(taskID int, now time.Time, hour, minute, second int) string {
	return fmt.Sprintf("%d:%04d-%02d-%02d %02d:%02d:%02d",
		taskID, now.Year(), int(now.Month()), now.Day(), hour, minute, second)
}

// cleanStaleScheduleKeys 清理超过 10 分钟的旧去重 key
func cleanStaleScheduleKeys(now time.Time) {
	scheduleRanMu.Lock()
	defer scheduleRanMu.Unlock()
	for k, t := range scheduleRanMap {
		if now.Sub(t) > 10*time.Minute {
			delete(scheduleRanMap, k)
		}
	}
}

// RunScheduledTasks 检查并触发所有到达定时时间的任务
func RunScheduledTasks(ctx context.Context, now time.Time, triggeredBy string, stateMgr *StateManager) []HistoryItem {
	cleanStaleScheduleKeys(now)

	stateMgr.Mu.RLock()
	tasks := make([]Task, len(stateMgr.State.Tasks))
	copy(tasks, stateMgr.State.Tasks)
	scheduleSettings := stateMgr.State.Settings.Schedule
	stateMgr.Mu.RUnlock()

	var dueTasks []Task

	for _, task := range tasks {
		if !task.Enabled {
			continue
		}

		hour, minute, second := scheduleSettings.Hour, scheduleSettings.Minute, scheduleSettings.Second
		shouldRun := scheduleSettings.Enabled

		if task.ScheduleEnabled {
			hour, minute, second = task.ScheduleHour, task.ScheduleMinute, task.ScheduleSecond
			shouldRun = true
		}

		if shouldRun && shouldTriggerAt(now, hour, minute, second) {
			key := scheduledRunKey(task.ID, now, hour, minute, second)
			scheduleRanMu.Lock()
			_, ran := scheduleRanMap[key]
			if !ran {
				scheduleRanMap[key] = now
				dueTasks = append(dueTasks, task)
			}
			scheduleRanMu.Unlock()
		}
	}

	if len(dueTasks) == 0 {
		return []HistoryItem{}
	}

	var results []HistoryItem
	for _, t := range dueTasks {
		res := ExecuteTask(ctx, t, triggeredBy, stateMgr)
		results = append(results, res)
	}

	return results
}

// SchedulerLoop 后台运行的秒级循环调度器
func SchedulerLoop(ctx context.Context, stateMgr *StateManager) {
	slog.Info("后台秒级定时调度引擎已启动")
	for {
		select {
		case <-ctx.Done():
			slog.Info("后台定时调度引擎已停止")
			return
		default:
			now := GetAppNow()
			RunScheduledTasks(ctx, now, "schedule", stateMgr)

			// 对齐下一次整秒时间点 (微秒级精准)
			nanoSec := time.Now().Nanosecond()
			sleepDuration := time.Duration(1000000000-nanoSec) * time.Nanosecond
			time.Sleep(sleepDuration)
		}
	}
}
