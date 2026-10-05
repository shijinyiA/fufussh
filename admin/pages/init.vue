<template>
  <div class="login-page">
    <div class="login-bg-shapes">
      <div class="shape shape-1"></div>
      <div class="shape shape-2"></div>
      <div class="shape shape-3"></div>
      <div class="shape shape-4"></div>
      <div class="grid-overlay"></div>
    </div>

    <div class="login-container">
      <div class="login-left">
        <div class="brand-area">
          <div class="brand-icon-wrap" style="filter: drop-shadow(0 8px 16px rgba(16,185,129,0.3));">
            <svg viewBox="0 0 48 48" fill="none" class="brand-svg">
              <rect width="48" height="48" rx="14" fill="#10b981"/>
              <path d="M24 14v20M14 24h20" stroke="#fff" stroke-width="3.5" stroke-linecap="round"/>
            </svg>
          </div>
          <h1 class="brand-title">SSH-WEB</h1>
          <p class="brand-subtitle">初始化设置</p>
        </div>

        <div class="feature-list">
          <div class="feature-item">
            <div class="feature-dot" style="background:rgba(255,255,255,0.6)"></div>
            <span>创建管理员账户</span>
          </div>
          <div class="feature-item">
            <div class="feature-dot" style="background:rgba(255,255,255,0.6)"></div>
            <span>设置安全密码</span>
          </div>
          <div class="feature-item">
            <div class="feature-dot" style="background:rgba(255,255,255,0.6)"></div>
            <span>开始管理后台</span>
          </div>
        </div>

        <div class="login-footer-text">First Time Setup</div>
      </div>

      <div class="login-right">
        <div class="login-card">
          <div class="login-header">
            <h2 class="login-title">初始化管理员</h2>
            <p class="login-desc">首次使用需要创建管理员账户</p>
          </div>

          <el-form @submit.prevent="handleInit" label-position="top" size="large">
            <el-form-item label="用户名（2-64位）">
              <el-input v-model="form.username" placeholder="请输入用户名" autofocus />
            </el-form-item>

            <el-form-item label="密码（至少6位）">
              <el-input v-model="form.password" type="password" placeholder="请输入密码" show-password />
            </el-form-item>

            <el-form-item label="确认密码">
              <el-input v-model="form.confirmPassword" type="password" placeholder="请再次输入密码"
                        show-password @keyup.enter="handleInit" />
            </el-form-item>

            <div v-if="passwordStrength" class="strength-area">
              <div class="strength-bar">
                <div v-for="i in 4" :key="i" class="strength-seg"
                     :class="[i <= passwordStrength.level ? 'filled' : '', passwordStrength.level <= 1 ? 'weak' : passwordStrength.level <= 2 ? 'medium' : passwordStrength.level <= 3 ? 'good' : 'strong']"></div>
              </div>
              <p class="strength-text" :class="'level-' + (passwordStrength.level || 0)">{{ passwordStrength.text }}</p>
            </div>

            <el-button type="success" :loading="loading" :disabled="!isFormValid"
                       class="login-btn login-btn-green" @click="handleInit">
              {{ loading ? '处理中...' : '初始化并进入' }}
            </el-button>
          </el-form>

          <el-alert v-if="errorMsg" :title="errorMsg" type="error" :closable="false" show-icon class="login-error" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Key } from '@element-plus/icons-vue'

definePageMeta({ layout: false })

const auth = useAdminAuth()
const router = useRouter()
const loading = ref(false)
const errorMsg = ref('')

const form = ref({ username: '', password: '', confirmPassword: '' })

const isFormValid = computed(() =>
  form.value.username.length >= 2 &&
  form.value.password.length >= 6 &&
  form.value.password === form.value.confirmPassword
)

const passwordStrength = computed(() => {
  const pwd = form.value.password
  if (!pwd) return null
  let score = 0
  if (pwd.length >= 6) score++
  if (pwd.length >= 10) score++
  if (/[A-Z]/.test(pwd) && /[a-z]/.test(pwd)) score++
  if (/\d/.test(pwd)) score++
  if (/[^A-Za-z0-9]/.test(pwd)) score++

  if (score <= 2) return { level: 1, text: '弱：建议混合大小写、数字和符号' }
  if (score <= 3) return { level: 2, text: '中等强度' }
  if (score <= 4) return { level: 3, text: '强度良好' }
  return { level: 4, text: '非常强' }
})

onMounted(async () => {
  const status = await auth.checkStatus()
  if (status.initialized) router.replace('/login')
})

async function handleInit() {
  if (form.value.password !== form.value.confirmPassword) { errorMsg.value = '两次密码不一致'; return }
  errorMsg.value = ''
  loading.value = true
  try {
    const result = await auth.initAdmin(form.value.username, form.value.password)
    if (result.success) {
      const lr = await auth.login(form.value.username, form.value.password)
      if (lr.success) router.push('/')
    } else errorMsg.value = result.message || '初始化失败'
  } catch (e: any) { errorMsg.value = e.response?.data?.message || '网络错误' }
  finally { loading.value = false }
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex; align-items: center; justify-content: center;
  background: linear-gradient(135deg, #f0fdf4 0%, #ecfdf5 30%, #d1fae5 60%, #f0fdf4 100%);
  position: relative; overflow: hidden; padding: 20px;
}

.login-bg-shapes { position: absolute; inset: 0; pointer-events: none; }

.shape { position: absolute; border-radius: 50%; opacity: 0.06; }
.shape-1 { width: 500px; height: 500px; background: #059669; top: -180px; right: -120px; }
.shape-2 { width: 350px; height: 350px; background: #10b981; bottom: -100px; left: -80px; }
.shape-3 { width: 200px; height: 200px; background: #34d399; top: 40%; right: 10%; }
.shape-4 { width: 150px; height: 150px; background: #6ee7b7; top: 15%; left: 15%; }

.grid-overlay {
  position: absolute; inset: 0;
  background-image:
    linear-gradient(rgba(5,150,105,0.03) 1px, transparent 1px),
    linear-gradient(90deg, rgba(5,150,105,0.03) 1px, transparent 1px);
  background-size: 40px 40px;
}

.login-container {
  position: relative; z-index: 1;
  display: flex; width: 100%;
  max-width: 880px; min-height: 520px;
  background: #fff; border-radius: 20px;
  box-shadow: 0 4px 6px -1px rgb(0 0 0 / 0.04), 0 10px 30px -4px rgb(5 150 105 / 0.08), 0 0 0 1px rgba(5,150,105,0.06);
  overflow: hidden;
}

.login-left {
  flex: 0 0 340px;
  background: linear-gradient(160deg, #065f46 0%, #059669 50%, #10b981 100%);
  padding: 44px 36px;
  display: flex; flex-direction: column; justify-content: space-between;
  position: relative; overflow: hidden;
}

.login-left::before {
  content: ''; position: absolute; top: -60px; right: -60px;
  width: 200px; height: 200px; border-radius: 50%; background: rgba(255,255,255,0.07);
}
.login-left::after {
  content: ''; position: absolute; bottom: -40px; left: -30px;
  width: 140px; height: 140px; border-radius: 50%; background: rgba(255,255,255,0.05);
}

.brand-area { position: relative; z-index: 1; }

.brand-icon-wrap { width: 56px; height: 56px; margin-bottom: 20px; }
.brand-svg { width: 100%; height: 100%; }

.brand-title { font-size: 26px; font-weight: 800; color: #fff; letter-spacing: -0.02em; line-height: 1.2; margin-bottom: 6px; }
.brand-subtitle { font-size: 14px; color: rgba(255,255,255,0.7); font-weight: 400; }

.feature-list { position: relative; z-index: 1; margin-top: 32px; }
.feature-item { display: flex; align-items: center; gap: 10px; padding: 8px 0; color: rgba(255,255,255,0.85); font-size: 13.5px; }
.feature-dot { width: 7px; height: 7px; border-radius: 50%; background: rgba(255,255,255,0.5); flex-shrink: 0; }

.login-footer-text { position: relative; z-index: 1; color: rgba(255,255,255,0.35); font-size: 11.5px; letter-spacing: 2px; text-transform: uppercase; font-weight: 600; }

.login-right { flex: 1; padding: 44px 40px; display: flex; align-items: center; justify-content: center; }
.login-card { width: 100%; max-width: 360px; }
.login-header { margin-bottom: 28px; }
.login-title { font-size: 22px; font-weight: 700; color: #111827; letter-spacing: -0.02em; margin-bottom: 6px; }
.login-desc { font-size: 14px; color: #9ca3af; }

.strength-area { margin-bottom: 16px; }
.strength-bar { display: flex; gap: 4px; margin-bottom: 6px; }
.strength-seg { flex: 1; height: 4px; border-radius: 2px; background: #e5e7eb; transition: all 0.25s ease; }
.strength-seg.filled.weak { background: #ef4444; }
.strength-seg.filled.medium { background: #f59e0b; }
.strength-seg.filled.good { background: #3b82f6; }
.strength-seg.filled.strong { background: #10b981; }
.strength-text { font-size: 12px; }
.strength-text.level-0,.strength-text.level-1 { color: #ef4444; }
.strength-text.level-2 { color: #f59e0b; }
.strength-text.level-3 { color: #3b82f6; }
.strength-text.level-4 { color: #10b981; }

.login-btn {
  width: 100% !important; height: 42px !important; border-radius: 10px !important;
  font-size: 15px !important; font-weight: 600 !important; letter-spacing: 0.05em; margin-top: 4px;
  box-shadow: 0 4px 12px rgba(37,99,235,0.25) !important; transition: all 0.2s ease !important;
}
.login-btn:hover:not(:disabled) { transform: translateY(-1px); box-shadow: 0 6px 18px rgba(37,99,235,0.35) !important; }
.login-btn-green {
  background: linear-gradient(135deg, #059669 0%, #10b981 100%) !important; border: none !important;
  box-shadow: 0 4px 12px rgba(5,150,105,0.25) !important;
}
.login-btn-green:hover:not(:disabled) { box-shadow: 0 6px 18px rgba(5,150,105,0.35) !important; }

.login-error { margin-top: 16px; }

@media (max-width: 768px) {
  .login-container { flex-direction: column; max-width: 420px; min-height: auto; }
  .login-left { flex: none; padding: 32px 28px 24px; }
  .feature-list { display: none; }
  .brand-area { text-align: center; }
  .brand-icon-wrap { margin: 0 auto 14px; }
  .login-footer-text { display: none; }
  .login-right { padding: 28px 24px 32px; }
}
</style>
