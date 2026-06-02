package service

import (
	"alert-ops/internal/alert/repo"
	"alert-ops/internal/model"
	"errors"
)

// RuleService 规则管理服务接口
type RuleService interface {
	Create(rule *model.AlertRule, channelIDs []uint) error
	GetByID(id uint) (*model.AlertRule, error)
	ListBySource(sourceID uint) ([]model.AlertRule, error)
	Update(rule *model.AlertRule, channelIDs []uint) error
	Delete(id uint) error
}

type ruleService struct {
	ruleRepo repo.AlertRuleRepo
}

func NewRuleService() RuleService {
	return &ruleService{
		ruleRepo: repo.NewAlertRuleRepo(),
	}
}

// Create 创建规则
func (s *ruleService) Create(rule *model.AlertRule, channelIDs []uint) error {
	if err := s.ruleRepo.Create(rule); err != nil {
		return errors.New("创建规则失败")
	}

	// 设置通道关联
	if len(channelIDs) > 0 {
		if err := s.ruleRepo.SetChannels(rule.ID, channelIDs); err != nil {
			return errors.New("设置通道关联失败")
		}
	}

	return nil
}

// GetByID 获取规则详情
func (s *ruleService) GetByID(id uint) (*model.AlertRule, error) {
	rule, err := s.ruleRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("规则不存在")
	}
	return rule, nil
}

// ListBySource 根据告警源ID获取规则列表
func (s *ruleService) ListBySource(sourceID uint) ([]model.AlertRule, error) {
	return s.ruleRepo.ListBySource(sourceID)
}

// Update 更新规则
func (s *ruleService) Update(rule *model.AlertRule, channelIDs []uint) error {
	if err := s.ruleRepo.Update(rule); err != nil {
		return errors.New("更新规则失败")
	}

	// 更新通道关联
	if channelIDs != nil {
		if err := s.ruleRepo.SetChannels(rule.ID, channelIDs); err != nil {
			return errors.New("更新通道关联失败")
		}
	}

	return nil
}

// Delete 删除规则
func (s *ruleService) Delete(id uint) error {
	return s.ruleRepo.Delete(id)
}
