import { defineStore } from 'pinia'

export const useAdminStore = defineStore('admin', {
  state: () => ({
    token: localStorage.getItem('admin_token') || '',
    user: JSON.parse(localStorage.getItem('admin_user') || 'null'),
    siteConfig: JSON.parse(localStorage.getItem('site_config') || '{}')
  }),
  getters: {
    isLoggedIn: (state) => !!state.token,
    siteName: (state) => state.siteConfig.site_name || '晨泽发卡'
  },
  actions: {
    setToken(token) {
      this.token = token
      localStorage.setItem('admin_token', token)
    },
    setUser(user) {
      this.user = user
      localStorage.setItem('admin_user', JSON.stringify(user))
    },
    logout() {
      this.token = ''
      this.user = null
      localStorage.removeItem('admin_token')
      localStorage.removeItem('admin_user')
    },
    setSiteConfig(config) {
      this.siteConfig = config
      localStorage.setItem('site_config', JSON.stringify(config))
    }
  }
})
