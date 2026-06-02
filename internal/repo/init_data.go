package repo

import (
	"alert-ops/internal/model"
	"alert-ops/pkg/password"
	"log"
)

func InitDefaultData() {
	// 检查是否已有管理员
	var count int64
	DB.Model(&model.User{}).Count(&count)

	if count > 0 {
		log.Println("数据库已有数据，跳过初始化")
		return
	}

	// ============================================
	// 创建角色（固定 ID）
	// ============================================
	adminRole := model.Role{
		ID:          1,
		Name:        "管理员",
		Code:        "admin",
		Description: "系统管理员，拥有所有权限",
		Status:      1,
	}

	userRole := model.Role{
		ID:          2,
		Name:        "普通用户",
		Code:        "user",
		Description: "普通用户，基本权限",
		Status:      1,
	}

	DB.FirstOrCreate(&adminRole, model.Role{Code: "admin"})
	DB.FirstOrCreate(&userRole, model.Role{Code: "user"})

	// ============================================
	// 创建权限（固定 ID，ParentID 直接写死）
	// ID 优化方案：
	//   1-9:     顶级菜单
	//   100-200: 系统管理（用户/角色/权限）
	// ============================================

	permissions := []model.Permission{
		// ============================================
		// 顶级菜单 (parent_id=0)
		// 排序：1-仪表盘，2-告警管理，3-系统管理
		// ============================================
		{ID: 1, ParentID: 0, Name: "仪表盘", Code: "k8s:dashboard", Type: "menu", Path: "/dashboard", Component: "Dashboard", Icon: "Odometer", Sort: 0, Hidden: false, Status: 1, APIPath: "/clusters/:id/namespaces", APIMethod: "GET"},
		{ID: 3, ParentID: 0, Name: "告警管理", Code: "alert", Type: "menu", Path: "/alert", Component: "Layout", Icon: "Bell", Sort: 1, Hidden: false, Status: 1, APIPath: "", APIMethod: ""},
		{ID: 2, ParentID: 0, Name: "系统管理", Code: "system", Type: "menu", Path: "/system", Component: "Layout", Icon: "Setting", Sort: 2, Hidden: false, Status: 1, APIPath: "", APIMethod: ""},

		// ============================================
		// 告警管理 (parent_id=3)
		// ============================================
		{ID: 200, ParentID: 3, Name: "告警源", Code: "alert:source:list", Type: "menu", Path: "/alert/sources", Component: "alert/AlertSourceList", Icon: "", Sort: 1, Hidden: false, Status: 1, APIPath: "/alert-sources", APIMethod: "GET"},
		{ID: 201, ParentID: 200, Name: "添加告警源", Code: "alert:source:add", Type: "button", Path: "", Component: "", Icon: "", Sort: 1, Hidden: false, Status: 1, APIPath: "/alert-sources", APIMethod: "POST"},
		{ID: 202, ParentID: 200, Name: "编辑告警源", Code: "alert:source:edit", Type: "button", Path: "", Component: "", Icon: "", Sort: 2, Hidden: false, Status: 1, APIPath: "/alert-sources/:id", APIMethod: "PUT"},
		{ID: 203, ParentID: 200, Name: "删除告警源", Code: "alert:source:delete", Type: "button", Path: "", Component: "", Icon: "", Sort: 3, Hidden: false, Status: 1, APIPath: "/alert-sources/:id", APIMethod: "DELETE"},

		{ID: 204, ParentID: 3, Name: "转发规则", Code: "alert:rule:list", Type: "menu", Path: "/alert/rules", Component: "alert/AlertRuleList", Icon: "", Sort: 2, Hidden: false, Status: 1, APIPath: "/alert-rules", APIMethod: "GET"},
		{ID: 205, ParentID: 204, Name: "添加规则", Code: "alert:rule:add", Type: "button", Path: "", Component: "", Icon: "", Sort: 1, Hidden: false, Status: 1, APIPath: "/alert-rules", APIMethod: "POST"},
		{ID: 206, ParentID: 204, Name: "编辑规则", Code: "alert:rule:edit", Type: "button", Path: "", Component: "", Icon: "", Sort: 2, Hidden: false, Status: 1, APIPath: "/alert-rules/:id", APIMethod: "PUT"},
		{ID: 207, ParentID: 204, Name: "删除规则", Code: "alert:rule:delete", Type: "button", Path: "", Component: "", Icon: "", Sort: 3, Hidden: false, Status: 1, APIPath: "/alert-rules/:id", APIMethod: "DELETE"},

		{ID: 208, ParentID: 3, Name: "消息模板", Code: "alert:template:list", Type: "menu", Path: "/alert/templates", Component: "alert/AlertTemplateList", Icon: "", Sort: 3, Hidden: false, Status: 1, APIPath: "/alert-templates", APIMethod: "GET"},
		{ID: 209, ParentID: 208, Name: "添加模板", Code: "alert:template:add", Type: "button", Path: "", Component: "", Icon: "", Sort: 1, Hidden: false, Status: 1, APIPath: "/alert-templates", APIMethod: "POST"},
		{ID: 210, ParentID: 208, Name: "编辑模板", Code: "alert:template:edit", Type: "button", Path: "", Component: "", Icon: "", Sort: 2, Hidden: false, Status: 1, APIPath: "/alert-templates/:id", APIMethod: "PUT"},
		{ID: 211, ParentID: 208, Name: "删除模板", Code: "alert:template:delete", Type: "button", Path: "", Component: "", Icon: "", Sort: 3, Hidden: false, Status: 1, APIPath: "/alert-templates/:id", APIMethod: "DELETE"},

		{ID: 212, ParentID: 3, Name: "发送通道", Code: "alert:channel:list", Type: "menu", Path: "/alert/channels", Component: "alert/AlertChannelList", Icon: "", Sort: 4, Hidden: false, Status: 1, APIPath: "/alert-channels", APIMethod: "GET"},
		{ID: 213, ParentID: 212, Name: "添加通道", Code: "alert:channel:add", Type: "button", Path: "", Component: "", Icon: "", Sort: 1, Hidden: false, Status: 1, APIPath: "/alert-channels", APIMethod: "POST"},
		{ID: 214, ParentID: 212, Name: "编辑通道", Code: "alert:channel:edit", Type: "button", Path: "", Component: "", Icon: "", Sort: 2, Hidden: false, Status: 1, APIPath: "/alert-channels/:id", APIMethod: "PUT"},
		{ID: 215, ParentID: 212, Name: "删除通道", Code: "alert:channel:delete", Type: "button", Path: "", Component: "", Icon: "", Sort: 3, Hidden: false, Status: 1, APIPath: "/alert-channels/:id", APIMethod: "DELETE"},

		{ID: 216, ParentID: 3, Name: "告警流水", Code: "alert:record:list", Type: "menu", Path: "/alert/records", Component: "alert/AlertRecordList", Icon: "", Sort: 5, Hidden: false, Status: 1, APIPath: "/alert-records", APIMethod: "GET"},
		{ID: 217, ParentID: 3, Name: "AI 配置", Code: "alert:ai-config:edit", Type: "menu", Path: "/alert/ai-config", Component: "alert/AIConfig", Icon: "", Sort: 6, Hidden: false, Status: 1, APIPath: "/ai-config", APIMethod: "PUT"},

		// ============================================
		// 系统管理 (parent_id=2)
		// ============================================
		// 用户管理
		{ID: 100, ParentID: 2, Name: "用户管理", Code: "system:user:list", Type: "menu", Path: "/system/users", Component: "system/UserList", Icon: "", Sort: 1, Hidden: false, Status: 1, APIPath: "/users", APIMethod: "GET"},
		{ID: 101, ParentID: 100, Name: "添加用户", Code: "system:user:add", Type: "button", Path: "", Component: "", Icon: "", Sort: 1, Hidden: false, Status: 1, APIPath: "/auth/register", APIMethod: "POST"},
		{ID: 102, ParentID: 100, Name: "编辑用户", Code: "system:user:edit", Type: "button", Path: "", Component: "", Icon: "", Sort: 2, Hidden: false, Status: 1, APIPath: "/users/:id", APIMethod: "PUT"},
		{ID: 103, ParentID: 100, Name: "删除用户", Code: "system:user:delete", Type: "button", Path: "", Component: "", Icon: "", Sort: 3, Hidden: false, Status: 1, APIPath: "/users/:id", APIMethod: "DELETE"},
		{ID: 104, ParentID: 100, Name: "分配角色", Code: "system:user:assign-role", Type: "button", Path: "", Component: "", Icon: "", Sort: 4, Hidden: false, Status: 1, APIPath: "/users/:id/roles", APIMethod: "POST"},
		{ID: 105, ParentID: 100, Name: "禁用/启用用户", Code: "system:user:toggle-status", Type: "button", Path: "", Component: "", Icon: "", Sort: 5, Hidden: false, Status: 1, APIPath: "/users/:id/status", APIMethod: "PUT"},
		{ID: 106, ParentID: 100, Name: "查看用户角色", Code: "system:user:role:list", Type: "button", Path: "", Component: "", Icon: "", Sort: 6, Hidden: false, Status: 1, APIPath: "/users/:id/roles", APIMethod: "GET"},
		{ID: 116, ParentID: 100, Name: "删除分配角色", Code: "system:user:role:delete-role", Type: "button", Path: "", Component: "", Icon: "", Sort: 7, Hidden: false, Status: 1, APIPath: "/users/:id/roles", APIMethod: "DELETE"},

		// 角色管理
		{ID: 107, ParentID: 2, Name: "角色管理", Code: "system:role:list", Type: "menu", Path: "/system/roles", Component: "system/RoleList", Icon: "", Sort: 2, Hidden: false, Status: 1, APIPath: "/roles", APIMethod: "GET"},
		{ID: 108, ParentID: 107, Name: "添加角色", Code: "system:role:add", Type: "button", Path: "", Component: "", Icon: "", Sort: 1, Hidden: false, Status: 1, APIPath: "/roles", APIMethod: "POST"},
		{ID: 109, ParentID: 107, Name: "删除角色", Code: "system:role:delete", Type: "button", Path: "", Component: "", Icon: "", Sort: 2, Hidden: false, Status: 1, APIPath: "/roles/:id", APIMethod: "DELETE"},
		{ID: 110, ParentID: 107, Name: "分配权限", Code: "system:role:assign-permission", Type: "button", Path: "", Component: "", Icon: "", Sort: 3, Hidden: false, Status: 1, APIPath: "/roles/:id/permissions", APIMethod: "POST"},
		{ID: 111, ParentID: 107, Name: "查看角色权限", Code: "system:role:permission:list", Type: "button", Path: "", Component: "", Icon: "", Sort: 4, Hidden: false, Status: 1, APIPath: "/roles/:id/permissions", APIMethod: "GET"},

		// 权限管理
		{ID: 112, ParentID: 2, Name: "权限管理", Code: "system:permission:list", Type: "menu", Path: "/system/permissions", Component: "system/PermissionList", Icon: "", Sort: 3, Hidden: false, Status: 1, APIPath: "/permissions", APIMethod: "GET"},
		{ID: 113, ParentID: 112, Name: "添加权限", Code: "system:permission:add", Type: "button", Path: "", Component: "", Icon: "", Sort: 1, Hidden: false, Status: 1, APIPath: "/permissions", APIMethod: "POST"},
		{ID: 114, ParentID: 112, Name: "编辑权限", Code: "system:permission:edit", Type: "button", Path: "", Component: "", Icon: "", Sort: 2, Hidden: false, Status: 1, APIPath: "/permissions/:id", APIMethod: "PUT"},
		{ID: 115, ParentID: 112, Name: "删除权限", Code: "system:permission:delete", Type: "button", Path: "", Component: "", Icon: "", Sort: 3, Hidden: false, Status: 1, APIPath: "/permissions/:id", APIMethod: "DELETE"},
	}

	// 使用 Save() 插入或更新（Save 会保留 ID）
	for _, perm := range permissions {
		DB.Save(&perm)
	}

	// ============================================
	// 分配权限给角色
	// ============================================

	// 管理员拥有所有权限
	var allPerms []model.Permission
	DB.Find(&allPerms)
	for _, perm := range allPerms {
		var count int64
		DB.Model(&model.RolePermission{}).Where("role_id = ? AND permission_id = ?", adminRole.ID, perm.ID).Count(&count)
		if count == 0 {
			DB.Create(&model.RolePermission{RoleID: adminRole.ID, PermissionID: perm.ID})
		}
	}

	// ============================================
	// 创建默认管理员账号
	// ============================================
	hashedPassword, err := password.HashPassword("admin123")
	if err != nil {
		log.Printf("密码加密失败: %v", err)
		return
	}

	adminUser := model.User{
		Username: "admin",
		Password: hashedPassword,
		Email:    "admin@example.com",
		Nickname: "系统管理员",
		Status:   1,
	}

	if err := DB.Create(&adminUser).Error; err != nil {
		log.Printf("创建管理员账号失败: %v", err)
		return
	}

	// 关联管理员角色
	userRoleRelation := model.UserRole{
		UserID: adminUser.ID,
		RoleID: adminRole.ID,
	}

	if err := DB.Create(&userRoleRelation).Error; err != nil {
		log.Printf("关联管理员角色失败: %v", err)
		return
	}

	log.Println("默认数据初始化成功")
	log.Println("管理员账号: admin")
	log.Println("管理员密码: admin123")
}
