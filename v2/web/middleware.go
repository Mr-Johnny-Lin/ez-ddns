// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package web

import (
	"net/http"
	"strings"

	"ez-ddns/v2/utils"
)

type AuthConfig struct {
	APIKey string
}

func NewAuthMiddleware(authConfig *AuthConfig) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if authConfig.APIKey == "" {
				next(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				JSONError(w, http.StatusUnauthorized, "缺少授权头")
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				JSONError(w, http.StatusUnauthorized, "授权格式错误")
				return
			}

			token := parts[1]
			if !utils.ValidateToken(token, authConfig.APIKey) {
				JSONError(w, http.StatusUnauthorized, "无效的API密钥")
				return
			}

			next(w, r)
		}
	}
}

func CorsMiddleware(next HandlerFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func LoggerMiddleware(next HandlerFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		next(w, r)
	}
}
