<template>
  <div class="forbidden-container">
    <el-result
      icon="error"
      title="403 无权限访问"
      sub-title="抱歉，您没有权限访问该页面，请联系管理员分配相应权限"
    >
      <template #extra>
        <el-button type="primary" @click="goHome">返回首页</el-button>
        <el-button @click="goBack">返回上一页</el-button>
      </template>
    </el-result>
  </div>
</template>

<script setup>
import { useRouter, useRoute } from 'vue-router'

const router = useRouter()
const route = useRoute()

const goHome = () => {
  router.push('/')
}

const goBack = () => {
  const redirect = route.query.redirect
  if (redirect && redirect !== '/403') {
    router.back()
  } else {
    router.push('/')
  }
}
</script>

<style scoped>
.forbidden-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: calc(100vh - 200px);
  padding: 40px;
}
</style>