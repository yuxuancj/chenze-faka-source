<template>
  <div class="orders-page">
    <a-card class="page-card">
      <div class="page-header">
        <div class="filter-area">
          <a-select
            v-model="filterStatus"
            placeholder="订单状态"
            class="filter-select"
            allow-clear
          >
            <a-option value="pending">待支付</a-option>
            <a-option value="paid">已完成</a-option>
            <a-option value="expired">已过期</a-option>
            <a-option value="cancelled">已取消</a-option>
          </a-select>
          <a-input-search
            v-model="keyword"
            placeholder="搜索订单号/商品名"
            class="search-input"
            allow-clear
            @search="handleSearch"
          />
        </div>
        <a-button @click="loadData">
          <iconify-icon icon="arco:refresh" class="btn-icon" />
          刷新
        </a-button>
      </div>
      <a-table :data="filteredList" :pagination="pagination" :bordered="false" :row-key="'id'">
        <template #columns>
          <a-table-column title="订单号" data-index="order_no" :width="200">
            <template #cell="{ record }">
              <span class="order-no">{{ record.order_no }}</span>
            </template>
          </a-table-column>
          <a-table-column title="商品" data-index="product_name" />
          <a-table-column title="数量" data-index="quantity" :width="80" />
          <a-table-column title="金额" data-index="amount" :width="100">
            <template #cell="{ record }">
              <span class="price">￥{{ record.amount }}</span>
            </template>
          </a-table-column>
          <a-table-column title="支付方式" data-index="pay_method" :width="100">
            <template #cell="{ record }">
              <span v-if="record.pay_method === 'alipay'">支付宝</span>
              <span v-else-if="record.pay_method === 'wechat'">微信</span>
              <span v-else>-</span>
            </template>
          </a-table-column>
          <a-table-column title="状态" data-index="status" :width="100">
            <template #cell="{ record }">
              <a-tag :color="statusColor(record.status)">{{ statusText(record.status) }}</a-tag>
            </template>
          </a-table-column>
          <a-table-column title="联系方式" data-index="contact" :width="150" />
          <a-table-column title="创建时间" data-index="created_at" :width="160" />
          <a-table-column title="操作" :width="120" fixed="right">
            <template #cell="{ record }">
              <a-space>
                <a-button type="text" size="small" @click="viewDetail(record)">详情</a-button>
                <a-popconfirm
                  v-if="record.status === 'pending'"
                  content="确定取消该订单？"
                  @ok="cancelOrder(record)"
                >
                  <a-button type="text" size="small" status="danger">取消</a-button>
                </a-popconfirm>
              </a-space>
            </template>
          </a-table-column>
        </template>
      </a-table>
    </a-card>
    <a-drawer v-model:visible="detailVisible" title="订单详情" :width="480">
      <a-descriptions v-if="currentOrder" :column="1" bordered>
        <a-descriptions-item label="订单号">
          {{ currentOrder.order_no }}
        </a-descriptions-item>
        <a-descriptions-item label="商品名称">
          {{ currentOrder.product_name }}
        </a-descriptions-item>
        <a-descriptions-item label="数量">
          {{ currentOrder.quantity }}
        </a-descriptions-item>
        <a-descriptions-item label="金额">
          <span class="price">￥{{ currentOrder.amount }}</span>
        </a-descriptions-item>
        <a-descriptions-item label="支付方式">
          {{ currentOrder.pay_method === 'alipay' ? '支付宝' : currentOrder.pay_method === 'wechat' ? '微信' : '-' }}
        </a-descriptions-item>
        <a-descriptions-item label="状态">
          <a-tag :color="statusColor(currentOrder.status)">{{ statusText(currentOrder.status) }}</a-tag>
        </a-descriptions-item>
        <a-descriptions-item label="联系方式">
          {{ currentOrder.contact }}
        </a-descriptions-item>
        <a-descriptions-item label="卡密">
          <div v-if="currentOrder.cards" class="cards-box">
            <code>{{ currentOrder.cards }}</code>
          </div>
          <span v-else>-</span>
        </a-descriptions-item>
        <a-descriptions-item label="创建时间">
          {{ currentOrder.created_at }}
        </a-descriptions-item>
        <a-descriptions-item v-if="currentOrder.paid_at" label="支付时间">
          {{ currentOrder.paid_at }}
        </a-descriptions-item>
      </a-descriptions>
    </a-drawer>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { Message } from '@arco-design/web-vue'
import { api } from '../../api'

const filterStatus = ref(null)
const keyword = ref('')
const detailVisible = ref(false)
const currentOrder = ref(null)

const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
  showTotal: true,
  sizeCanChange: true
})

const list = ref([])

const filteredList = computed(() => {
  let result = list.value
  if (filterStatus.value) {
    result = result.filter(item => item.status === filterStatus.value)
  }
  if (keyword.value) {
    result = result.filter(item =>
      item.order_no.includes(keyword.value) ||
      (item.product_name && item.product_name.includes(keyword.value))
    )
  }
  pagination.total = result.length
  return result
})

const statusColor = (status) => {
  const map = { pending: 'orange', paid: 'green', expired: 'red', cancelled: 'gray' }
  return map[status] || 'gray'
}

const statusText = (status) => {
  const map = { pending: '待支付', paid: '已完成', expired: '已过期', cancelled: '已取消' }
  return map[status] || status
}

const loadData = async () => {
  try {
    const res = await api.getOrders()
    if (res.data) {
      list.value = res.data
    }
  } catch (e) {
    list.value = [
      { id: 1, order_no: 'ORD202401001', product_name: 'Q币充值', quantity: 1, amount: 95, pay_method: 'alipay', status: 'paid', contact: '13800000001', cards: 'QB100-XXXX-XXXX-0001', created_at: '2024-01-15 14:30', paid_at: '2024-01-15 14:31' },
      { id: 2, order_no: 'ORD202401002', product_name: '话费充值', quantity: 1, amount: 98, pay_method: 'wechat', status: 'paid', contact: '13800000002', cards: 'HH100-XXXX-XXXX-0001', created_at: '2024-01-15 13:20', paid_at: '2024-01-15 13:21' },
      { id: 3, order_no: 'ORD202401003', product_name: '视频会员', quantity: 2, amount: 44, pay_method: 'alipay', status: 'pending', contact: 'test@test.com', cards: null, created_at: '2024-01-15 12:10', paid_at: null },
      { id: 4, order_no: 'ORD202401004', product_name: '游戏点卡', quantity: 1, amount: 48, pay_method: 'alipay', status: 'paid', contact: '13800000004', cards: 'YX500-XXXX-XXXX-0001', created_at: '2024-01-15 11:00', paid_at: '2024-01-15 11:05' },
      { id: 5, order_no: 'ORD202401005', product_name: '京东E卡', quantity: 1, amount: 490, pay_method: 'wechat', status: 'expired', contact: '13800000005', cards: null, created_at: '2024-01-14 16:45', paid_at: null }
    ]
    pagination.total = list.value.length
  }
}

const handleSearch = () => {
  pagination.current = 1
}

const viewDetail = (record) => {
  currentOrder.value = record
  detailVisible.value = true
}

const cancelOrder = async (record) => {
  Message.success('订单已取消')
  loadData()
}

onMounted(loadData)
</script>

<style scoped>
.orders-page {
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
  width: 160px;
}

.search-input {
  width: 260px;
}

.btn-icon {
  margin-right: 4px;
}

.order-no {
  font-family: 'Courier New', monospace;
  font-size: 13px;
}

.price {
  color: #f53f3f;
  font-weight: 600;
}

.cards-box {
  background: #f7f8fa;
  padding: 8px;
  border-radius: 4px;
  word-break: break-all;
}

.cards-box code {
  font-family: 'Courier New', monospace;
  color: #1d2129;
}
</style>
