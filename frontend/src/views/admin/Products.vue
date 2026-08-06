<template>
  <div class="products-page">
    <a-card class="page-card">
      <div class="page-header">
        <a-input-search
          v-model="keyword"
          placeholder="搜索商品名称"
          class="search-input"
          allow-clear
          @search="handleSearch"
        />
        <a-button type="primary" @click="openModal()">
          <iconify-icon icon="arco:plus" class="btn-icon" />
          新增商品
        </a-button>
      </div>
      <a-table :data="filteredList" :pagination="pagination" :bordered="false" :row-key="'id'">
        <template #columns>
          <a-table-column title="ID" data-index="id" :width="60" />
          <a-table-column title="商品名称" data-index="name" />
          <a-table-column title="描述" data-index="description" :ellipsis="true" />
          <a-table-column title="价格" data-index="price">
            <template #cell="{ record }">
              <span class="price">￥{{ record.price }}</span>
            </template>
          </a-table-column>
          <a-table-column title="状态" data-index="status">
            <template #cell="{ record }">
              <a-switch v-model="record.status" @change="toggleStatus(record)" />
            </template>
          </a-table-column>
          <a-table-column title="创建时间" data-index="created_at" :width="160" />
          <a-table-column title="操作" :width="160" fixed="right">
            <template #cell="{ record }">
              <a-space>
                <a-button type="text" size="small" @click="openModal(record)">编辑</a-button>
                <a-popconfirm content="确定删除？" @ok="handleDelete(record)">
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
      :title="editing ? '编辑商品' : '新增商品'"
      @ok="handleSubmit"
      @cancel="modalVisible = false"
      :ok-loading="submitting"
    >
      <a-form :model="formData" layout="vertical">
        <a-form-item field="name" label="商品名称" required>
          <a-input v-model="formData.name" placeholder="请输入商品名称" />
        </a-form-item>
        <a-form-item field="description" label="商品描述">
          <a-textarea v-model="formData.description" placeholder="请输入商品描述" :auto-size="{ minRows: 2, maxRows: 4 }" />
        </a-form-item>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item field="price" label="售价" required>
              <a-input-number v-model="formData.price" :min="0" :precision="2" class="full-width" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item field="cost_price" label="成本价">
              <a-input-number v-model="formData.cost_price" :min="0" :precision="2" class="full-width" />
            </a-form-item>
          </a-col>
        </a-row>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item field="icon" label="图标">
              <a-input v-model="formData.icon" placeholder="arco:gift" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item field="color" label="主题色">
              <a-input v-model="formData.color" placeholder="#2a7fff" />
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
  description: '',
  price: 0,
  cost_price: 0,
  icon: 'arco:gift',
  color: '#2a7fff',
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
    const res = await api.getProducts()
    if (res.data) {
      list.value = res.data
    }
  } catch (e) {
    list.value = [
      { id: 1, name: 'Q币充值', description: '100Q币 官方直充', price: 95, cost_price: 90, icon: 'arco:gift', color: '#2a7fff', status: true, created_at: '2024-01-15 10:00' },
      { id: 2, name: '话费充值', description: '100元话费 秒到账', price: 98, cost_price: 95, icon: 'arco:phone', color: '#52c41a', status: true, created_at: '2024-01-14 15:30' },
      { id: 3, name: '视频会员', description: '腾讯视频VIP月卡', price: 22, cost_price: 18, icon: 'arco:play-circle', color: '#fa541c', status: true, created_at: '2024-01-13 09:20' }
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
      name: '', description: '', price: 0, cost_price: 0,
      icon: 'arco:gift', color: '#2a7fff', status: true
    })
  }
  modalVisible.value = true
}

const handleSubmit = async () => {
  if (!formData.name) {
    Message.warning('请输入商品名称')
    return
  }
  submitting.value = true
  try {
    if (editing.value) {
      await api.updateProduct(editing.value.id, formData)
      Message.success('更新成功')
    } else {
      await api.createProduct(formData)
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
    await api.deleteProduct(record.id)
    Message.success('删除成功')
    loadData()
  } catch (e) {
    Message.error(e.message || '删除失败')
  }
}

const toggleStatus = async (record) => {
  try {
    await api.updateProduct(record.id, { status: record.status })
    Message.success('状态更新成功')
  } catch (e) {
    Message.error(e.message || '更新失败')
    record.status = !record.status
  }
}

onMounted(loadData)
</script>

<style scoped>
.products-page {
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

.price {
  color: #f53f3f;
  font-weight: 600;
}

.full-width {
  width: 100%;
}

.status-label {
  margin-left: 8px;
  font-size: 13px;
  color: #86909c;
}
</style>
