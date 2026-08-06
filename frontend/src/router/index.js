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
    path: '/install-page',
    component: () => import('../views/Install.vue'),
    name: 'install-page'
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
        meta: { title: '仪表盘' }
      },
      {
        path: 'categories',
        name: 'categories',
        component: () => import('../views/admin/Categories.vue'),
        meta: { title: '商品分类' }
      },
      {
        path: 'products',
        name: 'products',
        component: () => import('../views/admin/Products.vue'),
        meta: { title: '商品管理' }
      },
      {
        path: 'cards',
        name: 'cards',
        component: () => import('../views/admin/Cards.vue'),
        meta: { title: '卡密管理' }
      },
      {
        path: 'orders',
        name: 'orders',
        component: () => import('../views/admin/Orders.vue'),
        meta: { title: '订单管理' }
      },
      {
        path: 'payments',
        name: 'payments',
        component: () => import('../views/admin/PaymentSettings.vue'),
        meta: { title: '支付接口' }
      },
      {
        path: 'emails',
        name: 'emails',
        component: () => import('../views/admin/MailSettings.vue'),
        meta: { title: '邮件系统' }
      },
      {
        path: 'nodes',
        name: 'nodes',
        component: () => import('../views/admin/NodeSettings.vue'),
        meta: { title: '节点管理' }
      },
      {
        path: 'logs',
        name: 'logs',
        component: () => import('../views/admin/Logs.vue'),
        meta: { title: '日志系统' }
      },
      {
        path: 'settings',
        name: 'settings',
        component: () => import('../views/admin/Settings.vue'),
        meta: { title: '系统设置' }
      }
    ]
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
