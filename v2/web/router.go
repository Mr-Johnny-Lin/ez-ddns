// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package web

import (
	"encoding/json"
	"net/http"
)

type HandlerFunc func(w http.ResponseWriter, r *http.Request)

type Middleware func(HandlerFunc) HandlerFunc

type Router struct {
	routes map[string]map[string]HandlerFunc
}

func NewRouter() *Router {
	return &Router{
		routes: make(map[string]map[string]HandlerFunc),
	}
}

func (r *Router) Get(path string, handler HandlerFunc) {
	r.addRoute("GET", path, handler)
}

func (r *Router) Post(path string, handler HandlerFunc) {
	r.addRoute("POST", path, handler)
}

func (r *Router) Put(path string, handler HandlerFunc) {
	r.addRoute("PUT", path, handler)
}

func (r *Router) Delete(path string, handler HandlerFunc) {
	r.addRoute("DELETE", path, handler)
}

func (r *Router) addRoute(method, path string, handler HandlerFunc) {
	if r.routes[method] == nil {
		r.routes[method] = make(map[string]HandlerFunc)
	}
	r.routes[method][path] = handler
}

func (r *Router) Use(middleware Middleware) {
	for method, paths := range r.routes {
		for path, handler := range paths {
			r.routes[method][path] = middleware(handler)
		}
	}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	method := req.Method
	path := req.URL.Path

	if routes, ok := r.routes[method]; ok {
		if handler, ok := routes[path]; ok {
			handler(w, req)
			return
		}

		for routePath, handler := range routes {
			if params, matched := matchPath(routePath, path); matched {
				query := req.URL.Query()
				for key, value := range params {
					query.Add(key, value)
				}
				req.URL.RawQuery = query.Encode()
				handler(w, req)
				return
			}
		}
	}

	http.NotFound(w, req)
}

func matchPath(routePath, reqPath string) (map[string]string, bool) {
	routeParts := splitPath(routePath)
	reqParts := splitPath(reqPath)

	if len(routeParts) != len(reqParts) {
		return nil, false
	}

	params := make(map[string]string)
	for i, routePart := range routeParts {
		if len(routePart) > 0 && routePart[0] == ':' {
			params[routePart[1:]] = reqParts[i]
		} else if routePart != reqParts[i] {
			return nil, false
		}
	}

	return params, true
}

func splitPath(path string) []string {
	if path == "/" {
		return []string{}
	}
	parts := []string{}
	start := 1
	for i := 1; i < len(path); i++ {
		if path[i] == '/' {
			parts = append(parts, path[start:i])
			start = i + 1
		}
	}
	if start < len(path) {
		parts = append(parts, path[start:])
	}
	return parts
}

func JSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func JSONError(w http.ResponseWriter, statusCode int, message string) {
	JSONResponse(w, statusCode, map[string]string{"error": message})
}
