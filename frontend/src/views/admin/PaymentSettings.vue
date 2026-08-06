<template>
  <div class="payments-page">
    <a-card class="page-card">
      <div class="page-header">
        <a-input-search
          v-model="keyword"
          placeholder="搜索支付通道名称"
          class="search-input"
          allow-clear
          @search="handleSearch"
        />
        <a-button type="primary" @click="openModal()">
          <iconify-icon icon="arco:plus" class="btn-icon" />
          新增支付通道
        </a-button>
      </div>
      <a-table :data="filteredList" :pagination="pagination" :bordered="false" :row-key="'id'">
        <template #columns>
          <a-table-column title="ID" data-index="id" :width="60" />
          <a-table-column title="通道名称" data-index="name" />
          <a-table-column title="类型" data-index="type" :width="120">
            <template #cell="{ record }">
              <a-tag :color="typeColor(record.type)">{{ typeText(record.type) }}</a-tag>
            </template>
          </a-table-column>
          <a-table-column title="费率" data-index="fee_rate" :width="100">
            <template #cell="{ record }">
              <span>{{ record.fee_rate }}%</span>
            </template>
          </a-table-column>
          <a-table-column title="状态" data-index="status" :width="120">
            <template #cell="{ record }">
              <a-switch v-model="record.status" @change="toggleStatus(record)" />
              <span class="status-text">{{ record.status ? '启用' : '禁用' }}</span>
            </template>
          </a-table-column>
          <a-table-column title="操作" :width="160" fixed="right">
            <template #cell="{ record }">
              <a-space>
                <a-button type="text" size="small" @click="openModal(record)">编辑</a-button>
                <a-popconfirm content="确定删除该支付通道？" @ok="handleDelete(record)">
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
      :title="editing ? '编辑支付通道' : '新增支付通道'"
      @ok="handleSubmit"
      @cancel="modalVisible = false"
      :ok-loading="submitting"
    >
      <a-form :model="formData" layout="vertical">
        <a-form-item field="name" label="通道名称" required>
          <a-input v-model="formData.name" placeholder="请输入通道名称" />
        </a-form-item>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item field="type" label="支付类型" required>
              <a-select v-model="formData.type" placeholder="请选择支付类型" class="full-width">
                <a-option value="alipay">支付宝</a-option>
                <a-option value="wechat">微信支付</a-option>
                <a-option value="stripe">Stripe</a-option>
                <a-option value="custom">自定义</a-option>
              </a-select>
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item field="fee_rate" label="费率(%)" required>
              <a-input-number v-model="formData.fee_rate" :min="0" :max="10" :precision="2" class="full-width" />
            </a-form-item>
          </a-col>
        </a-row>
        <a-form-item field="config" label="配置信息">
          <a-textarea
            v-model="formData.config"
            placeholder='请输入JSON配置，如 {"app_id":"xxx","secret":"xxx"}'
            :auto-size="{ minRows: 3, maxRows: 6 }"
          />
        </a-form-item>
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
  type: 'alipay',
  fee_rate: 0,
  config: '',
  status: true
})

const typeColor = (type) => {
  const map = { alipay: 'blue', wechat: 'green', stripe: 'purple', custom: 'gray' }
  return map[type] || 'gray'
}

const typeText = (type) => {
  const map = { alipay: '支付宝', wechat: '微信支付', stripe: 'Stripe', custom: '自定义' }
  return map[type] || type
}

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
    const res = await api.adminGetPayments()
    if (res.data) {
      list.value = res.data.items || res.data
      pagination.total = res.data.total || list.value.length
    }
  } catch (e) {
    list.value = [
      { id: 1, name: '支付宝通道', type: 'alipay', fee_rate: 1.0, config: '{"app_id":"2021000000"}', status: true },
      { id: 2, name: '微信支付通道', type: 'wechat', fee_rate: 0.6, config: '{"mch_id":"1600000000"}', status: true },
      { id: 3, name: 'Stripe通道', type: 'stripe', fee_rate: 2.9, config: '{"api_key":"sk_live_xxx"}', status: false },
      { id: 4, name: '第三方通道', type: 'custom', fee_rate: 1.5, config: '{"api_url":"https://example.com"}', status: true }
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
      name: '', type: 'alipay', fee_rate: 0, config: '', status: true
    })
  }
  modalVisible.value = true
}

const handleSubmit = async () => {
  if (!formData.name) {
    Message.warning('请输入通道名称')
    return
  }
  submitting.value = true
  try {
    if (editing.value) {
      await api.adminUpdatePayment(formData)
      Message.success('更新成功')
    } else {
      await api.adminCreatePayment(formData)
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
    await api.adminDeletePayment(record.id)
    Message.success('删除成功')
    loadData()
  } catch (e) {
    Message.error(e.message || '删除失败')
  }
}

const toggleStatus = async (record) => {
  try {
    await api.adminUpdatePayment({ id: record.id, status: record.status })
    Message.success('状态更新成功')
  } catch (e) {
    Message.error(e.message || '更新失败')
    record.status = !record.status
  }
}

onMounted(loadData)
</script>

<style scoped>
.payments-page {
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

.status-text {
  margin-left: 8px;
  font-size: 13px;
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
