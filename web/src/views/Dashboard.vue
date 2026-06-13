<template>
  <div class="alert-dashboard">
    <!-- 筛选栏 -->
    <el-card class="filter-card">
      <div class="filter-row">
        <div class="filter-left">
          <el-select v-model="selectedSourceId" placeholder="告警源" @change="refreshAll" clearable style="width:200px;">
            <el-option label="全部" :value="0" />
            <el-option v-for="s in sources" :key="s.id" :label="s.name" :value="s.id" />
          </el-select>
          <el-select v-model="selectedDays" @change="refreshAll" style="width:120px;margin-left:8px;">
            <el-option label="近24小时" :value="1" />
            <el-option label="近7天" :value="7" />
            <el-option label="近30天" :value="30" />
          </el-select>
          <el-select v-model="selectedSeverity" @change="refreshTrend" clearable placeholder="级别筛选" style="width:120px;margin-left:8px;">
            <el-option label="全部" value="" />
            <el-option label="critical" value="critical" />
            <el-option label="warning" value="warning" />
            <el-option label="info" value="info" />
          </el-select>
        </div>
        <div class="filter-right">
          <el-button @click="refreshAll" :loading="loading">刷新</el-button>
        </div>
      </div>
    </el-card>

    <!-- 数字卡片 -->
    <div class="stat-cards">
      <div class="stat-card card-blue">
        <div class="card-icon">📊</div>
        <div class="card-body">
          <div class="card-value">{{ summary.today_count }}</div>
          <div class="card-label">今日告警数</div>
        </div>
      </div>
      <div class="stat-card card-red">
        <div class="card-icon">🔥</div>
        <div class="card-body">
          <div class="card-value">{{ summary.firing_count }}</div>
          <div class="card-label">当前 Firing</div>
        </div>
      </div>
      <div class="stat-card card-orange">
        <div class="card-icon">❌</div>
        <div class="card-body">
          <div class="card-value">{{ summary.failed_count }}</div>
          <div class="card-label">发送失败</div>
        </div>
      </div>
      <div class="stat-card card-purple">
        <div class="card-icon">🔇</div>
        <div class="card-body">
          <div class="card-value">{{ summary.suppressed_count }}</div>
          <div class="card-label">被抑制告警</div>
        </div>
      </div>
    </div>

    <!-- 图表区 -->
    <div class="charts-row">
      <!-- 趋势折线图 -->
      <el-card class="chart-card chart-wide">
        <template #header>
          <span class="chart-title">📈 告警趋势（近{{ selectedDays }}天）</span>
        </template>
        <div class="chart-container">
          <svg :viewBox="`0 0 ${chartWidth} ${chartHeight}`" class="trend-chart" v-if="dailyTrend.length > 0">
            <!-- Y轴刻度 -->
            <line v-for="i in 5" :key="'grid'+i"
              :x1="60" :y1="20 + (i-1) * (chartHeight - 60) / 4"
              :x2="chartWidth - 20" :y2="20 + (i-1) * (chartHeight - 60) / 4"
              stroke="#eee" stroke-width="1" />
            <text v-for="i in 5" :key="'ylab'+i"
              :x="55" :y="20 + (i-1) * (chartHeight - 60) / 4 + 4"
              text-anchor="end" font-size="10" fill="#999">
              {{ yMax - Math.round((i-1) * yMax / 4) }}
            </text>

            <!-- X轴日期标签 -->
            <text v-for="(item, idx) in dailyTrend" :key="'xlab'+idx"
              :x="60 + idx * (chartWidth - 80) / (dailyTrend.length - 1 || 1)"
              :y="chartHeight - 5" text-anchor="middle" font-size="9" fill="#999"
              v-if="dailyTrend.length <= 14 || idx % Math.ceil(dailyTrend.length / 10) === 0">
              {{ formatDateLabel(item.date) }}
            </text>

            <!-- Firing 折线 -->
            <polyline :points="firingPoints" fill="none" stroke="#f56c6c" stroke-width="2.5" />
            <circle v-for="(item, idx) in dailyTrend" :key="'fdot'+idx"
              :cx="60 + idx * (chartWidth - 80) / (dailyTrend.length - 1 || 1)"
              :cy="20 + (chartHeight - 60) * (1 - (item.firing || 0) / (yMax || 1))"
              r="3" fill="#f56c6c" />

            <!-- Resolved 折线 -->
            <polyline :points="resolvedPoints" fill="none" stroke="#67c23a" stroke-width="2.5" />
            <circle v-for="(item, idx) in dailyTrend" :key="'rdot'+idx"
              :cx="60 + idx * (chartWidth - 80) / (dailyTrend.length - 1 || 1)"
              :cy="20 + (chartHeight - 60) * (1 - (item.resolved || 0) / (yMax || 1))"
              r="3" fill="#67c23a" />
          </svg>

          <div class="chart-legend">
            <span class="legend-item"><span class="dot" style="background:#f56c6c;"></span> Firing</span>
            <span class="legend-item"><span class="dot" style="background:#67c23a;"></span> Resolved</span>
          </div>

          <el-empty v-if="dailyTrend.length === 0 && !loading" description="暂无数据" :image-size="60" />
        </div>
      </el-card>

      <!-- Top 10 告警 -->
      <el-card class="chart-card chart-wide">
        <template #header>
          <span class="chart-title">🏆 Top 10 告警名称</span>
        </template>
        <div class="chart-container">
          <div class="bar-chart" v-if="topAlerts.length > 0">
            <div v-for="(item, idx) in topAlerts" :key="idx" class="bar-row">
              <span class="bar-label" :title="item.name">{{ idx + 1 }}. {{ truncateName(item.name, 30) }}</span>
              <div class="bar-track">
                <div class="bar-fill" :style="{width: (item.count / topMax * 100) + '%', background: barColor(idx)}"></div>
              </div>
              <span class="bar-count">{{ item.count }}</span>
            </div>
          </div>
          <el-empty v-if="topAlerts.length === 0 && !loading" description="暂无数据" :image-size="60" />
        </div>
      </el-card>

      <!-- 级别分布饼图 -->
      <el-card class="chart-card">
        <template #header>
          <span class="chart-title">🎯 告警级别分布</span>
        </template>
        <div class="chart-container">
          <svg viewBox="0 0 200 200" class="pie-chart" v-if="severityData.length > 0">
            <g v-for="(slice, idx) in pieSlices" :key="'slice'+idx">
              <path :d="slice.path" :fill="slice.color" stroke="#fff" stroke-width="2">
                <title>{{ slice.name }}: {{ slice.count }}</title>
              </path>
            </g>
            <text x="100" y="95" text-anchor="middle" font-size="14" font-weight="bold" fill="#333">
              {{ totalSeverity }}
            </text>
            <text x="100" y="112" text-anchor="middle" font-size="10" fill="#999">总计</text>
          </svg>
          <div class="pie-legend">
            <div v-for="item in severityData" :key="item.name" class="legend-item">
              <span class="dot" :style="{background: severityColor(item.name)}"></span>
              <span>{{ severityLabel(item.name) }} ({{ item.count }})</span>
            </div>
          </div>
          <el-empty v-if="severityData.length === 0 && !loading" description="暂无数据" :image-size="60" />
        </div>
      </el-card>

      <!-- 发送状态分布 -->
      <el-card class="chart-card">
        <template #header>
          <span class="chart-title">📬 发送状态分布</span>
        </template>
        <div class="chart-container">
          <div class="status-bars" v-if="sendStatusData.length > 0">
            <div v-for="item in sendStatusData" :key="item.name" class="status-row">
              <span class="status-label">{{ sendStatusLabel(item.name) }}</span>
              <div class="status-track">
                <div class="status-fill" :style="{width: (item.count / statusMax * 100) + '%', background: sendStatusColor(item.name)}"></div>
              </div>
              <span class="status-count">{{ item.count }}</span>
            </div>
          </div>
          <el-empty v-if="sendStatusData.length === 0 && !loading" description="暂无数据" :image-size="60" />
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import * as alertApi from '@/api/alert'

const loading = ref(false)
const sources = ref([])
const selectedSourceId = ref(0)
const selectedDays = ref(7)
const selectedSeverity = ref('')

// 数据
const summary = ref({ today_count: 0, firing_count: 0, failed_count: 0, suppressed_count: 0 })
const dailyTrend = ref([])
const severityData = ref([])
const topAlerts = ref([])
const sendStatusData = ref([])

// 图表尺寸
const chartWidth = 600
const chartHeight = 250

// 标签映射
const severityLabel = (s) => ({ critical: '严重', warning: '警告', info: '提示', unknown: '未知' })[s] || s
const severityColor = (s) => ({ critical: '#f56c6c', warning: '#e6a23c', info: '#909399', unknown: '#c0c4cc' })[s] || '#c0c4cc'

const sendStatusLabel = (s) => ({ pending: '待发送', sent: '已发送', failed: '失败', suppressed: '已抑制', skipped: '已跳过' })[s] || s
const sendStatusColor = (s) => ({ pending: '#909399', sent: '#67c23a', failed: '#f56c6c', suppressed: '#e6a23c', skipped: '#c0c4cc' })[s] || '#c0c4cc'

const barColor = (idx) => {
  const colors = ['#409eff', '#67c23a', '#e6a23c', '#f56c6c', '#909399', '#409eff', '#67c23a', '#e6a23c', '#f56c6c', '#909399']
  return colors[idx % 10]
}

// 折线图 - Y轴最大值
const yMax = computed(() => {
  if (dailyTrend.value.length === 0) return 1
  let max = 1
  dailyTrend.value.forEach(d => {
    if (d.firing > max) max = d.firing
    if (d.resolved > max) max = d.resolved
  })
  return Math.ceil(max * 1.2)
})

// 折线图 - Firing 坐标点
const firingPoints = computed(() => {
  return dailyTrend.value.map((item, idx) => {
    const x = 60 + idx * (chartWidth - 80) / (dailyTrend.value.length - 1 || 1)
    const y = 20 + (chartHeight - 60) * (1 - (item.firing || 0) / (yMax.value || 1))
    return `${x},${y}`
  }).join(' ')
})

// 折线图 - Resolved 坐标点
const resolvedPoints = computed(() => {
  return dailyTrend.value.map((item, idx) => {
    const x = 60 + idx * (chartWidth - 80) / (dailyTrend.value.length - 1 || 1)
    const y = 20 + (chartHeight - 60) * (1 - (item.resolved || 0) / (yMax.value || 1))
    return `${x},${y}`
  }).join(' ')
})

// 饼图数据
const totalSeverity = computed(() => severityData.value.reduce((s, i) => s + i.count, 0))
const pieSlices = computed(() => {
  const total = totalSeverity.value || 1
  let startAngle = -Math.PI / 2
  return severityData.value.map(item => {
    const angle = (item.count / total) * Math.PI * 2
    const x1 = 100 + 70 * Math.cos(startAngle)
    const y1 = 100 + 70 * Math.sin(startAngle)
    const x2 = 100 + 70 * Math.cos(startAngle + angle)
    const y2 = 100 + 70 * Math.sin(startAngle + angle)
    const largeArc = angle > Math.PI ? 1 : 0
    const path = `M 100 100 L ${x1} ${y1} A 70 70 0 ${largeArc} 1 ${x2} ${y2} Z`
    const slice = { path, color: severityColor(item.name), name: item.name, count: item.count }
    startAngle += angle
    return slice
  })
})

// Top 告警最大值
const topMax = computed(() => {
  if (topAlerts.value.length === 0) return 1
  return Math.max(...topAlerts.value.map(i => i.count))
})

// 发送状态最大值
const statusMax = computed(() => {
  if (sendStatusData.value.length === 0) return 1
  return Math.max(...sendStatusData.value.map(i => i.count))
})

// 格式化日期标签
const formatDateLabel = (date) => {
  if (!date) return ''
  const parts = date.split('-')
  if (parts.length === 3) return parts[1] + '/' + parts[2]
  return date
}

// 截断告警名称
const truncateName = (name, max) => {
  if (!name) return '-'
  return name.length > max ? name.slice(0, max) + '...' : name
}

// 加载告警源列表
const loadSources = async () => {
  try {
    const res = await alertApi.getAlertSources({ page: 1, page_size: 100 })
    sources.value = res.data.list || []
  } catch (e) { /* handled */ }
}

// 刷新汇总卡片
const fetchSummary = async () => {
  try {
    const params = { source_id: selectedSourceId.value, days: selectedDays.value }
    const res = await alertApi.getStatsSummary(params)
    summary.value = res.data || { today_count: 0, firing_count: 0, failed_count: 0, suppressed_count: 0 }
  } catch (e) { /* handled */ }
}

// 刷新趋势
const fetchTrend = async () => {
  try {
    const params = { source_id: selectedSourceId.value, days: selectedDays.value }
    if (selectedSeverity.value) params.severity = selectedSeverity.value
    const res = await alertApi.getStatsDailyTrend(params)
    dailyTrend.value = res.data || []
  } catch (e) { /* handled */ }
}

// 刷新级别分布
const fetchSeverity = async () => {
  try {
    const params = { source_id: selectedSourceId.value, days: selectedDays.value }
    const res = await alertApi.getStatsBySeverity(params)
    severityData.value = res.data || []
  } catch (e) { /* handled */ }
}

// 刷新 Top 告警
const fetchTopAlerts = async () => {
  try {
    const params = { source_id: selectedSourceId.value, days: selectedDays.value, limit: 10 }
    const res = await alertApi.getStatsTopAlerts(params)
    topAlerts.value = res.data || []
  } catch (e) { /* handled */ }
}

// 刷新发送状态
const fetchSendStatus = async () => {
  try {
    const params = { source_id: selectedSourceId.value, days: selectedDays.value }
    const res = await alertApi.getStatsBySendStatus(params)
    sendStatusData.value = res.data || []
  } catch (e) { /* handled */ }
}

// 刷新趋势（级别变化时）
const refreshTrend = () => {
  fetchTrend()
}

// 全部刷新
const refreshAll = async () => {
  loading.value = true
  await Promise.all([fetchSummary(), fetchTrend(), fetchSeverity(), fetchTopAlerts(), fetchSendStatus()])
  loading.value = false
}

onMounted(async () => {
  await loadSources()
  refreshAll()
})
</script>

<style scoped>
.alert-dashboard { padding: 4px 0; }

/* 筛选栏 */
.filter-card { margin-bottom: 16px; }
.filter-row { display: flex; justify-content: space-between; align-items: center; }
.filter-left { display: flex; align-items: center; }

/* 数字卡片 */
.stat-cards { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; margin-bottom: 16px; }
.stat-card { display: flex; align-items: center; padding: 20px; border-radius: 8px; color: #fff; }
.card-blue { background: linear-gradient(135deg, #409eff, #337ecc); }
.card-red { background: linear-gradient(135deg, #f56c6c, #c45656); }
.card-orange { background: linear-gradient(135deg, #e6a23c, #b88230); }
.card-purple { background: linear-gradient(135deg, #9b59b6, #7d3c98); }
.card-icon { font-size: 32px; margin-right: 16px; }
.card-body { flex: 1; }
.card-value { font-size: 28px; font-weight: bold; line-height: 1.2; }
.card-label { font-size: 13px; opacity: 0.9; margin-top: 4px; }

/* 图表行 */
.charts-row { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }
.chart-card { min-height: 0; }
.chart-card .chart-container { min-height: 200px; display: flex; flex-direction: column; align-items: center; justify-content: center; }
.chart-title { font-weight: bold; font-size: 14px; }

/* 宽图占满两列 */
.chart-wide { grid-column: span 2; }

/* 趋势图 */
.trend-chart { width: 100%; max-height: 280px; }

/* 图例 */
.chart-legend { display: flex; justify-content: center; gap: 20px; margin-top: 8px; }
.pie-legend { display: flex; flex-wrap: wrap; justify-content: center; gap: 12px; margin-top: 8px; }
.legend-item { display: flex; align-items: center; gap: 4px; font-size: 12px; }
.dot { display: inline-block; width: 10px; height: 10px; border-radius: 50%; }

/* 饼图 */
.pie-chart { width: 180px; height: 180px; }

/* 柱状图 */
.bar-chart { width: 100%; padding: 8px 0; }
.bar-row { display: flex; align-items: center; margin-bottom: 8px; }
.bar-label { width: 200px; font-size: 12px; text-align: right; padding-right: 8px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; color: #666; }
.bar-track { flex: 1; height: 20px; background: #f0f2f5; border-radius: 4px; overflow: hidden; }
.bar-fill { height: 100%; border-radius: 4px; transition: width 0.5s ease; min-width: 4px; }
.bar-count { width: 50px; font-size: 12px; font-weight: bold; text-align: right; padding-left: 8px; color: #333; }

/* 发送状态条 */
.status-bars { width: 100%; padding: 8px 16px; }
.status-row { display: flex; align-items: center; margin-bottom: 12px; }
.status-label { width: 60px; font-size: 12px; color: #666; text-align: right; padding-right: 12px; }
.status-track { flex: 1; height: 24px; background: #f0f2f5; border-radius: 4px; overflow: hidden; }
.status-fill { height: 100%; border-radius: 4px; transition: width 0.5s ease; min-width: 4px; display: flex; align-items: center; padding-left: 8px; }
.status-count { width: 50px; font-size: 13px; font-weight: bold; text-align: right; padding-left: 8px; color: #333; }

/* 响应式 */
@media (max-width: 1200px) {
  .stat-cards { grid-template-columns: repeat(2, 1fr); }
  .charts-row { grid-template-columns: 1fr; }
  .chart-wide { grid-column: span 1; }
}
@media (max-width: 768px) {
  .stat-cards { grid-template-columns: 1fr; }
  .bar-label { width: 120px; }
}
</style>
