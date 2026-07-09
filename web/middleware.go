// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package web

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"ez-ddns/utils"
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

type responseRecorder struct {
	http.ResponseWriter
	body   bytes.Buffer
	status int
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func NewLoggerMiddleware(logger utils.Logger) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			var reqBody []byte
			if r.Body != nil {
				reqBody, _ = io.ReadAll(r.Body)
				r.Body = io.NopCloser(bytes.NewBuffer(reqBody))
			}

			recorder := &responseRecorder{
				ResponseWriter: w,
				status:         http.StatusOK,
			}

			r = r.WithContext(utils.ContextWithLogger(r.Context(), logger))

			next(recorder, r)

			logger.Debug("API Request - Method: %s, Path: %s, Status: %d, RequestBody: %s, ResponseBody: %s",
				r.Method,
				r.URL.Path,
				recorder.status,
				string(reqBody),
				recorder.body.String(),
			)
		}
	}
}
