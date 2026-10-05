<template>
  <div class="direct-create-page">
    <div class="loading-content">
      <div class="spinner"></div>
      <h3>正在创建连接...</h3>
      <p>{{ status }}</p>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()
const status = ref('正在准备...')

const createConnection = async () => {
  const { host, port, username, password } = route.query

  if (!host || !username || !password) {
    status.value = '参数错误：缺少必要的连接信息'
    setTimeout(() => {
      router.push('/login')
    }, 2000)
    return
  }

  status.value = '正在连接服务器...'

  try {
    const params = new URLSearchParams({
      host,
      port: port || '22',
      username,
      password
    })

    const res = await fetch(`/api/direct/connect?${params}`, {
      method: 'POST'
    })
    const data = await res.json()

    if (data.success && data.token) {
      status.value = '连接已创建，正在跳转...'
      setTimeout(() => {
        router.push(`/direct/${data.token}`)
      }, 500)
    } else {
      throw new Error(data.message || '创建连接失败')
    }
  } catch (e) {
    status.value = `创建连接失败: ${e.message}`
    setTimeout(() => {
      router.push('/login')
    }, 2000)
  }
}

onMounted(() => {
  createConnection()
})
</script>

<style scoped>
.direct-create-page {
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #0f0f1a;
}

.loading-content {
  text-align: center;
  padding: 20px;
}

.spinner {
  width: 44px;
  height: 44px;
  border: 3px solid rgba(255,255,255,0.1);
  border-top-color: #00d4ff;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin: 0 auto 20px;
}

h3 {
  font-size: 17px;
  color: #e0e0e0;
  margin-bottom: 8px;
}

p {
  font-size: 13px;
  color: #888;
  margin-bottom: 22px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}
</style>
