import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    component: () => import('../views/Home.vue'),
    name: 'home'
  },
  {
    path: '/product/:id',
    component: () => import('../views/ProductDetail.vue'),
    name: 'product-detail'
  },
  {
    path: '/checkout',
    component: () => import('../views/Checkout.vue'),
    name: 'checkout'
  },
  {
    path: '/query',
    component: () => import('../views/Query.vue'),
    name: 'query'
  },
  {
    path: '/login',
    component: () => import('../views/Login.vue'),
    name: 'login'
  },
  {
    path: '/install',
    component: () => import('../views/Install.vue'),
    name: 'install'
  },
  {
    path: '/admin',
    component: () => import('../components/layout/AdminLayout.vue'),
    redirect: '/admin/dashboard',
    meta: { requiresAuth: true },
    children: [
      {
        path: 'dashboard',
        name: 'dashboard',
        component: () => import('../views/admin/Dashboard.vue'),
        meta: { title: '仪表盘', icon: 'arco:dashboard' }
      },
      {
        path: 'products',
        name: 'products',
        component: () => import('../views/admin/Products.vue'),
        meta: { title: '商品管理', icon: 'arco:apps' }
      },
      {
        path: 'cards',
        name: 'cards',
        component: () => import('../views/admin/Cards.vue'),
        meta: { title: '卡密管理', icon: 'arco:file' }
      },
      {
        path: 'orders',
        name: 'orders',
        component: () => import('../views/admin/Orders.vue'),
        meta: { title: '订单管理', icon: 'arco:order' }
      },
      {
        path: 'settings',
        name: 'settings',
        component: () => import('../views/admin/Settings.vue'),
        meta: { title: '系统设置', icon: 'arco:settings' }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('admin_token')
  if (to.meta.requiresAuth !== false && to.path.startsWith('/admin')) {
    if (!token) {
      next({ name: 'login' })
    } else {
      next()
    }
  } else {
    next()
  }
})

export default router
