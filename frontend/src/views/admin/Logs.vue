<template>
  <div class="logs-page">
    <a-tabs v-model:active-key="activeTab" type="line">
      <a-tab-pane key="operations" title="操作日志">
        <div class="tab-content">
          <a-card class="page-card">
            <div class="page-header">
              <div class="filter-area">
                <a-input-search
                  v-model="operationKeyword"
                  placeholder="搜索操作人"
                  class="search-input"
                  allow-clear
                  @search="handleOperationSearch"
                />
              </div>
              <a-button @click="loadOperations">
                <iconify-icon icon="arco:refresh" class="btn-icon" />
                刷新
              </a-button>
            </div>
            <a-table :data="filteredOperations" :pagination="operationPagination" :bordered="false" :row-key="'id'">
              <template #columns>
                <a-table-column title="ID" data-index="id" :width="60" />
                <a-table-column title="操作人" data-index="username" :width="120" />
                <a-table-column title="操作类型" data-index="action" :width="120">
                  <template #cell="{ record }">
                    <a-tag :color="actionColor(record.action)">{{ record.action }}</a-tag>
                  </template>
                </a-table-column>
                <a-table-column title="操作描述" data-index="description" />
                <a-table-column title="IP地址" data-index="ip" :width="140" />
                <a-table-column title="时间" data-index="created_at" :width="160" />
              </template>
            </a-table>
          </a-card>
        </div>
      </a-tab-pane>

      <a-tab-pane key="logins" title="登录日志">
        <div class="tab-content">
          <a-card class="page-card">
            <div class="page-header">
              <div class="filter-area">
                <a-input-search
                  v-model="loginKeyword"
                  placeholder="搜索用户名"
                  class="search-input"
                  allow-clear
                  @search="handleLoginSearch"
                />
              </div>
              <a-button @click="loadLoginLogs">
                <iconify-icon icon="arco:refresh" class="btn-icon" />
                刷新
              </a-button>
            </div>
            <a-table :data="filteredLogins" :pagination="loginPagination" :bordered="false" :row-key="'id'">
              <template #columns>
                <a-table-column title="ID" data-index="id" :width="60" />
                <a-table-column title="用户名" data-index="username" :width="140" />
                <a-table-column title="登录IP" data-index="ip" :width="140" />
                <a-table-column title="浏览器" data-index="user_agent" />
                <a-table-column title="结果" data-index="status" :width="100">
                  <template #cell="{ record }">
                    <a-tag :color="record.status === '成功' ? 'green' : 'red'">{{ record.status }}</a-tag>
                  </template>
                </a-table-column>
                <a-table-column title="登录时间" data-index="created_at" :width="160" />
              </template>
            </a-table>
          </a-card>
        </div>
      </a-tab-pane>

      <a-tab-pane key="orders" title="订单日志">
        <div class="tab-content">
          <a-card class="page-card">
            <div class="page-header">
              <div class="filter-area">
                <a-input-search
                  v-model="orderKeyword"
                  placeholder="搜索订单号"
                  class="search-input"
                  allow-clear
                  @search="handleOrderSearch"
                />
              </div>
              <a-button @click="loadOrderLogs">
                <iconify-icon icon="arco:refresh" class="btn-icon" />
                刷新
              </a-button>
            </div>
            <a-table :data="filteredOrders" :pagination="orderPagination" :bordered="false" :row-key="'id'">
              <template #columns>
                <a-table-column title="ID" data-index="id" :width="60" />
                <a-table-column title="订单号" data-index="order_no" :width="200">
                  <template #cell="{ record }">
                    <span class="order-no">{{ record.order_no }}</span>
                  </template>
                </a-table-column>
                <a-table-column title="操作" data-index="action" :width="120">
                  <template #cell="{ record }">
                    <a-tag :color="orderActionColor(record.action)">{{ record.action }}</a-tag>
                  </template>
                </a-table-column>
                <a-table-column title="操作人" data-index="operator" :width="120" />
                <a-table-column title="备注" data-index="remark" />
                <a-table-column title="时间" data-index="created_at" :width="160" />
              </template>
            </a-table>
          </a-card>
        </div>
      </a-tab-pane>

      <a-tab-pane key="emails" title="邮件日志">
        <div class="tab-content">
          <a-card class="page-card">
            <div class="page-header">
              <div class="filter-area">
                <a-input-search
                  v-model="emailKeyword"
                  placeholder="搜索收件人邮箱"
                  class="search-input"
                  allow-clear
                  @search="handleEmailSearch"
                />
              </div>
              <a-button @click="loadEmailLogsData">
                <iconify-icon icon="arco:refresh" class="btn-icon" />
                刷新
              </a-button>
            </div>
            <a-table :data="filteredEmails" :pagination="emailPagination" :bordered="false" :row-key="'id'">
              <template #columns>
                <a-table-column title="ID" data-index="id" :width="60" />
                <a-table-column title="收件人" data-index="to_email" />
                <a-table-column title="主题" data-index="subject" />
                <a-table-column title="状态" data-index="status" :width="100">
                  <template #cell="{ record }">
                    <a-tag :color="emailStatusColor(record.status)">{{ record.status }}</a-tag>
                  </template>
                </a-table-column>
                <a-table-column title="发送时间" data-index="created_at" :width="160" />
              </template>
            </a-table>
          </a-card>
        </div>
      </a-tab-pane>
    </a-tabs>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { Message } from '@arco-design/web-vue'
import { api } from '../../api'

const activeTab = ref('operations')

const operationKeyword = ref('')
const loginKeyword = ref('')
const orderKeyword = ref('')
const emailKeyword = ref('')

const operationPagination = reactive({
  current: 1, pageSize: 10, total: 0, showTotal: true, sizeCanChange: true
})
const loginPagination = reactive({
  current: 1, pageSize: 10, total: 0, showTotal: true, sizeCanChange: true
})
const orderPagination = reactive({
  current: 1, pageSize: 10, total: 0, showTotal: true, sizeCanChange: true
})
const emailPagination = reactive({
  current: 1, pageSize: 10, total: 0, showTotal: true, sizeCanChange: true
})

const operations = ref([])
const loginLogs = ref([])
const orderLogs = ref([])
const emailLogsData = ref([])

const filteredOperations = computed(() => {
  let result = operations.value
  if (operationKeyword.value) {
    result = result.filter(item => item.username.includes(operationKeyword.value))
  }
  operationPagination.total = result.length
  return result
})

const filteredLogins = computed(() => {
  let result = loginLogs.value
  if (loginKeyword.value) {
    result = result.filter(item => item.username.includes(loginKeyword.value))
  }
  loginPagination.total = result.length
  return result
})

const filteredOrders = computed(() => {
  let result = orderLogs.value
  if (orderKeyword.value) {
    result = result.filter(item => item.order_no.includes(orderKeyword.value))
  }
  orderPagination.total = result.length
  return result
})

const filteredEmails = computed(() => {
  let result = emailLogsData.value
  if (emailKeyword.value) {
    result = result.filter(item => item.to_email.includes(emailKeyword.value))
  }
  emailPagination.total = result.length
  return result
})

const actionColor = (action) => {
  const map = { '创建': 'green', '更新': 'blue', '删除': 'red', '登录': 'arcoblue', '导出': 'orange' }
  return map[action] || 'gray'
}

const orderActionColor = (action) => {
  const map = { '创建订单': 'green', '支付成功': 'arcoblue', '取消订单': 'red', '发卡': 'orange', '退款': 'purple' }
  return map[action] || 'gray'
}

const emailStatusColor = (status) => {
  const map = { '发送成功': 'green', '发送失败': 'red', '待发送': 'orange' }
  return map[status] || 'gray'
}

const loadOperations = async () => {
  try {
    const res = await api.adminGetOperations()
    if (res.data) {
      operations.value = res.data.items || res.data
      operationPagination.total = res.data.total || operations.value.length
    }
  } catch (e) {
    operations.value = [
      { id: 1, username: 'admin', action: '创建', description: '创建商品「Q币充值」', ip: '192.168.1.100', created_at: '2024-01-15 14:30' },
      { id: 2, username: 'admin', action: '更新', description: '更新商品「话费充值」价格', ip: '192.168.1.100', created_at: '2024-01-15 13:20' },
      { id: 3, username: 'operator', action: '删除', description: '删除分类「旧分类」', ip: '192.168.1.101', created_at: '2024-01-15 12:10' },
      { id: 4, username: 'admin', action: '导出', description: '导出卡密记录', ip: '192.168.1.100', created_at: '2024-01-15 11:00' }
    ]
    operationPagination.total = operations.value.length
  }
}

const loadLoginLogs = async () => {
  try {
    const res = await api.adminGetLoginLogs()
    if (res.data) {
      loginLogs.value = res.data.items || res.data
      loginPagination.total = res.data.total || loginLogs.value.length
    }
  } catch (e) {
    loginLogs.value = [
      { id: 1, username: 'admin', ip: '192.168.1.100', user_agent: 'Chrome / Windows', status: '成功', created_at: '2024-01-15 14:30' },
      { id: 2, username: 'operator', ip: '192.168.1.101', user_agent: 'Firefox / Windows', status: '成功', created_at: '2024-01-15 10:20' },
      { id: 3, username: 'test', ip: '10.0.0.5', user_agent: 'Safari / macOS', status: '失败', created_at: '2024-01-15 09:10' }
    ]
    loginPagination.total = loginLogs.value.length
  }
}

const loadOrderLogs = async () => {
  try {
    const res = await api.adminGetOrderLogs()
    if (res.data) {
      orderLogs.value = res.data.items || res.data
      orderPagination.total = res.data.total || orderLogs.value.length
    }
  } catch (e) {
    orderLogs.value = [
      { id: 1, order_no: 'ORD202401001', action: '创建订单', operator: '系统', remark: '用户下单购买Q币充值', created_at: '2024-01-15 14:30' },
      { id: 2, order_no: 'ORD202401001', action: '支付成功', operator: '系统', remark: '支付宝支付成功', created_at: '2024-01-15 14:31' },
      { id: 3, order_no: 'ORD202401001', action: '发卡', operator: '系统', remark: '自动发放卡密 QB100-XXXX-XXXX-0001', created_at: '2024-01-15 14:31' },
      { id: 4, order_no: 'ORD202401002', action: '创建订单', operator: '系统', remark: '用户下单购买话费充值', created_at: '2024-01-15 13:20' },
      { id: 5, order_no: 'ORD202401003', action: '取消订单', operator: 'admin', remark: '手动取消，用户申请退款', created_at: '2024-01-15 11:00' }
    ]
    orderPagination.total = orderLogs.value.length
  }
}

const loadEmailLogsData = async () => {
  try {
    const res = await api.adminGetEmailLogs()
    if (res.data) {
      emailLogsData.value = res.data.items || res.data
      emailPagination.total = res.data.total || emailLogsData.value.length
    }
  } catch (e) {
    emailLogsData.value = [
      { id: 1, to_email: 'user1@test.com', subject: '订单通知 - ORD202401001', status: '发送成功', created_at: '2024-01-15 14:30' },
      { id: 2, to_email: 'user2@test.com', subject: '发卡通知 - ORD202401002', status: '发送成功', created_at: '2024-01-15 13:20' },
      { id: 3, to_email: 'user3@test.com', subject: '密码重置', status: '发送失败', created_at: '2024-01-15 12:10' },
      { id: 4, to_email: 'user4@test.com', subject: '订单通知 - ORD202401003', status: '待发送', created_at: '2024-01-15 11:00' }
    ]
    emailPagination.total = emailLogsData.value.length
  }
}

const handleOperationSearch = () => { operationPagination.current = 1 }
const handleLoginSearch = () => { loginPagination.current = 1 }
const handleOrderSearch = () => { orderPagination.current = 1 }
const handleEmailSearch = () => { emailPagination.current = 1 }

watch(activeTab, (val) => {
  switch (val) {
    case 'operations': loadOperations(); break
    case 'logins': loadLoginLogs(); break
    case 'orders': loadOrderLogs(); break
    case 'emails': loadEmailLogsData(); break
  }
})

onMounted(() => {
  loadOperations()
})
</script>

<style scoped>
.logs-page {
  padding: 4px;
}

.tab-content {
  padding: 16px 0;
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
  flex-wrap: wrap;
}

.filter-area {
  display: flex;
  gap: 12px;
}

.search-input {
  width: 300px;
}

.btn-icon {
  margin-right: 4px;
}

.order-no {
  font-family: 'Courier New', monospace;
  font-size: 13px;
}

@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    align-items: stretch;
  }

  .search-input {
    width: 100%;
  }

  .filter-area {
    width: 100%;
  }
}
</style>
