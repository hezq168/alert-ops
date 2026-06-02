package handler

import (
	"alert-ops/internal/alert/service"
	"alert-ops/internal/model"
	"alert-ops/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type RuleHandler struct {
	ruleService service.RuleService
}

func NewRuleHandler() *RuleHandler {
	return &RuleHandler{
		ruleService: service.NewRuleService(),
	}
}

// ListRules 获取规则列表（按告警源）
// @Summary 获取规则列表
// @Description 根据告警源ID获取规则列表
// @Tags 规则管理
// @Produce json
// @Security BearerAuth
// @Param source_id query int true "告警源ID"
// @Success 200 {object} response.Response
// @Router /api/v1/rules [get]
func (h *RuleHandler) ListRules(c *gin.Context) {
	sourceIDStr := c.Query("source_id")
	if sourceIDStr == "" {
		response.BadRequest(c, "source_id 不能为空")
		return
	}
	sourceID, err := strconv.ParseUint(sourceIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的 source_id")
		return
	}

	list, err := h.ruleService.ListBySource(uint(sourceID))
	if err != nil {
		response.InternalServerError(c, "获取规则列表失败")
		return
	}
	response.Success(c, list)
}

// CreateRule 创建规则
// @Summary 创建规则
// @Description 创建新的告警规则
// @Tags 规则管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param data body object true "规则信息"
// @Success 200 {object} response.Response
// @Router /api/v1/rules [post]
func (h *RuleHandler) CreateRule(c *gin.Context) {
	var req struct {
		model.AlertRule
		ChannelIDs []uint `json:"channel_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	if err := h.ruleService.Create(&req.AlertRule, req.ChannelIDs); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, req.AlertRule)
}

// GetRule 获取规则详情
// @Summary 获取规则详情
// @Description 根据ID获取规则详情
// @Tags 规则管理
// @Produce json
// @Security BearerAuth
// @Param id path int true "规则ID"
// @Success 200 {object} response.Response
// @Router /api/v1/rules/{id} [get]
func (h *RuleHandler) GetRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	rule, err := h.ruleService.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, rule)
}

// UpdateRule 更新规则
// @Summary 更新规则
// @Description 更新指定规则的信息
// @Tags 规则管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "规则ID"
// @Param data body object true "规则信息"
// @Success 200 {object} response.Response
// @Router /api/v1/rules/{id} [put]
func (h *RuleHandler) UpdateRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	var req struct {
		model.AlertRule
		ChannelIDs []uint `json:"channel_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	req.AlertRule.ID = uint(id)
	if err := h.ruleService.Update(&req.AlertRule, req.ChannelIDs); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, nil)
}

// DeleteRule 删除规则
// @Summary 删除规则
// @Description 删除指定的规则
// @Tags 规则管理
// @Produce json
// @Security BearerAuth
// @Param id path int true "规则ID"
// @Success 200 {object} response.Response
// @Router /api/v1/rules/{id} [delete]
func (h *RuleHandler) DeleteRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	if err := h.ruleService.Delete(uint(id)); err != nil {
		response.InternalServerError(c, "删除规则失败")
		return
	}
	response.Success(c, nil)
}
