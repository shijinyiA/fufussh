<template>
  <div class="admin-layout">
    <aside class="admin-sidebar" :class="{ collapsed: isCollapsed, 'is-mobile-open': mobileSidebarOpen }">
      <div class="sidebar-brand">
        <span v-show="!isCollapsed" class="brand-text">SSH-WEB</span>
      </div>

      <nav class="sidebar-nav">
        <NuxtLink v-for="item in menuItems" :key="item.to" :to="item.to" class="sidebar-link"
                  :title="item.label" @click="closeMobileSidebar">
          <el-icon><component :is="item.icon" /></el-icon>
          <span v-show="!isCollapsed" class="link-label">{{ item.label }}</span>
        </NuxtLink>
      </nav>

      <div class="sidebar-footer">
        <button class="collapse-btn desktop-only" @click="isCollapsed = !isCollapsed" :title="isCollapsed ? '展开' : '收起'">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
               :style="{ transform: isCollapsed ? 'rotate(180deg)' : '' }">
            <polyline points="15 18 9 12 15 6"/>
          </svg>
        </button>
      </div>
    </aside>

    <div class="admin-body">
      <header class="admin-topbar">
        <div class="topbar-left">
          <button class="hamburger mobile-only" @click="mobileSidebarOpen = true">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round">
              <line x1="3" y1="6" x2="21" y2="6"/><line x1="3" y1="12" x2="21" y2="12"/><line x1="3" y1="18" x2="21" y2="18"/>
            </svg>
          </button>
          <span class="topbar-breadcrumb">{{ currentRouteLabel }}</span>
        </div>
        <div class="topbar-right">
          <el-dropdown trigger="click" @command="handleCommand">
            <div class="user-area">
              <div class="user-avatar">{{ auth.user.username?.charAt(0)?.toUpperCase() || 'A' }}</div>
              <span class="user-name">{{ auth.user.username }}</span>
              <el-icon><ArrowDown /></el-icon>
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="password">修改密码</el-dropdown-item>
                <el-dropdown-item command="logout" divided>退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </header>

      <main class="admin-content">
        <slot />
      </main>
    </div>

    <div v-if="mobileSidebarOpen" class="sidebar-mask" @click="mobileSidebarOpen = false"></div>

    <el-dialog v-model="passwordModalOpen" title="修改密码" width="420px" :close-on-click-modal="false">
      <el-form @submit.prevent="handleChangePassword" label-position="top">
        <el-form-item label="原密码"><el-input v-model="passwordForm.oldPassword" type="password" show-password /></el-form-item>
        <el-form-item label="新密码（至少6位）"><el-input v-model="passwordForm.newPassword" type="password" show-password /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="passwordModalOpen = false">取消</el-button>
        <el-button type="primary" :loading="passwordLoading" @click="handleChangePassword">确认修改</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ArrowDown } from '@element-plus/icons-vue'
import { User as UserIcon, Monitor, Setting, Document, DataAnalysis, Tickets, Files } from '@element-plus/icons-vue'

const route = useRoute()
const isCollapsed = ref(false)
const mobileSidebarOpen = ref(false)
const passwordModalOpen = ref(false)
const passwordLoading = ref(false)
const passwordForm = ref({ oldPassword: '', newPassword: '' })
const auth = useAdminAuth()

const menuItems = [
  { to: '/', label: '仪表盘', icon: Monitor },
  { to: '/users', label: '用户管理', icon: UserIcon },
  { to: '/servers', label: '服务器列表', icon: DataAnalysis },
  { to: '/settings', label: '系统设置', icon: Setting },
  { to: '/commands', label: '快捷命令', icon: Tickets },
  { to: '/database', label: '数据库管理', icon: Files },
  { to: '/logs', label: '操作日志', icon: Document },
]

const routeLabels: Record<string, string> = {
  '/': '仪表盘',
  '/users': '用户管理',
  '/servers': '服务器列表',
  '/settings': '系统设置',
  '/commands': '快捷命令',
  '/database': '数据库管理',
  '/logs': '操作日志',
}

const currentRouteLabel = computed(() => routeLabels[route.path] || '管理后台')

function closeMobileSidebar() {
  if (window.innerWidth < 1024) mobileSidebarOpen.value = false
}

function handleCommand(command: string) {
  if (command === 'password') { passwordForm.value = { oldPassword: '', newPassword: '' }; passwordModalOpen.value = true }
  if (command === 'logout') auth.logout()
}

async function handleChangePassword() {
  if (!passwordForm.value.oldPassword || !passwordForm.value.newPassword || passwordForm.value.newPassword.length < 6) { ElMessage.error('请填写完整信息且密码至少6位'); return }
  passwordLoading.value = true
  try {
    const { adminAPI } = await import('~/api')
    const res = await adminAPI.changePassword(passwordForm.value)
    res.data.success ? (ElMessage.success('密码修改成功'), passwordModalOpen.value = false) : ElMessage.error(res.data.message || '修改失败')
  } catch (e: any) { ElMessage.error(e.response?.data?.message || '请求失败') }
  finally { passwordLoading.value = false }
}
</script>

<style scoped>
.admin-layout {
  display: flex;
  min-height: 100vh;
  background: #f0f2f5;
}

.admin-sidebar {
  width: 200px;
  min-width: 200px;
  background: linear-gradient(180deg, #1a2332 0%, #202b3d 100%);
  display: flex;
  flex-direction: column;
  transition: width 0.22s ease, min-width 0.22s ease;
  position: sticky;
  top: 0;
  height: 100vh;
  z-index: 30;
  flex-shrink: 0;
}
.admin-sidebar.collapsed { width: 60px; min-width: 60px; }

.sidebar-brand {
  height: 52px;
  display: flex;
  align-items: center;
  padding: 0 16px;
  border-bottom: 1px solid rgba(255,255,255,0.06);
  flex-shrink: 0;
}
.brand-text {
  font-size: 15px;
  font-weight: 700;
  color: #e8ecf1;
  letter-spacing: 1px;
  white-space: nowrap;
}

.sidebar-nav {
  flex: 1;
  padding: 8px 6px;
  overflow-y: auto;
  overflow-x: hidden;
}
.sidebar-link {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border-radius: 6px;
  font-size: 13.5px;
  color: #a3b1c2;
  text-decoration: none;
  transition: all 0.15s ease;
  margin-bottom: 1px;
  cursor: pointer;
}
.sidebar-link .el-icon {
  font-size: 17px;
  color: #73849a;
  flex-shrink: 0;
  transition: color 0.15s;
}
.link-label { white-space: nowrap; }

.sidebar-link:hover {
  background: rgba(255,255,255,0.06);
  color: #d4dde8;
}
.sidebar-link:hover .el-icon { color: #a3b1c2; }

.sidebar-link.router-link-active,
.sidebar-link.router-link-exact-active {
  background: rgba(37,99,235,0.2);
  color: #60a5fa;
  font-weight: 600;
}
.sidebar-link.router-link-active .el-icon,
.sidebar-link.router-link-exact-active .el-icon {
  color: #60a5fa;
}

.sidebar-footer {
  padding: 10px 14px;
  border-top: 1px solid rgba(255,255,255,0.06);
  flex-shrink: 0;
}
.collapse-btn {
  width: 100%;
  height: 32px;
  display: flex; align-items: center; justify-content: center;
  border-radius: 6px;
  border: 1px solid rgba(255,255,255,0.08);
  background: transparent;
  color: #73849a;
  cursor: pointer;
  transition: all 0.15s;
}
.collapse-btn:hover { background: rgba(255,255,255,0.06); color: #a3b1c2; }
.collapse-btn svg { width: 14px; height: 14px; transition: transform 0.22s ease; }

.admin-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  overflow: hidden;
}

.admin-topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 52px;
  padding: 0 20px;
  background: #fff;
  border-bottom: 1px solid #e8ecf1;
  flex-shrink: 0;
  gap: 12px;
}
.topbar-left {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}
.topbar-breadcrumb {
  font-size: 14px;
  font-weight: 600;
  color: #1e293b;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.hamburger {
  display: none;
  width: 34px; height: 34px;
  align-items: center; justify-content: center;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  background: #fff;
  color: #475569;
  cursor: pointer;
  flex-shrink: 0;
}
.hamburger:hover { background: #f8fafc; }
.hamburger svg { width: 17px; height: 17px; }

.topbar-right { flex-shrink: 0; }
.user-area {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 10px 4px 4px;
  border-radius: 20px;
  cursor: pointer;
  transition: background 0.15s;
}
.user-area:hover { background: #f5f7fa; }
.user-avatar {
  width: 28px; height: 28px;
  border-radius: 50%;
  background: linear-gradient(135deg, #2563eb, #3b82f6);
  color: #fff;
  font-size: 12px;
  font-weight: 700;
  display: flex; align-items: center; justify-content: center;
  flex-shrink: 0;
}
.user-name {
  font-size: 13px;
  color: #334155;
  max-width: 90px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.user-area .el-icon { font-size: 11px; color: #94a3b8; }

.admin-content {
  flex: 1;
  padding: 20px;
  overflow-y: auto;
  overflow-x: hidden;
}

.sidebar-mask { display: none; }

@media (max-width: 1023px) {
  .admin-layout { position: relative; overflow-x: hidden; }
  .admin-sidebar {
    position: fixed;
    left: 0; top: 0; bottom: 0;
    z-index: 50;
    width: 240px; min-width: 240px;
    transform: translateX(-100%);
    transition: transform 0.28s cubic-bezier(0.4, 0, 0.2, 1);
    box-shadow: none;
  }
  .admin-sidebar.is-mobile-open {
    transform: translateX(0);
    box-shadow: 8px 0 40px rgba(0,0,0,0.25);
  }
  .admin-sidebar.collapsed { width: 240px; min-width: 240px; }
  .hamburger { display: flex; }
  .desktop-only { display: none !important; }
  .sidebar-mask {
    display: block;
    position: fixed;
    inset: 0;
    z-index: 45;
    background: rgba(0,0,0,0.45);
    backdrop-filter: blur(3px);
  }
  .user-name { display: none; }
  .admin-content { padding: 14px 12px; }
  .admin-topbar { padding: 0 12px; }
}

@media (max-width: 480px) {
  .admin-content { padding: 12px 8px; }
  .admin-topbar { padding: 0 8px; }
  .topbar-breadcrumb { font-size: 13px; }
}
</style>
