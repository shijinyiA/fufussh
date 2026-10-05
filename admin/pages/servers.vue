<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="page-title mb-0">服务器列表</h1>
        <p class="page-subtitle mt-0 mb-4">管理SSH连接配置</p>
      </div>
      <el-button type="primary" @click="openCreateModal"><el-icon class="mr-1"><Plus /></el-icon>添加服务器</el-button>
    </div>

    <div class="bg-white rounded-xl border border-gray-200 shadow-sm overflow-hidden">
      <el-table :data="servers" v-loading="loading" stripe style="width: 100%">
        <el-table-column prop="name" label="名称" min-width="140">
          <template #default="{ row }">
            <span class="font-medium">{{ row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="host" label="主机" min-width="160">
          <template #default="{ row }">
            <code class="text-sm bg-gray-100 px-2 py-0.5 rounded">{{ row.host }}:{{ row.port || 22 }}</code>
          </template>
        </el-table-column>
        <el-table-column prop="user" label="用户" width="120">
          <template #default="{ row }">{{ row.user || '-' }}</template>
        </el-table-column>
        <el-table-column label="状态" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 'online' ? 'success' : 'info'" size="small" effect="light">
              {{ row.status === 'online' ? '在线' : '离线' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" align="center">
          <template #default="{ row }">
            <el-dropdown trigger="click" @command="(cmd: string) => handleAction(cmd, row)">
              <el-button type="primary" link size="small"><el-icon><MoreFilled /></el-icon></el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="edit"><el-icon><EditPen /></el-icon>编辑</el-dropdown-item>
                  <el-dropdown-item command="test"><el-icon><Connection /></el-icon>测试连接</el-dropdown-item>
                  <el-dropdown-item command="delete" divided><el-icon color="#ef4444"><Delete /></el-icon>删除</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </template>
        </el-table-column>
      </el-table>

      <div v-if="total > pageSize" class="px-6 py-4 flex justify-center border-t border-gray-100">
        <el-pagination v-model:current-page="currentPage" :total="total" :page-size="pageSize"
                       layout="prev, pager, next" small @current-change="loadServers" />
      </div>
    </div>

    <el-dialog v-model="modalOpen" :title="editingServer ? '编辑服务器' : '添加服务器'" width="520px" :close-on-click-modal="false">
      <el-form label-position="top">
        <el-form-item label="名称" required><el-input v-model="form.name" placeholder="如：生产服务器" /></el-form-item>
        <div class="grid grid-cols-2 gap-4">
          <el-form-item label="主机地址" required><el-input v-model="form.host" placeholder="IP或域名" /></el-form-item>
          <el-form-item label="端口"><el-input-number v-model="form.port" :min="1" :max="65535" :step="1" style="width: 100%" /></el-form-item>
        </div>
        <el-form-item label="用户名"><el-input v-model="form.user" placeholder="SSH登录用户" /></el-form-item>
        <el-form-item label="备注"><el-input v-model="form.description" type="textarea" :rows="2" placeholder="可选描述信息" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="modalOpen = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { Plus, MoreFilled, EditPen, Connection, Delete } from '@element-plus/icons-vue'
import { ElMessageBox } from 'element-plus'

const { adminAPI } = await import('~/api')

const loading = ref(false)
const saving = ref(false)
const servers = ref<any[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = 20
const modalOpen = ref(false)
const editingServer = ref<any>(null)
const form = ref({ name: '', host: '', port: 22, user: '', description: '' })

async function loadServers() {
  loading.value = true
  try {
    const res = await adminAPI.servers.list(currentPage.value, pageSize)
    if (res.data.success) { servers.value = res.data.data || []; total.value = res.data.total || 0 }
  } catch (e) { console.error(e) }
  finally { loading.value = false }
}

function openCreateModal() { editingServer.value = null; form.value = { name: '', host: '', port: 22, user: '', description: '' }; modalOpen.value = true }

function handleAction(cmd: string, server: any) {
  if (cmd === 'edit') { editingServer.value = server; form.value = { name: server.name, host: server.host, port: server.port || 22, user: server.user || '', description: server.description || '' }; modalOpen.value = true }
  else if (cmd === 'test') ElMessage.info('连接测试功能开发中')
  else if (cmd === 'delete') handleDelete(server)
}

async function handleSave() {
  if (!form.value.name || !form.value.host) return ElMessage.error('请填写名称和主机地址')
  saving.value = true
  try {
    const res = editingServer.value
      ? await adminAPI.servers.update(editingServer.id, form.value)
      : await adminAPI.servers.create(form.value)
    if (res.data.success) { ElMessage.success(res.data.message); modalOpen.value = false; loadServers() }
    else ElMessage.error(res.data.message)
  } catch (e: any) { ElMessage.error(e.response?.data?.message || '操作失败') }
  finally { saving.value = false }
}

async function handleDelete(server: any) {
  try {
    await ElMessageBox.confirm(`确定要删除服务器 ${server.name} 吗？`, '确认删除', { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' })
    const res = await adminAPI.servers.delete(server.id)
    if (res.data.success) { ElMessage.success('已删除'); loadServers() }
    else ElMessage.error(res.data.message)
  } catch (e: any) { if (e !== 'cancel') ElMessage.error(e.response?.data?.message || '删除失败') }
}

onMounted(() => loadServers())
</script>
