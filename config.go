package main

import (
	"os"
	"strings"
	"time"
)

const (
	defaultTimezone      = "Asia/Shanghai"
	defaultAdminPassword = "admin123456"
	defaultAddr          = "0.0.0.0:8080"
)

// GetAppTimezone 获取应用配置的时区
func GetAppTimezone() *time.Location {
	tzName := strings.TrimSpace(os.Getenv("APP_TIMEZONE"))
	if tzName == "" {
		tzName = defaultTimezone
	}
	loc, err := time.LoadLocation(tzName)
	if err != nil {
		loc, err = time.LoadLocation(defaultTimezone)
		if err != nil {
			return time.Local
		}
	}
	return loc
}

// GetAppNow 获取带指定时区信息的当前时间
func GetAppNow() time.Time {
	return time.Now().In(GetAppTimezone())
}

// FormatNow 格式化当前时间为标准日期时间字符串 (YYYY-MM-DD HH:MM:SS)
func FormatNow() string {
	return GetAppNow().Format("2006-01-02 15:04:05")
}

// FormatNowMS 格式化当前时间为带毫秒的日期时间字符串 (YYYY-MM-DD HH:MM:SS.mmm)
func FormatNowMS() string {
	return GetAppNow().Format("2006-01-02 15:04:05.000")
}

// GetDefaultPassword 获取系统默认初始密码
func GetDefaultPassword() string {
	val := strings.TrimSpace(os.Getenv("QIANGDAO_DEFAULT_PASSWORD"))
	if val != "" {
		return val
	}
	return defaultAdminPassword
}

// IsDocsEnabled 判断是否开启 Swagger/Docs 自动文档
func IsDocsEnabled() bool {
	val := strings.ToLower(strings.TrimSpace(os.Getenv("ENABLE_DOCS")))
	return val == "true" || val == "1" || val == "yes" || val == "on"
}
