package gorm_logger

import (
	"alert-ops/internal/config"
	"context"
	"go.uber.org/zap"
	"gorm.io/gorm/logger"
	"time"
)

type GormZapLogger struct {
	logLevel logger.LogLevel // 日志级别
}

func NewGormZapLogger() logger.Interface {
	lvl := logger.Silent
	if config.Conf.Mode == "dev" || config.Conf.Mode == "debug" {
		lvl = logger.Info // 开发或调试模式输出SQL
	}
	return &GormZapLogger{logLevel: lvl}
}

func (l *GormZapLogger) LogMode(level logger.LogLevel) logger.Interface {
	l.logLevel = level
	return l
}

func (l *GormZapLogger) Info(ctx context.Context, msg string, args ...interface{}) {
	zap.L().Info(msg, zap.Any("args", args))
}

func (l *GormZapLogger) Warn(ctx context.Context, msg string, args ...interface{}) {
	zap.L().Warn(msg, zap.Any("args", args))
}

func (l *GormZapLogger) Error(ctx context.Context, msg string, args ...interface{}) {
	zap.L().Error(msg, zap.Any("args", args))
}

func (l *GormZapLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rows int64), err error) {
	//  Silent 级别不输出任何日志
	if l.logLevel <= logger.Silent {
		return
	}

	sql, rows := fc()
	elapsed := time.Since(begin)

	zap.L().Info("[GORM SQL]",
		zap.String("sql", sql),
		zap.Duration("cost", elapsed),
		zap.Int64("rows", rows),
		zap.Error(err),
	)
}
