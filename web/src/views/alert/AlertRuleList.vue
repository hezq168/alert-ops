<template>
  <div class="alert-rule-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <div class="header-left">
            <span class="title">转发规则</span>
            <el-select v-model="selectedSourceId" placeholder="选择告警源" @change="fetchList" style="width:220px;margin-left:12px;" clearable>
              <el-option v-for="s in sources" :key="s.id" :label="s.name" :value="s.id" />
            </el-select>
          </div>
          <el-button type="primary" @click="showAddDialog" :disabled="!selectedSourceId" v-permission="'alert:rule:add'">
            <el-icon><Plus /></el-icon>添加规则
          </el-button>
        </div>
      </template>

      <el-table :data="list" stripe v-loading="loading">
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="name" label="规则名称" min-width="150" />
        <el-table-column label="类型" width="100">
          <template #default="{ row }">
            <el-tag :type="ruleTypeColor(row.rule_type)" size="small">{{ ruleTypeLabel(row.rule_type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="匹配条件" min-width="200">
          <template #default="{ row }">
            <template v-if="row.match_labels">
              <el-tag v-for="(v, k) in parseJSON(row.match_labels)" :key="k" size="small" style="margin-right:4px;">
                {{ k }}={{ v }}
              </el-tag>
            </template>
            <span v-else style="color:#999;">匹配所有</span>
          </template>
        </el-table-column>
        <el-table-column label="时间规则" width="160">
          <template #default="{ row }">
            <template v-if="row.rule_type === 'time'">
              {{ row.work_time_start || '09:00' }} - {{ row.work_time_end || '18:00' }}
              <el-tag v-if="row.suppress_off_hours" type="warning" size="small" style="margin-left:4px;">抑制</el-tag>
            </template>
            <span v-else style="color:#999;">-</span>
          </template>
        </el-table-column>
        <el-table-column label="AI分析" width="80">
          <template #default="{ row }">
            <el-tag :type="row.ai_enabled ? 'success' : 'info'" size="small">{{ row.ai_enabled ? '启用' : '关闭' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="通道" min-width="150">
          <template #default="{ row }">
            <el-tag v-for="ch in (row.channels || [])" :key="ch.id" size="small" style="margin-right:4px;">
              {{ ch.name }}
            </el-tag>
            <span v-if="!row.channels || row.channels.length === 0" style="color:#999;">使用源默认通道</span>
          </template>
        </el-table-column>
        <el-table-column prop="priority" label="优先级" width="80" sortable />
        <el-table-column label="状态" width="70">
          <template #default="{ row }">
            <el-tag :type="row.enabled ? 'success' : 'danger'" size="small">{{ row.enabled ? '启用' : '禁用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="primary" @click="handleEdit(row)" v-permission="'alert:rule:edit'">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)" v-permission="'alert:rule:delete'">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 添加/编辑弹窗 -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑规则' : '添加规则'" width="650px">
      <el-form :model="form" label-width="110px">
        <el-form-item label="规则名称" required>
          <el-input v-model="form.name" placeholder="如：生产环境严重告警转发" />
        </el-form-item>
        <el-form-item label="规则类型" required>
          <el-select v-model="form.rule_type" style="width:100%">
            <el-option label="默认规则（直接转发）" value="default" />
            <el-option label="时间规则（工作时间窗口）" value="time" />
            <el-option label="AI 规则（AI分析后发送）" value="ai" />
          </el-select>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="匹配标签">
          <div v-for="(item, idx) in matchLabels" :key="idx" style="display:flex;gap:8px;margin-bottom:4px;align-items:center;">
            <el-input v-model="item.key" placeholder="标签名" style="width:150px;" />
            <el-input v-model="item.value" placeholder="标签值" style="width:200px;" />
            <el-button type="danger" size="small" @click="matchLabels.splice(idx,1)">×</el-button>
          </div>
          <div style="display:flex;align-items:center;height:32px;">
            <el-button size="small" @click="matchLabels.push({key:'',value:''})">+ 添加匹配条件</el-button>
          </div>
        </el-form-item>
        <template v-if="form.rule_type === 'time'">
          <el-form-item label="工作时间">
            <el-time-picker v-model="workTimeRange" is-range format="HH:mm" value-format="HH:mm" range-separator="至" />
          </el-form-item>
          <el-form-item label="非工作时间抑制">
            <el-switch v-model="form.suppress_off_hours" />
            <span style="margin-left:8px;color:#999;font-size:12px;">开启后，非工作时间告警将被抑制，上班后统一发送</span>
          </el-form-item>
        </template>
        <template v-if="form.rule_type === 'ai'">
          <el-form-item label="启用AI分析">
            <el-switch v-model="form.ai_enabled" />
          </el-form-item>
          <el-form-item label="AI提示词">
            <el-input v-model="form.ai_prompt_template" type="textarea" :rows="3" placeholder="自定义AI分析提示词，留空使用默认" />
          </el-form-item>
        </template>
        <el-form-item label="消息模板">
          <el-select v-model="form.template_id" clearable placeholder="选择模板（可选）" style="width:100%">
            <el-option v-for="t in templates" :key="t.id" :label="t.name" :value="t.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="发送通道">
          <el-select v-model="form.channel_ids" multiple placeholder="选择通道（可选，不选则使用源默认通道）" style="width:100%">
            <el-option v-for="ch in channels" :key="ch.id" :label="`${ch.name} (${ch.type})`" :value="ch.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="优先级">
          <el-input-number v-model="form.priority" :min="0" :max="100" />
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="form.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import * as alertApi from '@/api/alert'

const route = useRoute()
const loading = ref(false)
const list = ref([])
const sources = ref([])
const templates = ref([])
const channels = ref([])
const selectedSourceId = ref(null)
const dialogVisible = ref(false)
const isEdit = ref(false)
const workTimeRange = ref([])
const matchLabels = ref([])

const form = ref({
  name: '', source_id: null, rule_type: 'default', description: '',
  match_labels: '', work_time_start: '09:00', work_time_end: '18:00',
  suppress_off_hours: false, ai_enabled: false, ai_prompt_template: '',
  template_id: null, channel_ids: [], priority: 0, enabled: true
})

const parseJSON = (str) => {
  try { return JSON.parse(str) } catch { return {} }
}

const ruleTypeLabel = (t) => ({ default: '默认', time: '时间', ai: 'AI分析' })[t] || t
const ruleTypeColor = (t) => ({ default: '', time: 'warning', ai: 'success' })[t] || ''

const loadSources = async () => {
  try {
    const res = await alertApi.getAlertSources({ page: 1, page_size: 100 })
    sources.value = res.data.list || []
  } catch (e) { /* handled */ }
}

const fetchList = async () => {
  if (!selectedSourceId.value) { list.value = []; return }
  loading.value = true
  try {
    const [ruleRes, tplRes, chRes] = await Promise.all([
      alertApi.getAlertRules({ source_id: selectedSourceId.value }),
      alertApi.getAlertTemplates({ source_id: selectedSourceId.value }),
      alertApi.getAlertChannels({ source_id: selectedSourceId.value })
    ])
    list.value = ruleRes.data || []
    templates.value = tplRes.data || []
    channels.value = chRes.data || []
  } catch (e) { /* handled */ }
  loading.value = false
}

const showAddDialog = () => {
  isEdit.value = false
  form.value = {
    name: '', source_id: selectedSourceId.value, rule_type: 'default', description: '',
    match_labels: '', work_time_start: '09:00', work_time_end: '18:00',
    suppress_off_hours: false, ai_enabled: false, ai_prompt_template: '',
    template_id: null, channel_ids: [], priority: 0, enabled: true
  }
  matchLabels.value = []
  workTimeRange.value = []
  dialogVisible.value = true
}

const handleEdit = (row) => {
  isEdit.value = true
  form.value = {
    ...row,
    template_id: row.template_id || null,
    channel_ids: (row.channels || []).map(c => c.id)
  }
  matchLabels.value = []
  if (row.match_labels) {
    const labels = parseJSON(row.match_labels)
    for (const [k, v] of Object.entries(labels)) {
      matchLabels.value.push({ key: k, value: v })
    }
  }
  workTimeRange.value = row.work_time_start && row.work_time_end
    ? [new Date(2000,0,1,...row.work_time_start.split(':')), new Date(2000,0,1,...row.work_time_end.split(':'))]
    : []
  dialogVisible.value = true
}

const handleSubmit = async () => {
  if (!form.value.name) { ElMessage.warning('请填写规则名称'); return }

  // 构建 match_labels JSON
  const labels = {}
  matchLabels.value.forEach(item => {
    if (item.key && item.value) labels[item.key] = item.value
  })
  form.value.match_labels = Object.keys(labels).length > 0 ? JSON.stringify(labels) : ''

  // 工作时间 -> HH:mm 字符串
  if (workTimeRange.value && workTimeRange.value.length === 2) {
    const fmt = (d) => {
      if (typeof d === 'string') return d
      if (d instanceof Date) {
        const h = String(d.getHours()).padStart(2, '0')
        const m = String(d.getMinutes()).padStart(2, '0')
        return `${h}:${m}`
      }
      return ''
    }
    form.value.work_time_start = fmt(workTimeRange.value[0])
    form.value.work_time_end = fmt(workTimeRange.value[1])
  }

  try {
    if (isEdit.value) {
      await alertApi.updateAlertRule(form.value.id, form.value)
      ElMessage.success('更新成功')
    } else {
      await alertApi.createAlertRule(form.value)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchList()
  } catch (e) { /* handled */ }
}

const handleDelete = async (row) => {
  await ElMessageBox.confirm('确定删除该规则吗？', '确认删除', { type: 'warning' })
  try {
    await alertApi.deleteAlertRule(row.id)
    ElMessage.success('删除成功')
    fetchList()
  } catch (e) { /* handled */ }
}

onMounted(async () => {
  await loadSources()
  // 从 URL 参数获取默认选中的告警源
  if (route.query.source_id) {
    selectedSourceId.value = Number(route.query.source_id)
  }
  if (selectedSourceId.value) {
    fetchList()
  }
})
</script>

<style scoped>
.card-header { display: flex; justify-content: space-between; align-items: center; }
.header-left { display: flex; align-items: center; }
.title { font-size: 16px; font-weight: bold; }
</style>
