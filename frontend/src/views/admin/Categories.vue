<template>
  <div class="categories-page">
    <a-card class="page-card">
      <div class="page-header">
        <a-input-search
          v-model="keyword"
          placeholder="搜索分类名称"
          class="search-input"
          allow-clear
          @search="handleSearch"
        />
        <a-button type="primary" @click="openModal()">
          <iconify-icon icon="arco:plus" class="btn-icon" />
          新增分类
        </a-button>
      </div>
      <a-table :data="filteredList" :pagination="pagination" :bordered="false" :row-key="'id'">
        <template #columns>
          <a-table-column title="ID" data-index="id" :width="60" />
          <a-table-column title="分类名称" data-index="name" />
          <a-table-column title="图标" data-index="icon" :width="120">
            <template #cell="{ record }">
              <span v-if="record.icon" class="icon-cell">
                <iconify-icon :icon="record.icon" :size="20" />
                <span class="icon-name">{{ record.icon }}</span>
              </span>
              <span v-else>-</span>
            </template>
          </a-table-column>
          <a-table-column title="排序" data-index="sort" :width="80" />
          <a-table-column title="状态" data-index="status" :width="100">
            <template #cell="{ record }">
              <a-tag :color="record.status ? 'green' : 'gray'">
                {{ record.status ? '启用' : '禁用' }}
              </a-tag>
            </template>
          </a-table-column>
          <a-table-column title="创建时间" data-index="created_at" :width="160" />
          <a-table-column title="操作" :width="160" fixed="right">
            <template #cell="{ record }">
              <a-space>
                <a-button type="text" size="small" @click="openModal(record)">编辑</a-button>
                <a-popconfirm content="确定删除该分类？" @ok="handleDelete(record)">
                  <a-button type="text" size="small" status="danger">删除</a-button>
                </a-popconfirm>
              </a-space>
            </template>
          </a-table-column>
        </template>
      </a-table>
    </a-card>

    <a-modal
      v-model:visible="modalVisible"
      :title="editing ? '编辑分类' : '新增分类'"
      @ok="handleSubmit"
      @cancel="modalVisible = false"
      :ok-loading="submitting"
    >
      <a-form :model="formData" layout="vertical">
        <a-form-item field="name" label="分类名称" required>
          <a-input v-model="formData.name" placeholder="请输入分类名称" />
        </a-form-item>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item field="icon" label="图标">
              <a-input v-model="formData.icon" placeholder="arco:apps" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item field="sort" label="排序">
              <a-input-number v-model="formData.sort" :min="0" :max="999" class="full-width" />
            </a-form-item>
          </a-col>
        </a-row>
        <a-form-item field="status" label="状态">
          <a-switch v-model="formData.status" />
          <span class="status-label">{{ formData.status ? '启用' : '禁用' }}</span>
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { Message } from '@arco-design/web-vue'
import { api } from '../../api'

const keyword = ref('')
const modalVisible = ref(false)
const editing = ref(null)
const submitting = ref(false)

const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
  showTotal: true,
  sizeCanChange: true
})

const list = ref([])

const formData = reactive({
  name: '',
  icon: '',
  sort: 0,
  status: true
})

const filteredList = computed(() => {
  let result = list.value
  if (keyword.value) {
    result = result.filter(item => item.name.includes(keyword.value))
  }
  pagination.total = result.length
  return result
})

const loadData = async () => {
  try {
    const res = await api.adminGetCategories()
    if (res.data) {
      list.value = res.data.items || res.data
      pagination.total = res.data.total || list.value.length
    }
  } catch (e) {
    list.value = [
      { id: 1, name: '虚拟商品', icon: 'arco:gift', sort: 1, status: true, created_at: '2024-01-15 10:00' },
      { id: 2, name: '充值服务', icon: 'arco:phone', sort: 2, status: true, created_at: '2024-01-14 15:30' },
      { id: 3, name: '会员卡', icon: 'arco:medal', sort: 3, status: true, created_at: '2024-01-13 09:20' },
      { id: 4, name: '游戏点卡', icon: 'arco:game', sort: 4, status: false, created_at: '2024-01-12 14:10' }
    ]
    pagination.total = list.value.length
  }
}

const handleSearch = () => {
  pagination.current = 1
}

const openModal = (record) => {
  if (record) {
    editing.value = record
    Object.assign(formData, { ...record })
  } else {
    editing.value = null
    Object.assign(formData, {
      name: '', icon: '', sort: 0, status: true
    })
  }
  modalVisible.value = true
}

const handleSubmit = async () => {
  if (!formData.name) {
    Message.warning('请输入分类名称')
    return
  }
  submitting.value = true
  try {
    if (editing.value) {
      await api.adminUpdateCategory(formData)
      Message.success('更新成功')
    } else {
      await api.adminCreateCategory(formData)
      Message.success('创建成功')
    }
    modalVisible.value = false
    loadData()
  } catch (e) {
    Message.error(e.message || '操作失败')
  } finally {
    submitting.value = false
  }
}

const handleDelete = async (record) => {
  try {
    await api.adminDeleteCategory(record.id)
    Message.success('删除成功')
    loadData()
  } catch (e) {
    Message.error(e.message || '删除失败')
  }
}

onMounted(loadData)
</script>

<style scoped>
.categories-page {
  padding: 4px;
}

.page-card {
  border-radius: 8px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  gap: 16px;
}

.search-input {
  width: 300px;
}

.btn-icon {
  margin-right: 4px;
}

.full-width {
  width: 100%;
}

.status-label {
  margin-left: 8px;
  font-size: 13px;
  color: #86909c;
}

.icon-cell {
  display: flex;
  align-items: center;
  gap: 6px;
}

.icon-name {
  font-size: 12px;
  color: #86909c;
}

@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    align-items: stretch;
  }

  .search-input {
    width: 100%;
  }
}
</style>
