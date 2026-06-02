package repo

import (
	"alert-ops/internal/model"
	"alert-ops/internal/repo"
	"time"
)

type UserRepo interface {
	Create(user *model.User) error
	GetByUsername(username string) (*model.User, error)
	GetByID(id uint) (*model.User, error)
	List(page, pageSize int) ([]model.User, int64, error)
	Update(user *model.User) error
	Delete(id uint) error
	UpdateLastLogin(id uint) error
	GetUserRoles(userID uint) ([]model.Role, error)
}

type userRepo struct{}

func NewUserRepo() UserRepo {
	return &userRepo{}
}

// Create 创建用户 - 把新用户保存到数据库
func (r *userRepo) Create(user *model.User) error {
	return repo.DB.Create(user).Error
}

// GetByUsername 根据用户名查找用户 - 登录时用
func (r *userRepo) GetByUsername(username string) (*model.User, error) {
	var user model.User
	err := repo.DB.Where("username = ?", username).Preload("Roles").First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByID 根据ID查找用户 - 获取用户详情时用
func (r *userRepo) GetByID(id uint) (*model.User, error) {
	var user model.User
	err := repo.DB.Where("id = ?", id).Preload("Roles").First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// List 分页查询用户列表 - 后台管理页面用
func (r *userRepo) List(page, pageSize int) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	offset := (page - 1) * pageSize

	repo.DB.Model(&model.User{}).Count(&total)

	err := repo.DB.Offset(offset).Limit(pageSize).Order("created_at DESC").Preload("Roles").Find(&users).Error
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// Update 更新用户信息 - 修改昵称、邮箱等
func (r *userRepo) Update(user *model.User) error {
	return repo.DB.Omit("created_at").Save(user).Error
}

// Delete 删除用户 - 软删除
func (r *userRepo) Delete(id uint) error {
	return repo.DB.Delete(&model.User{}, id).Error
}

// UpdateLastLogin 更新最后登录时间 - 登录成功后调用
func (r *userRepo) UpdateLastLogin(id uint) error {
	now := time.Now()
	return repo.DB.Model(&model.User{}).Where("id = ?", id).Update("last_login", now).Error
}

// GetUserRoles 获取用户的角色列表 - 权限判断时用
func (r *userRepo) GetUserRoles(userID uint) ([]model.Role, error) {
	var roles []model.Role
	err := repo.DB.Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", userID).
		Find(&roles).Error
	return roles, err
}
