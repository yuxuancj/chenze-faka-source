<template>
  <div class="admin-root">
    <!-- Mobile overlay -->
    <div
      v-if="isMobile && !collapsed"
      class="mobile-overlay"
      @click="collapsed = true"
    ></div>

    <!-- Sidebar -->
    <aside
      class="admin-sider"
      :class="{
        'sider-mobile': isMobile,
        'sider-hidden-mobile': isMobile && collapsed,
        'sider-collapsed-desktop': !isMobile && collapsed
      }"
    >
      <div class="logo">
        <div class="logo-icon">
          <iconify-icon icon="arco:shop" :size="22" />
        </div>
        <span v-if="!collapsed" class="logo-text">{{ siteName }}</span>
      </div>

      <nav class="admin-menu">
        <div class="menu-section">
          <div class="menu-item" :class="{ active: route.path === '/admin/dashboard' }" @click="navigate('/admin/dashboard')">
            <iconify-icon icon="arco:dashboard" :size="18" />
            <span v-if="!collapsed">仪表盘</span>
          </div>
        </div>

        <div class="menu-group">
          <div class="menu-group-title" @click="toggleGroup('product')">
            <iconify-icon icon="arco:apps" :size="18" />
            <span v-if="!collapsed" class="group-name">商品与运营</span>
            <iconify-icon v-if="!collapsed" :icon="openGroups.product ? 'arco:icon-up' : 'arco:icon-down'" :size="14" class="arrow" />
          </div>
          <div v-if="openGroups.product && !collapsed" class="sub-menu">
            <div class="menu-item" :class="{ active: route.path === '/admin/categories' }" @click="navigate('/admin/categories')">
              <iconify-icon icon="arco:menu" :size="16" />
              <span>商品分类</span>
            </div>
            <div class="menu-item" :class="{ active: route.path === '/admin/products' }" @click="navigate('/admin/products')">
              <iconify-icon icon="arco:apps" :size="16" />
              <span>商品管理</span>
            </div>
          </div>
        </div>

        <div class="menu-group">
          <div class="menu-group-title" @click="toggleGroup('order')">
            <iconify-icon icon="arco:file" :size="18" />
            <span v-if="!collapsed" class="group-name">卡密与订单</span>
            <iconify-icon v-if="!collapsed" :icon="openGroups.order ? 'arco:icon-up' : 'arco:icon-down'" :size="14" class="arrow" />
          </div>
          <div v-if="openGroups.order && !collapsed" class="sub-menu">
            <div class="menu-item" :class="{ active: route.path === '/admin/cards' }" @click="navigate('/admin/cards')">
              <iconify-icon icon="arco:file" :size="16" />
              <span>卡密管理</span>
            </div>
            <div class="menu-item" :class="{ active: route.path === '/admin/orders' }" @click="navigate('/admin/orders')">
              <iconify-icon icon="arco:order" :size="16" />
              <span>订单管理</span>
            </div>
          </div>
        </div>

        <div class="menu-group">
          <div class="menu-group-title" @click="toggleGroup('pay')">
            <iconify-icon icon="arco:pay-circle" :size="18" />
            <span v-if="!collapsed" class="group-name">支付与通知</span>
            <iconify-icon v-if="!collapsed" :icon="openGroups.pay ? 'arco:icon-up' : 'arco:icon-down'" :size="14" class="arrow" />
          </div>
          <div v-if="openGroups.pay && !collapsed" class="sub-menu">
            <div class="menu-item" :class="{ active: route.path === '/admin/payments' }" @click="navigate('/admin/payments')">
              <iconify-icon icon="arco:pay-circle" :size="16" />
              <span>支付接口</span>
            </div>
            <div class="menu-item" :class="{ active: route.path === '/admin/emails' }" @click="navigate('/admin/emails')">
              <iconify-icon icon="arco:mail" :size="16" />
              <span>邮件系统</span>
            </div>
            <div class="menu-item" :class="{ active: route.path === '/admin/nodes' }" @click="navigate('/admin/nodes')">
              <iconify-icon icon="arco:cloud" :size="16" />
              <span>节点管理</span>
            </div>
          </div>
        </div>

        <div class="menu-section">
          <div class="menu-item" :class="{ active: route.path === '/admin/logs' }" @click="navigate('/admin/logs')">
            <iconify-icon icon="arco:file-one" :size="18" />
            <span v-if="!collapsed">日志系统</span>
          </div>
        </div>

        <div class="menu-section">
          <div class="menu-item" :class="{ active: route.path === '/admin/settings' }" @click="navigate('/admin/settings')">
            <iconify-icon icon="arco:settings" :size="18" />
            <span v-if="!collapsed">系统设置</span>
          </div>
        </div>
      </nav>
    </aside>

    <!-- Main content -->
    <div class="admin-main" :class="{ 'main-full': isDesktopCollapsed }">
      <!-- Header -->
      <header class="admin-header">
        <div class="header-left">
          <button class="toggle-btn" @click="toggleSidebar">
            <iconify-icon :icon="collapsed ? 'arco:menu-unfold' : 'arco:menu-fold'" :size="20" />
          </button>
          <span class="header-title">{{ currentTitle }}</span>
        </div>
        <div class="header-right">
          <div class="user-dropdown" @click="showUserMenu = !showUserMenu">
            <div class="user-avatar">{{ userName.charAt(0) }}</div>
            <span class="user-name">{{ userName }}</span>
            <iconify-icon icon="arco:icon-down" :size="12" />
          </div>
          <div v-if="showUserMenu" class="user-menu">
            <div class="user-menu-item" @click="handleProfile">
              <iconify-icon icon="arco:user" :size="14" />
              <span>个人设置</span>
            </div>
            <div class="user-menu-item logout" @click="handleLogout">
              <iconify-icon icon="arco:poweroff" :size="14" />
              <span>退出登录</span>
            </div>
          </div>
        </div>
      </header>

      <!-- Content -->
      <main class="admin-content">
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </main>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, reactive, onMounted, onUnmounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

const collapsed = ref(false)
const isMobile = ref(false)
const showUserMenu = ref(false)

const openGroups = reactive({
  product: false,
  order: false,
  pay: false
})

const siteName = computed(() => {
  try {
    const config = JSON.parse(localStorage.getItem('site_config') || '{}')
    return config.site_name || '晨泽发卡'
  } catch {
    return '晨泽发卡'
  }
})

const userName = computed(() => {
  try {
    const user = JSON.parse(localStorage.getItem('admin_user') || 'null')
    return user?.username || '管理员'
  } catch {
    return '管理员'
  }
})

const currentTitle = computed(() => {
  const map = {
    '/admin/dashboard': '仪表盘',
    '/admin/categories': '商品分类',
    '/admin/products': '商品管理',
    '/admin/cards': '卡密管理',
    '/admin/orders': '订单管理',
    '/admin/payments': '支付接口',
    '/admin/emails': '邮件系统',
    '/admin/nodes': '节点管理',
    '/admin/logs': '日志系统',
    '/admin/settings': '系统设置'
  }
  return map[route.path] || '管理后台'
})

const isDesktopCollapsed = computed(() => {
  return !isMobile.value && collapsed.value
})

const toggleGroup = (key) => {
  openGroups[key] = !openGroups[key]
}

const toggleSidebar = () => {
  collapsed.value = !collapsed.value
}

const navigate = (path) => {
  if (path !== route.path) {
    router.push(path)
    if (isMobile.value) {
      collapsed.value = true
    }
  }
}

const handleProfile = () => {
  showUserMenu.value = false
  navigate('/admin/settings')
}

const handleLogout = () => {
  showUserMenu.value = false
  if (confirm('确定要退出登录吗？')) {
    localStorage.removeItem('admin_token')
    localStorage.removeItem('admin_user')
    router.push('/login')
  }
}

const checkMobile = () => {
  isMobile.value = window.innerWidth < 992
  if (isMobile.value) {
    collapsed.value = true
  }
}

watch(() => route.path, (path) => {
  if (isMobile.value) {
    collapsed.value = true
  }
  // Auto-open relevant group
  if (['/admin/categories', '/admin/products'].includes('/' + path)) {
    openGroups.product = true
  } else if (['/admin/cards', '/admin/orders'].includes('/' + path)) {
    openGroups.order = true
  } else if (['/admin/payments', '/admin/emails', '/admin/nodes'].includes('/' + path)) {
    openGroups.pay = true
  }
})

onMounted(() => {
  checkMobile()
  window.addEventListener('resize', checkMobile)
  // Auto-open relevant group on mount
  const path = route.path
  if (['/admin/categories', '/admin/products'].includes(path)) {
    openGroups.product = true
  } else if (['/admin/cards', '/admin/orders'].includes(path)) {
    openGroups.order = true
  } else if (['/admin/payments', '/admin/emails', '/admin/nodes'].includes(path)) {
    openGroups.pay = true
  }
})

onUnmounted(() => {
  window.removeEventListener('resize', checkMobile)
})

// Close user menu on click outside
const closeMenu = (e) => {
  if (!e.target.closest('.user-dropdown') && !e.target.closest('.user-menu')) {
    showUserMenu.value = false
  }
}
onMounted(() => {
  document.addEventListener('click', closeMenu)
})
onUnmounted(() => {
  document.removeEventListener('click', closeMenu)
})
</script>

<style scoped>
.admin-root {
  display: flex;
  min-height: 100vh;
  background: #f2f3f5;
}

/* Sidebar */
.admin-sider {
  width: 220px;
  background: #fff;
  border-right: 1px solid #e5e6eb;
  display: flex;
  flex-direction: column;
  transition: width 0.2s, transform 0.2s;
  flex-shrink: 0;
}

.admin-sider:has(.logo) {
  height: 100vh;
  position: sticky;
  top: 0;
}

.admin-sider .logo {
  height: 60px;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 20px;
  border-bottom: 1px solid #e5e6eb;
  overflow: hidden;
  white-space: nowrap;
}

.logo-icon {
  width: 32px;
  height: 32px;
  flex-shrink: 0;
  background: linear-gradient(135deg, #2a7fff, #1a5fcc);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
}

.logo-text {
  font-size: 16px;
  font-weight: 600;
  color: #1d2129;
}

.admin-menu {
  flex: 1;
  padding: 8px;
  overflow-y: auto;
}

.menu-section {
  margin-bottom: 4px;
}

.menu-group {
  margin-bottom: 4px;
}

.menu-group-title {
  display: flex;
  align-items: center;
  gap: 10px;
  height: 40px;
  padding: 0 12px;
  cursor: pointer;
  color: #4e5969;
  border-radius: 6px;
  transition: background 0.15s;
}

.menu-group-title:hover {
  background: #f2f3f5;
  color: #1d2129;
}

.menu-group-title .group-name {
  flex: 1;
  font-size: 14px;
}

.menu-group-title .arrow {
  color: #86909c;
}

.menu-item {
  display: flex;
  align-items: center;
  gap: 10px;
  height: 40px;
  padding: 0 12px;
  cursor: pointer;
  color: #4e5969;
  border-radius: 6px;
  transition: background 0.15s, color 0.15s;
  font-size: 14px;
  white-space: nowrap;
}

.menu-item:hover {
  background: #f2f3f5;
  color: #1d2129;
}

.menu-item.active {
  background: #e8f3ff;
  color: #165dff;
  font-weight: 500;
}

.sub-menu {
  padding-left: 12px;
  margin-top: 2px;
}

/* Collapsed state */
.admin-sider:has(.admin-root .logo) {
  /* This won't work with scoped, using below instead */
}

/* Main */
.admin-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  transition: margin-left 0.2s;
}

.admin-main.main-full {
  /* When sidebar is collapsed on desktop */
}

/* Header */
.admin-header {
  height: 60px;
  background: #fff;
  border-bottom: 1px solid #e5e6eb;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  position: sticky;
  top: 0;
  z-index: 100;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.toggle-btn {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  background: transparent;
  border-radius: 6px;
  cursor: pointer;
  color: #4e5969;
  transition: background 0.15s;
}

.toggle-btn:hover {
  background: #f2f3f5;
  color: #1d2129;
}

.header-title {
  font-size: 16px;
  font-weight: 600;
  color: #1d2129;
}

.header-right {
  position: relative;
}

.user-dropdown {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 4px 10px;
  border-radius: 6px;
  transition: background 0.15s;
}

.user-dropdown:hover {
  background: #f2f3f5;
}

.user-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: linear-gradient(135deg, #2a7fff, #1a5fcc);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  font-weight: 600;
}

.user-name {
  font-size: 14px;
  color: #1d2129;
}

.user-menu {
  position: absolute;
  right: 0;
  top: calc(100% + 8px);
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
  min-width: 160px;
  padding: 4px;
  z-index: 200;
}

.user-menu-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  color: #4e5969;
  transition: background 0.15s;
}

.user-menu-item:hover {
  background: #f2f3f5;
  color: #1d2129;
}

.user-menu-item.logout:hover {
  background: #ffece8;
  color: #f53f3f;
}

/* Content */
.admin-content {
  flex: 1;
  padding: 20px;
  overflow-y: auto;
}

/* Mobile overlay */
.mobile-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
  z-index: 1000;
  transition: opacity 0.2s;
}

/* Animations */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.15s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

/* Desktop collapsed */
.sider-collapsed-desktop {
  width: 60px !important;
}

.sider-collapsed-desktop .logo-text,
.sider-collapsed-desktop .group-name,
.sider-collapsed-desktop .menu-item span,
.sider-collapsed-desktop .menu-group-title .arrow {
  display: none;
}

.sider-collapsed-desktop .menu-item,
.sider-collapsed-desktop .menu-group-title {
  justify-content: center;
  padding: 0;
}

@media (max-width: 991px) {
  .admin-root {
    flex-direction: column;
  }

  .admin-sider {
    position: fixed;
    left: 0;
    top: 0;
    bottom: 0;
    width: 220px;
    z-index: 1100;
    box-shadow: 2px 0 8px rgba(0, 0, 0, 0.15);
    transform: translateX(0);
    transition: transform 0.25s ease;
  }

  .admin-sider.sider-hidden-mobile {
    transform: translateX(-100%);
  }

  .admin-sider .logo {
    height: 56px;
  }

  .admin-content {
    padding: 12px;
  }

  .admin-header {
    height: 52px;
    padding: 0 12px;
  }

  .header-title {
    font-size: 15px;
  }

  .user-name {
    display: none;
  }

  .toggle-btn {
    width: 32px;
    height: 32px;
  }
}

@media (max-width: 480px) {
  .admin-content {
    padding: 10px 8px;
  }
}
</style>
