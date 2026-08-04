package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var sharedHTTPClient = &http.Client{
	Timeout: 10 * time.Second,
}

// SendNotifications 异步推送任务日志结果通知
func SendNotifications(item HistoryItem, notify NotifySettings) {
	if item.Status == "success" && !notify.NotifyOnSuccess {
		return
	}
	if item.Status == "failed" && !notify.NotifyOnFailure {
		return
	}

	msg := fmt.Sprintf("%s\nTask: %s\nBy: %s\nStatus code: %d\nDuration: %d ms\nMessage: %s\nPreview:\n%s",
		strings.ToUpper(item.Status),
		item.TaskName,
		item.TriggeredBy,
		item.StatusCode,
		item.ResponseTimeMS,
		item.Message,
		item.ResponsePreview,
	)

	// Telegram 通知推送
	if notify.TelegramEnabled && notify.TelegramBotToken != "" && notify.TelegramChatID != "" {
		tgURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", notify.TelegramBotToken)
		go func() {
			formData := url.Values{}
			formData.Set("chat_id", notify.TelegramChatID)
			formData.Set("text", msg)
			resp, err := sharedHTTPClient.PostForm(tgURL, formData)
			if err == nil {
				_ = resp.Body.Close()
			} else {
				slog.Error("Telegram 通知发送失败", "error", err)
			}
		}()
	}

	// Webhook 通知推送
	if notify.WebhookEnabled && notify.WebhookURL != "" {
		go func() {
			payload := map[string]interface{}{
				"title":   item.Status,
				"message": msg,
				"item":    item,
			}
			data, err := json.Marshal(payload)
			if err != nil {
				return
			}
			resp, err := sharedHTTPClient.Post(notify.WebhookURL, "application/json", bytes.NewBuffer(data))
			if err == nil {
				_ = resp.Body.Close()
			} else {
				slog.Error("Webhook 通知发送失败", "error", err)
			}
		}()
	}
}

// SendTestTelegramNotification 发送 Telegram 测试通知
func SendTestTelegramNotification(notify NotifySettings) (bool, string) {
	if !notify.TelegramEnabled || notify.TelegramBotToken == "" || notify.TelegramChatID == "" {
		return false, "Telegram 未配置或未启用，请先保存通知设置"
	}

	tgURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", notify.TelegramBotToken)
	msg := "This is a test notification from QianDao V2."

	formData := url.Values{}
	formData.Set("chat_id", notify.TelegramChatID)
	formData.Set("text", msg)

	resp, err := sharedHTTPClient.PostForm(tgURL, formData)
	if err != nil {
		return false, fmt.Sprintf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return false, fmt.Sprintf("Telegram API 返回异常: %s", string(bodyBytes))
	}

	return true, ""
}
