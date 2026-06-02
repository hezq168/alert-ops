package adapter

import "time"

// ============================================
// Alertmanager Webhook 原始数据结构
// ============================================

// AMWebhookPayload Alertmanager webhook 请求体
type AMWebhookPayload struct {
	Receiver          string            `json:"receiver"`
	Status            string            `json:"status"` // firing | resolved
	Alerts            []AMAlert         `json:"alerts"`
	GroupLabels       map[string]string `json:"groupLabels"`
	CommonLabels      map[string]string `json:"commonLabels"`
	CommonAnnotations map[string]string `json:"commonAnnotations"`
	ExternalURL       string            `json:"externalURL"`
	Version           string            `json:"version"`
	GroupKey          string            `json:"groupKey"`
	TruncatedAlerts   int               `json:"truncatedAlerts"`
}

// AMAlert 单条告警
type AMAlert struct {
	Status       string            `json:"status"` // firing | resolved
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     time.Time         `json:"startsAt"`
	EndsAt       time.Time         `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
	Fingerprint  string            `json:"fingerprint"`
}

// ============================================
// 统一内部告警格式（适配后）
// ============================================

// NormalizedAlert 标准化后的告警结构
type NormalizedAlert struct {
	AlertName   string            `json:"alert_name"`
	Status      string            `json:"status"` // firing | resolved
	Severity    string            `json:"severity"`
	Instance    string            `json:"instance"`
	Summary     string            `json:"summary"`
	Description string            `json:"description"`
	StartsAt    time.Time         `json:"starts_at"`
	EndsAt      time.Time         `json:"ends_at"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	RawData     string            `json:"raw_data"` // 原始JSON
}

// ============================================
// 告警源适配器接口（预留扩展）
// ============================================

// AlertAdapter 告警源适配器接口
// 后续添加云告警/Zabbix 时实现此接口即可
type AlertAdapter interface {
	// Parse 解析原始 HTTP Body 为标准化告警列表
	Parse(body []byte) ([]*NormalizedAlert, error)
	// Type 返回适配器类型
	Type() string
}

// ============================================
// Alertmanager 适配器实现
// ============================================

type AlertmanagerAdapter struct{}

func NewAlertmanagerAdapter() *AlertmanagerAdapter {
	return &AlertmanagerAdapter{}
}

func (a *AlertmanagerAdapter) Type() string {
	return "alertmanager"
}

func (a *AlertmanagerAdapter) Parse(body []byte) ([]*NormalizedAlert, error) {
	// 先用 map 解析 JSON（避免 time.Time 解析失败导致全部丢弃）
	var raw map[string]interface{}
	if err := jsonUnmarshal(body, &raw); err != nil {
		return nil, err
	}

	// 用 map 方式逐条提取，容忍个别字段类型不匹配
	alertsRaw, ok := raw["alerts"].([]interface{})
	if !ok {
		return []*NormalizedAlert{}, nil
	}

	var result []*NormalizedAlert
	for _, item := range alertsRaw {
		alertMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		na := convertAMAlert(alertMap)
		if na != nil {
			result = append(result, na)
		}
	}

	return result, nil
}

func convertAMAlert(m map[string]interface{}) *NormalizedAlert {
	na := &NormalizedAlert{}

	// 提取基本字段
	if v, ok := m["status"].(string); ok {
		na.Status = v
	}
	if v, ok := m["startsAt"].(string); ok {
		t, err := time.Parse(time.RFC3339, v)
		if err == nil {
			na.StartsAt = t
		}
	}
	if v, ok := m["endsAt"].(string); ok {
		t, err := time.Parse(time.RFC3339, v)
		if err == nil {
			na.EndsAt = t
		}
	}

	// 提取 labels
	na.Labels = extractStringMap(m, "labels")
	if na.Labels != nil {
		na.AlertName = na.Labels["alertname"]
		na.Severity = na.Labels["severity"]
		na.Instance = na.Labels["instance"]
	}

	// 提取 annotations
	na.Annotations = extractStringMap(m, "annotations")
	if na.Annotations != nil {
		na.Summary = na.Annotations["summary"]
		na.Description = na.Annotations["description"]
	}

	// 保存原始JSON
	rawBytes, _ := jsonMarshal(m)
	na.RawData = string(rawBytes)

	return na
}

func extractStringMap(m map[string]interface{}, key string) map[string]string {
	if v, ok := m[key].(map[string]interface{}); ok {
		result := make(map[string]string)
		for k, val := range v {
			if s, ok := val.(string); ok {
				result[k] = s
			}
		}
		return result
	}
	return nil
}
