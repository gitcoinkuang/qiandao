package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	sessionMu sync.RWMutex
	sessions  = make(map[string]time.Time)
)

// HashPassword 计算密码的 SHA256 哈希值
func HashPassword(password string) string {
	h := sha256.Sum256([]byte(password))
	return hex.EncodeToString(h[:])
}

// GenerateToken 生成指定字节长度的随机 16 进制 Token 字符串
func GenerateToken(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)
}

// CreateSession 创建 Session，有效期 7 天
func CreateSession(token string) {
	sessionMu.Lock()
	defer sessionMu.Unlock()
	sessions[token] = GetAppNow().Add(7 * 24 * time.Hour)
}

// DeleteSession 删除指定的 Session Token
func DeleteSession(token string) {
	sessionMu.Lock()
	defer sessionMu.Unlock()
	delete(sessions, token)
}

// IsAuthenticated 判断 Gin 请求上下文是否已登录认证
func IsAuthenticated(c *gin.Context, securityEnabled bool) bool {
	if !securityEnabled {
		return true
	}
	cookieVal, err := c.Cookie("qiandao_session")
	if err != nil || cookieVal == "" {
		return false
	}

	sessionMu.RLock()
	expiresAt, exists := sessions[cookieVal]
	sessionMu.RUnlock()

	if !exists {
		return false
	}

	if GetAppNow().After(expiresAt) {
		DeleteSession(cookieVal)
		return false
	}

	return true
}

// CheckPassword 校验明文密码是否与哈希结果一致
func CheckPassword(password string, storedHash string, securityEnabled bool) bool {
	if !securityEnabled {
		return true
	}
	return storedHash == HashPassword(password)
}
