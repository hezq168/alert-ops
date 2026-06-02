package model

import (
	"time"

	"gorm.io/gorm"
)

// ============================================
// 告警源实例（一个 slug 一个独立配置单元）
// ============================================
type AlertSource struct {
	ID            uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	Name          string         `json:"name" gorm:"type:varchar(100);not null;comment:'告警源名称'"`
	Slug          string         `json:"slug" gorm:"type:varchar(100);uniqueIndex;not null;comment:'唯一标识(用于webhook URL)'"`
	Type          string         `json:"type" gorm:"type:varchar(50);not null;comment:'告警源类型: alertmanager/aliyun/tencent/aws/zabbix'"`
	Description   string         `json:"description" gorm:"type:varchar(500);comment:'描述'"`
	Enabled       bool           `json:"enabled" gorm:"default:true;comment:'是否启用'"`
	ContinueMatch bool           `json:"continue_match" gorm:"default:false;comment:'匹配到规则后是否继续往下匹配'"`
	Config        string         `json:"config" gorm:"type:text;comment:'额外配置JSON'"`
	CreatedAt     time.Time      `json:"created_at" gorm:"comment:'创建时间'"`
	UpdatedAt     time.Time      `json:"updated_at" gorm:"comment:'更新时间'"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index;comment:'删除时间'"`

	// 关联
	Rules     []AlertRule     `json:"rules,omitempty" gorm:"foreignKey:SourceID"`
	Templates []AlertTemplate `json:"templates,omitempty" gorm:"foreignKey:SourceID"`
	Channels  []AlertChannel  `json:"channels,omitempty" gorm:"foreignKey:SourceID"`
}

// ============================================
// 转发规则（归属于告警源实例）
// ============================================
type AlertRule struct {
	ID          uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	SourceID    uint   `json:"source_id" gorm:"not null;index;comment:'告警源ID'"`
	Name        string `json:"name" gorm:"type:varchar(100);not null;comment:'规则名称'"`
	Description string `json:"description" gorm:"type:varchar(500);comment:'规则描述'"`
	RuleType    string `json:"rule_type" gorm:"type:varchar(20);not null;default:'default';comment:'规则类型: default/time/ai'"`
	Enabled     bool   `json:"enabled" gorm:"default:true;comment:'是否启用'"`
	Priority    int    `json:"priority" gorm:"default:0;comment:'优先级(数字越大越优先)'"`

	// 匹配条件（JSON格式: {"severity":"critical","alertname":"CPUHigh"}）
	MatchLabels string `json:"match_labels" gorm:"type:text;comment:'匹配标签JSON'"`

	// 时间规则相关
	WorkTimeStart    string `json:"work_time_start" gorm:"type:varchar(10);comment:'工作时间开始 HH:mm'"`
	WorkTimeEnd      string `json:"work_time_end" gorm:"type:varchar(10);comment:'工作时间结束 HH:mm'"`
	SuppressOffHours bool   `json:"suppress_off_hours" gorm:"default:false;comment:'是否抑制非工作时间告警'"`

	// AI 规则相关
	AIEnabled        bool   `json:"ai_enabled" gorm:"default:false;comment:'是否启用AI分析'"`
	AIPromptTemplate string `json:"ai_prompt_template" gorm:"type:text;comment:'AI分析提示词模板'"`

	// 关联的模板
	TemplateID *uint `json:"template_id" gorm:"comment:'消息模板ID'"`

	CreatedAt time.Time      `json:"created_at" gorm:"comment:'创建时间'"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"comment:'更新时间'"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index;comment:'删除时间'"`

	// 关联
	Source   AlertSource    `json:"source,omitempty" gorm:"foreignKey:SourceID"`
	Template *AlertTemplate `json:"template,omitempty" gorm:"foreignKey:TemplateID"`
	Channels []AlertChannel `json:"channels,omitempty" gorm:"many2many:rule_channels;joinForeignKey:rule_id;joinReferences:channel_id"`
}

// ============================================
// 规则-通道关联表
// ============================================
type RuleChannel struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	RuleID    uint      `json:"rule_id" gorm:"not null;index;comment:'规则ID'"`
	ChannelID uint      `json:"channel_id" gorm:"not null;index;comment:'通道ID'"`
	CreatedAt time.Time `json:"created_at" gorm:"comment:'创建时间'"`
}

// ============================================
// 消息模板（归属于告警源实例）
// ============================================
type AlertTemplate struct {
	ID          uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	SourceID    uint           `json:"source_id" gorm:"not null;index;comment:'告警源ID'"`
	Name        string         `json:"name" gorm:"type:varchar(100);not null;comment:'模板名称'"`
	ChannelType string         `json:"channel_type" gorm:"type:varchar(20);not null;default:'feishu';comment:'通道类型: feishu/webhook'"`
	Type        string         `json:"type" gorm:"type:varchar(20);not null;default:'card';comment:'消息类型: text/card'"`
	TitleTpl    string         `json:"title_tpl" gorm:"type:varchar(500);comment:'标题模板'"`
	ContentTpl  string         `json:"content_tpl" gorm:"type:text;not null;comment:'内容模板'"`
	Variables   string         `json:"variables" gorm:"type:text;comment:'可用变量列表JSON'"`
	CreatedAt   time.Time      `json:"created_at" gorm:"comment:'创建时间'"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"comment:'更新时间'"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index;comment:'删除时间'"`

	Source AlertSource `json:"source,omitempty" gorm:"foreignKey:SourceID"`
}

// ============================================
// 发送通道（归属于告警源实例）
// ============================================
type AlertChannel struct {
	ID         uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	SourceID   uint           `json:"source_id" gorm:"not null;index;comment:'告警源ID'"`
	Name       string         `json:"name" gorm:"type:varchar(100);not null;comment:'通道名称'"`
	Type       string         `json:"type" gorm:"type:varchar(20);not null;comment:'通道类型: feishu/webhook/dingtalk/wecom/email'"`
	WebhookURL string         `json:"webhook_url" gorm:"type:varchar(500);not null;comment:'Webhook地址'"`
	Secret     string         `json:"secret,omitempty" gorm:"type:varchar(200);comment:'签名密钥'"`
	Config     string         `json:"config" gorm:"type:text;comment:'额外配置JSON'"`
	Enabled    bool           `json:"enabled" gorm:"default:true;comment:'是否启用'"`
	CreatedAt  time.Time      `json:"created_at" gorm:"comment:'创建时间'"`
	UpdatedAt  time.Time      `json:"updated_at" gorm:"comment:'更新时间'"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index;comment:'删除时间'"`

	Source AlertSource `json:"source,omitempty" gorm:"foreignKey:SourceID"`
}

// ============================================
// 告警记录流水
// ============================================
type AlertRecord struct {
	ID               uint       `json:"id" gorm:"primaryKey;autoIncrement"`
	SourceID         uint       `json:"source_id" gorm:"not null;index;comment:'告警源ID'"`
	RuleID           *uint      `json:"rule_id" gorm:"index;comment:'匹配的规则ID'"`
	AlertName        string     `json:"alert_name" gorm:"type:varchar(200);comment:'告警名称'"`
	Status           string     `json:"status" gorm:"type:varchar(20);not null;comment:'告警状态: firing/resolved'"`
	Severity         string     `json:"severity" gorm:"type:varchar(20);comment:'告警级别'"`
	Instance         string     `json:"instance" gorm:"type:varchar(200);comment:'告警实例'"`
	Summary          string     `json:"summary" gorm:"type:text;comment:'告警摘要'"`
	Description      string     `json:"description" gorm:"type:text;comment:'告警描述'"`
	RawData          string     `json:"raw_data" gorm:"type:longtext;comment:'原始告警数据JSON'"`
	FormattedMessage string     `json:"formatted_message" gorm:"type:longtext;comment:'格式化后的消息'"`
	AISuggestion     string     `json:"ai_suggestion" gorm:"type:text;comment:'AI建议'"`
	SendStatus       string     `json:"send_status" gorm:"type:varchar(20);default:'pending';comment:'发送状态: pending/sent/failed/suppressed'"`
	SendError        string     `json:"send_error" gorm:"type:text;comment:'发送失败原因'"`
	SentAt           *time.Time `json:"sent_at" gorm:"comment:'发送时间'"`
	CreatedAt        time.Time  `json:"created_at" gorm:"comment:'创建时间'"`
}

// ============================================
// 被抑制的告警队列（非工作时间抑制，上班统一发送）
// ============================================
type SuppressedAlert struct {
	ID              uint       `json:"id" gorm:"primaryKey;autoIncrement"`
	RecordID        uint       `json:"record_id" gorm:"not null;index;comment:'告警记录ID'"`
	RuleID          uint       `json:"rule_id" gorm:"not null;index;comment:'规则ID'"`
	SourceID        uint       `json:"source_id" gorm:"not null;index;comment:'告警源ID'"`
	SuppressReason  string     `json:"suppress_reason" gorm:"type:varchar(200);comment:'抑制原因'"`
	ScheduledSendAt *time.Time `json:"scheduled_send_at" gorm:"comment:'计划发送时间'"`
	Status          string     `json:"status" gorm:"type:varchar(20);default:'pending';comment:'状态: pending/sent/cancelled'"`
	CreatedAt       time.Time  `json:"created_at" gorm:"comment:'创建时间'"`
	UpdatedAt       time.Time  `json:"updated_at" gorm:"comment:'更新时间'"`
}
