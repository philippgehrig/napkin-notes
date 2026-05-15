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
      name: 'gallery',
      component: () => import('../views/GalleryView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/note/:id?',
      name: 'editor',
      component: () => import('../views/NoteEditorView.vue'),
      meta: { requiresAuth: true },
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
    return { name: 'gallery' }
  }
})

export default router
