<template>
  <div class="cards-page">
    <a-card class="page-card">
      <div class="page-header">
        <div class="filter-area">
          <a-select
            v-model="filterProduct"
            placeholder="选择商品"
            class="filter-select"
            allow-clear
          >
            <a-option v-for="p in products" :key="p.id" :value="p.id">{{ p.name }}</a-option>
          </a-select>
          <a-button type="primary" @click="showImportModal = true">
            <iconify-icon icon="arco:upload" class="btn-icon" />
            批量导入
          </a-button>
        </div>
        <a-space>
          <a-tag color="arcoblue">总卡密：{{ totalCards }}</a-tag>
          <a-tag color="green">可用：{{ availableCards }}</a-tag>
          <a-tag color="gray">已使用：{{ usedCards }}</a-tag>
        </a-space>
      </div>
      <a-table :data="filteredList" :pagination="pagination" :bordered="false" :row-key="'id'">
        <template #columns>
          <a-table-column title="ID" data-index="id" :width="60" />
          <a-table-column title="卡密内容" data-index="card_content">
            <template #cell="{ record }">
              <code class="card-content">{{ record.card_content }}</code>
            </template>
          </a-table-column>
          <a-table-column title="所属商品" data-index="product_name" />
          <a-table-column title="状态" data-index="status">
            <template #cell="{ record }">
              <a-tag :color="record.status === 'available' ? 'green' : 'gray'">
                {{ record.status === 'available' ? '未使用' : '已使用' }}
              </a-tag>
            </template>
          </a-table-column>
          <a-table-column title="使用时间" data-index="used_at" :width="160" />
          <a-table-column title="创建时间" data-index="created_at" :width="160" />
          <a-table-column title="操作" :width="100" fixed="right">
            <template #cell="{ record }">
              <a-popconfirm content="确定删除？" @ok="handleDelete(record)">
                <a-button type="text" size="small" status="danger">删除</a-button>
              </a-popconfirm>
            </template>
          </a-table-column>
        </template>
      </a-table>
    </a-card>
    <a-modal
      v-model:visible="showImportModal"
      title="批量导入卡密"
      @ok="handleImport"
      @cancel="showImportModal = false"
      :ok-loading="importing"
    >
      <a-form :model="importForm" layout="vertical">
        <a-form-item field="product_id" label="所属商品" required>
          <a-select v-model="importForm.product_id" placeholder="请选择商品">
            <a-option v-for="p in products" :key="p.id" :value="p.id">{{ p.name }}</a-option>
          </a-select>
        </a-form-item>
        <a-form-item field="cards" label="卡密内容" required>
          <a-textarea
            v-model="importForm.cards"
            placeholder="每行一个卡密，支持批量粘贴"
            :auto-size="{ minRows: 6, maxRows: 12 }"
          />
        </a-form-item>
      </a-form>
      <a-alert type="info" class="import-tip">
        提示：每行输入一个卡密，可直接粘贴文本
      </a-alert>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { Message } from '@arco-design/web-vue'
import { api } from '../../api'

const filterProduct = ref(null)
const showImportModal = ref(false)
const importing = ref(false)

const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
  showTotal: true,
  sizeCanChange: true
})

const products = ref([])
const list = ref([])

const importForm = reactive({
  product_id: null,
  cards: ''
})

const filteredList = computed(() => {
  let result = list.value
  if (filterProduct.value) {
    result = result.filter(item => item.product_id === filterProduct.value)
  }
  pagination.total = result.length
  return result
})

const totalCards = computed(() => list.value.length)
const availableCards = computed(() => list.value.filter(i => i.status === 'available').length)
const usedCards = computed(() => list.value.filter(i => i.status === 'used').length)

const loadData = async () => {
  try {
    const [cardsRes, productsRes] = await Promise.all([
      api.getCards(),
      api.getProducts()
    ])
    if (cardsRes.data) list.value = cardsRes.data
    if (productsRes.data) products.value = productsRes.data
  } catch (e) {
    products.value = [
      { id: 1, name: 'Q币充值' },
      { id: 2, name: '话费充值' },
      { id: 3, name: '视频会员' }
    ]
    list.value = [
      { id: 1, card_content: 'QB100-XXXX-XXXX-0001', product_id: 1, product_name: 'Q币充值', status: 'available', used_at: null, created_at: '2024-01-15 10:00' },
      { id: 2, card_content: 'QB100-XXXX-XXXX-0002', product_id: 1, product_name: 'Q币充值', status: 'used', used_at: '2024-01-15 14:30', created_at: '2024-01-15 10:00' },
      { id: 3, card_content: 'HH100-XXXX-XXXX-0001', product_id: 2, product_name: '话费充值', status: 'available', used_at: null, created_at: '2024-01-14 15:30' },
      { id: 4, card_content: 'SPVIP-XXXX-XXXX-0001', product_id: 3, product_name: '视频会员', status: 'available', used_at: null, created_at: '2024-01-13 09:20' }
    ]
    pagination.total = list.value.length
  }
}

const handleImport = async () => {
  if (!importForm.product_id) {
    Message.warning('请选择商品')
    return
  }
  if (!importForm.cards.trim()) {
    Message.warning('请输入卡密内容')
    return
  }
  importing.value = true
  try {
    await api.importCards({
      product_id: importForm.product_id,
      cards: importForm.cards.split('\n').filter(c => c.trim())
    })
    Message.success('导入成功')
    showImportModal.value = false
    importForm.product_id = null
    importForm.cards = ''
    loadData()
  } catch (e) {
    Message.error(e.message || '导入失败')
  } finally {
    importing.value = false
  }
}

const handleDelete = async (record) => {
  try {
    await api.deleteCard(record.id)
    Message.success('删除成功')
    loadData()
  } catch (e) {
    Message.error(e.message || '删除失败')
  }
}

onMounted(loadData)
</script>

<style scoped>
.cards-page {
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
  flex-wrap: wrap;
  gap: 16px;
}

.filter-area {
  display: flex;
  gap: 12px;
}

.filter-select {
  width: 200px;
}

.btn-icon {
  margin-right: 4px;
}

.card-content {
  background: #f7f8fa;
  padding: 4px 8px;
  border-radius: 4px;
  font-family: 'Courier New', monospace;
  font-size: 13px;
  color: #1d2129;
}

.import-tip {
  margin-top: 12px;
}
</style>
