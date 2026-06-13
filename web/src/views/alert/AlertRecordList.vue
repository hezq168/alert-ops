<template>
  <div class="alert-record-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <div class="header-left">
            <span class="title">告警流水</span>
          </div>
          <div class="header-right">
            <el-select v-model="selectedSourceId" placeholder="告警源" @change="fetchList" clearable style="width:200px;">
              <el-option v-for="s in sources" :key="s.id" :label="s.name" :value="s.id" />
            </el-select>
            <el-select v-model="filterStatus" placeholder="发送状态" @change="fetchList" clearable style="width:130px;margin-left:8px;">
              <el-option label="待发送" value="pending" />
              <el-option label="已发送" value="sent" />
              <el-option label="发送失败" value="failed" />
              <el-option label="已抑制" value="suppressed" />
            </el-select>
            <el-select v-model="filterAcknowledged" placeholder="确认状态" clearable style="width:130px;margin-left:8px;">
              <el-option label="待确认" value="false" />
              <el-option label="已确认" value="true" />
            </el-select>
            <el-button @click="fetchList" style="margin-left:8px;">查询</el-button>
          </div>
        </div>
      </template>

      <el-table :data="filteredList" stripe v-loading="loading" max-height="600">
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="alert_name" label="告警名称" show-overflow-tooltip width="160" />
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 'firing' ? 'danger' : 'success'" size="small">
              {{ row.status === 'firing' ? '触发' : '恢复' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="级别" width="80">
          <template #default="{ row }">
            <el-tag :type="severityColor(row.severity)" size="small">{{ row.severity || '-' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="instance" label="实例" show-overflow-tooltip min-width="120" />
        <el-table-column prop="summary" label="摘要" show-overflow-tooltip min-width="200" />
        <el-table-column label="确认状态" width="100">
          <template #default="{ row }">
            <el-tag v-if="row.acknowledged" type="success" size="small">已确认</el-tag>
            <el-tag v-else-if="row.status === 'firing'" type="warning" size="small">待确认</el-tag>
            <span v-else style="color:#909399;font-size:12px;">-</span>
          </template>
        </el-table-column>
        <el-table-column label="发送状态" width="90">
          <template #default="{ row }">
            <el-tooltip :content="row.send_error" :disabled="!row.send_error" placement="top">
              <el-tag :type="sendStatusColor(row.send_status)" size="small">{{ sendStatusLabel(row.send_status) }}</el-tag>
            </el-tooltip>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="时间" width="170">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="230" fixed="right">
          <template #default="{ row }">
            <div class="action-btns">
              <el-button size="small" type="primary" @click="showDetail(row)">详情</el-button>
              <el-button v-if="row.status === 'firing' && !row.acknowledged" size="small" type="warning" @click="handleAck(row)">确认</el-button>
              <template v-if="row.status === 'firing' && row.acknowledged">
                <el-button size="small" type="info" @click="handleUnack(row)">撤销</el-button>
                <el-button size="small" type="success" @click="handleNote(row)">{{ row.process_note ? '修改备注' : '填写备注' }}</el-button>
              </template>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next"
          @change="fetchList"
        />
      </div>
    </el-card>

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="告警详情" width="950px">
      <template v-if="detail">
        <el-descriptions :column="2" size="small" :labelStyle="{width:'100px',fontWeight:'bold',color:'#606266'}">
          <el-descriptions-item label="告警名称">{{ detail.alert_name }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="detail.status === 'firing' ? 'danger' : 'success'" size="small">{{ detail.status }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="级别">{{ detail.severity }}</el-descriptions-item>
          <el-descriptions-item label="实例">{{ detail.instance }}</el-descriptions-item>
          <el-descriptions-item label="发送状态">
            <el-tag :type="sendStatusColor(detail.send_status)" size="small">{{ sendStatusLabel(detail.send_status) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
          <el-descriptions-item label="确认状态">
            <el-tag v-if="detail.acknowledged" type="success" size="small">已确认</el-tag>
            <el-tag v-else-if="detail.status === 'firing'" type="warning" size="small">待确认</el-tag>
            <span v-else>-</span>
          </el-descriptions-item>
          <el-descriptions-item v-if="detail.acknowledged_by" label="确认人">{{ detail.acknowledged_by }}</el-descriptions-item>
          <el-descriptions-item v-if="detail.acknowledged_at" label="确认时间">{{ formatTime(detail.acknowledged_at) }}</el-descriptions-item>
        </el-descriptions>
        <el-divider style="margin:12px 0;" />
        <el-descriptions :column="1" size="small" :labelStyle="{width:'100px',fontWeight:'bold',color:'#606266'}">
          <el-descriptions-item label="摘要">{{ detail.summary }}</el-descriptions-item>
          <el-descriptions-item label="描述">{{ detail.description }}</el-descriptions-item>
          <el-descriptions-item v-if="detail.process_note" label="处理备注">{{ detail.process_note }}</el-descriptions-item>
          <el-descriptions-item v-if="detail.send_error" label="错误信息">
            <span style="color:red;">{{ detail.send_error }}</span>
          </el-descriptions-item>
          <el-descriptions-item v-if="detail.ai_suggestion" label="AI建议">
            {{ detail.ai_suggestion }}
          </el-descriptions-item>
        </el-descriptions>
        <div v-if="detail.formatted_message" style="margin-top:16px;">
          <div style="font-size:14px;font-weight:600;margin-bottom:8px;color:#606266;">格式化消息预览</div>
          <el-input :model-value="detail.formatted_message" type="textarea" :rows="8" readonly style="font-size:13px;font-family:monospace;" />
        </div>
        <div style="margin-top:16px;text-align:right;">
          <el-button v-if="detail.status === 'firing' && !detail.acknowledged" type="warning" @click="handleAckFromDetail">确认告警</el-button>
          <template v-if="detail.status === 'firing' && detail.acknowledged">
            <el-button type="info" @click="handleUnackFromDetail">取消确认</el-button>
            <el-button type="success" @click="handleNoteFromDetail">{{ detail.process_note ? '修改备注' : '填写备注' }}</el-button>
          </template>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import * as alertApi from '@/api/alert'

const loading = ref(false)
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const sources = ref([])
const selectedSourceId = ref(null)
const filterStatus = ref('')
const filterAcknowledged = ref('')
const detailVisible = ref(false)
const detail = ref(null)

const sendStatusLabel = (s) => ({ pending: '待发送', sent: '已发送', failed: '失败', suppressed: '已抑制' })[s] || s
const sendStatusColor = (s) => ({ pending: 'info', sent: 'success', failed: 'danger', suppressed: 'warning' })[s] || ''
const severityColor = (s) => ({ critical: 'danger', warning: 'warning', info: 'info' })[s] || ''

// 纯前端过滤：已确认/待确认
const filteredList = computed(() => {
  if (!filterAcknowledged.value) return list.value
  if (filterAcknowledged.value === 'true') {
    return list.value.filter(r => r.acknowledged === true)
  }
  // 'false' = 待确认：未确认 + firing 中（排除已恢复的）
  return list.value.filter(r => !r.acknowledged && r.status === 'firing')
})

const loadSources = async () => {
  try {
    const res = await alertApi.getAlertSources({ page: 1, page_size: 100 })
    sources.value = res.data.list || []
  } catch (e) { /* handled */ }
}

const fetchList = async () => {
  loading.value = true
  try {
    const params = { page: page.value, page_size: pageSize.value }
    if (selectedSourceId.value) params.source_id = selectedSourceId.value
    if (filterStatus.value) params.status = filterStatus.value
    // filterAcknowledged 改为前端 computed 过滤，不传给后端
    const res = await alertApi.getAlertRecords(params)
    list.value = res.data.list || []
    total.value = res.data.total || 0
  } catch (e) { /* handled */ }
  loading.value = false
}

const showDetail = (row) => {
  detail.value = row
  detailVisible.value = true
}

// 确认告警（列表操作列）
const handleAck = async (row) => {
  try {
    await ElMessageBox.confirm('确认收到此告警并开始处理？', '确认告警', { type: 'warning' })
  } catch { return }
  try {
    await alertApi.ackAlert(row.id, { user: 'admin' })
    ElMessage.success('已确认')
    row.acknowledged = true
    row.acknowledged_by = 'admin'
    row.acknowledged_at = new Date().toISOString()
    if (detail.value && detail.value.id === row.id) {
      detail.value.acknowledged = true
      detail.value.acknowledged_by = 'admin'
      detail.value.acknowledged_at = new Date().toISOString()
    }
  } catch { ElMessage.error('确认失败') }
}

// 填备注（列表操作列）
const handleNote = async (row) => {
  try {
    const { value } = await ElMessageBox.prompt('请输入处理备注', '处理备注', {
      confirmButtonText: '保存',
      inputValue: row.process_note || '',
      inputType: 'textarea'
    })
    if (value != null) {
      await alertApi.updateAlertNote(row.id, { note: value })
      ElMessage.success('备注已保存')
      row.process_note = value
      if (detail.value && detail.value.id === row.id) {
        detail.value.process_note = value
      }
    }
  } catch { /* cancelled */ }
}

// 确认告警（详情弹窗内）
const handleAckFromDetail = () => {
  if (detail.value) handleAck(detail.value)
}

// 取消确认（列表操作列）
const handleUnack = async (row) => {
  try {
    await ElMessageBox.confirm('确定要取消此告警的确认状态吗？', '取消确认', { type: 'info' })
  } catch { return }
  try {
    await alertApi.unackAlert(row.id)
    ElMessage.success('已取消确认')
    row.acknowledged = false
    row.acknowledged_by = ''
    row.acknowledged_at = null
    if (detail.value && detail.value.id === row.id) {
      detail.value.acknowledged = false
      detail.value.acknowledged_by = ''
      detail.value.acknowledged_at = null
    }
  } catch { ElMessage.error('取消确认失败') }
}

// 填备注（详情弹窗内）
const handleNoteFromDetail = () => {
  if (detail.value) handleNote(detail.value)
}

// 取消确认（详情弹窗内）
const handleUnackFromDetail = () => {
  if (detail.value) handleUnack(detail.value)
}

const formatTime = (t) => {
  if (!t) return '-'
  return new Date(t).toLocaleString('zh-CN')
}

onMounted(async () => {
  await loadSources()
  fetchList()
})
</script>

<style scoped>
.card-header { display: flex; justify-content: space-between; align-items: center; }
.header-left { display: flex; align-items: center; }
.header-right { display: flex; align-items: center; }
.title { font-size: 16px; font-weight: bold; }
.pagination { margin-top: 16px; display: flex; justify-content: flex-end; }
.action-btns { display: flex; gap: 4px; flex-wrap: nowrap; white-space: nowrap; }
</style>
