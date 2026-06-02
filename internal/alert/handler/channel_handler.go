package handler

import (
	"alert-ops/internal/alert/service"
	"alert-ops/internal/model"
	"alert-ops/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ChannelHandler struct {
	channelService service.ChannelService
}

func NewChannelHandler() *ChannelHandler {
	return &ChannelHandler{
		channelService: service.NewChannelService(),
	}
}

// ListChannels 获取通道列表（按告警源）
// @Summary 获取通道列表
// @Description 根据告警源ID获取通道列表
// @Tags 通道管理
// @Produce json
// @Security BearerAuth
// @Param source_id query int true "告警源ID"
// @Success 200 {object} response.Response
// @Router /api/v1/channels [get]
func (h *ChannelHandler) ListChannels(c *gin.Context) {
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

	list, err := h.channelService.ListBySource(uint(sourceID))
	if err != nil {
		response.InternalServerError(c, "获取通道列表失败")
		return
	}
	response.Success(c, list)
}

// CreateChannel 创建通道
// @Summary 创建通道
// @Description 创建新的告警通道
// @Tags 通道管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param data body model.AlertChannel true "通道信息"
// @Success 200 {object} response.Response
// @Router /api/v1/channels [post]
func (h *ChannelHandler) CreateChannel(c *gin.Context) {
	var req model.AlertChannel
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	if err := h.channelService.Create(&req); err != nil {
		response.InternalServerError(c, "创建通道失败: "+err.Error())
		return
	}
	response.Success(c, req)
}

// GetChannel 获取通道详情
// @Summary 获取通道详情
// @Description 根据ID获取通道详情
// @Tags 通道管理
// @Produce json
// @Security BearerAuth
// @Param id path int true "通道ID"
// @Success 200 {object} response.Response
// @Router /api/v1/channels/{id} [get]
func (h *ChannelHandler) GetChannel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	ch, err := h.channelService.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, ch)
}

// UpdateChannel 更新通道
// @Summary 更新通道
// @Description 更新指定通道的信息
// @Tags 通道管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "通道ID"
// @Param data body model.AlertChannel true "通道信息"
// @Success 200 {object} response.Response
// @Router /api/v1/channels/{id} [put]
func (h *ChannelHandler) UpdateChannel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	var req model.AlertChannel
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	req.ID = uint(id)
	if err := h.channelService.Update(&req); err != nil {
		response.InternalServerError(c, "更新通道失败")
		return
	}
	response.Success(c, nil)
}

// DeleteChannel 删除通道
// @Summary 删除通道
// @Description 删除指定的通道
// @Tags 通道管理
// @Produce json
// @Security BearerAuth
// @Param id path int true "通道ID"
// @Success 200 {object} response.Response
// @Router /api/v1/channels/{id} [delete]
func (h *ChannelHandler) DeleteChannel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	if err := h.channelService.Delete(uint(id)); err != nil {
		response.InternalServerError(c, "删除通道失败")
		return
	}
	response.Success(c, nil)
}

// TestSend 测试通道发送
// @Summary 测试通道发送
// @Description 向指定通道发送一条测试消息
// @Tags 通道管理
// @Produce json
// @Security BearerAuth
// @Param id path int true "通道ID"
// @Success 200 {object} response.Response
// @Router /api/v1/alert-channels/{id}/test [post]
func (h *ChannelHandler) TestSend(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	ch, err := h.channelService.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, "通道不存在")
		return
	}

	result := h.channelService.Send(ch,
		"[测试消息] 告警通道连通性测试",
		"这是一条测试消息，如果您收到此消息，说明通道配置正确，消息发送功能正常。",
		"firing", "info",
	)

	if result.Success {
		response.Success(c, gin.H{
			"success": true,
			"channel": ch.Name,
			"type":    ch.Type,
		})
	} else {
		response.InternalServerError(c, "发送失败: "+result.Error)
	}
}
