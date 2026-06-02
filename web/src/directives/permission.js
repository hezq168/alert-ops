import { useUserStore } from '@/stores/user'

export default {
  mounted(el, binding) {
    const userStore = useUserStore()
    const { value } = binding

    if (value && value instanceof Array && value.length > 0) {
      // 支持数组形式：v-permission="['k8s:cluster:add', 'k8s:cluster:edit']"
      const hasPermission = userStore.permissions.some(permission => {
        return value.includes(permission.code)
      })

      if (!hasPermission) {
        el.parentNode && el.parentNode.removeChild(el)
      }
    } else {
      // 支持字符串形式：v-permission="'k8s:cluster:add'"
      const hasPermission = userStore.hasPermission(value)

      if (!hasPermission) {
        el.parentNode && el.parentNode.removeChild(el)
      }
    }
  }
}