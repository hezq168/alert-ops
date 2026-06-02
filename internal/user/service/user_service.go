package service

import (
	"alert-ops/internal/model"
	dbRepo "alert-ops/internal/repo"
	"alert-ops/internal/user/repo"
	"alert-ops/internal/util"
	"alert-ops/pkg/jwt"
	pwd "alert-ops/pkg/password"
	"errors"
)

type UserService interface {
	Register(username, password, email, phone, nickname string) (*model.User, error)
	Login(username, password string) (string, *model.User, error)
	GetUser(id uint) (*model.User, error)
	ListUsers(page, pageSize int) ([]model.User, int64, error)
	UpdateUser(user *model.User) error
	DeleteUser(id uint) error
	ChangePassword(userID uint, oldPassword, newPassword string) error
	UpdateUserStatus(userID uint, status int) error
	GetUserPermissions(userID uint) ([]model.Permission, error)
}

type userService struct {
	userRepo repo.UserRepo
}

func NewUserService() UserService {
	return &userService{
		userRepo: repo.NewUserRepo(),
	}
}

// Register 注册：检查用户名、加密密码、创建用户
func (s *userService) Register(username, password, email, phone, nickname string) (*model.User, error) {
	// 1. 检查用户名是否已存在（err == nil 说明找到了，即用户名已占用）
	existingUser, _ := s.userRepo.GetByUsername(username)
	if existingUser != nil {
		return nil, errors.New("用户名已存在")
	}

	// 2. 加密密码
	hashedPassword, err := pwd.HashPassword(password)
	if err != nil {
		return nil, errors.New("密码加密失败")
	}

	// 3. 创建用户
	user := &model.User{
		Username: username,
		Password: hashedPassword,
		Email:    email,
		Phone:    phone,
		Nickname: nickname,
		Status:   1,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, errors.New("创建用户失败")
	}

	// 4. 返回用户信息（不包含密码）
	user.Password = ""
	return user, nil
}

// Login 登录：验证用户名密码、生成 Token
func (s *userService) Login(username, password string) (string, *model.User, error) {
	// 1. 查找用户
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		return "", nil, errors.New("用户名或密码错误")
	}

	// 2. 验证密码
	if !pwd.CheckPassword(password, user.Password) {
		return "", nil, errors.New("用户名或密码错误")
	}

	// 3. 检查用户状态
	if user.Status != 1 {
		return "", nil, errors.New("账号已被禁用")
	}

	// 4. 更新最后登录时间
	s.userRepo.UpdateLastLogin(user.ID)

	// 5. 生成 Token
	token, err := jwt.GenerateToken(uint(int64(user.ID)), user.Username)
	if err != nil {
		return "", nil, errors.New("生成 Token 失败")
	}

	// 6. 返回 Token 和用户信息（不包含密码）
	user.Password = ""
	return token, user, nil
}

// GetUser  获取用户详情
func (s *userService) GetUser(id uint) (*model.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("用户不存在")
	}
	user.Password = ""
	return user, nil
}

// ListUsers 分页查询用户列表
func (s *userService) ListUsers(page, pageSize int) ([]model.User, int64, error) {
	users, total, err := s.userRepo.List(page, pageSize)
	if err != nil {
		return nil, 0, errors.New("获取用户列表失败")
	}

	// 移除密码字段
	for i := range users {
		users[i].Password = ""
	}

	return users, total, nil
}

// UpdateUser 更新用户信息
func (s *userService) UpdateUser(user *model.User) error {
	existingUser, err := s.userRepo.GetByID(user.ID)
	if err != nil {
		return errors.New("用户不存在")
	}

	// 如果修改了密码，需要重新加密
	if user.Password != "" {
		hashedPassword, err := pwd.HashPassword(user.Password)
		if err != nil {
			return errors.New("密码加密失败")
		}
		user.Password = hashedPassword
	} else {
		// 保持原密码
		user.Password = existingUser.Password
	}

	return s.userRepo.Update(user)
}

// DeleteUser 删除用户
func (s *userService) DeleteUser(id uint) error {
	return s.userRepo.Delete(id)
}

// ChangePassword 修改密码
func (s *userService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	if !pwd.CheckPassword(oldPassword, user.Password) {
		return errors.New("原密码错误")
	}

	hashedPassword, err := pwd.HashPassword(newPassword)
	if err != nil {
		return errors.New("密码加密失败")
	}

	user.Password = hashedPassword
	return s.userRepo.Update(user)
}

// UpdateUserStatus 更新用户状态
func (s *userService) UpdateUserStatus(userID uint, status int) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	user.Status = status
	user.Roles = nil
	return s.userRepo.Update(user)
}

// GetUserPermissions 获取用户权限列表
func (s *userService) GetUserPermissions(userID uint) ([]model.Permission, error) {
	// 通过用户ID查询角色，再查询角色的权限
	var permissions []model.Permission

	err := dbRepo.DB.Table("permissions").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
		Where("user_roles.user_id = ? AND permissions.status = 1", userID).
		Find(&permissions).Error

	if err != nil {
		return nil, errors.New("获取权限失败")
	}

	// 补父级菜单
	permissions, err = util.AppendParentPermissions(permissions)
	if err != nil {
		return nil, err
	}

	return permissions, nil
}
