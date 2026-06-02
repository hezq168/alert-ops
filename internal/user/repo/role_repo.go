package repo

import (
	"alert-ops/internal/model"
	"alert-ops/internal/repo"
)

type RoleRepo interface {
	Create(role *model.Role) error
	GetByID(id uint) (*model.Role, error)
	GetByCode(code string) (*model.Role, error)
	List(page, pageSize int) ([]model.Role, int64, error)
	Delete(id uint) error
}

type roleRepo struct{}

func NewRoleRepo() RoleRepo {
	return &roleRepo{}
}

func (r *roleRepo) Create(role *model.Role) error {
	return repo.DB.Create(role).Error
}

func (r *roleRepo) GetByID(id uint) (*model.Role, error) {
	var role model.Role
	err := repo.DB.First(&role, id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepo) GetByCode(code string) (*model.Role, error) {
	var role model.Role
	err := repo.DB.Where("code = ?", code).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepo) List(page, pageSize int) ([]model.Role, int64, error) {
	var roles []model.Role
	var total int64

	offset := (page - 1) * pageSize

	repo.DB.Model(&model.Role{}).Where("status = ?", 1).Count(&total)

	err := repo.DB.Where("status = ?", 1).
		Offset(offset).Limit(pageSize).
		Order("created_at DESC").
		Find(&roles).Error

	return roles, total, err
}

func (r *roleRepo) Delete(id uint) error {
	return repo.DB.Delete(&model.Role{}, id).Error
}
