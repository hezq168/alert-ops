package handler

import (
	"alert-ops/internal/alert/service"
	"alert-ops/internal/model"
	"alert-ops/internal/repo"
	"alert-ops/pkg/response"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AIConfigHandler struct {
	aiService service.AIService
}

func NewAIConfigHandler(aiSvc service.AIService) *AIConfigHandler {
	return &AIConfigHandler{aiService: aiSvc}
}

// GetConfig 获取当前 AI 配置
// @Summary 获取AI配置
// @Tags AI配置
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Router /api/v1/ai-config [get]
func (h *AIConfigHandler) GetConfig(c *gin.Context) {
	var cfg model.AIConfig
	result := repo.DB.First(&cfg)
	if result.Error != nil {
		if result.RowsAffected == 0 {
			response.Success(c, nil)
			return
		}
		zap.L().Error("获取AI配置失败", zap.Error(result.Error))
		response.InternalServerError(c, "获取AI配置失败")
		return
	}
	// 隐藏 API Key 部分字符（仅用于前端回显，实际值保持不变）
	displayCfg := cfg
	if len(displayCfg.APIKey) > 8 {
		displayCfg.APIKey = displayCfg.APIKey[:4] + "****" + displayCfg.APIKey[len(displayCfg.APIKey)-4:]
	}
	response.Success(c, displayCfg)
}

// UpdateConfig 更新 AI 配置并热切换 provider
// @Summary 更新AI配置
// @Tags AI配置
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param config body model.AIConfig true "AI配置"
// @Success 200 {object} response.Response
// @Router /api/v1/ai-config [put]
func (h *AIConfigHandler) UpdateConfig(c *gin.Context) {
	var input model.AIConfig
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 基本校验
	if input.Provider == "" {
		response.BadRequest(c, "provider 不能为空")
		return
	}
	if input.APIKey == "" {
		response.BadRequest(c, "api_key 不能为空")
		return
	}
	if input.BaseURL == "" {
		response.BadRequest(c, "base_url 不能为空")
		return
	}
	if input.Model == "" {
		response.BadRequest(c, "model 不能为空")
		return
	}

	// 查找或创建记录（只保留一条，ID=1）
	var cfg model.AIConfig
	result := repo.DB.First(&cfg)
	if result.Error != nil {
		// 不存在则创建
		input.ID = 1
		if err := repo.DB.Create(&input).Error; err != nil {
			zap.L().Error("创建AI配置失败", zap.Error(err))
			response.InternalServerError(c, "保存配置失败")
			return
		}
	} else {
		// 如果 API Key 未修改（仍是打码值），沿用数据库中的真实值
		apiKey := input.APIKey
		if strings.Contains(apiKey, "****") {
			apiKey = cfg.APIKey
		}
		// 更新
		updates := map[string]interface{}{
			"provider": input.Provider,
			"api_key":  apiKey,
			"base_url": input.BaseURL,
			"model":    input.Model,
		}
		if err := repo.DB.Model(&cfg).Updates(updates).Error; err != nil {
			zap.L().Error("更新AI配置失败", zap.Error(err))
			response.InternalServerError(c, "保存配置失败")
			return
		}
	}

	// 热切换 provider（用真实 API Key）
	realAPIKey := input.APIKey
	if strings.Contains(realAPIKey, "****") {
		realAPIKey = cfg.APIKey
	}
	newProvider := service.CreateProvider(input.Provider, realAPIKey, input.BaseURL, input.Model)
	if newProvider != nil {
		h.aiService.SetProvider(newProvider)
		zap.L().Info("AI provider 热切换成功",
			zap.String("provider", input.Provider),
			zap.String("model", input.Model),
		)
	} else {
		zap.L().Warn("AI provider 创建失败，未切换", zap.String("provider", input.Provider))
	}

	response.SuccessWithMessage(c, "AI配置已更新并生效", nil)
}

// TestConnection 测试 AI 连接（使用数据库中已保存的配置）
// @Summary 测试AI连接
// @Tags AI配置
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Router /api/v1/ai-config/test [post]
func (h *AIConfigHandler) TestConnection(c *gin.Context) {
	// 直接从数据库获取已保存的配置
	var cfg model.AIConfig
	if err := repo.DB.First(&cfg).Error; err != nil {
		response.BadRequest(c, "未找到AI配置，请先保存配置后再测试")
		return
	}

	if cfg.Provider == "" || cfg.APIKey == "" || cfg.BaseURL == "" || cfg.Model == "" {
		response.BadRequest(c, "AI配置不完整，请完善后重试")
		return
	}

	p := service.CreateProvider(cfg.Provider, cfg.APIKey, cfg.BaseURL, cfg.Model)
	if p == nil {
		response.InternalServerError(c, "不支持的 AI 提供商: "+cfg.Provider)
		return
	}

	result, err := p.Analyze("ping", "", "请回复pong")
	if err != nil {
		zap.L().Error("AI连接测试失败",
			zap.String("provider", cfg.Provider),
			zap.String("model", cfg.Model),
			zap.Error(err),
		)
		response.InternalServerError(c, "连接测试失败: "+err.Error())
		return
	}

	zap.L().Info("AI连接测试成功",
		zap.String("provider", cfg.Provider),
		zap.String("model", cfg.Model),
		zap.Int("result_len", len(result)),
	)
	response.Success(c, gin.H{
		"success": true,
		"model":   cfg.Model,
		"result":  result,
	})
}
