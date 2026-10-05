<template>
  <div class="login-page">

    <div class="login-bg-shapes">
      <div class="shape shape-1"></div>
      <div class="shape shape-2"></div>
      <div class="shape shape-3"></div>
      <div class="shape shape-4"></div>
      <div class="grid-overlay"></div>
    </div>

    <img v-if="settings.background_url" :src="settings.background_url" class="bg-image" alt=""
         :style="{ filter: `blur(${settings.background_blur || 0}px) brightness(0.5)` }" />

    <div class="login-container">

      <div class="login-left">
        <div class="brand-area">
          <img
            :src="settings.logo_url || 'https://logo.fufuidc.com/logo3.png'"
            class="brand-logo"
            alt="logo"
          />
          <h1 class="brand-title">{{ settings.site_title || 'SSH-WEB' }}</h1>
          <p class="brand-subtitle">云端SSH终端管理平台</p>
        </div>

        <div class="login-footer-text" v-if="settings.footer_text">
          {{ settings.footer_text }}
        </div>
      </div>

      <div class="login-right">
        <div class="login-card">
          <el-tabs v-model="activeTab" stretch class="login-tabs">
            <el-tab-pane label="登录" name="login">
              <div class="tab-header">
                <h2 class="tab-title">欢迎回来</h2>
                <p class="tab-desc">请输入您的账户信息</p>
              </div>
              <el-form ref="loginRef" :model="loginForm" :rules="loginRules" label-position="top" size="large" @submit.prevent="handleLogin">
                <el-form-item prop="user_id">
                  <el-input v-model="loginForm.user_id" placeholder="用户ID" :prefix-icon="User" @keyup.enter="handleLogin" />
                </el-form-item>
                <el-form-item prop="password">
                  <el-input v-model="loginForm.password" type="password" placeholder="密码" :prefix-icon="Lock" show-password @keyup.enter="handleLogin" />
                </el-form-item>

                <el-form-item v-if="captchaEnabled" label="安全验证" class="captcha-form-item">
                  <div id="captcha-container"></div>
                  <div v-if="captchaError" class="captcha-error">{{ captchaError }}</div>
                </el-form-item>
                <el-button type="primary" :loading="loading" class="login-btn" @click="handleLogin">
                  {{ loading ? '登录中...' : '登 录' }}
                </el-button>
              </el-form>
            </el-tab-pane>

            <el-tab-pane v-if="settings.allow_register !== false" label="注册" name="register">
              <div class="tab-header">
                <h2 class="tab-title">创建账户</h2>
                <p class="tab-desc">注册一个新的SSH-WEB账户</p>
              </div>
              <el-form ref="regRef" :model="regForm" :rules="regRules" label-position="top" size="large" @submit.prevent="handleRegister">
                <el-form-item prop="user_id">
                  <el-input v-model="regForm.user_id" placeholder="用户ID（2-32个字符）" :prefix-icon="User" />
                </el-form-item>
                <el-form-item prop="password">
                  <el-input v-model="regForm.password" type="password" placeholder="密码（4-64个字符）" :prefix-icon="Lock" show-password @keyup.enter="handleRegister" />
                </el-form-item>
                <el-form-item prop="confirmPassword">
                  <el-input v-model="regForm.confirmPassword" type="password" placeholder="确认密码" :prefix-icon="Lock" show-password @keyup.enter="handleRegister" />
                </el-form-item>
                <el-button type="primary" :loading="loading" class="login-btn" @click="handleRegister">
                  {{ loading ? '注册中...' : '注 册' }}
                </el-button>
              </el-form>
            </el-tab-pane>
          </el-tabs>
        </div>
      </div>
    </div>

    <div class="page-footer">
      <span>若叶睦&锦衣</span>
      <span v-if="settings.footer_text"> | {{ settings.footer_text }}</span>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, watch, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { User, Lock } from '@element-plus/icons-vue'
import { authAPI } from '../api'
import { useSystemSettings } from '../composables/useSystemSettings'

const router = useRouter()
const { settings } = useSystemSettings()
const activeTab = ref('login')
const loading = ref(false)
const loginRef = ref(null)
const regRef = ref(null)

let captchaInstance = null
const captchaError = ref('')
const captchaEnabled = ref(false)
const captchaResult = ref(null)

const loginForm = reactive({ user_id: '', password: '' })
const regForm = reactive({ user_id: '', password: '', confirmPassword: '' })

const validateConfirmPass = (rule, value, callback) => {
  if (value !== regForm.password) callback(new Error('两次密码不一致'))
  else callback()
}

const loginRules = {
  user_id: [{ required: true, message: '请输入用户ID', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

const regRules = {
  user_id: [
    { required: true, message: '请输入用户ID', trigger: 'blur' },
    { min: 2, max: 32, message: '长度2-32个字符', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 4, max: 64, message: '长度4-64个字符', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    { validator: validateConfirmPass, trigger: 'blur' }
  ]
}

function loadGeetestSDK() {
  return new Promise((resolve, reject) => {
    if (window.initGeetest4) {
      resolve()
      return
    }
    const script = document.createElement('script')
    script.src = 'https://static.geetest.com/v4/gt4.js'
    script.async = true
    script.onload = () => resolve()
    script.onerror = () => reject(new Error('极验SDK加载失败'))
    document.head.appendChild(script)
  })
}

async function initCaptcha() {
  if (!settings.value.captcha_id) return
  try {
    await loadGeetestSDK()
    if (captchaInstance) {
      try { captchaInstance.destroy() } catch (e) {  }
    }
    window.initGeetest4({
      captchaId: settings.value.captcha_id,
      product: 'float',
      nativeButton: { width: '100%', height: '40px' }
    }, function (captcha) {
      captchaInstance = captcha
      captchaInstance.appendTo('#captcha-container')
      captchaInstance.onSuccess(function () {
        const result = captchaInstance.getValidate()
        if (result) {
          captchaResult.value = {
            lot_number: result.lot_number,
            captcha_output: result.captcha_output,
            pass_token: result.pass_token,
            gen_time: result.gen_time
          }
          captchaError.value = ''
        }
      })
      captchaInstance.onError(function () {
        captchaError.value = '验证码加载异常，请刷新重试'
      })
    })
  } catch (e) {
    captchaError.value = '验证码SDK加载失败'
  }
}

watch(() => settings.value, (val) => {
  captchaEnabled.value = !!(val.captcha_enabled && val.captcha_id)
}, { immediate: true, deep: true })

watch([captchaEnabled, activeTab], async ([enabled]) => {
  if (enabled && activeTab.value === 'login') {
    await nextTick()
    initCaptcha()
  }
})

onMounted(() => {
  if (captchaEnabled.value) {
    nextTick(() => initCaptcha())
  }
})

const handleLogin = async () => {
  if (!loginRef.value) return
  await loginRef.value.validate(async (valid) => {
    if (!valid) return

    if (captchaEnabled.value && !captchaResult.value) {
      ElMessage.warning('请完成安全验证')
      return
    }

    loading.value = true
    try {
      const payload = {
        user_id: loginForm.user_id,
        password: loginForm.password
      }
      if (captchaResult.value) {
        payload.captcha = captchaResult.value
      }
      const res = await authAPI.login(payload)
      if (res.data.success) { ElMessage.success('登录成功'); router.push('/') }
      else { ElMessage.error(res.data.message || '登录失败') }
    } catch { ElMessage.error('登录失败，请检查网络') }
    finally { loading.value = false }
  })
}

const handleRegister = async () => {
  if (!regRef.value) return
  await regRef.value.validate(async (valid) => {
    if (!valid) return
    loading.value = true
    try {
      const res = await authAPI.register({ user_id: regForm.user_id, password: regForm.password })
      if (res.data.success) { ElMessage.success('注册成功'); router.push('/') }
      else { ElMessage.error(res.data.message || '注册失败') }
    } catch { ElMessage.error('注册失败，请检查网络') }
    finally { loading.value = false }
  })
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #f0f5ff 0%, #e8f0fe 30%, #dbeafe 60%, #eff6ff 100%);
  position: relative;
  overflow: hidden;
  padding: 20px;
}

.login-bg-shapes {
  position: absolute;
  inset: 0;
  pointer-events: none;
  z-index: 0;
}
.shape {
  position: absolute;
  border-radius: 50%;
  opacity: 0.06;
}
.shape-1 { width: 500px; height: 500px; background: #2563eb; top: -180px; right: -120px; }
.shape-2 { width: 350px; height: 350px; background: #3b82f6; bottom: -100px; left: -80px; }
.shape-3 { width: 200px; height: 200px; background: #60a5fa; top: 40%; right: 10%; }
.shape-4 { width: 150px; height: 150px; background: #93c5fd; top: 15%; left: 15%; }

.grid-overlay {
  position: absolute; inset: 0;
  background-image:
    linear-gradient(rgba(37,99,235,0.03) 1px, transparent 1px),
    linear-gradient(90deg, rgba(37,99,235,0.03) 1px, transparent 1px);
  background-size: 40px 40px;
}

.bg-image {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  object-position: center;
  z-index: 0;
}

.login-container {
  position: relative;
  z-index: 1;
  display: flex;
  width: 100%;
  max-width: 880px;
  min-height: 520px;
  background: #fff;
  border-radius: 20px;
  box-shadow:
    0 4px 6px -1px rgb(0 0 0 / 0.04),
    0 10px 30px -4px rgb(37 99 235 / 0.08),
    0 0 0 1px rgba(37, 99, 235, 0.06);
  overflow: hidden;
  animation: fadeIn 0.6s ease;
}

.login-left {
  flex: 0 0 340px;
  background: linear-gradient(160deg, #1e40af 0%, #2563eb 50%, #3b82f6 100%);
  padding: 44px 36px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  position: relative;
  overflow: hidden;
}
.login-left::before {
  content: '';
  position: absolute;
  top: -60px; right: -60px;
  width: 200px; height: 200px;
  border-radius: 50%;
  background: rgba(255,255,255,0.07);
}
.login-left::after {
  content: '';
  position: absolute;
  bottom: -40px; left: -30px;
  width: 140px; height: 140px;
  border-radius: 50%;
  background: rgba(255,255,255,0.05);
}

.brand-area {
  position: relative;
  z-index: 1;
}
.brand-logo {
  width: 64px;
  height: 64px;
  object-fit: contain;
  border-radius: 16px;
  margin-bottom: 18px;
  filter: drop-shadow(0 8px 16px rgba(0,0,0,0.15));
}
.brand-title {
  font-size: 26px;
  font-weight: 800;
  color: #fff;
  letter-spacing: -0.02em;
  line-height: 1.2;
  margin-bottom: 6px;
  margin-top: 0;
}
.brand-subtitle {
  font-size: 14px;
  color: rgba(255,255,255,0.7);
  font-weight: 400;
  margin: 0;
}

.login-footer-text {
  position: relative;
  z-index: 1;
  color: rgba(255,255,255,0.35);
  font-size: 11.5px;
  letter-spacing: 2px;
  text-transform: uppercase;
  font-weight: 600;
}

.login-right {
  flex: 1;
  padding: 44px 40px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.login-card {
  width: 100%;
  max-width: 380px;
}

.login-tabs {
  width: 100%;
}
.tab-header {
  margin-bottom: 24px;
}
.tab-title {
  font-size: 22px;
  font-weight: 700;
  color: #111827;
  letter-spacing: -0.02em;
  margin: 0 0 6px 0;
}
.tab-desc {
  font-size: 14px;
  color: #9ca3af;
  font-weight: 400;
  margin: 0;
}

.login-tabs :deep(.el-tabs__item) {
  font-size: 15px;
  font-weight: 600;
  color: #94a3b8;
  padding: 0 4px;
  transition: color 0.2s;
}
.login-tabs :deep(.el-tabs__item.is-active) {
  color: #2563eb;
}
.login-tabs :deep(.el-tabs__active-bar) {
  background-color: #2563eb;
  height: 2.5px;
}
.login-tabs :deep(.el-tabs__nav-wrap::after) {
  background-color: #e5e7eb;
  height: 1px;
}

.captcha-form-item {
  margin-bottom: 16px;
}
.captcha-form-item :deep(.el-form-item__label) {
  font-size: 13px;
  color: #6b7280;
  padding-bottom: 4px;
}
.captcha-error {
  color: #f56c6c;
  font-size: 12px;
  margin-top: 4px;
}

.login-btn {
  width: 100% !important;
  height: 42px !important;
  border-radius: 10px !important;
  font-size: 15px !important;
  font-weight: 600 !important;
  letter-spacing: 0.05em;
  margin-top: 4px;
  background: linear-gradient(135deg, #2563eb 0%, #3b82f6 100%) !important;
  border: none !important;
  box-shadow: 0 4px 12px rgba(37,99,235,0.25) !important;
  transition: all 0.2s ease !important;
}
.login-btn:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 6px 18px rgba(37,99,235,0.35) !important;
}
.login-btn:active:not(:disabled) {
  transform: translateY(0);
}

.page-footer {
  position: absolute;
  bottom: 20px;
  left: 50%;
  transform: translateX(-50%);
  color: rgba(0,0,0,0.35);
  font-size: 12px;
  text-align: center;
  z-index: 1;
  letter-spacing: 1px;
  white-space: nowrap;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(12px); }
  to { opacity: 1; transform: translateY(0); }
}

@media (max-width: 768px) {
  .login-container {
    flex-direction: column;
    max-width: 420px;
    min-height: auto;
  }
  .login-left {
    flex: none;
    padding: 28px 24px 20px;
  }
  .brand-logo {
    width: 52px;
    height: 52px;
    margin-bottom: 14px;
  }
  .brand-title { font-size: 22px; }
  .brand-area { text-align: center; }
  .brand-logo { margin-left: auto; margin-right: auto; }
  .login-footer-text { display: none; }
  .login-right { padding: 28px 24px 32px; }
  .tab-title { font-size: 20px; }
  .page-footer { font-size: 11px; }
}

@media (max-width: 480px) {
  .login-page { padding: 12px; }
  .login-container { border-radius: 14px; }
  .login-left { padding: 24px 20px 16px; }
  .login-right { padding: 24px 18px 28px; }
  .login-btn { height: 40px !important; font-size: 14px !important; }
}
</style>
