<template>
  <div class="direct-page">
    <div class="connecting-overlay" v-if="status === 'connecting'">
      <div class="connecting-content">
        <div class="spinner"></div>
        <h3>正在连接到 {{ sessionInfo?.host || '...' }}</h3>
        <p>{{ connectStatus }}</p>
        <div class="connect-steps">
          <div class="step" :class="{ active: connectStep >= 1, done: connectStep > 1 }"><div class="step-dot"></div><span>解析主机地址</span></div>
          <div class="step" :class="{ active: connectStep >= 2, done: connectStep > 2 }"><div class="step-dot"></div><span>建立SSH连接</span></div>
          <div class="step" :class="{ active: connectStep >= 3, done: connectStep > 3 }"><div class="step-dot"></div><span>身份认证</span></div>
          <div class="step" :class="{ active: connectStep >= 4, done: connectStep > 4 }"><div class="step-dot"></div><span>启动终端</span></div>
        </div>
      </div>
    </div>

    <div class="error-overlay" v-if="status === 'error'">
      <div class="error-content">
        <el-icon :size="48" color="var(--danger)"><CircleCloseFilled /></el-icon>
        <h3>连接失败</h3>
        <p>{{ error }}</p>
        <div class="error-actions">
          <el-button type="primary" @click="retry">重新连接</el-button>
        </div>
      </div>
    </div>

    <div class="toolbar" v-if="status === 'connected'">
      <div class="toolbar-left">
        <span class="server-info">
          <el-icon><Monitor /></el-icon>
          {{ sessionInfo?.host }}:{{ sessionInfo?.port || 22 }}
        </span>
      </div>
      <div class="toolbar-center">
        <el-button-group size="small">
          <el-tooltip content="列出文件 (ls -la)" placement="bottom">
            <el-button @click="sendCommand('ls -la')">ls</el-button>
          </el-tooltip>
          <el-tooltip content="查看当前目录 (pwd)" placement="bottom">
            <el-button @click="sendCommand('pwd')">pwd</el-button>
          </el-tooltip>
          <el-tooltip content="系统信息" placement="bottom">
            <el-button @click="sendCommand('uname -a && cat /etc/os-release && free -h && df -h')"><el-icon><Cpu /></el-icon></el-button>
          </el-tooltip>
          <el-divider direction="vertical" />
          <el-dropdown trigger="click" @command="handleQuickCommand">
            <el-button size="small">快捷命令<el-icon class="el-icon--right"><ArrowDown /></el-icon></el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item
                  v-for="(cmd, index) in quickCommandsList"
                  :key="index"
                  :command="cmd.command"
                >
                  {{ cmd.name }}
                </el-dropdown-item>
                <el-dropdown-item v-if="quickCommandsList.length === 0" disabled>暂无快捷命令</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </el-button-group>
      </div>
      <div class="toolbar-right">
        <el-button size="small" @click="showHardwareMonitor = !showHardwareMonitor" :icon="DataAnalysis" :type="showHardwareMonitor ? 'primary' : ''">监控</el-button>
        <el-button size="small" @click="toggleSFTP" :icon="FolderOpened" :type="sftpOpen ? 'primary' : ''">
          {{ sftpOpen ? '关闭文件' : '文件管理' }}
        </el-button>
        <el-button size="small" @click="showSettings = true" :icon="Setting">设置</el-button>
      </div>
    </div>

    <div class="main-layout" :class="{ 'sftp-open': sftpOpen, 'settings-open': showSettings }">
      <div class="terminal-area">
        <div class="terminal-container" ref="terminalContainer" :style="bgStyle">
          <div ref="terminalElement" class="terminal-element"></div>
        </div>
      </div>

      <transition name="slide-sftp">
        <div class="sftp-panel" v-show="sftpOpen && status === 'connected'">
          <div class="sftp-header">
            <span class="sftp-title">文件管理</span>
            <el-button text size="small" :icon="Close" @click="sftpOpen = false" class="sftp-close-btn" />
          </div>

          <div class="sftp-toolbar">
            <el-breadcrumb separator="/">
              <el-breadcrumb-item @click="navigateToPath('/')"><el-icon><HomeFilled /></el-icon></el-breadcrumb-item>
              <template v-for="(seg, i) in pathSegments" :key="i">
                <el-breadcrumb-item @click="navigateToPath(buildPath(i + 1))">{{ seg }}</el-breadcrumb-item>
              </template>
            </el-breadcrumb>
            <div class="toolbar-actions">
              <el-button text size="small" :icon="RefreshRight" @click="refreshList" :loading="sftpLoading" />
              <el-dropdown trigger="click" @command="handleSFTPAction">
                <el-button text size="small" :icon="Plus" />
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="mkdir" :icon="FolderAdd">新建目录</el-dropdown-item>
                    <el-dropdown-item command="upload" :icon="Upload">上传文件</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </div>

          <div class="sftp-file-list" v-loading="sftpLoading" element-loading-text="加载中...">
            <div
              v-for="file in sftpFiles"
              :key="file.path"
              class="file-item"
              :class="{ selected: selectedFile?.path === file.path }"
              @click="selectedFile = file"
              @dblclick="handleFileDblClick(file)"
            >
              <el-icon :size="18" :color="getFileIconColor(file)">
                <component :is="file.is_dir ? Folder : Document" />
              </el-icon>
              <div class="file-info">
                <span class="file-name">{{ file.name }}</span>
                <span class="file-meta">{{ file.is_dir ? '目录' : formatSize(file.size) }}</span>
              </div>
              <el-dropdown trigger="click" @command="(cmd) => handleFileAction(cmd, file)">
                <el-button text size="small" :icon="MoreFilled" @click.stop />
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item v-if="file.is_dir" command="open" :icon="FolderOpened">打开</el-dropdown-item>
                    <el-dropdown-item v-if="!file.is_dir" command="download" :icon="Download">下载</el-dropdown-item>
                    <el-dropdown-item v-if="!file.is_dir" command="edit" :icon="Edit">编辑</el-dropdown-item>
                    <el-dropdown-item command="rename" :icon="EditPen">重命名</el-dropdown-item>
                    <el-dropdown-item command="delete" :icon="Delete" style="color: var(--danger)">删除</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
            <div class="empty-files" v-if="!sftpLoading && sftpFiles.length === 0">
              <el-icon :size="32" color="var(--text-muted)"><FolderOpened /></el-icon>
              <p>空目录</p>
            </div>
          </div>
        </div>
      </transition>
    </div>

    <button class="mobile-sftp-toggle" v-if="status === 'connected' && !sftpOpen && !showSettings && !keyboardVisible" @click="sftpOpen = true">
      <el-icon :size="18"><FolderOpened /></el-icon>
      文件
    </button>

    <button class="mobile-quick-cmd-toggle" v-if="status === 'connected' && !keyboardVisible" @click="showMobileQuickCmd = !showMobileQuickCmd">
      <el-icon :size="18"><Cpu /></el-icon>
      快捷
    </button>

    <div class="mobile-quick-cmd-panel" v-show="showMobileQuickCmd">
      <div class="mobile-quick-cmd-header">快捷命令</div>
      <div class="mobile-quick-cmd-list">
        <button @click="sendCommand('ls -la')">ls</button>
        <button @click="sendCommand('pwd')">pwd</button>
        <button @click="sendCommand('uname -a && cat /etc/os-release && free -h && df -h')"><el-icon><Cpu /></el-icon></button>
      </div>
      <div class="mobile-quick-cmd-divider"></div>
      <button
        v-for="(cmd, index) in quickCommandsList"
        :key="index"
        class="mobile-quick-item"
        @click="handleQuickCommand(cmd.command); showMobileQuickCmd = false"
      >
        {{ cmd.name }}
      </button>
      <button v-if="quickCommandsList.length === 0" class="mobile-quick-item" disabled>暂无快捷命令</button>
    </div>

    <div class="mobile-extra-keys" v-if="status === 'connected'">
      <button class="extra-key" :class="{ active: ctrlPressed }" @click="toggleCtrl">CTRL</button>
      <button class="extra-key" :class="{ active: altPressed }" @click="toggleAlt">ALT</button>
      <button class="extra-key" @click="sendKey('ESC')">ESC</button>
      <button class="extra-key" @click="sendKey('TAB')">TAB</button>
      <button class="extra-key special" @click="sendCtrlC">CTRL+C</button>
      <button class="extra-key" @click="sendKey('UP')">↑</button>
      <button class="extra-key" @click="sendKey('DOWN')">↓</button>
      <button class="extra-key" @click="sendKey('LEFT')">←</button>
      <button class="extra-key" @click="sendKey('RIGHT')">→</button>
    </div>

    <el-drawer v-model="showSettings" title="终端设置" direction="rtl" size="320px">
      <div class="settings-content">
        <div class="settings-section">
          <h4>外观</h4>
          <div class="setting-item">
            <label>背景图片</label>
            <div class="bg-url-row">
              <el-input v-model="bgImageUrl" placeholder="输入图片URL" clearable @change="applyBgImage">
                <template #append>
                  <el-button @click="applyBgImage">应用</el-button>
                </template>
              </el-input>
            </div>
            <div class="bg-quick-actions">
              <el-button size="small" type="primary" plain @click="useRandomBg">随机动漫图</el-button>
              <el-button size="small" v-if="bgImageUrl" :icon="RefreshRight" @click="refreshRandomBg" title="刷新随机图片">刷新图片</el-button>
              <el-button size="small" v-if="bgImageUrl" type="danger" plain @click="clearBgImage">清除背景</el-button>
            </div>
          </div>
          <div class="setting-item">
            <label>背景透明度</label>
            <el-slider v-model="bgOpacity" :min="0" :max="100" :step="5" @change="applyBgImage" />
            <span class="value-label">{{ bgOpacity }}%</span>
          </div>
          <div class="setting-item">
            <label>模糊效果</label>
            <el-switch v-model="bgBlur" @change="applyBgImage" />
            <span class="value-label">{{ bgBlur ? '开启' : '关闭' }}</span>
          </div>
        </div>

        <div class="settings-section">
          <h4>字体</h4>
          <div class="setting-item">
            <label>字体大小</label>
            <el-slider v-model="fontSize" :min="10" :max="24" :step="1" @change="applyFontSize" />
            <span class="value-label">{{ fontSize }}px</span>
          </div>
          <div class="setting-item">
            <label>字体</label>
            <el-select v-model="fontFamily" @change="applyFontFamily">
              <el-option value="'Cascadia Code', 'Fira Code', monospace" label="Cascadia Code" />
              <el-option value="'Fira Code', monospace" label="Fira Code" />
              <el-option value="'JetBrains Mono', monospace" label="JetBrains Mono" />
              <el-option value="'Consolas', monospace" label="Consolas" />
              <el-option value="monospace" label="系统默认" />
            </el-select>
          </div>
        </div>

        <div class="settings-section">
          <h4>光标</h4>
          <div class="setting-item">
            <label>光标样式</label>
            <el-radio-group v-model="cursorStyle" @change="applyCursorStyle">
              <el-radio value="block">方块</el-radio>
              <el-radio value="underline">下划线</el-radio>
              <el-radio value="bar">竖线</el-radio>
            </el-radio-group>
          </div>
          <div class="setting-item">
            <label>光标闪烁</label>
            <el-switch v-model="cursorBlink" @change="applyCursorStyle" />
          </div>
        </div>
      </div>
    </el-drawer>

    <el-dialog v-model="showMkdir" title="新建目录" width="400px" destroy-on-close>
      <el-input v-model="newDirName" placeholder="目录名称" @keyup.enter="doMkdir" />
      <template #footer>
        <el-button @click="showMkdir = false">取消</el-button>
        <el-button type="primary" @click="doMkdir">创建</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showRename" title="重命名" width="400px" destroy-on-close>
      <el-input v-model="renameName" placeholder="新名称" @keyup.enter="doRename" />
      <template #footer>
        <el-button @click="showRename = false">取消</el-button>
        <el-button type="primary" @click="doRename">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showEditor" title="编辑文件" width="700px" destroy-on-close>
      <el-input v-model="editorContent" type="textarea" :rows="18" :placeholder="'正在编辑: ' + editingFileName" />
      <template #footer>
        <el-button @click="showEditor = false">取消</el-button>
        <el-button type="primary" @click="saveEditedFile" :loading="saving">保存</el-button>
      </template>
    </el-dialog>

    <input type="file" ref="fileInput" style="display:none" @change="handleFileUpload" />

    <HardwareMonitor
      v-if="status === 'connected' && showHardwareMonitor"
      :token="token"
      :visible="showHardwareMonitor"
      @close="showHardwareMonitor = false"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { WebLinksAddon } from '@xterm/addon-web-links'
import { SearchAddon } from '@xterm/addon-search'
import { ElMessage, ElMessageBox } from 'element-plus'
import '@xterm/xterm/css/xterm.css'
import {
  FolderOpened, Setting, Monitor, Cpu, ArrowDown, Close,
  HomeFilled, RefreshRight, Plus, FolderAdd, Upload, Download,
  Edit, EditPen, Delete, MoreFilled, Folder, Document, DataAnalysis
} from '@element-plus/icons-vue'
import HardwareMonitor from '../components/HardwareMonitor.vue'

const route = useRoute()
const router = useRouter()
const token = route.params.token

const status = ref('loading')
const error = ref('')
const connectStatus = ref('')
const connectStep = ref(0)
const sessionInfo = ref(null)
const terminalContainer = ref(null)
const terminalElement = ref(null)
const showSettings = ref(false)
const showHardwareMonitor = ref(false)
const showMobileQuickCmd = ref(false)
const ctrlPressed = ref(false)
const altPressed = ref(false)
const keyboardVisible = ref(false)
const bgImageUrl = ref('')
const bgOpacity = ref(30)
const bgBlur = ref(false)
const fontSize = ref(14)
const fontFamily = ref("'Cascadia Code', 'Fira Code', monospace")
const cursorStyle = ref('bar')
const cursorBlink = ref(true)

const sftpOpen = ref(false)
const sftpFiles = ref([])
const sftpLoading = ref(false)
const selectedFile = ref(null)
const currentPath = ref('/')
const showMkdir = ref(false)
const newDirName = ref('')
const showRename = ref(false)
const renameName = ref('')
const showEditor = ref(false)
const editorContent = ref('')
const editingFileName = ref('')
const editingFilePath = ref('')
const saving = ref(false)
const fileInput = ref(null)

let term = null, fitAddon = null, ws = null, pingInterval = null, sftpWs = null
let pendingCallbacks = []

const bgStyle = computed(() => {
  if (!bgImageUrl.value) return {}
  return {
    backgroundImage: `url(${bgImageUrl.value})`,
    backgroundSize: 'cover',
    backgroundPosition: 'center',
    backgroundRepeat: 'no-repeat'
  }
})

const pathSegments = computed(() => {
  const p = currentPath.value
  if (p === '/') return []
  return p.split('/').filter(Boolean)
})

const buildPath = (idx) => '/' + pathSegments.value.slice(0, idx).join('/')

const defaultTheme = {
  background: '#0f0f1a',
  foreground: '#e0e0e0',
  cursor: '#00d4ff',
  cursorAccent: '#0f0f1a',
  selectionBackground: 'rgba(0,212,255,0.3)',
  black: '#0f0f1a',
  red: '#ff4757',
  green: '#2ed573',
  yellow: '#ffa502',
  blue: '#3B82F6',
  magenta: '#a855f7',
  cyan: '#00d4ff',
  white: '#e0e0e0',
  brightBlack: '#5a6480',
  brightRed: '#ff6b81',
  brightGreen: '#7bed9f',
  brightYellow: '#ffc048',
  brightBlue: '#60a5fa',
  brightMagenta: '#c084fc',
  brightCyan: '#22d3ee',
  brightWhite: '#ffffff'
}

const getWsUrl = (sessionToken) => {
  const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  return `${proto}//${window.location.host}/api/direct/ws/terminal?token=${sessionToken}&cols=${term?.cols || 80}&rows=${term?.rows || 24}`
}

const getSftpWsUrl = () => {
  const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  return `${proto}//${window.location.host}/api/direct/ws/sftp?token=${token}`
}

const initTerminal = () => {
  term = new Terminal({
    cursorBlink: cursorBlink.value,
    cursorStyle: cursorStyle.value,
    fontSize: fontSize.value,
    fontFamily: fontFamily.value,
    theme: defaultTheme,
    scrollback: 10000,
    convertEol: true
  })
  fitAddon = new FitAddon()
  term.loadAddon(fitAddon)
  term.loadAddon(new WebLinksAddon())
  term.loadAddon(new SearchAddon())
  term.open(terminalElement.value)
  nextTick(() => { fitAddon.fit(); applyBgImage() })
}

const applyBgImage = () => {
  if (!terminalContainer.value) return
  const xtermEl = terminalElement.value?.querySelector('.xterm')
  if (bgImageUrl.value) {
    terminalContainer.value.style.backgroundImage = `url(${bgImageUrl.value})`
    terminalContainer.value.style.backgroundSize = 'cover'
    terminalContainer.value.style.backgroundPosition = 'center'
    terminalContainer.value.style.backgroundRepeat = 'no-repeat'
    terminalContainer.value.style.setProperty('--bg-overlay-opacity', bgOpacity.value / 100)
    terminalContainer.value.classList.add('has-bg')
    if (bgBlur.value) {
      terminalContainer.value.style.backdropFilter = 'blur(4px)'
    } else {
      terminalContainer.value.style.backdropFilter = ''
    }
    if (term) {
      term.options.theme.background = 'transparent'
    }
    if (xtermEl) {
      xtermEl.style.background = 'transparent'
    }
    const viewport = terminalElement.value?.querySelector('.xterm-viewport')
    if (viewport) {
      viewport.style.background = 'transparent'
    }
  } else {
    terminalContainer.value.style.backgroundImage = ''
    terminalContainer.value.style.backdropFilter = ''
    terminalContainer.value.classList.remove('has-bg')
    if (term) {
      term.options.theme.background = defaultTheme.background
    }
    if (xtermEl) {
      xtermEl.style.background = ''
    }
    const viewport = terminalElement.value?.querySelector('.xterm-viewport')
    if (viewport) {
      viewport.style.background = ''
    }
  }
}

const clearBgImage = () => {
  bgImageUrl.value = ''
  applyBgImage()
}

const RANDOM_BG_API = 'https://www.loliapi.com/acg/'

const useRandomBg = () => {
  bgImageUrl.value = RANDOM_BG_API + '?t=' + Date.now()
  applyBgImage()
}

const refreshRandomBg = () => {
  if (bgImageUrl.value && bgImageUrl.value.includes('loliapi.com')) {
    bgImageUrl.value = RANDOM_BG_API + '?t=' + Date.now()
    applyBgImage()
  } else {
    useRandomBg()
  }
}

const toggleCtrl = () => {
  ctrlPressed.value = !ctrlPressed.value
  if (ctrlPressed.value) altPressed.value = false
}

const toggleAlt = () => {
  altPressed.value = !altPressed.value
  if (altPressed.value) ctrlPressed.value = false
}

const sendKey = (key) => {
  if (!term || !ws || ws.readyState !== WebSocket.OPEN) return

  let data = ''
  const ctrl = ctrlPressed.value
  const alt = altPressed.value

  switch (key) {
    case 'ESC': data = '\x1b'; break
    case 'TAB': data = '\t'; break
    case 'UP': data = '\x1b[A'; break
    case 'DOWN': data = '\x1b[B'; break
    case 'LEFT': data = '\x1b[D'; break
    case 'RIGHT': data = '\x1b[C'; break
  }

  if (ctrl && data) {
    data = '\x1b' + data
  }
  if (alt && data) {
    data = '\x1b' + data
  }

  ws.send(JSON.stringify({ type: 'input', data }))
  ctrlPressed.value = false
  altPressed.value = false
}

const sendCtrlC = () => {
  if (ws && ws.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify({ type: 'input', data: '\x03' }))
  }
  ctrlPressed.value = false
  altPressed.value = false
}

const applyFontSize = () => {
  if (term) term.options.fontSize = fontSize.value
  if (fitAddon) try { fitAddon.fit() } catch {}
}

const applyFontFamily = () => {
  if (term) term.options.fontFamily = fontFamily.value
}

const applyCursorStyle = () => {
  if (term) {
    term.options.cursorStyle = cursorStyle.value
    term.options.cursorBlink = cursorBlink.value
  }
}

const sendCommand = (cmd) => {
  if (term && ws?.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify({ type: 'input', data: cmd + '\r' }))
  }
}

const quickCommandsList = ref([])

const loadQuickCommands = async () => {
  try {
    const res = await fetch('/api/quick-commands')
    const data = await res.json()
    if (data.success && data.commands) {
      quickCommandsList.value = data.commands
    }
  } catch (err) {
    console.error('加载快捷命令失败', err)
  }
}

const handleQuickCommand = (command) => {
  if (command) {
    sendCommand(command)
  }
}

const toggleSFTP = () => {
  sftpOpen.value = !sftpOpen.value
  if (sftpOpen.value && !sftpWs) {
    initSFTP()
  }
  nextTick(() => { if (fitAddon) try { fitAddon.fit() } catch {} })
}

const initSFTP = () => {
  closeSFTP()
  sftpWs = new WebSocket(getSftpWsUrl())
  sftpWs.onmessage = (event) => {
    try {
      const msg = JSON.parse(event.data)
      if (msg.type === 'connected') {
        listDirectory(currentPath.value)
        return
      }
      if (msg.type === 'error') {
        ElMessage.error(msg.data || 'SFTP错误')
        return
      }

      const cb = pendingCallbacks.shift()
      if (msg.files) sftpFiles.value = msg.files
      if (cb) {
        msg.success ? cb.resolve(msg) : cb.reject(new Error(msg.message || '操作失败'))
      } else if (msg.message && !msg.files) {
        msg.success ? ElMessage.success(msg.message) : ElMessage.error(msg.message)
      }
    } catch {}
  }
  sftpWs.onerror = () => ElMessage.error('SFTP连接失败')
  sftpWs.onclose = () => {}
}

const closeSFTP = () => {
  if (sftpWs) { sftpWs.close(); sftpWs = null; pendingCallbacks = [] }
}

const sendSFTPRequest = (req) => new Promise((resolve, reject) => {
  pendingCallbacks.push({ resolve, reject })
  if (sftpWs?.readyState === WebSocket.OPEN) sftpWs.send(JSON.stringify(req))
  else { pendingCallbacks.pop(); reject(new Error('SFTP未连接')) }
})

const listDirectory = async (path) => {
  sftpLoading.value = true
  try { await sendSFTPRequest({ action: 'list', path }) }
  catch (e) { ElMessage.error('获取目录列表失败') }
  finally { setTimeout(() => { sftpLoading.value = false }, 200) }
}

const navigateToPath = (path) => { currentPath.value = path; listDirectory(path); selectedFile.value = null }
const refreshList = () => listDirectory(currentPath.value)

const handleFileDblClick = (file) => {
  if (file.is_dir) navigateToPath(file.path)
}

const handleSFTPAction = (command) => {
  if (command === 'mkdir') { newDirName.value = ''; showMkdir.value = true }
  else if (command === 'upload') { fileInput.value?.click() }
}

const doMkdir = async () => {
  if (!newDirName.value) return
  try {
    await sendSFTPRequest({ action: 'mkdir', path: currentPath.value === '/' ? `/${newDirName.value}` : `${currentPath.value}/${newDirName.value}` })
    showMkdir.value = false
    refreshList()
  } catch (e) { ElMessage.error(e.message || '创建目录失败') }
}

const handleFileAction = async (command, file) => {
  switch (command) {
    case 'open':
      if (file.is_dir) navigateToPath(file.path)
      break
    case 'download':
      try {
        const res = await sendSFTPRequest({ action: 'download', path: file.path })
        if (res.data) {
          const byteChars = atob(res.data)
          const bytes = new Uint8Array(byteChars.length)
          for (let i = 0; i < byteChars.length; i++) bytes[i] = byteChars.charCodeAt(i)
          const blob = new Blob([bytes])
          const url = URL.createObjectURL(blob)
          const a = document.createElement('a')
          a.href = url; a.download = file.name; a.click()
          URL.revokeObjectURL(url)
        }
      } catch (e) { ElMessage.error(e.message || '下载失败') }
      break
    case 'edit':
      try {
        const res = await sendSFTPRequest({ action: 'read', path: file.path })
        if (res.data) {
          editingFileName.value = file.name
          editingFilePath.value = file.path
          editorContent.value = atob(res.data)
          showEditor.value = true
        }
      } catch (e) { ElMessage.error(e.message || '读取文件失败') }
      break
    case 'rename':
      renameName.value = file.name
      showRename.value = true
      selectedFile.value = file
      break
    case 'delete':
      try {
        await ElMessageBox.confirm(`确定删除 ${file.name}?`, '确认删除', { type: 'warning' })
        await sendSFTPRequest({ action: 'rm', path: file.path })
        refreshList()
      } catch {}
      break
  }
}

const doRename = async () => {
  if (!renameName.value || !selectedFile.value) return
  try {
    const dir = selectedFile.value.path.substring(0, selectedFile.value.path.lastIndexOf('/'))
    const newPath = dir ? `${dir}/${renameName.value}` : `/${renameName.value}`
    await sendSFTPRequest({ action: 'rename', path: selectedFile.value.path, dst: newPath })
    showRename.value = false
    refreshList()
  } catch (e) { ElMessage.error(e.message || '重命名失败') }
}

const saveEditedFile = async () => {
  saving.value = true
  try {
    const content = btoa(unescape(encodeURIComponent(editorContent.value)))
    await sendSFTPRequest({ action: 'upload', path: editingFilePath.value, content })
    showEditor.value = false
    ElMessage.success('保存成功')
  } catch (e) { ElMessage.error(e.message || '保存失败') }
  finally { saving.value = false }
}

const handleFileUpload = async (e) => {
  const file = e.target.files[0]
  if (!file) return
  try {
    const reader = new FileReader()
    reader.onload = async (ev) => {
      const base64 = ev.target.result.split(',')[1]
      const path = currentPath.value === '/' ? `/${file.name}` : `${currentPath.value}/${file.name}`
      try {
        await sendSFTPRequest({ action: 'upload', path, content: base64 })
        refreshList()
        ElMessage.success('上传成功')
      } catch (err) { ElMessage.error(err.message || '上传失败') }
    }
    reader.readAsDataURL(file)
  } finally { e.target.value = '' }
}

const formatSize = (size) => {
  if (!size || size <= 0) return '-'
  const units = ['B', 'KB', 'MB', 'GB']
  let i = 0; let s = size
  while (s >= 1024 && i < units.length - 1) { s /= 1024; i++ }
  return i === 0 ? `${s} ${units[i]}` : `${s.toFixed(1)} ${units[i]}`
}

const getFileIconColor = (file) => {
  if (file.is_dir) return '#F59E0B'
  const ext = file.name.split('.').pop().toLowerCase()
  const colors = { js: '#F7DF1E', ts: '#3178C6', py: '#3776AB', go: '#00ADD8', rs: '#DEA584', sh: '#4EAA25', json: '#000', md: '#083FA1', html: '#E34F26', css: '#1572B6', yml: '#CB171E', yaml: '#CB171E', sql: '#336791', java: '#ED8B00', php: '#777BB4', rb: '#CC342D', vue: '#42B883', c: '#555555', cpp: '#00599C' }
  return colors[ext] || '#94A3B8'
}

const connectWebSocket = (sessionToken) => {
  if (ws) { ws.close(); ws = null }

  const steps = [
    { step: 1, msg: '正在解析主机地址...', delay: 400 },
    { step: 2, msg: '正在建立SSH连接...', delay: 800 },
    { step: 3, msg: '正在进行身份认证...', delay: 600 },
    { step: 4, msg: '正在启动终端会话...', delay: 400 }
  ]
  steps.forEach(({ step, msg, delay }) => setTimeout(() => {
    if (status.value === 'connecting') { connectStep.value = step; connectStatus.value = msg }
  }, delay))

  ws = new WebSocket(getWsUrl(sessionToken))

  ws.onmessage = (event) => {
    try {
      const msg = JSON.parse(event.data)
      switch (msg.type) {
        case 'connected':
          status.value = 'connected'
          connectStep.value = 4
          startPing()
          break
        case 'output':
          if (term) term.write(msg.data)
          break
        case 'error':
          status.value = 'error'
          error.value = msg.data
          break
        case 'disconnected':
          status.value = 'error'
          error.value = '连接已断开'
          break
        case 'pong':
          break
      }
    } catch {
      if (term) term.write(event.data)
    }
  }

  ws.onerror = () => {
    status.value = 'error'
    error.value = 'WebSocket连接失败'
  }

  ws.onclose = () => {
    if (status.value === 'connected') {
      status.value = 'error'
      error.value = '连接已关闭'
    }
    stopPing()
  }

  term.onData((data) => {
    if (ws?.readyState === WebSocket.OPEN) ws.send(JSON.stringify({ type: 'input', data }))
  })
  term.onResize(({ cols, rows }) => {
    if (ws?.readyState === WebSocket.OPEN) ws.send(JSON.stringify({ type: 'resize', cols, rows }))
  })
}

const startPing = () => {
  stopPing()
  pingInterval = setInterval(() => {
    if (ws?.readyState === WebSocket.OPEN) ws.send(JSON.stringify({ type: 'ping' }))
  }, 30000)
}

const stopPing = () => {
  if (pingInterval) { clearInterval(pingInterval); pingInterval = null }
}

const handleResize = () => { if (fitAddon) try { fitAddon.fit() } catch {} }

const fetchSession = async () => {
  try {
    const res = await fetch(`/api/direct/session/${token}`)
    const data = await res.json()
    if (data.success) return data
    else throw new Error(data.message || '获取会话失败')
  } catch (e) { throw e }
}

const retry = () => {
  error.value = ''
  if (sessionInfo.value) {
    status.value = 'connecting'
    connectWebSocket(token)
  }
}

const loadAndConnect = async () => {
  if (!token) {
    status.value = 'error'
    error.value = '缺少token参数'
    return
  }

  status.value = 'connecting'
  connectStatus.value = '正在获取会话信息...'

  try {
    const data = await fetchSession()
    sessionInfo.value = {
      host: data.host,
      port: data.port,
      username: data.username,
      password: data.password
    }
    connectWebSocket(token)
  } catch (e) {
    status.value = 'error'
    error.value = e.message || '获取会话失败，请检查链接是否过期'
  }
}

onMounted(async () => {
  initTerminal()
  await loadQuickCommands()
  await loadAndConnect()
  window.addEventListener('resize', handleResize)
  if (window.visualViewport) {
    window.visualViewport.addEventListener('resize', handleViewportResize)
  }
})

onBeforeUnmount(() => {
  stopPing()
  closeSFTP()
  if (ws) ws.close()
  if (term) term.dispose()
  window.removeEventListener('resize', handleResize)
  if (window.visualViewport) {
    window.visualViewport.removeEventListener('resize', handleViewportResize)
  }
})

const handleViewportResize = () => {
  if (window.visualViewport) {
    keyboardVisible.value = window.visualViewport.height < window.innerHeight * 0.8
  }
}
</script>

<style scoped>
.direct-page {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: #0f0f1a;
  position: relative;
  overflow: hidden;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background: rgba(30, 32, 48, 0.8);
  backdrop-filter: blur(10px);
  border-bottom: 1px solid rgba(255,255,255,0.1);
  flex-shrink: 0;
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.toolbar-center {
  display: flex;
  align-items: center;
  gap: 8px;
}

.server-info {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--text-muted);
}

.toolbar-right {
  display: flex;
  gap: 8px;
}

.main-layout {
  display: flex;
  flex: 1;
  overflow: hidden;
  position: relative;
}

.terminal-area {
  flex: 1;
  min-width: 0;
  position: relative;
  overflow: hidden;
}

.terminal-container {
  width: 100%;
  height: 100%;
  padding: 6px;
  overflow: hidden;
  position: relative;
}

.terminal-container::before {
  content: '';
  position: absolute;
  inset: 0;
  background: rgba(15, 15, 26, var(--bg-overlay-opacity, 0));
  pointer-events: none;
  z-index: 0;
  opacity: 0;
  transition: opacity 0.3s ease;
}
.terminal-container.has-bg::before {
  opacity: 1;
}

.terminal-container:has(> div) .terminal-element {
  position: relative;
  z-index: 1;
}

.terminal-element {
  width: 100%;
  height: 100%;
  overflow: hidden;
}

.sftp-panel {
  width: 340px;
  min-width: 280px;
  max-width: 45vw;
  background: var(--bg-card);
  border-left: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.sftp-header {
  height: 42px;
  padding: 0 12px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid var(--border);
  flex-shrink: 0;
}

.sftp-title {
  font-weight: 600;
  font-size: 14px;
  color: var(--text-primary);
}

.sftp-close-btn { color: var(--text-muted); }

.sftp-toolbar {
  height: 40px;
  padding: 0 10px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid var(--border);
  flex-shrink: 0;
  gap: 8px;
}

.sftp-toolbar :deep(.el-breadcrumb) { flex: 1; min-width: 0; }
.sftp-toolbar :deep(.el-breadcrumb__inner) { cursor: pointer; }
.toolbar-actions { display: flex; gap: 2px; flex-shrink: 0; }

.sftp-file-list {
  flex: 1;
  overflow-y: auto;
  padding: 4px 0;
  -webkit-overflow-scrolling: touch;
}

.file-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 7px 12px;
  cursor: pointer;
  transition: background 0.15s;
}

.file-item:hover { background: var(--bg-hover); }
.file-item.selected { background: var(--accent-dim); }

.file-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 1px;
}

.file-name {
  font-size: 13px;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.file-meta { font-size: 11px; color: var(--text-muted); }

.empty-files {
  text-align: center;
  padding: 40px 16px;
  color: var(--text-muted);
}

.empty-files p { margin-top: 8px; font-size: 13px; }

.mobile-sftp-toggle {
  display: none;
  position: fixed;
  bottom: 20px;
  right: 16px;
  z-index: 100;
  background: var(--accent);
  color: #fff;
  border: none;
  border-radius: 20px;
  padding: 10px 18px;
  font-size: 13px;
  cursor: pointer;
  align-items: center;
  gap: 6px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.3);
}

.mobile-quick-cmd-toggle {
  display: none;
  position: fixed;
  bottom: 80px;
  right: 16px;
  z-index: 100;
  background: #3B82F6;
  color: #fff;
  border: none;
  border-radius: 20px;
  padding: 10px 18px;
  font-size: 13px;
  cursor: pointer;
  align-items: center;
  gap: 6px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.3);
}

.mobile-quick-cmd-panel {
  display: none;
  position: fixed;
  bottom: 130px;
  right: 16px;
  z-index: 200;
  background: #1e1e2f;
  border-radius: 12px;
  padding: 8px 0;
  min-width: 180px;
  max-height: 60vh;
  overflow-y: auto;
  box-shadow: 0 8px 32px rgba(0,0,0,0.5);
}

.mobile-quick-cmd-header {
  padding: 8px 16px;
  font-size: 12px;
  color: var(--text-muted);
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 1px;
}

.mobile-quick-cmd-list {
  display: flex;
  gap: 6px;
  padding: 4px 12px 8px;
}

.mobile-quick-cmd-list button {
  flex: 1;
  background: rgba(59,130,246,0.15);
  color: #60a5fa;
  border: 1px solid rgba(59,130,246,0.3);
  border-radius: 8px;
  padding: 8px 4px;
  font-size: 13px;
  cursor: pointer;
  text-align: center;
}

.mobile-quick-cmd-divider {
  height: 1px;
  background: rgba(255,255,255,0.08);
  margin: 4px 12px;
}

.mobile-quick-item {
  display: block;
  width: 100%;
  background: transparent;
  color: var(--text-primary);
  border: none;
  padding: 10px 16px;
  font-size: 14px;
  text-align: left;
  cursor: pointer;
  transition: background 0.15s;
}

.mobile-quick-item:hover { background: rgba(59,130,246,0.12); }
.mobile-quick-item:active { background: rgba(59,130,246,0.2); }

.mobile-extra-keys {
  display: none;
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  z-index: 400;
  background: #1a1a2e;
  border-top: 1px solid rgba(255,255,255,0.1);
  padding: 6px 4px;
  display: none;
  gap: 4px;
  flex-wrap: wrap;
  justify-content: center;
}

.extra-key {
  background: #2d2d44;
  color: #e0e0e0;
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 6px;
  padding: 8px 12px;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  min-width: 44px;
  text-align: center;
  transition: all 0.15s;
}

.extra-key:active {
  background: #3d3d5c;
  transform: scale(0.95);
}

.extra-key.active {
  background: #3B82F6;
  border-color: #3B82F6;
  color: #fff;
}

.extra-key.special {
  background: rgba(239,68,68,0.2);
  border-color: rgba(239,68,68,0.4);
  color: #f87171;
}

.extra-key.special:active {
  background: rgba(239,68,68,0.4);
}

.connecting-overlay, .error-overlay {
  position: absolute;
  inset: 0;
  z-index: 100;
  background: rgba(15,15,26,0.95);
  display: flex;
  align-items: center;
  justify-content: center;
  animation: fadeIn 0.3s ease;
}

.connecting-content, .error-content {
  text-align: center;
  max-width: 90%;
  padding: 20px;
}

.spinner {
  width: 44px;
  height: 44px;
  border: 3px solid var(--border);
  border-top-color: var(--accent);
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin: 0 auto 20px;
}

.connecting-content h3 { font-size: 17px; color: var(--text-primary); margin-bottom: 6px; }
.connecting-content p { font-size: 13px; color: var(--text-muted); margin-bottom: 28px; }

.connect-steps {
  display: flex;
  flex-direction: column;
  gap: 11px;
  align-items: flex-start;
  max-width: 190px;
  margin: 0 auto;
}

.step {
  display: flex;
  align-items: center;
  gap: 9px;
  font-size: 13px;
  color: var(--text-muted);
  transition: var(--transition);
}

.step.active { color: var(--text-primary); }
.step.done { color: var(--success); }

.step-dot {
  width: 9px;
  height: 9px;
  border-radius: 50%;
  background: var(--border);
  transition: var(--transition);
  flex-shrink: 0;
}

.step.active .step-dot { background: var(--accent); animation: pulse 1s infinite; }
.step.done .step-dot { background: var(--success); }

.error-content h3 { font-size: 19px; color: var(--text-primary); margin: 14px 0 7px; }
.error-content p { font-size: 13px; color: var(--text-muted); margin-bottom: 22px; max-width: 360px; word-break: break-word; }
.error-actions { display: flex; gap: 12px; justify-content: center; }

.settings-content { padding: 0 16px; }
.settings-section { margin-bottom: 24px; }
.settings-section h4 { font-size: 14px; color: var(--text-primary); margin-bottom: 12px; padding-bottom: 8px; border-bottom: 1px solid var(--border); }
.setting-item { margin-bottom: 16px; }
.setting-item label { display: block; font-size: 13px; color: var(--text-muted); margin-bottom: 6px; }
.setting-tip { font-size: 11px; color: var(--text-muted); margin-top: 4px; }
.value-label { font-size: 12px; color: var(--accent); margin-left: 8px; }

.bg-url-row { margin-bottom: 8px; }
.bg-quick-actions { display: flex; gap: 6px; flex-wrap: wrap; }

.slide-sftp-enter-active { transition: transform 0.25s ease, opacity 0.25s ease; }
.slide-sftp-leave-active { transition: transform 0.2s ease, opacity 0.2s ease; }
.slide-sftp-enter-from { transform: translateX(30px); opacity: 0; }
.slide-sftp-leave-to { transform: translateX(30px); opacity: 0; }

@media (max-width: 768px) {
  .toolbar-center { display: none; }
  .toolbar-right :deep(.el-button) { padding: 6px !important; font-size: 12px; }

  .sftp-panel {
    position: absolute;
    right: 0;
    top: 0;
    bottom: 0;
    z-index: 150;
    width: 88vw;
    max-width: 400px;
    box-shadow: -4px 0 24px rgba(0,0,0,0.5);
    border-left: none;
  }

  .mobile-sftp-toggle { display: flex; }
  .mobile-quick-cmd-toggle { display: flex; }
  .mobile-quick-cmd-panel { display: block; }
  .mobile-extra-keys { display: flex; }
  .terminal-area { padding-bottom: 50px; }
  .main-layout.sftp-open .terminal-area { filter: brightness(0.55); pointer-events: none; }
}

@media (max-width: 480px) {
  .toolbar { padding: 6px 8px; }
  .server-info { font-size: 12px; }
  .sftp-header { height: 38px; padding: 0 10px; }
  .sftp-toolbar { height: 36px; padding: 0 8px; }
  .file-item { padding: 8px 10px; gap: 6px; }
  .file-name { font-size: 12px; }
  .mobile-sftp-toggle { bottom: 50px; right: 10px; padding: 8px 14px; font-size: 12px; }
}

@keyframes spin { to { transform: rotate(360deg); } }
@keyframes pulse { 0%, 100% { opacity: 1; } 50% { opacity: 0.4; } }
@keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
</style>
