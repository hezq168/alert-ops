package scheduler

import (
	"alert-ops/internal/alert/service"
	"time"

	"go.uber.org/zap"
)

// SuppressedSender 抑制告警定时发送器
type SuppressedSender struct {
	engineSvc service.EngineService
	interval  time.Duration
	stopCh    chan struct{}
}

func NewSuppressedSender(engineSvc service.EngineService) *SuppressedSender {
	return &SuppressedSender{
		engineSvc: engineSvc,
		interval:  5 * time.Minute, // 每5分钟检查一次
		stopCh:    make(chan struct{}),
	}
}

// Start 启动定时任务
func (s *SuppressedSender) Start() {
	zap.L().Info("抑制告警定时发送任务已启动", zap.Duration("interval", s.interval))

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.checkAndSend()
		case <-s.stopCh:
			zap.L().Info("抑制告警定时发送任务已停止")
			return
		}
	}
}

// Stop 停止定时任务
func (s *SuppressedSender) Stop() {
	close(s.stopCh)
}

// checkAndSend 检查并发送被抑制的告警
func (s *SuppressedSender) checkAndSend() {
	// FlushSuppressedAlerts 内部会根据每条规则的 WorkTimeStart/WorkTimeEnd
	// 判断当前是否在对应规则的工作时间内，非工作时间自动跳过
	if err := s.engineSvc.FlushSuppressedAlerts(); err != nil {
		zap.L().Error("发送抑制告警失败", zap.Error(err))
	}
}

// StartSuppressedSender 启动抑制告警发送器（goroutine）
func StartSuppressedSender(engineSvc service.EngineService) {
	sender := NewSuppressedSender(engineSvc)
	go sender.Start()
}
