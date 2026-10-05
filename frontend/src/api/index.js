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
    if (error.response && error.response.status === 401) {
      if (!window.location.pathname.startsWith('/login')) {
        window.location.href = '/login'
      }
    }
    return Promise.reject(error)
  }
)

export const authAPI = {
  register: (data) => api.post('/auth/register', data),
  login: (data) => api.post('/auth/login', data),
  logout: () => api.post('/auth/logout'),
  checkAuth: () => api.get('/auth/check')
}

export const userAPI = {
  getProfile: () => api.get('/user/profile'),
  updateProfile: (data) => api.put('/user/profile', data),
  changePassword: (data) => api.post('/user/password', data)
}

export const serverAPI = {
  list: () => api.get('/servers'),
  get: (id) => api.get(`/servers/${id}`),
  create: (data) => api.post('/servers', data),
  update: (id, data) => api.put(`/servers/${id}`, data),
  delete: (id) => api.delete(`/servers/${id}`),
  test: (id) => api.post(`/servers/${id}/test`),
  groups: () => api.get('/servers/groups'),
  share: (id) => api.post(`/servers/${id}/share`)
}

export const sessionAPI = { list: () => api.get('/sessions') }

export const settingsAPI = {
  get: (serverId) => api.get(`/settings/${serverId}`),
  save: (serverId, data) => api.post(`/settings/${serverId}`, data),
}

export const systemSettingsAPI = {
  get: () => api.get('/system-settings'),
}

export default api
