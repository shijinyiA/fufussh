<template>
  <div class="shared-connect-page">
    <div class="connect-card" v-if="!loading && serverInfo">
      <div class="card-header">
        <h2>{{ serverInfo.server_name }}</h2>
        <p class="server-info">{{ serverInfo.username }}@{{ serverInfo.host }}:{{ serverInfo.port }}</p>
      </div>

      <div class="warning-notice">
        <el-icon><WarningFilled /></el-icon>
        <span>这是一次性分享链接，连接后将自动失效</span>
      </div>

      <div class="connect-actions">
        <el-button type="primary" size="large" @click="connectToServer" :loading="connecting" style="width: 100%">
          <el-icon><Connection /></el-icon>
          立即连接
        </el-button>
      </div>
    </div>

    <div class="error-card" v-if="!loading && error">
      <el-icon :size="48" color="#ef4444"><CircleCloseFilled /></el-icon>
      <h3>{{ error }}</h3>
      <p>请向分享者索取新的链接</p>
    </div>

    <div class="loading-card" v-if="loading">
      <el-icon class="is-loading" :size="48"><Loading /></el-icon>
      <p>正在验证分享链接...</p>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { WarningFilled, CircleCloseFilled, Loading, Connection } from '@element-plus/icons-vue'
import api from '../api'

const route = useRoute()
const router = useRouter()

const loading = ref(true)
const connecting = ref(false)
const error = ref('')
const serverInfo = ref(null)
const token = ref('')

onMounted(async () => {
  token.value = route.params.token
  if (!token.value) {
    error.value = '无效的分享链接'
    loading.value = false
    return
  }

  try {
    const res = await api.get(`/share/${token.value}`)
    if (res.data.success) {
      serverInfo.value = res.data
    } else {
      error.value = res.data.message || '分享链接无效'
    }
  } catch (err) {
    if (err.response?.status === 410) {
      error.value = '此分享链接已过期或已被使用'
    } else {
      error.value = '验证分享链接失败'
    }
  } finally {
    loading.value = false
  }
})

const connectToServer = async () => {
  connecting.value = true
  try {
    const res = await api.post('/direct/connect', {
      host: serverInfo.value.host,
      port: serverInfo.value.port,
      username: serverInfo.value.username,
      password: serverInfo.value.password,
      private_key: serverInfo.value.private_key,
      share_token: token.value
    })

    if (res.data.success) {
      router.push(`/direct/${res.data.token}`)
    } else {
      ElMessage.error(res.data.message || '连接失败')
    }
  } catch (err) {
    ElMessage.error('连接服务器失败')
  } finally {
    connecting.value = false
  }
}
</script>

<style scoped>
.shared-connect-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #0f0f1a 0%, #1a1a2e 100%);
  padding: 20px;
}

.connect-card, .error-card, .loading-card {
  background: #fff;
  border-radius: 16px;
  padding: 32px;
  width: 100%;
  max-width: 400px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
}

.card-header {
  text-align: center;
  margin-bottom: 24px;
}

.card-header h2 {
  margin: 0 0 8px 0;
  font-size: 24px;
  color: #1f2937;
}

.server-info {
  margin: 0;
  color: #6b7280;
  font-family: 'Courier New', monospace;
}

.warning-notice {
  display: flex;
  align-items: center;
  gap: 8px;
  background: #fef3c7;
  color: #92400e;
  padding: 12px 16px;
  border-radius: 8px;
  margin-bottom: 24px;
  font-size: 14px;
}

.connect-actions {
  margin-top: 16px;
}

.error-card {
  text-align: center;
}

.error-card h3 {
  margin: 16px 0 8px 0;
  color: #ef4444;
}

.error-card p {
  margin: 0;
  color: #6b7280;
}

.loading-card {
  text-align: center;
}

.loading-card p {
  margin-top: 16px;
  color: #6b7280;
}

.is-loading {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
