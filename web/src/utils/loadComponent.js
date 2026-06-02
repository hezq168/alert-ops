// 使用 Vite 的 import.meta.glob 预加载所有 views 下的组件
const modules = import.meta.glob('@/views/**/*.vue')

// 调试：查看所有可用的模块路径
console.log('可用的组件模块:', Object.keys(modules))

/**
 * 动态加载组件
 * @param {string} componentPath - 组件路径（相对于 views 目录）
 * @returns {Function} 组件加载函数
 */
export function loadComponent(componentPath) {
  if (!componentPath) {
    console.warn('组件路径为空，使用默认 Layout')
    return () => import('@/views/Layout.vue')
  }
  
  // Layout 特殊处理
  if (componentPath === 'Layout') {
    return () => import('@/views/Layout.vue')
  }
  
  // 构建完整的模块路径
  const modulePath = `/src/views/${componentPath}.vue`
  
  // 在预加载的模块中查找
  const loader = modules[modulePath]
  
  if (loader) {
    console.log(`✅ 找到组件: ${componentPath}`)
    return loader
  }
  
  // 如果找不到，尝试其他可能的路径格式
  console.warn(`⚠️ 组件未找到: ${componentPath}`)
  console.warn('期望路径:', modulePath)
  console.warn('可用路径示例:', Object.keys(modules).slice(0, 5))
  
  // 返回 404 组件
  return () => import('@/views/404.vue')
}