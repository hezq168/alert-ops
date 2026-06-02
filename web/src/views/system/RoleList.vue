<template>
  <div class="role-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span class="title">角色管理</span>
          <el-button type="primary" @click="showAddDialog" v-permission="'system:role:add'">
            <el-icon><Plus /></el-icon>
            添加角色
          </el-button>
        </div>
      </template>

      <el-table :data="roles" stripe v-loading="loading">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="角色名称" width="150" />
        <el-table-column prop="code" label="角色编码" width="150" />
        <el-table-column prop="description" label="描述" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150">
          <template #default="{ row }">
            <el-button size="small" type="danger" @click="handleDelete(row.id)" v-permission="'system:role:delete'">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div style="display: flex; justify-content: flex-end; margin-top: 16px">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="loadRoles"
          @current-change="loadRoles"
        />
      </div>

      <!-- 添加角色对话框 -->
      <el-dialog v-model="dialogVisible" title="添加角色" width="500px">
        <el-form :model="roleForm" label-width="80px">
          <el-form-item label="角色名称" required>
            <el-input v-model="roleForm.name" />
          </el-form-item>
          <el-form-item label="角色编码" required>
            <el-input v-model="roleForm.code" />
          </el-form-item>
          <el-form-item label="描述">
            <el-input v-model="roleForm.description" type="textarea" :rows="3" />
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSubmit">确定</el-button>
        </template>
      </el-dialog>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getRoles, createRole, deleteRole } from '@/api/user'

const loading = ref(false)
const dialogVisible = ref(false)
const roles = ref([])

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const roleForm = reactive({
  name: '',
  code: '',
  description: ''
})

const loadRoles = async () => {
  loading.value = true
  try {
    const res = await getRoles({ page: pagination.page, page_size: pagination.pageSize })
    roles.value = res.data.list || []
    pagination.total = res.data.total || 0
  } catch (error) {
    console.error(error)
  } finally {
    loading.value = false
  }
}

const showAddDialog = () => {
  Object.assign(roleForm, {
    name: '',
    code: '',
    description: ''
  })
  dialogVisible.value = true
}

const handleSubmit = async () => {
  try {
    await createRole(roleForm)
    ElMessage.success('添加成功')
    dialogVisible.value = false
    pagination.page = 1
    loadRoles()
  } catch (error) {
    console.error(error)
  }
}

const handleDelete = async (id) => {
  try {
    await ElMessageBox.confirm('确定要删除这个角色吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    await deleteRole(id)
    
    ElMessage.success('删除成功')

    // 如果当前页只剩一条且不是第一页，回到上一页
    if (roles.value.length === 1 && pagination.page > 1) {
      pagination.page--
    }
    loadRoles()
  } catch (error) {
    if (error !== 'cancel') {
      console.error(error)
    }
  }
}

onMounted(() => {
  loadRoles()
})
</script>

<style scoped>
.role-list {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.title {
  font-size: 18px;
  font-weight: bold;
}
</style>