// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package web

import (
	"context"
	"encoding/json"
	"net/http"

	"ez-ddns/dao"
	"ez-ddns/model"
	"ez-ddns/utils"
	"ez-ddns/web/dto"
	"ez-ddns/web/service"
)

type APIHandlers struct {
	configService   service.ConfigService
	domainRepo      dao.DomainRepository
	systemRepo      dao.SystemRepository
	ddnsService     DDNSHandler
	autoDDNSService AutoDDNSServiceHandler
}

type DDNSHandler interface {
	HandleConfig(ctx context.Context, configID string)
}

type AutoDDNSServiceHandler interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context)
}

func NewAPIHandlers(configService *service.ConfigService, domainRepo dao.DomainRepository, systemRepo dao.SystemRepository, ddnsService DDNSHandler, autoDDNSService AutoDDNSServiceHandler) *APIHandlers {
	return &APIHandlers{
		configService:   *configService,
		domainRepo:      domainRepo,
		systemRepo:      systemRepo,
		ddnsService:     ddnsService,
		autoDDNSService: autoDDNSService,
	}
}

func (h *APIHandlers) RegisterRoutes(router *Router) {
	router.Post("/api/config", h.CreateConfig)
	router.Get("/api/config/:id", h.GetConfig)
	router.Get("/api/configs", h.ListConfigs)
	router.Put("/api/config/:id", h.UpdateConfig)
	router.Delete("/api/config/:id", h.DeleteConfig)

	router.Post("/api/domain", h.CreateDomain)
	router.Put("/api/domain/:id", h.UpdateDomain)
	router.Delete("/api/domain/:id", h.DeleteDomain)

	router.Get("/api/system", h.GetSystem)
	router.Put("/api/system", h.UpdateSystem)

	router.Post("/api/ddns/:configID", h.TriggerDDNS)

	// AutoDDNS 接口
	router.Post("/api/auto-ddns/start", h.StartAutoDDNS)
	router.Post("/api/auto-ddns/stop", h.StopAutoDDNS)
}

func (h *APIHandlers) CreateConfig(w http.ResponseWriter, r *http.Request) {
	var req dto.ConfigDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "请求体解析失败")
		return
	}

	resp, err := h.configService.CreateConfigWithDomains(r.Context(), &req)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "创建配置失败: "+err.Error())
		return
	}

	JSONResponse(w, http.StatusCreated, resp)
}

func (h *APIHandlers) GetConfig(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/config/"):]
	resp, err := h.configService.GetConfigWithDomains(r.Context(), id)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "读取配置失败: "+err.Error())
		return
	}
	if resp == nil {
		JSONError(w, http.StatusNotFound, "配置不存在")
		return
	}

	resp.AccessKeyId = utils.MaskSecret(resp.AccessKeyId)
	resp.AccessKeySecret = utils.MaskSecret(resp.AccessKeySecret)

	JSONResponse(w, http.StatusOK, resp)
}

func (h *APIHandlers) ListConfigs(w http.ResponseWriter, r *http.Request) {
	resp, err := h.configService.GetAllConfigs(r.Context())
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "读取配置列表失败: "+err.Error())
		return
	}

	for i := range resp {
		resp[i].AccessKeyId = utils.MaskSecret(resp[i].AccessKeyId)
		resp[i].AccessKeySecret = utils.MaskSecret(resp[i].AccessKeySecret)
		resp[i].Domains = nil
	}

	JSONResponse(w, http.StatusOK, resp)
}

func (h *APIHandlers) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/config/"):]
	var req dto.ConfigDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "请求体解析失败")
		return
	}
	req.ID = id

	resp, err := h.configService.UpdateConfigWithDomains(r.Context(), &req)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "更新配置失败: "+err.Error())
		return
	}

	JSONResponse(w, http.StatusOK, resp)
}

func (h *APIHandlers) DeleteConfig(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/config/"):]
	if err := h.configService.DeleteConfigWithDomains(r.Context(), id); err != nil {
		JSONError(w, http.StatusInternalServerError, "删除配置失败: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *APIHandlers) CreateDomain(w http.ResponseWriter, r *http.Request) {
	var domain dto.DomainDTO
	if err := json.NewDecoder(r.Body).Decode(&domain); err != nil {
		JSONError(w, http.StatusBadRequest, "请求体解析失败")
		return
	}

	modelDomain := &model.DomainConfig{
		ID:         domain.ID,
		ConfigID:   domain.ConfigID,
		DomainName: domain.DomainName,
		RR:         domain.RR,
		IPType:     model.ParseIPType(domain.IPType),
	}

	if err := h.domainRepo.Create(r.Context(), modelDomain); err != nil {
		JSONError(w, http.StatusInternalServerError, "创建域名失败: "+err.Error())
		return
	}

	resp := dto.DomainDTO{
		ID:         modelDomain.ID,
		ConfigID:   modelDomain.ConfigID,
		DomainName: modelDomain.DomainName,
		RR:         modelDomain.RR,
		IPType:     modelDomain.IPType.String(),
	}

	JSONResponse(w, http.StatusCreated, resp)
}

func (h *APIHandlers) UpdateDomain(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/domain/"):]
	var domain dto.DomainDTO
	if err := json.NewDecoder(r.Body).Decode(&domain); err != nil {
		JSONError(w, http.StatusBadRequest, "请求体解析失败")
		return
	}
	domain.ID = id

	modelDomain := &model.DomainConfig{
		ID:         domain.ID,
		ConfigID:   domain.ConfigID,
		DomainName: domain.DomainName,
		RR:         domain.RR,
		IPType:     model.ParseIPType(domain.IPType),
	}

	if err := h.domainRepo.Update(r.Context(), modelDomain); err != nil {
		JSONError(w, http.StatusInternalServerError, "更新域名失败: "+err.Error())
		return
	}

	resp := dto.DomainDTO{
		ID:         modelDomain.ID,
		ConfigID:   modelDomain.ConfigID,
		DomainName: modelDomain.DomainName,
		RR:         modelDomain.RR,
		IPType:     modelDomain.IPType.String(),
	}

	JSONResponse(w, http.StatusOK, resp)
}

func (h *APIHandlers) DeleteDomain(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/domain/"):]
	if err := h.domainRepo.Delete(r.Context(), id); err != nil {
		JSONError(w, http.StatusInternalServerError, "删除域名失败: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *APIHandlers) GetSystem(w http.ResponseWriter, r *http.Request) {
	system, err := h.systemRepo.Read(r.Context())
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "读取系统配置失败: "+err.Error())
		return
	}

	systemDTO := &dto.SystemDTO{
		LogLevel: system.LogLevel,
	}

	JSONResponse(w, http.StatusOK, systemDTO)
}

func (h *APIHandlers) UpdateSystem(w http.ResponseWriter, r *http.Request) {
	var systemDTO dto.SystemDTO
	if err := json.NewDecoder(r.Body).Decode(&systemDTO); err != nil {
		JSONError(w, http.StatusBadRequest, "请求体解析失败")
		return
	}

	system := &model.SystemConfig{
		LogLevel: systemDTO.LogLevel,
	}

	if err := h.systemRepo.Update(r.Context(), system); err != nil {
		JSONError(w, http.StatusInternalServerError, "更新系统配置失败: "+err.Error())
		return
	}

	JSONResponse(w, http.StatusOK, systemDTO)
}

func (h *APIHandlers) TriggerDDNS(w http.ResponseWriter, r *http.Request) {
	configID := r.URL.Path[len("/api/ddns/"):]

	logger := utils.LoggerFromContext(r.Context())
	bgCtx := utils.ContextWithLogger(context.Background(), logger)

	go h.ddnsService.HandleConfig(bgCtx, configID)

	JSONResponse(w, http.StatusAccepted, map[string]string{
		"message":  "DDNS 更新任务已触发",
		"configID": configID,
	})
}

// StartAutoDDNS 启动自动DDNS服务
func (h *APIHandlers) StartAutoDDNS(w http.ResponseWriter, r *http.Request) {
	logger := utils.LoggerFromContext(r.Context())
	bgCtx := utils.ContextWithLogger(context.Background(), logger)

	if err := h.autoDDNSService.Start(bgCtx); err != nil {
		JSONError(w, http.StatusInternalServerError, "启动自动DDNS服务失败: "+err.Error())
		return
	}

	JSONResponse(w, http.StatusOK, map[string]string{
		"message": "自动DDNS服务已启动",
	})
}

// StopAutoDDNS 停止自动DDNS服务
func (h *APIHandlers) StopAutoDDNS(w http.ResponseWriter, r *http.Request) {
	h.autoDDNSService.Stop(r.Context())

	JSONResponse(w, http.StatusOK, map[string]string{
		"message": "自动DDNS服务已停止",
	})
}
