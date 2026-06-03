package repo

import (
	"alert-ops/internal/model"
	Repo "alert-ops/internal/repo"
	"errors"

	"gorm.io/gorm"
)

// ============================================
// 告警源 Repo
// ============================================
type AlertSourceRepo interface {
	Create(s *model.AlertSource) error
	GetByID(id uint) (*model.AlertSource, error)
	GetBySlug(slug string) (*model.AlertSource, error)
	List(page, pageSize int) ([]model.AlertSource, int64, error)
	Update(s *model.AlertSource) error
	Delete(id uint) error
}

type alertSourceRepo struct {
}

func NewAlertSourceRepo() AlertSourceRepo {
	return &alertSourceRepo{}
}

// Create 创建告警源
func (r *alertSourceRepo) Create(s *model.AlertSource) error {
	return Repo.DB.Create(s).Error
}

// GetByID 根据ID获取告警源
func (r *alertSourceRepo) GetByID(id uint) (*model.AlertSource, error) {
	var s model.AlertSource
	err := Repo.DB.First(&s, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// GetBySlug 根据slug获取告警源
func (r *alertSourceRepo) GetBySlug(slug string) (*model.AlertSource, error) {
	var s model.AlertSource
	err := Repo.DB.Where("slug = ?", slug).First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// List 分页查询告警源列表
func (r *alertSourceRepo) List(page, pageSize int) ([]model.AlertSource, int64, error) {
	var list []model.AlertSource
	var total int64
	db := Repo.DB.Model(&model.AlertSource{})
	db.Count(&total)
	err := db.Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&list).Error
	return list, total, err
}

// Update 更新告警源（使用 Updates + map，避免 GORM Save 的零值陷阱）
func (r *alertSourceRepo) Update(s *model.AlertSource) error {
	return Repo.DB.Model(&model.AlertSource{}).Where("id = ?", s.ID).Updates(map[string]interface{}{
		"name":           s.Name,
		"slug":           s.Slug,
		"type":           s.Type,
		"description":    s.Description,
		"enabled":        s.Enabled,
		"continue_match": s.ContinueMatch,
		"config":         s.Config,
	}).Error
}

// Delete 删除告警源
func (r *alertSourceRepo) Delete(id uint) error {
	return Repo.DB.Delete(&model.AlertSource{}, id).Error
}

// ============================================
// 规则 Repo
// ============================================
type AlertRuleRepo interface {
	Create(rule *model.AlertRule) error
	GetByID(id uint) (*model.AlertRule, error)
	ListBySource(sourceID uint) ([]model.AlertRule, error)
	Update(rule *model.AlertRule) error
	Delete(id uint) error
	SetChannels(ruleID uint, channelIDs []uint) error
}

type alertRuleRepo struct {
}

func NewAlertRuleRepo() AlertRuleRepo {
	return &alertRuleRepo{}
}

// Create 创建规则
func (r *alertRuleRepo) Create(rule *model.AlertRule) error {
	return Repo.DB.Create(rule).Error
}

// GetByID 根据ID获取规则
func (r *alertRuleRepo) GetByID(id uint) (*model.AlertRule, error) {
	var rule model.AlertRule
	err := Repo.DB.Preload("Channels").Preload("Template").First(&rule, id).Error
	return &rule, err
}

// ListBySource 根据告警源ID获取规则列表
func (r *alertRuleRepo) ListBySource(sourceID uint) ([]model.AlertRule, error) {
	var rules []model.AlertRule
	err := Repo.DB.Where("source_id = ?", sourceID).
		Preload("Channels").Preload("Template").
		Order("priority DESC, id ASC").Find(&rules).Error
	return rules, err
}

// Update 更新规则（使用 Updates + map，避免 GORM Save 的零值陷阱）
func (r *alertRuleRepo) Update(rule *model.AlertRule) error {
	// 兜底：时间字段截断到 varchar(10)，防止前端传入完整时间戳
	ws := rule.WorkTimeStart
	if len(ws) > 10 {
		ws = ws[:10]
	}
	we := rule.WorkTimeEnd
	if len(we) > 10 {
		we = we[:10]
	}

	return Repo.DB.Model(&model.AlertRule{}).Where("id = ?", rule.ID).Updates(map[string]interface{}{
		"source_id":          rule.SourceID,
		"name":               rule.Name,
		"description":        rule.Description,
		"rule_type":          rule.RuleType,
		"enabled":            rule.Enabled,
		"priority":           rule.Priority,
		"match_labels":       rule.MatchLabels,
		"work_time_start":    ws,
		"work_time_end":      we,
		"suppress_off_hours": rule.SuppressOffHours,
		"ai_enabled":         rule.AIEnabled,
		"ai_prompt_template": rule.AIPromptTemplate,
		"template_id":        rule.TemplateID,
	}).Error
}

// Delete 删除规则
func (r *alertRuleRepo) Delete(id uint) error {
	return Repo.DB.Delete(&model.AlertRule{}, id).Error
}

// SetChannels 设置规则关联的通道
func (r *alertRuleRepo) SetChannels(ruleID uint, channelIDs []uint) error {
	tx := Repo.DB.Begin()
	// 清除旧关联
	if err := tx.Where("rule_id = ?", ruleID).Delete(&model.RuleChannel{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	// 添加新关联
	for _, cid := range channelIDs {
		if err := tx.Create(&model.RuleChannel{RuleID: ruleID, ChannelID: cid}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

// ============================================
// 模板 Repo
// ============================================
type AlertTemplateRepo interface {
	Create(t *model.AlertTemplate) error
	GetByID(id uint) (*model.AlertTemplate, error)
	ListBySource(sourceID uint) ([]model.AlertTemplate, error)
	Update(t *model.AlertTemplate) error
	Delete(id uint) error
}

type alertTemplateRepo struct {
}

func NewAlertTemplateRepo() AlertTemplateRepo {
	return &alertTemplateRepo{}
}

// Create 创建模板
func (r *alertTemplateRepo) Create(t *model.AlertTemplate) error {
	return Repo.DB.Create(t).Error
}

// GetByID 根据ID获取模板
func (r *alertTemplateRepo) GetByID(id uint) (*model.AlertTemplate, error) {
	var t model.AlertTemplate
	err := Repo.DB.First(&t, id).Error
	return &t, err
}

// ListBySource 根据告警源ID获取模板列表
func (r *alertTemplateRepo) ListBySource(sourceID uint) ([]model.AlertTemplate, error) {
	var list []model.AlertTemplate
	err := Repo.DB.Where("source_id = ?", sourceID).Order("id DESC").Find(&list).Error
	return list, err
}

// Update 更新模板（使用 Updates + map，避免 GORM Save 的零值陷阱）
func (r *alertTemplateRepo) Update(t *model.AlertTemplate) error {
	return Repo.DB.Model(&model.AlertTemplate{}).Where("id = ?", t.ID).Updates(map[string]interface{}{
		"source_id":    t.SourceID,
		"name":         t.Name,
		"channel_type": t.ChannelType,
		"type":         t.Type,
		"title_tpl":    t.TitleTpl,
		"content_tpl":  t.ContentTpl,
		"variables":    t.Variables,
	}).Error
}

// Delete 删除模板
func (r *alertTemplateRepo) Delete(id uint) error {
	return Repo.DB.Delete(&model.AlertTemplate{}, id).Error
}

// ============================================
// 通道 Repo
// ============================================
type AlertChannelRepo interface {
	Create(ch *model.AlertChannel) error
	GetByID(id uint) (*model.AlertChannel, error)
	ListBySource(sourceID uint) ([]model.AlertChannel, error)
	Update(ch *model.AlertChannel) error
	Delete(id uint) error
}

type alertChannelRepo struct {
}

func NewAlertChannelRepo() AlertChannelRepo {
	return &alertChannelRepo{}
}

// Create 创建通道
func (r *alertChannelRepo) Create(ch *model.AlertChannel) error {
	return Repo.DB.Create(ch).Error
}

// GetByID 根据ID获取通道
func (r *alertChannelRepo) GetByID(id uint) (*model.AlertChannel, error) {
	var ch model.AlertChannel
	err := Repo.DB.First(&ch, id).Error
	return &ch, err
}

// ListBySource 根据告警源ID获取通道列表
func (r *alertChannelRepo) ListBySource(sourceID uint) ([]model.AlertChannel, error) {
	var list []model.AlertChannel
	err := Repo.DB.Where("source_id = ?", sourceID).Order("id DESC").Find(&list).Error
	return list, err
}

// Update 更新通道（使用 Updates + map，避免 GORM Save 的零值陷阱）
func (r *alertChannelRepo) Update(ch *model.AlertChannel) error {
	updates := map[string]interface{}{
		"source_id":   ch.SourceID,
		"name":        ch.Name,
		"type":        ch.Type,
		"webhook_url": ch.WebhookURL,
		"config":      ch.Config,
		"enabled":     ch.Enabled,
	}
	// 只有填了 secret 才更新，空字符串保留原有密钥
	if ch.Secret != "" {
		updates["secret"] = ch.Secret
	}
	return Repo.DB.Model(&model.AlertChannel{}).Where("id = ?", ch.ID).Updates(updates).Error
}

// Delete 删除通道
func (r *alertChannelRepo) Delete(id uint) error {
	return Repo.DB.Delete(&model.AlertChannel{}, id).Error
}

// ============================================
// 告警记录 Repo
// ============================================
type AlertRecordRepo interface {
	Create(record *model.AlertRecord) error
	GetByID(id uint) (*model.AlertRecord, error)
	Update(record *model.AlertRecord) error
	List(sourceID uint, page, pageSize int, status string) ([]model.AlertRecord, int64, error)
}

type alertRecordRepo struct {
}

func NewAlertRecordRepo() AlertRecordRepo {
	return &alertRecordRepo{}
}

// Create 创建告警记录
func (r *alertRecordRepo) Create(record *model.AlertRecord) error {
	return Repo.DB.Create(record).Error
}

// GetByID 根据ID获取告警记录
func (r *alertRecordRepo) GetByID(id uint) (*model.AlertRecord, error) {
	var record model.AlertRecord
	err := Repo.DB.First(&record, id).Error
	return &record, err
}

// Update 更新告警记录（仅更新发送相关字段，避免零值被跳过）
func (r *alertRecordRepo) Update(record *model.AlertRecord) error {
	return Repo.DB.Model(&model.AlertRecord{}).Where("id = ?", record.ID).Updates(map[string]interface{}{
		"send_status":       record.SendStatus,
		"send_error":        record.SendError,
		"formatted_message": record.FormattedMessage,
		"ai_suggestion":     record.AISuggestion,
		"sent_at":           record.SentAt,
	}).Error
}

// List 分页查询告警记录列表
func (r *alertRecordRepo) List(sourceID uint, page, pageSize int, status string) ([]model.AlertRecord, int64, error) {
	var list []model.AlertRecord
	var total int64
	query := Repo.DB.Model(&model.AlertRecord{})
	if sourceID > 0 {
		query = query.Where("source_id = ?", sourceID)
	}
	if status != "" {
		query = query.Where("send_status = ?", status)
	}
	query.Count(&total)
	err := query.Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&list).Error
	return list, total, err
}

// ============================================
// 抑制告警 Repo
// ============================================
type SuppressedAlertRepo interface {
	Create(sa *model.SuppressedAlert) error
	GetPending() ([]model.SuppressedAlert, error)
	GetByID(id uint) (*model.SuppressedAlert, error)
	UpdateStatus(id uint, status string) error
}

type suppressedAlertRepo struct {
}

func NewSuppressedAlertRepo() SuppressedAlertRepo {
	return &suppressedAlertRepo{}
}

// Create 创建抑制告警记录
func (r *suppressedAlertRepo) Create(sa *model.SuppressedAlert) error {
	return Repo.DB.Create(sa).Error
}

// GetPending 获取待发送的抑制告警列表
func (r *suppressedAlertRepo) GetPending() ([]model.SuppressedAlert, error) {
	var list []model.SuppressedAlert
	err := Repo.DB.Where("status = ?", "pending").
		Order("id ASC").Find(&list).Error
	return list, err
}

// GetByID 根据ID获取抑制告警
func (r *suppressedAlertRepo) GetByID(id uint) (*model.SuppressedAlert, error) {
	var sa model.SuppressedAlert
	err := Repo.DB.First(&sa, id).Error
	return &sa, err
}

// UpdateStatus 更新抑制告警状态
func (r *suppressedAlertRepo) UpdateStatus(id uint, status string) error {
	return Repo.DB.Model(&model.SuppressedAlert{}).Where("id = ?", id).Update("status", status).Error
}
