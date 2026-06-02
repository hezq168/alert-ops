package handler

import (
	"alert-ops/internal/alert/service"
	"alert-ops/internal/model"
	"alert-ops/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TemplateHandler struct {
	templateService service.TemplateService
}

func NewTemplateHandler() *TemplateHandler {
	return &TemplateHandler{
		templateService: service.NewTemplateService(),
	}
}

// ListTemplates 获取模板列表（按告警源）
// @Summary 获取模板列表
// @Description 根据告警源ID获取模板列表
// @Tags 模板管理
// @Produce json
// @Security BearerAuth
// @Param source_id query int true "告警源ID"
// @Success 200 {object} response.Response
// @Router /api/v1/templates [get]
func (h *TemplateHandler) ListTemplates(c *gin.Context) {
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

	list, err := h.templateService.ListBySource(uint(sourceID))
	if err != nil {
		response.InternalServerError(c, "获取模板列表失败")
		return
	}
	response.Success(c, list)
}

// CreateTemplate 创建模板
// @Summary 创建模板
// @Description 创建新的告警模板
// @Tags 模板管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param data body model.AlertTemplate true "模板信息"
// @Success 200 {object} response.Response
// @Router /api/v1/templates [post]
func (h *TemplateHandler) CreateTemplate(c *gin.Context) {
	var req model.AlertTemplate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	if err := h.templateService.Create(&req); err != nil {
		response.InternalServerError(c, "创建模板失败: "+err.Error())
		return
	}
	response.Success(c, req)
}

// GetTemplate 获取模板详情
// @Summary 获取模板详情
// @Description 根据ID获取模板详情
// @Tags 模板管理
// @Produce json
// @Security BearerAuth
// @Param id path int true "模板ID"
// @Success 200 {object} response.Response
// @Router /api/v1/templates/{id} [get]
func (h *TemplateHandler) GetTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	tpl, err := h.templateService.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, tpl)
}

// UpdateTemplate 更新模板
// @Summary 更新模板
// @Description 更新指定模板的信息
// @Tags 模板管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "模板ID"
// @Param data body model.AlertTemplate true "模板信息"
// @Success 200 {object} response.Response
// @Router /api/v1/templates/{id} [put]
func (h *TemplateHandler) UpdateTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	var req model.AlertTemplate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	req.ID = uint(id)
	if err := h.templateService.Update(&req); err != nil {
		response.InternalServerError(c, "更新模板失败")
		return
	}
	response.Success(c, nil)
}

// DeleteTemplate 删除模板
// @Summary 删除模板
// @Description 删除指定的模板
// @Tags 模板管理
// @Produce json
// @Security BearerAuth
// @Param id path int true "模板ID"
// @Success 200 {object} response.Response
// @Router /api/v1/templates/{id} [delete]
func (h *TemplateHandler) DeleteTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	if err := h.templateService.Delete(uint(id)); err != nil {
		response.InternalServerError(c, "删除模板失败")
		return
	}
	response.Success(c, nil)
}
