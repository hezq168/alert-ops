/**
 * 动态路由管理器
 */

import router from '@/router'
import { useUserStore } from '@/stores/user'

class DynamicRouteManager {
  constructor() {
    this.addedRoutes = []
  }

  /**
   * 添加动态路由
   * @param {Array} menus - 菜单列表
   */
  async addRoutes(menus) {
    const userStore = useUserStore()
    
    if (!menus || menus.length === 0) {
      console.warn('没有可添加的菜单数据')
      return []
    }

    // 如果已经添加过，先移除
    if (this.addedRoutes.length > 0) {
      this.removeRoutes()
    }

    // 导入动态路由生成函数
    const { addDynamicRoutes } = await import('@/router/dynamicRoutes')
    
    // 添加新路由
    this.addedRoutes = await addDynamicRoutes(menus, router)
    userStore.setRoutesAdded(true)
    
    console.log(`已添加 ${this.addedRoutes.length} 个动态路由`)
    return this.addedRoutes
  }

  /**
   * 移除所有动态路由
   */
  removeRoutes() {
    const { removeDynamicRoutes } = require('@/router/dynamicRoutes')
    const routeNames = this.addedRoutes.map(route => route.name).filter(Boolean)
    
    removeDynamicRoutes(router, routeNames)
    this.addedRoutes = []
    
    const userStore = useUserStore()
    userStore.setRoutesAdded(false)
    
    console.log('已移除所有动态路由')
  }

  /**
   * 重置路由（登出时调用）
   */
  reset() {
    this.removeRoutes()
  }

  /**
   * 获取已添加的路由数量
   */
  getRouteCount() {
    return this.addedRoutes.length
  }

  /**
   * 检查路由是否已添加
   */
  isAdded() {
    return this.addedRoutes.length > 0
  }
}

// 创建单例
const dynamicRouteManager = new DynamicRouteManager()

export default dynamicRouteManager