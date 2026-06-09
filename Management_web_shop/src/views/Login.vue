<template>
  <div class="login-page">
    <section class="auth-panel">
      <div class="brand">
        <img src="/title.jpg" alt="logo" class="brand-logo" />
      </div>

      <el-form ref="loginFormRef" :model="loginForm" :rules="loginRules" @submit.prevent="handleLogin">
        <el-form-item prop="mobile">
          <el-input v-model="loginForm.mobile" placeholder="手机号" prefix-icon="Iphone" size="large" />
        </el-form-item>
        <el-form-item prop="password">
          <el-input
            v-model="loginForm.password"
            type="password"
            placeholder="密码"
            prefix-icon="Lock"
            size="large"
            show-password
            @keyup.enter="handleLogin"
          />
        </el-form-item>
        <el-button class="submit-button" type="primary" size="large" :loading="loading" @click="handleLogin">
          登录
        </el-button>
      </el-form>

      <div class="register-link">
        <el-button type="text" @click="goToRegister">注册</el-button>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, FormInstance } from 'element-plus'
import type { FormRules } from 'element-plus'
import { backendLogin, saveBackendSession } from '@/api'

const router = useRouter()
const loginFormRef = ref<FormInstance>()
const loading = ref(false)

const mobileRule = /^1[3-9]\d{9}$/

const loginForm = reactive({
  mobile: '',
  password: ''
})

const mobileRules = [
  { required: true, message: '请输入手机号', trigger: 'blur' },
  { pattern: mobileRule, message: '手机号格式不正确', trigger: 'blur' }
]

const passwordRules = [
  { required: true, message: '请输入密码', trigger: 'blur' },
  { min: 6, message: '密码至少6位', trigger: 'blur' }
]

const loginRules: FormRules = {
  mobile: mobileRules,
  password: passwordRules
}

const handleLogin = async () => {
  if (!loginFormRef.value) return
  const valid = await loginFormRef.value.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    const res = await backendLogin(loginForm)
    saveBackendSession(res.data.user)
    ElMessage.success('登录成功')
    router.push('/dashboard')
  } catch (error: any) {
    ElMessage.error(error?.response?.data?.msg || '登录失败')
  } finally {
    loading.value = false
  }
}

const goToRegister = () => {
  router.push('/register')
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: grid;
  place-items: center;
  padding: 24px;
  background: #eef2f7;
}

.auth-panel {
  width: min(380px, 100%);
  padding: 28px;
  background: #ffffff;
  border: 1px solid #dfe5ee;
  border-radius: 8px;
}

.brand {
  display: flex;
  justify-content: center;
  margin-bottom: 18px;
}

.brand-logo {
  width: 64px;
  height: 64px;
  object-fit: contain;
}

.submit-button {
  width: 100%;
  margin-top: 4px;
}

.register-link {
  text-align: center;
  margin-top: 16px;
}

.register-link :deep(.el-button--text) {
  font-size: 14px;
  color: #409eff;
}

@media (max-width: 520px) {
  .auth-panel {
    padding: 24px;
  }
}
</style>