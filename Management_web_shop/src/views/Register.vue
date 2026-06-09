<template>
  <div class="register-page">
    <section class="auth-panel">
      <div class="brand">
        <img src="/title.jpg" alt="logo" class="brand-logo" />
      </div>

      <el-form ref="registerFormRef" :model="registerForm" :rules="registerRules" @submit.prevent="handleRegister">
        <el-form-item prop="mobile">
          <el-input v-model="registerForm.mobile" placeholder="手机号" prefix-icon="Iphone" size="large" />
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
          注册
        </el-button>
      </el-form>

      <div class="login-link">
        <el-button type="text" @click="goToLogin">返回登录</el-button>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, FormInstance } from 'element-plus'
import type { FormRules } from 'element-plus'
import { backendRegisterByPhone, saveBackendSession, sendBackendRegisterCaptcha } from '@/api'

const router = useRouter()
const registerFormRef = ref<FormInstance>()
const loading = ref(false)
const captchaLoading = ref(false)
const countdown = ref(0)

const mobileRule = /^1[3-9]\d{9}$/

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

const registerRules: FormRules = {
  mobile: mobileRules,
  password: passwordRules,
  captcha: [
    { required: true, message: '请输入验证码', trigger: 'blur' },
    { len: 6, message: '验证码为6位数字', trigger: 'blur' }
  ]
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

const goToLogin = () => {
  router.push('/login')
}
</script>

<style scoped>
.register-page {
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

.login-link {
  text-align: center;
  margin-top: 16px;
}

.login-link :deep(.el-button--text) {
  font-size: 14px;
  color: #409eff;
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