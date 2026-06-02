package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	Username  string         `json:"username" gorm:"type:varchar(50);uniqueIndex;not null;comment:'用户名'"`
	Password  string         `json:"-" gorm:"type:varchar(255);not null;comment:'密码（bcrypt加密）'"`
	Email     string         `json:"email" gorm:"type:varchar(100);uniqueIndex;comment:'邮箱'"`
	Phone     string         `json:"phone" gorm:"type:varchar(20);comment:'手机号'"`
	Nickname  string         `json:"nickname" gorm:"type:varchar(50);comment:'昵称'"`
	Avatar    string         `json:"avatar" gorm:"type:varchar(255);comment:'头像URL'"`
	Status    int            `json:"status" gorm:"type:tinyint;default:1;comment:'状态：1-正常，0-禁用'"`
	LastLogin *time.Time     `json:"last_login" gorm:"comment:'最后登录时间'"`
	CreatedAt time.Time      `json:"created_at" gorm:"comment:'创建时间'"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"comment:'更新时间'"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index;comment:'删除时间'"`

	Roles []Role `json:"roles,omitempty" gorm:"many2many:user_roles;"`
}

type Role struct {
	ID          uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string         `json:"name" gorm:"type:varchar(50);uniqueIndex;not null;comment:'角色名称'"`
	Code        string         `json:"code" gorm:"type:varchar(50);uniqueIndex;not null;comment:'角色编码'"`
	Description string         `json:"description" gorm:"type:varchar(200);comment:'描述'"`
	Status      int            `json:"status" gorm:"type:tinyint;default:1;comment:'状态：1-启用，0-禁用'"`
	CreatedAt   time.Time      `json:"created_at" gorm:"comment:'创建时间'"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"comment:'更新时间'"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index;comment:'删除时间'"`
}

type UserRole struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    uint      `json:"user_id" gorm:"not null;index;comment:'用户ID'"`
	RoleID    uint      `json:"role_id" gorm:"not null;index;comment:'角色ID'"`
	CreatedAt time.Time `json:"created_at" gorm:"comment:'创建时间'"`
}

// Permission 权限表
type Permission struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Name      string    `json:"name" gorm:"type:varchar(50);not null;comment:'权限名称'"`
	Code      string    `json:"code" gorm:"type:varchar(100);uniqueIndex;not null;comment:'权限编码'"`
	Type      string    `json:"type" gorm:"type:varchar(20);not null;comment:'类型：menu-菜单，button-按钮'"`
	ParentID  uint      `json:"parent_id" gorm:"default:0;comment:'父级ID'"`
	Path      string    `json:"path" gorm:"type:varchar(200);comment:'路由路径'"`
	Component string    `json:"component" gorm:"type:varchar(100);comment:'组件路径'"`
	Icon      string    `json:"icon" gorm:"type:varchar(50);comment:'图标'"`
	Sort      int       `json:"sort" gorm:"default:0;comment:'排序'"`
	Hidden    bool      `json:"hidden" gorm:"default:false;comment:'是否隐藏：true-隐藏，false-显示'"`
	Status    int       `json:"status" gorm:"type:tinyint;default:1;comment:'状态：1-启用，0-禁用'"`
	CreatedAt time.Time `json:"created_at" gorm:"comment:'创建时间'"`
	UpdatedAt time.Time `json:"updated_at" gorm:"comment:'更新时间'"`
	APIPath   string    `json:"api_path" gorm:"type:varchar(200);comment:'api路由路径'"`
	APIMethod string    `json:"api_method" gorm:"type:varchar(10);comment:'api请求方法'"`
}

// RolePermission 角色权限关联表
type RolePermission struct {
	ID           uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	RoleID       uint      `json:"role_id" gorm:"not null;index;comment:'角色ID'"`
	PermissionID uint      `json:"permission_id" gorm:"not null;index;comment:'权限ID'"`
	CreatedAt    time.Time `json:"created_at" gorm:"comment:'创建时间'"`
}
