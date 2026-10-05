<template>
  <div class="db-page">
    <div class="page-header">
      <div class="header-left">
        <h3>数据库管理</h3>
        <span class="header-desc">数据表状态、备份与恢复</span>
      </div>
      <el-button type="primary" :loading="backingUp" @click="handleBackup">
        <el-icon><Download /></el-icon>备份数据库
      </el-button>
    </div>

    <div class="stats-row">
      <div class="mini-stat">
        <span class="mini-label">数据库类型</span>
        <span class="mini-value type-tag">MySQL</span>
      </div>
      <div class="mini-stat">
        <span class="mini-label">数据表数</span>
        <span class="mini-value">{{ dbInfo.tableCount }}</span>
      </div>
      <div class="mini-stat">
        <span class="mini-label">总记录数</span>
        <span class="mini-value">{{ formatNumber(dbInfo.totalRows) }}</span>
      </div>
      <div class="mini-stat">
        <span class="mini-label">总大小</span>
        <span class="mini-value">{{ dbInfo.totalSize }} MB</span>
      </div>
    </div>

    <div class="table-card">
      <div class="card-header">
        <h4>数据表列表</h4>
        <el-button :icon="Refresh" text size="small" @click="loadDbInfo">刷新</el-button>
      </div>

      <div v-if="loadingTables" class="loading-area"><el-icon class="is-loading" :size="24"><Loading /></el-icon></div>

      <template v-else>
        <div class="table-list-header">
          <span class="th-name">表名</span>
          <span class="th-rows">行数</span>
          <span class="th-size">大小</span>
          <span class="th-time">创建时间</span>
          <span class="th-action">操作</span>
        </div>
        <div class="table-list-body">
          <div v-for="(t, i) in tables" :key="t.name" class="table-row" :class="{ odd: i % 2 === 1 }">
            <span class="td-name"><code>{{ t.name }}</code></span>
            <span class="td-rows">{{ formatNumber(t.rows) }}</span>
            <span class="td-size">{{ t.size }} KB</span>
            <span class="td-time">{{ t.created_at || '-' }}</span>
            <span class="td-action">
              <button class="link-btn" @click="browseTable(t.name)">浏览</button>
              <button class="link-btn danger" @click="confirmDeleteTable(t.name)">清空</button>
            </span>
          </div>
        </div>

        <div v-if="tables.length === 0" class="empty-state">
          <el-icon :size="36" color="#c0c4cc"><FolderOpened /></el-icon>
          <p>暂无数据表</p>
        </div>
      </template>
    </div>

    <el-dialog v-model="browseOpen" :title="'浏览: ' + browseTableTitle" width="90%" top="5vh"
               :close-on-click-modal="false" destroy-on-close>
      <div v-if="browseLoading" class="text-center py-8"><el-icon class="is-loading" :size="28"><Loading /></el-icon></div>
      <template v-else>
        <div class="mb-3 flex items-center justify-between">
          <span class="text-sm text-gray-500">共 {{ browseTotal }} 条记录</span>
          <el-pagination v-model:current-page="browsePage" :total="browseTotal" :page-size="50"
                         layout="prev, pager, next" small @current-change="doBrowseTable" />
        </div>
        <div class="overflow-auto border rounded-lg" style="max-height: 55vh;">
          <el-table :data="browseRows" stripe size="small" border>
            <el-table-column v-for="col in browseColumns" :key="col" :prop="col" :label="col" min-width="120" show-overflow-tooltip />
            <el-table-column label="操作" width="80" align="center" fixed="right">
              <template #default="{ row }">
                <el-button type="danger" link size="small" @click="deleteRow(browseTableTitle, row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { Download, Refresh, Loading, FolderOpened } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'

definePageMeta({ middleware: ['admin-auth'] })
const { adminAPI } = await import('~/api')

const loadingTables = ref(false)
const backingUp = ref(false)
const tables = ref<any[]>([])
const dbInfo = ref<any>({ tableCount: 0, totalRows: 0, totalSize: 0 })

const browseOpen = ref(false)
const browseLoading = ref(false)
const browseTableTitle = ref('')
const browseRows = ref<any[]>([])
const browseColumns = ref<string[]>([])
const browseTotal = ref(0)
const browsePage = ref(1)

async function loadDbInfo() {
  loadingTables.value = true
  try {
    const res = await adminAPI.database.status()
    const d = res.data
    if (d.success) {
      tables.value = d.tables || []
      dbInfo.value = {
        tableCount: d.table_count || 0,
        totalRows: d.total_rows || 0,
        totalSize: d.total_size || 0
      }
    }
  } catch (e) { console.error(e) }
  finally { loadingTables.value = false }
}

function formatNumber(n: number | string): string {
  return Number(n).toLocaleString()
}

async function handleBackup() {
  backingUp.value = true
  try {
    const res = await adminAPI.backup()
    const blob = new Blob([res.data], { type: 'application/octet-stream' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `backup_${new Date().toISOString().slice(0,10)}.sql`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
    ElMessage.success('备份完成，文件已下载')
  } catch (e: any) {
    ElMessage.error(e.response?.data?.message || '备份失败')
  }
  finally { backingUp.value = false }
}

async function browseTable(name: string) {
  browseTableTitle.value = name
  browsePage.value = 1
  browseOpen.value = true
  await doBrowseTable()
}

async function doBrowseTable() {
  browseLoading.value = true
  try {
    const res = await adminAPI.database.browseTable(browseTableTitle.value, browsePage.value, 50)
    const d = res.data
    if (d.success) {
      browseColumns.value = d.columns || []
      browseRows.value = d.rows || []
      browseTotal.value = d.total || 0
    }
  } catch (e) { console.error(e) }
  finally { browseLoading.value = false }
}

async function deleteRow(table: string, row: any) {
  if (!row) return
  const keys = Object.keys(row)
  if (!keys.length) return
  try {
    await ElMessageBox.confirm('确定要删除这条记录吗？', '确认删除', { type: 'warning' })
    await adminAPI.database.deleteRow(table, row[keys[0]])
    ElMessage.success('已删除')
    doBrowseTable()
    loadDbInfo()
  } catch (e: any) { if (e !== 'cancel') ElMessage.error('删除失败') }
}

async function confirmDeleteTable(name: string) {
  try {
    await ElMessageBox.confirm(`确定要清空表 "${name}" 的所有数据吗？此操作不可恢复！`, '危险操作', {
      confirmButtonText: '确认清空', cancelButtonText: '取消', type: 'error'
    })
    await adminAPI.database.query({ sql: `TRUNCATE TABLE ${name}` })
    ElMessage.success(`${name} 已清空`)
    loadDbInfo()
  } catch (e: any) { if (e !== 'cancel') ElMessage.error('操作失败') }
}

onMounted(() => loadDbInfo())
</script>

<style scoped>
.db-page { display: flex; flex-direction: column; gap: 18px; }

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
.header-desc { font-size: 13px; color: #94a3b8; }

.stats-row {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 14px;
}
.mini-stat {
  background: #fff;
  border-radius: 10px;
  padding: 16px 20px;
  border: 1px solid #eef2f7;
  box-shadow: 0 1px 3px rgba(0,0,0,0.04);
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.mini-label {
  font-size: 12.5px;
  color: #94a3b8;
  font-weight: 500;
}
.mini-value {
  font-size: 22px;
  font-weight: 800;
  color: #1e293b;
}
.type-tag {
  background: linear-gradient(135deg, #f97316, #fb923c);
  color: #fff;
  padding: 2px 12px;
  border-radius: 5px;
  font-size: 14px;
  letter-spacing: 0.5px;
}

.table-card {
  background: #fff;
  border-radius: 10px;
  border: 1px solid #eef2f7;
  box-shadow: 0 1px 3px rgba(0,0,0,0.04);
  overflow: hidden;
}
.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid #f1f5f9;
}
.card-header h4 {
  font-size: 15px;
  font-weight: 700;
  color: #1e293b;
  margin: 0;
}
.loading-area {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 50px 0;
  color: #c0c4cc;
}

.table-list-header {
  display: flex;
  align-items: center;
  padding: 11px 20px;
  background: #f8fafc;
  border-bottom: 1px solid #f1f5f9;
  font-size: 12.5px;
  font-weight: 600;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.3px;
}
.th-name { width: 200px; flex-shrink: 0; }
.th-rows { width: 100px; flex-shrink: 0; text-align: right; }
.th-size { width: 100px; flex-shrink: 0; text-align: right; }
.th-time { flex: 1; min-width: 140px; }
.th-action { width: 130px; flex-shrink: 0; text-align: right; }

.table-list-body { padding: 2px 0; }
.table-row {
  display: flex;
  align-items: center;
  padding: 11px 20px;
  transition: background 0.12s;
}
.table-row.odd { background: #fafbfc; }
.table-row:hover { background: #eff6ff; }

.td-name { width: 200px; flex-shrink: 0; }
.td-name code {
  font-size: 13px;
  font-weight: 600;
  color: #2563eb;
  background: #eff6ff;
  padding: 3px 8px;
  border-radius: 5px;
}
.td-rows { width: 100px; flex-shrink: 0; text-align: right; font-size: 13px; font-weight: 600; color: #334155; }
.td-size { width: 100px; flex-shrink: 0; text-align: right; font-size: 13px; color: #64748b; }
.td-time { flex: 1; min-width: 140px; font-size: 12.5px; color: #94a3b8; }
.td-action { width: 130px; flex-shrink: 0; text-align: right; display: flex; gap: 6px; justify-content: flex-end; }

.link-btn {
  font-size: 12px;
  padding: 3px 10px;
  border: none;
  background: transparent;
  color: #2563eb;
  cursor: pointer;
  border-radius: 4px;
  font-weight: 600;
}
.link-btn:hover { background: #eff6ff; }
.link-btn.danger { color: #ef4444; }
.link-btn.danger:hover { background: #fef2f2; }

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 40px 20px;
  color: #94a3b8;
}
.empty-state p { font-size: 14px; font-weight: 600; margin: 0; color: #64748b; }

@media (max-width: 1023px) {
  .stats-row { grid-template-columns: repeat(2, 1fr); }
  .th-name, .td-name { width: 150px; }
  .th-action, .td-action { width: 110px; }
}
@media (max-width: 767px) {
  .stats-row { grid-template-columns: repeat(2, 1fr); }
  .table-list-header { display: none; }
  .table-row { flex-wrap: wrap; gap: 6px; padding: 12px 16px; }
  .td-name { width: 100%; order: -1; }
  .td-rows, .td-size, .td-time, .td-action { width: auto; }
}
</style>
