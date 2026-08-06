<template>
  <a-config-provider>
    <a-layout class="admin-layout">
      <a-layout-sider
        v-model:collapsed="collapsed"
        :collapsible="true"
        :breakpoint="'xl'"
        :collapsed-width="60"
        :width="220"
        class="admin-sider"
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
          @select="handleSelect"
          class="admin-menu"
        >
          <a-menu-item v-for="item in menuItems" :key="item.path">
            <iconify-icon :icon="item.icon" :size="18" />
            <template #title>{{ item.title }}</template>
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
  </a-config-provider>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAdminStore } from '../../stores'
import AppHeader from './AppHeader.vue'

const collapsed = ref(false)
const route = useRoute()
const router = useRouter()
const store = useAdminStore()

const siteName = computed(() => store.siteName)

const menuItems = [
  { path: '/admin/dashboard', title: '仪表盘', icon: 'arco:dashboard' },
  { path: '/admin/products', title: '商品管理', icon: 'arco:apps' },
  { path: '/admin/cards', title: '卡密管理', icon: 'arco:file' },
  { path: '/admin/orders', title: '订单管理', icon: 'arco:order' },
  { path: '/admin/settings', title: '系统设置', icon: 'arco:settings' }
]

const selectedKeys = computed(() => [route.path])
const openKeys = ref([route.path])

const breadcrumb = computed(() => {
  const matched = route.matched.filter(item => item.meta && item.meta.title)
  return matched.map(item => ({
    title: item.meta.title,
    path: item.path
  }))
})

const handleOpenChange = (keys) => {
  openKeys.value = keys
}

const handleSelect = (key) => {
  router.push(key)
}
</script>

<style scoped>
.admin-layout {
  height: 100vh;
}

.admin-sider {
  background: var(--color-bg-2);
  border-right: 1px solid var(--color-border-2);
  overflow: hidden;
}

.logo {
  height: 64px;
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
  font-size: 18px;
  font-weight: 600;
  color: var(--color-text-1);
}

.admin-menu {
  background: transparent;
  border-right: none;
  padding: 12px 8px;
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
</style>
