import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/Login.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/',
    name: 'Home',
    component: () => import('../views/Home.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/profile',
    name: 'Profile',
    component: () => import('../views/Profile.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/terminal/:id',
    name: 'Terminal',
    component: () => import('../views/Terminal.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/direct/:token',
    name: 'DirectConnect',
    component: () => import('../views/DirectConnect.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/direct/connect',
    name: 'DirectConnectCreate',
    component: () => import('../views/DirectConnectCreate.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/sftp/:token',
    name: 'DirectSFTP',
    component: () => import('../views/FileManager.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/share/:token',
    name: 'SharedConnect',
    component: () => import('../views/SharedConnect.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/'
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach(async (to, from, next) => {
  if (to.meta.requiresAuth !== false) {
    try {
      const { authAPI } = await import('../api')
      const res = await authAPI.checkAuth()
      if (res.data.success) next()
      else next('/login')
    } catch {
      next('/login')
    }
  } else {
    next()
  }
})

export default router
