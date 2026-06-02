<template>
  <el-container class="layout-container">
    <!-- 顶部全局栏 -->
    <el-header class="global-header" height="60px">
      <div class="header-left">
        <div class="logo">
          <span class="logo-icon">🚀</span>
          <span class="logo-text" v-if="!isCollapse">Alert-Ops</span>
        </div>
        
      </div>

      <div class="header-right">
        <el-input v-model="searchQuery" placeholder="搜索资源 (Ctrl+K)" prefix-icon="Search" class="global-search" clearable />
        <el-dropdown @command="handleCommand">
          <span class="user-info">
            <el-avatar :size="32" :icon="UserFilled" />
            <span class="username">{{ userStore.userInfo.username }}</span>
            <el-icon><ArrowDown /></el-icon>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="logout" divided>退出登录</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </el-header>

    <el-container class="main-container">
      <!-- 侧边功能菜单 (动态生成) -->
      <el-aside :width="isCollapse ? '64px' : '220px'" class="sidebar">
        <div class="collapse-btn" @click="toggleCollapse">
          <el-icon><Fold v-if="!isCollapse" /><Expand v-else /></el-icon>
        </div>
        
        <el-menu
          :default-active="activeMenu"
          :collapse="isCollapse"
          :collapse-transition="false"
          router
          background-color="#304156"
          text-color="#bfcbd9"
          active-text-color="#409EFF"
        >
          <!-- 递归渲染菜单组件 -->
          <menu-item v-for="menu in menuTree" :key="menu.id" :menu="menu" />
        </el-menu>
      </el-aside>

      <!-- 主内容区 -->
      <el-main class="main-content">
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { computed, ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { ElMessageBox } from 'element-plus'
import { UserFilled, ArrowDown, Fold, Expand, Search } from '@element-plus/icons-vue'
import MenuItem from './components/MenuItem.vue' // 我们需要创建一个递归组件

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const isCollapse = ref(false)
const searchQuery = ref('')
const activeMenu = computed(() => route.path)

// 构建菜单树
const menuTree = computed(() => {
  const menus = userStore.menus || []
  const map = {}
  const roots = []
  
  // 1. 初始化所有菜单节点
  menus.forEach(m => {
    map[m.id] = { ...m, children: [] }
  })
  
  // 2. 构建父子关系
  menus.forEach(m => {
    if (m.parent_id === 0) {
      roots.push(map[m.id])
    } else if (map[m.parent_id]) {
      map[m.parent_id].children.push(map[m.id])
    }
  })
  
  // 3. 递归排序函数：对每个节点及其子节点进行排序
  function sortMenuTree(menuList) {
    // 按 sort 字段升序排序
    menuList.sort((a, b) => (a.sort || 0) - (b.sort || 0))
    
    // 递归排序子菜单
    menuList.forEach(menu => {
      if (menu.children && menu.children.length > 0) {
        sortMenuTree(menu.children)
      }
    })
    
    return menuList
  }
  
  // 4. 对根菜单进行排序（会自动递归排序所有子菜单）
  return sortMenuTree(roots)
})

const toggleCollapse = () => isCollapse.value = !isCollapse.value


const handleCommand = async (command) => {
  if (command === 'logout') {
    await ElMessageBox.confirm('确定退出吗？', '提示', { type: 'warning' })
    userStore.logout()
    router.push('/login')
  }
}

</script>

<style scoped>
.layout-container { 
  height: 100vh; 
  display: flex; 
  flex-direction: column; 
}

/* 顶部全局栏样式 */
.global-header { 
  background: #fff; 
  border-bottom: 1px solid #dcdfe6; 
  display: flex; 
  justify-content: space-between; 
  align-items: center; 
  padding: 0 20px; 
  z-index: 10;
  box-shadow: 0 1px 4px rgba(0, 21, 41, 0.08);
  height: 60px !important; /* 强制高度 */
}

.header-left { 
  display: flex; 
  align-items: center; 
  gap: 20px; 
  flex: 1;
}

.logo { 
  display: flex; 
  align-items: center; 
  gap: 10px; 
  font-size: 18px;
  font-weight: 600; 
  color: #304156; 
  white-space: nowrap;
}

.logo-icon {
  font-size: 24px;
}

.cluster-selector { 
  width: 220px; 
}

/* 修复右侧重叠问题 */
.header-right { 
  display: flex; 
  align-items: center; 
  gap: 15px; 
  flex-shrink: 0; /* 防止被压缩 */
}

.global-search { 
  width: 250px; 
}

/* 修复用户信息重叠 */
.user-info {
  display: flex; 
  align-items: center; 
  gap: 6px; 
  cursor: pointer; 
  padding: 6px 10px; 
  border-radius: 20px;
  transition: background-color 0.3s;
  white-space: nowrap;
}

.user-info:hover {
  background-color: #f5f7fa;
}

.username {
  font-size: 14px;
  color: #606266;
}

/* 主体容器 */
.main-container { 
  flex: 1; 
  overflow: hidden; 
  display: flex;
}

.sidebar { 
  background: #304156; 
  transition: width 0.3s; 
  display: flex;
  flex-direction: column;
  overflow-y: auto;
}

.collapse-btn { 
  height: 40px; 
  display: flex; 
  align-items: center; 
  justify-content: center; 
  color: #bfcbd9; 
  cursor: pointer; 
  border-bottom: 1px solid #2b3a4d;
}

.collapse-btn:hover {
  background-color: #263445;
  color: #fff;
}

/* 菜单样式调整 */
:deep(.el-menu) { 
  border-right: none; 
}

:deep(.el-sub-menu__title), :deep(.el-menu-item) {
  display: flex !important;
  align-items: center;
}

.main-content { 
  background: #f0f2f5; 
  padding: 20px; 
  overflow-y: auto; 
  flex: 1;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>