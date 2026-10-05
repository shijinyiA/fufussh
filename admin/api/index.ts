import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  timeout: 30000,
  withCredentials: true,
  headers: { 'Content-Type': 'application/json' }
})

api.interceptors.response.use(
  response => response,
  error => {
    if (error.response?.status === 401) {
      localStorage.removeItem('admin_token')

      if (typeof window !== 'undefined' && !window.location.pathname.startsWith('/admin/login')) {
        window.location.href = '/admin/login'
      }
    }
    return Promise.reject(error)
  }
)

function getAdminToken() {
  if (typeof window !== 'undefined') {
    return localStorage.getItem('admin_token') || ''
  }
  return ''
}

export const adminAPI = {
  status: () => api.get('/admin/status'),
  init: (data) => api.post('/admin/init', data),
  login: (data) => api.post('/admin/login', data),
  check: () => api.get('/admin/check', {
    headers: { Authorization: `Bearer ${getAdminToken()}` }
  }),
  changePassword: (data) => api.post('/admin/change-password', data, {
    headers: { Authorization: `Bearer ${getAdminToken()}` }
  }),
  stats: () => api.get('/admin/stats', {
    headers: { Authorization: `Bearer ${getAdminToken()}` }
  }),
  systemInfo: () => api.get('/admin/system-info', {
    headers: { Authorization: `Bearer ${getAdminToken()}` }
  }),
  users: {
    list: (page = 1, pageSize = 20) => api.get(`/admin/users?page=${page}&page_size=${pageSize}`, {
      headers: { Authorization: `Bearer ${getAdminToken()}` }
    }),
    create: (data) => api.post('/admin/users', data, {
      headers: { Authorization: `Bearer ${getAdminToken()}` }
    }),
    update: (id, data) => api.put(`/admin/users/${id}`, data, {
      headers: { Authorization: `Bearer ${getAdminToken()}` }
    }),
    resetPassword: (id, password) => api.put(`/admin/users/${id}/reset-password`, { password }, {
      headers: { Authorization: `Bearer ${getAdminToken()}` }
    }),
    delete: (id) => api.delete(`/admin/users/${id}`, {
      headers: { Authorization: `Bearer ${getAdminToken()}` }
    })
  },
  servers: {
    list: (page = 1, pageSize = 20) => api.get(`/admin/servers?page=${page}&page_size=${pageSize}`, {
      headers: { Authorization: `Bearer ${getAdminToken()}` }
    }),
    delete: (id) => api.delete(`/admin/servers/${id}`, {
      headers: { Authorization: `Bearer ${getAdminToken()}` }
    })
  },
  settings: {
    get: () => api.get('/admin/settings', {
      headers: { Authorization: `Bearer ${getAdminToken()}` }
    }),
    save: (data) => api.post('/admin/settings', data, {
      headers: { Authorization: `Bearer ${getAdminToken()}` }
    })
  },
  commands: {
    get: () => api.get('/admin/commands', {
      headers: { Authorization: `Bearer ${getAdminToken()}` }
    }),
    save: (commands) => api.post('/admin/commands', { commands }, {
      headers: { Authorization: `Bearer ${getAdminToken()}` }
    })
  },
  database: {
    status: () => api.get('/admin/database/status', {
      headers: { Authorization: `Bearer ${getAdminToken()}` }
    }),
    tables: () => api.get('/admin/database/tables', {
      headers: { Authorization: `Bearer ${getAdminToken()}` }
    }),
    backup: () => api.get('/admin/database/backup', {
      headers: { Authorization: `Bearer ${getAdminToken()}` },
      responseType: 'blob'
    }),
    browseTable: (table, page = 1, pageSize = 50) => api.get(`/admin/database/tables/${table}?page=${page}&page_size=${pageSize}`, {
      headers: { Authorization: `Bearer ${getAdminToken()}` }
    }),
    query: (sql) => api.post('/admin/database/query', { sql }, {
      headers: { Authorization: `Bearer ${getAdminToken()}` }
    }),
    deleteRow: (table, id) => api.delete(`/admin/database/tables/${table}/rows`, {
      headers: { Authorization: `Bearer ${getAdminToken()}` },
      data: { id }
    })
  },
  logs: {
    list: (page = 1, pageSize = 50) => api.get(`/admin/logs?page=${page}&page_size=${pageSize}`, {
      headers: { Authorization: `Bearer ${getAdminToken()}` }
    })
  },
  upload: (file) => {
    const formData = new FormData()
    formData.append('file', file)
    return api.post('/admin/upload', formData, {
      headers: {
        'Authorization': `Bearer ${getAdminToken()}`,
        'Content-Type': 'multipart/form-data'
      }
    })
  }
}

export default api
