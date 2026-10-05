<template>
  <div class="commands-page">
    <div class="page-header">
      <div class="header-left">
        <h3>快捷命令</h3>
        <span class="header-desc">管理预设的SSH命令模板</span>
      </div>
      <el-button type="primary" @click="openCreateModal"><el-icon><Plus /></el-icon>添加命令</el-button>
    </div>

    <div class="cmd-list-card">
      <div v-if="loading" class="loading-area"><el-icon class="is-loading" :size="24"><Loading /></el-icon></div>

      <template v-else-if="commands.length > 0">
        <div class="cmd-list-header">
          <span class="col-sort">#</span>
          <span class="col-name">名称</span>
          <span class="col-cmd">执行命令</span>
          <span class="col-action">操作</span>
        </div>
        <div class="cmd-list-body">
          <div v-for="(cmd, index) in commands" :key="cmd.id || index" class="cmd-row">
            <span class="col-sort">
              <span class="sort-badge">{{ index + 1 }}</span>
            </span>
            <span class="col-name">
              <span class="name-text">{{ cmd.name }}</span>
            </span>
            <span class="col-cmd">
              <code class="cmd-code">{{ cmd.command }}</code>
            </span>
            <span class="col-action">
              <button class="action-btn edit" @click="openEditModal(cmd, index)" title="编辑">
                <el-icon :size="14"><EditPen /></el-icon>编辑
              </button>
              <button class="action-btn copy" @click="copyCommand(cmd.command)" title="复制">
                <el-icon :size="14"><CopyDocument /></el-icon>复制
              </button>
              <button class="action-btn up" :disabled="index === 0" @click="moveUp(index)" title="上移">
                <el-icon :size="14"><Top /></el-icon>
              </button>
              <button class="action-btn down" :disabled="index === commands.length - 1" @click="moveDown(index)" title="下移">
                <el-icon :size="14"><Bottom /></el-icon>
              </button>
              <button class="action-btn del" @click="handleDelete(cmd, index)" title="删除">
                <el-icon :size="14"><Delete /></el-icon>删除
              </button>
            </span>
          </div>
        </div>
      </template>

      <div v-else class="empty-state">
        <el-icon :size="40" color="#c0c4cc"><Tickets /></el-icon>
        <p>暂无快捷命令</p>
        <span>点击右上角添加，设置后将在 SSH 终端中显示</span>
      </div>
    </div>

    <el-dialog v-model="modalOpen" :title="editingIndex !== null ? '编辑命令' : '添加命令'" width="540px"
               :close-on-click-modal="false" destroy-on-close>
      <el-form label-position="top">
        <el-form-item label="命令名称" required>
          <el-input v-model="form.name" placeholder="如：查看磁盘使用率" autofocus />
        </el-form-item>
        <el-form-item label="执行命令" required>
          <el-input v-model="form.command" type="textarea" :rows="4"
                    placeholder="df -h&#10;&#10;支持多行命令" class="font-mono text-xs" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="modalOpen = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">{{ editingIndex !== null ? '更新' : '添加' }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { Plus, EditPen, CopyDocument, Delete, Tickets, Loading, Top, Bottom } from '@element-plus/icons-vue'
import { ElMessageBox } from 'element-plus'

definePageMeta({ middleware: ['admin-auth'] })
const { adminAPI } = await import('~/api')

const loading = ref(false)
const saving = ref(false)
const commands = ref<any[]>([])
const modalOpen = ref(false)
const editingIndex = ref<number | null>(null)
const form = ref({ name: '', command: '' })

async function loadCommands() {
  loading.value = true
  try {
    const res = await adminAPI.commands.get()
    if (res.data.success && res.data.data) commands.value = res.data.data
  } catch (e) { console.error(e) }
  finally { loading.value = false }
}

async function saveAllCommands() {
  saving.value = true
  try {
    const payload = commands.value.map((cmd, i) => ({
      name: cmd.name,
      command: cmd.command,
      sortOrder: (i + 1) * 10
    }))
    const res = await adminAPI.commands.save(payload)
    if (res.data.success) return true
    else { ElMessage.error(res.data.message); return false }
  } catch (e: any) { ElMessage.error(e.response?.data?.message || '保存失败'); return false }
  finally { saving.value = false }
}

function openCreateModal() { editingIndex.value = null; form.value = { name: '', command: '' }; modalOpen.value = true }
function openEditModal(cmd: any, index: number) { editingIndex.value = index; form.value = { name: cmd.name, command: cmd.command }; modalOpen.value = true }

function copyCommand(text: string) {
  navigator.clipboard.writeText(text)
  ElMessage.success('已复制到剪贴板')
}

async function moveUp(index: number) {
  if (index <= 0) return
  const tmp = commands.value[index]
  commands.value.splice(index, 1)
  commands.value.splice(index - 1, 0, tmp)
  await saveAllCommands()
}

async function moveDown(index: number) {
  if (index >= commands.value.length - 1) return
  const tmp = commands.value[index]
  commands.value.splice(index, 1)
  commands.value.splice(index + 1, 0, tmp)
  await saveAllCommands()
}

async function handleSave() {
  if (!form.value.name.trim()) return ElMessage.error('请输入命令名称')
  if (!form.value.command.trim()) return ElMessage.error('请输入执行命令')
  const newCmd = { name: form.value.name.trim(), command: form.value.command.trim() }
  if (editingIndex.value !== null) commands.value[editingIndex.value] = newCmd
  else commands.value.push(newCmd)
  const ok = await saveAllCommands()
  if (ok) { modalOpen.value = false; loadCommands() }
}

async function handleDelete(cmd: any, index: number) {
  try {
    await ElMessageBox.confirm(`确定要删除命令 "${cmd.name}" 吗？`, '确认删除', {
      confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning'
    })
    commands.value.splice(index, 1)
    await saveAllCommands()
    ElMessage.success('已删除')
  } catch (e: any) { if (e !== 'cancel') ElMessage.error('操作失败') }
}

onMounted(() => loadCommands())
</script>

<style scoped>
.commands-page { display: flex; flex-direction: column; gap: 18px; }

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.header-left h3 {
  font-size: 16px;
  font-weight: 700;
  color: #1e293b;
  margin: 0 0 2px 0;
}
.header-desc {
  font-size: 13px;
  color: #94a3b8;
}

.cmd-list-card {
  background: #fff;
  border-radius: 10px;
  border: 1px solid #eef2f7;
  box-shadow: 0 1px 3px rgba(0,0,0,0.04);
  overflow: hidden;
}
.loading-area {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 50px 0;
  color: #c0c4cc;
}

.cmd-list-header {
  display: flex;
  align-items: center;
  padding: 12px 20px;
  background: #f8fafc;
  border-bottom: 1px solid #f1f5f9;
  font-size: 12.5px;
  font-weight: 600;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.3px;
}
.col-sort { width: 56px; flex-shrink: 0; text-align: center; }
.col-name { width: 180px; flex-shrink: 0; }
.col-cmd { flex: 1; min-width: 0; }
.col-action { width: 280px; flex-shrink: 0; text-align: right; }

.cmd-list-body { padding: 4px 0; }
.cmd-row {
  display: flex;
  align-items: center;
  padding: 12px 20px;
  border-bottom: 1px solid #f8fafc;
  transition: background 0.12s;
}
.cmd-row:last-child { border-bottom: none; }
.cmd-row:hover { background: #fafbfc; }

.sort-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 26px; height: 26px;
  border-radius: 7px;
  background: #f1f5f9;
  font-size: 12px;
  font-weight: 700;
  color: #64748b;
}
.name-text {
  font-size: 13.5px;
  font-weight: 600;
  color: #334155;
}
.cmd-code {
  display: block;
  font-size: 12.5px;
  color: #64748b;
  background: #f8fafc;
  padding: 6px 10px;
  border-radius: 6px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 100%;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

.col-action {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 4px;
}
.action-btn {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  padding: 5px 10px;
  border-radius: 6px;
  border: 1px solid transparent;
  background: transparent;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.12s;
  color: #64748b;
}
.action-btn:hover:not(:disabled) { background: #f1f5f9; }
.action-btn.edit:hover { color: #2563eb; background: #eff6ff; border-color: #dbeafe; }
.action-btn.copy:hover { color: #059669; background: #ecfdf5; border-color: #d1fae5; }
.action-btn.up:hover, .action-btn.down:hover { color: #2563eb; background: #eff6ff; }
.action-btn.del:hover { color: #ef4444; background: #fef2f2; border-color: #fecaca; }
.action-btn:disabled { opacity: 0.35; cursor: not-allowed; }

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 50px 20px;
  color: #94a3b8;
}
.empty-state p {
  font-size: 14px;
  font-weight: 600;
  margin: 0;
  color: #64748b;
}
.empty-state span { font-size: 13px; }

@media (max-width: 1023px) {
  .col-name { width: 130px; }
  .col-action { width: 220px; }
  .cmd-code { font-size: 11.5px; padding: 5px 8px; }
}
@media (max-width: 767px) {
  .cmd-list-header { display: none; }
  .cmd-row { flex-wrap: wrap; gap: 8px; padding: 14px 16px; }
  .col-sort { width: auto; order: -1; }
  .col-name { width: 100%; order: 0; }
  .col-cmd { width: 100%; order: 1; }
  .col-action { width: 100%; order: 2; justify-content: flex-start; }
  .cmd-code { white-space: normal; }
}
</style>
