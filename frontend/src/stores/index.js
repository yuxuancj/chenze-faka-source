import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useAdminStore = defineStore('admin', () => {
  const token = ref(localStorage.getItem('admin_token') || '')
  const user = ref(JSON.parse(localStorage.getItem('admin_user') || 'null'))
  const siteConfig = ref(JSON.parse(localStorage.getItem('site_config') || '{}'))

  const isLoggedIn = computed(() => !!token.value)
  const siteName = computed(() => siteConfig.value.site_name || '晨泽发卡')

  function setToken(val) {
    token.value = val
    localStorage.setItem('admin_token', val)
  }

  function setUser(val) {
    user.value = val
    localStorage.setItem('admin_user', JSON.stringify(val))
  }

  function logout() {
    token.value = ''
    user.value = null
    localStorage.removeItem('admin_token')
    localStorage.removeItem('admin_user')
  }

  function setSiteConfig(config) {
    siteConfig.value = config
    localStorage.setItem('site_config', JSON.stringify(config))
  }

  return {
    token,
    user,
    siteConfig,
    isLoggedIn,
    siteName,
    setToken,
    setUser,
    logout,
    setSiteConfig
  }
})
