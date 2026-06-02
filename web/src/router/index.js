import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { addDynamicRoutes } from './dynamicRoutes'
import Layout from '@/views/Layout.vue'

// 1. 静态路由（不需要权限的基础路由）
export const constantRoutes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { title: '登录' }
  },
  {
    path: '/403',
    name: 'Forbidden',
    component: () => import('@/views/403.vue'),
    meta: { title: '无权限访问' }
  },
  {
    path: '/404',
    name: 'NotFound',
    component: () => import('@/views/404.vue'),
    meta: { title: '页面不存在' }
  }
]

// 2. 基础布局路由（作为动态路由的容器）
export const layoutRoute = {
  path: '/',
  name: 'Layout', // 必须有 name，方便动态路由挂载子路由
  component: Layout,
  redirect: '/dashboard', // 默认首页，如果后端没配 dashboard，可以改成 /k8s/clusters
  meta: { requiresAuth: true },
  children: [
    // 集群详情   
  ] // 动态路由会挂载到这里
}

const router = createRouter({
  history: createWebHistory(),
  routes: [...constantRoutes, layoutRoute]
})

// 白名单
const whiteList = ['/login', '/403', '/404']
let hasAddedDynamicRoutes = false

router.beforeEach(async (to, from) => {
  const userStore = useUserStore()
  
  // 设置标题
  document.title = to.meta.title ? `${to.meta.title} - Kube-Ops` : 'Kube-Ops'
  
  if (whiteList.includes(to.path)) {
    return true
  }
  
  if (!userStore.isLoggedIn) {
    return { path: '/login', query: { redirect: to.fullPath } }
  }
  
  if (to.path === '/login') {
    return '/'
  }
  
  // 核心：动态添加路由
  if (userStore.isLoggedIn && !hasAddedDynamicRoutes) {
    try {
      if (!userStore.userInfo.id || userStore.menus.length === 0) {
        await userStore.initUserInfo()
      }
      
      if (userStore.menus && userStore.menus.length > 0) {
        await addDynamicRoutes(userStore.menus, router)
        hasAddedDynamicRoutes = true
        userStore.setRoutesAdded(true)
        
        // 重新导航以匹配新路由
        return { ...to, replace: true }
      }
    } catch (error) {
      console.error('动态路由添加失败:', error)
      userStore.logout()
      return '/login'
    }
  }
  
  // 权限检查
  if (to.meta.permission) {
    if (!userStore.hasPermission(to.meta.permission)) {
      return { path: '/403', query: { redirect: to.fullPath } }
    }
  }
  
  return true
})

export default router