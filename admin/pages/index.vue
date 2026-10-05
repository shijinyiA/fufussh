<template>
  <div class="dashboard-page">
    <div class="stats-grid">
      <div class="stat-card stat-blue">
        <div class="stat-icon"><el-icon><User /></el-icon></div>
        <div class="stat-info">
          <span class="stat-num">{{ stats.userCount }}</span>
          <span class="stat-label">注册用户</span>
        </div>
      </div>
      <div class="stat-card stat-green">
        <div class="stat-icon"><el-icon><Monitor /></el-icon></div>
        <div class="stat-info">
          <span class="stat-num">{{ stats.serverCount }}</span>
          <span class="stat-label">服务器数量</span>
        </div>
      </div>
      <div class="stat-card stat-orange">
        <div class="stat-icon"><el-icon><Tools /></el-icon></div>
        <div class="stat-info">
          <span class="stat-num">{{ stats.commandCount }}</span>
          <span class="stat-label">快捷命令</span>
        </div>
      </div>
    </div>

    <div class="dashboard-row">
      <div class="dash-card quick-actions">
        <div class="card-header">
          <h3>快捷操作</h3>
          <el-button :icon="Refresh" text size="small" @click="loadAll">刷新</el-button>
        </div>
        <div class="action-grid">
          <NuxtLink to="/users" class="action-item">
            <el-icon color="#2563eb"><User /></el-icon>
            <span>用户管理</span>
          </NuxtLink>
          <NuxtLink to="/servers" class="action-item">
            <el-icon color="#059669"><Monitor /></el-icon>
            <span>服务器</span>
          </NuxtLink>
          <NuxtLink to="/settings" class="action-item">
            <el-icon color="#7c3aed"><Setting /></el-icon>
            <span>系统设置</span>
          </NuxtLink>
          <NuxtLink to="/commands" class="action-item">
            <el-icon color="#ea580c"><Tickets /></el-icon>
            <span>快捷命令</span>
          </NuxtLink>
        </div>
      </div>

      <div class="dash-card system-info">
        <div class="card-header">
          <h3>系统信息</h3>
        </div>
        <div class="info-list">
          <div class="info-row"><span class="info-label">版本</span><span class="info-value version-tag">V2.1.0 LTS</span></div>
          <div class="info-row">
            <span class="info-label">内存占用</span>
            <div class="memory-area">
              <span class="mem-text">{{ sysInfo.memoryUsed }} MB / {{ sysInfo.memoryTotal }} MB</span>
              <div class="mem-bar">
                <div class="mem-fill" :style="{ width: sysInfo.memoryUsage + '%' }"></div>
              </div>
            </div>
          </div>
          <div class="info-row"><span class="info-label">数据库</span><span class="info-value">MySQL</span></div>
          <div class="info-row"><span class="info-label">状态</span><span class="info-status online">正常运行</span></div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { User, Monitor, Tools, Refresh, Setting, Tickets } from '@element-plus/icons-vue'

definePageMeta({ middleware: ['admin-auth'] })
const { adminAPI } = await import('~/api')

const stats = ref({ userCount: 0, serverCount: 0, commandCount: 0 })
const sysInfo = ref({ memoryTotal: 0, memoryUsed: 0, memoryUsage: 0 })

async function loadStats() {
  try {
    const res = await adminAPI.stats()
    const d = res.data
    if (d.success) {
      stats.value = {
        userCount: Number(d.user_count) || 0,
        serverCount: Number(d.server_count) || 0,
        commandCount: Number(d.command_count) || 0
      }
    }
  } catch (e) { console.error(e) }
}

async function loadSysInfo() {
  try {
    const res = await adminAPI.systemInfo()
    const d = res.data
    if (d.success) {
      sysInfo.value = {
        memoryTotal: Math.round(Number(d.memory_total) || 0),
        memoryUsed: Math.round(Number(d.memory_used) || 0),
        memoryUsage: Number(d.memory_usage) || 0
      }
    }
  } catch (e) { console.error(e) }
}

async function loadAll() { await Promise.all([loadStats(), loadSysInfo()]) }

onMounted(() => loadAll())
</script>

<style scoped>
.dashboard-page { display: flex; flex-direction: column; gap: 20px; }

.stats-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
}
.stat-card {
  background: #fff;
  border-radius: 10px;
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 16px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.04), 0 1px 2px rgba(0,0,0,0.02);
  border: 1px solid #eef2f7;
  transition: transform 0.15s, box-shadow 0.15s;
}
.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(0,0,0,0.08);
}

.stat-icon {
  width: 48px; height: 48px;
  border-radius: 12px;
  display: flex; align-items: center; justify-content: center;
  font-size: 22px;
  flex-shrink: 0;
}
.stat-blue .stat-icon { background: #eff6ff; color: #2563eb; }
.stat-green .stat-icon { background: #ecfdf5; color: #059669; }
.stat-orange .stat-icon { background: #fff7ed; color: #ea580c; }

.stat-info { display: flex; flex-direction: column; gap: 2px; min-width: 0; }
.stat-num {
  font-size: 26px;
  font-weight: 800;
  line-height: 1.2;
  letter-spacing: -0.5px;
}
.stat-blue .stat-num { color: #1e40af; }
.stat-green .stat-num { color: #047857; }
.stat-orange .stat-num { color: #c2410c; }
.stat-label {
  font-size: 13px;
  color: #94a3b8;
  font-weight: 500;
}

.dashboard-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.dash-card {
  background: #fff;
  border-radius: 10px;
  padding: 20px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.04), 0 1px 2px rgba(0,0,0,0.02);
  border: 1px solid #eef2f7;
}
.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 18px;
}
.card-header h3 {
  font-size: 15px;
  font-weight: 700;
  color: #1e293b;
  margin: 0;
}

.action-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
}
.action-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 14px;
  border-radius: 8px;
  border: 1px solid #f1f5f9;
  text-decoration: none;
  font-size: 13.5px;
  font-weight: 600;
  color: #334155;
  transition: all 0.15s;
}
.action-item:hover {
  background: #f8fafc;
  border-color: #e2e8f0;
  transform: translateX(2px);
}
.action-item .el-icon { font-size: 18px; }

.info-list { display: flex; flex-direction: column; gap: 14px; }
.info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 14px;
  border-bottom: 1px solid #f5f7fa;
}
.info-row:last-child { border-bottom: none; padding-bottom: 0; }
.info-label { font-size: 13px; color: #94a3b8; font-weight: 500; }
.info-value { font-size: 13.5px; color: #334155; font-weight: 600; }
.version-tag {
  background: linear-gradient(135deg, #2563eb, #3b82f6);
  color: #fff;
  padding: 2px 10px;
  border-radius: 5px;
  font-size: 12.5px;
  letter-spacing: 0.5px;
}
.info-status {
  font-size: 12.5px;
  font-weight: 600;
  padding: 2px 10px;
  border-radius: 20px;
}
.info-status.online {
  background: #ecfdf5;
  color: #047857;
}

.memory-area {
  display: flex;
  flex-direction: column;
  gap: 5px;
  min-width: 140px;
  max-width: 180px;
}
.mem-text {
  font-size: 12px;
  font-weight: 600;
  color: #475569;
}
.mem-bar {
  width: 100%;
  height: 6px;
  background: #f1f5f9;
  border-radius: 3px;
  overflow: hidden;
}
.mem-fill {
  height: 100%;
  border-radius: 3px;
  background: linear-gradient(90deg, #2563eb, #60a5fa);
  transition: width 0.5s ease;
  min-width: 8px;
}

@media (max-width: 1023px) {
  .stats-grid { grid-template-columns: repeat(3, 1fr); }
  .dashboard-row { grid-template-columns: 1fr; }
}
@media (max-width: 480px) {
  .stats-grid { grid-template-columns: 1fr; }
  .stat-num { font-size: 22px; }
  .action-grid { grid-template-columns: 1fr; }
}
</style>
