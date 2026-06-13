import request from '../utils/request'

// ============================================
// 告警源
// ============================================
export function getAlertSources(params) {
  return request({ url: '/alert-sources', method: 'get', params })
}

export function createAlertSource(data) {
  return request({ url: '/alert-sources', method: 'post', data })
}

export function getAlertSource(id) {
  return request({ url: `/alert-sources/${id}`, method: 'get' })
}

export function updateAlertSource(id, data) {
  return request({ url: `/alert-sources/${id}`, method: 'put', data })
}

export function deleteAlertSource(id) {
  return request({ url: `/alert-sources/${id}`, method: 'delete' })
}

// ============================================
// 规则
// ============================================
export function getAlertRules(params) {
  return request({ url: '/alert-rules', method: 'get', params })
}

export function createAlertRule(data) {
  return request({ url: '/alert-rules', method: 'post', data })
}

export function getAlertRule(id) {
  return request({ url: `/alert-rules/${id}`, method: 'get' })
}

export function updateAlertRule(id, data) {
  return request({ url: `/alert-rules/${id}`, method: 'put', data })
}

export function deleteAlertRule(id) {
  return request({ url: `/alert-rules/${id}`, method: 'delete' })
}

// ============================================
// 模板
// ============================================
export function getAlertTemplates(params) {
  return request({ url: '/alert-templates', method: 'get', params })
}

export function createAlertTemplate(data) {
  return request({ url: '/alert-templates', method: 'post', data })
}

export function getAlertTemplate(id) {
  return request({ url: `/alert-templates/${id}`, method: 'get' })
}

export function updateAlertTemplate(id, data) {
  return request({ url: `/alert-templates/${id}`, method: 'put', data })
}

export function deleteAlertTemplate(id) {
  return request({ url: `/alert-templates/${id}`, method: 'delete' })
}

// ============================================
// 通道
// ============================================
export function getAlertChannels(params) {
  return request({ url: '/alert-channels', method: 'get', params })
}

export function createAlertChannel(data) {
  return request({ url: '/alert-channels', method: 'post', data })
}

export function getAlertChannel(id) {
  return request({ url: `/alert-channels/${id}`, method: 'get' })
}

export function updateAlertChannel(id, data) {
  return request({ url: `/alert-channels/${id}`, method: 'put', data })
}

export function deleteAlertChannel(id) {
  return request({ url: `/alert-channels/${id}`, method: 'delete' })
}

export function testAlertChannel(id) {
  return request({ url: `/alert-channels/${id}/test`, method: 'post' })
}

// ============================================
// 告警流水
// ============================================
export function getAlertRecords(params) {
  return request({ url: '/alert-records', method: 'get', params })
}

export function ackAlert(id, data) {
  return request({ url: `/alert-records/${id}/ack`, method: 'post', data })
}

export function unackAlert(id) {
  return request({ url: `/alert-records/${id}/unack`, method: 'post' })
}

export function updateAlertNote(id, data) {
  return request({ url: `/alert-records/${id}/note`, method: 'post', data })
}

// ============================================
// AI 配置
// ============================================
export function getAIConfig() {
  return request({ url: '/ai-config', method: 'get' })
}

export function updateAIConfig(data) {
  return request({ url: '/ai-config', method: 'put', data })
}

export function testAIConnection(data) {
  return request({ url: '/ai-config/test', method: 'post', data })
}

// ============================================
// 告警统计 Dashboard
// ============================================
export function getStatsSummary(params) {
  return request({ url: '/alert-records/stats/summary', method: 'get', params })
}

export function getStatsDailyTrend(params) {
  return request({ url: '/alert-records/stats/daily-trend', method: 'get', params })
}

export function getStatsBySeverity(params) {
  return request({ url: '/alert-records/stats/by-severity', method: 'get', params })
}

export function getStatsTopAlerts(params) {
  return request({ url: '/alert-records/stats/top-alerts', method: 'get', params })
}

export function getStatsBySendStatus(params) {
  return request({ url: '/alert-records/stats/by-status', method: 'get', params })
}
