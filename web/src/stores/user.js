import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { getUserInfo } from '@/api/auth'

export const useUserStore = defineStore('user', () => {
  const token = ref(localStorage.getItem('token') || '')
  const userInfo = ref(JSON.parse(localStorage.getItem('user') || '{}'))
  const permissions = ref(JSON.parse(localStorage.getItem('permissions') || '[]'))
  const menus = ref(JSON.parse(localStorage.getItem('menus') || '[]'))
  const routesAdded = ref(false) // 标记动态路由是否已添加
  
  const isLoggedIn = computed(() => !!token.value)

  function setToken(newToken) {
    token.value = newToken
    localStorage.setItem('token', newToken)
  }

  function setUserInfo(info) {
    userInfo.value = info
    localStorage.setItem('user', JSON.stringify(info))
  }

  function setPermissions(perms) {
    permissions.value = perms
    localStorage.setItem('permissions', JSON.stringify(perms))
  }

  function setMenus(menuList) {
    menus.value = menuList
    localStorage.setItem('menus', JSON.stringify(menuList))
  }

  function setRoutesAdded(status) {
    routesAdded.value = status
  }

  // 检查是否有某个权限
  function hasPermission(permissionCode) {
    if (!permissionCode) return true
    
    // 如果是超级管理员，拥有所有权限
    if (userInfo.value.is_admin || userInfo.value.roles?.some(r => r.code === 'super_admin')) {
      return true
    }
    
    return permissions.value.some(p => p.code === permissionCode)
  }

  // 检查是否有任意一个权限
  function hasAnyPermission(permissionCodes) {
    if (!permissionCodes || permissionCodes.length === 0) return true
    
    return permissionCodes.some(code => hasPermission(code))
  }

  // 检查是否有所有权限
  function hasAllPermissions(permissionCodes) {
    if (!permissionCodes || permissionCodes.length === 0) return true
    
    return permissionCodes.every(code => hasPermission(code))
  }

  // 初始化用户信息（从后端获取最新数据）
  async function initUserInfo() {
    if (!token.value) return
    
    try {
      const res = await getUserInfo()
      if (res.data) {
        setUserInfo(res.data.user)
        setPermissions(res.data.permissions || [])
        setMenus(res.data.menus || [])
      }
    } catch (error) {
      console.error('获取用户信息失败:', error)
      logout()
    }
  }

  function logout() {
    token.value = ''
    userInfo.value = {}
    permissions.value = []
    menus.value = []
    routesAdded.value = false
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    localStorage.removeItem('permissions')
    localStorage.removeItem('menus')
  }

  return {
    token,
    userInfo,
    permissions,
    menus,
    routesAdded,
    isLoggedIn,
    setToken,
    setUserInfo,
    setPermissions,
    setMenus,
    setRoutesAdded,
    hasPermission,
    hasAnyPermission,
    hasAllPermissions,
    initUserInfo,
    logout
  }
})