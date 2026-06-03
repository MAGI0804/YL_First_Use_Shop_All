<template>
  <div class="login-page">
    <section class="auth-panel">
      <div class="brand">
        <img src="/title.jpg" alt="logo" class="brand-logo" />
        <div>
          <h1>优蓝童装管理后台</h1>
          <p>手机号密码登录，首次登录需短信验证码激活</p>
        </div>
      </div>

      <el-tabs v-model="activeTab" class="auth-tabs">
        <el-tab-pane label="密码登录" name="login">
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
        </el-tab-pane>

        <el-tab-pane label="首次注册" name="register">
          <el-form ref="registerFormRef" :model="registerForm" :rules="registerRules" @submit.prevent="handleRegister">
            <el-form-item prop="mobile">
              <el-input v-model="registerForm.mobile" placeholder="管理员已添加的手机号" prefix-icon="Iphone" size="large" />
            </el-form-item>
            <el-form-item prop="password">
              <el-input
                v-model="registerForm.password"
                type="password"
                placeholder="设置登录密码"
                prefix-icon="Lock"
                size="large"
                show-password
              />
            </el-form-item>
            <el-form-item prop="captcha">
              <div class="captcha-row">
                <el-input v-model="registerForm.captcha" placeholder="短信验证码" prefix-icon="Key" size="large" />
                <el-button :disabled="countdown > 0" :loading="captchaLoading" size="large" @click="sendCaptcha">
                  {{ countdown > 0 ? `${countdown}s` : '获取验证码' }}
                </el-button>
              </div>
            </el-form-item>
            <el-button class="submit-button" type="primary" size="large" :loading="loading" @click="handleRegister">
              注册并登录
            </el-button>
          </el-form>
        </el-tab-pane>
      </el-tabs>
    </section>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, FormInstance } from 'element-plus'
import type { FormRules } from 'element-plus'
import { backendLogin, backendRegisterByPhone, saveBackendSession, sendBackendRegisterCaptcha } from '@/api'

const router = useRouter()
const activeTab = ref('login')
const loginFormRef = ref<FormInstance>()
const registerFormRef = ref<FormInstance>()
const loading = ref(false)
const captchaLoading = ref(false)
const countdown = ref(0)

const mobileRule = /^1[3-9]\d{9}$/

const loginForm = reactive({
  mobile: '',
  password: ''
})

const registerForm = reactive({
  mobile: '',
  password: '',
  captcha: ''
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

const registerRules: FormRules = {
  mobile: mobileRules,
  password: passwordRules,
  captcha: [
    { required: true, message: '请输入验证码', trigger: 'blur' },
    { len: 6, message: '验证码为6位数字', trigger: 'blur' }
  ]
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

const sendCaptcha = async () => {
  if (!mobileRule.test(registerForm.mobile)) {
    ElMessage.warning('请先输入正确的手机号')
    return
  }
  captchaLoading.value = true
  try {
    await sendBackendRegisterCaptcha({ mobile: registerForm.mobile })
    ElMessage.success('验证码已发送')
    countdown.value = 60
    const timer = window.setInterval(() => {
      countdown.value -= 1
      if (countdown.value <= 0) window.clearInterval(timer)
    }, 1000)
  } catch (error: any) {
    ElMessage.error(error?.response?.data?.msg || '验证码发送失败')
  } finally {
    captchaLoading.value = false
  }
}

const handleRegister = async () => {
  if (!registerFormRef.value) return
  const valid = await registerFormRef.value.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    const res = await backendRegisterByPhone(registerForm)
    saveBackendSession(res.data.user)
    ElMessage.success('注册成功')
    router.push('/dashboard')
  } catch (error: any) {
    ElMessage.error(error?.response?.data?.msg || '注册失败')
  } finally {
    loading.value = false
  }
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
  width: min(440px, 100%);
  padding: 32px;
  background: #ffffff;
  border: 1px solid #dfe5ee;
  border-radius: 8px;
  box-shadow: 0 12px 36px rgba(20, 34, 58, 0.12);
}

.brand {
  display: flex;
  gap: 14px;
  align-items: center;
  margin-bottom: 24px;
}

.brand-logo {
  width: 48px;
  height: 48px;
  object-fit: contain;
}

.brand h1 {
  margin: 0;
  font-size: 22px;
  line-height: 1.25;
  color: #172033;
}

.brand p {
  margin: 6px 0 0;
  font-size: 13px;
  color: #6b7280;
}

.auth-tabs :deep(.el-tabs__header) {
  margin-bottom: 22px;
}

.captcha-row {
  width: 100%;
  display: grid;
  grid-template-columns: 1fr 116px;
  gap: 10px;
}

.submit-button {
  width: 100%;
  margin-top: 4px;
}

@media (max-width: 520px) {
  .auth-panel {
    padding: 24px;
  }

  .captcha-row {
    grid-template-columns: 1fr;
  }
}
</style>
