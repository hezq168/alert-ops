package repo

import (
	"alert-ops/internal/config"
	"alert-ops/internal/model"
	"alert-ops/pkg/gorm_logger"
	"fmt"
	"go.uber.org/zap"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDatabase() error {
	var err error

	mysqlConfig := config.Conf.MysqlConfig
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		mysqlConfig.Username,
		mysqlConfig.Password,
		mysqlConfig.Host,
		mysqlConfig.Port,
		mysqlConfig.Dbname,
	)

	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		// 这里替换成 zap 日志
		Logger: gorm_logger.NewGormZapLogger(),
	})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	sqlDB.SetMaxIdleConns(mysqlConfig.MaxIdleConns)
	sqlDB.SetMaxOpenConns(mysqlConfig.MaxOpenConns)
	maxLifetime, err := time.ParseDuration(mysqlConfig.MaxLifetime)
	if err != nil {
		return err
	}
	sqlDB.SetConnMaxLifetime(maxLifetime)

	err = DB.AutoMigrate(
		&model.User{},
		&model.Role{},
		&model.UserRole{},
		&model.RolePermission{},
		&model.Permission{},
		// 告警模块
		&model.AlertSource{},
		&model.AlertRule{},
		&model.RuleChannel{},
		&model.AlertTemplate{},
		&model.AlertChannel{},
		&model.AlertRecord{},
		&model.SuppressedAlert{},
		// AI 配置
		&model.AIConfig{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	// 手动补列（GORM AutoMigrate 在某些情况下可能不会添加所有列）
	migrateColumns()

	zap.L().Info("MySQL database initialized successfully")
	return nil
}

// migrateColumns 手动补列，处理 AutoMigrate 可能遗漏的字段
func migrateColumns() {
	type columnDef struct {
		table string
		name  string
		sql   string
	}
	columns := []columnDef{
		{"alert_rules", "ai_prompt_template", "ALTER TABLE `alert_rules` ADD COLUMN `ai_prompt_template` TEXT COMMENT 'AI分析提示词模板';"},
		{"alert_sources", "continue_match", "ALTER TABLE `alert_sources` ADD COLUMN `continue_match` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '匹配到规则后是否继续往下匹配';"},
	}

	for _, col := range columns {
		if DB.Migrator().HasColumn(col.table, col.name) {
			continue
		}
		if err := DB.Exec(col.sql).Error; err != nil {
			zap.L().Warn("手动添加列失败", zap.String("table", col.table), zap.String("column", col.name), zap.Error(err))
		} else {
			zap.L().Info("手动添加列成功", zap.String("table", col.table), zap.String("column", col.name))
		}
	}

	// 为已有数据库补充 AI 配置菜单权限
	migrateAIConfigMenu()
}

// migrateAIConfigMenu 为已有数据库补充 AI 配置菜单权限
func migrateAIConfigMenu() {
	var count int64
	DB.Model(&model.Permission{}).Where("code = ?", "alert:ai-config:edit").Count(&count)
	if count > 0 {
		return
	}

	aiConfigPerm := model.Permission{
		ID:        217,
		ParentID:  3,
		Name:      "AI 配置",
		Code:      "alert:ai-config:edit",
		Type:      "menu",
		Path:      "/alert/ai-config",
		Component: "alert/AIConfig",
		Icon:      "",
		Sort:      6,
		Hidden:    false,
		Status:    1,
		APIPath:   "/ai-config",
		APIMethod: "PUT",
	}
	if err := DB.Save(&aiConfigPerm).Error; err != nil {
		zap.L().Warn("添加AI配置菜单权限失败", zap.Error(err))
		return
	}

	// 给所有管理员角色分配该权限
	var roles []model.Role
	DB.Where("code = ?", "admin").Find(&roles)
	for _, role := range roles {
		var rpCount int64
		DB.Model(&model.RolePermission{}).Where("role_id = ? AND permission_id = ?", role.ID, 217).Count(&rpCount)
		if rpCount == 0 {
			DB.Create(&model.RolePermission{RoleID: role.ID, PermissionID: 217})
		}
	}

	zap.L().Info("AI配置菜单权限补充完成")
}
