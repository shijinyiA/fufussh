export default defineNuxtRouteMiddleware(async (to) => {
  if (to.path === '/login' || to.path === '/init') return

  const auth = useAdminAuth()
  const status = await auth.checkStatus()

  if (!status.initialized && to.path !== '/init') {
    return navigateTo('/init')
  }

  if (status.initialized) {
    const isValid = await auth.checkAuth()
    if (!isValid && to.path !== '/login') {
      return navigateTo('/login')
    }
  }
})
