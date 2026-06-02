<template>
  <div class="alert-channel-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <div class="header-left">
            <span class="title">发送通道</span>
            <el-select v-model="selectedSourceId" placeholder="选择告警源" @change="fetchList" style="width:220px;margin-left:12px;" clearable>
              <el-option v-for="s in sources" :key="s.id" :label="s.name" :value="s.id" />
            </el-select>
          </div>
          <el-button type="primary" @click="showAddDialog" :disabled="!selectedSourceId" v-permission="'alert:channel:add'">
            <el-icon><Plus /></el-icon>添加通道
          </el-button>
        </div>
      </template>

      <el-table :data="list" stripe v-loading="loading">
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="name" label="通道名称" min-width="150" />
        <el-table-column label="类型" width="100">
          <template #default="{ row }">
            <el-tag :type="row.type === 'feishu' ? 'success' : ''">{{ channelTypeLabel(row.type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="webhook_url" label="Webhook地址" show-overflow-tooltip min-width="300" />
        <el-table-column label="签名" width="80">
          <template #default="{ row }">
            <el-tag :type="row.secret ? 'success' : 'info'" size="small">{{ row.secret ? '已配置' : '未配置' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="70">
          <template #default="{ row }">
            <el-tag :type="row.enabled ? 'success' : 'danger'" size="small">{{ row.enabled ? '启用' : '禁用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="170">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="success" @click="handleTest(row)">测试</el-button>
            <el-button size="small" type="primary" @click="handleEdit(row)" v-permission="'alert:channel:edit'">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)" v-permission="'alert:channel:delete'">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 添加/编辑弹窗 -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑通道' : '添加通道'" width="550px">
      <el-form :model="form" label-width="110px">
        <el-form-item label="通道名称" required>
          <el-input v-model="form.name" placeholder="如：生产告警飞书群" />
        </el-form-item>
        <el-form-item label="通道类型" required>
          <el-select v-model="form.type" style="width:100%">
            <el-option label="飞书机器人" value="feishu" />
            <el-option label="自定义 Webhook" value="webhook" />
            <el-option label="钉钉机器人" value="dingtalk" />
            <el-option label="企业微信机器人" value="wecom" />
            <el-option label="邮件（预留）" value="email" disabled />
          </el-select>
        </el-form-item>
        <el-form-item label="Webhook URL" required>
          <el-input v-model="form.webhook_url" placeholder="https://open.feishu.cn/open-apis/bot/v2/hook/xxx" />
          <div v-if="form.type === 'feishu'" style="color:#999;font-size:12px;margin-top:4px;">
            飞书群 → 群设置 → 机器人 → 添加自定义机器人 → 复制 Webhook 地址
          </div>
          <div v-if="form.type === 'dingtalk'" style="color:#999;font-size:12px;margin-top:4px;">
            钉钉群 → 群设置 → 智能群助手 → 添加机器人 → 自定义 → 复制 Webhook 地址
          </div>
          <div v-if="form.type === 'wecom'" style="color:#999;font-size:12px;margin-top:4px;">
            企业微信群 → 群设置 → 群机器人 → 添加 → 复制 Webhook 地址
          </div>
        </el-form-item>
        <el-form-item label="签名密钥">
          <el-input v-model="form.secret" :placeholder="form.type === 'dingtalk' ? '钉钉机器人安全设置中的加签密钥' : '飞书机器人安全设置中的签名校验密钥'" show-password />
          <div style="color:#999;font-size:12px;margin-top:4px;">可选，用于签名校验，增强安全性</div>
        </el-form-item>
        <el-form-item label="@提醒手机号" v-if="form.type === 'dingtalk'">
          <el-input v-model="form.at_mobiles" placeholder="多个用英文逗号分隔，如：13800138000,13900139000" />
          <div style="color:#999;font-size:12px;margin-top:4px;">选填，消息发送时自动 @ 这些手机号对应的群成员</div>
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
  name: '', source_id: null, type: 'feishu',
  webhook_url: '', secret: '', enabled: true, at_mobiles: ''
})

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
    const res = await alertApi.getAlertChannels({ source_id: selectedSourceId.value })
    list.value = res.data || []
  } catch (e) { /* handled */ }
  loading.value = false
}

const showAddDialog = () => {
  isEdit.value = false
  form.value = { name: '', source_id: selectedSourceId.value, type: 'feishu', webhook_url: '', secret: '', enabled: true, at_mobiles: '' }
  dialogVisible.value = true
}

const handleEdit = (row) => {
  isEdit.value = true
  form.value = { ...row, secret: '', at_mobiles: '' }
  // 从 Config 回显 @ 手机号
  if (row.type === 'dingtalk' && row.config) {
    try {
      const cfg = JSON.parse(row.config)
      form.value.at_mobiles = (cfg.at_mobiles || []).join(',')
    } catch (e) { /* ignore */ }
  }
  dialogVisible.value = true
}

const handleSubmit = async () => {
  if (!form.value.name || !form.value.webhook_url) {
    ElMessage.warning('请填写必填项')
    return
  }
  // 将 at_mobiles 序列化到 Config 字段
  if (form.value.type === 'dingtalk' && form.value.at_mobiles) {
    const mobiles = form.value.at_mobiles.split(',').map(s => s.trim()).filter(Boolean)
    form.value.config = JSON.stringify({ at_mobiles: mobiles })
  }
  try {
    if (isEdit.value) {
      await alertApi.updateAlertChannel(form.value.id, form.value)
      ElMessage.success('更新成功')
    } else {
      await alertApi.createAlertChannel(form.value)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchList()
  } catch (e) { /* handled */ }
}

const handleTest = async (row) => {
  try {
    await ElMessageBox.confirm(`确定向通道「${row.name}」发送一条测试消息吗？`, '测试发送', { type: 'info' })
  } catch {
    return // 用户取消
  }
  try {
    await alertApi.testAlertChannel(row.id)
    ElMessage.success(`测试消息已成功发送到「${row.name}」`)
  } catch (e) {
    ElMessage.error('测试发送失败：' + (e?.message || '未知错误'))
  }
}

const handleDelete = async (row) => {
  await ElMessageBox.confirm('确定删除该通道吗？', '确认删除', { type: 'warning' })
  try {
    await alertApi.deleteAlertChannel(row.id)
    ElMessage.success('删除成功')
    fetchList()
  } catch (e) { /* handled */ }
}

const formatTime = (t) => {
  if (!t) return '-'
  return new Date(t).toLocaleString('zh-CN')
}

onMounted(loadSources)
</script>

<style scoped>
.card-header { display: flex; justify-content: space-between; align-items: center; }
.header-left { display: flex; align-items: center; }
.title { font-size: 16px; font-weight: bold; }
</style>
