<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="page-title mb-0">用户管理</h1>
        <p class="page-subtitle mt-0 mb-4">管理系统注册用户</p>
      </div>
      <el-button type="primary" @click="openCreateModal"><el-icon class="mr-1"><Plus /></el-icon>添加用户</el-button>
    </div>

    <div class="bg-white rounded-xl border border-gray-200 shadow-sm overflow-hidden">
      <div class="px-6 py-4 border-b border-gray-100 flex items-center justify-between">
        <span class="text-sm text-gray-500">共 {{ total }} 个用户</span>
        <el-input v-model="searchText" placeholder="搜索用户..." clearable style="width: 240px" />
      </div>

      <el-table :data="filteredUsers" v-loading="loading" stripe style="width: 100%">
        <el-table-column prop="id" label="ID" width="220">
          <template #default="{ row }">
            <span class="font-mono text-xs">{{ row.id }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="username" label="用户名" min-width="150">
          <template #default="{ row }">
            <div class="flex items-center gap-2">
              <el-avatar v-if="row.avatar" :src="row.avatar" :size="28" />
              <span>{{ row.username || row.id }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            <span class="text-gray-500 text-sm">{{ row.created_at || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="160" align="center">
          <template #default="{ row }">
            <el-dropdown trigger="click" @command="(cmd: string) => handleAction(cmd, row)">
              <el-button type="primary" link size="small"><el-icon><MoreFilled /></el-icon></el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="edit"><el-icon><EditPen /></el-icon>编辑信息</el-dropdown-item>
                  <el-dropdown-item command="resetPwd"><el-icon><Key /></el-icon>重置密码</el-dropdown-item>
                  <el-dropdown-item command="delete" divided><el-icon color="#ef4444"><Delete /></el-icon>删除用户</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </template>
        </el-table-column>
      </el-table>

      <div v-if="total > pageSize" class="px-6 py-4 flex justify-center border-t border-gray-100">
        <el-pagination v-model:current-page="currentPage" :total="total" :page-size="pageSize"
                       layout="prev, pager, next" small @current-change="loadUsers" />
      </div>
    </div>

    <el-dialog v-model="createModalOpen" :title="editingUser ? '编辑用户' : '添加用户'" width="480px" :close-on-click-modal="false">
      <el-form label-position="top">
        <el-form-item label="用户ID（2-64位）" required>
          <el-input v-model="userForm.id" placeholder="唯一标识符" :disabled="!!editingUser" />
        </el-form-item>
        <el-form-item label="用户名">
          <el-input v-model="userForm.username" placeholder="显示名称" />
        </el-form-item>
        <el-form-item label="头像URL">
          <el-input v-model="userForm.avatar" placeholder="https://..." />
        </el-form-item>
        <el-form-item v-if="!editingUser" label="密码（至少4位）" required>
          <el-input v-model="userForm.password" type="password" show-password placeholder="登录密码" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createModalOpen = false">取消</el-button>
        <el-button type="primary" :loading="formLoading" @click="handleSaveUser">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="resetPwdOpen" title="重置密码" width="420px" :close-on-click-modal="false">
      <el-form label-position="top">
        <el-form-item label="新密码（至少4位）">
          <el-input v-model="newPassword" type="password" show-password />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="resetPwdOpen = false">取消</el-button>
        <el-button type="primary" :loading="resetLoading" @click="confirmResetPassword">确认重置</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { Plus, MoreFilled, EditPen, Key, Delete } from '@element-plus/icons-vue'
import { ElMessageBox } from 'element-plus'

const { adminAPI } = await import('~/api')

const loading = ref(false)
const users = ref<any[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = 20
const searchText = ref('')

const createModalOpen = ref(false)
const formLoading = ref(false)
const editingUser = ref<any>(null)
const userForm = ref({ id: '', username: '', avatar: '', password: '' })
const resetPwdOpen = ref(false)
const resetLoading = ref(false)
const resetTarget = ref<any>(null)
const newPassword = ref('')

const filteredUsers = computed(() => {
  if (!searchText.value) return users.value
  const q = searchText.value.toLowerCase()
  return users.value.filter((u: any) => (u.username || u.id).toLowerCase().includes(q) || u.id.toLowerCase().includes(q))
})

async function loadUsers() {
  loading.value = true
  try {
    const res = await adminAPI.users.list(currentPage.value, pageSize)
    if (res.data.success) { users.value = res.data.data || []; total.value = res.data.total || 0 }
  } catch (e) { console.error(e) }
  finally { loading.value = false }
}

function openCreateModal() { editingUser.value = null; userForm.value = { id: '', username: '', avatar: '', password: '' }; createModalOpen.value = true }

function handleAction(command: string, user: any) {
  if (command === 'edit') { editingUser.value = user; userForm.value = { id: user.id, username: user.username, avatar: user.avatar, password: '' }; createModalOpen.value = true }
  else if (command === 'resetPwd') { resetTarget.value = user; newPassword.value = ''; resetPwdOpen.value = true }
  else if (command === 'delete') handleDelete(user)
}

async function handleSaveUser() {
  if (!userForm.value.id || (!editingUser.value && !userForm.value.password)) return ElMessage.error('请填写必填项')
  formLoading.value = true
  try {
    const res = editingUser.value
      ? await adminAPI.users.update(userForm.value.id, { username: userForm.value.username, avatar: userForm.value.avatar })
      : await adminAPI.users.create(userForm.value)
    if (res.data.success) { ElMessage.success(res.data.message); createModalOpen.value = false; loadUsers() }
    else ElMessage.error(res.data.message)
  } catch (e: any) { ElMessage.error(e.response?.data?.message || '操作失败') }
  finally { formLoading.value = false }
}

async function confirmResetPassword() {
  if (!newPassword.value || newPassword.value.length < 4) return ElMessage.error('密码至少4位')
  resetLoading.value = true
  try {
    const res = await adminAPI.users.resetPassword(resetTarget.value.id, newPassword.value)
    if (res.data.success) { ElMessage.success('密码已重置'); resetPwdOpen.value = false }
    else ElMessage.error(res.data.message)
  } catch (e: any) { ElMessage.error(e.response?.data?.message || '操作失败') }
  finally { resetLoading.value = false }
}

async function handleDelete(user: any) {
  try {
    await ElMessageBox.confirm(`确定要删除用户 ${user.username || user.id} 吗？此操作不可恢复！`, '确认删除', { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' })
    const res = await adminAPI.users.delete(user.id)
    if (res.data.success) { ElMessage.success('用户已删除'); loadUsers() }
    else ElMessage.error(res.data.message)
  } catch (e: any) { if (e !== 'cancel') ElMessage.error(e.response?.data?.message || '删除失败') }
}

onMounted(() => loadUsers())
</script>
