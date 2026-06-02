<template>
  <div class="alert-source-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span class="title">告警源管理</span>
          <el-button type="primary" @click="showAddDialog" v-permission="'alert:source:add'">
            <el-icon><Plus /></el-icon>添加告警源
          </el-button>
        </div>
      </template>

      <el-table :data="list" stripe v-loading="loading">
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="name" label="名称" min-width="150" />
        <el-table-column label="Webhook URL" min-width="280">
          <template #default="{ row }">
            <el-tag type="info" style="font-family: monospace; font-size: 12px;">
              POST /api/v1/webhook/alertmanager/{{ row.slug }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="type" label="类型" width="130">
          <template #default="{ row }">
            <el-tag>{{ row.type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" show-overflow-tooltip min-width="150" />
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.enabled ? 'success' : 'danger'" size="small">
              {{ row.enabled ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="170">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="primary" @click="handleEdit(row)" v-permission="'alert:source:edit'">编辑</el-button>
            <el-button size="small" type="warning" @click="goToRules(row)">规则</el-button>
            <el-button size="small" type="success" @click="goToChannels(row)">通道</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)" v-permission="'alert:source:delete'">删除</el-button>
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

    <!-- 添加/编辑弹窗 -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑告警源' : '添加告警源'" width="550px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="名称" required>
          <el-input v-model="form.name" placeholder="如：生产环境-Alertmanager" />
        </el-form-item>
        <el-form-item label="唯一标识(slug)" required>
          <el-input v-model="form.slug" placeholder="如：prod" :disabled="isEdit" />
          <div style="color:#999;font-size:12px;margin-top:4px;">
            用于生成 webhook URL：/api/v1/webhook/alertmanager/{{ form.slug || 'slug' }}
          </div>
        </el-form-item>
        <el-form-item label="类型" required>
          <el-select v-model="form.type" style="width:100%">
            <el-option label="Alertmanager" value="alertmanager" />
            <el-option label="阿里云云监控" value="aliyun" disabled />
            <el-option label="腾讯云监控" value="tencent" disabled />
            <el-option label="AWS CloudWatch" value="aws" disabled />
            <el-option label="Zabbix" value="zabbix" disabled />
          </el-select>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="2" placeholder="告警源描述" />
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="form.enabled" />
        </el-form-item>
        <el-form-item label="继续匹配">
          <el-switch v-model="form.continue_match" />
          <span style="margin-left:8px;color:#999;font-size:12px;">开启后匹配到规则仍继续往下匹配，关闭则命中第一条后停止</span>
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
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import * as alertApi from '@/api/alert'

const router = useRouter()
const loading = ref(false)
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const dialogVisible = ref(false)
const isEdit = ref(false)
const form = ref({
  name: '', slug: '', type: 'alertmanager', description: '', enabled: true, continue_match: false
})

const fetchList = async () => {
  loading.value = true
  try {
    const res = await alertApi.getAlertSources({ page: page.value, page_size: pageSize.value })
    list.value = res.data.list || []
    total.value = res.data.total || 0
  } catch (e) { /* handled by interceptor */ }
  loading.value = false
}

const showAddDialog = () => {
  isEdit.value = false
  form.value = { name: '', slug: '', type: 'alertmanager', description: '', enabled: true, continue_match: false }
  dialogVisible.value = true
}

const handleEdit = (row) => {
  isEdit.value = true
  form.value = { ...row }
  dialogVisible.value = true
}

const handleSubmit = async () => {
  if (!form.value.name || !form.value.slug || !form.value.type) {
    ElMessage.warning('请填写必填项')
    return
  }
  try {
    if (isEdit.value) {
      await alertApi.updateAlertSource(form.value.id, form.value)
      ElMessage.success('更新成功')
    } else {
      await alertApi.createAlertSource(form.value)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchList()
  } catch (e) { /* handled */ }
}

const handleDelete = async (row) => {
  await ElMessageBox.confirm('确定删除该告警源吗？删除后关联的规则、模板、通道也将无法使用。', '确认删除', { type: 'warning' })
  try {
    await alertApi.deleteAlertSource(row.id)
    ElMessage.success('删除成功')
    fetchList()
  } catch (e) { /* handled */ }
}

const goToRules = (row) => {
  router.push({ path: '/alert/rules', query: { source_id: row.id, source_name: row.name } })
}

const goToChannels = (row) => {
  router.push({ path: '/alert/channels', query: { source_id: row.id, source_name: row.name } })
}

const formatTime = (t) => {
  if (!t) return '-'
  return new Date(t).toLocaleString('zh-CN')
}

onMounted(fetchList)
</script>

<style scoped>
.card-header { display: flex; justify-content: space-between; align-items: center; }
.title { font-size: 16px; font-weight: bold; }
.pagination { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
