<template>
  <div class="terminal-page">

    <div class="terminal-topbar">
      <div class="topbar-left">
        <el-button text @click="goBack" :icon="ArrowLeft">返回</el-button>
        <div class="connection-info" v-if="serverInfo">
          <div class="status-dot" :class="connected ? 'connected' : 'disconnected'"></div>
          <span class="server-label">{{ serverInfo.name }}</span>
          <span class="server-addr">{{ serverInfo.username }}@{{ serverInfo.host }}:{{ serverInfo.port }}</span>
        </div>
      </div>
      <div class="topbar-center">
        <el-button-group size="small" v-if="connected">
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
      <div class="topbar-right">
        <el-button text @click="showHardwareMonitor = !showHardwareMonitor" :icon="DataAnalysis" :type="showHardwareMonitor ? 'primary' : ''">
          监控
        </el-button>
        <el-button text @click="toggleSFTP" :icon="FolderOpened" :type="sftpOpen ? 'primary' : ''">
          {{ sftpOpen ? '关闭文件' : '文件管理' }}
        </el-button>
        <el-button text @click="showSettingsPanel = true" :icon="Setting">设置</el-button>
        <el-button text @click="reconnect" :icon="Refresh" :loading="connecting">重连</el-button>
      </div>
    </div>

    <div class="main-layout" :class="{ 'sftp-open': sftpOpen, 'settings-open': showSettingsPanel }">

      <div class="terminal-area">

        <div class="connecting-overlay" v-if="connecting">
          <div class="connecting-content">
            <div class="spinner"></div>
            <h3>正在连接到 {{ serverInfo?.name || '...' }}</h3>
            <p>{{ connectStatus }}</p>
            <div class="connect-steps">
              <div class="step" :class="{ active: connectStep >= 1, done: connectStep > 1 }"><div class="step-dot"></div><span>解析主机地址</span></div>
              <div class="step" :class="{ active: connectStep >= 2, done: connectStep > 2 }"><div class="step-dot"></div><span>建立SSH连接</span></div>
              <div class="step" :class="{ active: connectStep >= 3, done: connectStep > 3 }"><div class="step-dot"></div><span>身份认证</span></div>
              <div class="step" :class="{ active: connectStep >= 4, done: connectStep > 4 }"><div class="step-dot"></div><span>启动终端</span></div>
            </div>
          </div>
        </div>

        <div class="error-overlay" v-if="error && !connecting">
          <div class="error-content">
            <el-icon :size="48" color="var(--danger)"><CircleCloseFilled /></el-icon>
            <h3>连接失败</h3>
            <p>{{ error }}</p>
            <div class="error-actions">
              <el-button type="primary" @click="reconnect">重新连接</el-button>
              <el-button @click="goBack">返回</el-button>
            </div>
          </div>
        </div>

        <div class="terminal-container" ref="terminalContainer">
          <div ref="terminalElement" class="terminal-element"></div>
        </div>
      </div>

      <transition name="slide-settings">
        <div class="settings-panel" v-show="showSettingsPanel">
          <div class="settings-header">
            <span class="settings-title">终端设置</span>
            <el-button text size="small" :icon="Close" @click="showSettingsPanel = false" class="settings-close-btn" />
          </div>

          <div class="settings-body">
            <el-form label-width="80px" label-position="left" size="small">
              <div class="settings-section">
                <div class="section-title">字体设置</div>
                <el-form-item label="字号">
                  <el-slider v-model="termSettings.fontSize" :min="10" :max="32" :step="1" show-input @change="applySettings" />
                </el-form-item>
                <el-form-item label="字体">
                  <el-select v-model="termSettings.fontFamily" @change="applySettings">
                    <el-option v-for="f in fontOptions" :key="f.value" :label="f.label" :value="f.value" />
                  </el-select>
                </el-form-item>
                <el-form-item label="字重">
                  <el-select v-model="termSettings.fontWeight" @change="applySettings">
                    <el-option label="正常" value="normal" />
                    <el-option label="粗体" value="bold" />
                    <el-option label="100" value="100" />
                    <el-option label="200" value="200" />
                    <el-option label="300" value="300" />
                    <el-option label="400" value="400" />
                    <el-option label="500" value="500" />
                    <el-option label="600" value="600" />
                    <el-option label="700" value="700" />
                  </el-select>
                </el-form-item>
                <el-form-item label="行高">
                  <el-slider v-model="termSettings.lineHeight" :min="1" :max="3" :step="0.1" show-input @change="applySettings" />
                </el-form-item>
                <el-form-item label="字间距">
                  <el-slider v-model="termSettings.letterSpacing" :min="0" :max="10" :step="1" show-input @change="applySettings" />
                </el-form-item>
              </div>

              <div class="settings-section">
                <div class="section-title">光标设置</div>
                <el-form-item label="光标样式">
                  <el-select v-model="termSettings.cursorStyle" @change="applySettings">
                    <el-option label="竖线 (bar)" value="bar" />
                    <el-option label="方块 (block)" value="block" />
                    <el-option label="下划线 (underline)" value="underline" />
                  </el-select>
                </el-form-item>
                <el-form-item label="光标闪烁">
                  <el-switch v-model="termSettings.cursorBlink" @change="applySettings" />
                </el-form-item>
                <el-form-item label="光标颜色">
                  <el-color-picker v-model="termSettings.cursorColor" @change="applySettings" />
                </el-form-item>
              </div>

              <div class="settings-section">
                <div class="section-title">滚动设置</div>
                <el-form-item label="回滚行数">
                  <el-input-number v-model="termSettings.scrollback" :min="100" :max="100000" :step="1000" @change="applySettings" />
                </el-form-item>
                <el-form-item label="提示音">
                  <el-select v-model="termSettings.bellStyle" @change="applySettings">
                    <el-option label="无" value="none" />
                    <el-option label="声音" value="sound" />
                    <el-option label="视觉" value="visual" />
                    <el-option label="两者" value="both" />
                  </el-select>
                </el-form-item>
              </div>

              <div class="settings-section">
                <div class="section-title">颜色设置</div>
                <el-form-item label="背景色">
                  <el-color-picker v-model="termSettings.background" @change="applySettings" />
                </el-form-item>
                <el-form-item label="前景色">
                  <el-color-picker v-model="termSettings.foreground" @change="applySettings" />
                </el-form-item>
                <el-form-item label="选中背景">
                  <el-input v-model="termSettings.selectionBg" @change="applySettings" />
                </el-form-item>
                <el-form-item label="允许透明">
                  <el-switch v-model="termSettings.allowTransparency" @change="applySettings" />
                </el-form-item>
              </div>

              <div class="settings-section">
                <div class="section-title">背景图片</div>
                <el-form-item label="图片URL">
                  <el-input v-model="termSettings.bgImageUrl" placeholder="输入图片URL" clearable @change="applyBgImage">
                    <template #append>
                      <el-button @click="applyBgImage">应用</el-button>
                    </template>
                  </el-input>
                </el-form-item>
                <el-form-item label="">
                  <div class="bg-quick-actions">
                    <el-button size="small" type="primary" plain @click="useRandomBg">随机动漫图</el-button>
                    <el-button size="small" v-if="termSettings.bgImageUrl" :icon="RefreshRight" @click="refreshRandomBg">刷新图片</el-button>
                    <el-button size="small" v-if="termSettings.bgImageUrl" type="danger" plain @click="clearBgImage">清除背景</el-button>
                  </div>
                </el-form-item>
                <el-form-item label="透明度">
                  <el-slider v-model="termSettings.bgOpacity" :min="0" :max="100" :step="5" @change="applyBgImage" />
                </el-form-item>
                <el-form-item label="模糊效果">
                  <el-switch v-model="termSettings.bgBlur" @change="applyBgImage" />
                </el-form-item>
              </div>

              <div class="settings-section">
                <div class="section-title">ANSI颜色</div>
                <div class="color-grid">
                  <div class="color-item" v-for="c in ansiColorFields" :key="c.key">
                    <span class="color-label">{{ c.label }}</span>
                    <el-color-picker v-model="termSettings[c.key]" size="small" @change="applySettings" />
                  </div>
                </div>
              </div>

              <div class="settings-section">
                <div class="section-title">预设主题</div>
                <div class="theme-presets">
                  <div class="theme-preset" v-for="preset in themePresets" :key="preset.name" @click="applyPreset(preset)">
                    <div class="preset-preview" :style="{ background: preset.background }">
                      <span :style="{ color: preset.foreground }">Aa</span>
                    </div>
                    <span class="preset-name">{{ preset.name }}</span>
                  </div>
                </div>
              </div>
            </el-form>

            <div class="settings-actions">
              <el-button size="small" @click="resetSettings">恢复默认</el-button>
              <el-button v-if="!isDirectConnect" type="primary" size="small" @click="saveSettingsToServer" :loading="savingSettings">保存设置</el-button>
            </div>
          </div>
        </div>
      </transition>

      <transition name="slide-sftp">
        <div class="sftp-panel" v-show="sftpOpen && connected">
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
              <el-button text size="small" :icon="RefreshRight" @click="refreshList" :loading="loading" />
              <el-dropdown trigger="click" @command="handleToolbarAction">
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

          <div class="sftp-file-list" v-loading="loading" element-loading-text="加载中...">
            <div
              v-for="file in files"
              :key="file.path"
              class="file-item"
              :class="{ selected: selectedFile?.path === file.path }"
              @click="selectedFile = file"
              @dblclick="handleFileDblClick(file)"
            >
              <el-icon :size="20" :color="getFileIconColor(file)">
                <component :is="file.is_dir ? Folder : Document" />
              </el-icon>
              <div class="file-info">
                <span class="file-name" :title="file.name">{{ file.name }}</span>
                <span class="file-meta">{{ formatSize(file.size) }}</span>
              </div>
              <el-dropdown trigger="click" @command="(cmd) => handleFileAction(cmd, file)" @click.stop>
                <el-button text size="small" :icon="MoreFilled" />
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item v-if="!file.is_dir" command="download" :icon="Download">下载</el-dropdown-item>
                    <el-dropdown-item v-if="!file.is_dir" command="edit" :icon="Edit">编辑</el-dropdown-item>
                    <el-dropdown-item command="rename" :icon="EditPen">重命名</el-dropdown-item>
                    <el-dropdown-item command="delete" :icon="Delete" divided danger>删除</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>

            <div v-if="!loading && files.length === 0" class="empty-files">
              <el-icon :size="40" color="var(--text-muted)"><FolderOpened /></el-icon>
              <p>空目录</p>
            </div>
          </div>
        </div>
      </transition>
    </div>

    <button class="mobile-sftp-toggle" v-if="connected && !sftpOpen && !showSettingsPanel && !keyboardVisible" @click="sftpOpen = true">
      <el-icon :size="18"><FolderOpened /></el-icon>
      文件
    </button>

    <button class="mobile-quick-cmd-toggle" v-if="connected && !keyboardVisible" @click="showMobileQuickCmd = !showMobileQuickCmd">
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

    <div class="mobile-extra-keys" v-if="connected">
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

    <el-dialog v-model="showEditor" title="编辑文件" width="700px" destroy-on-close>
      <el-input v-model="editorContent" type="textarea" :rows="18" :placeholder="'正在编辑: ' + editingFileName" />
      <template #footer>
        <el-button @click="showEditor = false">取消</el-button>
        <el-button type="primary" @click="saveEditedFile" :loading="saving">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showMkdir" title="新建目录" width="400px" destroy-on-close>
      <el-input v-model="newDirName" placeholder="目录名称" @keyup.enter="doMkdir" />
      <template #footer>
        <el-button @click="showMkdir = false">取消</el-button>
        <el-button type="primary" @click="doMkdir">创建</el-button>
      </template>
    </el-dialog>

    <HardwareMonitor
      v-if="connected && showHardwareMonitor"
      :serverId="serverId"
      :visible="showHardwareMonitor"
      @close="showHardwareMonitor = false"
    />
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  ArrowLeft, FolderOpened, Refresh, CircleCloseFilled,
  Close, HomeFilled, RefreshRight, Plus, FolderAdd, Upload,
  Folder, Document, Download, Edit, EditPen, Delete, MoreFilled, Setting,
  Cpu, ArrowDown, Monitor, DataAnalysis
} from '@element-plus/icons-vue'
import { serverAPI, settingsAPI } from '../api'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { WebLinksAddon } from '@xterm/addon-web-links'
import { SearchAddon } from '@xterm/addon-search'
import '@xterm/xterm/css/xterm.css'
import HardwareMonitor from '../components/HardwareMonitor.vue'

const route = useRoute()
const router = useRouter()
const serverId = route.params.id

const serverInfo = ref(null)
const connected = ref(false)
const connecting = ref(false)
const error = ref('')
const connectStep = ref(0)
const connectStatus = ref('')
const terminalContainer = ref(null)
const terminalElement = ref(null)
const sftpOpen = ref(false)
const showSettingsPanel = ref(false)
const showMobileQuickCmd = ref(false)
const showHardwareMonitor = ref(false)
const ctrlPressed = ref(false)
const altPressed = ref(false)
const keyboardVisible = ref(false)
const savingSettings = ref(false)

let term = null, fitAddon = null, pingInterval = null, sftpWs = null
const ws = ref(null)

const isDirectConnect = computed(() => route.path.startsWith('/direct'))
let autoSaveTimer = null

const defaultSettings = {
  fontSize: 14,
  fontFamily: '"Cascadia Code", "Fira Code", "JetBrains Mono", Menlo, Monaco, monospace',
  fontWeight: 'normal',
  lineHeight: 1,
  cursorStyle: 'bar',
  cursorBlink: true,
  cursorColor: '#00d4ff',
  scrollback: 10000,
  background: '#0f0f1a',
  foreground: '#e0e0e0',
  theme_black: '#0f0f1a',
  theme_red: '#ff4757',
  theme_green: '#2ed573',
  theme_yellow: '#ffa502',
  theme_blue: '#3B82F6',
  theme_magenta: '#a855f7',
  theme_cyan: '#00d4ff',
  theme_white: '#e0e0e0',
  theme_bright_black: '#5a6480',
  theme_bright_red: '#ff6b81',
  theme_bright_green: '#7bed9f',
  theme_bright_yellow: '#ffc048',
  theme_bright_blue: '#60a5fa',
  theme_bright_magenta: '#c084fc',
  theme_bright_cyan: '#22d3ee',
  theme_bright_white: '#ffffff',
  selectionBg: 'rgba(0,212,255,0.3)',
  letterSpacing: 0,
  bellStyle: 'none',
  allowTransparency: true,
  bgImageUrl: '',
  bgOpacity: 30,
  bgBlur: false
}

const termSettings = reactive({ ...defaultSettings })

const fontOptions = [
  { label: 'Cascadia Code', value: '"Cascadia Code", Menlo, Monaco, monospace' },
  { label: 'Fira Code', value: '"Fira Code", Menlo, Monaco, monospace' },
  { label: 'JetBrains Mono', value: '"JetBrains Mono", Menlo, Monaco, monospace' },
  { label: 'Source Code Pro', value: '"Source Code Pro", Menlo, Monaco, monospace' },
  { label: 'Consolas', value: 'Consolas, "Courier New", monospace' },
  { label: 'Monaco', value: 'Monaco, Menlo, monospace' },
  { label: 'Menlo', value: 'Menlo, Monaco, monospace' },
  { label: 'Courier New', value: '"Courier New", monospace' },
  { label: 'Ubuntu Mono', value: '"Ubuntu Mono", monospace' },
  { label: 'Noto Sans Mono', value: '"Noto Sans Mono", monospace' },
]

const ansiColorFields = [
  { key: 'theme_black', label: 'Black' },
  { key: 'theme_red', label: 'Red' },
  { key: 'theme_green', label: 'Green' },
  { key: 'theme_yellow', label: 'Yellow' },
  { key: 'theme_blue', label: 'Blue' },
  { key: 'theme_magenta', label: 'Magenta' },
  { key: 'theme_cyan', label: 'Cyan' },
  { key: 'theme_white', label: 'White' },
  { key: 'theme_bright_black', label: 'Bright Black' },
  { key: 'theme_bright_red', label: 'Bright Red' },
  { key: 'theme_bright_green', label: 'Bright Green' },
  { key: 'theme_bright_yellow', label: 'Bright Yellow' },
  { key: 'theme_bright_blue', label: 'Bright Blue' },
  { key: 'theme_bright_magenta', label: 'Bright Magenta' },
  { key: 'theme_bright_cyan', label: 'Bright Cyan' },
  { key: 'theme_bright_white', label: 'Bright White' },
]

const themePresets = [
  {
    name: '暗夜蓝', background: '#0f0f1a', foreground: '#e0e0e0', cursorColor: '#00d4ff',
    selectionBg: 'rgba(0,212,255,0.3)',
    theme_black: '#0f0f1a', theme_red: '#ff4757', theme_green: '#2ed573', theme_yellow: '#ffa502',
    theme_blue: '#3B82F6', theme_magenta: '#a855f7', theme_cyan: '#00d4ff', theme_white: '#e0e0e0',
    theme_bright_black: '#5a6480', theme_bright_red: '#ff6b81', theme_bright_green: '#7bed9f',
    theme_bright_yellow: '#ffc048', theme_bright_blue: '#60a5fa', theme_bright_magenta: '#c084fc',
    theme_bright_cyan: '#22d3ee', theme_bright_white: '#ffffff'
  },
  {
    name: '经典黑', background: '#000000', foreground: '#f0f0f0', cursorColor: '#ffffff',
    selectionBg: 'rgba(255,255,255,0.3)',
    theme_black: '#000000', theme_red: '#cd0000', theme_green: '#00cd00', theme_yellow: '#cdcd00',
    theme_blue: '#0000ee', theme_magenta: '#cd00cd', theme_cyan: '#00cdcd', theme_white: '#e5e5e5',
    theme_bright_black: '#7f7f7f', theme_bright_red: '#ff0000', theme_bright_green: '#00ff00',
    theme_bright_yellow: '#ffff00', theme_bright_blue: '#5c5cff', theme_bright_magenta: '#ff00ff',
    theme_bright_cyan: '#00ffff', theme_bright_white: '#ffffff'
  },
  {
    name: 'Solarized Dark', background: '#002b36', foreground: '#839496', cursorColor: '#93a1a1',
    selectionBg: 'rgba(147,161,161,0.3)',
    theme_black: '#073642', theme_red: '#dc322f', theme_green: '#859900', theme_yellow: '#b58900',
    theme_blue: '#268bd2', theme_magenta: '#d33682', theme_cyan: '#2aa198', theme_white: '#eee8d5',
    theme_bright_black: '#002b36', theme_bright_red: '#cb4b16', theme_bright_green: '#586e75',
    theme_bright_yellow: '#657b83', theme_bright_blue: '#839496', theme_bright_magenta: '#6c71c4',
    theme_bright_cyan: '#93a1a1', theme_bright_white: '#fdf6e3'
  },
  {
    name: 'Solarized Light', background: '#fdf6e3', foreground: '#657b83', cursorColor: '#586e75',
    selectionBg: 'rgba(88,110,117,0.3)',
    theme_black: '#073642', theme_red: '#dc322f', theme_green: '#859900', theme_yellow: '#b58900',
    theme_blue: '#268bd2', theme_magenta: '#d33682', theme_cyan: '#2aa198', theme_white: '#eee8d5',
    theme_bright_black: '#002b36', theme_bright_red: '#cb4b16', theme_bright_green: '#586e75',
    theme_bright_yellow: '#657b83', theme_bright_blue: '#839496', theme_bright_magenta: '#6c71c4',
    theme_bright_cyan: '#93a1a1', theme_bright_white: '#fdf6e3'
  },
  {
    name: 'Dracula', background: '#282a36', foreground: '#f8f8f2', cursorColor: '#f8f8f2',
    selectionBg: 'rgba(248,248,242,0.3)',
    theme_black: '#21222c', theme_red: '#ff5555', theme_green: '#50fa7b', theme_yellow: '#f1fa8c',
    theme_blue: '#bd93f9', theme_magenta: '#ff79c6', theme_cyan: '#8be9fd', theme_white: '#f8f8f2',
    theme_bright_black: '#6272a4', theme_bright_red: '#ff6e6e', theme_bright_green: '#69ff94',
    theme_bright_yellow: '#ffffa5', theme_bright_blue: '#d6acff', theme_bright_magenta: '#ff92df',
    theme_bright_cyan: '#a4ffff', theme_bright_white: '#ffffff'
  },
  {
    name: 'Monokai', background: '#272822', foreground: '#f8f8f2', cursorColor: '#f8f8f0',
    selectionBg: 'rgba(248,248,240,0.3)',
    theme_black: '#272822', theme_red: '#f92672', theme_green: '#a6e22e', theme_yellow: '#f4bf75',
    theme_blue: '#66d9ef', theme_magenta: '#ae81ff', theme_cyan: '#a1efe4', theme_white: '#f8f8f2',
    theme_bright_black: '#75715e', theme_bright_red: '#fd971f', theme_bright_green: '#a6e22e',
    theme_bright_yellow: '#e6db74', theme_bright_blue: '#66d9ef', theme_bright_magenta: '#ae81ff',
    theme_bright_cyan: '#a1efe4', theme_bright_white: '#f9f8f5'
  },
  {
    name: 'Nord', background: '#2e3440', foreground: '#d8dee9', cursorColor: '#d8dee9',
    selectionBg: 'rgba(216,222,233,0.3)',
    theme_black: '#3b4252', theme_red: '#bf616a', theme_green: '#a3be8c', theme_yellow: '#ebcb8b',
    theme_blue: '#81a1c1', theme_magenta: '#b48ead', theme_cyan: '#88c0d0', theme_white: '#e5e9f0',
    theme_bright_black: '#4c566a', theme_bright_red: '#bf616a', theme_bright_green: '#a3be8c',
    theme_bright_yellow: '#ebcb8b', theme_bright_blue: '#81a1c1', theme_bright_magenta: '#b48ead',
    theme_bright_cyan: '#8fbcbb', theme_bright_white: '#eceff4'
  },
  {
    name: 'One Dark', background: '#282c34', foreground: '#abb2bf', cursorColor: '#528bff',
    selectionBg: 'rgba(82,139,255,0.3)',
    theme_black: '#282c34', theme_red: '#e06c75', theme_green: '#98c379', theme_yellow: '#e5c07b',
    theme_blue: '#61afef', theme_magenta: '#c678dd', theme_cyan: '#56b6c2', theme_white: '#abb2bf',
    theme_bright_black: '#5c6370', theme_bright_red: '#e06c75', theme_bright_green: '#98c379',
    theme_bright_yellow: '#e5c07b', theme_bright_blue: '#61afef', theme_bright_magenta: '#c678dd',
    theme_bright_cyan: '#56b6c2', theme_bright_white: '#ffffff'
  },
]

const applyPreset = (preset) => {
  Object.keys(preset).forEach(key => {
    if (key !== 'name') {
      termSettings[key] = preset[key]
    }
  })
  applySettings()
}

const goBack = () => router.push('/')

const getWsUrl = () => {
  const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  return `${proto}//${window.location.host}/api/ws/terminal?server_id=${serverId}&cols=${term?.cols || 80}&rows=${term?.rows || 24}`
}

const getSftpWsUrl = () => {
  const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  return `${proto}//${window.location.host}/api/ws/sftp?server_id=${serverId}`
}

const buildTheme = () => ({
  background: termSettings.background,
  foreground: termSettings.foreground,
  cursor: termSettings.cursorColor,
  cursorAccent: termSettings.background,
  selectionBackground: termSettings.selectionBg,
  black: termSettings.theme_black,
  red: termSettings.theme_red,
  green: termSettings.theme_green,
  yellow: termSettings.theme_yellow,
  blue: termSettings.theme_blue,
  magenta: termSettings.theme_magenta,
  cyan: termSettings.theme_cyan,
  white: termSettings.theme_white,
  brightBlack: termSettings.theme_bright_black,
  brightRed: termSettings.theme_bright_red,
  brightGreen: termSettings.theme_bright_green,
  brightYellow: termSettings.theme_bright_yellow,
  brightBlue: termSettings.theme_bright_blue,
  brightMagenta: termSettings.theme_bright_magenta,
  brightCyan: termSettings.theme_bright_cyan,
  brightWhite: termSettings.theme_bright_white,
})

const initTerminal = () => {
  term = new Terminal({
    cursorBlink: termSettings.cursorBlink,
    cursorStyle: termSettings.cursorStyle,
    fontSize: termSettings.fontSize,
    fontFamily: termSettings.fontFamily,
    fontWeight: termSettings.fontWeight,
    lineHeight: termSettings.lineHeight,
    letterSpacing: termSettings.letterSpacing,
    theme: buildTheme(),
    allowTransparency: termSettings.allowTransparency,
    scrollback: termSettings.scrollback,
    bellStyle: termSettings.bellStyle,
    convertEol: true
  })
  fitAddon = new FitAddon()
  term.loadAddon(fitAddon)
  term.loadAddon(new WebLinksAddon())
  term.loadAddon(new SearchAddon())
  term.open(terminalElement.value)
  nextTick(() => { fitAddon.fit(); applyBgImage() })
  term.onData((data) => {
    if (ws.value?.readyState === WebSocket.OPEN) ws.value.send(JSON.stringify({ type: 'input', data }))
  })
  term.onResize(({ cols, rows }) => {
    if (ws.value?.readyState === WebSocket.OPEN) ws.value.send(JSON.stringify({ type: 'resize', cols, rows }))
  })
}

const applySettings = () => {
  if (!term) return
  term.options.fontSize = termSettings.fontSize
  term.options.fontFamily = termSettings.fontFamily
  term.options.fontWeight = termSettings.fontWeight
  term.options.lineHeight = termSettings.lineHeight
  term.options.letterSpacing = termSettings.letterSpacing
  term.options.cursorStyle = termSettings.cursorStyle
  term.options.cursorBlink = termSettings.cursorBlink
  term.options.scrollback = termSettings.scrollback
  term.options.bellStyle = termSettings.bellStyle
  term.options.allowTransparency = termSettings.allowTransparency

  const newTheme = buildTheme()
  if (termSettings.bgImageUrl) {
    newTheme.background = 'transparent'
  }
  term.options.theme = newTheme

  if (termSettings.bgImageUrl) {
    const xtermEl = terminalElement.value?.querySelector('.xterm')
    const viewport = terminalElement.value?.querySelector('.xterm-viewport')
    if (xtermEl) xtermEl.style.background = 'transparent'
    if (viewport) viewport.style.background = 'transparent'
  }

  nextTick(() => { if (fitAddon) try { fitAddon.fit() } catch {} })
  autoSaveSettings()
}

const applyBgImage = () => {
  if (!terminalContainer.value) return
  const xtermEl = terminalElement.value?.querySelector('.xterm')
  const viewport = terminalElement.value?.querySelector('.xterm-viewport')
  if (termSettings.bgImageUrl) {
    terminalContainer.value.style.backgroundImage = `url(${termSettings.bgImageUrl})`
    terminalContainer.value.style.backgroundSize = 'cover'
    terminalContainer.value.style.backgroundPosition = 'center'
    terminalContainer.value.style.backgroundRepeat = 'no-repeat'
    terminalContainer.value.style.setProperty('--bg-overlay-opacity', termSettings.bgOpacity / 100)
    terminalContainer.value.classList.add('has-bg')
    if (termSettings.bgBlur) {
      terminalContainer.value.style.backdropFilter = 'blur(4px)'
    } else {
      terminalContainer.value.style.backdropFilter = ''
    }
    if (term) {
      term.options.theme = { ...term.options.theme, background: 'transparent' }
    }
    if (xtermEl) {
      xtermEl.style.background = 'transparent'
    }
    if (viewport) {
      viewport.style.background = 'transparent'
    }
  } else {
    terminalContainer.value.style.backgroundImage = ''
    terminalContainer.value.style.backdropFilter = ''
    terminalContainer.value.classList.remove('has-bg')
    if (term) {
      term.options.theme = buildTheme()
    }
    if (xtermEl) {
      xtermEl.style.background = ''
    }
    if (viewport) {
      viewport.style.background = ''
    }
  }
  autoSaveSettings()
}

const clearBgImage = () => {
  termSettings.bgImageUrl = ''
  applyBgImage()
}

const RANDOM_BG_API = 'https://www.loliapi.com/acg/'

const useRandomBg = () => {
  termSettings.bgImageUrl = RANDOM_BG_API + '?t=' + Date.now()
  applyBgImage()
}

const refreshRandomBg = () => {
  if (termSettings.bgImageUrl && termSettings.bgImageUrl.includes('loliapi.com')) {
    termSettings.bgImageUrl = RANDOM_BG_API + '?t=' + Date.now()
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
  if (!term || !ws.value || ws.value.readyState !== WebSocket.OPEN) return

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

  ws.value.send(JSON.stringify({ type: 'input', data }))
  ctrlPressed.value = false
  altPressed.value = false
}

const sendCtrlC = () => {
  if (ws.value && ws.value.readyState === WebSocket.OPEN) {
    ws.value.send(JSON.stringify({ type: 'input', data: '\x03' }))
  }
  ctrlPressed.value = false
  altPressed.value = false
}

const sendCommand = (cmd) => {
  if (term && ws.value?.readyState === WebSocket.OPEN) {
    ws.value.send(JSON.stringify({ type: 'input', data: cmd + '\r' }))
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

const resetSettings = () => {
  Object.assign(termSettings, { ...defaultSettings })
  applySettings()
}

const loadSettingsFromServer = async () => {
  try {
    const res = await settingsAPI.get(serverId)
    if (res.data.success && res.data.settings) {
      const s = res.data.settings
      Object.keys(defaultSettings).forEach(key => {
        if (s[key] !== undefined && s[key] !== null) {
          termSettings[key] = s[key]
        }
      })
    }
  } catch {}
}

const saveSettingsToServer = async () => {
  savingSettings.value = true
  try {
    await settingsAPI.save(serverId, { ...termSettings })
    ElMessage.success('设置已保存')
  } catch {
    ElMessage.error('保存设置失败')
  } finally {
    savingSettings.value = false
  }
}

const autoSaveSettings = () => {
  if (isDirectConnect.value || !serverId) return
  if (autoSaveTimer) clearTimeout(autoSaveTimer)
  autoSaveTimer = setTimeout(async () => {
    try {
      await settingsAPI.save(serverId, { ...termSettings })
    } catch {}
  }, 800)
}

const connectWebSocket = () => {
  if (ws.value) { ws.value.close(); ws.value = null }
  connecting.value = true; error.value = ''; connectStep.value = 0; connectStatus.value = '正在初始化...'
  const steps = [
    { step: 1, msg: '正在解析主机地址...', delay: 400 },
    { step: 2, msg: '正在建立SSH连接...', delay: 800 },
    { step: 3, msg: '正在进行身份认证...', delay: 600 },
    { step: 4, msg: '正在启动终端会话...', delay: 400 }
  ]
  steps.forEach(({ step, msg, delay }) => setTimeout(() => {
    if (connecting.value) { connectStep.value = step; connectStatus.value = msg }
  }, delay))

  ws.value = new WebSocket(getWsUrl())
  ws.value.onopen = () => {}
  ws.value.onmessage = (event) => {
    try {
      const msg = JSON.parse(event.data)
      switch (msg.type) {
        case 'connected':
          connecting.value = false; connected.value = true; connectStep.value = 4
          ElMessage.success(msg.data || '连接成功')
          startPing(); initSFTP()
          break
        case 'output': if (term) term.write(msg.data); break
        case 'error':
          connecting.value = false; connected.value = false; error.value = msg.data
          ElMessage.error(msg.data); break
        case 'disconnected': connected.value = false; ElMessage.warning('连接已断开'); break
        case 'pong': break
      }
    } catch { if (term) term.write(event.data) }
  }
  ws.value.onerror = () => { connecting.value = false; connected.value = false; error.value = 'WebSocket连接失败'; ElMessage.error('连接失败') }
  ws.value.onclose = () => {
    connected.value = false
    if (term) term.write('\r\n\x1b[33m--- 连接已关闭 ---\x1b[0m\r\n')
    stopPing(); closeSFTP()
  }
}

const reconnect = () => { error.value = ''; if (term) term.clear(); connectWebSocket() }

const startPing = () => {
  stopPing()
  pingInterval = setInterval(() => { if (ws.value?.readyState === WebSocket.OPEN) ws.value.send(JSON.stringify({ type: 'ping' })) }, 30000)
}
const stopPing = () => { if (pingInterval) { clearInterval(pingInterval); pingInterval = null } }

const files = ref([])
const loading = ref(false)
const currentPath = ref('/')
const selectedFile = ref(null)
const showEditor = ref(false)
const editorContent = ref('')
const editingFileName = ref('')
const saving = ref(false)
const showMkdir = ref(false)
const newDirName = ref('')
let pendingCallbacks = []

const pathSegments = computed(() => currentPath.value.split('/').filter(Boolean))

const buildPath = (depth) => '/' + pathSegments.value.slice(0, depth).join('/')

const initSFTP = () => {
  closeSFTP()
  sftpWs = new WebSocket(getSftpWsUrl())
  sftpWs.onmessage = (event) => {
    try {
      const msg = JSON.parse(event.data)
      if (msg.type === 'connected') { listDirectory(currentPath.value); return }
      if (msg.type === 'error') { ElMessage.error(msg.data || 'SFTP错误'); return }

      const cb = pendingCallbacks.shift()
      if (msg.files) files.value = msg.files
      if (cb) { msg.success ? cb.resolve(msg) : cb.reject(new Error(msg.message || '操作失败')) }
      else if (msg.message && !msg.files) { msg.success ? ElMessage.success(msg.message) : ElMessage.error(msg.message) }
    } catch {}
  }
  sftpWs.onerror = () => ElMessage.error('SFTP连接失败')
  sftpWs.onclose = () => {}
}

const closeSFTP = () => { if (sftpWs) { sftpWs.close(); sftpWs = null; pendingCallbacks = [] } }

const sendSFTPRequest = (req) => new Promise((resolve, reject) => {
  pendingCallbacks.push({ resolve, reject })
  if (sftpWs?.readyState === WebSocket.OPEN) sftpWs.send(JSON.stringify(req))
  else { pendingCallbacks.pop(); reject(new Error('SFTP未连接')) }
})

const listDirectory = async (path) => {
  loading.value = true
  try { await sendSFTPRequest({ action: 'list', path }) }
  catch (e) { ElMessage.error('获取目录列表失败') }
  finally { setTimeout(() => { loading.value = false }, 200) }
}

const navigateToPath = (path) => { currentPath.value = path; listDirectory(path); selectedFile.value = null }
const refreshList = () => listDirectory(currentPath.value)

const handleFileDblClick = (file) => {
  if (file.is_dir) navigateToPath(file.path)
}

const handleFileAction = async (command, file) => {
  switch (command) {
    case 'download':
      try {
        const res = await sendSFTPRequest({ action: 'download', path: file.path })
        if (res.data) {
          const byteChars = atob(res.data)
          const bytes = new Uint8Array(byteChars.length)
          for (let i = 0; i < byteChars.length; i++) bytes[i] = byteChars.charCodeAt(i)
          const blob = new Blob([bytes])
          const url = URL.createObjectURL(blob)
          const a = document.createElement('a'); a.href = url; a.download = file.name; a.click(); URL.revokeObjectURL(url)
          ElMessage.success('下载成功')
        }
      } catch (e) { ElMessage.error('下载失败') }
      break
    case 'edit':
      try {
        const res = await sendSFTPRequest({ action: 'read', path: file.path })
        editorContent.value = res.data ? atob(res.data) : ''
        editingFileName.value = file.name
        showEditor.value = true
      } catch (e) { ElMessage.error('读取文件失败') }
      break
    case 'delete':
      try {
        await ElMessageBox.confirm(`确定删除 "${file.name}" 吗？`, '确认删除', { type: 'warning' })
        await sendSFTPRequest({ action: 'rm', path: file.path })
      } catch {}
      break
    case 'rename':
      try {
        const { value } = await ElMessageBox.prompt('输入新名称', '重命名', { inputValue: file.name })
        if (value && value !== file.name) {
          const parentPath = file.path.substring(0, file.path.lastIndexOf('/'))
          await sendSFTPRequest({ action: 'rename', path: file.path, dst: parentPath + '/' + value })
        }
      } catch {}
      break
  }
}

const saveEditedFile = async () => {
  saving.value = true
  try {
    const content = btoa(unescape(encodeURIComponent(editorContent.value)))
    await sendSFTPRequest({ action: 'write', path: currentPath.value + '/' + editingFileName.value, content })
    showEditor.value = false; ElMessage.success('保存成功')
  } catch (e) { ElMessage.error('保存失败') }
  finally { saving.value = false }
}

const handleToolbarAction = async (cmd) => {
  switch (cmd) {
    case 'mkdir':
      newDirName.value = ''; showMkdir.value = true; break
    case 'upload':
      const input = document.createElement('input'); input.type = 'file'; input.multiple = true
      input.onchange = async (e) => {
        for (const f of e.target.files) {
          const reader = new FileReader()
          reader.onload = async () => {
            const content = btoa(reader.result)
            await sendSFTPRequest({ action: 'upload', path: currentPath.value, name: f.name, content })
          }
          reader.readAsBinaryString(f)
        }
      }; input.click()
      break
  }
}

const doMkdir = async () => {
  if (!newDirName.value.trim()) return
  try { await sendSFTPRequest({ action: 'mkdir', path: currentPath.value + '/' + newDirName.value.trim() }); showMkdir.value = false }
  catch (e) { ElMessage.error('创建失败') }
}

const toggleSFTP = () => { sftpOpen.value = !sftpOpen.value }

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

const handleResize = () => { if (fitAddon) try { fitAddon.fit() } catch {} }

onMounted(async () => {
  try { const res = await serverAPI.get(serverId); if (res.data.success) serverInfo.value = res.data.server } catch {}
  await loadSettingsFromServer()
  loadQuickCommands()
  initTerminal(); connectWebSocket()
  window.addEventListener('resize', handleResize)
  if (window.visualViewport) {
    window.visualViewport.addEventListener('resize', handleViewportResize)
  }
})

onBeforeUnmount(() => {
  stopPing(); closeSFTP()
  if (ws.value) ws.value.close()
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
.terminal-page { height: 100vh; height: 100dvh; display: flex; flex-direction: column; background: #0f0f1a; position: relative; overflow: hidden; }

.terminal-topbar {
  height: 44px; min-height: 44px; background: var(--bg-secondary); border-bottom: 1px solid var(--border);
  display: flex; align-items: center; justify-content: space-between; padding: 0 12px; flex-shrink: 0;
}
.topbar-left { display: flex; align-items: center; gap: 10px; overflow: hidden; }
.connection-info { display: flex; align-items: center; gap: 8px; overflow: hidden; }
.status-dot { width: 8px; height: 8px; border-radius: 50%; transition: var(--transition); flex-shrink: 0; }
.status-dot.connected { background: var(--success); box-shadow: 0 0 8px rgba(34,197,94,0.5); }
.status-dot.disconnected { background: var(--danger); }
.server-label { font-weight: 600; font-size: 13px; color: var(--text-primary); white-space: nowrap; }
.server-addr { font-size: 11px; color: var(--text-muted); font-family: 'Courier New', monospace; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.topbar-right { display: flex; align-items: center; gap: 4px; flex-shrink: 0; }
.topbar-center { display: flex; align-items: center; flex-shrink: 0; }

.main-layout { display: flex; flex: 1; overflow: hidden; position: relative; }
.terminal-area { flex: 1; min-width: 0; position: relative; overflow: hidden; }
.terminal-container { width: 100%; height: 100%; padding: 6px; overflow: hidden; position: relative; }
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
.terminal-element { width: 100%; height: 100%; overflow: hidden; }

.settings-panel {
  width: 360px; min-width: 300px; max-width: 45vw;
  background: var(--bg-card); border-left: 1px solid var(--border);
  display: flex; flex-direction: column; overflow: hidden;
}
.settings-header {
  height: 42px; padding: 0 12px; display: flex; align-items: center; justify-content: space-between;
  border-bottom: 1px solid var(--border); flex-shrink: 0;
}
.settings-title { font-weight: 600; font-size: 14px; color: var(--text-primary); }
.settings-close-btn { color: var(--text-muted); }
.settings-body {
  flex: 1; overflow-y: auto; padding: 12px 16px;
}
.settings-section {
  margin-bottom: 16px; padding-bottom: 12px; border-bottom: 1px solid var(--border);
}
.settings-section:last-of-type { border-bottom: none; }
.section-title {
  font-size: 13px; font-weight: 600; color: var(--accent); margin-bottom: 10px;
  padding-left: 2px;
}

.bg-quick-actions { display: flex; gap: 6px; flex-wrap: wrap; }
.settings-body :deep(.el-form-item) { margin-bottom: 10px; }
.settings-body :deep(.el-form-item__label) { font-size: 12px; color: var(--text-secondary); }
.settings-body :deep(.el-slider) { padding-right: 50px; }
.settings-actions {
  display: flex; gap: 8px; justify-content: flex-end; padding-top: 12px;
  border-top: 1px solid var(--border);
}

.color-grid {
  display: grid; grid-template-columns: repeat(2, 1fr); gap: 6px;
}
.color-item {
  display: flex; align-items: center; gap: 6px;
}
.color-label {
  font-size: 11px; color: var(--text-secondary); min-width: 70px;
}

.theme-presets {
  display: grid; grid-template-columns: repeat(4, 1fr); gap: 8px;
}
.theme-preset {
  display: flex; flex-direction: column; align-items: center; gap: 4px;
  cursor: pointer; padding: 4px; border-radius: 6px; transition: var(--transition);
}
.theme-preset:hover { background: var(--bg-hover); }
.preset-preview {
  width: 100%; height: 32px; border-radius: 4px; display: flex; align-items: center;
  justify-content: center; font-size: 13px; font-weight: 600; border: 1px solid var(--border);
}
.preset-name { font-size: 10px; color: var(--text-muted); text-align: center; }

.sftp-panel {
  width: 340px; min-width: 280px; max-width: 45vw;
  background: var(--bg-card); border-left: 1px solid var(--border);
  display: flex; flex-direction: column; overflow: hidden;
}
.sftp-header {
  height: 42px; padding: 0 12px; display: flex; align-items: center; justify-content: space-between;
  border-bottom: 1px solid var(--border); flex-shrink: 0;
}
.sftp-title { font-weight: 600; font-size: 14px; color: var(--text-primary); }
.sftp-close-btn { color: var(--text-muted); }

.sftp-toolbar {
  height: 40px; padding: 0 10px; display: flex; align-items: center; justify-content: space-between;
  border-bottom: 1px solid var(--border); flex-shrink: 0; gap: 8px;
}
.sftp-toolbar :deep(.el-breadcrumb) { flex: 1; min-width: 0; }
.sftp-toolbar :deep(.el-breadcrumb__inner) { cursor: pointer; }
.toolbar-actions { display: flex; gap: 2px; flex-shrink: 0; }

.sftp-file-list { flex: 1; overflow-y: auto; padding: 4px 0; -webkit-overflow-scrolling: touch; }
.file-item {
  display: flex; align-items: center; gap: 8px; padding: 7px 12px; cursor: pointer;
  transition: background 0.15s;
}
.file-item:hover { background: var(--bg-hover); }
.file-item.selected { background: var(--accent-dim); }
.file-info { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 1px; }
.file-name { font-size: 13px; color: var(--text-primary); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.file-meta { font-size: 11px; color: var(--text-muted); }
.empty-files { text-align: center; padding: 40px 16px; color: var(--text-muted); }
.empty-files p { margin-top: 8px; font-size: 13px; }

.slide-sftp-enter-active { transition: transform 0.25s ease, opacity 0.25s ease; }
.slide-sftp-leave-active { transition: transform 0.2s ease, opacity 0.2s ease; }
.slide-sftp-enter-from { transform: translateX(30px); opacity: 0; }
.slide-sftp-leave-to { transform: translateX(30px); opacity: 0; }

.slide-settings-enter-active { transition: transform 0.25s ease, opacity 0.25s ease; }
.slide-settings-leave-active { transition: transform 0.2s ease, opacity 0.2s ease; }
.slide-settings-enter-from { transform: translateX(30px); opacity: 0; }
.slide-settings-leave-to { transform: translateX(30px); opacity: 0; }

.mobile-sftp-toggle {
  display: none; position: fixed; bottom: 60px; right: 16px; z-index: 200;
  background: var(--accent); color: #fff; border: none; border-radius: 24px;
  padding: 10px 18px; font-size: 13px; cursor: pointer; box-shadow: 0 4px 12px rgba(0,0,0,0.3);
  gap: 6px; align-items: center;
}

.mobile-quick-cmd-toggle {
  display: none;
  position: fixed;
  bottom: 120px;
  right: 16px;
  z-index: 200;
  background: #3B82F6;
  color: #fff;
  border: none;
  border-radius: 24px;
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
  bottom: 170px;
  right: 16px;
  z-index: 300;
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
  position: absolute; inset: 0; z-index: 100;
  background: rgba(15,15,26,0.95); display: flex; align-items: center; justify-content: center;
  animation: fadeIn 0.3s ease;
}
.connecting-content, .error-content { text-align: center; max-width: 90%; padding: 20px; }
.spinner {
  width: 44px; height: 44px; border: 3px solid var(--border); border-top-color: var(--accent);
  border-radius: 50%; animation: spin 1s linear infinite; margin: 0 auto 20px;
}
.connecting-content h3 { font-size: 17px; color: var(--text-primary); margin-bottom: 6px; word-break: break-word; }
.connecting-content p { font-size: 13px; color: var(--text-muted); margin-bottom: 28px; }
.connect-steps { display: flex; flex-direction: column; gap: 11px; align-items: flex-start; max-width: 190px; margin: 0 auto; }
.step { display: flex; align-items: center; gap: 9px; font-size: 13px; color: var(--text-muted); transition: var(--transition); }
.step.active { color: var(--text-primary); }
.step.done { color: var(--success); }
.step-dot { width: 9px; height: 9px; border-radius: 50%; background: var(--border); transition: var(--transition); flex-shrink: 0; }
.step.active .step-dot { background: var(--accent); animation: pulse 1s infinite; }
.step.done .step-dot { background: var(--success); }
.error-content h3 { font-size: 19px; color: var(--text-primary); margin: 14px 0 7px; }
.error-content p { font-size: 13px; color: var(--text-muted); margin-bottom: 22px; max-width: 360px; word-break: break-word; }
.error-actions { display: flex; gap: 12px; justify-content: center; flex-wrap: wrap; }

@keyframes spin { to { transform: rotate(360deg); } }
@keyframes pulse { 0%, 100% { opacity: 1; } 50% { opacity: 0.4; } }
@keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }

@media (max-width: 768px) {
  .terminal-topbar {
    height: 48px; padding: 0 8px;
    gap: 4px;
  }
  .topbar-left { gap: 6px; }
  .topbar-center { display: none; }
  .server-addr { display: none; }
  .topbar-right { gap: 2px; }
  .topbar-right :deep(.el-button) { padding: 6px !important; font-size: 12px; }

  .terminal-container { padding: 3px; }

  .sftp-panel {
    position: absolute; right: 0; top: 0; bottom: 0; z-index: 150;
    width: 88vw; max-width: 400px; box-shadow: -4px 0 24px rgba(0,0,0,0.5);
    border-left: none;
  }
  .settings-panel {
    position: absolute; right: 0; top: 0; bottom: 0; z-index: 150;
    width: 88vw; max-width: 400px; box-shadow: -4px 0 24px rgba(0,0,0,0.5);
    border-left: none;
  }
  .mobile-sftp-toggle { display: flex; }
  .mobile-quick-cmd-toggle { display: flex; }
  .mobile-quick-cmd-panel { display: block; }
  .mobile-extra-keys { display: flex; }
  .terminal-area { padding-bottom: 50px; }
  .main-layout.sftp-open .terminal-area { filter: brightness(0.55); pointer-events: none; }
  .main-layout.settings-open .terminal-area { filter: brightness(0.55); pointer-events: none; }

  .connecting-content h3 { font-size: 15px; }
  .connect-steps { max-width: 160px; }
  .step { font-size: 12px; gap: 7px; }
  .error-content p { max-width: 260px; font-size: 12px; }
}

@media (max-width: 480px) {
  .terminal-topbar { height: 44px; padding: 0 6px; }
  .server-label { font-size: 12px; }
  .sftp-header { height: 38px; padding: 0 10px; }
  .sftp-toolbar { height: 36px; padding: 0 8px; }
  .sftp-file-list { padding: 2px 0; }
  .file-item { padding: 8px 10px; gap: 6px; }
  .file-name { font-size: 12px; }
  .mobile-sftp-toggle { bottom: 50px; right: 10px; padding: 8px 14px; font-size: 12px; }
  .mobile-quick-cmd-toggle { bottom: 100px; right: 10px; padding: 8px 14px; font-size: 12px; }
  .mobile-quick-cmd-panel { bottom: 150px; right: 10px; min-width: 160px; }
  .settings-panel { width: 95vw; }
  .theme-presets { grid-template-columns: repeat(3, 1fr); }
}
</style>
