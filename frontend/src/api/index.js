import axios from 'axios'

const request = axios.create({
  baseURL: '',
  timeout: 15000
})

request.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('admin_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

request.interceptors.response.use(
  (response) => {
    const res = response.data
    if (res.code !== undefined && res.code !== 0 && res.code !== 200) {
      return Promise.reject(new Error(res.message || '请求失败'))
    }
    return res
  },
  (error) => {
    if (error.response && error.response.status === 401) {
      localStorage.removeItem('admin_token')
      localStorage.removeItem('admin_user')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export const api = {
  getSiteConfig: () => request.get('/api/site/config'),
  getCaptcha: () => request.get('/api/captcha'),

  login: (data) => request.post('/api/auth/login', data),
  register: (data) => request.post('/api/auth/register', data),

  getProducts: (params) => request.get('/api/products', { params }),
  getProductsOnShelf: () => request.get('/api/products/on-shelf'),
  getProductsGrouped: () => request.get('/api/products/on-shelf-grouped'),
  getProduct: (id) => request.get(`/api/products/${id}`),

  createOrder: (data) => request.post('/api/orders', data),
  queryOrders: (params) => request.get('/api/orders/query', { params }),
  getOrder: (orderNo) => request.get(`/api/orders/${orderNo}`),

  getCardCount: (productId) => request.get(`/api/cards/product/${productId}/count`),

  getDashboard: () => request.get('/api/admin/dashboard'),
  getOrderStatusCounts: () => request.get('/api/admin/dashboard/order-status'),

  adminGetCategories: (params) => request.get('/api/admin/categories', { params }),
  adminGetAllCategories: () => request.get('/api/admin/categories/all'),
  adminCreateCategory: (data) => request.post('/api/admin/categories', data),
  adminUpdateCategory: (data) => request.put('/api/admin/categories', data),
  adminDeleteCategory: (id) => request.delete(`/api/admin/categories/${id}`),

  adminGetProducts: (params) => request.get('/api/admin/products', { params }),
  adminCreateProduct: (data) => request.post('/api/admin/products', data),
  adminUpdateProduct: (data) => request.put('/api/admin/products', data),
  adminDeleteProduct: (id) => request.delete(`/api/admin/products/${id}`),

  adminGetCards: (params) => request.get('/api/admin/cards', { params }),
  adminImportCards: (data) => request.post('/api/admin/cards/import', data),
  adminDeleteCard: (id) => request.delete(`/api/admin/cards/${id}`),
  adminExportCards: (params) => request.get('/api/admin/cards/export', { params }),

  adminGetOrders: (params) => request.get('/api/admin/orders', { params }),
  adminGetOrderLogs: (params) => request.get('/api/admin/orders/logs', { params }),

  adminGetPayments: (params) => request.get('/api/admin/payments', { params }),
  adminGetAllPayments: () => request.get('/api/admin/payments/all'),
  adminCreatePayment: (data) => request.post('/api/admin/payments', data),
  adminUpdatePayment: (data) => request.put('/api/admin/payments', data),
  adminDeletePayment: (id) => request.delete(`/api/admin/payments/${id}`),

  adminGetEmails: (params) => request.get('/api/admin/emails', { params }),
  adminCreateEmail: (data) => request.post('/api/admin/emails', data),
  adminUpdateEmail: (data) => request.put('/api/admin/emails', data),
  adminDeleteEmail: (id) => request.delete(`/api/admin/emails/${id}`),
  adminTestEmail: (id) => request.post(`/api/admin/emails/test/${id}`),
  adminGetEmailLogs: (params) => request.get('/api/admin/emails/logs', { params }),

  adminGetOperations: (params) => request.get('/api/admin/logs/operations', { params }),
  adminGetLoginLogs: (params) => request.get('/api/admin/logs/logins', { params }),

  adminGetNodes: (params) => request.get('/api/admin/nodes', { params }),
  adminCreateNode: (data) => request.post('/api/admin/nodes', data),
  adminUpdateNode: (data) => request.put('/api/admin/nodes', data),
  adminDeleteNode: (id) => request.delete(`/api/admin/nodes/${id}`),
  adminPingNode: (id) => request.post(`/api/admin/nodes/ping/${id}`),

  adminGetSettings: () => request.get('/api/admin/settings'),
  adminUpdateSettings: (data) => request.put('/api/admin/settings', data),

  adminGetVersion: () => request.get('/api/admin/upgrade/version'),
  adminCheckUpdate: () => request.get('/api/admin/upgrade/check'),
  adminUploadPackage: (data) => request.post('/api/admin/upgrade/upload', data),
  adminApplyUpgrade: (data) => request.post('/api/admin/upgrade/apply', data),
  adminUpgradeLogs: (params) => request.get('/api/admin/upgrade/logs', { params }),

  adminUploadFile: (data) => request.post('/api/admin/upload', data, {
    headers: { 'Content-Type': 'multipart/form-data' }
  }),
  adminListFiles: (params) => request.get('/api/admin/upload', { params }),
  adminDeleteFile: (id) => request.delete(`/api/admin/upload/${id}`),

  install: (data) => request.post('/api/install', data),
  getLicenseStatus: () => request.get('/api/install/license-status'),
  verifyLicense: (data) => request.post('/api/install/verify-license', data),
  testDatabase: (data) => request.post('/api/install/test-database', data),
  checkEnv: () => request.get('/api/install/env')
}

export default request
