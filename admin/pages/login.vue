<template>
  <div class="login-page">
    <div class="login-bg-shapes">
      <div class="shape shape-1"></div>
      <div class="shape shape-2"></div>
      <div class="shape shape-3"></div>
      <div class="shape shape-4"></div>
      <div class="grid-overlay"></div>
    </div>

    <img v-if="bgUrl" :src="bgUrl" class="bg-image" alt=""
         :style="{ filter: `blur(${bgBlur}px) brightness(0.5)` }" />

    <div class="login-container">
      <div class="login-left">
        <div class="brand-area">
          <h1 class="brand-title">芙芙云计算</h1>
          <p class="brand-subtitle">SSH-WEB 管理控制台</p>
        </div>

        <div class="login-footer-text">
          版权所有©芙芙云计算（www.fufuidc.com）2.1.0LTS · AUTHOR by QQ:2049963372
          Developer 锦衣喵
        </div>
      </div>

      <div class="login-right">
        <div class="login-card">
          <div class="login-header">
            <h2 class="login-title">管理员登录</h2>
            <p class="login-desc">请输入您的管理员账户信息</p>
          </div>

          <el-form @submit.prevent="handleLogin" label-position="top" size="large">
            <el-form-item label="用户名">
              <el-input v-model="form.username" placeholder="请输入用户名" :prefix-icon="User" autofocus />
            </el-form-item>

            <el-form-item label="密码">
              <el-input v-model="form.password" type="password" placeholder="请输入密码"
                        :prefix-icon="Lock" show-password @keyup.enter="handleLogin" />
            </el-form-item>

            <el-button type="primary" :loading="loading" class="login-btn" @click="handleLogin">
              {{ loading ? '登录中...' : '登 录' }}
            </el-button>
          </el-form>

          <el-alert v-if="errorMsg" :title="errorMsg" type="error" :closable="false" show-icon class="login-error" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { User, Lock } from '@element-plus/icons-vue'

definePageMeta({ layout: false })

const auth = useAdminAuth()
const router = useRouter()
const loading = ref(false)
const errorMsg = ref('')

const form = ref({ username: '', password: '' })

const bgUrl = ref('')
const bgBlur = ref(0)

onMounted(async () => {
  const status = await auth.checkStatus()
  if (!status.initialized) { router.replace('/init'); return }

  try {
    const res = await $fetch('/api/system-settings')
    if (res.success && res.settings) {
      const s = res.settings
      bgUrl.value = s.background_url || ''
      bgBlur.value = s.background_blur || 0
    }
  } catch (e) {
    console.warn('加载系统设置失败', e)
  }
})

async function handleLogin() {
  if (!form.value.username || !form.value.password) { errorMsg.value = '请输入用户名和密码'; return }

  errorMsg.value = ''
  loading.value = true
  try {
    const result = await auth.login({ username: form.value.username, password: form.value.password })
    if (result.success) router.push('/')
    else errorMsg.value = result.message || '登录失败'
  } catch (e: any) { errorMsg.value = e.response?.data?.message || '网络错误，请重试' }
  finally { loading.value = false }
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

.login-bg-shapes { position: absolute; inset: 0; pointer-events: none; }

.shape {
  position: absolute;
  border-radius: 50%;
  opacity: 0.06;
}
.shape-1 {
  width: 500px; height: 500px;
  background: #2563eb;
  top: -180px; right: -120px;
}
.shape-2 {
  width: 350px; height: 350px;
  background: #3b82f6;
  bottom: -100px; left: -80px;
}
.shape-3 {
  width: 200px; height: 200px;
  background: #60a5fa;
  top: 40%; right: 10%;
}
.shape-4 {
  width: 150px; height: 150px;
  background: #93c5fd;
  top: 15%; left: 15%;
}

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

.brand-area { position: relative; z-index: 1; }

.brand-icon-wrap {
  width: 56px; height: 56px;
  margin-bottom: 20px;
  filter: drop-shadow(0 8px 16px rgba(0,0,0,0.15));
}

.brand-svg { width: 100%; height: 100%; }

.brand-title {
  font-size: 26px;
  font-weight: 800;
  color: #fff;
  letter-spacing: -0.02em;
  line-height: 1.2;
  margin-bottom: 6px;
}

.brand-subtitle {
  font-size: 14px;
  color: rgba(255,255,255,0.7);
  font-weight: 400;
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

.login-card { width: 100%; max-width: 360px; }

.login-header { margin-bottom: 28px; }

.login-title {
  font-size: 22px;
  font-weight: 700;
  color: #111827;
  letter-spacing: -0.02em;
  margin-bottom: 6px;
}

.login-desc {
  font-size: 14px;
  color: #9ca3af;
  font-weight: 400;
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

.login-error { margin-top: 16px; }

@media (max-width: 768px) {
  .login-container {
    flex-direction: column;
    max-width: 420px;
    min-height: auto;
  }
  .login-left {
    flex: none;
    padding: 32px 28px 24px;
  }
  .brand-area { text-align: center; }
  .brand-icon-wrap { margin: 0 auto 14px; }
  .login-footer-text { display: none; }
  .login-right { padding: 28px 24px 32px; }
}
</style>
