<template>
  <a-layout class="admin-layout">
    <div
      v-if="isMobile && !collapsed"
      class="mobile-overlay"
      @click="collapsed = true"
    ></div>

    <a-layout-sider
      v-model:collapsed="collapsed"
      :collapsible="true"
      :breakpoint="'lg'"
      :collapsed-width="60"
      :width="230"
      class="admin-sider"
      :class="{ 'sider-mobile': isMobile }"
    >
      <div class="logo">
        <div class="logo-icon">
          <iconify-icon icon="arco:shop" :size="24" />
        </div>
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
            <iconify-icon icon="arco:order" :size="16" />
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
      <app-header
        v-model:collapsed="collapsed"
        :breadcrumb="breadcrumb"
        :site-name="siteName"
      />
      <a-layout-content class="admin-content">
        <div class="content-wrapper">
          <router-view v-slot="{ Component }">
            <transition name="fade" mode="out-in">
              <component :is="Component" />
            </transition>
          </router-view>
        </div>
      </a-layout-content>
    </a-layout>
  </a-layout>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import AppHeader from './AppHeader.vue'

const collapsed = ref(false)
const route = useRoute()
const router = useRouter()
const isMobile = ref(false)

const siteName = computed(() => {
  try {
    const config = JSON.parse(localStorage.getItem('site_config') || '{}')
    return config.site_name || '晨泽发卡'
  } catch {
    return '晨泽发卡'
  }
})

const selectedKeys = computed(() => [route.path])

const openKeys = ref(() => {
  const path = route.path
  const map = {
    '/admin/dashboard': [],
    '/admin/categories': ['product'],
    '/admin/products': ['product'],
    '/admin/cards': ['order'],
    '/admin/orders': ['order'],
    '/admin/payments': ['pay'],
    '/admin/emails': ['pay'],
    '/admin/nodes': ['pay'],
    '/admin/logs': [],
    '/admin/settings': []
  }
  return map[path] || []
})

const breadcrumb = computed(() => {
  return route.matched
    .filter(item => item.meta && item.meta.title)
    .map(item => ({ title: item.meta.title, path: item.path }))
})

const handleOpenChange = (keys) => {
  openKeys.value = keys
}

const handleMenuItemClick = (key) => {
  const path = typeof key === 'string' ? key : (Array.isArray(key) ? key[0] : key)
  if (path && typeof path === 'string' && path.startsWith('/')) {
    if (path !== route.path) {
      router.push(path)
      if (isMobile.value) {
        collapsed.value = true
      }
    }
  }
}

const checkMobile = () => {
  isMobile.value = window.innerWidth < 992
  if (isMobile.value) {
    collapsed.value = true
  }
}

watch(() => route.path, () => {
  if (isMobile.value) {
    collapsed.value = true
  }
})

onMounted(() => {
  checkMobile()
  window.addEventListener('resize', checkMobile)
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
  overflow: hidden;
  transition: width 0.2s;
}

.sider-mobile {
  position: fixed;
  left: 0;
  top: 0;
  bottom: 0;
  z-index: 1100;
  box-shadow: 2px 0 8px rgba(0, 0, 0, 0.15);
}

.mobile-overlay {
  position: fixed;
  left: 0;
  right: 0;
  top: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.45);
  z-index: 1050;
  transition: opacity 0.2s;
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
  font-size: 17px;
  font-weight: 600;
  color: var(--color-text-1);
}

.admin-menu {
  background: transparent;
  border-right: none;
  padding: 8px;
}

.admin-content {
  background: var(--color-fill-2);
  padding: 0;
  overflow: auto;
}

.content-wrapper {
  padding: 20px;
  min-height: 100%;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

@media (max-width: 992px) {
  .admin-layout {
    height: auto;
    min-height: 100vh;
  }
  .content-wrapper {
    padding: 12px;
  }
}

@media (max-width: 768px) {
  .admin-layout {
    height: auto;
    min-height: 100vh;
  }
  .content-wrapper {
    padding: 12px 10px;
  }
}
</style>
