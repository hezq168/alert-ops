package handler

import (
	"alert-ops/internal/alert/service"
	"alert-ops/internal/model"
	"alert-ops/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SourceHandler struct {
	sourceService service.SourceService
}

func NewSourceHandler() *SourceHandler {
	return &SourceHandler{
		sourceService: service.NewSourceService(),
	}
}

// ListSources 获取告警源列表
// @Summary 获取告警源列表
// @Description 分页获取告警源列表
// @Tags 告警源管理
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Success 200 {object} response.Response
// @Router /api/v1/sources [get]
func (h *SourceHandler) ListSources(c *gin.Context) {
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

	list, total, err := h.sourceService.List(page, pageSize)
	if err != nil {
		response.InternalServerError(c, "获取告警源列表失败")
		return
	}
	response.Success(c, gin.H{
		"list":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// CreateSource 创建告警源
// @Summary 创建告警源
// @Description 创建新的告警源
// @Tags 告警源管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param data body model.AlertSource true "告警源信息"
// @Success 200 {object} response.Response
// @Router /api/v1/sources [post]
func (h *SourceHandler) CreateSource(c *gin.Context) {
	var req model.AlertSource
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	if err := h.sourceService.Create(&req); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, req)
}

// GetSource 获取告警源详情
// @Summary 获取告警源详情
// @Description 根据ID获取告警源详情
// @Tags 告警源管理
// @Produce json
// @Security BearerAuth
// @Param id path int true "告警源ID"
// @Success 200 {object} response.Response
// @Router /api/v1/sources/{id} [get]
func (h *SourceHandler) GetSource(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	source, err := h.sourceService.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, source)
}

// UpdateSource 更新告警源
// @Summary 更新告警源
// @Description 更新指定告警源的信息
// @Tags 告警源管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "告警源ID"
// @Param data body model.AlertSource true "告警源信息"
// @Success 200 {object} response.Response
// @Router /api/v1/sources/{id} [put]
func (h *SourceHandler) UpdateSource(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	var req model.AlertSource
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	req.ID = uint(id)
	if err := h.sourceService.Update(&req); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, nil)
}

// DeleteSource 删除告警源
// @Summary 删除告警源
// @Description 删除指定的告警源
// @Tags 告警源管理
// @Produce json
// @Security BearerAuth
// @Param id path int true "告警源ID"
// @Success 200 {object} response.Response
// @Router /api/v1/sources/{id} [delete]
func (h *SourceHandler) DeleteSource(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	if err := h.sourceService.Delete(uint(id)); err != nil {
		response.InternalServerError(c, "删除告警源失败")
		return
	}
	response.Success(c, nil)
}
