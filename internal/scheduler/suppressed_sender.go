package scheduler

import (
	alertRepo "alert-ops/internal/alert/repo"
	"alert-ops/internal/alert/service"
	"alert-ops/internal/model"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

// SuppressedSender 抑制告警定时发送器（不依赖 EngineService，直接操作 repo + channelSvc）
type SuppressedSender struct {
	suppressedRepo alertRepo.SuppressedAlertRepo
	recordRepo     alertRepo.AlertRecordRepo
	ruleRepo       alertRepo.AlertRuleRepo
	channelSvc     service.ChannelService
	interval       time.Duration
	stopCh         chan struct{}
}

func NewSuppressedSender() *SuppressedSender {
	return &SuppressedSender{
		suppressedRepo: alertRepo.NewSuppressedAlertRepo(),
		recordRepo:     alertRepo.NewAlertRecordRepo(),
		ruleRepo:       alertRepo.NewAlertRuleRepo(),
		channelSvc:     service.NewChannelService(),
		interval:       5 * time.Minute, // 每5分钟检查一次
		stopCh:         make(chan struct{}),
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
			s.flush()
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

// flush 发送所有被抑制的告警
func (s *SuppressedSender) flush() {
	suppressed, err := s.suppressedRepo.GetPending()
	if err != nil {
		zap.L().Error("获取抑制告警失败", zap.Error(err))
		return
	}

	if len(suppressed) == 0 {
		zap.L().Info("无待发送的抑制告警")
		return
	}

	zap.L().Info("开始发送抑制告警", zap.Int("count", len(suppressed)))

	// 1. 按 (RuleID, SourceID) 分组
	groups, err := s.groupByRule(suppressed)
	if err != nil {
		zap.L().Error("分组抑制告警失败", zap.Error(err))
		return
	}

	zap.L().Info("抑制告警分组完成", zap.Int("group_count", len(groups)))

	// 2. 每组发送一张汇总卡片
	now := time.Now()
	for _, g := range groups {
		if err := s.sendSummary(&g, now); err != nil {
			zap.L().Error("发送抑制告警汇总失败",
				zap.Uint("rule_id", g.RuleID),
				zap.Uint("source_id", g.SourceID),
				zap.Error(err),
			)
		}
	}
}

// suppressedGroup 抑制告警分组（按 RuleID + SourceID）
type suppressedGroup struct {
	RuleID   uint
	SourceID uint
	Rule     *model.AlertRule
	Records  []*model.AlertRecord
	SAIDs    []uint // suppressed_alert IDs
}

// alertKey 告警去重键（alertname + instance）
type alertKey struct {
	AlertName string
	Instance  string
}

// alertSummary 去重后的告警摘要
type alertSummary struct {
	AlertName string
	Instance  string
	Severity  string
	Count     int
	FirstTime time.Time
}

// groupByRule 将抑制告警按 (RuleID, SourceID) 分组，组内去重统计
func (s *SuppressedSender) groupByRule(suppressed []model.SuppressedAlert) ([]suppressedGroup, error) {
	type groupKey struct {
		RuleID   uint
		SourceID uint
	}

	groupMap := make(map[groupKey]*suppressedGroup)

	for _, sa := range suppressed {
		key := groupKey{RuleID: sa.RuleID, SourceID: sa.SourceID}

		g, ok := groupMap[key]
		if !ok {
			// 首次出现，获取规则信息
			rule, err := s.ruleRepo.GetByID(sa.RuleID)
			if err != nil {
				zap.L().Error("获取规则失败，跳过该抑制记录",
					zap.Uint("rule_id", sa.RuleID),
					zap.Uint("sa_id", sa.ID),
					zap.Error(err),
				)
				continue
			}

			// 检查当前是否已过规则的工作开始时间，未到则跳过
			if !isAfterWorkStart(rule) {
				zap.L().Info("当前未到规则的工作开始时间，跳过",
					zap.Uint("rule_id", rule.ID),
					zap.String("rule_name", rule.Name),
					zap.String("work_start", rule.WorkTimeStart),
					zap.String("work_end", rule.WorkTimeEnd),
				)
				continue
			}

			g = &suppressedGroup{
				RuleID:   sa.RuleID,
				SourceID: sa.SourceID,
				Rule:     rule,
			}
			groupMap[key] = g
		}

		// 获取告警记录
		record, err := s.recordRepo.GetByID(sa.RecordID)
		if err != nil {
			zap.L().Error("获取告警记录失败，跳过",
				zap.Uint("record_id", sa.RecordID),
				zap.Error(err),
			)
			continue
		}

		g.Records = append(g.Records, record)
		g.SAIDs = append(g.SAIDs, sa.ID)
	}

	var groups []suppressedGroup
	for _, g := range groupMap {
		groups = append(groups, *g)
	}

	return groups, nil
}

// dedupAndSummarize 对组内告警记录按 (alertname, instance) 去重统计
func (s *SuppressedSender) dedupAndSummarize(records []*model.AlertRecord) []alertSummary {
	summaryMap := make(map[alertKey]*alertSummary)

	for _, r := range records {
		key := alertKey{AlertName: r.AlertName, Instance: r.Instance}
		as, ok := summaryMap[key]
		if !ok {
			as = &alertSummary{
				AlertName: r.AlertName,
				Instance:  r.Instance,
				Severity:  r.Severity,
				Count:     1,
				FirstTime: r.CreatedAt,
			}
			summaryMap[key] = as
		} else {
			as.Count++
			if r.CreatedAt.Before(as.FirstTime) {
				as.FirstTime = r.CreatedAt
			}
		}
	}

	var result []alertSummary
	for _, v := range summaryMap {
		result = append(result, *v)
	}
	return result
}

// sendSummary 发送一组抑制告警的汇总卡片
func (s *SuppressedSender) sendSummary(g *suppressedGroup, now time.Time) error {
	// 去重统计
	summaries := s.dedupAndSummarize(g.Records)

	zap.L().Info("发送抑制告警汇总",
		zap.Uint("rule_id", g.RuleID),
		zap.Uint("source_id", g.SourceID),
		zap.Int("total_records", len(g.Records)),
		zap.Int("dedup_count", len(summaries)),
		zap.Int("channel_count", len(g.Rule.Channels)),
	)

	// 构建汇总标题
	suppressReason := "非工作时间"
	if len(g.Records) > 0 {
		// 从第一个 SuppressedAlert 获取抑制原因
		sa, _ := s.suppressedRepo.GetByID(g.SAIDs[0])
		if sa != nil && sa.SuppressReason != "" {
			suppressReason = sa.SuppressReason
		}
	}

	title := fmt.Sprintf("📋 非工作时间告警汇总 - %s", g.Rule.Name)

	// 构建汇总卡片内容
	content := s.buildSummaryContent(suppressReason, summaries, len(g.Records))

	// 通过规则关联的通道发送
	allSent := true
	for _, ch := range g.Rule.Channels {
		if !ch.Enabled {
			continue
		}
		// 汇总卡片的 severity 取最高级别
		highestSev := getHighestSeverity(summaries)
		result := s.channelSvc.SendWithRetry(&ch, title, content, "firing", highestSev)
		if !result.Success {
			allSent = false
			zap.L().Error("汇总卡片发送失败",
				zap.String("channel", ch.Name),
				zap.String("error", result.Error),
			)
		}
	}

	// 更新所有关联的记录和抑制记录状态
	if allSent {
		for _, r := range g.Records {
			r.SendStatus = "sent"
			r.SentAt = &now
			s.recordRepo.Update(r)
		}
		for _, id := range g.SAIDs {
			s.suppressedRepo.UpdateStatus(id, "sent")
		}
	}

	return nil
}

// buildSummaryContent 构建汇总卡片内容
func (s *SuppressedSender) buildSummaryContent(suppressReason string, summaries []alertSummary, totalCount int) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("⏱ **抑制原因**：%s\n", suppressReason))
	sb.WriteString(fmt.Sprintf("📊 **共收到 %d 条抑制告警**，去重后 %d 条：\n\n", totalCount, len(summaries)))

	for _, as := range summaries {
		emoji := severityEmojiStr(as.Severity)
		sb.WriteString(fmt.Sprintf("%s **%s** - %s（触发 %d 次，首次 %s）\n",
			emoji,
			as.AlertName,
			as.Instance,
			as.Count,
			as.FirstTime.Format("01-02 15:04"),
		))
	}

	sb.WriteString(fmt.Sprintf("\n⏰ 汇总发送时间：%s", time.Now().Format("2006-01-02 15:04:05")))

	return sb.String()
}

// ============================================
// 工具函数
// ============================================

// isAfterWorkStart 判断当前时间是否已过规则的工作开始时间
func isAfterWorkStart(rule *model.AlertRule) bool {
	if rule.WorkTimeStart == "" || rule.WorkTimeEnd == "" {
		return true
	}
	now := time.Now()
	currentTime := now.Format("15:04")
	return currentTime >= rule.WorkTimeStart
}

// severityEmojiStr 告警级别对应的 emoji
func severityEmojiStr(severity string) string {
	switch strings.ToLower(severity) {
	case "critical":
		return "🔴"
	case "warning":
		return "🟡"
	case "info":
		return "🔵"
	default:
		return "🟢"
	}
}

// getHighestSeverity 获取最高告警级别
func getHighestSeverity(summaries []alertSummary) string {
	severityOrder := map[string]int{"critical": 3, "warning": 2, "info": 1}
	highest := "info"
	highestLevel := 0
	for _, as := range summaries {
		level, ok := severityOrder[strings.ToLower(as.Severity)]
		if ok && level > highestLevel {
			highestLevel = level
			highest = as.Severity
		}
	}
	return highest
}

// StartSuppressedSender 启动抑制告警发送器（goroutine）
func StartSuppressedSender() {
	sender := NewSuppressedSender()
	go sender.Start()
}
