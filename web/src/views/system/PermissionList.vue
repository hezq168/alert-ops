<template>
  <div class="permission-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span class="title">权限管理</span>
          <el-button type="primary" @click="showAddDialog" v-permission="'system:permission:add'">
            <el-icon><Plus /></el-icon>
            新增权限
          </el-button>
        </div>
      </template>

      <el-row :gutter="20">
        <!-- 左侧：角色列表 -->
        <el-col :span="6">
          <el-card shadow="never">
            <template #header>
              <span>选择角色</span>
            </template>
            <el-menu :default-active="String(selectedRoleId)" @select="handleSelectRole">
              <el-menu-item 
                v-for="role in roles" 
                :key="role.id" 
                :index="String(role.id)"
              >
                {{ role.name }}
              </el-menu-item>
            </el-menu>
          </el-card>
        </el-col>

        <!-- 右侧：权限配置 -->
        <el-col :span="18">
          <el-card shadow="never" v-if="selectedRoleId">
            <template #header>
              <div class="card-header">
                <span>配置权限 - {{ currentRole?.name }}</span>
                <el-button type="primary" @click="handleSave" :loading="saving" v-permission="'system:role:assign-permission'">
                  保存
                </el-button>
              </div>
            </template>

            <el-tree
              ref="treeRef"
              :data="permissionTree"
              show-checkbox
              node-key="id"
              :props="{ label: 'name', children: 'children' }"
              :default-checked-keys="checkedKeys"
              style="max-height: 500px; overflow-y: auto"
            >
              <template #default="{ node, data }">
                <span class="custom-tree-node">
                  <span>{{ node.label }}</span>
                  <span>{{ data.api_path }}</span>
                  <span>{{ data.api_method }}</span>
                  <span>
                    <el-tag size="small" :type="data.type === 'menu' ? '' : 'warning'" style="margin-left: 8px">
                      {{ data.type === 'menu' ? '菜单' : '按钮' }}
                    </el-tag>
                    <el-button 
                      link 
                      type="primary" 
                      size="small" 
                      @click.stop="handleEditPermission(data)"
                      style="margin-left: 8px" v-permission="'system:permission:edit'"
                    >
                      编辑
                    </el-button>
                    <el-button 
                      link 
                      type="danger" 
                      size="small" 
                      @click.stop="handleDeletePermission(data)"
                      style="margin-left: 8px" v-permission="'system:permission:delete'"
                    >
                      删除
                    </el-button>
                  </span>
                </span>
              </template>
            </el-tree>
          </el-card>
          
          <el-empty v-else description="请选择一个角色" />
        </el-col>
      </el-row>

      <!-- 新增/编辑权限对话框 -->
      <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑权限' : '新增权限'" width="500px">
        <el-form :model="permissionForm" label-width="100px">
          <el-form-item label="权限名称" required>
            <el-input v-model="permissionForm.name" placeholder="例如：用户管理" />
          </el-form-item>
          <el-form-item label="权限编码" required>
            <el-input v-model="permissionForm.code" placeholder="例如：system:user:list" />
            <div class="form-tip">格式：模块:功能:操作，如 system:user:add</div>
          </el-form-item>
          <el-form-item label="权限类型" required>
            <el-radio-group v-model="permissionForm.type">
              <el-radio label="menu">菜单</el-radio>
              <el-radio label="button">按钮</el-radio>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="父级权限">
            <el-select v-model="permissionForm.parent_id" placeholder="选择父级" clearable style="width: 100%">
              <el-option
                v-for="perm in parentPermissions"
                :key="perm.id"
                :label="perm.name"
                :value="perm.id"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="路由路径" v-if="permissionForm.type === 'menu'">
            <el-input v-model="permissionForm.path" placeholder="例如：/system/users" />
          </el-form-item>
          <el-form-item label="图标" v-if="permissionForm.type === 'menu'">
            <el-input v-model="permissionForm.icon" placeholder="例如：User" />
          </el-form-item>
          <el-form-item label="API请求方法">
            <el-input v-model="permissionForm.api_method" placeholder="例如：GET,POST,PUT,DELETE" />
          </el-form-item>         
          <el-form-item label="API路由">
            <el-input v-model="permissionForm.api_path" placeholder="例如：/api/system/users" />
          </el-form-item>
                 
          <el-form-item label="排序">
            <el-input-number v-model="permissionForm.sort" :min="0" :max="999" style="width: 100%" />
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSubmit" :loading="submitLoading">确定</el-button>
        </template>
      </el-dialog>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getRoles } from '@/api/user'
import { getAllPermissions, getRolePermissions, assignPermissions, createPermission, updatePermission, deletePermission } from '@/api/permission'

const roles = ref([])
const selectedRoleId = ref(null)
const currentRole = ref(null)
const allPermissions = ref([])
const rolePermissions = ref([])
const treeRef = ref(null)
const saving = ref(false)

// 对话框相关
const dialogVisible = ref(false)
const submitLoading = ref(false)
const isEdit = ref(false)
const permissionForm = ref({
  id: null,
  name: '',
  code: '',
  type: 'menu',
  parent_id: null,
  path: '',
  icon: '',
  sort: 0,
  api_path: '',
  api_method: ''
})

// 可选的父级权限（排除按钮类型和当前编辑的权限）
const parentPermissions = computed(() => {
  let perms = allPermissions.value.filter(p => p.type === 'menu')
  
  // 如果是编辑模式，排除当前权限本身（避免循环引用）
  if (isEdit.value && permissionForm.value.id) {
    perms = perms.filter(p => p.id !== permissionForm.value.id)
  }
  
  return perms
})

// 将权限列表转换为树形结构
const permissionTree = computed(() => {
  const map = {}
  const tree = []
  
  // 先创建所有节点的映射
  allPermissions.value.forEach(perm => {
    map[perm.id] = { ...perm, children: [] }
  })
  
  // 构建树形结构
  allPermissions.value.forEach(perm => {
    // 防止循环引用：如果父级是自己，强制设为顶级
    if (perm.parent_id === perm.id) {
      perm.parent_id = 0
    }
    
    if (perm.parent_id === 0 || !map[perm.parent_id]) {
      tree.push(map[perm.id])
    } else {
      const parent = map[perm.parent_id]
      if (parent) {
        parent.children.push(map[perm.id])
      }
    }
  })
  
  return tree
})

// 已选中的权限 ID
const checkedKeys = computed(() => {
  return rolePermissions.value.map(p => p.id)
})

// 监听选中权限变化，更新树的选中状态
watch(checkedKeys, (newKeys) => {
  if (treeRef.value) {
    // 清除所有选中
    treeRef.value.setCheckedKeys([])
    // 设置新的选中状态
    treeRef.value.setCheckedKeys(newKeys)
  }
}, { deep: true })

// 加载角色列表
const loadRoles = async () => {
  try {
    const res = await getRoles()
    roles.value = res.data.list || res.data || []
  } catch (error) {
    console.error(error)
  }
}

// 加载所有权限
const loadAllPermissions = async () => {
  try {
    const res = await getAllPermissions()
    allPermissions.value = res.data || []
  } catch (error) {
    console.error(error)
  }
}

// 选择角色
const handleSelectRole = async (roleId) => {
  selectedRoleId.value = Number(roleId)
  currentRole.value = roles.value.find(r => r.id === selectedRoleId.value)
  
  // 加载该角色的权限
  try {
    const res = await getRolePermissions(selectedRoleId.value)
    rolePermissions.value = res.data || []
  } catch (error) {
    console.error(error)
  }
}

// 保存权限
const handleSave = async () => {
  if (!treeRef.value) return
  
  saving.value = true
  try {
    // 只获取完全选中的节点（不包括半选中）
    const checkedKeys = treeRef.value.getCheckedKeys()
    
    await assignPermissions(selectedRoleId.value, checkedKeys)
    ElMessage.success('保存成功')
    
    // 重新加载
    const res = await getRolePermissions(selectedRoleId.value)
    rolePermissions.value = res.data || []
  } catch (error) {
    console.error(error)
  } finally {
    saving.value = false
  }
}

// 显示新增对话框
const showAddDialog = () => {
  isEdit.value = false
  permissionForm.value = {
    id: null,
    name: '',
    code: '',
    type: 'menu',
    parent_id: null,
    path: '',
    icon: '',
    sort: 0,
    api_path: '',
    api_method: ''
  }
  dialogVisible.value = true
}

// 显示编辑对话框
const handleEditPermission = (permission) => {
  isEdit.value = true
  permissionForm.value = {
    id: permission.id,
    name: permission.name,
    code: permission.code,
    type: permission.type,
    parent_id: permission.parent_id || null,
    path: permission.path || '',
    icon: permission.icon || '',
    sort: permission.sort || 0,
    api_path: permission.api_path || '',
    api_method: permission.api_method || ''
  }
  dialogVisible.value = true
}

// 提交表单
const handleSubmit = async () => {
  // 简单验证
  if (!permissionForm.value.name || !permissionForm.value.code) {
    ElMessage.warning('请填写必填项')
    return
  }

  submitLoading.value = true
  try {
    if (isEdit.value) {
      // 编辑
      await updatePermission(permissionForm.value.id, permissionForm.value)
      ElMessage.success('更新成功')
    } else {
      // 新增
      await createPermission(permissionForm.value)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    
    // 重新加载权限列表
    await loadAllPermissions()
  } catch (error) {
    console.error(error)
  } finally {
    submitLoading.value = false
  }
}

// 删除权限
const handleDeletePermission = async (permission) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除权限 "${permission.name}" 吗？`,
      '提示',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    await deletePermission(permission.id)
    ElMessage.success('删除成功')
    
    // 重新加载
    await loadAllPermissions()
  } catch (error) {
    if (error !== 'cancel') {
      console.error(error)
    }
  }
}

onMounted(() => {
  loadRoles()
  loadAllPermissions()
})
</script>

<style scoped>
.permission-list {
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

.custom-tree-node {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 14px;
  padding-right: 8px;
}

.form-tip {
  font-size: 12px;
  color: #999;
  margin-top: 5px;
}
</style>