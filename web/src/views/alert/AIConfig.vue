<template>
  <div class="ai-config-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span class="title">AI 分析配置</span>
        </div>
      </template>

      <el-form :model="form" label-width="120px" style="max-width: 600px">
        <el-form-item label="AI 提供商" required>
          <el-select v-model="form.provider" style="width: 100%">
            <el-option label="OpenAI 兼容（阿里百炼/DeepSeek/等）" value="openai" />
            <el-option label="CodeBuddy" value="codebuddy" />
          </el-select>
        </el-form-item>

        <el-form-item label="API Key" required>
          <el-input v-model="form.api_key" type="password" show-password placeholder="请输入 API Key" />
        </el-form-item>

        <el-form-item label="API 地址" required>
          <el-input v-model="form.base_url" placeholder="如 https://dashscope.aliyuncs.com/compatible-mode/v1" />
          <div class="form-tip">
            常用地址：阿里百炼 dashscope.aliyuncs.com/compatible-mode/v1 |
            DeepSeek api.deepseek.com/v1 |
            OpenAI api.openai.com/v1
          </div>
        </el-form-item>

        <el-form-item label="模型名称" required>
          <el-input v-model="form.model" placeholder="如 qwen-turbo / deepseek-chat / gpt-4o" />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="handleSave" :loading="saving">保存并生效</el-button>
          <el-button @click="handleTest" :loading="testing" type="success" plain>测试连接</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getAIConfig, updateAIConfig, testAIConnection } from '@/api/alert'

const form = reactive({
  provider: 'openai',
  api_key: '',
  base_url: '',
  model: ''
})

const saving = ref(false)
const testing = ref(false)

onMounted(async () => {
  try {
    const res = await getAIConfig()
    if (res.data) {
      Object.assign(form, res.data)
    }
  } catch (e) {
    // 未配置时不报错
  }
})

async function handleSave() {
  if (!form.provider || !form.api_key || !form.base_url || !form.model) {
    ElMessage.warning('请填写完整的配置信息')
    return
  }
  saving.value = true
  try {
    await updateAIConfig({ ...form })
    ElMessage.success('AI 配置已保存并生效')
  } catch (e) {
    ElMessage.error('保存失败：' + (e?.message || '未知错误'))
  } finally {
    saving.value = false
  }
}

async function handleTest() {
  if (!form.provider || !form.api_key || !form.base_url || !form.model) {
    ElMessage.warning('请填写完整的配置信息后再测试')
    return
  }
  testing.value = true
  try {
    const res = await testAIConnection({ ...form })
    ElMessage.success(`连接成功！模型：${res.data.model}，响应：${res.data.result}`)
  } catch (e) {
    ElMessage.error('连接测试失败：' + (e?.message || '未知错误'))
  } finally {
    testing.value = false
  }
}
</script>

<style scoped>
.ai-config-page {
  padding: 20px;
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.title {
  font-size: 18px;
  font-weight: 600;
}
.form-tip {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
  line-height: 1.6;
}
</style>
