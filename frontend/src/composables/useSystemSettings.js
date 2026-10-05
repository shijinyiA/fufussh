import { ref, watch } from 'vue'
import { systemSettingsAPI } from '../api'

const settings = ref({
  site_title: '芙芙云 SSH Terminal',
  site_description: '',
  site_keywords: '',
  site_author: '',
  logo_url: 'https://logo.fufuidc.com/logo3.png',
  favicon_url: '/favicon.svg',
  background_url: 'https://www.loliapi.com/acg/',
  background_blur: 0,
  theme_color: '#3B82F6',
  custom_css: '',
  custom_js: '',
  allow_register: true,
  captcha_enabled: false,
  captcha_id: '',
  captcha_key: ''
})

let loaded = false
let loadPromise = null

export function useSystemSettings() {

  async function loadSettings() {
    if (loaded) return settings.value
    if (loadPromise) return loadPromise

    loadPromise = systemSettingsAPI.get()
      .then(res => {
        if (res.data.success && res.data.settings) {
          Object.assign(settings.value, res.data.settings)
          applySettings(settings.value)
          loaded = true
        }
        return settings.value
      })
      .catch(err => {
        console.warn('加载系统设置失败，使用默认值:', err)
        applySettings(settings.value)
        return settings.value
      })
      .finally(() => {
        loadPromise = null
      })

    return loadPromise
  }

  function applySettings(s) {
    if (s.site_title) {
      document.title = s.site_title
    }

    if (s.favicon_url && s.favicon_url !== '/favicon.svg') {
      let link = document.querySelector("link[rel*='icon']") || document.createElement('link')
      link.type = 'image/x-icon'
      link.rel = 'shortcut icon'
      link.href = s.favicon_url
      document.getElementsByTagName('head')[0].appendChild(link)
    }

    if (s.theme_color) {
      document.documentElement.style.setProperty('--accent', s.theme_color)

      let metaThemeColor = document.querySelector('meta[name="theme-color"]')
      if (!metaThemeColor) {
        metaThemeColor = document.createElement('meta')
        metaThemeColor.name = 'theme-color'
        document.head.appendChild(metaThemeColor)
      }
      metaThemeColor.content = s.theme_color
    }

    if (s.custom_css) {
      let customStyle = document.getElementById('system-custom-css')
      if (!customStyle) {
        customStyle = document.createElement('style')
        customStyle.id = 'system-custom-css'
        document.head.appendChild(customStyle)
      }
      customStyle.textContent = s.custom_css
    }
  }

  watch(() => settings.value, (newVal) => {
    applySettings(newVal)
  }, { deep: true })

  return {
    settings,
    loadSettings,
    applySettings
  }
}

export default useSystemSettings
