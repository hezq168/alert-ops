package service

import (
	"alert-ops/internal/alert/repo"
	"alert-ops/internal/model"
	"errors"
)

// SourceService 告警源管理服务接口
type SourceService interface {
	Create(source *model.AlertSource) error
	GetByID(id uint) (*model.AlertSource, error)
	List(page, pageSize int) ([]model.AlertSource, int64, error)
	Update(source *model.AlertSource) error
	Delete(id uint) error
}

type sourceService struct {
	sourceRepo repo.AlertSourceRepo
}

func NewSourceService() SourceService {
	return &sourceService{
		sourceRepo: repo.NewAlertSourceRepo(),
	}
}

// Create 创建告警源
func (s *sourceService) Create(source *model.AlertSource) error {
	if source.Slug == "" {
		return errors.New("slug 不能为空")
	}

	// 检查 slug 是否已存在
	existing, _ := s.sourceRepo.GetBySlug(source.Slug)
	if existing != nil {
		return errors.New("slug 已存在")
	}

	return s.sourceRepo.Create(source)
}

// GetByID 获取告警源详情
func (s *sourceService) GetByID(id uint) (*model.AlertSource, error) {
	source, err := s.sourceRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("告警源不存在")
	}
	return source, nil
}

// List 分页查询告警源列表
func (s *sourceService) List(page, pageSize int) ([]model.AlertSource, int64, error) {
	return s.sourceRepo.List(page, pageSize)
}

// Update 更新告警源
func (s *sourceService) Update(source *model.AlertSource) error {
	// 检查告警源是否存在
	_, err := s.sourceRepo.GetByID(source.ID)
	if err != nil {
		return errors.New("告警源不存在")
	}
	return s.sourceRepo.Update(source)
}

// Delete 删除告警源
func (s *sourceService) Delete(id uint) error {
	return s.sourceRepo.Delete(id)
}
