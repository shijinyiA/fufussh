<template>
  <div
    class="hardware-monitor"
    :class="{ collapsed: isCollapsed }"
    :style="{ left: position.x + 'px', top: position.y + 'px' }"
    v-show="visible"
  >
    <div class="monitor-header" @mousedown="startDrag">
      <span class="monitor-title">
        <el-icon><Monitor /></el-icon>
        系统监控
      </span>
      <div class="monitor-actions">
        <el-button text size="small" @click="toggleCollapse">
          <el-icon><component :is="isCollapsed ? 'ArrowDown' : 'ArrowUp'" /></el-icon>
        </el-button>
        <el-button text size="small" @click="close">
          <el-icon><Close /></el-icon>
        </el-button>
      </div>
    </div>

    <div class="monitor-body" v-show="!isCollapsed">
      <div class="monitor-section loading-section" v-if="loading">
        <div class="loading-spinner"></div>
        <span>获取中...</span>
      </div>

      <template v-else>
        <div class="monitor-section">
          <div class="section-header">
            <span class="section-label">CPU使用率</span>
            <span class="section-value" :class="getUsageClass(cpuUsage)">{{ cpuUsage }}%</span>
          </div>
          <div class="usage-bar">
            <div class="usage-fill" :style="{ width: cpuUsage + '%' }" :class="getUsageClass(cpuUsage)"></div>
          </div>
        </div>

        <div class="monitor-section">
          <div class="section-header">
            <span class="section-label">内存使用</span>
            <span class="section-value" :class="getUsageClass(memoryUsage)">{{ memoryUsage }}%</span>
          </div>
          <div class="usage-bar">
            <div class="usage-fill" :style="{ width: memoryUsage + '%' }" :class="getUsageClass(memoryUsage)"></div>
          </div>
          <div class="section-detail">{{ memoryUsed }} / {{ memoryTotal }}</div>
        </div>

        <div class="monitor-section">
          <div class="section-header">
            <span class="section-label">磁盘使用</span>
            <span class="section-value" :class="getUsageClass(diskUsage)">{{ diskUsage }}%</span>
          </div>
          <div class="usage-bar">
            <div class="usage-fill" :style="{ width: diskUsage + '%' }" :class="getUsageClass(diskUsage)"></div>
          </div>
          <div class="section-detail">{{ diskUsed }} / {{ diskTotal }}</div>
        </div>

        <div class="monitor-section uptime-section">
          <span class="section-label">运行时间</span>
          <span class="uptime-value">{{ uptime }}</span>
        </div>
      </template>

      <div class="monitor-footer">
        <el-button size="small" text @click="refresh" :loading="loading">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
        <span class="last-update">上次更新: {{ lastUpdate }}</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { Monitor, Close, ArrowDown, ArrowUp, Refresh } from '@element-plus/icons-vue'
import axios from 'axios'

const props = defineProps({
  serverId: String,
  token: String,
  visible: Boolean
})

const emit = defineEmits(['close'])

const isCollapsed = ref(false)
const loading = ref(false)
const position = ref({ x: 20, y: 80 })
const dragStart = ref({ x: 0, y: 0, elX: 0, elY: 0 })
const isDragging = ref(false)

const cpuUsage = ref(0)
const memoryUsage = ref(0)
const memoryUsed = ref('0 MB')
const memoryTotal = ref('0 MB')
const diskUsage = ref(0)
const diskUsed = ref('0 GB')
const diskTotal = ref('0 GB')
const uptime = ref('-')
const lastUpdate = ref('-')

let autoRefreshTimer = null

const toggleCollapse = () => {
  isCollapsed.value = !isCollapsed.value
}

const close = () => {
  emit('close')
}

const startDrag = (e) => {
  if (e.target.closest('.monitor-actions')) return
  isDragging.value = true
  dragStart.value = {
    x: e.clientX,
    y: e.clientY,
    elX: position.value.x,
    elY: position.value.y
  }
  document.addEventListener('mousemove', onDrag)
  document.addEventListener('mouseup', stopDrag)
}

const onDrag = (e) => {
  if (!isDragging.value) return
  const dx = e.clientX - dragStart.value.x
  const dy = e.clientY - dragStart.value.y
  position.value = {
    x: Math.max(0, dragStart.value.elX + dx),
    y: Math.max(0, dragStart.value.elY + dy)
  }
}

const stopDrag = () => {
  isDragging.value = false
  document.removeEventListener('mousemove', onDrag)
  document.removeEventListener('mouseup', stopDrag)
}

const getUsageClass = (usage) => {
  if (usage >= 90) return 'critical'
  if (usage >= 70) return 'warning'
  return 'normal'
}

const refresh = async () => {
  if (loading.value) return
  if (!props.serverId && !props.token) return

  loading.value = true

  try {
    const params = {}
    if (props.serverId) params.server_id = props.serverId
    if (props.token) params.token = props.token

    const res = await axios.get('/api/hardware-stats', {
      params,
      timeout: 15000
    })

    if (res.data && res.data.success && res.data.data) {
      const data = res.data.data

      if (data.cpu_usage !== undefined) cpuUsage.value = Math.round(data.cpu_usage)
      if (data.memory_usage !== undefined) memoryUsage.value = data.memory_usage
      if (data.memory_used) memoryUsed.value = data.memory_used
      if (data.memory_total) memoryTotal.value = data.memory_total
      if (data.disk_usage !== undefined) diskUsage.value = data.disk_usage
      if (data.disk_used) diskUsed.value = data.disk_used
      if (data.disk_total) diskTotal.value = data.disk_total
      if (data.uptime) uptime.value = data.uptime

      lastUpdate.value = new Date().toLocaleTimeString()
    }
  } catch (e) {
    console.error('[HardwareMonitor] fetch error:', e)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  refresh()
  autoRefreshTimer = setInterval(refresh, 30000)
})

onBeforeUnmount(() => {
  if (autoRefreshTimer) {
    clearInterval(autoRefreshTimer)
  }
})

defineExpose({ refresh })
</script>

<style scoped>
.hardware-monitor {
  position: fixed;
  width: 260px;
  background: rgba(15, 15, 26, 0.95);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);
  z-index: 1000;
  backdrop-filter: blur(10px);
  overflow: hidden;
}

.monitor-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 12px;
  background: rgba(255, 255, 255, 0.05);
  cursor: move;
  user-select: none;
}

.monitor-title {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  font-weight: 500;
  color: #fff;
}

.monitor-actions {
  display: flex;
  gap: 2px;
}

.monitor-actions .el-button {
  padding: 4px;
  color: rgba(255, 255, 255, 0.6);
}

.monitor-actions .el-button:hover {
  color: #fff;
}

.monitor-body {
  padding: 12px;
}

.monitor-section {
  margin-bottom: 14px;
}

.monitor-section:last-child {
  margin-bottom: 0;
}

.loading-section {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px 0;
  color: rgba(255, 255, 255, 0.6);
  font-size: 13px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 6px;
}

.section-label {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.6);
}

.section-value {
  font-size: 13px;
  font-weight: 600;
}

.section-value.normal { color: #22c55e; }
.section-value.warning { color: #f59e0b; }
.section-value.critical { color: #ef4444; }

.usage-bar {
  height: 6px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 3px;
  overflow: hidden;
}

.usage-fill {
  height: 100%;
  border-radius: 3px;
  transition: width 0.3s ease;
}

.usage-fill.normal { background: linear-gradient(90deg, #22c55e, #4ade80); }
.usage-fill.warning { background: linear-gradient(90deg, #f59e0b, #fbbf24); }
.usage-fill.critical { background: linear-gradient(90deg, #ef4444, #f87171); }

.section-detail {
  font-size: 11px;
  color: rgba(255, 255, 255, 0.4);
  margin-top: 4px;
}

.uptime-section {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 8px;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

.uptime-value {
  font-size: 12px;
  color: #60a5fa;
}

.monitor-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 12px;
  padding-top: 10px;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

.last-update {
  font-size: 10px;
  color: rgba(255, 255, 255, 0.3);
}

.loading-spinner {
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255, 255, 255, 0.1);
  border-top-color: #3b82f6;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin-right: 8px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.hardware-monitor.collapsed .monitor-body {
  display: none;
}
</style>
