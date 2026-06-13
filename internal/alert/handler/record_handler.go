package handler

import (
	"alert-ops/internal/alert/service"
	"alert-ops/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ackRequest 确认告警请求体
type ackRequest struct {
	User string `json:"user" binding:"required"`
}

// noteRequest 处理备注请求体
type noteRequest struct {
	Note string `json:"note" binding:"required"`
}

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

// GetStatsSummary 获取统计汇总数字卡片
// @Summary 告警统计汇总
// @Description 获取今日告警数、firing数、失败数、抑制数
// @Tags 告警统计
// @Produce json
// @Security BearerAuth
// @Param source_id query int false "告警源ID(0=全部)"
// @Param days query int false "统计天数(默认7)"
// @Success 200 {object} response.Response
// @Router /api/v1/alert-records/stats/summary [get]
func (h *RecordHandler) GetStatsSummary(c *gin.Context) {
	sourceID, days := h.parseStatsParams(c)
	data, err := h.engineSvc.GetStatsSummary(sourceID, days)
	if err != nil {
		response.InternalServerError(c, "获取统计汇总失败")
		return
	}
	response.Success(c, data)
}

// GetStatsDailyTrend 获取近N天告警趋势
// @Summary 告警趋势
// @Description 获取近N天每天的 firing/resolved 数量
// @Tags 告警统计
// @Produce json
// @Security BearerAuth
// @Param source_id query int false "告警源ID(0=全部)"
// @Param days query int false "统计天数(默认7)"
// @Param severity query string false "告警级别筛选"
// @Success 200 {object} response.Response
// @Router /api/v1/alert-records/stats/daily-trend [get]
func (h *RecordHandler) GetStatsDailyTrend(c *gin.Context) {
	sourceID, days := h.parseStatsParams(c)
	severity := c.Query("severity")
	data, err := h.engineSvc.GetStatsDailyTrend(sourceID, days, severity)
	if err != nil {
		response.InternalServerError(c, "获取告警趋势失败")
		return
	}
	response.Success(c, data)
}

// GetStatsBySeverity 按告警级别统计
// @Summary 告警级别分布
// @Description 按告警级别统计数量
// @Tags 告警统计
// @Produce json
// @Security BearerAuth
// @Param source_id query int false "告警源ID(0=全部)"
// @Param days query int false "统计天数(默认7)"
// @Success 200 {object} response.Response
// @Router /api/v1/alert-records/stats/by-severity [get]
func (h *RecordHandler) GetStatsBySeverity(c *gin.Context) {
	sourceID, days := h.parseStatsParams(c)
	data, err := h.engineSvc.GetStatsBySeverity(sourceID, days)
	if err != nil {
		response.InternalServerError(c, "获取级别分布失败")
		return
	}
	response.Success(c, data)
}

// GetStatsTopAlerts Top N 告警名称
// @Summary Top告警
// @Description 获取触发次数最多的 Top N 告警
// @Tags 告警统计
// @Produce json
// @Security BearerAuth
// @Param source_id query int false "告警源ID(0=全部)"
// @Param days query int false "统计天数(默认7)"
// @Param limit query int false "Top数量(默认10)"
// @Success 200 {object} response.Response
// @Router /api/v1/alert-records/stats/top-alerts [get]
func (h *RecordHandler) GetStatsTopAlerts(c *gin.Context) {
	sourceID, days := h.parseStatsParams(c)
	limit := 10
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 50 {
			limit = parsed
		}
	}
	data, err := h.engineSvc.GetStatsTopAlerts(sourceID, days, limit)
	if err != nil {
		response.InternalServerError(c, "获取Top告警失败")
		return
	}
	response.Success(c, data)
}

// GetStatsBySendStatus 按发送状态统计
// @Summary 发送状态分布
// @Description 按发送状态统计告警数量
// @Tags 告警统计
// @Produce json
// @Security BearerAuth
// @Param source_id query int false "告警源ID(0=全部)"
// @Param days query int false "统计天数(默认7)"
// @Success 200 {object} response.Response
// @Router /api/v1/alert-records/stats/by-status [get]
func (h *RecordHandler) GetStatsBySendStatus(c *gin.Context) {
	sourceID, days := h.parseStatsParams(c)
	data, err := h.engineSvc.GetStatsBySendStatus(sourceID, days)
	if err != nil {
		response.InternalServerError(c, "获取状态分布失败")
		return
	}
	response.Success(c, data)
}

// AckAlert 确认告警
// @Summary 确认告警
// @Description 对 firing 状态的告警进行人工确认
// @Tags 告警流水
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "告警流水ID"
// @Param body body ackRequest true "确认人信息"
// @Success 200 {object} response.Response
// @Router /api/v1/alert-records/:id/ack [post]
func (h *RecordHandler) AckAlert(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的记录ID")
		return
	}

	var req ackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请提供确认人信息(user)")
		return
	}

	if err := h.engineSvc.AckAlert(uint(id), req.User); err != nil {
		zap.L().Error("确认告警失败", zap.Uint64("record_id", id), zap.Error(err))
		response.InternalServerError(c, "确认告警失败")
		return
	}

	response.Success(c, gin.H{"msg": "已确认"})
}

// UnackAlert 取消确认告警
// @Summary 取消确认
// @Description 取消告警的人工确认状态
// @Tags 告警流水
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "告警流水ID"
// @Success 200 {object} response.Response
// @Router /api/v1/alert-records/:id/unack [post]
func (h *RecordHandler) UnackAlert(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的记录ID")
		return
	}

	if err := h.engineSvc.UnackAlert(uint(id)); err != nil {
		zap.L().Error("取消确认失败", zap.Uint64("record_id", id), zap.Error(err))
		response.InternalServerError(c, "取消确认失败")
		return
	}

	response.Success(c, gin.H{"msg": "已取消确认"})
}

// UpdateAlertNote 更新处理备注
// @Summary 更新处理备注
// @Description 给告警记录添加处理备注
// @Tags 告警流水
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "告警流水ID"
// @Param body body noteRequest true "处理备注"
// @Success 200 {object} response.Response
// @Router /api/v1/alert-records/:id/note [post]
func (h *RecordHandler) UpdateAlertNote(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的记录ID")
		return
	}

	var req noteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请提供处理备注(note)")
		return
	}

	if err := h.engineSvc.UpdateAlertNote(uint(id), req.Note); err != nil {
		zap.L().Error("更新处理备注失败", zap.Uint64("record_id", id), zap.Error(err))
		response.InternalServerError(c, "更新处理备注失败")
		return
	}

	response.Success(c, gin.H{"msg": "备注已保存"})
}

// parseStatsParams 解析统计通用参数
func (h *RecordHandler) parseStatsParams(c *gin.Context) (sourceID uint, days int) {
	days = 7
	if d := c.Query("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 && parsed <= 90 {
			days = parsed
		}
	}
	if sid := c.Query("source_id"); sid != "" {
		if parsed, err := strconv.ParseUint(sid, 10, 32); err == nil {
			sourceID = uint(parsed)
		}
	}
	return
}
