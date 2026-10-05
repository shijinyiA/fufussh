import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { serverAPI } from '../api'

export const useServerStore = defineStore('server', () => {
  const servers = ref([])
  const loading = ref(false)
  const currentGroup = ref('')
  const searchQuery = ref('')

  const filteredServers = computed(() => {
    let result = servers.value

    if (currentGroup.value) {
      result = result.filter(s =>
        (s.group || '默认') === currentGroup.value
      )
    }

    if (searchQuery.value) {
      const query = searchQuery.value.toLowerCase()
      result = result.filter(s =>
        s.name.toLowerCase().includes(query) ||
        s.host.toLowerCase().includes(query) ||
        s.username.toLowerCase().includes(query) ||
        (s.tags && s.tags.some(t => t.toLowerCase().includes(query)))
      )
    }

    return result
  })

  const groups = computed(() => {
    const groupMap = {}
    servers.value.forEach(s => {
      const g = s.group || '默认'
      groupMap[g] = (groupMap[g] || 0) + 1
    })
    return Object.entries(groupMap).map(([name, count]) => ({ name, count }))
  })

  async function fetchServers() {
    loading.value = true
    try {
      const res = await serverAPI.list()
      if (res.data.success) {
        servers.value = res.data.servers
      }
    } catch (err) {
      console.error('Failed to fetch servers:', err)
    } finally {
      loading.value = false
    }
  }

  async function createServer(data) {
    const res = await serverAPI.create(data)
    if (res.data.success) {
      await fetchServers()
    }
    return res.data
  }

  async function updateServer(id, data) {
    const res = await serverAPI.update(id, data)
    if (res.data.success) {
      await fetchServers()
    }
    return res.data
  }

  async function deleteServer(id) {
    const res = await serverAPI.delete(id)
    if (res.data.success) {
      await fetchServers()
    }
    return res.data
  }

  async function testConnection(id) {
    const res = await serverAPI.test(id)
    return res.data
  }

  return {
    servers,
    loading,
    currentGroup,
    searchQuery,
    filteredServers,
    groups,
    fetchServers,
    createServer,
    updateServer,
    deleteServer,
    testConnection
  }
})
