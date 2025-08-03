package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"ahs/internal/service"
	"go.uber.org/zap"
)

// Handler HTTP处理器
type Handler struct {
	workflowService *service.WorkflowService
	logger          *zap.Logger
}

// New 创建新的HTTP处理器
func New(workflowService *service.WorkflowService, logger *zap.Logger) *Handler {
	return &Handler{
		workflowService: workflowService,
		logger:          logger,
	}
}

// Health 健康检查处理器
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "仅支持 GET 请求", http.StatusMethodNotAllowed)
		return
	}

	workflows := h.workflowService.ListWorkflows()
	resp := map[string]interface{}{
		"status":         "ok",
		"version":        "1.0.0",
		"time":           time.Now().Format(time.RFC3339),
		"workflows":      len(workflows),
		"workflow_names": workflows,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("编码健康检查响应失败", zap.Error(err))
		http.Error(w, "内部服务器错误", http.StatusInternalServerError)
	}
}

// ListWorkflows 列出工作流处理器
func (h *Handler) ListWorkflows(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "仅支持 GET 请求", http.StatusMethodNotAllowed)
		return
	}

	workflows := h.workflowService.ListWorkflows()
	resp := map[string]interface{}{
		"workflows": workflows,
		"count":     len(workflows),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("编码工作流列表响应失败", zap.Error(err))
		http.Error(w, "内部服务器错误", http.StatusInternalServerError)
	}
}

// WorkflowInfo 获取工作流信息处理器
func (h *Handler) WorkflowInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "仅支持 GET 请求", http.StatusMethodNotAllowed)
		return
	}

	// 从路径提取工作流名称
	path := strings.TrimPrefix(r.URL.Path, "/api/workflows/")
	if path == "" {
		http.Error(w, "缺少工作流名称", http.StatusBadRequest)
		return
	}

	info, err := h.workflowService.GetWorkflowInfo(path)
	if err != nil {
		if err == service.ErrWorkflowNotFound {
			http.Error(w, "工作流未找到", http.StatusNotFound)
		} else {
			h.logger.Error("获取工作流信息失败", zap.Error(err), zap.String("workflow", path))
			http.Error(w, "获取工作流信息失败", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(info); err != nil {
		h.logger.Error("编码工作流信息响应失败", zap.Error(err))
		http.Error(w, "内部服务器错误", http.StatusInternalServerError)
	}
}

// Execute 执行工作流处理器
func (h *Handler) Execute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "仅支持 POST 请求", http.StatusMethodNotAllowed)
		return
	}

	var req service.WorkflowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("解析请求失败", zap.Error(err))
		http.Error(w, "请求格式错误", http.StatusBadRequest)
		return
	}

	// 执行工作流
	ctx := r.Context()
	if req.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(req.Timeout)*time.Second)
		defer cancel()
	}

	resp, err := h.workflowService.Execute(ctx, req)
	if err != nil {
		switch err {
		case service.ErrWorkflowNotFound:
			http.Error(w, "工作流未找到", http.StatusNotFound)
		case service.ErrInvalidRequest:
			http.Error(w, "无效的请求", http.StatusBadRequest)
		default:
			h.logger.Error("执行工作流失败", zap.Error(err), zap.String("workflow", req.Workflow))
			http.Error(w, "执行工作流失败", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("编码执行响应失败", zap.Error(err))
		http.Error(w, "内部服务器错误", http.StatusInternalServerError)
	}
}

// ExecuteStream 流式执行工作流处理器
func (h *Handler) ExecuteStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "仅支持 POST 请求", http.StatusMethodNotAllowed)
		return
	}

	// 设置流式响应头
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "流式传输不支持", http.StatusInternalServerError)
		return
	}

	var req service.WorkflowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorEvent(w, "请求格式错误")
		flusher.Flush()
		return
	}

	// 执行流式工作流
	ctx := r.Context()
	if req.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(req.Timeout)*time.Second)
		defer cancel()
	}

	err := h.workflowService.ExecuteStream(ctx, req, func(data string, done bool, err error) {
		if err != nil {
			h.sendErrorEvent(w, err.Error())
			flusher.Flush()
			return
		}

		if done {
			h.sendEvent(w, "done", data)
		} else {
			h.sendEvent(w, "data", data)
		}
		flusher.Flush()
	})

	if err != nil {
		switch err {
		case service.ErrWorkflowNotFound:
			h.sendErrorEvent(w, "工作流未找到")
		case service.ErrInvalidRequest:
			h.sendErrorEvent(w, "无效的请求")
		default:
			h.logger.Error("流式执行工作流失败", zap.Error(err), zap.String("workflow", req.Workflow))
			h.sendErrorEvent(w, "执行工作流失败")
		}
		flusher.Flush()
		return
	}
}

// sendEvent 发送 SSE 事件
func (h *Handler) sendEvent(w io.Writer, event, data string) {
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, data)
}

// sendErrorEvent 发送错误事件
func (h *Handler) sendErrorEvent(w io.Writer, message string) {
	errorData, _ := json.Marshal(map[string]string{"error": message})
	fmt.Fprintf(w, "event: error\ndata: %s\n\n", errorData)
}
