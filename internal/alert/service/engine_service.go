package service

import (
	"alert-ops/internal/alert/adapter"
	alertRepo "alert-ops/internal/alert/repo"
	"alert-ops/internal/model"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"go.uber.org/zap"
)

// ErrNoAlerts 解析后告警数量为0
var ErrNoAlerts = errors.New("no alerts parsed from webhook body")

// EngineService 规则引擎核心服务接口（引擎 + webhook + 记录查询）
type EngineService interface {
	// 核心引擎
	ProcessAlert(source *model.AlertSource, alert *adapter.NormalizedAlert) error
	FlushSuppressedAlerts() error
	// Webhook 接收
	ReceiveAlertmanager(slug string, body []byte) error
	ParseAlertmanagerBody(body io.Reader) (*adapter.AMWebhookPayload, error)
	// 记录查询
	ListRecords(sourceID uint, page, pageSize int, status string) ([]model.AlertRecord, int64, error)
	GetRecordByID(id uint) (*model.AlertRecord, error)
}

type engineService struct {
	sourceRepo     alertRepo.AlertSourceRepo
	ruleRepo       alertRepo.AlertRuleRepo
	templateRepo   alertRepo.AlertTemplateRepo
	channelRepo    alertRepo.AlertChannelRepo
	recordRepo     alertRepo.AlertRecordRepo
	suppressedRepo alertRepo.SuppressedAlertRepo
	templateSvc    TemplateService
	channelSvc     ChannelService
	aiSvc          AIService
}

func NewEngineService(aiSvc AIService) EngineService {
	return &engineService{
		sourceRepo:     alertRepo.NewAlertSourceRepo(),
		ruleRepo:       alertRepo.NewAlertRuleRepo(),
		templateRepo:   alertRepo.NewAlertTemplateRepo(),
		channelRepo:    alertRepo.NewAlertChannelRepo(),
		recordRepo:     alertRepo.NewAlertRecordRepo(),
		suppressedRepo: alertRepo.NewSuppressedAlertRepo(),
		templateSvc:    NewTemplateService(),
		channelSvc:     NewChannelService(),
		aiSvc:          aiSvc,
	}
}

// ProcessAlert 处理单条标准化告警
func (s *engineService) ProcessAlert(source *model.AlertSource, alert *adapter.NormalizedAlert) error {
	zap.L().Info("=== 规则引擎开始处理 ===",
		zap.Uint("source_id", source.ID),
		zap.String("source_slug", source.Slug),
		zap.String("alert_name", alert.AlertName),
		zap.String("status", alert.Status),
		zap.String("severity", alert.Severity),
	)

	// 1. 加载该告警源的所有已启用规则
	rules, err := s.ruleRepo.ListBySource(source.ID)
	if err != nil {
		zap.L().Error("加载规则失败", zap.Uint("source_id", source.ID), zap.Error(err))
		return fmt.Errorf("加载规则失败: %w", err)
	}

	zap.L().Info("加载到规则",
		zap.Uint("source_id", source.ID),
		zap.Int("rule_count", len(rules)),
	)

	// 2. 筛选匹配的规则
	matchedRules := s.matchRules(rules, alert)
	zap.L().Info("规则匹配结果",
		zap.Int("total_rules", len(rules)),
		zap.Int("matched_rules", len(matchedRules)),
	)

	if len(matchedRules) == 0 {
		zap.L().Info("无匹配规则，走默认发送逻辑",
			zap.String("alert_name", alert.AlertName),
		)
		return s.processDefaultRule(source, alert)
	}

	// 3. 按优先级执行规则
	for i, rule := range matchedRules {
		if !rule.Enabled {
			zap.L().Info("规则已禁用，跳过",
				zap.Uint("rule_id", rule.ID),
				zap.String("rule_name", rule.Name),
			)
			continue
		}

		zap.L().Info("执行规则",
			zap.Int("index", i+1),
			zap.Uint("rule_id", rule.ID),
			zap.String("rule_name", rule.Name),
			zap.String("rule_type", rule.RuleType),
			zap.Int("priority", rule.Priority),
			zap.Int("channel_count", len(rule.Channels)),
		)

		switch rule.RuleType {
		case "time":
			if err := s.processTimeRule(source, &rule, alert); err != nil {
				zap.L().Error("时间规则处理失败",
					zap.Uint("rule_id", rule.ID),
					zap.Error(err),
				)
			}
		case "ai":
			if err := s.processAIRule(source, &rule, alert); err != nil {
				zap.L().Error("AI规则处理失败",
					zap.Uint("rule_id", rule.ID),
					zap.Error(err),
				)
			}
		default: // "default"
			if err := s.processDefaultRuleWithRule(source, &rule, alert); err != nil {
				zap.L().Error("默认规则处理失败",
					zap.Uint("rule_id", rule.ID),
					zap.Error(err),
				)
			}
		}

		// 告警源配置为 first-match 模式（continue_match=false），执行完第一条匹配规则后停止
		if !source.ContinueMatch {
			zap.L().Info("first-match 模式，停止继续匹配",
				zap.Uint("rule_id", rule.ID),
				zap.String("rule_name", rule.Name),
			)
			break
		}
	}

	zap.L().Info("=== 规则引擎处理完成 ===",
		zap.String("alert_name", alert.AlertName),
	)
	return nil
}

// matchRules 匹配规则（基于 match_labels）
func (s *engineService) matchRules(rules []model.AlertRule, alert *adapter.NormalizedAlert) []model.AlertRule {
	var matched []model.AlertRule
	for _, rule := range rules {
		matchResult := s.matchLabels(rule.MatchLabels, alert.Labels)
		zap.L().Info("规则标签匹配",
			zap.Uint("rule_id", rule.ID),
			zap.String("rule_name", rule.Name),
			zap.String("match_labels", rule.MatchLabels),
			zap.Any("alert_labels", alert.Labels),
			zap.Bool("matched", matchResult),
		)
		if matchResult {
			matched = append(matched, rule)
		}
	}
	return matched
}

// matchLabels 检查告警标签是否匹配规则定义的匹配条件
func (s *engineService) matchLabels(matchLabelsJSON string, alertLabels map[string]string) bool {
	if matchLabelsJSON == "" {
		return true // 没有匹配条件，匹配所有
	}

	var matchLabels map[string]string
	if err := json.Unmarshal([]byte(matchLabelsJSON), &matchLabels); err != nil {
		zap.L().Error("解析匹配标签失败", zap.Error(err))
		return false
	}

	for k, v := range matchLabels {
		if alertVal, ok := alertLabels[k]; !ok || alertVal != v {
			return false
		}
	}
	return true
}

// ============================================
// 默认规则处理：直接渲染模板 + 发送
// ============================================

func (s *engineService) processDefaultRule(source *model.AlertSource, alert *adapter.NormalizedAlert) error {
	zap.L().Info(">>> 默认发送（无匹配规则）", zap.String("alert", alert.AlertName))

	// 创建告警记录（仅记录，不发送）
	record := s.createAlertRecord(source.ID, nil, alert)
	record.SendStatus = "skipped"
	record.SendError = "无匹配规则且未配置默认通道"

	// 使用默认模板渲染
	_, content := s.templateSvc.Render(alert, DefaultTitleTpl, DefaultFeishuCardTpl)
	record.FormattedMessage = content

	if err := s.recordRepo.Create(record); err != nil {
		return fmt.Errorf("创建告警记录失败: %w", err)
	}

	zap.L().Info("<<< 无匹配规则，跳过发送", zap.String("alert", alert.AlertName))
	return nil
}

func (s *engineService) processDefaultRuleWithRule(source *model.AlertSource, rule *model.AlertRule, alert *adapter.NormalizedAlert) error {
	zap.L().Info(">>> 默认规则发送",
		zap.Uint("rule_id", rule.ID),
		zap.String("rule_name", rule.Name),
		zap.Int("rule_channels", len(rule.Channels)),
	)

	// 创建告警记录
	record := s.createAlertRecord(source.ID, &rule.ID, alert)
	if err := s.recordRepo.Create(record); err != nil {
		return fmt.Errorf("创建告警记录失败: %w", err)
	}

	// 发送（根据通道类型选择模板渲染）
	if len(rule.Channels) > 0 {
		zap.L().Info("通过规则关联通道发送", zap.Int("channel_count", len(rule.Channels)))
		for _, ch := range rule.Channels {
			if !ch.Enabled {
				zap.L().Info("通道已禁用，跳过", zap.String("channel_name", ch.Name))
				continue
			}
			// 根据通道类型选择模板渲染
			title, content := s.renderForChannel(rule.Template, alert, ch.Type, "")
			record.FormattedMessage = content

			zap.L().Info("发送到通道",
				zap.String("channel_name", ch.Name),
				zap.String("channel_type", ch.Type),
				zap.String("webhook_url", ch.WebhookURL),
			)
			result := s.channelSvc.SendWithRetry(&ch, title, content, alert.Status, alert.Severity)
			if !result.Success {
				zap.L().Error("通道发送失败",
					zap.String("channel", ch.Name),
					zap.String("error", result.Error),
				)
				record.SendError = result.Error
				record.SendStatus = "failed"
			} else {
				zap.L().Info("通道发送成功", zap.String("channel", ch.Name))
				record.SendStatus = "sent"
			}
		}
	} else {
		zap.L().Info("规则未关联通道，跳过发送")
		record.SendStatus = "skipped"
		record.SendError = "规则未关联任何通道"
	}

	now := time.Now()
	record.SentAt = &now
	s.recordRepo.Update(record)

	zap.L().Info("<<< 默认规则发送完成",
		zap.String("send_status", record.SendStatus),
	)
	return nil
}

// ============================================
// 时间规则处理：工作时间直接发送，非工作时间抑制
// ============================================

func (s *engineService) processTimeRule(source *model.AlertSource, rule *model.AlertRule, alert *adapter.NormalizedAlert) error {
	// 创建告警记录
	record := s.createAlertRecord(source.ID, &rule.ID, alert)
	if err := s.recordRepo.Create(record); err != nil {
		return fmt.Errorf("创建告警记录失败: %w", err)
	}

	if rule.SuppressOffHours && !s.isWorkTime(rule) {
		// 非工作时间，抑制告警
		zap.L().Info("非工作时间，抑制告警",
			zap.String("alert", alert.AlertName),
			zap.String("work_start", rule.WorkTimeStart),
			zap.String("work_end", rule.WorkTimeEnd),
		)
		record.SendStatus = "suppressed"

		// 计算下次上班时间
		scheduledTime := s.nextWorkTime(rule)
		sa := &model.SuppressedAlert{
			RecordID:        record.ID,
			RuleID:          rule.ID,
			SourceID:        source.ID,
			SuppressReason:  fmt.Sprintf("非工作时间(%s-%s)，抑制后上班统一发送", rule.WorkTimeStart, rule.WorkTimeEnd),
			ScheduledSendAt: &scheduledTime,
			Status:          "pending",
		}
		if err := s.suppressedRepo.Create(sa); err != nil {
			zap.L().Error("创建抑制记录失败", zap.Error(err))
		}

		s.recordRepo.Update(record)
		return nil
	}

	// 工作时间，直接发送
	return s.processDefaultRuleWithRule(source, rule, alert)
}

func (s *engineService) isWorkTime(rule *model.AlertRule) bool {
	if rule.WorkTimeStart == "" || rule.WorkTimeEnd == "" {
		return true // 没配置时间窗口，默认工作时间
	}

	now := time.Now()
	currentTime := now.Format("15:04")
	return currentTime >= rule.WorkTimeStart && currentTime <= rule.WorkTimeEnd
}

// isAfterWorkStart 判断当前时间是否已过规则的工作开始时间
// 用于 FlushSuppressedAlerts 场景：补发只需过了开始时间即可，不限制结束时间
func (s *engineService) isAfterWorkStart(rule *model.AlertRule) bool {
	if rule.WorkTimeStart == "" || rule.WorkTimeEnd == "" {
		return true
	}
	now := time.Now()
	currentTime := now.Format("15:04")
	return currentTime >= rule.WorkTimeStart
}

func (s *engineService) nextWorkTime(rule *model.AlertRule) time.Time {
	now := time.Now()
	startParts := parseTimeStr(rule.WorkTimeStart)
	if startParts == nil {
		return now.Add(10 * time.Minute)
	}

	next := time.Date(now.Year(), now.Month(), now.Day(), startParts[0], startParts[1], 0, 0, now.Location())
	if now.After(next) {
		// 今天上班时间已过，明天
		next = next.Add(24 * time.Hour)
	}
	// 跳过周末
	for next.Weekday() == time.Saturday || next.Weekday() == time.Sunday {
		next = next.Add(24 * time.Hour)
	}
	return next
}

func parseTimeStr(s string) []int {
	if len(s) < 5 {
		return nil
	}
	var h, m int
	fmt.Sscanf(s, "%d:%d", &h, &m)
	return []int{h, m}
}

// ============================================
// AI 规则处理：调用 AI 分析后发送
// ============================================

func (s *engineService) processAIRule(source *model.AlertSource, rule *model.AlertRule, alert *adapter.NormalizedAlert) error {
	// 创建告警记录
	record := s.createAlertRecord(source.ID, &rule.ID, alert)
	if err := s.recordRepo.Create(record); err != nil {
		return fmt.Errorf("创建告警记录失败: %w", err)
	}

	// 调用 AI 分析
	aiSuggestion := ""
	// 只有 firing 状态才调用 AI 分析，恢复不调用AI分析
	if rule.AIEnabled && alert.Status == "firing" {
		var err error
		aiSuggestion, err = s.aiSvc.Analyze(alert.Summary, alert.Description, rule.AIPromptTemplate)
		if err != nil {
			zap.L().Error("AI分析失败", zap.Error(err))
			aiSuggestion = fmt.Sprintf("AI分析失败: %v", err)
		}
		record.AISuggestion = aiSuggestion
	}

	// 发送（根据通道类型选择模板渲染，带 AI 建议）
	if len(rule.Channels) > 0 {
		for _, ch := range rule.Channels {
			if !ch.Enabled {
				continue
			}
			// 根据通道类型选择模板渲染
			title, content := s.renderForChannel(rule.Template, alert, ch.Type, aiSuggestion)
			record.FormattedMessage = content

			result := s.channelSvc.SendWithRetry(&ch, title, content, alert.Status, alert.Severity)
			if !result.Success {
				record.SendError = result.Error
				record.SendStatus = "failed"
			} else {
				record.SendStatus = "sent"
			}
		}
	} else {
		zap.L().Info("规则未关联通道，跳过发送")
		record.SendStatus = "skipped"
		record.SendError = "规则未关联任何通道"
	}

	now := time.Now()
	record.SentAt = &now
	s.recordRepo.Update(record)

	return nil
}

// ============================================
// 辅助方法
// ============================================

func (s *engineService) createAlertRecord(sourceID uint, ruleID *uint, alert *adapter.NormalizedAlert) *model.AlertRecord {
	return &model.AlertRecord{
		SourceID:    sourceID,
		RuleID:      ruleID,
		AlertName:   alert.AlertName,
		Status:      alert.Status,
		Severity:    alert.Severity,
		Instance:    alert.Instance,
		Summary:     alert.Summary,
		Description: alert.Description,
		RawData:     alert.RawData,
		SendStatus:  "pending",
	}
}

func (s *engineService) sendToAllChannels(sourceID uint, title, content string, record *model.AlertRecord) {
	channels, err := s.channelRepo.ListBySource(sourceID)
	if err != nil {
		zap.L().Error("获取通道列表失败", zap.Uint("source_id", sourceID), zap.Error(err))
		record.SendStatus = "failed"
		record.SendError = "获取通道列表失败"
		return
	}

	zap.L().Info("获取到告警源下所有通道",
		zap.Uint("source_id", sourceID),
		zap.Int("total_channels", len(channels)),
	)

	enabledCount := 0
	allSuccess := true
	for _, ch := range channels {
		if !ch.Enabled {
			zap.L().Info("通道已禁用，跳过",
				zap.String("channel_name", ch.Name),
				zap.String("channel_type", ch.Type),
			)
			continue
		}
		enabledCount++
		zap.L().Info("发送到通道",
			zap.String("channel_name", ch.Name),
			zap.String("channel_type", ch.Type),
			zap.String("webhook_url", ch.WebhookURL),
		)
		result := s.channelSvc.SendWithRetry(&ch, title, content, record.Status, record.Severity)
		if !result.Success {
			allSuccess = false
			zap.L().Error("通道发送失败",
				zap.String("channel", ch.Name),
				zap.String("error", result.Error),
			)
			if record.SendError != "" {
				record.SendError += "; "
			}
			record.SendError += result.Error
		} else {
			zap.L().Info("通道发送成功", zap.String("channel", ch.Name))
		}
	}

	zap.L().Info("通道发送汇总",
		zap.Int("total", len(channels)),
		zap.Int("enabled", enabledCount),
		zap.Bool("all_success", allSuccess),
	)

	if allSuccess && enabledCount > 0 {
		record.SendStatus = "sent"
	} else {
		record.SendStatus = "failed"
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

// FlushSuppressedAlerts 发送所有被抑制的告警（定时任务调用）
// 按 (RuleID, SourceID) 分组，每组构建一张汇总卡片发送
func (s *engineService) FlushSuppressedAlerts() error {
	suppressed, err := s.suppressedRepo.GetPending()
	if err != nil {
		return fmt.Errorf("获取抑制告警失败: %w", err)
	}

	if len(suppressed) == 0 {
		zap.L().Info("无待发送的抑制告警")
		return nil
	}

	zap.L().Info("开始发送抑制告警", zap.Int("count", len(suppressed)))

	// 1. 按 (RuleID, SourceID) 分组
	groups, err := s.groupSuppressedByRule(suppressed)
	if err != nil {
		return fmt.Errorf("分组抑制告警失败: %w", err)
	}

	zap.L().Info("抑制告警分组完成", zap.Int("group_count", len(groups)))

	// 2. 每组发送一张汇总卡片
	now := time.Now()
	for _, g := range groups {
		if err := s.sendSuppressedSummary(&g, now); err != nil {
			zap.L().Error("发送抑制告警汇总失败",
				zap.Uint("rule_id", g.RuleID),
				zap.Uint("source_id", g.SourceID),
				zap.Error(err),
			)
		}
	}

	return nil
}

// groupSuppressedByRule 将抑制告警按 (RuleID, SourceID) 分组，组内去重统计
func (s *engineService) groupSuppressedByRule(suppressed []model.SuppressedAlert) ([]suppressedGroup, error) {
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
			if !s.isAfterWorkStart(rule) {
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
func (s *engineService) dedupAndSummarize(records []*model.AlertRecord) []alertSummary {
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

// sendSuppressedSummary 发送一组抑制告警的汇总卡片
func (s *engineService) sendSuppressedSummary(g *suppressedGroup, now time.Time) error {
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
		highestSev := s.getHighestSeverity(summaries)
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
func (s *engineService) buildSummaryContent(suppressReason string, summaries []alertSummary, totalCount int) string {
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
func (s *engineService) getHighestSeverity(summaries []alertSummary) string {
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

// ============================================
// Webhook 接收（从 webhook_service.go 合并）
// ============================================

// ReceiveAlertmanager 接收 Alertmanager webhook
func (s *engineService) ReceiveAlertmanager(slug string, body []byte) error {
	zap.L().Info("开始处理 Webhook",
		zap.String("slug", slug),
		zap.Int("body_len", len(body)),
	)

	// 1. 根据 slug 查找告警源实例
	source, err := s.sourceRepo.GetBySlug(slug)
	if err != nil {
		zap.L().Error("告警源不存在", zap.String("slug", slug), zap.Error(err))
		return fmt.Errorf("告警源不存在: %s", slug)
	}

	zap.L().Info("找到告警源",
		zap.Uint("source_id", source.ID),
		zap.String("source_name", source.Name),
		zap.Bool("enabled", source.Enabled),
	)

	if !source.Enabled {
		zap.L().Warn("告警源已禁用", zap.String("slug", slug))
		return fmt.Errorf("告警源已禁用: %s", slug)
	}

	// 2. 使用适配器解析告警
	zap.L().Info("原始请求体",
		zap.String("slug", slug),
		zap.String("raw_body", string(body)),
	)

	adp := adapter.NewAlertmanagerAdapter()
	alerts, err := adp.Parse(body)
	if err != nil {
		zap.L().Error("解析告警数据失败", zap.String("slug", slug), zap.Error(err))
		return fmt.Errorf("解析告警数据失败: %w", err)
	}

	zap.L().Info("解析告警完成",
		zap.String("slug", slug),
		zap.Int("alert_count", len(alerts)),
	)

	if len(alerts) == 0 {
		zap.L().Warn("解析后告警数量为0",
			zap.String("slug", slug),
			zap.String("raw_body", string(body)),
		)
		return ErrNoAlerts
	}

	// 3. 异步逐条处理告警，webhook 立即返回
	zap.L().Info("开始异步处理告警",
		zap.String("slug", slug),
		zap.Int("alert_count", len(alerts)),
	)

	for i, alert := range alerts {
		idx := i + 1
		a := *alert // 解引用拷贝，避免闭包共享指针
		go func(index int, alertItem adapter.NormalizedAlert) {
			zap.L().Info("开始处理第N条告警",
				zap.Int("index", index),
				zap.Int("total", len(alerts)),
				zap.String("alert_name", alertItem.AlertName),
				zap.String("status", alertItem.Status),
				zap.String("severity", alertItem.Severity),
				zap.String("instance", alertItem.Instance),
			)

			if err := s.ProcessAlert(source, &alertItem); err != nil {
				zap.L().Error("处理告警失败",
					zap.String("alert", alertItem.AlertName),
					zap.Error(err),
				)
			}
		}(idx, a)
	}

	zap.L().Info("Webhook 已接收，异步处理中",
		zap.String("slug", slug),
		zap.Int("alert_count", len(alerts)),
	)
	return nil
}

// ParseAlertmanagerBody 解析 Alertmanager 请求体（用于预览）
func (s *engineService) ParseAlertmanagerBody(body io.Reader) (*adapter.AMWebhookPayload, error) {
	var payload adapter.AMWebhookPayload
	data, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	return &payload, nil
}

// ============================================
// 记录查询（从 record_service.go 合并）
// ============================================

// ListRecords 分页查询告警记录列表
func (s *engineService) ListRecords(sourceID uint, page, pageSize int, status string) ([]model.AlertRecord, int64, error) {
	return s.recordRepo.List(sourceID, page, pageSize, status)
}

// GetRecordByID 根据ID获取告警记录
func (s *engineService) GetRecordByID(id uint) (*model.AlertRecord, error) {
	return s.recordRepo.GetByID(id)
}

// renderForChannel 根据通道类型选择对应模板渲染
// 钉钉使用 DefaultDingtalkTpl（\n 显式换行），其他通道使用 DefaultFeishuCardTpl
func (s *engineService) renderForChannel(tpl *model.AlertTemplate, alert *adapter.NormalizedAlert, channelType string, aiSuggestion string) (title, content string) {
	if tpl != nil && tpl.ChannelType == channelType {
		// 有自定义模板且通道类型匹配，使用自定义模板
		title, content = s.templateSvc.RenderFromTemplate(alert, tpl, aiSuggestion)
	} else {
		// 根据通道类型选择默认模板
		contentTpl := DefaultFeishuCardTpl
		if channelType == "dingtalk" {
			contentTpl = DefaultDingtalkTpl
		} else if channelType == "wecom" {
			contentTpl = DefaultWecomTpl
		}

		title, content = s.templateSvc.Render(alert, DefaultTitleTpl, contentTpl)
		if aiSuggestion != "" {
			content += fmt.Sprintf("\n\n---\n🤖 AI分析建议：\n%s", aiSuggestion)
		}
	}

	// 钉钉 Markdown 需要 \n\n 才能换行，将单个 \n 替换为 \n\n
	if channelType == "dingtalk" {
		content = strings.ReplaceAll(content, "\n", "\n\n")
	}

	return title, content
}
