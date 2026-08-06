<template>
  <div class="dashboard-page">
    <a-row :gutter="[16, 16]" class="stat-row">
      <a-col :xs="12" :sm="6">
        <a-card class="stat-card" class-stat="blue">
          <div class="stat-content">
            <div class="stat-info">
              <p class="stat-label">总商品数</p>
              <h2 class="stat-value">{{ stats.products }}</h2>
            </div>
            <div class="stat-icon blue">
              <iconify-icon icon="arco:apps" :size="28" />
            </div>
          </div>
          <p class="stat-trend">
            <span class="trend-up">
              <iconify-icon icon="arco:rise" /> 12%
            </span>
            较上月
          </p>
        </a-card>
      </a-col>
      <a-col :xs="12" :sm="6">
        <a-card class="stat-card">
          <div class="stat-content">
            <div class="stat-info">
              <p class="stat-label">总订单数</p>
              <h2 class="stat-value">{{ stats.orders }}</h2>
            </div>
            <div class="stat-icon green">
              <iconify-icon icon="arco:order" :size="28" />
            </div>
          </div>
          <p class="stat-trend">
            <span class="trend-up">
              <iconify-icon icon="arco:rise" /> 8%
            </span>
            较上月
          </p>
        </a-card>
      </a-col>
      <a-col :xs="12" :sm="6">
        <a-card class="stat-card">
          <div class="stat-content">
            <div class="stat-info">
              <p class="stat-label">总用户数</p>
              <h2 class="stat-value">{{ stats.users }}</h2>
            </div>
            <div class="stat-icon orange">
              <iconify-icon icon="arco:user-group" :size="28" />
            </div>
          </div>
          <p class="stat-trend">
            <span class="trend-up">
              <iconify-icon icon="arco:rise" /> 5%
            </span>
            较上月
          </p>
        </a-card>
      </a-col>
      <a-col :xs="12" :sm="6">
        <a-card class="stat-card">
          <div class="stat-content">
            <div class="stat-info">
              <p class="stat-label">总收入</p>
              <h2 class="stat-value">￥{{ stats.revenue }}</h2>
            </div>
            <div class="stat-icon purple">
              <iconify-icon icon="arco:money" :size="28" />
            </div>
          </div>
          <p class="stat-trend">
            <span class="trend-up">
              <iconify-icon icon="arco:rise" /> 15%
            </span>
            较上月
          </p>
        </a-card>
      </a-col>
    </a-row>
    <a-row :gutter="[16, 16]" class="content-row">
      <a-col :xs="24" :lg="16">
        <a-card title="最近订单" class="content-card">
          <a-table :data="recentOrders" :pagination="false" :bordered="false">
            <template #columns>
              <a-table-column title="订单号" data-index="order_no" :width="180" />
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
              <a-table-column title="时间" data-index="created_at" :width="160" />
            </template>
          </a-table>
        </a-card>
      </a-col>
      <a-col :xs="24" :lg="8">
        <a-card title="热销商品" class="content-card">
          <div class="hot-products">
            <div v-for="(item, index) in hotProducts" :key="index" class="hot-product-item">
              <span class="rank" :class="'rank-' + (index + 1)">{{ index + 1 }}</span>
              <div class="product-info">
                <span class="product-name">{{ item.name }}</span>
                <span class="product-sales">销量 {{ item.sales }}</span>
              </div>
              <span class="product-revenue">￥{{ item.revenue }}</span>
            </div>
          </div>
        </a-card>
      </a-col>
    </a-row>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { api } from '../../api'

const stats = reactive({
  products: 0,
  orders: 0,
  users: 0,
  revenue: 0
})

const recentOrders = ref([])
const hotProducts = ref([])

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
    const res = await api.getDashboard()
    const d = res.data || res
    if (d) {
      stats.products = d.product_count || 0
      stats.orders = d.order_count || 0
      stats.users = d.user_count || 0
      stats.revenue = d.total_revenue || 0
      recentOrders.value = d.recent_orders || []
      hotProducts.value = d.top_products || []
      if (d.sales_trend && Array.isArray(d.sales_trend)) {
        // handle sales trend if needed
      }
    }
  } catch (e) {
    // keep default mock data
  }
}

onMounted(loadData)
</script>

<style scoped>
.dashboard-page {
  padding: 4px;
}

.stat-row {
  margin-bottom: 16px;
}

.stat-card {
  border-radius: 8px;
}

.stat-content {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}

.stat-info {
  flex: 1;
}

.stat-label {
  font-size: 13px;
  color: #86909c;
  margin: 0 0 8px 0;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: #1d2129;
  margin: 0;
}

.stat-icon {
  width: 56px;
  height: 56px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
}

.stat-icon.blue { background: linear-gradient(135deg, #2a7fff, #5ba3ff); }
.stat-icon.green { background: linear-gradient(135deg, #52c41a, #7cc54b); }
.stat-icon.orange { background: linear-gradient(135deg, #fa8c16, #ffa940); }
.stat-icon.purple { background: linear-gradient(135deg, #722ed1, #9254de); }

.stat-trend {
  font-size: 12px;
  color: #86909c;
  margin: 12px 0 0 0;
  display: flex;
  align-items: center;
  gap: 6px;
}

.trend-up {
  color: #52c41a;
  display: flex;
  align-items: center;
  gap: 2px;
}

.content-row {
  margin-bottom: 16px;
}

.content-card {
  border-radius: 8px;
}

.price {
  color: #f53f3f;
  font-weight: 600;
}

.hot-products {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.hot-product-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px;
  background: #f7f8fa;
  border-radius: 6px;
}

.rank {
  width: 28px;
  height: 28px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: 600;
  background: #e5e6eb;
  color: #86909c;
  flex-shrink: 0;
}

.rank.rank-1 { background: #ff4d4f; color: #fff; }
.rank.rank-2 { background: #fa8c16; color: #fff; }
.rank.rank-3 { background: #fadb14; color: #fff; }

.product-info {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.product-name {
  font-size: 14px;
  color: #1d2129;
}

.product-sales {
  font-size: 12px;
  color: #86909c;
}

.product-revenue {
  font-size: 14px;
  font-weight: 600;
  color: #f53f3f;
}
</style>
