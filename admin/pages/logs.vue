<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="page-title mb-0">操作日志</h1>
        <p class="page-subtitle mt-0 mb-4">系统操作审计记录</p>
      </div>
      <el-button type="danger" plain :loading="clearing" @click="handleClear"><el-icon class="mr-1"><Delete /></el-icon>清空日志</el-button>
    </div>

    <div class="bg-white rounded-xl border border-gray-200 shadow-sm overflow-hidden">
      <div class="px-6 py-4 border-b border-gray-100 flex items-center justify-between flex-wrap gap-3">
        <span class="text-sm text-gray-500">共 {{ total }} 条记录</span>
        <div class="flex items-center gap-3">
          <el-select v-model="filterType" clearable placeholder="操作类型" size="default" style="width: 140px">
            <el-option v-for="t in actionTypes" :key="t" :label="t" :value="t" />
          </el-select>
          <el-date-picker v-model="dateRange" type="daterange" range-separator="至" start-placeholder="开始日期" end-placeholder="结束日期"
                          value-format="YYYY-MM-DD" size="default" style="width: 260px" />
        </div>
      </div>

      <el-table :data="logs" v-loading="loading" stripe style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="action" label="操作类型" width="130">
          <template #default="{ row }">
            <el-tag :type="getActionType(row.action).tagType" size="small" effect="light">{{ row.action }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="target" label="目标" min-width="150">
          <template #default="{ row }"><span class="font-mono text-sm">{{ row.target || '-' }}</span></template>
        </el-table-column>
        <el-table-column prop="operator" label="操作人" width="140" />
        <el-table-column prop="ip" label="IP地址" width="140">
          <template #default="{ row }"><code class="text-xs bg-gray-100 px-1.5 py-0.5 rounded">{{ row.ip || '-' }}</code></template>
        </el-table-column>
        <el-table-column prop="created_at" label="时间" min-width="170">
          <template #default="{ row }"><span class="text-gray-500 text-sm">{{ row.created_at || '-' }}</span></template>
        </el-table-column>
      </el-table>

      <div v-if="total > pageSize" class="px-6 py-4 flex justify-center border-t border-gray-100">
        <el-pagination v-model:current-page="currentPage" :total="total" :page-size="pageSize"
                       layout="prev, pager, next" small @current-change="loadLogs" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Delete } from '@element-plus/icons-vue'
import { ElMessageBox } from 'element-plus'

definePageMeta({ middleware: ['admin-auth'] })
const { adminAPI } = await import('~/api')

const loading = ref(false)
const clearing = ref(false)
const logs = ref<any[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = 25
const filterType = ref('')
const dateRange = ref<[string, string] | null>(null)
const actionTypes = ['login', 'logout', 'create', 'update', 'delete', 'backup', 'restore']

function getActionType(action: string): { tagType: '' | 'success' | 'warning' | 'danger' | 'info' } {
  const map: Record<string, any> = { login: 'success', logout: 'info', create: 'primary', update: 'warning', delete: 'danger', backup: '', restore: '' }
  return map[action] || 'info'
}

async function loadLogs() {
  loading.value = true
  try {
    const params: Record<string, any> = { page: currentPage.value, page_size: pageSize }
    if (filterType.value) params.action_type = filterType.value
    if (dateRange.value?.[0]) params.start_date = dateRange.value[0]
    if (dateRange.value?.[1]) params.end_date = dateRange.value[1]
    const res = await adminAPI.logs.list(params)
    if (res.data.success) { logs.value = res.data.data || []; total.value = res.data.total || 0 }
  } catch (e) { console.error(e) }
  finally { loading.value = false }
}

async function handleClear() {
  try {
    await ElMessageBox.confirm('确定要清空所有操作日志吗？此操作不可恢复！', '确认清空', { confirmButtonText: '清空', cancelButtonText: '取消', type: 'warning' })
    clearing.value = true
    const res = await adminAPI.logs.clear()
    clearing.value = false
    if (res.data.success) { ElMessage.success('日志已清空'); loadLogs() }
    else ElMessage.error(res.data.message)
  } catch (e: any) { clearing.value = false; if (e !== 'cancel') ElMessage.error(e.response?.data?.message || '清空失败') }
}

onMounted(() => loadLogs())

watch([filterType, dateRange], () => { currentPage.value = 1; loadLogs() })
</script>
