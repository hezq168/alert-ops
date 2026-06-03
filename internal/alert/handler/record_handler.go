package handler

import (
	"alert-ops/internal/alert/service"
	"alert-ops/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type RecordHandler struct {
	engineSvc service.EngineService
}

func NewRecordHandler(engineSvc service.EngineService) *RecordHandler {
	return &RecordHandler{
		engineSvc: engineSvc,
	}
}

// ListRecords 获取告警流水列表
// @Summary 获取告警流水列表
// @Description 分页查询告警发送流水记录
// @Tags 告警流水
// @Produce json
// @Security BearerAuth
// @Param source_id query int false "告警源ID"
// @Param status query string false "发送状态"
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Success 200 {object} response.Response
// @Router /api/v1/records [get]
func (h *RecordHandler) ListRecords(c *gin.Context) {
	sourceIDStr := c.Query("source_id")
	status := c.Query("status")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	var sourceID uint
	if sourceIDStr != "" {
		sid, err := strconv.ParseUint(sourceIDStr, 10, 32)
		if err != nil {
			response.BadRequest(c, "无效的 source_id")
			return
		}
		sourceID = uint(sid)
	}

	list, total, err := h.engineSvc.ListRecords(sourceID, page, pageSize, status)
	if err != nil {
		response.InternalServerError(c, "获取告警流水失败")
		return
	}
	response.Success(c, gin.H{
		"list":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// AnalyzeRecord AI 分析告警流水
// @Summary AI 分析告警
// @Description 使用 AI 分析指定告警流水记录
// @Tags 告警流水
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "告警流水ID"
// @Success 200 {object} response.Response
// @Router /api/v1/alert-records/:id/analyze [post]
func (h *RecordHandler) AnalyzeRecord(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		zap.L().Warn("AI分析请求：无效的记录ID", zap.String("id", idStr))
		response.BadRequest(c, "无效的记录ID")
		return
	}

	zap.L().Info("收到AI分析请求", zap.Uint64("record_id", id))

	// 获取告警记录详情
	record, err := h.engineSvc.GetRecordByID(uint(id))
	if err != nil {
		zap.L().Error("AI分析：获取告警记录失败", zap.Uint64("record_id", id), zap.Error(err))
		response.InternalServerError(c, "获取告警记录失败")
		return
	}
	if record == nil {
		zap.L().Warn("AI分析：告警记录不存在", zap.Uint64("record_id", id))
		response.NotFound(c, "告警记录不存在")
		return
	}

	zap.L().Info("AI分析：已获取告警记录",
		zap.Uint("record_id", record.ID),
		zap.String("alert_name", record.AlertName),
		zap.String("severity", record.Severity),
	)

	// 调用 AI 分析（自动降级：数据库 → 配置文件）
	result, err := service.AnalyzeWithFallback(record.AlertName, record.Severity, record.Instance, nil, record.Summary, record.Description, "")
	if err != nil {
		zap.L().Error("AI分析失败",
			zap.Uint("record_id", record.ID),
			zap.Error(err),
		)
		response.InternalServerError(c, "AI分析失败: "+err.Error())
		return
	}

	zap.L().Info("AI分析请求完成",
		zap.Uint("record_id", record.ID),
		zap.Int("result_len", len(result)),
	)

	response.Success(c, gin.H{
		"analysis": result,
	})
}
