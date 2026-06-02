import Layout from '@/views/Layout.vue'
import { loadComponent } from '@/utils/loadComponent'

/**
 * 将后端返回的菜单数据转换为 Vue Router 路由配置
 */
function convertMenuToRoute(menu) {
  // 1. 过滤掉按钮类型的权限，它们不需要生成路由
  if (menu.type !== 'menu') {
    return null
  }

  const isLayout = menu.component === 'Layout'
  
  // 2. 处理路径：如果是 Layout 且没有 path，给一个占位符以便挂载子路由
  let routePath = menu.path
  if (isLayout && !routePath) {
    routePath = `/layout-${menu.id}`
  } else if (!routePath) {
    // 如果没有路径且不是 Layout，则无法生成路由
    return null
  }
  
  const route = {
    path: routePath.startsWith('/') ? routePath : `/${routePath}`,
    name: menu.code ? menu.code.replace(/:/g, '_') : undefined,
    component: loadComponent(menu.component),
    meta: {
      title: menu.name,
      icon: menu.icon,
      permission: menu.code, // 用于路由守卫的权限检查
      hidden: menu.hidden || false
    }
  }
  
  // 3. 递归处理子菜单
  if (menu.children && menu.children.length > 0) {
    const children = menu.children
      .map(child => convertMenuToRoute(child))
      .filter(r => r !== null)
    
    if (children.length > 0) {
      if (isLayout) {
        // 如果是 Layout，直接返回子路由数组（扁平化）
        return children
      } else {
        route.children = children
      }
    }
  }
  
  // 4. 如果是一个没有子项的 Layout，则不生成路由
  if (isLayout && (!menu.children || menu.children.length === 0)) {
    return null
  }
  
  return route
}

/**
 * 格式化菜单树（确保父子关系正确）
 */
export function formatMenus(menus) {
  const map = {}
  const roots = []
  
  menus.forEach(m => {
    map[m.id] = { ...m, children: [] }
  })
  
  menus.forEach(m => {
    if (m.parent_id === 0) {
      roots.push(map[m.id])
    } else if (map[m.parent_id]) {
      map[m.parent_id].children.push(map[m.id])
    }
  })
  
  // 按 sort 排序
  return roots.sort((a, b) => a.sort - b.sort)
}

/**
 * 生成路由配置
 */
export function generateRoutes(menus) {
  return menus
    .map(menu => convertMenuToRoute(menu))
    .filter(route => route !== null)
}

/**
 * 动态添加路由到路由器实例
 */
export async function addDynamicRoutes(menus, router) {
  const formattedMenus = formatMenus(menus)
  const dynamicRoutes = generateRoutes(formattedMenus)
  
  const addedRoutes = []
  
  /**
   * 递归添加路由
   * @param {Object|Array} route - 路由对象或数组
   * @param {String} parentRouteName - 父路由名称，默认为 'Layout'
   */
  function addRouteRecursively(route, parentRouteName = 'Layout') {
    if (Array.isArray(route)) {
      // 如果是数组（通常是 Layout 的子路由），遍历添加
      route.forEach(r => addRouteRecursively(r, parentRouteName))
    } else if (route) {
      // 添加到指定的父路由下
      router.addRoute(parentRouteName, route)
      addedRoutes.push(route)
      console.log(`✅ 已添加路由: ${route.path} (name: ${route.name})`)
      
      // 如果该路由还有子路由，继续递归添加
      if (route.children && route.children.length > 0) {
        route.children.forEach(child => {
          addRouteRecursively(child, route.name)
        })
      }
    }
  }
  
  dynamicRoutes.forEach(route => {
    addRouteRecursively(route)
  })
  
  console.log('✅ 动态路由添加完成，共添加:', addedRoutes.length, '个路由')
  return addedRoutes
}

/**
 * 移除所有动态路由
 * @param {Object} router - Vue Router 实例
 * @param {Array} routeNames - 要移除的路由名称列表
 */
export function removeDynamicRoutes(router, routeNames = []) {
  routeNames.forEach(name => {
    if (name) {
      router.removeRoute(name)
    }
  })
  console.log('动态路由已移除')
}

export default {
  generateRoutes,
  formatMenus,
  addDynamicRoutes,
  removeDynamicRoutes
}