package handler

import (
	"alert-ops/internal/alert/service"
	"alert-ops/pkg/response"
	"errors"
	"io"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type WebhookHandler struct {
	engineSvc service.EngineService
}

func NewWebhookHandler(engineSvc service.EngineService) *WebhookHandler {
	return &WebhookHandler{
		engineSvc: engineSvc,
	}
}

// ReceiveAlertmanager 接收 Alertmanager webhook
// @Summary 接收 Alertmanager Webhook
// @Description 接收 Alertmanager 发送的告警回调
// @Tags Webhook
// @Accept json
// @Produce json
// @Param slug path string true "告警源 slug"
// @Success 200 {object} response.Response
// @Router /api/v1/webhook/alertmanager/{slug} [post]
func (h *WebhookHandler) ReceiveAlertmanager(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		response.BadRequest(c, "slug 不能为空")
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		zap.L().Error("读取 Webhook 请求体失败", zap.String("slug", slug), zap.Error(err))
		response.BadRequest(c, "读取请求体失败")
		return
	}

	zap.L().Info("收到 Webhook 请求",
		zap.String("slug", slug),
		zap.Int("body_size", len(body)),
	)

	// 异步处理：webhook 立即返回 200，实际处理在后台 goroutine 执行
	if err := h.engineSvc.ReceiveAlertmanager(slug, body); err != nil {
		// 告警数量为0属于业务异常，HTTP 200 + 业务码 40001
		if errors.Is(err, service.ErrNoAlerts) {
			zap.L().Warn("Webhook 解析后告警数量为0", zap.String("slug", slug))
			response.BusinessError(c, 40001, "解析后告警数量为0，请检查请求体格式")
			return
		}
		zap.L().Error("Webhook 处理失败", zap.String("slug", slug), zap.Error(err))
		response.Error(c, 500, err.Error())
		return
	}

	response.SuccessWithMessage(c, "ok", nil)
}

// ReceiveWebhook 通用 webhook 接收（预留云告警等扩展）
// @Summary 通用 Webhook 接收
// @Description 接收各类告警源的 Webhook 回调
// @Tags Webhook
// @Accept json
// @Produce json
// @Param type path string true "告警源类型"
// @Param slug path string true "告警源 slug"
// @Success 200 {object} response.Response
// @Router /api/v1/webhook/{type}/{slug} [post]
func (h *WebhookHandler) ReceiveWebhook(c *gin.Context) {
	webhookType := c.Param("type")
	slug := c.Param("slug")

	// 目前只支持 alertmanager
	switch webhookType {
	case "alertmanager":
		h.ReceiveAlertmanager(c)
	default:
		response.BadRequest(c, "不支持的告警源类型: "+webhookType)
	}

	_ = slug // 统一入口暂时用 type 分发
}
