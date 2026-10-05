<template>
  <div class="file-manager-page">

    <div class="fm-topbar">
      <div class="topbar-left">
        <el-button text @click="goBack" :icon="ArrowLeft">返回</el-button>
        <span class="fm-title">文件管理 - {{ serverInfo?.name || '' }}</span>
      </div>
      <div class="topbar-right">
        <el-button text @click="openTerminal" :icon="Monitor">终端</el-button>
      </div>
    </div>

    <div class="fm-toolbar">
      <div class="toolbar-left">
        <el-button @click="goUp" :icon="Top" :disabled="currentPath === '/'" size="small">上级目录</el-button>
        <el-button @click="goHome" :icon="HomeFilled" size="small">根目录</el-button>
        <el-button @click="refreshList" :icon="Refresh" :loading="loading" size="small">刷新</el-button>
      </div>
      <div class="toolbar-center">
        <div class="path-breadcrumb">
          <span
            v-for="(segment, index) in pathSegments"
            :key="index"
            class="path-segment"
            @click="navigateToPath(segment.path)"
          >
            <el-icon v-if="index > 0"><ArrowRight /></el-icon>
            {{ segment.name }}
          </span>
        </div>
      </div>
      <div class="toolbar-right">
        <el-button @click="showNewFolderDialog = true" :icon="FolderAdd" size="small">新建文件夹</el-button>
        <el-button @click="triggerUpload" :icon="Upload" size="small">上传</el-button>
        <input ref="fileInput" type="file" style="display: none" @change="handleFileUpload" multiple />
      </div>
    </div>

    <div class="fm-search">
      <el-input
        v-model="searchQuery"
        placeholder="搜索文件名..."
        :prefix-icon="Search"
        clearable
        size="small"
        @keyup.enter="searchFiles"
      />
    </div>

    <div class="fm-content" v-loading="loading">
      <table class="file-table">
        <thead>
          <tr>
            <th class="col-icon"></th>
            <th class="col-name" @click="sortBy('name')">
              名称
              <el-icon v-if="sortField === 'name'" size="12">
                <SortUp v-if="sortOrder === 'asc'" />
                <SortDown v-else />
              </el-icon>
            </th>
            <th class="col-size" @click="sortBy('size')">
              大小
              <el-icon v-if="sortField === 'size'" size="12">
                <SortUp v-if="sortOrder === 'asc'" />
                <SortDown v-else />
              </el-icon>
            </th>
            <th class="col-mode">权限</th>
            <th class="col-time" @click="sortBy('mod_time')">
              修改时间
              <el-icon v-if="sortField === 'mod_time'" size="12">
                <SortUp v-if="sortOrder === 'asc'" />
                <SortDown v-else />
              </el-icon>
            </th>
            <th class="col-actions">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="currentPath !== '/'" @click="goUp" class="file-row">
            <td class="col-icon"><el-icon><Folder /></el-icon></td>
            <td class="col-name" colspan="4">..</td>
            <td></td>
          </tr>
          <tr
            v-for="file in sortedFiles"
            :key="file.path"
            class="file-row"
            :class="{ selected: selectedFile?.path === file.path }"
            @click="selectedFile = file"
            @dblclick="handleDoubleClick(file)"
          >
            <td class="col-icon">
              <el-icon :color="getFileIconColor(file)">
                <Folder v-if="file.is_dir" />
                <Document v-else-if="isTextFile(file.name)" />
                <Picture v-else-if="isImageFile(file.name)" />
                <VideoPlay v-else-if="isVideoFile(file.name)" />
                <Document v-else />
              </el-icon>
            </td>
            <td class="col-name">
              <span class="file-name" :class="{ 'is-dir': file.is_dir }">
                {{ file.name }}
              </span>
            </td>
            <td class="col-size">
              <span v-if="file.is_dir">-</span>
              <span v-else>{{ formatSize(file.size) }}</span>
            </td>
            <td class="col-mode">{{ file.mode }}</td>
            <td class="col-time">{{ formatDate(file.mod_time) }}</td>
            <td class="col-actions">
              <el-dropdown trigger="click" @command="handleFileAction($event, file)">
                <el-button text size="small" :icon="MoreFilled" @click.stop />
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item v-if="!file.is_dir" command="download" :icon="Download">下载</el-dropdown-item>
                    <el-dropdown-item v-if="isTextFile(file.name)" command="edit" :icon="Edit">编辑</el-dropdown-item>
                    <el-dropdown-item command="rename" :icon="EditPen">重命名</el-dropdown-item>
                    <el-dropdown-item command="delete" :icon="Delete" divided danger>删除</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </td>
          </tr>
          <tr v-if="files.length === 0 && !loading">
            <td colspan="6" class="empty-row">
              <div class="empty-files">
                <el-icon :size="48" color="var(--text-muted)"><FolderOpened /></el-icon>
                <p>空目录</p>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="fm-statusbar">
      <span>{{ files.length }} 项</span>
      <span v-if="selectedFile">选中: {{ selectedFile.name }}</span>
    </div>

    <el-dialog v-model="showNewFolderDialog" title="新建文件夹" width="400px" destroy-on-close>
      <el-input v-model="newFolderName" placeholder="文件夹名称" @keyup.enter="createFolder" />
      <template #footer>
        <el-button @click="showNewFolderDialog = false">取消</el-button>
        <el-button type="primary" @click="createFolder" :loading="actionLoading">创建</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showRenameDialog" title="重命名" width="400px" destroy-on-close>
      <el-input v-model="renameName" placeholder="新名称" @keyup.enter="renameFile" />
      <template #footer>
        <el-button @click="showRenameDialog = false">取消</el-button>
        <el-button type="primary" @click="renameFile" :loading="actionLoading">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="showEditorDialog"
      :title="`编辑 - ${editingFileName}`"
      width="80%"
      top="5vh"
      destroy-on-close
    >
      <div class="editor-container">
        <el-input
          v-model="editingContent"
          type="textarea"
          :rows="30"
          class="code-editor"
        />
      </div>
      <template #footer>
        <el-button @click="showEditorDialog = false">取消</el-button>
        <el-button type="primary" @click="saveFileContent" :loading="actionLoading">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showSearchDialog" title="搜索结果" width="600px">
      <div class="search-results" v-if="searchResults.length">
        <div v-for="file in searchResults" :key="file.path" class="search-item" @click="navigateToPath(file.path)">
          <el-icon><Folder v-if="file.is_dir" /><Document v-else /></el-icon>
          <span>{{ file.path }}</span>
        </div>
      </div>
      <div v-else class="empty-search">
        <p>未找到匹配文件</p>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  ArrowLeft, ArrowRight, Top, HomeFilled, Refresh, FolderAdd,
  Upload, Download, Edit, EditPen, Delete, MoreFilled, Folder,
  Document, Picture, VideoPlay, FolderOpened, Monitor, Search,
  SortUp, SortDown
} from '@element-plus/icons-vue'
import { serverAPI } from '../api'

const route = useRoute()
const router = useRouter()

const serverId = route.params.id
const isDirectConnection = route.params.token || route.path.startsWith('/sftp/')
const directToken = route.params.token || serverId
const serverInfo = ref(null)
const files = ref([])
const currentPath = ref('/')
const loading = ref(false)
const actionLoading = ref(false)
const selectedFile = ref(null)
const sortField = ref('name')
const sortOrder = ref('asc')
const searchQuery = ref('')

const showNewFolderDialog = ref(false)
const newFolderName = ref('')
const showRenameDialog = ref(false)
const renameName = ref('')
const renameFileRef = ref(null)
const showEditorDialog = ref(false)
const editingFileName = ref('')
const editingFilePath = ref('')
const editingContent = ref('')
const showSearchDialog = ref(false)
const searchResults = ref([])
const fileInput = ref(null)

let ws = null

const sortedFiles = computed(() => {
  const list = [...files.value]
  list.sort((a, b) => {

    if (a.is_dir && !b.is_dir) return -1
    if (!a.is_dir && b.is_dir) return 1

    let cmp = 0
    switch (sortField.value) {
      case 'name': cmp = a.name.localeCompare(b.name); break
      case 'size': cmp = a.size - b.size; break
      case 'mod_time': cmp = new Date(a.mod_time) - new Date(b.mod_time); break
    }
    return sortOrder.value === 'asc' ? cmp : -cmp
  })
  return list
})

const pathSegments = computed(() => {
  const segments = currentPath.value.split('/').filter(Boolean)
  let path = ''
  const result = [{ name: '/', path: '/' }]
  segments.forEach(seg => {
    path += '/' + seg
    result.push({ name: seg, path })
  })
  return result
})

const sortBy = (field) => {
  if (sortField.value === field) {
    sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortField.value = field
    sortOrder.value = 'asc'
  }
}

const connectSFTP = () => {
  const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const wsUrl = isDirectConnection
    ? `${proto}//${window.location.host}/api/direct/ws/sftp?token=${directToken}`
    : `${proto}//${window.location.host}/api/ws/sftp?server_id=${serverId}`
  console.log('SFTP connecting to:', wsUrl)
  ws = new WebSocket(wsUrl)

  ws.onopen = () => {
    console.log('SFTP connected')
    listDirectory(currentPath.value)
  }

  ws.onmessage = (event) => {
    try {
      const msg = JSON.parse(event.data)
      console.log('SFTP message:', msg)
      handleSFTPMessage(msg)
    } catch (e) {
      console.error('SFTP message parse error:', e)
    }
  }

  ws.onerror = (e) => {
    console.error('SFTP error:', e)
    ElMessage.error('SFTP连接失败')
  }

  ws.onclose = (e) => {
    console.log('SFTP closed:', e.code, e.reason)
  }
}

let pendingCallbacks = []
let messageQueue = []

const sendSFTPRequest = (req) => {
  return new Promise((resolve, reject) => {
    pendingCallbacks.push({ resolve, reject })
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify(req))
    } else {
      pendingCallbacks.pop()
      reject(new Error('SFTP未连接'))
    }
  })
}

const handleSFTPMessage = (msg) => {
  if (msg.type === 'connected') return
  if (msg.type === 'error') {
    const callback = pendingCallbacks.shift()
    if (callback) {
      callback.reject(new Error(msg.data || 'SFTP错误'))
    } else {
      ElMessage.error(msg.data || 'SFTP错误')
    }
    return
  }

  if (msg.success !== undefined) {
    const callback = pendingCallbacks.shift()

    if (msg.files) {
      files.value = msg.files
    }

    if (callback) {
      if (msg.success) {
        callback.resolve(msg)
      } else {
        callback.reject(new Error(msg.message || '操作失败'))
      }
    } else {
      if (msg.message && !msg.files) {
        if (msg.success) {
          ElMessage.success(msg.message)
        } else {
          ElMessage.error(msg.message)
        }
      }
    }
  }
}

const listDirectory = async (path) => {
  loading.value = true
  try {
    await sendSFTPRequest({ action: 'list', path })
  } catch (e) {
    ElMessage.error('获取目录列表失败')
  } finally {
    setTimeout(() => { loading.value = false }, 300)
  }
}

const goUp = () => {
  const parent = currentPath.value.substring(0, currentPath.value.lastIndexOf('/'))
  navigateToPath(parent || '/')
}

const goHome = () => {
  navigateToPath('/')
}

const navigateToPath = (path) => {
  if (selectedFile.value?.is_dir && path === selectedFile.value.path) {

  }
  currentPath.value = path
  listDirectory(path)
  selectedFile.value = null
}

const refreshList = () => {
  listDirectory(currentPath.value)
}

const handleDoubleClick = (file) => {
  if (file.is_dir) {
    navigateToPath(file.path)
  } else if (isTextFile(file.name)) {
    openFileEditor(file)
  } else {
    ElMessage.info('双击文件夹进入，双击文本文件编辑')
  }
}

const handleFileAction = async (command, file) => {
  switch (command) {
    case 'download':
      downloadFile(file)
      break
    case 'edit':
      openFileEditor(file)
      break
    case 'rename':
      renameFileRef.value = file
      renameName.value = file.name
      showRenameDialog.value = true
      break
    case 'delete':
      try {
        await ElMessageBox.confirm(
          `确定要删除 "${file.name}" 吗？${file.is_dir ? '这将递归删除目录下所有文件！' : ''}`,
          '确认删除',
          { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' }
        )
        await sendSFTPRequest({ action: 'rm', path: file.path })
        refreshList()
      } catch {}
      break
  }
}

const createFolder = async () => {
  if (!newFolderName.value) {
    ElMessage.warning('请输入文件夹名称')
    return
  }
  actionLoading.value = true
  try {
    const path = currentPath.value === '/'
      ? `/${newFolderName.value}`
      : `${currentPath.value}/${newFolderName.value}`
    await sendSFTPRequest({ action: 'mkdir', path })
    showNewFolderDialog.value = false
    newFolderName.value = ''
    refreshList()
  } catch (e) {
    ElMessage.error('创建文件夹失败')
  } finally {
    actionLoading.value = false
  }
}

const renameFile = async () => {
  if (!renameName.value) {
    ElMessage.warning('请输入新名称')
    return
  }
  actionLoading.value = true
  try {
    const parentDir = renameFileRef.value.path.substring(0, renameFileRef.value.path.lastIndexOf('/'))
    const newPath = parentDir ? `${parentDir}/${renameName.value}` : `/${renameName.value}`
    await sendSFTPRequest({ action: 'rename', path: renameFileRef.value.path, dst: newPath })
    showRenameDialog.value = false
    refreshList()
  } catch (e) {
    ElMessage.error('重命名失败')
  } finally {
    actionLoading.value = false
  }
}

const downloadFile = (file) => {
  if (isDirectConnection) {
    window.open(`/api/direct/sftp/download?token=${directToken}&path=${encodeURIComponent(file.path)}`, '_blank')
  } else {
    window.open(`/api/sftp/download?server_id=${serverId}&path=${encodeURIComponent(file.path)}`, '_blank')
  }
}

const openFileEditor = async (file) => {
  try {
    const res = await sendSFTPRequest({ action: 'read', path: file.path })
    editingFileName.value = file.name
    editingFilePath.value = file.path

    try {
      editingContent.value = atob(res.data || '')
    } catch {
      editingContent.value = res.data || ''
    }
    showEditorDialog.value = true
  } catch (e) {
    ElMessage.error('打开文件失败')
  }
}

const saveFileContent = async () => {
  actionLoading.value = true
  try {
    const content = btoa(unescape(encodeURIComponent(editingContent.value)))
    await sendSFTPRequest({ action: 'write', path: editingFilePath.value, content })
    showEditorDialog.value = false
    refreshList()
  } catch (e) {
    ElMessage.error('保存文件失败')
  } finally {
    actionLoading.value = false
  }
}

const triggerUpload = () => {
  fileInput.value?.click()
}

const handleFileUpload = async (event) => {
  const files = event.target.files
  if (!files || files.length === 0) return

  for (const file of files) {
    try {
      const reader = new FileReader()
      reader.onload = async (e) => {
        const base64 = e.target.result.split(',')[1]
        try {
          await sendSFTPRequest({
            action: 'upload',
            path: currentPath.value,
            name: file.name,
            content: base64
          })
          refreshList()
        } catch (err) {
          ElMessage.error(`上传 ${file.name} 失败`)
        }
      }
      reader.readAsDataURL(file)
    } catch (e) {
      ElMessage.error(`读取 ${file.name} 失败`)
    }
  }

  event.target.value = ''
}

const searchFiles = async () => {
  if (!searchQuery.value) return
  try {
    const res = await sendSFTPRequest({ action: 'search', path: currentPath.value, name: searchQuery.value })
    if (res.files) {
      searchResults.value = res.files
      showSearchDialog.value = true
    }
  } catch (e) {
    ElMessage.error('搜索失败')
  }
}

const formatSize = (bytes) => {
  if (bytes < 0) return '-'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let size = bytes
  let unitIndex = 0
  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024
    unitIndex++
  }
  return unitIndex === 0 ? `${size} B` : `${size.toFixed(1)} ${units[unitIndex]}`
}

const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  return d.toLocaleString('zh-CN')
}

const isTextFile = (name) => {
  const exts = ['.txt', '.log', '.conf', '.cfg', '.ini', '.yml', '.yaml', '.json', '.xml',
    '.sh', '.bash', '.py', '.js', '.ts', '.go', '.java', '.c', '.cpp', '.h',
    '.css', '.scss', '.html', '.vue', '.jsx', '.tsx', '.md', '.sql', '.env',
    '.toml', '.properties', '.dockerfile', '.gitignore', '.makefile', '.rs']
  const lower = name.toLowerCase()
  return exts.some(ext => lower.endsWith(ext)) || name.startsWith('.')
}

const isImageFile = (name) => {
  const exts = ['.jpg', '.jpeg', '.png', '.gif', '.svg', '.webp', '.bmp', '.ico']
  return exts.some(ext => name.toLowerCase().endsWith(ext))
}

const isVideoFile = (name) => {
  const exts = ['.mp4', '.avi', '.mkv', '.mov', '.wmv', '.flv']
  return exts.some(ext => name.toLowerCase().endsWith(ext))
}

const getFileIconColor = (file) => {
  if (file.is_dir) return '#FFA500'
  if (isImageFile(file.name)) return '#2ed573'
  if (isVideoFile(file.name)) return '#ff4757'
  if (isTextFile(file.name)) return '#00d4ff'
  return '#8892b0'
}

const goBack = () => {
  router.push('/')
}

const openTerminal = () => {
  router.push(`/terminal/${serverId}`)
}

onMounted(async () => {
  if (!isDirectConnection) {
    try {
      const res = await serverAPI.get(serverId)
      if (res.data.success) {
        serverInfo.value = res.data.server
      }
    } catch {}
  }

  connectSFTP()
})

onBeforeUnmount(() => {
  if (ws) {
    ws.close()
  }
})
</script>

<style scoped>
.file-manager-page {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: var(--bg-primary);
}

.fm-topbar {
  height: 48px;
  background: var(--bg-secondary);
  border-bottom: 1px solid var(--border);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  flex-shrink: 0;
}

.topbar-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.fm-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.fm-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 16px;
  background: var(--bg-secondary);
  border-bottom: 1px solid var(--border);
  gap: 16px;
  flex-shrink: 0;
}

.toolbar-left,
.toolbar-right {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

.toolbar-center {
  flex: 1;
  overflow: hidden;
}

.path-breadcrumb {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  overflow-x: auto;
  white-space: nowrap;
}

.path-segment {
  display: flex;
  align-items: center;
  gap: 4px;
  color: var(--text-secondary);
  cursor: pointer;
  padding: 2px 6px;
  border-radius: 4px;
  transition: var(--transition);
}

.path-segment:hover {
  background: var(--bg-hover);
  color: var(--accent);
}

.fm-search {
  padding: 8px 16px;
  background: var(--bg-secondary);
  border-bottom: 1px solid var(--border);
  flex-shrink: 0;
}

.fm-content {
  flex: 1;
  overflow-y: auto;
  padding: 0;
}

.file-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.file-table th {
  text-align: left;
  padding: 10px 12px;
  background: var(--bg-secondary);
  color: var(--text-muted);
  font-weight: 500;
  cursor: pointer;
  user-select: none;
  position: sticky;
  top: 0;
  z-index: 1;
  border-bottom: 1px solid var(--border);
}

.file-table th:hover {
  color: var(--text-primary);
}

.col-icon { width: 40px; }
.col-name { min-width: 200px; }
.col-size { width: 100px; }
.col-mode { width: 120px; }
.col-time { width: 180px; }
.col-actions { width: 60px; }

.file-row {
  transition: background 0.15s;
  border-bottom: 1px solid var(--border);
  cursor: pointer;
}

.file-row:hover {
  background: var(--bg-hover);
}

.file-row.selected {
  background: var(--accent-dim);
}

.file-row td {
  padding: 8px 12px;
  color: var(--text-primary);
}

.file-name.is-dir {
  color: #FFA500;
  font-weight: 500;
}

.empty-row {
  text-align: center;
}

.empty-files {
  padding: 60px 0;
  color: var(--text-muted);
}

.empty-files p {
  margin-top: 12px;
}

.fm-statusbar {
  height: 28px;
  background: var(--bg-secondary);
  border-top: 1px solid var(--border);
  display: flex;
  align-items: center;
  padding: 0 16px;
  gap: 16px;
  font-size: 12px;
  color: var(--text-muted);
  flex-shrink: 0;
}

.editor-container {
  max-height: 70vh;
  overflow: auto;
}

.code-editor :deep(.el-textarea__inner) {
  font-family: 'Courier New', monospace;
  font-size: 14px;
  line-height: 1.5;
  background: var(--bg-primary);
  color: var(--text-primary);
  min-height: 500px !important;
}

.search-results {
  max-height: 400px;
  overflow-y: auto;
}

.search-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  color: var(--text-secondary);
  transition: var(--transition);
}

.search-item:hover {
  background: var(--bg-hover);
  color: var(--accent);
}

.empty-search {
  text-align: center;
  padding: 40px;
  color: var(--text-muted);
}
</style>
