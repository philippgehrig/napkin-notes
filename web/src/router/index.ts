import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('../views/LoginView.vue'),
      meta: { guest: true },
    },
    {
      path: '/register',
      name: 'register',
      component: () => import('../views/RegisterView.vue'),
      meta: { guest: true },
    },
    {
      path: '/',
      name: 'napkin',
      component: () => import('../views/NapkinPageView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/napkin/:id',
      name: 'napkin-edit',
      component: () => import('../views/NapkinPageView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/gallery',
      name: 'gallery',
      component: () => import('../views/GalleryView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/note/:id?',
      name: 'editor',
      redirect: (to) => {
        const id = to.params.id
        return id ? `/napkin/${id}` : '/'
      },
    },
    {
      path: '/trash',
      name: 'trash',
      component: () => import('../views/TrashView.vue'),
      meta: { requiresAuth: true },
    },
  ],
})

router.beforeEach((to) => {
  const token = localStorage.getItem('access_token')

  if (to.meta.requiresAuth && !token) {
    return { name: 'login' }
  }

  if (to.meta.guest && token) {
    return { name: 'napkin' }
  }
})

export default router
