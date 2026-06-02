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
            <el-button @click="fetchList" style="margin-left:8px;">查询</el-button>
          </div>
        </div>
      </template>

      <el-table :data="list" stripe v-loading="loading" max-height="600">
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="alert_name" label="告警名称" min-width="150" />
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
        <el-table-column label="发送状态" width="90">
          <template #default="{ row }">
            <el-tag :type="sendStatusColor(row.send_status)" size="small">{{ sendStatusLabel(row.send_status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="send_error" label="错误信息" show-overflow-tooltip min-width="150" />
        <el-table-column prop="created_at" label="时间" width="170">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="primary" @click="showDetail(row)">详情</el-button>
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
    <el-dialog v-model="detailVisible" title="告警详情" width="700px">
      <template v-if="detail">
        <el-descriptions :column="2" border>
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
          <el-descriptions-item label="摘要" :span="2">{{ detail.summary }}</el-descriptions-item>
          <el-descriptions-item label="描述" :span="2">{{ detail.description }}</el-descriptions-item>
          <el-descriptions-item v-if="detail.send_error" label="错误" :span="2">
            <span style="color:red;">{{ detail.send_error }}</span>
          </el-descriptions-item>
          <el-descriptions-item v-if="detail.ai_suggestion" label="AI建议" :span="2">
            {{ detail.ai_suggestion }}
          </el-descriptions-item>
        </el-descriptions>
        <div v-if="detail.formatted_message" style="margin-top:16px;">
          <h4>格式化消息预览</h4>
          <el-input :model-value="detail.formatted_message" type="textarea" :rows="8" readonly />
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import * as alertApi from '@/api/alert'

const loading = ref(false)
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const sources = ref([])
const selectedSourceId = ref(null)
const filterStatus = ref('')
const detailVisible = ref(false)
const detail = ref(null)

const sendStatusLabel = (s) => ({ pending: '待发送', sent: '已发送', failed: '失败', suppressed: '已抑制' })[s] || s
const sendStatusColor = (s) => ({ pending: 'info', sent: 'success', failed: 'danger', suppressed: 'warning' })[s] || ''
const severityColor = (s) => ({ critical: 'danger', warning: 'warning', info: 'info' })[s] || ''

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
</style>
