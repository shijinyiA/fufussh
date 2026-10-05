import { defineNuxtConfig } from 'nuxt/config'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'

export default defineNuxtConfig({
  compatibilityDate: '2025-01-01',
  devtools: { enabled: true },
  app: {
    baseURL: '/admin/',
    buildAssetsDir: 'assets/',
    head: {
      title: '管理员后台 - SSH-WEB',
      meta: [
        { name: 'viewport', content: 'width=device-width, initial-scale=1' }
      ]
    }
  },
  ssr: false,
  nitro: {
    output: {
      dir: '../backend/admin_dist',
      publicDir: '../backend/admin_dist'
    },
    preset: 'static',
    publicAssets: []
  },
  css: ['~/assets/css/admin.css'],
  runtimeConfig: {
    public: {
      apiBase: '/api'
    }
  },
  vite: {
    plugins: [
      AutoImport({
        resolvers: [ElementPlusResolver()],
        dts: false
      }),
      Components({
        resolvers: [ElementPlusResolver({ importStyle: false })],
        dts: false
      })
    ],
    ssr: {
      noExternal: ['element-plus', '@element-plus/icons-vue']
    }
  },
  build: {
    transpile: ['element-plus', '@element-plus/icons-vue']
  }
})
