import 'element-plus/dist/index.css'

import {
  Lock, Key, User, Monitor, Setting, Document, DataAnalysis,
  Tickets, Files, Plus, MoreFilled, EditPen, Delete, Connection,
  Promotion, Brush, Tools, CopyDocument, FolderOpened, Download, Upload,
  Picture
} from '@element-plus/icons-vue'

const icons = {
  Lock, Key, User, Monitor, Setting, Document, DataAnalysis,
  Tickets, Files, Plus, MoreFilled, EditPen, Delete, Connection,
  Promotion, Brush, Tools, CopyDocument, FolderOpened, Download, Upload,
  Picture
}

export default defineNuxtPlugin((nuxtApp) => {
  for (const [key, component] of Object.entries(icons)) {
    nuxtApp.vueApp.component(key, component)
  }
})
