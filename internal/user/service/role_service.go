package service

import (
	"alert-ops/internal/model"
	"alert-ops/internal/repo"
	userRepo "alert-ops/internal/user/repo"
	"errors"
)

type RoleService interface {
	CreateRole(name, code, description string) (*model.Role, error)
	ListRoles(page, pageSize int) ([]model.Role, int64, error)
	DeleteRole(id uint) error
	AssignRole(userID, roleID uint) error
	GetUserRoles(userID uint) ([]model.Role, error)
	RemoveUserRole(userID, roleID uint) error
	GetRolePermissions(roleID uint) ([]model.Permission, error)
	AssignPermissions(roleID uint, permissionIDs []uint) error
	ListAllPermissions() ([]model.Permission, error)
	CreatePermission(name, code, permType, apiMethod, apiPath string, parentID uint, path, icon string, sort int) (*model.Permission, error)
	DeletePermission(id uint) error
	UpdatePermission(id uint, name, code, permType, apiMethod, apiPath string, parentID uint, path, icon string, sort int) error
}

type roleService struct {
	roleRepo userRepo.RoleRepo
	userRepo userRepo.UserRepo
}

func NewRoleService() RoleService {
	return &roleService{
		roleRepo: userRepo.NewRoleRepo(),
		userRepo: userRepo.NewUserRepo(),
	}
}

func (s *roleService) CreateRole(name, code, description string) (*model.Role, error) {
	// 检查角色编码是否已存在
	existingRole, _ := s.roleRepo.GetByCode(code)
	if existingRole != nil {
		return nil, errors.New("角色编码已存在")
	}

	role := &model.Role{
		Name:        name,
		Code:        code,
		Description: description,
		Status:      1,
	}

	if err := s.roleRepo.Create(role); err != nil {
		return nil, errors.New("创建角色失败")
	}

	return role, nil
}

func (s *roleService) ListRoles(page, pageSize int) ([]model.Role, int64, error) {
	return s.roleRepo.List(page, pageSize)
}

func (s *roleService) AssignRole(userID, roleID uint) error {
	// 检查用户是否存在
	_, err := s.userRepo.GetByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	// 检查角色是否存在
	_, err = s.roleRepo.GetByID(roleID)
	if err != nil {
		return errors.New("角色不存在")
	}

	// 检查是否已分配
	var count int64
	repo.DB.Model(&model.UserRole{}).Where("user_id = ? AND role_id = ?", userID, roleID).Count(&count)
	if count > 0 {
		return errors.New("该用户已拥有此角色")
	}

	// 分配角色
	userRole := model.UserRole{
		UserID: userID,
		RoleID: roleID,
	}

	return repo.DB.Create(&userRole).Error
}

func (s *roleService) GetUserRoles(userID uint) ([]model.Role, error) {
	return s.userRepo.GetUserRoles(userID)
}

func (s *roleService) RemoveUserRole(userID, roleID uint) error {
	// 检查用户是否存在
	_, err := s.userRepo.GetByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	// 检查角色是否存在
	_, err = s.roleRepo.GetByID(roleID)
	if err != nil {
		return errors.New("角色不存在")
	}

	// 删除用户角色关联
	result := repo.DB.Where("user_id = ? AND role_id = ?", userID, roleID).Delete(&model.UserRole{})
	if result.Error != nil {
		return errors.New("移除角色失败")
	}

	if result.RowsAffected == 0 {
		return errors.New("该用户没有此角色")
	}

	return nil
}

func (s *roleService) DeleteRole(id uint) error {
	// 检查是否有用户使用此角色
	var count int64
	repo.DB.Model(&model.UserRole{}).Where("role_id = ?", id).Count(&count)
	if count > 0 {
		return errors.New("该角色正在被使用，无法删除")
	}

	return s.roleRepo.Delete(id)
}

// GetRolePermissions 获取角色的权限列表
func (s *roleService) GetRolePermissions(roleID uint) ([]model.Permission, error) {
	var permissions []model.Permission

	// 查询角色真实权限
	err := repo.DB.Table("permissions").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id = ? AND permissions.status = 1", roleID).
		Order("permissions.sort ASC").
		Find(&permissions).Error

	if err != nil {
		return nil, errors.New("获取角色权限失败")
	}

	return permissions, nil
}

// AssignPermissions 为角色分配权限（先删除旧的，再添加新的）
func (s *roleService) AssignPermissions(roleID uint, permissionIDs []uint) error {
	// 检查角色是否存在
	_, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		return errors.New("角色不存在")
	}

	// 删除该角色的所有权限
	repo.DB.Where("role_id = ?", roleID).Delete(&model.RolePermission{})

	// 去重
	uniqueIDs := make(map[uint]bool)
	for _, permID := range permissionIDs {
		uniqueIDs[permID] = true
	}

	// 添加新权限
	for permID := range uniqueIDs {
		rolePerm := model.RolePermission{
			RoleID:       roleID,
			PermissionID: permID,
		}
		repo.DB.Create(&rolePerm)
	}

	return nil
}

// ListAllPermissions 获取所有权限
func (s *roleService) ListAllPermissions() ([]model.Permission, error) {
	var permissions []model.Permission

	err := repo.DB.Where("status = 1").Order("sort ASC").Find(&permissions).Error
	if err != nil {
		return nil, errors.New("获取权限列表失败")
	}

	return permissions, nil
}

// CreatePermission 创建权限
func (s *roleService) CreatePermission(name, code, permType, apiPath, apiMethod string, parentID uint, path, icon string, sort int) (*model.Permission, error) {
	// 检查权限编码是否已存在
	var count int64
	repo.DB.Model(&model.Permission{}).Where("code = ?", code).Count(&count)
	if count > 0 {
		return nil, errors.New("权限编码已存在")
	}

	permission := &model.Permission{
		Name:      name,
		Code:      code,
		Type:      permType,
		ParentID:  parentID,
		Path:      path,
		Icon:      icon,
		Sort:      sort,
		Status:    1,
		APIPath:   apiPath,
		APIMethod: apiMethod,
	}

	if err := repo.DB.Create(permission).Error; err != nil {
		return nil, errors.New("创建权限失败")
	}

	return permission, nil
}

// DeletePermission 删除权限
func (s *roleService) DeletePermission(id uint) error {
	// 检查是否有角色使用此权限
	var count int64
	repo.DB.Model(&model.RolePermission{}).Where("permission_id = ?", id).Count(&count)
	if count > 0 {
		return errors.New("该权限正在被使用，无法删除")
	}

	if err := repo.DB.Delete(&model.Permission{}, id).Error; err != nil {
		return errors.New("删除权限失败")
	}

	return nil
}

// UpdatePermission 更新权限
func (s *roleService) UpdatePermission(id uint, name, code, permType, apiPath, apiMethod string, parentID uint, path, icon string, sort int) error {
	// 检查权限是否存在
	var permission model.Permission
	if err := repo.DB.First(&permission, id).Error; err != nil {
		return errors.New("权限不存在")
	}

	// 防止循环引用：不能将自己设置为父级
	if parentID == id {
		return errors.New("不能将自己设置为父级权限")
	}

	// 如果修改了编码，检查新编码是否已被其他权限使用
	if permission.Code != code {
		var count int64
		repo.DB.Model(&model.Permission{}).Where("code = ? AND id != ?", code, id).Count(&count)
		if count > 0 {
			return errors.New("权限编码已存在")
		}
	}

	// 更新字段
	permission.Name = name
	permission.Code = code
	permission.Type = permType
	permission.ParentID = parentID
	permission.Path = path
	permission.Icon = icon
	permission.Sort = sort
	permission.APIPath = apiPath
	permission.APIMethod = apiMethod

	if err := repo.DB.Omit("created_at").Save(&permission).Error; err != nil {
		return errors.New("更新权限失败")
	}

	return nil
}
