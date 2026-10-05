<template>
  <div class="home-page">

    <aside class="sidebar">
      <div class="sidebar-search">
        <el-input
          v-model="serverStore.searchQuery"
          placeholder="搜索服务器..."
          :prefix-icon="Search"
          clearable
          size="default"
        />
      </div>

      <div class="sidebar-groups">
        <div
          class="group-item"
          :class="{ active: !serverStore.currentGroup }"
          @click="serverStore.currentGroup = ''"
        >
          <el-icon><Monitor /></el-icon>
          <span>全部服务器</span>
          <el-badge :value="serverStore.servers.length" :max="99" />
        </div>
        <div
          v-for="group in serverStore.groups"
          :key="group.name"
          class="group-item"
          :class="{ active: serverStore.currentGroup === group.name }"
          @click="serverStore.currentGroup = group.name"
        >
          <el-icon><Folder /></el-icon>
          <span>{{ group.name }}</span>
          <el-badge :value="group.count" :max="99" />
        </div>
      </div>

      <div class="sidebar-footer">
        <div class="user-info" @click="router.push('/profile')">
          <el-icon><User /></el-icon>
          <span>{{ username }} - 设置</span>
        </div>
        <div class="user-info logout" @click="handleLogout">
          <el-icon><SwitchButton /></el-icon>
          <span>退出登录</span>
        </div>
      </div>
    </aside>

    <main class="main-content">
      <div class="content-header">
        <h2>服务器列表</h2>
        <el-button type="primary" @click="showAddDialog = true" :icon="Plus">
          添加服务器
        </el-button>
      </div>

      <div class="server-grid" v-loading="serverStore.loading">
        <div
          v-for="server in serverStore.filteredServers"
          :key="server.id"
          class="server-card"
          :style="{ '--card-color': server.color || '#3B82F6' }"
          @animationend="server._animating = false"
        >
          <div class="card-accent"></div>
          <div class="card-body">
            <div class="card-header">
              <div class="server-icon" :style="{ background: server.color || '#3B82F6' }">
                {{ server.name.charAt(0).toUpperCase() }}
              </div>
              <div class="card-actions">
                <el-dropdown trigger="click" @command="handleCommand($event, server)">
                  <el-button text size="small" :icon="MoreFilled" />
                  <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item command="edit" :icon="Edit">编辑</el-dropdown-item>
                      <el-dropdown-item command="files" :icon="FolderOpened">文件管理</el-dropdown-item>
                      <el-dropdown-item command="share" :icon="Share">分享连接</el-dropdown-item>
                      <el-dropdown-item command="test" :icon="Connection">测试连接</el-dropdown-item>
                      <el-dropdown-item command="delete" :icon="Delete" divided danger>删除</el-dropdown-item>
                    </el-dropdown-menu>
                  </template>
                </el-dropdown>
              </div>
            </div>

            <h3 class="server-name">{{ server.name }}</h3>
            <p class="server-info">
              <el-icon><User /></el-icon>
              {{ server.username }}@{{ server.host }}:{{ server.port }}
            </p>

            <div class="card-tags" v-if="server.tags && server.tags.length">
              <el-tag
                v-for="tag in server.tags"
                :key="tag"
                size="small"
                effect="dark"
                type="info"
              >
                {{ tag }}
              </el-tag>
            </div>

            <p class="server-desc" v-if="server.description">{{ server.description }}</p>

            <div class="card-footer">
              <el-button
                type="primary"
                size="default"
                @click="connectServer(server)"
                :icon="VideoPlay"
                class="connect-btn"
              >
                连接终端
              </el-button>
            </div>
          </div>
        </div>

        <div v-if="!serverStore.loading && serverStore.filteredServers.length === 0" class="empty-state">
          <div class="empty-icon">
            <el-icon :size="64" color="var(--text-muted)"><Monitor /></el-icon>
          </div>
          <h3>{{ serverStore.searchQuery ? '未找到匹配的服务器' : '暂无服务器' }}</h3>
          <p>{{ serverStore.searchQuery ? '尝试修改搜索关键词' : '点击右上角按钮添加第一台服务器' }}</p>
          <el-button v-if="!serverStore.searchQuery" type="primary" @click="showAddDialog = true" :icon="Plus">
            添加服务器
          </el-button>
        </div>
      </div>
    </main>

    <el-dialog
      v-model="showAddDialog"
      :title="editingServer ? '编辑服务器' : '添加服务器'"
      width="520px"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <el-form
        ref="serverFormRef"
        :model="serverForm"
        :rules="serverRules"
        label-width="90px"
        label-position="left"
      >
        <el-form-item label="名称" prop="name">
          <el-input v-model="serverForm.name" placeholder="服务器显示名称" />
        </el-form-item>
        <el-form-item label="主机" prop="host">
          <el-input v-model="serverForm.host" placeholder="IP地址或域名" />
        </el-form-item>
        <el-form-item label="端口" prop="port">
          <el-input-number v-model="serverForm.port" :min="1" :max="65535" :step="1" />
        </el-form-item>
        <el-form-item label="用户名" prop="username">
          <el-input v-model="serverForm.username" placeholder="SSH登录用户名" />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input v-model="serverForm.password" type="password" show-password placeholder="SSH登录密码" />
        </el-form-item>
        <el-form-item label="私钥">
          <el-input
            v-model="serverForm.private_key"
            type="textarea"
            :rows="3"
            placeholder="SSH私钥内容（可选，优先使用私钥认证）"
          />
        </el-form-item>
        <el-form-item label="分组">
          <el-input v-model="serverForm.group" placeholder="分组名称（可选）" />
        </el-form-item>
        <el-form-item label="标签">
          <el-select
            v-model="serverForm.tags"
            multiple
            filterable
            allow-create
            default-first-option
            placeholder="添加标签"
          >
            <el-option v-for="tag in commonTags" :key="tag" :label="tag" :value="tag" />
          </el-select>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="serverForm.description" type="textarea" :rows="2" placeholder="服务器描述（可选）" />
        </el-form-item>
        <el-form-item label="颜色">
          <el-color-picker v-model="serverForm.color" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="showAddDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSubmitServer" :loading="submitting">
          {{ editingServer ? '保存' : '添加' }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showShareDialog" title="分享服务器连接" width="450px">
      <div class="share-content">
        <p class="share-warning">
          <el-icon><WarningFilled /></el-icon>
          此链接仅可使用一次，有效期24小时
        </p>
        <el-input v-model="shareUrl" readonly>
          <template #append>
            <el-button @click="copyShareUrl">复制</el-button>
          </template>
        </el-input>
        <p class="share-hint">将此链接发送给需要访问的人</p>
      </div>
    </el-dialog>

    <div class="page-footer">
      <span>若叶睦&锦衣</span>
      <span v-if="footerText"> | {{ footerText }}</span>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Search, Monitor, Folder, Plus, Edit, Delete, MoreFilled,
  VideoPlay, FolderOpened, Connection, User, SwitchButton, Share, WarningFilled
} from '@element-plus/icons-vue'
import { useServerStore } from '../stores/server'
import { authAPI, serverAPI } from '../api'

const router = useRouter()
const serverStore = useServerStore()

const username = computed(() => {
  return document.cookie
    .split('; ')
    .find(row => row.startsWith('web_ssh_user='))
    ?.split('=')[1] || '用户'
})

const showAddDialog = ref(false)
const editingServer = ref(null)
const submitting = ref(false)
const serverFormRef = ref(null)

const showShareDialog = ref(false)
const shareUrl = ref('')
const footerText = ref('')

const commonTags = ['生产环境', '测试环境', '开发环境', '数据库', 'Web', 'API', '内部', '外部']

const serverForm = reactive({
  name: '',
  host: '',
  port: 22,
  username: '',
  password: '',
  private_key: '',
  group: '',
  tags: [],
  description: '',
  color: '#3B82F6'
})

const serverRules = {
  name: [{ required: true, message: '请输入服务器名称', trigger: 'blur' }],
  host: [{ required: true, message: '请输入主机地址', trigger: 'blur' }],
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }]
}

const resetForm = () => {
  Object.assign(serverForm, {
    name: '',
    host: '',
    port: 22,
    username: '',
    password: '',
    private_key: '',
    group: '',
    tags: [],
    description: '',
    color: '#3B82F6'
  })
  editingServer.value = null
}

const connectServer = (server) => {
  router.push(`/terminal/${server.id}`)
}

const openFileManager = (server) => {
  router.push(`/files/${server.id}`)
}

const handleCommand = async (command, server) => {
  switch (command) {
    case 'edit':
      editingServer.value = server
      Object.assign(serverForm, {
        name: server.name,
        host: server.host,
        port: server.port,
        username: server.username,
        password: '',
        private_key: '',
        group: server.group || '',
        tags: server.tags || [],
        description: server.description || '',
        color: server.color || '#3B82F6'
      })
      showAddDialog.value = true
      break

    case 'share':
      try {
        const res = await serverAPI.share(server.id)
        if (res.data.success) {
          shareUrl.value = window.location.origin + res.data.share_url
          showShareDialog.value = true
        } else {
          ElMessage.error(res.data.message || '创建分享链接失败')
        }
      } catch (err) {
        ElMessage.error('创建分享链接失败')
      }
      break

    case 'test':
      try {
        ElMessage.info('正在测试连接...')
        const result = await serverStore.testConnection(server.id)
        if (result.success) {
          ElMessage.success('连接测试成功！')
        } else {
          ElMessage.error(`连接失败: ${result.message}`)
        }
      } catch (err) {
        ElMessage.error('连接测试失败')
      }
      break

    case 'delete':
      try {
        await ElMessageBox.confirm(
          `确定要删除服务器 "${server.name}" 吗？`,
          '确认删除',
          { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' }
        )
        await serverStore.deleteServer(server.id)
        ElMessage.success('删除成功')
      } catch {}
      break
  }
}

const handleSubmitServer = async () => {
  if (!serverFormRef.value) return

  await serverFormRef.value.validate(async (valid) => {
    if (!valid) return

    submitting.value = true
    try {
      if (editingServer.value) {
        const result = await serverStore.updateServer(editingServer.value.id, serverForm)
        if (result.success) {
          ElMessage.success('更新成功')
          showAddDialog.value = false
          resetForm()
        } else {
          ElMessage.error(result.message || '更新失败')
        }
      } else {
        const result = await serverStore.createServer({ ...serverForm })
        if (result.success) {
          ElMessage.success('添加成功')
          showAddDialog.value = false
          resetForm()
        } else {
          ElMessage.error(result.message || '添加失败')
        }
      }
    } catch (err) {
      ElMessage.error('操作失败')
    } finally {
      submitting.value = false
    }
  })
}

const handleLogout = async () => {
  try {
    await authAPI.logout()
    router.push('/login')
  } catch {}
}

const copyShareUrl = async () => {
  try {
    await navigator.clipboard.writeText(shareUrl.value)
    ElMessage.success('链接已复制到剪贴板')
  } catch {
    ElMessage.error('复制失败，请手动复制')
  }
}

import { watch } from 'vue'
watch(showAddDialog, (val) => {
  if (!val) resetForm()
})

onMounted(async () => {
  serverStore.fetchServers()

  try {
    const res = await fetch('/api/system-settings')
    const data = await res.json()
    if (data.success && data.settings?.footer_text) {
      footerText.value = data.settings.footer_text
    }
  } catch (err) {
    console.error('加载设置失败', err)
  }
})
</script>

<style scoped>
.home-page {
  height: 100vh;
  display: flex;
  background: var(--bg-primary);
}

.sidebar {
  width: 260px;
  background: #ffffff;
  border-right: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
  box-shadow: 2px 0 12px rgba(37, 99, 235, 0.04);
}

.sidebar-search {
  padding: 16px 16px 8px;
}

.sidebar-groups {
  flex: 1;
  overflow-y: auto;
  padding: 0 12px;
}

.group-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: var(--transition);
  color: var(--text-secondary);
  margin-bottom: 4px;
}

.group-item:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.group-item.active {
  background: var(--accent-dim);
  color: var(--accent);
}

.group-item span {
  flex: 1;
  font-size: 14px;
}

.sidebar-footer {
  padding: 16px;
  border-top: 1px solid var(--border);
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--text-muted);
  cursor: pointer;
  padding: 8px;
  border-radius: var(--radius-sm);
  transition: var(--transition);
}

.user-info:hover {
  background: var(--bg-hover);
  color: var(--danger);
}

.user-info.logout:hover {
  color: var(--danger);
}

.share-content {
  padding: 8px 0;
}

.share-warning {
  display: flex;
  align-items: center;
  gap: 8px;
  background: #fef3c7;
  color: #92400e;
  padding: 12px 16px;
  border-radius: 8px;
  margin-bottom: 16px;
  font-size: 14px;
}

.share-hint {
  margin-top: 12px;
  font-size: 13px;
  color: var(--text-muted);
}

.main-content {
  flex: 1;
  overflow-y: auto;
  padding: 24px 32px;
}

.content-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.content-header h2 {
  font-size: 22px;
  color: var(--text-primary);
}

.server-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 20px;
}

.server-card {
  background: #ffffff;
  border: 1px solid var(--border);
  border-radius: var(--radius);
  overflow: hidden;
  transition: var(--transition);
  position: relative;
  animation: fadeIn 0.4s ease;
}

.server-card:hover {
  border-color: var(--card-color);
  transform: translateY(-2px);
  box-shadow: 0 8px 32px rgba(37, 99, 235, 0.12);
}

.card-accent {
  height: 3px;
  background: var(--card-color);
}

.card-body {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 12px;
}

.server-icon {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-weight: bold;
  font-size: 18px;
}

.card-actions {
  opacity: 0;
  transition: opacity 0.2s;
}

.server-card:hover .card-actions {
  opacity: 1;
}

.server-name {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 6px;
}

.server-info {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--text-secondary);
  margin-bottom: 10px;
  font-family: 'Courier New', monospace;
}

.card-tags {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
  margin-bottom: 10px;
}

.server-desc {
  font-size: 12px;
  color: var(--text-muted);
  margin-bottom: 16px;
  line-height: 1.5;
}

.card-footer {
  display: flex;
  gap: 10px;
}

.connect-btn {
  flex: 1;
}

.empty-state {
  grid-column: 1 / -1;
  text-align: center;
  padding: 80px 20px;
}

.empty-icon {
  margin-bottom: 20px;
}

.empty-state h3 {
  font-size: 18px;
  color: var(--text-secondary);
  margin-bottom: 8px;
}

.empty-state p {
  font-size: 14px;
  color: var(--text-muted);
  margin-bottom: 20px;
}

.page-footer {
  text-align: center;
  padding: 16px;
  font-size: 12px;
  color: var(--text-muted);
  border-top: 1px solid var(--border);
}

@media (max-width: 768px) {
  .home-page { flex-direction: column; }
  .sidebar {
    width: 100%; height: auto; max-height: none;
    border-right: none; border-bottom: 1px solid var(--border);
    flex-direction: row; align-items: center; flex-wrap: wrap;
    padding: 0;
  }
  .sidebar-search {
    flex: 1; min-width: 0; padding: 10px 12px;
  }
  .sidebar-groups {
    display: none;
  }
  .sidebar-footer {
    border-top: none; padding: 8px 10px;
  }
  .user-info { font-size: 11px; padding: 6px 8px; }

  .main-content { padding: 16px; overflow-y: auto; -webkit-overflow-scrolling: touch; }
  .content-header h2 { font-size: 18px; margin-bottom: 4px; }
  .server-grid { grid-template-columns: 1fr; gap: 14px; }
  .server-card:hover { transform: none; box-shadow: none; }
  .card-actions { opacity: 1; }
  .card-body { padding: 16px; }
  .server-name { font-size: 15px; }
  .server-info { font-size: 12px; }
}

@media (max-width: 480px) {
  .main-content { padding: 12px; }
  .server-grid { gap: 12px; }
  .card-body { padding: 14px; }
  .server-icon { width: 34px; height: 34px; font-size: 15px; }
  .page-footer { padding: 12px; font-size: 11px; }
}
</style>
