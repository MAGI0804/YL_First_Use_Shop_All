<template>
  <div class="forgot-password-container">
    <div class="forgot-password-box">
      <div class="back-link" @click="goBack">
        <el-icon><ArrowLeft /></el-icon>
        <span>返回登录</span>
      </div>

      <div class="forgot-header">
        <h1 class="title">找回密码</h1>
        <p class="subtitle">请输入您的手机号码和验证码</p>
      </div>

      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        class="forgot-form"
        @submit.prevent="handleSubmit"
      >
        <el-form-item prop="phone">
          <el-input
            v-model="form.phone"
            placeholder="请输入手机号"
            prefix-icon="Iphone"
            size="large"
            maxlength="11"
          />
        </el-form-item>

        <el-form-item prop="code">
          <div class="code-input-wrapper">
            <el-input
              v-model="form.code"
              placeholder="请输入验证码"
              prefix-icon="Key"
              size="large"
              maxlength="6"
              @keyup.enter="handleSubmit"
            />
            <el-button
              size="large"
              class="code-button"
              :disabled="countdown > 0"
              @click="sendCode"
            >
              {{ countdown > 0 ? `${countdown}s` : '获取验证码' }}
            </el-button>
          </div>
        </el-form-item>

        <el-button
          type="primary"
          size="large"
          class="submit-button"
          :loading="loading"
          @click="handleSubmit"
        >
          确 定
        </el-button>
      </el-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, FormInstance } from 'element-plus'
import { ArrowLeft } from '@element-plus/icons-vue'
import type { FormRules } from 'element-plus'

const router = useRouter()
const formRef = ref<FormInstance>()
const loading = ref(false)
const countdown = ref(0)

const form = reactive({
  phone: '',
  code: ''
})

const rules: FormRules = {
  phone: [
    { required: true, message: '请输入手机号', trigger: 'blur' },
    { pattern: /^1[3-9]\d{9}$/, message: '请输入正确的手机号', trigger: 'blur' }
  ],
  code: [
    { required: true, message: '请输入验证码', trigger: 'blur' },
    { len: 6, message: '验证码为6位数字', trigger: 'blur' }
  ]
}

let timer: ReturnType<typeof setInterval> | null = null

const sendCode = () => {
  if (!formRef.value) return

  formRef.value.validateField('phone', (valid) => {
    if (valid) {
      countdown.value = 60
      timer = setInterval(() => {
        countdown.value--
        if (countdown.value <= 0 && timer) {
          clearInterval(timer)
          timer = null
        }
      }, 1000)
      ElMessage.success('验证码已发送')
    }
  })
}

const handleSubmit = async () => {
  if (!formRef.value) return

  await formRef.value.validate((valid) => {
    if (valid) {
      loading.value = true
      setTimeout(() => {
        loading.value = false
        ElMessage.success('密码重置成功，请使用新密码登录')
        router.push('/login')
      }, 1000)
    }
  })
}

const goBack = () => {
  router.push('/login')
}
</script>

<style scoped>
.forgot-password-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #f5f7fa 0%, #e4e7ed 100%);
}

.forgot-password-box {
  width: 400px;
  padding: 40px;
  background: #ffffff;
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.08);
  position: relative;
}

.back-link {
  position: absolute;
  top: 20px;
  left: 20px;
  display: flex;
  align-items: center;
  gap: 4px;
  color: #666666;
  cursor: pointer;
  font-size: 14px;
  transition: color 0.3s ease;
}

.back-link:hover {
  color: #87ceeb;
}

.forgot-header {
  text-align: center;
  margin-bottom: 32px;
  margin-top: 20px;
}

.title {
  font-size: 24px;
  font-weight: 600;
  color: #1a1a1a;
  margin: 0 0 8px 0;
}

.subtitle {
  font-size: 14px;
  color: #999999;
  margin: 0;
}

.forgot-form {
  margin-top: 24px;
}

.forgot-form :deep(.el-input__wrapper) {
  padding: 4px 12px;
}

.forgot-form :deep(.el-input__inner) {
  height: 40px;
}

.code-input-wrapper {
  display: flex;
  gap: 12px;
}

.code-input-wrapper :deep(.el-input) {
  flex: 1;
}

.code-button {
  min-width: 120px;
  background: #ffffff;
  border: 1px solid #e4e7ed;
  color: #1a1a1a;
  border-radius: 8px;
  transition: all 0.3s ease;
}

.code-button:hover:not(:disabled) {
  border-color: #87ceeb;
  color: #87ceeb;
}

.code-button:disabled {
  background: #f5f7fa;
  color: #999999;
  border-color: #e4e7ed;
}

.submit-button {
  width: 100%;
  margin-top: 16px;
  background: #87ceeb;
  border-color: #87ceeb;
  font-size: 16px;
  font-weight: 500;
  border-radius: 8px;
  transition: all 0.3s ease;
}

.submit-button:hover {
  background: #5bc0de;
  border-color: #5bc0de;
  transform: translateY(-1px);
}

.submit-button:active {
  transform: translateY(0);
}
</style>
