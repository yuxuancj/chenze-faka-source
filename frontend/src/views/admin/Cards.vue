<template>
  <div class="cards-page">
    <a-tabs v-model:active-key="activeTab" type="line">
      <a-tab-pane key="manage" title="卡密管理">
        <div class="tab-content">
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
                <a-button @click="exportAvailableCards">
                  <iconify-icon icon="arco:download" class="btn-icon" />
                  导出可用卡密
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
        </div>
      </a-tab-pane>

      <a-tab-pane key="issue" title="发卡中心">
        <div class="tab-content">
          <a-row :gutter="16">
            <a-col :xs="24" :md="12">
              <a-card class="page-card">
                <template #title>
                  <span class="card-title">发卡操作</span>
                </template>
                <a-form :model="issueForm" layout="vertical" ref="issueFormRef">
                  <a-form-item field="product_id" label="选择商品" required>
                    <a-select v-model="issueForm.product_id" placeholder="请选择商品" @change="updateAvailableCount">
                      <a-option v-for="p in products" :key="p.id" :value="p.id">
                        {{ p.name }} (可用: {{ getAvailableCount(p.id) }})
                      </a-option>
                    </a-select>
                  </a-form-item>
                  <a-form-item field="quantity" label="发卡数量" required>
                    <a-input-number
                      v-model="issueForm.quantity"
                      :min="1"
                      :max="maxIssueCount"
                      style="width: 100%"
                    />
                  </a-form-item>
                  <a-form-item field="customer" label="客户/备注">
                    <a-input v-model="issueForm.customer" placeholder="填写客户名称或备注信息" />
                  </a-form-item>
                  <a-form-item field="order_no" label="订单号（可选）">
                    <a-input v-model="issueForm.order_no" placeholder="关联订单号" />
                  </a-form-item>
                  <a-form-item>
                    <a-space>
                      <a-button type="primary" :loading="issuing" @click="handleIssue">
                        <iconify-icon icon="arco:send" class="btn-icon" />
                        确认发卡
                      </a-button>
                      <a-button @click="resetIssueForm">重置</a-button>
                    </a-space>
                  </a-form-item>
                </a-form>
                <a-alert type="warning" class="issue-tip">
                  提示：发卡后卡密状态将变为"已使用"，请谨慎操作
                </a-alert>
              </a-card>
            </a-col>
            <a-col :xs="24" :md="12">
              <a-card class="page-card">
                <template #title>
                  <span class="card-title">发卡结果</span>
                </template>
                <div v-if="issuedCards.length > 0" class="issue-result">
                  <div class="result-info">
                    <a-tag color="green">成功发放：{{ issuedCards.length }} 张卡密</a-tag>
                    <a-tag v-if="issueForm.product_id" color="arcoblue">
                      商品：{{ getProductName(issueForm.product_id) }}
                    </a-tag>
                  </div>
                  <div class="card-list">
                    <div v-for="(card, index) in issuedCards" :key="index" class="card-item">
                      <span class="card-index">#{{ index + 1 }}</span>
                      <code class="card-content">{{ card.card_content }}</code>
                      <a-button type="text" size="mini" @click="copyCard(card.card_content)">复制</a-button>
                    </div>
                  </div>
                  <a-space class="result-actions">
                    <a-button type="primary" @click="copyAllCards">复制全部</a-button>
                    <a-button @click="clearResult">清空结果</a-button>
                  </a-space>
                </div>
                <a-empty v-else description="暂无发卡记录" />
              </a-card>
            </a-col>
          </a-row>
        </div>
      </a-tab-pane>

      <a-tab-pane key="records" title="发卡记录">
        <div class="tab-content">
          <a-card class="page-card">
            <div class="page-header">
              <div class="filter-area">
                <a-select
                  v-model="recordFilter.product_id"
                  placeholder="商品筛选"
                  class="filter-select"
                  allow-clear
                >
                  <a-option v-for="p in products" :key="p.id" :value="p.id">{{ p.name }}</a-option>
                </a-select>
                <a-select
                  v-model="recordFilter.status"
                  placeholder="状态筛选"
                  class="filter-select-sm"
                  allow-clear
                >
                  <a-option value="used">已发放</a-option>
                  <a-option value="available">未发放</a-option>
                </a-select>
                <a-button @click="exportRecords">
                  <iconify-icon icon="arco:download" class="btn-icon" />
                  导出记录
                </a-button>
              </div>
              <a-space>
                <a-tag color="arcoblue">记录总数：{{ totalRecords }}</a-tag>
              </a-space>
            </div>
            <a-table :data="filteredRecords" :pagination="recordPagination" :bordered="false" :row-key="'id'">
              <template #columns>
                <a-table-column title="ID" data-index="id" :width="60" />
                <a-table-column title="卡密" data-index="card_content">
                  <template #cell="{ record }">
                    <code class="card-content">{{ record.card_content }}</code>
                  </template>
                </a-table-column>
                <a-table-column title="商品" data-index="product_name" />
                <a-table-column title="状态" data-index="status">
                  <template #cell="{ record }">
                    <a-tag :color="record.status === 'used' ? 'orange' : 'green'">
                      {{ record.status === 'used' ? '已发放' : '未发放' }}
                    </a-tag>
                  </template>
                </a-table-column>
                <a-table-column title="客户/备注" data-index="customer" />
                <a-table-column title="发放时间" data-index="used_at" :width="160" />
                <a-table-column title="创建时间" data-index="created_at" :width="160" />
              </template>
            </a-table>
          </a-card>
        </div>
      </a-tab-pane>
    </a-tabs>

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

const activeTab = ref('manage')
const filterProduct = ref(null)
const showImportModal = ref(false)
const importing = ref(false)
const issuing = ref(false)
const issueFormRef = ref(null)

const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
  showTotal: true,
  sizeCanChange: true
})

const recordPagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
  showTotal: true,
  sizeCanChange: true
})

const products = ref([])
const list = ref([])
const issuedCards = ref([])

const importForm = reactive({
  product_id: null,
  cards: ''
})

const issueForm = reactive({
  product_id: null,
  quantity: 1,
  customer: '',
  order_no: ''
})

const recordFilter = reactive({
  product_id: null,
  status: null
})

const filteredList = computed(() => {
  let result = list.value
  if (filterProduct.value) {
    result = result.filter(item => item.product_id === filterProduct.value)
  }
  pagination.total = result.length
  return result
})

const filteredRecords = computed(() => {
  let result = list.value
  if (recordFilter.value.product_id) {
    result = result.filter(item => item.product_id === recordFilter.value.product_id)
  }
  if (recordFilter.value.status) {
    result = result.filter(item => item.status === recordFilter.value.status)
  }
  recordPagination.total = result.length
  return result
})

const totalCards = computed(() => list.value.length)
const availableCards = computed(() => list.value.filter(i => i.status === 'available').length)
const usedCards = computed(() => list.value.filter(i => i.status === 'used').length)
const totalRecords = computed(() => list.value.length)

const maxIssueCount = computed(() => {
  if (!issueForm.product_id) return 100
  return getAvailableCount(issueForm.product_id) || 1
})

const getAvailableCount = (productId) => {
  return list.value.filter(i => i.product_id === productId && i.status === 'available').length
}

const getProductName = (productId) => {
  const product = products.value.find(p => p.id === productId)
  return product ? product.name : '-'
}

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
      { id: 1, card_content: 'QB100-XXXX-XXXX-0001', product_id: 1, product_name: 'Q币充值', status: 'available', customer: '', order_no: '', used_at: null, created_at: '2024-01-15 10:00' },
      { id: 2, card_content: 'QB100-XXXX-XXXX-0002', product_id: 1, product_name: 'Q币充值', status: 'used', customer: '张三', order_no: 'ORD001', used_at: '2024-01-15 14:30', created_at: '2024-01-15 10:00' },
      { id: 3, card_content: 'HH100-XXXX-XXXX-0001', product_id: 2, product_name: '话费充值', status: 'available', customer: '', order_no: '', used_at: null, created_at: '2024-01-14 15:30' },
      { id: 4, card_content: 'SPVIP-XXXX-XXXX-0001', product_id: 3, product_name: '视频会员', status: 'available', customer: '', order_no: '', used_at: null, created_at: '2024-01-13 09:20' }
    ]
    pagination.total = list.value.length
    recordPagination.total = list.value.length
  }
}

const updateAvailableCount = () => {
  issueForm.quantity = 1
}

const handleIssue = async () => {
  if (!issueForm.product_id) {
    Message.warning('请选择商品')
    return
  }
  if (!issueForm.quantity || issueForm.quantity < 1) {
    Message.warning('请输入发卡数量')
    return
  }
  
  const available = getAvailableCount(issueForm.product_id)
  if (issueForm.quantity > available) {
    Message.warning(`可用卡密不足，当前只有 ${available} 张`)
    return
  }
  
  issuing.value = true
  try {
    // 获取可用卡密
    const availableCardsList = list.value
      .filter(i => i.product_id === issueForm.product_id && i.status === 'available')
      .slice(0, issueForm.quantity)
    
    // 模拟发卡（实际应调用后端API）
    issuedCards.value = availableCardsList.map(card => ({
      ...card,
      customer: issueForm.customer,
      order_no: issueForm.order_no
    }))
    
    Message.success(`成功发放 ${issuedCards.value.length} 张卡密`)
    
    // 更新卡密状态
    availableCardsList.forEach(card => {
      const target = list.value.find(i => i.id === card.id)
      if (target) {
        target.status = 'used'
        target.used_at = new Date().toLocaleString('zh-CN', { hour12: false })
        target.customer = issueForm.customer
        target.order_no = issueForm.order_no
      }
    })
    
    pagination.total = list.value.length
    recordPagination.total = list.value.length
    
    // 重置表单
    issueForm.quantity = 1
  } catch (e) {
    Message.error(e.message || '发卡失败')
  } finally {
    issuing.value = false
  }
}

const resetIssueForm = () => {
  issueForm.product_id = null
  issueForm.quantity = 1
  issueForm.customer = ''
  issueForm.order_no = ''
}

const copyCard = (content) => {
  navigator.clipboard.writeText(content).then(() => {
    Message.success('已复制到剪贴板')
  }).catch(() => {
    Message.error('复制失败')
  })
}

const copyAllCards = () => {
  const text = issuedCards.value.map(c => c.card_content).join('\n')
  navigator.clipboard.writeText(text).then(() => {
    Message.success(`已复制 ${issuedCards.value.length} 个卡密`)
  }).catch(() => {
    Message.error('复制失败')
  })
}

const clearResult = () => {
  issuedCards.value = []
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

const exportAvailableCards = () => {
  const available = list.value.filter(i => i.status === 'available')
  if (available.length === 0) {
    Message.warning('没有可用的卡密')
    return
  }
  const text = available.map(c => c.card_content).join('\n')
  const blob = new Blob([text], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `available_cards_${new Date().toISOString().slice(0, 10)}.txt`
  a.click()
  URL.revokeObjectURL(url)
  Message.success('导出成功')
}

const exportRecords = () => {
  const records = filteredRecords.value
  if (records.length === 0) {
    Message.warning('没有可导出的记录')
    return
  }
  
  let csv = '\uFEFF' + 'ID,卡密,商品,状态,客户/备注,发放时间,创建时间\n'
  records.forEach(r => {
    csv += `${r.id},${r.card_content},${r.product_name},${r.status === 'used' ? '已发放' : '未发放'},${r.customer || '-'},${r.used_at || '-'},${r.created_at}\n`
  })
  
  const blob = new Blob([csv], { type: 'text/csv' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `card_records_${new Date().toISOString().slice(0, 10)}.csv`
  a.click()
  URL.revokeObjectURL(url)
  Message.success('导出成功')
}

onMounted(loadData)
</script>

<style scoped>
.cards-page {
  padding: 4px;
}

.tab-content {
  padding: 16px 0;
}

.page-card {
  border-radius: 8px;
}

.card-title {
  font-size: 16px;
  font-weight: 600;
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
  flex-wrap: wrap;
}

.filter-select {
  width: 200px;
}

.filter-select-sm {
  width: 150px;
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

.issue-tip {
  margin-top: 16px;
}

.issue-result {
  margin-top: 8px;
}

.result-info {
  margin-bottom: 16px;
  display: flex;
  gap: 8px;
}

.card-list {
  max-height: 300px;
  overflow-y: auto;
  border: 1px solid #e5e6eb;
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 16px;
}

.card-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 0;
  border-bottom: 1px solid #f2f3f5;
}

.card-item:last-child {
  border-bottom: none;
}

.card-index {
  color: #86909c;
  font-size: 12px;
  min-width: 30px;
}

.card-item .card-content {
  flex: 1;
}

.result-actions {
  justify-content: flex-end;
}

@media (max-width: 768px) {
  .filter-select,
  .filter-select-sm {
    width: 100%;
  }
  
  .filter-area {
    width: 100%;
  }
}
</style>
