package service

import (
	"alert-ops/internal/alert/repo"
	"alert-ops/internal/model"
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

// ChannelService 通道服务接口（CRUD + 发送）
type ChannelService interface {
	// CRUD
	Create(channel *model.AlertChannel) error
	GetByID(id uint) (*model.AlertChannel, error)
	ListBySource(sourceID uint) ([]model.AlertChannel, error)
	Update(channel *model.AlertChannel) error
	Delete(id uint) error
	// 发送
	Send(channel *model.AlertChannel, title, content, status, severity string) *SendResult
	// 带重试的发送（最多重试 maxRetries 次，带退避策略）
	SendWithRetry(channel *model.AlertChannel, title, content, status, severity string) *SendResult
}

type channelService struct {
	channelRepo repo.AlertChannelRepo
	client      *http.Client
}

func NewChannelService() ChannelService {
	return &channelService{
		channelRepo: repo.NewAlertChannelRepo(),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ============================================
// CRUD 操作
// ============================================

// Create 创建通道
func (s *channelService) Create(channel *model.AlertChannel) error {
	return s.channelRepo.Create(channel)
}

// GetByID 获取通道详情
func (s *channelService) GetByID(id uint) (*model.AlertChannel, error) {
	ch, err := s.channelRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("通道不存在")
	}
	return ch, nil
}

// ListBySource 根据告警源ID获取通道列表
func (s *channelService) ListBySource(sourceID uint) ([]model.AlertChannel, error) {
	return s.channelRepo.ListBySource(sourceID)
}

// Update 更新通道
func (s *channelService) Update(channel *model.AlertChannel) error {
	return s.channelRepo.Update(channel)
}

// Delete 删除通道
func (s *channelService) Delete(id uint) error {
	return s.channelRepo.Delete(id)
}

// ============================================
// 发送操作
// ============================================

// SendResult 发送结果
type SendResult struct {
	Success bool
	Error   string
	Channel string
}

// Send 向指定通道发送消息
func (s *channelService) Send(channel *model.AlertChannel, title, content, status, severity string) *SendResult {
	switch channel.Type {
	case "feishu":
		return s.sendFeishu(channel, title, content, status, severity)
	case "webhook":
		return s.sendWebhook(channel, title, content)
	case "dingtalk":
		return s.sendDingtalk(channel, title, content, status)
	case "wecom":
		return s.sendWecom(channel, title, content, status, severity)
	default:
		return &SendResult{
			Success: false,
			Error:   fmt.Sprintf("不支持的通道类型: %s", channel.Type),
			Channel: channel.Name,
		}
	}
}

// SendWithRetry 带重试的发送，最多重试 3 次，带指数退避策略
func (s *channelService) SendWithRetry(channel *model.AlertChannel, title, content, status, severity string) *SendResult {
	const maxRetries = 3
	// 退避间隔：第1次重试等2秒，第2次等5秒，第3次等10秒
	retryBackoff := []time.Duration{2 * time.Second, 5 * time.Second, 10 * time.Second}

	var lastResult *SendResult

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			waitTime := retryBackoff[attempt-1]
			zap.L().Info("重试发送",
				zap.Int("attempt", attempt),
				zap.Int("max_retries", maxRetries),
				zap.String("channel", channel.Name),
				zap.Duration("wait_before_retry", waitTime),
			)
			time.Sleep(waitTime)
		}

		result := s.Send(channel, title, content, status, severity)
		if result.Success {
			if attempt > 0 {
				zap.L().Info("重试发送成功",
					zap.Int("attempt", attempt),
					zap.String("channel", channel.Name),
				)
			}
			return result
		}

		lastResult = result
		zap.L().Warn("发送失败，准备重试",
			zap.Int("attempt", attempt+1),
			zap.Int("max_retries", maxRetries),
			zap.String("channel", channel.Name),
			zap.String("error", result.Error),
		)
	}

	zap.L().Error("所有重试均失败",
		zap.Int("total_attempts", maxRetries+1),
		zap.String("channel", channel.Name),
		zap.String("last_error", lastResult.Error),
	)
	return lastResult
}

// ============================================
// 飞书发送
// ============================================

// sendFeishu 发送飞书机器人消息（支持签名校验）
func (s *channelService) sendFeishu(channel *model.AlertChannel, title, content, status, severity string) *SendResult {
	payload := s.buildFeishuCard(title, content, status, severity)
	return s.postJSON(channel.WebhookURL, channel.Secret, payload, channel.Name)
}

// buildFeishuCard 构建飞书交互式卡片消息
func (s *channelService) buildFeishuCard(title, content, status, severity string) map[string]interface{} {
	// 将内容中的 **bold** 转为飞书支持的格式
	content = convertMarkdownToFeishu(content)

	// 根据状态选颜色
	headerColor := "red" // firing 红色
	if status == "resolved" {
		headerColor = "green" // 恢复绿色
	} else {
		switch strings.ToLower(severity) {
		case "critical":
			headerColor = "red"
		case "warning":
			headerColor = "yellow"
		case "info":
			headerColor = "blue"
		}
	}

	return map[string]interface{}{
		"msg_type": "interactive",
		"card": map[string]interface{}{
			"header": map[string]interface{}{
				"title": map[string]interface{}{
					"tag":     "plain_text",
					"content": title,
				},
				"template": headerColor,
			},
			"elements": []map[string]interface{}{
				{
					"tag":     "markdown",
					"content": content,
				},
				{
					"tag": "hr",
				},
				{
					"tag": "note",
					"elements": []map[string]interface{}{
						{
							"tag":     "plain_text",
							"content": fmt.Sprintf("⏰ 发送时间：%s", time.Now().Format("2006-01-02 15:04:05")),
						},
					},
				},
			},
		},
	}
}

// ============================================
// 自定义 Webhook 发送
// ============================================

func (s *channelService) sendWebhook(channel *model.AlertChannel, title, content string) *SendResult {
	payload := map[string]interface{}{
		"title":     title,
		"content":   content,
		"timestamp": time.Now().Format(time.RFC3339),
	}
	return s.postJSON(channel.WebhookURL, channel.Secret, payload, channel.Name)
}

// ============================================
// 钉钉发送
// ============================================

// sendDingtalk 发送钉钉机器人消息（支持 Markdown 格式 + 签名校验 + @提醒）
func (s *channelService) sendDingtalk(channel *model.AlertChannel, title, content, status string) *SendResult {
	// 从 Config 解析 @ 手机号
	var atMobiles []string
	if channel.Config != "" {
		var cfg map[string]interface{}
		if err := json.Unmarshal([]byte(channel.Config), &cfg); err == nil {
			if mobiles, ok := cfg["at_mobiles"].([]interface{}); ok {
				for _, m := range mobiles {
					if s, ok := m.(string); ok {
						atMobiles = append(atMobiles, s)
					}
				}
			}
		}
	}

	payload := s.buildDingtalkMarkdown(title, content, status, atMobiles)

	// 钉钉签名：timestamp + "\n" + secret 做 HmacSHA256 + Base64
	webhookURL := channel.WebhookURL
	if channel.Secret != "" {
		timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
		sign := genDingtalkSign(timestamp, channel.Secret)
		// 追加 timestamp 和 sign 参数到 URL
		if strings.Contains(webhookURL, "?") {
			webhookURL += "&timestamp=" + timestamp + "&sign=" + sign
		} else {
			webhookURL += "?timestamp=" + timestamp + "&sign=" + sign
		}
		zap.L().Info("钉钉签名已设置",
			zap.String("timestamp", timestamp),
			zap.String("sign_prefix", sign[:16]+"..."),
		)
	}

	return s.postJSON(webhookURL, "", payload, channel.Name)
}

// buildDingtalkMarkdown 构建钉钉 Markdown 消息体（含 @ 提醒）
func (s *channelService) buildDingtalkMarkdown(title, content, status string, atMobiles []string) map[string]interface{} {
	// 将 HTML <br> 转为钉钉支持的换行符 \n
	//content = strings.ReplaceAll(content, "<br>", "\n")
	//content = strings.ReplaceAll(content, "<br/>", "\n")
	//content = strings.ReplaceAll(content, "<br />", "\n")

	// 钉钉 Markdown 格式：msgtype=markdown，正文用 ## 大标题
	// 标题按状态着色：恢复=绿色，触发=红色
	var titleColor string
	if status == "resolved" {
		titleColor = "#00C853"
	} else {
		titleColor = "#FF0000"
	}
	coloredTitle := fmt.Sprintf("<font color=\"%s\">%s</font>", titleColor, title)
	markdownText := fmt.Sprintf("## %s\n\n%s\n\n---\n\n⏰ 发送时间：%s", coloredTitle, content, time.Now().Format("2006-01-02 15:04:05"))

	// 在 markdown 文本末尾追加 @ 标记
	if len(atMobiles) > 0 {
		for _, mobile := range atMobiles {
			markdownText += "\n@" + mobile
		}
	}

	payload := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]interface{}{
			"title": title,
			"text":  markdownText,
		},
	}

	// 钉钉 @ 功能：atMobiles + isAtAll
	if len(atMobiles) > 0 {
		payload["at"] = map[string]interface{}{
			"atMobiles": atMobiles,
			"isAtAll":   false,
		}
	}

	return payload
}

// severityColor 返回告警级别对应的颜色（钉钉 <font> 标签用）
func severityColor(severity string) string {
	switch strings.ToLower(severity) {
	case "critical":
		return "#FF0000"
	case "warning":
		return "#FF8C00"
	case "info":
		return "#008000"
	default:
		return "#0000FF"
	}
}

// genDingtalkSign 生成钉钉签名：HmacSHA256(timestamp + "\n" + secret) -> Base64 -> URLEncode
func genDingtalkSign(timestamp, secret string) string {
	// 钉钉加签：把 timestamp+"\n"+密钥 当做签名字符串，使用 HmacSHA256 计算
	stringToSign := timestamp + "\n" + secret
	// 密钥作为 HMAC 的 key
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(stringToSign))
	data := h.Sum(nil)
	// Base64 encode
	sign := base64.StdEncoding.EncodeToString(data)
	// URL encode（使用标准库）
	sign = urlEncode(sign)
	return sign
}

func urlEncode(s string) string {
	// 使用 net/url 标准库进行 URL 编码
	encoded := url.QueryEscape(s)
	return encoded
}

// ============================================
// 企业微信发送
// ============================================

// sendWecom 发送企业微信群机器人消息（支持 markdown 格式）
func (s *channelService) sendWecom(channel *model.AlertChannel, title, content, status, severity string) *SendResult {
	payload := s.buildWecomMarkdown(title, content, status, severity)
	return s.postJSON(channel.WebhookURL, "", payload, channel.Name)
}

// buildWecomMarkdown 构建企业微信 Markdown 消息体
// 企业微信支持的字体颜色：info=绿色, warning=橙红色, comment=灰色, red=红色
func (s *channelService) buildWecomMarkdown(title, content, status, severity string) map[string]interface{} {
	var titleColor string
	if status == "resolved" {
		titleColor = "info" // 绿色
	} else {
		switch strings.ToLower(severity) {
		case "critical":
			titleColor = "red" // 红色
		case "warning":
			titleColor = "warning" // 橙红色
		default:
			titleColor = "info" // 蓝色（企业微信 info 实际显示为蓝色）
		}
	}

	coloredTitle := fmt.Sprintf("<font color=\"%s\">%s</font>", titleColor, title)
	markdownText := fmt.Sprintf("# %s\n\n%s\n\n---\n\n⏰ 发送时间：%s", coloredTitle, content, time.Now().Format("2006-01-02 15:04:05"))

	return map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]interface{}{
			"content": markdownText,
		},
	}
}

// ============================================
// 通用 HTTP POST
// ============================================

func (s *channelService) postJSON(webhookURL, secret string, payload interface{}, channelName string) *SendResult {
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		zap.L().Error("JSON序列化失败", zap.String("channel", channelName), zap.Error(err))
		return &SendResult{Success: false, Error: fmt.Sprintf("JSON序列化失败: %v", err), Channel: channelName}
	}

	zap.L().Info("准备发送 HTTP 请求",
		zap.String("channel", channelName),
		zap.String("url", webhookURL),
		zap.Bool("has_secret", secret != ""),
		zap.Int("body_size", len(bodyBytes)),
	)

	req, err := http.NewRequest("POST", webhookURL, bytes.NewReader(bodyBytes))
	if err != nil {
		zap.L().Error("创建HTTP请求失败", zap.String("channel", channelName), zap.Error(err))
		return &SendResult{Success: false, Error: fmt.Sprintf("创建请求失败: %v", err), Channel: channelName}
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	// 飞书签名
	if secret != "" {
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		sign := genFeishuSign(timestamp, secret)
		req.Header.Set("timestamp", timestamp)
		req.Header.Set("sign", sign)
		zap.L().Info("飞书签名已设置",
			zap.String("timestamp", timestamp),
			zap.String("sign_prefix", sign[:16]+"..."),
		)
	}

	zap.L().Info("发送 HTTP 请求",
		zap.String("channel", channelName),
		zap.String("url", webhookURL),
		zap.String("method", "POST"),
	)

	resp, err := s.client.Do(req)
	if err != nil {
		zap.L().Error("HTTP请求发送失败",
			zap.String("channel", channelName),
			zap.String("url", webhookURL),
			zap.Error(err),
		)
		return &SendResult{Success: false, Error: fmt.Sprintf("请求失败: %v", err), Channel: channelName}
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	zap.L().Info("HTTP 响应",
		zap.String("channel", channelName),
		zap.Int("status", resp.StatusCode),
		zap.String("body", string(respBody)),
	)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// 解析响应体，检查飞书/Webhook 返回的业务状态码
		var result map[string]interface{}
		if json.Unmarshal(respBody, &result) == nil {
			// 飞书返回格式: {"code": 0, "msg": "success"} 或 {"StatusCode": 0, "StatusMessage": "success"}
			if code, ok := getNumericField(result, "StatusCode"); ok && code != 0 {
				msg := getStringField(result, "StatusMessage")
				zap.L().Error("发送失败（业务状态码异常）",
					zap.String("channel", channelName),
					zap.Float64("StatusCode", code),
					zap.String("StatusMessage", msg),
				)
				return &SendResult{Success: false, Error: fmt.Sprintf("StatusCode=%v, StatusMessage=%s", code, msg), Channel: channelName}
			}
			if code, ok := getNumericField(result, "code"); ok && code != 0 {
				msg := getStringField(result, "msg")
				zap.L().Error("发送失败（业务状态码异常）",
					zap.String("channel", channelName),
					zap.Float64("code", code),
					zap.String("msg", msg),
				)
				return &SendResult{Success: false, Error: fmt.Sprintf("code=%v, msg=%s", code, msg), Channel: channelName}
			}
		}
		zap.L().Info("发送成功", zap.String("channel", channelName))
		return &SendResult{Success: true, Channel: channelName}
	}

	zap.L().Error("发送失败（HTTP状态码异常）",
		zap.String("channel", channelName),
		zap.Int("status", resp.StatusCode),
		zap.String("body", string(respBody)),
	)
	return &SendResult{Success: false, Error: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(respBody)), Channel: channelName}
}

// ============================================
// 飞书签名生成
// ============================================

func genFeishuSign(timestamp, secret string) string {
	msg := timestamp + "\n" + secret
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(msg))
	data := h.Sum(nil)
	return base64.StdEncoding.EncodeToString(data)
}

// ============================================
// 工具函数
// ============================================

func convertMarkdownToFeishu(md string) string {
	// 飞书 markdown 支持基本语法，简单处理
	// 保留 **bold**、换行等
	return md
}

// getNumericField 从 map 中提取数值字段（兼容 int/float64）
func getNumericField(m map[string]interface{}, key string) (float64, bool) {
	v, ok := m[key]
	if !ok {
		return 0, false
	}
	switch val := v.(type) {
	case float64:
		return val, true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case json.Number:
		f, _ := val.Float64()
		return f, true
	default:
		return 0, false
	}
}

// getStringField 从 map 中提取字符串字段
func getStringField(m map[string]interface{}, key string) string {
	v, ok := m[key]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return fmt.Sprintf("%v", v)
	}
	return s
}
