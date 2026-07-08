// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package web

import (
	"net/http"
)

type Server struct {
	router     *Router
	apiKey     string
	addr       string
	httpServer *http.Server
}

type ServerConfig struct {
	Addr   string
	APIKey string
}

func NewServer(config *ServerConfig, handlers *APIHandlers) *Server {
	router := NewRouter()
	handlers.RegisterRoutes(router)

	router.Use(CorsMiddleware)
	router.Use(NewAuthMiddleware(&AuthConfig{APIKey: config.APIKey}))

	return &Server{
		router: router,
		addr:   config.Addr,
	}
}

func (s *Server) Start() error {
	s.httpServer = &http.Server{
		Addr:    s.addr,
		Handler: s.router,
	}

	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop() error {
	if s.httpServer != nil {
		return s.httpServer.Close()
	}
	return nil
}
