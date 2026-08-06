<template>
  <a-layout-header class="admin-header" :bordered="true">
    <div class="header-left">
      <a-button
        class="collapse-btn"
        type="text"
        size="large"
        @click="$emit('update:collapsed', !collapsed)"
      >
        <iconify-icon :icon="collapsed ? 'arco:menu-unfold' : 'arco:menu-fold'" :size="20" />
      </a-button>
      <span v-if="siteName" class="header-title">{{ siteName }}</span>
    </div>
    <div class="header-right">
      <a-dropdown>
        <div class="user-info">
          <a-avatar :size="32" style="background-color: #2a7fff">
            {{ userName.charAt(0) }}
          </a-avatar>
          <span class="user-name">{{ userName }}</span>
          <iconify-icon icon="arco:icon-down" :size="12" />
        </div>
        <template #content>
          <a-doption @click="handleProfile">
            <iconify-icon icon="arco:user" :size="14" style="margin-right: 6px" />
            个人设置
          </a-doption>
          <a-doption @click="handleLogout">
            <iconify-icon icon="arco:poweroff" :size="14" style="margin-right: 6px" />
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

const userName = computed(() => {
  try {
    const user = JSON.parse(localStorage.getItem('admin_user') || 'null')
    return user?.username || '管理员'
  } catch {
    return '管理员'
  }
})

const handleProfile = () => {
  router.push('/admin/settings')
}

const handleLogout = () => {
  Modal.confirm({
    title: '退出登录',
    content: '确定要退出登录吗？',
    onOk: () => {
      localStorage.removeItem('admin_token')
      localStorage.removeItem('admin_user')
      router.push('/login')
    }
  })
}
</script>

<style scoped>
.admin-header {
  background: #fff;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  height: 60px;
  line-height: 60px;
  position: sticky;
  top: 0;
  z-index: 100;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.collapse-btn {
  font-size: 20px;
}

.header-title {
  font-size: 16px;
  font-weight: 600;
  color: #1d2129;
}

.header-right {
  display: flex;
  align-items: center;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 6px;
  transition: background 0.2s;
}

.user-info:hover {
  background: #f2f3f5;
}

.user-name {
  font-size: 14px;
  color: #1d2129;
}

@media (max-width: 768px) {
  .admin-header {
    padding: 0 8px;
    height: 52px;
    line-height: 52px;
  }
  
  .header-title {
    font-size: 14px;
  }
  
  .user-name {
    display: none;
  }
  
  .header-right .user-info {
    padding: 4px;
  }
  
  .header-right .user-info span:not(.user-name) {
    display: none;
  }
}
</style>
