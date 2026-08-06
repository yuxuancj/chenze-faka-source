<template>
  <a-layout class="admin-layout">
    <a-layout-sider
      v-model:collapsed="collapsed"
      :collapsible="true"
      :breakpoint="'lg'"
      :collapsed-width="60"
      :width="240"
      class="admin-sider"
    >
      <div class="logo">
        <iconify-icon icon="arco:shop" :size="24" />
        <span v-if="!collapsed" class="logo-text">{{ siteName }}</span>
      </div>

      <a-menu
        :collapsed="collapsed"
        :selected-keys="selectedKeys"
        :open-keys="openKeys"
        @open-keys-change="handleOpenChange"
        @menu-item-click="handleMenuItemClick"
        class="admin-menu"
      >
        <a-menu-item key="/admin/dashboard">
          <iconify-icon icon="arco:dashboard" :size="18" />
          <template #title>仪表盘</template>
        </a-menu-item>

        <a-sub-menu key="product">
          <template #title>
            <iconify-icon icon="arco:apps" :size="18" />
            <span>商品与运营</span>
          </template>
          <a-menu-item key="/admin/categories">
            <iconify-icon icon="arco:menu" :size="16" />
            <template #title>商品分类</template>
          </a-menu-item>
          <a-menu-item key="/admin/products">
            <iconify-icon icon="arco:apps" :size="16" />
            <template #title>商品管理</template>
          </a-menu-item>
        </a-sub-menu>

        <a-sub-menu key="order">
          <template #title>
            <iconify-icon icon="arco:file" :size="18" />
            <span>卡密与订单</span>
          </template>
          <a-menu-item key="/admin/cards">
            <iconify-icon icon="arco:file" :size="16" />
            <template #title>卡密管理</template>
          </a-menu-item>
          <a-menu-item key="/admin/orders">
            <iconify-icon icon="arco:icon-list" :size="16" />
            <template #title>订单管理</template>
          </a-menu-item>
        </a-sub-menu>

        <a-sub-menu key="pay">
          <template #title>
            <iconify-icon icon="arco:pay-circle" :size="18" />
            <span>支付与通知</span>
          </template>
          <a-menu-item key="/admin/payments">
            <iconify-icon icon="arco:pay-circle" :size="16" />
            <template #title>支付接口</template>
          </a-menu-item>
          <a-menu-item key="/admin/emails">
            <iconify-icon icon="arco:mail" :size="16" />
            <template #title>邮件系统</template>
          </a-menu-item>
          <a-menu-item key="/admin/nodes">
            <iconify-icon icon="arco:cloud" :size="16" />
            <template #title>节点管理</template>
          </a-menu-item>
        </a-sub-menu>

        <a-menu-item key="/admin/logs">
          <iconify-icon icon="arco:file-one" :size="18" />
          <template #title>日志系统</template>
        </a-menu-item>

        <a-menu-item key="/admin/settings">
          <iconify-icon icon="arco:settings" :size="18" />
          <template #title>系统设置</template>
        </a-menu-item>
      </a-menu>
    </a-layout-sider>

    <a-layout>
      <a-layout-header class="admin-header">
        <div class="header-left">
          <span class="header-title">{{ currentTitle }}</span>
        </div>
        <div class="header-right">
          <a-dropdown>
            <div class="user-info">
              <div class="user-avatar">{{ userName.charAt(0) }}</div>
              <span class="user-name">{{ userName }}</span>
            </div>
            <template #content>
              <a-doption @click="handleProfile">
                <iconify-icon icon="arco:user" :size="14" /> 个人设置
              </a-doption>
              <a-doption @click="handleLogout">
                <iconify-icon icon="arco:poweroff" :size="14" /> 退出登录
              </a-doption>
            </template>
          </a-dropdown>
        </div>
      </a-layout-header>

      <a-layout-content class="admin-content">
        <router-view v-slot="{ Component }">
          <component :is="Component" />
        </router-view>
      </a-layout-content>
    </a-layout>
  </a-layout>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

const collapsed = ref(false)
const openKeys = ref([])

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

const selectedKeys = computed(() => [route.path])

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

const handleOpenChange = (keys) => {
  openKeys.value = keys
}

const handleMenuItemClick = (key) => {
  if (!key) return
  let path = ''
  if (Array.isArray(key)) {
    path = key.find(k => typeof k === 'string' && k.startsWith('/')) || ''
  } else if (typeof key === 'string') {
    path = key
  }
  if (path && path !== route.path) {
    router.push(path).then(() => {
      if (window.innerWidth < 992) {
        collapsed.value = true
      }
    }).catch(() => {})
  }
}

const handleProfile = () => {
  router.push('/admin/settings')
}

const handleLogout = () => {
  if (confirm('确定要退出登录吗？')) {
    localStorage.removeItem('admin_token')
    localStorage.removeItem('admin_user')
    router.push('/login')
  }
}

const checkMobile = () => {
  if (window.innerWidth < 992) {
    collapsed.value = true
  }
}

watch(() => route.path, (path) => {
  if (window.innerWidth < 992) {
    collapsed.value = true
  }
  if (['/admin/categories', '/admin/products'].includes(path)) {
    openKeys.value = ['product']
  } else if (['/admin/cards', '/admin/orders'].includes(path)) {
    openKeys.value = ['order']
  } else if (['/admin/payments', '/admin/emails', '/admin/nodes'].includes(path)) {
    openKeys.value = ['pay']
  }
})

onMounted(() => {
  checkMobile()
  window.addEventListener('resize', checkMobile)
  const path = route.path
  if (['/admin/categories', '/admin/products'].includes(path)) {
    openKeys.value = ['product']
  } else if (['/admin/cards', '/admin/orders'].includes(path)) {
    openKeys.value = ['order']
  } else if (['/admin/payments', '/admin/emails', '/admin/nodes'].includes(path)) {
    openKeys.value = ['pay']
  }
})

onUnmounted(() => {
  window.removeEventListener('resize', checkMobile)
})
</script>

<style scoped>
.admin-layout {
  height: 100vh;
}

.admin-sider {
  background: var(--color-bg-2);
  border-right: 1px solid var(--color-border-2);
}

.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 0 20px;
  border-bottom: 1px solid var(--color-border-2);
  overflow: hidden;
  white-space: nowrap;
}

.logo-text {
  font-size: 17px;
  font-weight: 600;
  color: var(--color-text-1);
}

.admin-menu {
  background: transparent;
  border-right: none;
  padding: 8px;
}

.admin-menu :deep(.arco-menu-item),
.admin-menu :deep(.arco-menu-submenu-title) {
  height: 42px;
  line-height: 42px;
  margin: 2px 0;
  border-radius: 6px;
}

.admin-menu :deep(.arco-menu-item:hover),
.admin-menu :deep(.arco-menu-submenu-title:hover) {
  background: var(--color-fill-2);
}

.admin-menu :deep(.arco-menu-selected) {
  background: rgb(var(--primary-1));
  color: rgb(var(--primary-6));
  font-weight: 500;
}

.admin-header {
  height: 60px;
  line-height: 60px;
  background: var(--color-bg-2);
  border-bottom: 1px solid var(--color-border-2);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.header-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--color-text-1);
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 4px 10px;
  border-radius: 6px;
  transition: background 0.15s;
}

.user-info:hover {
  background: var(--color-fill-2);
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
  color: var(--color-text-1);
}

.admin-content {
  background: var(--color-fill-2);
  padding: 0;
}

.admin-content :deep(> div) {
  padding: 20px;
  min-height: 100%;
}

@media (max-width: 992px) {
  .admin-content :deep(> div) {
    padding: 12px;
  }
  .user-name {
    display: none;
  }
}
</style>
