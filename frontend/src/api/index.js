import axios from 'axios'
import { useAdminStore } from '../stores'

const request = axios.create({
  baseURL: '',
  timeout: 15000
})

request.interceptors.request.use(
  (config) => {
    const store = useAdminStore()
    if (store.token) {
      config.headers.Authorization = `Bearer ${store.token}`
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
      const store = useAdminStore()
      store.logout()
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export const api = {
  getSiteConfig: () => request.get('/api/site/config'),

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

  getDashboard: () => request.get('/api/admin/system/status'),

  adminGetProducts: (params) => request.get('/api/admin/products', { params }),
  adminCreateProduct: (data) => request.post('/api/admin/products', data),
  adminUpdateProduct: (data) => request.put('/api/admin/products', data),
  adminDeleteProduct: (id) => request.delete(`/api/admin/products/${id}`),

  adminGetCards: (params) => request.get('/api/admin/cards', { params }),
  adminImportCards: (data) => request.post('/api/admin/cards/import', data),
  adminDeleteCard: (id) => request.delete(`/api/admin/cards/${id}`),

  adminGetOrders: (params) => request.get('/api/admin/orders', { params }),

  adminGetSettings: () => request.get('/api/admin/settings'),
  adminUpdateSettings: (data) => request.put('/api/admin/settings', data),

  install: (data) => request.post('/api/install', data),
  getLicenseStatus: () => request.get('/api/install/license-status'),
  verifyLicense: (data) => request.post('/api/install/verify-license', data)
}

export default request
