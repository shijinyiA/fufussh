<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="page-title">系统设置</h1>
        <p class="page-subtitle">网站外观、品牌和功能配置</p>
      </div>
    </div>

    <el-tabs v-model="activeTab" type="border-card" class="settings-tabs" @tab-change="onTabChange">
      <el-tab-pane label="站点美化" name="beautify">
        <div class="tab-content-area">
          <el-form label-position="top" class="max-w-3xl">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-5">
              <el-form-item label="网站标题">
                <el-input v-model="form.site_title" placeholder="如：芙芙云 SSH Terminal">
                  <template #suffix><span class="text-xs text-gray-300 truncate max-w-[120px] block">{{ current.site_title }}</span></template>
                </el-input>
              </el-form-item>
              <el-form-item label="页脚文字">
                <el-input v-model="form.footer_text" placeholder="显示在登录页底部">
                  <template #suffix><span class="text-xs text-gray-300 truncate max-w-[120px] block">{{ current.footer_text }}</span></template>
                </el-input>
              </el-form-item>
            </div>

            <div class="space-y-2 mt-4 mb-2">
              <span class="text-xs font-medium text-gray-400 uppercase tracking-wider flex items-center gap-1"><el-icon :size="14"><Picture /></el-icon> 图片资源</span>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-5">
              <el-form-item label="Logo 图片 URL">
                <el-input v-model="form.logo_url" placeholder="https://example.com/logo.png">
                  <template #append v-if="form.logo_url"><el-button text @click="openPreview(form.logo_url)">预览</el-button></template>
                </el-input>
                <div class="mt-1.5 flex items-center gap-2">
                  <span class="text-[11px] text-gray-400 bg-gray-50 px-2 py-0.5 rounded font-mono max-w-full truncate">当前: {{ current.logo_url }}</span>
                  <div v-if="form.logo_url" class="preview-image-wrapper !w-10 !h-10 !rounded-lg shrink-0 cursor-pointer" @click="openPreview(form.logo_url)">
                    <img :src="form.logo_url" alt="" @error="(e) => (e.target as HTMLImageElement).style.display='none'" />
                  </div>
                </div>
              </el-form-item>

              <el-form-item label="Favicon 图标 URL">
                <el-input v-model="form.favicon_url" placeholder="/favicon.svg 或 https://...">
                  <template #append v-if="form.favicon_url"><el-button text @click="openPreview(form.favicon_url)">预览</el-button></template>
                </el-input>
                <div class="mt-1.5 flex items-center gap-2">
                  <span class="text-[11px] text-gray-400 bg-gray-50 px-2 py-0.5 rounded font-mono max-w-full truncate">当前: {{ current.favicon_url }}</span>
                  <div v-if="form.favicon_url" class="preview-image-wrapper !w-8 !h-8 !rounded-md shrink-0 cursor-pointer" @click="openPreview(form.favicon_url)">
                    <img :src="form.favicon_url" alt="" @error="(e) => (e.target as HTMLImageElement).style.display='none'" />
                  </div>
                </div>
              </el-form-item>
            </div>

            <el-form-item label="登录页背景图片 URL">
              <el-input v-model="form.background_url" placeholder="https://www.loliapi.com/acg/ 或自定义地址">
                <template #append v-if="form.background_url"><el-button text @click="openPreview(form.background_url)">预览</el-button></template>
              </el-input>
              <div class="mt-1.5 flex items-center gap-3 flex-wrap">
                <span class="text-[11px] text-gray-400 bg-gray-50 px-2 py-0.5 rounded font-mono max-w-[350px] truncate inline-block">当前: {{ current.background_url }}</span>
                <div v-if="form.background_url" class="preview-image-wrapper !w-24 !h-14 shrink-0 cursor-pointer" style="width:96px;height:56px;" @click="openPreview(form.background_url)">
                  <img :src="form.background_url" alt="" @error="(e) => (e.target as HTMLImageElement).style.display='none'" />
                </div>
              </div>
            </el-form-item>
          </el-form>

          <div class="save-bar">
            <el-button @click="resetTab('beautify')">重置</el-button>
            <el-button type="primary" :loading="savingTab === 'beautify'" @click="saveTab('beautify')">保存站点美化</el-button>
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="外观设置" name="appearance">
        <div class="tab-content-area">
          <el-form label-position="top" class="max-w-3xl">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-5">
              <el-form-item label="主题色">
                <div class="flex items-center gap-3">
                  <el-color-picker v-model="form.theme_color" show-alpha :predefined-colors="['#2563eb','#16a34a','#9333ea','#ea580c','#dc2626','#0891b2']" />
                  <code class="text-sm font-mono bg-gray-100 px-2 py-0.5 rounded text-gray-700">{{ form.theme_color }}</code>
                  <span class="text-[11px] text-gray-400">默认: {{ current.theme_color }}</span>
                </div>
              </el-form-item>
              <el-form-item label="圆角大小">
                <div class="flex items-center gap-3 flex-1">
                  <el-slider v-model="borderRadius" :min="0" :max="16" :step="1" show-input :input-size="'small'" input-style="width: 60px;" />
                </div>
              </el-form-item>
              <el-form-item label="登录页背景模糊">
                <div class="flex items-center gap-3 flex-1">
                  <el-slider v-model="backgroundBlur" :min="0" :max="20" :step="1" show-input :input-size="'small'" input-style="width: 60px;" />
                  <span class="text-xs text-gray-400 ml-1">{{ backgroundBlur > 0 ? backgroundBlur + 'px 模糊' : '不模糊' }}</span>
                </div>
              </el-form-item>
            </div>
          </el-form>

          <div class="save-bar">
            <el-button @click="resetTab('appearance')">重置</el-button>
            <el-button type="primary" :loading="savingTab === 'appearance'" @click="saveTab('appearance')">保存外观设置</el-button>
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="自定义代码" name="custom-code">
        <div class="tab-content-area">
          <p class="text-sm text-gray-500 mb-4 -mt-1">添加的自定义 CSS 和 JS 将注入到用户前端页面中，可用于深度定制外观和添加脚本功能。</p>
          <el-form label-position="top" class="max-w-3xl">
            <el-form-item label="自定义 CSS">
              <div class="relative">
                <el-input v-model="form.custom_css" type="textarea" :rows="6"
                          placeholder=".login-page { background: linear-gradient(135deg, #667eea, #764ba2); }&#10;.login-card { border-radius: 20px; }"
                          class="font-mono text-xs" />
                <div class="absolute top-2 right-2 z-10">
                  <el-tag size="small" :type="form.custom_css ? 'success' : 'info'">{{ form.custom_css ? form.custom_css.length + ' 字符' : '未设置' }}</el-tag>
                </div>
              </div>
              <div class="mt-1 flex items-center gap-2">
                <span class="text-[11px] text-gray-400">当前状态:</span>
                <el-tag size="small" :type="current.custom_css ? 'success' : 'info'">{{ current.custom_css ? current.custom_css.length + ' 字符已生效' : '无' }}</el-tag>
                <button v-if="current.custom_css" type="button" class="text-[11px] text-blue-500 hover:text-blue-700 ml-auto" @click="showCurrentCSS = !showCurrentCSS">
                  {{ showCurrentCSS ? '收起' : '查看当前CSS' }}
                </button>
              </div>
              <pre v-if="showCurrentCSS && current.custom_css" class="mt-2 p-3 bg-gray-900 text-green-400 rounded-lg text-xs overflow-x-auto font-mono leading-relaxed max-h-40 overflow-y-auto">{{ current.custom_css }}</pre>
            </el-form-item>

            <el-form-item label="自定义 JavaScript">
              <div class="relative">
                <el-input v-model="form.custom_js" type="textarea" :rows="5"
                          placeholder="// 自定义脚本示例&#10;console.log('Custom JS loaded');"
                          class="font-mono text-xs" />
                <div class="absolute top-2 right-2 z-10">
                  <el-tag size="small" :type="form.custom_js ? 'warning' : 'info'">{{ form.custom_js ? form.custom_js.length + ' 字符' : '未设置' }}</el-tag>
                </div>
              </div>
              <div class="mt-1 flex items-center gap-2">
                <span class="text-[11px] text-gray-400">当前状态:</span>
                <el-tag size="small" :type="current.custom_js ? 'warning' : 'info'">{{ current.custom_js ? current.custom_js.length + ' 字符已生效' : '无' }}</el-tag>
                <button v-if="current.custom_js" type="button" class="text-[11px] text-blue-500 hover:text-blue-700 ml-auto" @click="showCurrentJS = !showCurrentJS">
                  {{ showCurrentJS ? '收起' : '查看当前JS' }}
                </button>
              </div>
              <pre v-if="showCurrentJS && current.custom_js" class="mt-2 p-3 bg-gray-900 text-yellow-400 rounded-lg text-xs overflow-x-auto font-mono leading-relaxed max-h-40 overflow-y-auto">{{ current.custom_js }}</pre>
            </el-form-item>
          </el-form>

          <div class="save-bar">
            <el-button @click="resetTab('custom-code')">重置</el-button>
            <el-button type="primary" :loading="savingTab === 'custom-code'" @click="saveTab('custom-code')">保存自定义代码</el-button>
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="SEO 设置" name="seo">
        <div class="tab-content-area">
          <el-form label-position="top" class="max-w-3xl">
            <el-form-item label="网站描述">
              <el-input v-model="form.site_description" type="textarea" :rows="2" placeholder="用于搜索引擎展示的描述信息" />
            </el-form-item>
            <el-form-item label="关键词">
              <el-tag v-for="(kw, i) in keywords" :key="i" closable @close="removeKeyword(i)" class="mr-1 mb-1">{{ kw }}</el-tag>
              <el-input v-if="showKwInput" v-model="newKeyword" size="small" style="width: 140px" @keyup.enter="addKeyword" @blur="addKeyword" />
              <el-button v-else size="small" @click="showKwInput = true">+ 添加关键词</el-button>
            </el-form-item>
            <el-form-item label="作者/版权">
              <el-input v-model="form.site_author" placeholder="网站作者或版权信息" />
            </el-form-item>
          </el-form>

          <div class="save-bar">
            <el-button @click="resetTab('seo')">重置</el-button>
            <el-button type="primary" :loading="savingTab === 'seo'" @click="saveTab('seo')">保存 SEO 设置</el-button>
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="功能设置" name="features">
        <div class="tab-content-area">
          <el-form label-position="top" class="max-w-3xl">
            <div class="grid grid-cols-1 md:grid-cols-3 gap-5">
              <el-form-item label="允许注册">
                <el-switch v-model="allowRegister" active-text="开启" inactive-text="关闭" />
              </el-form-item>
              <el-form-item label="最大会话数">
                <el-input-number v-model="maxSessions" :min="1" :max="50" controls-position="right" />
              </el-form-item>
              <el-form-item label="超时时间(分钟)">
                <el-input-number v-model="sessionTimeout" :min="5" :max="1440" :step="5" controls-position="right" />
              </el-form-item>
            </div>

          </el-form>

          <div class="save-bar">
            <el-button @click="resetTab('features')">重置</el-button>
            <el-button type="primary" :loading="savingTab === 'features'" @click="saveTab('features')">保存功能设置</el-button>
          </div>
        </div>
      </el-tab-pane>
    </el-tabs>

    <el-dialog v-model="previewVisible" title="图片预览" width="680px" destroy-on-close>
      <div class="text-center p-4 bg-gray-50 rounded-xl">
        <img :src="previewUrl" alt="预览" class="max-w-full max-h-[420px] mx-auto rounded-lg shadow-md object-contain"
             @error="(e) => { ElMessage.error('图片加载失败'); previewVisible = false }" />
      </div>
      <div class="mt-3 text-center">
        <code class="text-xs text-gray-500 bg-gray-100 px-3 py-1.5 rounded-lg inline-block break-all max-w-full">{{ previewUrl }}</code>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { Monitor, Picture } from '@element-plus/icons-vue'

definePageMeta({ middleware: ['admin-auth'] })
const { adminAPI } = await import('~/api')

const activeTab = ref('beautify')
const savingTab = ref('')
const previewVisible = ref(false)
const previewUrl = ref('')
const borderRadius = ref(8)
const showCurrentCSS = ref(false)
const showCurrentJS = ref(false)

const current = ref({
  site_title: '', logo_url: '', favicon_url: '', background_url: '',
  theme_color: '#2563eb', custom_css: '', custom_js: '', footer_text: ''
})

const form = ref({
  site_title: '', site_description: '', site_keywords: '', site_author: '',
  logo_url: '', favicon_url: '', background_url: '', theme_color: '#2563eb',
  custom_css: '', custom_js: '', footer_text: ''
})

const allowRegister = ref(true)
const maxSessions = ref(10)
const sessionTimeout = ref(30)
const backgroundBlur = ref(0)
const keywords = ref<string[]>([])
const showKwInput = ref(false)
const newKeyword = ref('')

function openPreview(url: string) {
  if (!url) return
  previewUrl.value = url
  previewVisible.value = true
}

function addKeyword() {
  if (newKeyword.value.trim()) { keywords.value.push(newKeyword.value.trim()); newKeyword.value = ''; showKwInput.value = false }
}
function removeKeyword(i: number) { keywords.value.splice(i, 1) }

async function loadSettings() {
  try {
    const res = await adminAPI.settings.get()
    if (res.data.success && res.data.data) {
      const d = res.data.data
      current.value = {
        site_title: d.site_title || 'SSH-WEB',
        logo_url: d.logo_url || 'https://logo.fufuidc.com/logo3.png',
        favicon_url: d.favicon_url || '/favicon.svg',
        background_url: d.background_url || 'https://www.loliapi.com/acg/',
        theme_color: d.theme_color || '#3B82F6',
        custom_css: d.custom_css || '',
        custom_js: d.custom_js || '',
        footer_text: d.footer_text || ''
      }
      form.value = { ...current.value,
        site_description: d.site_description || '',
        site_keywords: d.site_keywords || '',
        site_author: d.site_author || ''
      }
      keywords.value = form.value.site_keywords ? form.value.site_keywords.split(',').map((k: string) => k.trim()).filter(Boolean) : []
      allowRegister.value = d.allow_register !== false
      maxSessions.value = d.max_sessions || 10
      sessionTimeout.value = d.session_timeout || 30
      borderRadius.value = d.border_radius || 8
      backgroundBlur.value = d.background_blur || 0
    }
  } catch (e) { console.error(e) }
}

function getPayloadForTab(tab: string): Record<string, any> {
  const base: Record<string, any> = {}
  if (tab === 'beautify') {
    base.site_title = form.value.site_title
    base.footer_text = form.value.footer_text
    base.logo_url = form.value.logo_url
    base.favicon_url = form.value.favicon_url
    base.background_url = form.value.background_url
  } else if (tab === 'appearance') {
    base.theme_color = form.value.theme_color
    base.border_radius = borderRadius.value
    base.background_blur = backgroundBlur.value
  } else if (tab === 'custom-code') {
    base.custom_css = form.value.custom_css
    base.custom_js = form.value.custom_js
  } else if (tab === 'seo') {
    base.site_description = form.value.site_description
    base.site_keywords = keywords.value.join(',')
    base.site_author = form.value.site_author
  } else if (tab === 'features') {
    base.allow_register = allowRegister.value
    base.max_sessions = maxSessions.value
    base.session_timeout = sessionTimeout.value
  }
  return base
}

async function saveTab(tab: string) {
  savingTab.value = tab
  try {
    const payload = getPayloadForTab(tab)
    const res = await adminAPI.settings.save(payload)
    if (res.data.success) {
      ElMessage.success('已保存')
      loadSettings()
    } else {
      ElMessage.error(res.data.message || '保存失败')
    }
  } catch (e: any) {
    ElMessage.error(e.response?.data?.message || '保存失败')
  } finally {
    savingTab.value = ''
  }
}

function resetTab(tab: string) {
  loadSettings()
}

function onTabChange() {}

onMounted(() => loadSettings())
</script>

<style scoped>
.settings-tabs {
  border-radius: 12px;
  box-shadow: 0 1px 3px rgb(0 0 0 / 0.06);
}
.settings-tabs :deep(.el-tabs__header) {
  background: #fafbfc;
  margin-bottom: 0;
}
.settings-tabs :deep(.el-tabs__item) {
  padding: 0 24px;
  height: 46px;
  line-height: 46px;
  font-size: 13.5px;
  font-weight: 500;
  color: #6b7280;
  transition: all 0.2s;
}
.settings-tabs :deep(.el-tabs__item:hover) {
  color: #2563eb;
}
.settings-tabs :deep(.el-tabs__item.is-active) {
  color: #2563eb;
  font-weight: 600;
}
.settings-tabs :deep(.el-tabs__active-bar) {
  background: #2563eb;
  height: 3px;
  border-radius: 2px;
}
.tab-content-area {
  padding: 28px 8px 8px;
}
.save-bar {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding-top: 16px;
  margin-top: 16px;
  border-top: 1px solid #f0f2f5;
}
.preview-image-wrapper {
  width: 40px; height: 40px; border-radius: 8px; overflow: hidden; border: 1px solid #e5e7eb;
  display: inline-flex; align-items: center; justify-content: center; background: #f9fafb;
}
.preview-image-wrapper img { max-width: 100%; max-height: 100%; object-fit: contain; }
</style>
