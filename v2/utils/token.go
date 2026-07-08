// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package utils

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func GenerateAuthToken(apiKey string) string {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := generateNonce()
	signature := GenerateSignature(timestamp, nonce, apiKey)
	return fmt.Sprintf("%s:%s:%s", timestamp, nonce, signature)
}

func GenerateSignature(timestamp, nonce, apiKey string) string {
	data := fmt.Sprintf("%s:%s", timestamp, nonce)
	h := hmac.New(sha256.New, []byte(apiKey))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func ValidateToken(token, apiKey string) bool {
	parts := strings.Split(token, ":")
	if len(parts) != 3 {
		return false
	}

	timestampStr := parts[0]
	nonce := parts[1]
	signature := parts[2]

	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return false
	}

	now := time.Now().Unix()
	if now-timestamp > 300 {
		return false
	}

	expectedSignature := GenerateSignature(timestampStr, nonce, apiKey)
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

func generateNonce() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
