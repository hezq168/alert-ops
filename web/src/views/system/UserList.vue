<template>
  <div class="user-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span class="title">用户管理</span>
          <el-button type="primary" @click="showAddDialog" v-permission="'system:user:add'">
            <el-icon><Plus /></el-icon>
            添加用户
          </el-button>
        </div>
      </template>

      <el-table :data="users" stripe v-loading="loading">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="username" label="用户名" width="150" />
        <el-table-column prop="nickname" label="昵称" width="150" />
        <el-table-column prop="email" label="邮箱" show-overflow-tooltip />
        <el-table-column prop="phone" label="手机号" show-overflow-tooltip />

        <!-- 角色 -->
        <el-table-column label="角色" width="200">
          <template #default="{ row }">
            <el-tag
              v-for="role in row.roles"
              :key="role.id"
              size="small"
              style="margin-right: 5px"
            >
              {{ role.name }}
            </el-tag>

            <span v-if="!row.roles || row.roles.length === 0" style="color: #999">
              未分配
            </span>
          </template>
        </el-table-column>

        <!-- 状态 -->
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">
              {{ row.status === 1 ? '正常' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>

        <!-- 操作 -->
        <el-table-column label="操作" width="320">
          <template #default="{ row }">
            <el-button size="small" type="primary" @click="handleEdit(row)" v-permission="'system:user:edit'">
              编辑
            </el-button>

            <el-button size="small" type="success" @click="showAssignRoleDialog(row)" v-permission="'system:user:assign-role'">
              分配角色
            </el-button>

            <el-button
              size="small"
              :type="row.status === 1 ? 'warning' : 'success'"
              @click="toggleStatus(row)" v-permission="'system:user:toggle-status'"
            >
              {{ row.status === 1 ? '禁用' : '启用' }}
            </el-button>

            <el-button size="small" type="danger" @click="handleDelete(row.id)" v-permission="'system:user:delete'">
              删除
            </el-button>
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
          @size-change="loadUsers"
          @current-change="loadUsers"
        />
      </div>

      <!-- 用户弹窗 -->
      <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑用户' : '添加用户'" width="500px">
        <el-form :model="userForm" label-width="80px">
          <el-form-item label="用户名" required>
            <el-input v-model="userForm.username" :disabled="isEdit" />
          </el-form-item>

          <el-form-item label="密码" v-if="!isEdit" required>
            <el-input v-model="userForm.password" type="password" />
          </el-form-item>

          <el-form-item label="昵称">
            <el-input v-model="userForm.nickname" />
          </el-form-item>

          <el-form-item label="邮箱">
            <el-input v-model="userForm.email" />
          </el-form-item>

          <el-form-item label="手机号">
            <el-input v-model="userForm.phone" />
          </el-form-item>
        </el-form>

        <template #footer>
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSubmit" :loading="submitLoading">
            确定
          </el-button>
        </template>
      </el-dialog>

      <!-- 角色弹窗 -->
      <el-dialog v-model="roleDialogVisible" title="分配角色" width="500px">
        <div v-if="currentUser" style="margin-bottom: 15px">
          <p>用户：<strong>{{ currentUser.username }}</strong></p>

          <p>当前角色：</p>
          <div style="margin-top: 10px">
            <el-tag
              v-for="role in currentUser.roles"
              :key="role.id"
              size="small"
              closable
              @close="handleRemoveRole(role)"
              style="margin-right: 5px"
            >
              {{ role.name }}
            </el-tag>

            <span v-if="!currentUser.roles || currentUser.roles.length === 0" style="color:#999">
              未分配
            </span>
          </div>
        </div>

        <el-divider />

        <el-form label-width="80px">
          <el-form-item label="添加角色">
            <el-select v-model="selectedRoleId" style="width: 100%">
              <el-option
                v-for="role in availableRoles"
                :key="role.id"
                :label="role.name"
                :value="role.id"
              />
            </el-select>
          </el-form-item>
        </el-form>

        <template #footer>
          <el-button @click="roleDialogVisible = false">关闭</el-button>
          <el-button type="primary" @click="handleAssignRole" :loading="assignLoading">
            添加
          </el-button>
        </template>
      </el-dialog>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue' // 确保引入了 Plus 图标
import {
  getUsers,
  createUser,
  updateUser,
  deleteUser,
  updateUserStatus,
  getRoles,
  assignRole,
  getUserRoles,
  removeUserRole
} from '@/api/user'


const loading = ref(false)
const submitLoading = ref(false)
const assignLoading = ref(false)

const dialogVisible = ref(false)
const roleDialogVisible = ref(false)

const isEdit = ref(false)

const users = ref([])
const allRoles = ref([])

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const currentUser = ref(null)
const selectedRoleId = ref(null)

const userForm = reactive({
  id: null,
  username: '',
  password: '',
  nickname: '',
  email: '',
  phone: ''
})

/* 可用角色 */
const availableRoles = computed(() => {
  if (!currentUser.value?.roles) return allRoles.value
  const ids = currentUser.value.roles.map(r => r.id)
  return allRoles.value.filter(r => !ids.includes(r.id))
})

/* ===== 用户列表（拉取数据） ===== */
const loadUsers = async () => {
  loading.value = true
  try {
    const res = await getUsers({ page: pagination.page, page_size: pagination.pageSize })
    users.value = res.data.list || []
    pagination.total = res.data.total || 0
  } finally {
    loading.value = false
  }
}

/* 获取所有角色（下拉列表用，不分页） */
const loadAllRoles = async () => {
  const res = await getRoles({ page: 1, page_size: 999 })
  allRoles.value = res.data.list || res.data || []
}

/* 新增弹窗 */
const showAddDialog = () => {
  isEdit.value = false
  Object.assign(userForm, {
    id: null,
    username: '',
    password: '',
    nickname: '',
    email: '',
    phone: ''
  })
  dialogVisible.value = true
}

/* 编辑弹窗 */
const handleEdit = (row) => {
  isEdit.value = true
  Object.assign(userForm, {
    id: row.id,
    username: row.username,
    password: '',
    nickname: row.nickname,
    email: row.email,
    phone: row.phone
  })
  dialogVisible.value = true
}

/* 提交表单（新增/修改） */
const handleSubmit = async () => {
  if (!userForm.username) return ElMessage.warning('请输入用户名')
  if (!isEdit.value && !userForm.password) return ElMessage.warning('请输入密码')

  submitLoading.value = true
  try {
    if (isEdit.value) {
      await updateUser(userForm.id, userForm)
      ElMessage.success('更新成功')
    } else {
      await createUser(userForm)
      ElMessage.success('添加成功')
      pagination.page = 1
    }
    dialogVisible.value = false
    loadUsers()
  } finally {
    submitLoading.value = false
  }
}

/* 分配角色弹窗 */
const showAssignRoleDialog = (user) => {
  currentUser.value = { ...user, roles: [...(user.roles || [])] }
  selectedRoleId.value = null
  roleDialogVisible.value = true
}

/* 分配角色 */
const handleAssignRole = async () => {
  if (!selectedRoleId.value) return ElMessage.warning('请选择角色')

  assignLoading.value = true
  try {
    await assignRole(currentUser.value.id, selectedRoleId.value)

    const res = await getUserRoles(currentUser.value.id)
    const newRoles = res.data || []

    currentUser.value.roles = newRoles

    const target = users.value.find(u => u.id === currentUser.value.id)
    if (target) target.roles = newRoles

    ElMessage.success('分配成功')
    selectedRoleId.value = null
  } finally {
    assignLoading.value = false
  }
}

/* 移除角色 */
const handleRemoveRole = async (role) => {
  try {
    await ElMessageBox.confirm('确定移除该角色？', '提示', { type: 'warning' })

    await removeUserRole(currentUser.value.id, role.id)

    const res = await getUserRoles(currentUser.value.id)
    const newRoles = res.data || []

    currentUser.value.roles = newRoles

    const target = users.value.find(u => u.id === currentUser.value.id)
    if (target) target.roles = newRoles
    ElMessage.success('移除成功')
  } catch (error) {
    if (error !== 'cancel') console.error(error)
  }
}

/* 状态切换 —— 修复局部对象污染导致的标签重复渲染 */
const toggleStatus = async (row) => {
  const newStatus = row.status === 1 ? 0 : 1

  try {
    await ElMessageBox.confirm(
      `确定要${newStatus === 1 ? '启用' : '禁用'}该用户吗？`, 
      '提示', 
      { type: 'warning' }
    )

    await updateUserStatus(row.id, newStatus)

    // 关键核心：通过重构整行对象引用，强制触发 Element 渲染树的清理与重置
    const targetIndex = users.value.findIndex(u => u.id === row.id)
    if (targetIndex !== -1) {
      users.value[targetIndex] = {
        ...users.value[targetIndex],
        status: newStatus
      }
    }

    ElMessage.success('更新成功')
  } catch (error) {
    if (error !== 'cancel') console.error(error)
  }
}

/* 删除 */
const handleDelete = async (id) => {
  try {
    await ElMessageBox.confirm('确定删除该用户吗？', '提示', { type: 'warning' })
    
    await deleteUser(id)
    ElMessage.success('删除成功')

    // 如果当前页只剩一条且不是第一页，回到上一页
    if (users.value.length === 1 && pagination.page > 1) {
      pagination.page--
    }
    loadUsers()

    if (currentUser.value && currentUser.value.id === id) {
      roleDialogVisible.value = false
      currentUser.value = null
    }
  } catch (error) {
    if (error !== 'cancel') console.error(error)
  }
}


onMounted(() => {
  loadUsers()
  loadAllRoles()
})
</script>

<style scoped>
.user-list {
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