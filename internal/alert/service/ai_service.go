package service

import (
	"alert-ops/internal/model"
	"alert-ops/internal/repo"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// ============================================
// AI 分析：查数据库 → 有则调用，没有则直接报错
// ============================================

// AnalyzeWithFallback 调用 AI 分析告警
func AnalyzeWithFallback(alertName, severity, instance string, labels map[string]string, summary, description, customPrompt string) (string, error) {
	var cfg model.AIConfig
	if err := repo.DB.First(&cfg).Error; err == nil && cfg.Provider != "" && cfg.APIKey != "" {
		p := createProvider(cfg.Provider, cfg.APIKey, cfg.BaseURL, cfg.Model)
		zap.L().Info("使用数据库AI配置",
			zap.String("provider", cfg.Provider),
			zap.String("model", cfg.Model),
		)
		return p.Analyze(alertName, severity, instance, labels, summary, description, customPrompt)
	}

	zap.L().Error("AI分析失败：数据库未配置AI，请在管理页面配置 AI 参数")
	return "", fmt.Errorf("AI 未配置，请在管理页面配置 AI 参数")
}

// CreateProvider 根据参数创建 AIProvider（供 handler 测试连接使用）
func CreateProvider(provider, apiKey, baseURL, model string) AIProvider {
	return createProvider(provider, apiKey, baseURL, model)
}

// createProvider 内部创建 provider
func createProvider(provider, apiKey, baseURL, model string) AIProvider {
	switch provider {
	case "openai":
		return &OpenAICompatibleProvider{
			APIKey:    apiKey,
			BaseURL:   baseURL,
			ModelName: model,
		}
	default:
		zap.L().Warn("未知AI提供商，回退到 OpenAI 兼容模式",
			zap.String("provider", provider),
		)
		return &OpenAICompatibleProvider{
			APIKey:    apiKey,
			BaseURL:   baseURL,
			ModelName: model,
		}
	}
}

// ============================================
// AI 提供商接口和实现（保持不变）
// ============================================

// AIProvider AI 提供商接口
type AIProvider interface {
	Analyze(alertName, severity, instance string, labels map[string]string, summary, description, customPrompt string) (string, error)
	Name() string
}

// ============================================
// CodeBuddy AI 提供商（预留实现）
// ============================================

type CodeBuddyProvider struct {
	APIKey    string `json:"api_key"`
	BaseURL   string `json:"base_url"`
	ModelName string `json:"model_name"`
	client    *http.Client
}

func (p *CodeBuddyProvider) Name() string {
	return "codebuddy"
}

func (p *CodeBuddyProvider) Analyze(alertName, severity, instance string, labels map[string]string, summary, description, customPrompt string) (string, error) {
	if p.APIKey == "" {
		return "", fmt.Errorf("CodeBuddy AI 未配置 API Key")
	}

	if p.client == nil {
		p.client = &http.Client{Timeout: 30 * time.Second}
	}

	if p.BaseURL == "" {
		p.BaseURL = "https://copilot.tencent.com"
	}
	if p.ModelName == "" {
		p.ModelName = "codebuddy-default"
	}

	// 构建提示词
	prompt := customPrompt
	if prompt == "" {
		prompt = buildAIPrompt(alertName, severity, instance, labels, summary, description)
	}

	// TODO: 对接 CodeBuddy OAuth2 API，获取 token 后调用 chat completions
	// 当前返回提示信息，表示 AI 分析功能待对接
	zap.L().Warn("CodeBuddy AI 尚未完成对接，请在 ai_service.go 中配置实际的 API 调用")
	return fmt.Sprintf("【AI分析功能待配置】\n请根据以下提示词调用 AI 接口：\n%s", prompt), nil
}

// ============================================
// 预留：OpenAI 兼容提供商
// ============================================

type OpenAICompatibleProvider struct {
	APIKey    string `json:"api_key"`
	BaseURL   string `json:"base_url"`   // 如 https://api.openai.com/v1
	ModelName string `json:"model_name"` // 如 gpt-4, gpt-3.5-turbo
	client    *http.Client
}

func (p *OpenAICompatibleProvider) Name() string {
	return "openai-compatible"
}

func (p *OpenAICompatibleProvider) Analyze(alertName, severity, instance string, labels map[string]string, summary, description, customPrompt string) (string, error) {
	if p.client == nil {
		p.client = &http.Client{Timeout: 30 * time.Second}
	}

	prompt := customPrompt
	if prompt == "" {
		prompt = buildAIPrompt(alertName, severity, instance, labels, summary, description)
	}

	reqBody := map[string]interface{}{
		"model": p.ModelName,
		"messages": []map[string]string{
			{"role": "system", "content": "你是一个专业的运维告警分析助手，请简洁分析告警原因并给出处理建议。"},
			{"role": "user", "content": prompt},
		},
		"max_tokens": 500,
	}

	bodyBytes, _ := json.Marshal(reqBody)
	apiURL := p.BaseURL + "/chat/completions"
	zap.L().Info("调用AI API",
		zap.String("url", apiURL),
		zap.String("model", p.ModelName),
		zap.Int("body_len", len(bodyBytes)),
	)

	req, err := http.NewRequest("POST", apiURL, bytes.NewReader(bodyBytes))
	if err != nil {
		zap.L().Error("创建AI请求失败", zap.Error(err))
		return "", fmt.Errorf("创建AI请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.APIKey)

	startTime := time.Now()
	resp, err := p.client.Do(req)
	elapsed := time.Since(startTime)
	if err != nil {
		zap.L().Error("AI API请求失败",
			zap.String("url", apiURL),
			zap.Duration("elapsed", elapsed),
			zap.Error(err),
		)
		return "", fmt.Errorf("AI请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	zap.L().Info("AI API响应",
		zap.Int("status_code", resp.StatusCode),
		zap.Duration("elapsed", elapsed),
		zap.Int("resp_len", len(respBody)),
	)

	if resp.StatusCode != 200 {
		zap.L().Error("AI API返回非200",
			zap.Int("status_code", resp.StatusCode),
			zap.String("body", string(respBody)),
		)
		return "", fmt.Errorf("AI返回错误 %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		zap.L().Error("AI响应JSON解析失败", zap.Error(err), zap.String("body", string(respBody)))
		return "", fmt.Errorf("AI响应解析失败: %w", err)
	}

	if len(result.Choices) > 0 {
		return result.Choices[0].Message.Content, nil
	}

	zap.L().Warn("AI返回空choices", zap.String("body", string(respBody)))
	return "AI未返回有效内容", nil
}

// buildAIPrompt 构建发送给 AI 的分析提示词
func buildAIPrompt(alertName, severity, instance string, labels map[string]string, summary, description string) string {
	labelsStr := ""
	for k, v := range labels {
		labelsStr += fmt.Sprintf("%s=%s, ", k, v)
	}
	return fmt.Sprintf(
		"请分析以下告警信息：\n"+
			"告警名称: %s\n级别: %s\n实例: %s\n标签: %s\n告警摘要: %s\n告警详情: %s\n\n请用中文回复。给出简洁的根因分析、解决建议（200字以内）尽量精简。",
		alertName, severity, instance, labelsStr, summary, description,
	)
}
