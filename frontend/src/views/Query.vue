<template>
  <div class="query-page">
    <div class="query-container">
      <div class="query-header">
        <h1 class="query-title">订单查询</h1>
        <p class="query-desc">输入订单号或联系方式查询您的订单</p>
      </div>
      <a-card class="query-card">
        <a-form :model="queryForm" layout="vertical" class="query-form">
          <a-form-item field="type" label="查询方式">
            <a-radio-group v-model="queryForm.type">
              <a-radio value="order_no">订单号</a-radio>
              <a-radio value="contact">联系方式</a-radio>
            </a-radio-group>
          </a-form-item>
          <a-form-item field="keyword" :label="queryForm.type === 'order_no' ? '订单号' : '联系方式'">
            <a-input
              v-model="queryForm.keyword"
              :placeholder="queryForm.type === 'order_no' ? '请输入订单号' : '请输入手机号或邮箱'"
              size="large"
            />
          </a-form-item>
          <a-button type="primary" size="large" long :loading="loading" @click="handleQuery">
            查询订单
          </a-button>
        </a-form>
      </a-card>
      <a-card v-if="orderList.length > 0" class="result-card" title="查询结果">
        <a-table :data="orderList" :pagination="false">
          <template #columns>
            <a-table-column title="订单号" data-index="order_no" />
            <a-table-column title="商品" data-index="product_name" />
            <a-table-column title="金额" data-index="amount">
              <template #cell="{ record }">
                <span class="price">￥{{ record.amount }}</span>
              </template>
            </a-table-column>
            <a-table-column title="状态" data-index="status">
              <template #cell="{ record }">
                <a-tag :color="statusColor(record.status)">{{ statusText(record.status) }}</a-tag>
              </template>
            </a-table-column>
            <a-table-column title="卡密" data-index="cards">
              <template #cell="{ record }">
                <a v-if="record.cards" @click="copyCards(record.cards)">复制卡密</a>
                <span v-else>-</span>
              </template>
            </a-table-column>
            <a-table-column title="创建时间" data-index="created_at" />
          </template>
        </a-table>
      </a-card>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { Message } from '@arco-design/web-vue'
import { api } from '../api'

const route = useRoute()

const loading = ref(false)
const orderList = ref([])

const queryForm = reactive({
  type: 'order_no',
  keyword: ''
})

const statusColor = (status) => {
  const map = { pending: 'orange', paid: 'green', expired: 'red', cancelled: 'gray' }
  return map[status] || 'gray'
}

const statusText = (status) => {
  const map = { pending: '待支付', paid: '已完成', expired: '已过期', cancelled: '已取消' }
  return map[status] || status
}

const handleQuery = async () => {
  if (!queryForm.keyword) {
    Message.warning('请输入查询内容')
    return
  }
  loading.value = true
  try {
    const res = await api.getOrder(queryForm.type === 'order_no' ? queryForm.keyword : queryForm.keyword)
    if (res.data) {
      orderList.value = Array.isArray(res.data) ? res.data : [res.data]
    } else {
      Message.warning('未找到相关订单')
    }
  } catch (e) {
    Message.error(e.message || '查询失败')
  } finally {
    loading.value = false
  }
}

const copyCards = (cards) => {
  navigator.clipboard.writeText(cards).then(() => {
    Message.success('卡密已复制到剪贴板')
  }).catch(() => {
    Message.error('复制失败，请手动复制')
  })
}

onMounted(() => {
  if (route.query.order_no) {
    queryForm.keyword = route.query.order_no
    queryForm.type = 'order_no'
    handleQuery()
  }
})
</script>

<style scoped>
.query-page {
  min-height: 100vh;
  background: linear-gradient(135deg, #f5f7fa 0%, #e4ecf7 100%);
  padding: 60px 20px;
}

.query-container {
  max-width: 800px;
  margin: 0 auto;
}

.query-header {
  text-align: center;
  margin-bottom: 40px;
}

.query-title {
  font-size: 32px;
  font-weight: 700;
  color: #1d2129;
  margin: 0 0 8px 0;
}

.query-desc {
  font-size: 15px;
  color: #86909c;
  margin: 0;
}

.query-card {
  border-radius: 12px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.06);
  margin-bottom: 24px;
}

.query-form {
  max-width: 500px;
  margin: 0 auto;
}

.result-card {
  border-radius: 12px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.06);
}

.price {
  color: #f53f3f;
  font-weight: 600;
}

@media (max-width: 768px) {
  .query-page {
    padding: 30px 12px;
  }
  
  .query-title {
    font-size: 24px;
  }
  
  .query-form {
    max-width: 100%;
  }
  
  .result-card :deep(.arco-table) {
    font-size: 13px;
  }
}

@media (max-width: 480px) {
  .query-title {
    font-size: 20px;
  }
  
  .query-desc {
    font-size: 13px;
  }
}
</style>
