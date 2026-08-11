<template>
  <a-layout class="admin-layout">
    <a-layout-sider
      v-model:collapsed="collapsed"
      :collapsible="true"
      :breakpoint="'lg'"
      :collapsed-width="60"
      :width="240"
      class="admin-sider"
      :class="{ 'sider-open': !collapsed, 'sider-closed-mobile': isMobile && collapsed }"
    >
      <div class="logo">
        <IconApps />
        <span v-if="!collapsed" class="logo-text">{{ siteName }}</span>
      </div>

      <a-menu
        :collapsed="collapsed"
        :selected-keys="selectedKeys"
        :open-keys="openKeys"
        @open-keys-change="handleOpenChange"
        class="admin-menu"
      >
        <a-menu-item key="/admin/dashboard" @click="navigate('/admin/dashboard')">
          <IconDashboard />
          <template #title>仪表盘</template>
        </a-menu-item>

        <a-sub-menu key="product">
          <template #title>
            <IconApps />
            <span>商品与运营</span>
          </template>
          <a-menu-item key="/admin/categories" @click="navigate('/admin/categories')">
            <IconMenu />
            <template #title>商品分类</template>
          </a-menu-item>
          <a-menu-item key="/admin/products" @click="navigate('/admin/products')">
            <IconApps />
            <template #title>商品管理</template>
          </a-menu-item>
        </a-sub-menu>

        <a-sub-menu key="order">
          <template #title>
            <IconFile />
            <span>卡密与订单</span>
          </template>
          <a-menu-item key="/admin/cards" @click="navigate('/admin/cards')">
            <IconFile />
            <template #title>卡密管理</template>
          </a-menu-item>
          <a-menu-item key="/admin/orders" @click="navigate('/admin/orders')">
            <IconList />
            <template #title>订单管理</template>
          </a-menu-item>
        </a-sub-menu>

        <a-sub-menu key="pay">
          <template #title>
            <IconAlipayCircle />
            <span>支付与通知</span>
          </template>
          <a-menu-item key="/admin/payments" @click="navigate('/admin/payments')">
            <IconAlipayCircle />
            <template #title>支付接口</template>
          </a-menu-item>
          <a-menu-item key="/admin/emails" @click="navigate('/admin/emails')">
            <IconEmail />
            <template #title>邮件系统</template>
          </a-menu-item>
          <a-menu-item key="/admin/nodes" @click="navigate('/admin/nodes')">
            <IconCloud />
            <template #title>节点管理</template>
          </a-menu-item>
        </a-sub-menu>

        <a-menu-item key="/admin/logs" @click="navigate('/admin/logs')">
          <IconHistory />
          <template #title>日志系统</template>
        </a-menu-item>

        <a-menu-item key="/admin/settings" @click="navigate('/admin/settings')">
          <IconSettings />
          <template #title>系统设置</template>
        </a-menu-item>
      </a-menu>
    </a-layout-sider>

    <div v-if="isMobile && !collapsed" class="sider-mask" @click="collapsed = true"></div>

    <a-layout>
      <a-layout-header class="admin-header">
        <div class="header-left">
          <a-button
            v-if="isMobile"
            type="text"
            size="small"
            class="menu-toggle-btn"
            @click="collapsed = !collapsed"
          >
            <IconMenuUnfold v-if="collapsed" />
            <IconMenuFold v-else />
          </a-button>
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
                <IconUser /> 个人设置
              </a-doption>
              <a-doption @click="handleLogout">
                <IconPoweroff /> 退出登录
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
import IconDashboard from '@arco-design/web-vue/es/icon/icon-dashboard/index.js'
import IconApps from '@arco-design/web-vue/es/icon/icon-apps/index.js'
import IconMenu from '@arco-design/web-vue/es/icon/icon-menu/index.js'
import IconFile from '@arco-design/web-vue/es/icon/icon-file/index.js'
import IconList from '@arco-design/web-vue/es/icon/icon-list/index.js'
import IconAlipayCircle from '@arco-design/web-vue/es/icon/icon-alipay-circle/index.js'
import IconEmail from '@arco-design/web-vue/es/icon/icon-email/index.js'
import IconCloud from '@arco-design/web-vue/es/icon/icon-cloud/index.js'
import IconHistory from '@arco-design/web-vue/es/icon/icon-history/index.js'
import IconSettings from '@arco-design/web-vue/es/icon/icon-settings/index.js'
import IconUser from '@arco-design/web-vue/es/icon/icon-user/index.js'
import IconPoweroff from '@arco-design/web-vue/es/icon/icon-poweroff/index.js'
import IconMenuFold from '@arco-design/web-vue/es/icon/icon-menu-fold/index.js'
import IconMenuUnfold from '@arco-design/web-vue/es/icon/icon-menu-unfold/index.js'

const route = useRoute()
const router = useRouter()

const collapsed = ref(false)
const openKeys = ref([])
const isMobile = ref(false)

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

const navigate = (path) => {
  if (!path) return
  if (path !== route.path) {
    router.push(path).then(() => {
      if (isMobile.value) {
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
  isMobile.value = window.innerWidth < 992
  if (isMobile.value) {
    collapsed.value = true
  }
}

watch(() => route.path, (path) => {
  if (isMobile.value) {
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
  transition: transform 0.2s ease;
  z-index: 100;
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

.menu-toggle-btn {
  display: none;
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
  overflow: auto;
}

.admin-content :deep(> div) {
  padding: 20px;
  min-height: 100%;
}

.sider-mask {
  display: none;
}

@media (max-width: 992px) {
  .menu-toggle-btn {
    display: flex;
  }

  .user-name {
    display: none;
  }

  .admin-sider {
    position: fixed !important;
    left: 0;
    top: 0;
    bottom: 0;
    z-index: 200;
    box-shadow: 2px 0 8px rgba(0, 0, 0, 0.1);
  }

  .admin-sider.sider-closed-mobile {
    transform: translateX(-100%);
  }

  .admin-sider.sider-open {
    transform: translateX(0);
  }

  .sider-mask {
    display: block;
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.45);
    z-index: 150;
    transition: opacity 0.2s;
  }

  .admin-content :deep(> div) {
    padding: 12px;
  }

  .admin-header {
    padding: 0 12px;
  }
}
</style>
