import { useState } from '#app'
import { adminAPI } from '~/api'

export const useAdminAuth = () => {
  const token = useState('admin_token', () =>
    typeof window !== 'undefined' ? localStorage.getItem('admin_token') || '' : ''
  )
  const user = useState('admin_user', () => ({ username: '', password_changed: true }))
  const initialized = useState('admin_initialized', () => true)

  async function checkStatus() {
    try {
      const res = await adminAPI.status()
      initialized.value = res.data.initialized
      return res.data
    } catch (e) {
      console.error('检查状态失败', e)
      return { initialized: false }
    }
  }

  async function initAdmin(username: string, password: string) {
    const res = await adminAPI.init({ username, password })
    if (res.data.success) {
      initialized.value = true
    }
    return res.data
  }

  async function login(payload: { username: string; password: string }) {
    const res = await adminAPI.login(payload)
    if (res.data.success) {
      token.value = res.data.token
      user.value = { username: res.data.username, password_changed: res.data.password_changed }
      if (typeof window !== 'undefined') {
        localStorage.setItem('admin_token', res.data.token)
      }
    }
    return res.data
  }

  async function checkAuth() {
    if (!token.value) return false
    try {
      const res = await adminAPI.check()
      if (res.data.success) {
        user.value = { username: res.data.username, password_changed: res.data.password_changed }
        return true
      }
      return false
    } catch {
      return false
    }
  }

  function logout() {
    token.value = ''
    user.value = { username: '', password_changed: true }
    if (typeof window !== 'undefined') {
      localStorage.removeItem('admin_token')
    }
  }

  return {
    token,
    user,
    initialized,
    checkStatus,
    initAdmin,
    login,
    checkAuth,
    logout
  }
}
