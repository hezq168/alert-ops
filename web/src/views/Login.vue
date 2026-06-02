<template>
  <div class="login-container">
    <div class="login-box">
      <h1 class="login-title">Alert-Ops</h1>
      <el-form :model="loginForm" label-width="80px">
        <el-form-item label="用户名">
          <el-input v-model="loginForm.username" placeholder="请输入用户名" :prefix-icon="User" class="form-input" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="loginForm.password" type="password" placeholder="请输入密码" :prefix-icon="Lock" class="form-input" @keyup.enter="handleLogin" />
        </el-form-item>
        <el-form-item label="验证码">
          <div class="captcha-row">
            <el-input v-model="loginForm.captchaCode" placeholder="请输入验证码" :prefix-icon="Key" maxlength="4" class="captcha-input" @keyup.enter="handleLogin" />
            <img :src="captchaImg" class="captcha-img" @click="refreshCaptcha" title="点击刷新验证码" />
          </div>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleLogin" :loading="loading" class="form-btn">登录</el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUserStore } from '../stores/user'
import { login, getCaptcha } from '../api/auth'
import { User, Lock, Key } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const loading = ref(false)

const loginForm = reactive({
  username: '',
  password: '',
  captchaId: '',
  captchaCode: ''
})

const captchaImg = ref('')

const refreshCaptcha = async () => {
  try {
    const res = await getCaptcha()
    loginForm.captchaId = res.data.captchaId
    captchaImg.value = res.data.captchaImg
  } catch {
    ElMessage.error('获取验证码失败')
  }
}

const handleLogin = async () => {
  if (!loginForm.username || !loginForm.password) {
    ElMessage.warning('请输入用户名和密码')
    return
  }
  if (!loginForm.captchaCode) {
    ElMessage.warning('请输入验证码')
    return
  }

  loading.value = true
  try {
    const res = await login(loginForm)
    userStore.setToken(res.data.token)
    userStore.setUserInfo(res.data.user)
    userStore.setPermissions(res.data.permissions || [])
    userStore.setMenus(res.data.menus || [])
    userStore.setRoutesAdded(false)

    ElMessage.success('登录成功')

    setTimeout(() => {
      const redirect = route.query.redirect
      if (redirect) {
        router.push(redirect)
      } else {
        router.push('/')
      }
    }, 300)
  } catch (error) {
    console.error('登录失败:', error)
    refreshCaptcha()
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  if (userStore.isLoggedIn) {
    router.push('/')
  }
  refreshCaptcha()
})
</script>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.login-box {
  background: white;
  padding: 40px 40px 30px;
  border-radius: 10px;
  box-shadow: 0 10px 40px rgba(0,0,0,0.1);
  width: 460px;
}

.login-title {
  text-align: center;
  font-size: 28px;
  color: #333;
  margin-bottom: 30px;
  font-weight: bold;
}

/* 统一所有输入框和按钮尺寸 */
.login-box :deep(.form-input) {
  width: 280px;
}

.login-box :deep(.el-input__wrapper) {
  height: 40px;
}

.login-box :deep(.form-btn) {
  height: 40px;
  width: 280px;
}

/* 验证码行同宽 */
.captcha-row {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 280px;
}

.captcha-input :deep(.el-input) {
  flex: 1;
}

.captcha-img {
  height: 40px;
  width: 120px;
  border-radius: 4px;
  cursor: pointer;
  border: 1px solid #dcdfe6;
  flex-shrink: 0;
}

.captcha-img:hover {
  border-color: #409eff;
}
</style>