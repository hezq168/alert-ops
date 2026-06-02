package service

import (
	"alert-ops/internal/alert/adapter"
	"alert-ops/internal/alert/repo"
	"alert-ops/internal/model"
	"errors"
	"fmt"
	"strings"
	"time"
)

// TemplateService 模板服务接口（CRUD + 渲染）
type TemplateService interface {
	// CRUD
	Create(template *model.AlertTemplate) error
	GetByID(id uint) (*model.AlertTemplate, error)
	ListBySource(sourceID uint) ([]model.AlertTemplate, error)
	Update(template *model.AlertTemplate) error
	Delete(id uint) error
	// 渲染
	RenderVariables(alert *adapter.NormalizedAlert) map[string]string
	Render(alert *adapter.NormalizedAlert, titleTpl, contentTpl string) (title, content string)
	RenderFromTemplate(alert *adapter.NormalizedAlert, tpl *model.AlertTemplate, aiSuggestion string) (title, content string)
}

type templateService struct {
	templateRepo repo.AlertTemplateRepo
}

func NewTemplateService() TemplateService {
	return &templateService{
		templateRepo: repo.NewAlertTemplateRepo(),
	}
}

// ============================================
// CRUD 操作
// ============================================

// Create 创建模板
func (s *templateService) Create(template *model.AlertTemplate) error {
	return s.templateRepo.Create(template)
}

// GetByID 获取模板详情
func (s *templateService) GetByID(id uint) (*model.AlertTemplate, error) {
	tpl, err := s.templateRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("模板不存在")
	}
	return tpl, nil
}

// ListBySource 根据告警源ID获取模板列表
func (s *templateService) ListBySource(sourceID uint) ([]model.AlertTemplate, error) {
	return s.templateRepo.ListBySource(sourceID)
}

// Update 更新模板
func (s *templateService) Update(template *model.AlertTemplate) error {
	return s.templateRepo.Update(template)
}

// Delete 删除模板
func (s *templateService) Delete(id uint) error {
	return s.templateRepo.Delete(id)
}

// ============================================
// 模板渲染
// ============================================

// RenderVariables 返回可用的模板变量列表
func (s *templateService) RenderVariables(alert *adapter.NormalizedAlert) map[string]string {
	vars := map[string]string{
		"alert_name":   alert.AlertName,   // 告警名称
		"status":       alert.Status,      // 告警状态
		"severity":     alert.Severity,    // 告警级别
		"instance":     alert.Instance,    // 告警实例
		"summary":      alert.Summary,     // 告警摘要
		"description":  alert.Description, // 告警详情
		"starts_at":    alert.StartsAt.Format("2006-01-02 15:04:05"),
		"current_time": time.Now().Format("2006-01-02 15:04:05"),
	}

	// 中文状态映射
	statusMap := map[string]string{
		"firing":   "触发",
		"resolved": "恢复",
	}
	vars["status_cn"] = statusMap[alert.Status]

	// 级别中文 + emoji
	severityMap := map[string]string{
		"critical": "严重",
		"warning":  "警告",
		"info":     "信息",
	}
	if cn, ok := severityMap[alert.Severity]; ok {
		vars["severity_cn"] = cn
	} else {
		vars["severity_cn"] = alert.Severity
	}

	// emoji
	vars["severity_emoji"] = s.severityEmoji(alert.Severity)
	vars["status_emoji"] = s.statusEmoji(alert.Status)

	// 级别颜色（钉钉 <font> 标签用）
	severityColorMap := map[string]string{
		"critical": "#FF0000",
		"warning":  "#FF8C00",
		"info":     "#008000",
	}
	if color, ok := severityColorMap[alert.Severity]; ok {
		vars["severity_color"] = color
	} else {
		vars["severity_color"] = "#0000FF"
	}

	// 状态颜色：触发=红色，恢复=绿色
	if alert.Status == "resolved" {
		vars["status_color"] = "#00C853"
	} else {
		vars["status_color"] = "#FF0000"
	}

	// 企业微信 <font color> 仅支持 info(绿)/comment(灰)/warning(橙红) 三种预设
	// resolved → info(绿)；firing 按 severity：critical→warning(红), warning→comment(灰), info→info(绿)
	if alert.Status == "resolved" {
		vars["wecom_color"] = "info"
	} else {
		wecomColorMap := map[string]string{
			"critical": "warning",
			"warning":  "comment",
			"info":     "info",
		}
		if color, ok := wecomColorMap[alert.Severity]; ok {
			vars["wecom_color"] = color
		} else {
			vars["wecom_color"] = "comment"
		}
	}

	return vars
}

func (s *templateService) severityEmoji(severity string) string {
	switch strings.ToLower(severity) {
	case "critical":
		return "🔴"
	case "warning":
		return "🟡"
	case "info":
		return "🟢"
	default:
		return "🔵"
	}
}

func (s *templateService) statusEmoji(status string) string {
	switch status {
	case "firing":
		return "🔥"
	case "resolved":
		return "✅"
	default:
		return "❓"
	}
}

// Render 渲染消息模板
// contentTpl 中的 {{.变量名}} 会被替换
func (s *templateService) Render(alert *adapter.NormalizedAlert, titleTpl, contentTpl string) (title, content string) {
	vars := s.RenderVariables(alert)

	title = titleTpl
	content = contentTpl

	for k, v := range vars {
		placeholder := fmt.Sprintf("{{.%s}}", k)
		title = strings.ReplaceAll(title, placeholder, v)
		content = strings.ReplaceAll(content, placeholder, v)
	}

	return title, content
}

// RenderFromTemplate 根据数据库模板渲染
func (s *templateService) RenderFromTemplate(alert *adapter.NormalizedAlert, tpl *model.AlertTemplate, aiSuggestion string) (title, content string) {
	title, content = s.Render(alert, tpl.TitleTpl, tpl.ContentTpl)

	// 如果有 AI 建议，追加到内容末尾
	if aiSuggestion != "" {
		content += fmt.Sprintf("\n\n---\n🤖 **AI 分析建议**：\n%s", aiSuggestion)
	}

	return title, content
}

// DefaultFeishuCardTpl 飞书卡片默认模板
const DefaultFeishuCardTpl = `{{.severity_emoji}} {{.status_emoji}} **[{{.status_cn}}] {{.alert_name}}**

**告警级别**：{{.severity_cn}}
**告警实例**：{{.instance}}
**触发时间**：{{.starts_at}}
**摘要**：{{.summary}}

**详情**：{{.description}}`

// DefaultDingtalkTpl 钉钉默认模板（Markdown + <font> 颜色标签，按告警级别着色）
const DefaultDingtalkTpl = "- **告警级别**：<font color=\"{{.severity_color}}\">{{.severity_cn}}</font>\n" +
	"- **告警实例**：{{.instance}}\n" +
	"- **触发时间**：{{.starts_at}}\n" +
	"- **摘要**：{{.summary}}\n\n" +
	"**详情**：\n{{.description}}"

// DefaultWecomTpl 企业微信默认模板（企业微信 markdown <font color> 仅支持 info/comment/warning 三种预设颜色）
// 使用 emoji + status_color 变量（已在 RenderVariables 中映射为企业微信兼容的颜色名）
// 企业微信 markdown 不支持无序列表 - 语法，使用引用块 > 格式
const DefaultWecomTpl = "# {{.status_cn}}：{{.alert_name}}\n" +
	"> 告警级别：<font color=\"{{.wecom_color}}\">{{.severity_cn}}</font>\n" +
	"> 告警实例：{{.instance}}\n" +
	"> 触发时间：{{.starts_at}}\n" +
	"> 摘要：{{.summary}}\n\n" +
	"**详情**：\n{{.description}}"

// DefaultFeishuTextTpl 飞书文本默认模板
const DefaultFeishuTextTpl = `{{.severity_emoji}} {{.status_emoji}} [{{.status_cn}}] {{.alert_name}}
级别: {{.severity_cn}} | 实例: {{.instance}}
时间: {{.starts_at}}
{{.summary}}
---
{{.description}}`

// DefaultWebhookTpl 自定义Webhook默认模板（JSON格式消息）
const DefaultWebhookTpl = `{
  "alert_name": "{{.alert_name}}",
  "status": "{{.status}}",
  "severity": "{{.severity}}",
  "instance": "{{.instance}}",
  "summary": "{{.summary}}",
  "description": "{{.description}}",
  "starts_at": "{{.starts_at}}",
  "current_time": "{{.current_time}}"
}`

// DefaultTitleTpl 默认标题模板
const DefaultTitleTpl = `{{.status_emoji}} [{{.status_cn}}] {{.alert_name}} - {{.severity_cn}}`
