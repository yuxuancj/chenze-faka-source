<template>
  <a-layout-header class="app-header">
    <div class="header-left">
      <a-button
        type="text"
        class="collapse-btn"
        @click="$emit('update:collapsed', !collapsed)"
      >
        <iconify-icon :icon="collapsed ? 'arco:menu-unfold' : 'arco:menu-fold'" :size="20" />
      </a-button>
      <a-breadcrumb class="header-breadcrumb">
        <a-breadcrumb-item v-for="(item, index) in breadcrumb" :key="index">
          {{ item.title }}
        </a-breadcrumb-item>
      </a-breadcrumb>
    </div>
    <div class="header-right">
      <a-input-search
        placeholder="搜索..."
        class="header-search"
        :search-button="true"
      />
      <a-dropdown>
        <div class="user-info">
          <a-avatar :size="32" class="user-avatar">
            <template #icon>
              <iconify-icon icon="arco:user" />
            </template>
          </a-avatar>
          <span class="user-name">{{ userName }}</span>
          <iconify-icon icon="arco:down" :size="12" />
        </div>
        <template #content>
          <a-doption @click="handleProfile">
            <iconify-icon icon="arco:user" class="dropdown-icon" />
            个人中心
          </a-doption>
          <a-doption @click="handleLogout">
            <iconify-icon icon="arco:poweroff" class="dropdown-icon" />
            退出登录
          </a-doption>
        </template>
      </a-dropdown>
    </div>
  </a-layout-header>
</template>

<script setup>
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAdminStore } from '../../stores'
import { Modal } from '@arco-design/web-vue'

defineProps({
  collapsed: Boolean,
  breadcrumb: {
    type: Array,
    default: () => []
  },
  siteName: {
    type: String,
    default: '晨泽发卡'
  }
})

defineEmits(['update:collapsed'])

const router = useRouter()
const store = useAdminStore()

const userName = computed(() => store.user?.username || '管理员')

const handleProfile = () => {
  router.push('/admin/settings')
}

const handleLogout = () => {
  Modal.confirm({
    title: '退出登录',
    content: '确定要退出登录吗？',
    onOk: () => {
      store.logout()
      router.push('/login')
    }
  })
}
</script>

<style scoped>
.app-header {
  height: 64px;
  background: var(--color-bg-2);
  border-bottom: 1px solid var(--color-border-2);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.04);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.collapse-btn {
  font-size: 20px;
  color: var(--color-text-1);
}

.header-breadcrumb {
  font-size: 14px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 20px;
}

.header-search {
  width: 240px;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 4px;
  transition: background 0.2s;
}

.user-info:hover {
  background: var(--color-fill-2);
}

.user-avatar {
  background: var(--color-primary-light-3);
  color: var(--color-primary);
}

.user-name {
  font-size: 14px;
  color: var(--color-text-1);
}

.dropdown-icon {
  margin-right: 8px;
}
</style>
