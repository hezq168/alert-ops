<template>
  <div class="alert-template-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <div class="header-left">
            <span class="title">消息模板</span>
            <el-select v-model="selectedSourceId" placeholder="选择告警源" @change="fetchList" style="width:220px;margin-left:12px;" clearable>
              <el-option v-for="s in sources" :key="s.id" :label="s.name" :value="s.id" />
            </el-select>
          </div>
          <el-button type="primary" @click="showAddDialog" :disabled="!selectedSourceId" v-permission="'alert:template:add'">
            <el-icon><Plus /></el-icon>添加模板
          </el-button>
        </div>
      </template>

      <el-table :data="list" stripe v-loading="loading">
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="name" label="模板名称" min-width="150" />
        <el-table-column label="通道类型" width="100">
          <template #default="{ row }">
            <el-tag>{{ channelTypeLabel(row.channel_type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="消息类型" width="90">
          <template #default="{ row }">
            <el-tag :type="row.type === 'card' ? 'success' : ''">{{ row.type === 'card' ? '卡片' : '文本' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="title_tpl" label="标题模板" show-overflow-tooltip min-width="200" />
        <el-table-column prop="content_tpl" label="内容模板" show-overflow-tooltip min-width="250" />
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="primary" @click="handleEdit(row)" v-permission="'alert:template:edit'">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)" v-permission="'alert:template:delete'">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 添加/编辑弹窗 -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑模板' : '添加模板'" width="700px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="模板名称" required>
          <el-input v-model="form.name" placeholder="如：飞书告警卡片模板" />
        </el-form-item>
        <el-form-item label="通道类型" required>
          <el-select v-model="form.channel_type" style="width:100%">
            <el-option label="飞书" value="feishu" />
            <el-option label="钉钉" value="dingtalk" />
            <el-option label="自定义Webhook" value="webhook" />
            <el-option label="企业微信" value="wecom" />
            <el-option label="邮件（预留）" value="email" disabled />
          </el-select>
        </el-form-item>
        <el-form-item label="消息类型" required>
          <el-select v-model="form.type" style="width:100%">
            <el-option label="交互式卡片" value="card" />
            <el-option label="纯文本" value="text" />
          </el-select>
        </el-form-item>
        <el-form-item label="标题模板">
          <el-input v-model="form.title_tpl" placeholder="{{.status_emoji}} [{{.status_cn}}] {{.alert_name}}" />
        </el-form-item>
        <el-form-item label="内容模板" required>
          <el-input v-model="form.content_tpl" type="textarea" :rows="8" placeholder="使用 {{.变量名}} 引用告警字段" />
        </el-form-item>
        <el-form-item label="可用变量">
          <div class="var-hints">
            <el-tag v-for="v in availableVars" :key="v" size="small" style="margin:2px;cursor:pointer;"
              @click="insertVar(v)">{{ v }}</el-tag>
          </div>
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
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import * as alertApi from '@/api/alert'

const loading = ref(false)
const list = ref([])
const sources = ref([])
const selectedSourceId = ref(null)
const dialogVisible = ref(false)
const isEdit = ref(false)

const form = ref({
  name: '', source_id: null, channel_type: 'feishu', type: 'card',
  title_tpl: '', content_tpl: ''
})

const availableVars = [
  '{{.alert_name}}', '{{.status}}', '{{.status_cn}}', '{{.status_emoji}}',
  '{{.severity}}', '{{.severity_cn}}', '{{.severity_emoji}}',
  '{{.severity_color}}', '{{.status_color}}',
  '{{.instance}}', '{{.summary}}', '{{.description}}',
  '{{.starts_at}}', '{{.current_time}}'
]

const channelTypeLabel = (t) => ({ feishu: '飞书', webhook: 'Webhook', dingtalk: '钉钉', wecom: '企微', email: '邮件' })[t] || t

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
    const res = await alertApi.getAlertTemplates({ source_id: selectedSourceId.value })
    list.value = res.data || []
  } catch (e) { /* handled */ }
  loading.value = false
}

const showAddDialog = () => {
  isEdit.value = false
  form.value = {
    name: '', source_id: selectedSourceId.value, channel_type: 'feishu',
    type: 'card', title_tpl: '', content_tpl: ''
  }
  dialogVisible.value = true
}

const handleEdit = (row) => {
  isEdit.value = true
  form.value = { ...row }
  dialogVisible.value = true
}

const insertVar = (v) => {
  form.value.content_tpl += v
}

const handleSubmit = async () => {
  if (!form.value.name || !form.value.content_tpl) {
    ElMessage.warning('请填写必填项')
    return
  }
  try {
    if (isEdit.value) {
      await alertApi.updateAlertTemplate(form.value.id, form.value)
      ElMessage.success('更新成功')
    } else {
      await alertApi.createAlertTemplate(form.value)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchList()
  } catch (e) { /* handled */ }
}

const handleDelete = async (row) => {
  await ElMessageBox.confirm('确定删除该模板吗？', '确认删除', { type: 'warning' })
  try {
    await alertApi.deleteAlertTemplate(row.id)
    ElMessage.success('删除成功')
    fetchList()
  } catch (e) { /* handled */ }
}

onMounted(loadSources)
</script>

<style scoped>
.card-header { display: flex; justify-content: space-between; align-items: center; }
.header-left { display: flex; align-items: center; }
.title { font-size: 16px; font-weight: bold; }
.var-hints { line-height: 2; }
</style>
