package model

import (
	"time"

	"gorm.io/gorm"
)

// AIConfig AI 配置（数据库中只保留一条记录，ID=1）
type AIConfig struct {
	ID        uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	Provider  string         `json:"provider" gorm:"type:varchar(50);not null;default:'openai';comment:'AI提供商: openai/codebuddy'"`
	APIKey    string         `json:"api_key" gorm:"type:varchar(500);not null;default:'';comment:'API Key'"`
	BaseURL   string         `json:"base_url" gorm:"type:varchar(500);not null;default:'';comment:'API 地址'"`
	Model     string         `json:"model" gorm:"type:varchar(100);not null;default:'';comment:'模型名称'"`
	CreatedAt time.Time      `json:"created_at" gorm:"comment:'创建时间'"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"comment:'更新时间'"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index;comment:'删除时间'"`
}
