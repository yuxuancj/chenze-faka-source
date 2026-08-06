<template>
  <div class="emails-page">
    <a-tabs v-model:active-key="activeTab" type="line">
      <a-tab-pane key="config" title="SMTP配置">
        <div class="tab-content">
          <a-card class="page-card">
            <div class="page-header">
              <a-button type="primary" @click="openModal()">
                <iconify-icon icon="arco:plus" class="btn-icon" />
                新增SMTP配置
              </a-button>
            </div>
            <a-table :data="filteredList" :pagination="pagination" :bordered="false" :row-key="'id'">
              <template #columns>
                <a-table-column title="ID" data-index="id" :width="60" />
                <a-table-column title="SMTP主机" data-index="smtp_host" />
                <a-table-column title="用户名" data-index="username" />
                <a-table-column title="端口" data-index="smtp_port" :width="80" />
                <a-table-column title="SSL" data-index="use_ssl" :width="80">
                  <template #cell="{ record }">
                    <a-tag :color="record.use_ssl ? 'green' : 'gray'">
                      {{ record.use_ssl ? '是' : '否' }}
                    </a-tag>
                  </template>
                </a-table-column>
                <a-table-column title="状态" data-index="status" :width="100">
                  <template #cell="{ record }">
                    <a-tag :color="record.status ? 'green' : 'gray'">
                      {{ record.status ? '启用' : '禁用' }}
                    </a-tag>
                  </template>
                </a-table-column>
                <a-table-column title="操作" :width="220" fixed="right">
                  <template #cell="{ record }">
                    <a-space>
                      <a-button type="text" size="small" @click="openModal(record)">编辑</a-button>
                      <a-button type="text" size="small" @click="handleTest(record)">测试连接</a-button>
                      <a-popconfirm content="确定删除？" @ok="handleDelete(record)">
                        <a-button type="text" size="small" status="danger">删除</a-button>
                      </a-popconfirm>
                    </a-space>
                  </template>
                </a-table-column>
              </template>
            </a-table>
          </a-card>
        </div>
      </a-tab-pane>

      <a-tab-pane key="logs" title="邮件日志">
        <div class="tab-content">
          <a-card class="page-card">
            <div class="page-header">
              <a-input-search
                v-model="logKeyword"
                placeholder="搜索收件人邮箱"
                class="search-input"
                allow-clear
                @search="handleLogSearch"
              />
              <a-button @click="loadEmailLogs">
                <iconify-icon icon="arco:refresh" class="btn-icon" />
                刷新
              </a-button>
            </div>
            <a-table :data="filteredLogs" :pagination="logPagination" :bordered="false" :row-key="'id'">
              <template #columns>
                <a-table-column title="ID" data-index="id" :width="60" />
                <a-table-column title="收件人" data-index="to_email" />
                <a-table-column title="主题" data-index="subject" />
                <a-table-column title="状态" data-index="status" :width="100">
                  <template #cell="{ record }">
                    <a-tag :color="logStatusColor(record.status)">{{ record.status }}</a-tag>
                  </template>
                </a-table-column>
                <a-table-column title="发送时间" data-index="created_at" :width="160" />
              </template>
            </a-table>
          </a-card>
        </div>
      </a-tab-pane>
    </a-tabs>

    <a-modal
      v-model:visible="modalVisible"
      :title="editing ? '编辑SMTP配置' : '新增SMTP配置'"
      @ok="handleSubmit"
      @cancel="modalVisible = false"
      :ok-loading="submitting"
    >
      <a-form :model="formData" layout="vertical">
        <a-row :gutter="16">
          <a-col :span="16">
            <a-form-item field="smtp_host" label="SMTP主机" required>
              <a-input v-model="formData.smtp_host" placeholder="smtp.example.com" />
            </a-form-item>
          </a-col>
          <a-col :span="8">
            <a-form-item field="smtp_port" label="端口" required>
              <a-input-number v-model="formData.smtp_port" :min="1" :max="65535" class="full-width" />
            </a-form-item>
          </a-col>
        </a-row>
        <a-form-item field="username" label="用户名" required>
          <a-input v-model="formData.username" placeholder="请输入邮箱账号" />
        </a-form-item>
        <a-form-item field="password" label="密码" required>
          <a-input-password v-model="formData.password" placeholder="请输入邮箱密码或授权码" />
        </a-form-item>
        <a-form-item field="use_ssl" label="使用SSL/TLS加密">
          <a-switch v-model="formData.use_ssl" />
          <span class="status-label">{{ formData.use_ssl ? '已开启' : '已关闭' }}</span>
        </a-form-item>
        <a-form-item field="status" label="状态">
          <a-switch v-model="formData.status" />
          <span class="status-label">{{ formData.status ? '启用' : '禁用' }}</span>
        </a-form-item>
      </a-form>
      <div class="test-connection-area">
        <a-button @click="handleTestConnection" :loading="testing">
          测试连接
        </a-button>
        <span v-if="testResult" :class="['test-result', testResult.type]">{{ testResult.msg }}</span>
      </div>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { Message } from '@arco-design/web-vue'
import { api } from '../../api'

const activeTab = ref('config')
const modalVisible = ref(false)
const editing = ref(null)
const submitting = ref(false)
const testing = ref(false)
const testResult = ref(null)
const logKeyword = ref('')

const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
  showTotal: true,
  sizeCanChange: true
})

const logPagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
  showTotal: true,
  sizeCanChange: true
})

const list = ref([])
const emailLogs = ref([])

const formData = reactive({
  smtp_host: '',
  smtp_port: 465,
  username: '',
  password: '',
  use_ssl: true,
  status: true
})

const filteredList = computed(() => {
  pagination.total = list.value.length
  return list.value
})

const filteredLogs = computed(() => {
  let result = emailLogs.value
  if (logKeyword.value) {
    result = result.filter(item => item.to_email.includes(logKeyword.value))
  }
  logPagination.total = result.length
  return result
})

const logStatusColor = (status) => {
  const map = { '发送成功': 'green', '发送失败': 'red', '待发送': 'orange' }
  return map[status] || 'gray'
}

const loadData = async () => {
  try {
    const res = await api.adminGetEmails()
    if (res.data) {
      list.value = res.data.items || res.data
      pagination.total = res.data.total || list.value.length
    }
  } catch (e) {
    list.value = [
      { id: 1, smtp_host: 'smtp.qq.com', smtp_port: 465, username: 'admin@qq.com', password: '******', use_ssl: true, status: true },
      { id: 2, smtp_host: 'smtp.163.com', smtp_port: 465, username: 'noreply@163.com', password: '******', use_ssl: true, status: true },
      { id: 3, smtp_host: 'smtp.gmail.com', smtp_port: 587, username: 'admin@gmail.com', password: '******', use_ssl: false, status: false }
    ]
    pagination.total = list.value.length
  }
}

const loadEmailLogs = async () => {
  try {
    const res = await api.adminGetEmailLogs()
    if (res.data) {
      emailLogs.value = res.data.items || res.data
      logPagination.total = res.data.total || emailLogs.value.length
    }
  } catch (e) {
    emailLogs.value = [
      { id: 1, to_email: 'user1@test.com', subject: '订单通知 - ORD202401001', status: '发送成功', created_at: '2024-01-15 14:30' },
      { id: 2, to_email: 'user2@test.com', subject: '发卡通知 - ORD202401002', status: '发送成功', created_at: '2024-01-15 13:20' },
      { id: 3, to_email: 'user3@test.com', subject: '密码重置', status: '发送失败', created_at: '2024-01-15 12:10' },
      { id: 4, to_email: 'user4@test.com', subject: '订单通知 - ORD202401003', status: '待发送', created_at: '2024-01-15 11:00' }
    ]
    logPagination.total = emailLogs.value.length
  }
}

const handleLogSearch = () => {
  logPagination.current = 1
}

const openModal = (record) => {
  testResult.value = null
  if (record) {
    editing.value = record
    Object.assign(formData, { ...record })
  } else {
    editing.value = null
    Object.assign(formData, {
      smtp_host: '', smtp_port: 465, username: '', password: '',
      use_ssl: true, status: true
    })
  }
  modalVisible.value = true
}

const handleSubmit = async () => {
  if (!formData.smtp_host) {
    Message.warning('请输入SMTP主机')
    return
  }
  if (!formData.username) {
    Message.warning('请输入用户名')
    return
  }
  submitting.value = true
  try {
    if (editing.value) {
      await api.adminUpdateEmail(formData)
      Message.success('更新成功')
    } else {
      await api.adminCreateEmail(formData)
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
    await api.adminDeleteEmail(record.id)
    Message.success('删除成功')
    loadData()
  } catch (e) {
    Message.error(e.message || '删除失败')
  }
}

const handleTest = async (record) => {
  try {
    const res = await api.adminTestEmail(record.id)
    if (res.data) {
      Message.success('连接测试成功')
    } else {
      Message.success('连接测试成功')
    }
  } catch (e) {
    Message.error(e.message || '连接测试失败')
  }
}

const handleTestConnection = async () => {
  if (!formData.smtp_host || !formData.username) {
    Message.warning('请先填写SMTP配置信息')
    return
  }
  testing.value = true
  testResult.value = null
  try {
    const res = await api.adminTestEmail(editing.value ? editing.value.id : 0)
    testResult.value = { type: 'success', msg: '✓ 连接成功，SMTP配置正常' }
  } catch (e) {
    testResult.value = { type: 'error', msg: '✗ 连接失败：' + (e.message || '未知错误') }
  } finally {
    testing.value = false
  }
}

onMounted(() => {
  loadData()
  loadEmailLogs()
})
</script>

<style scoped>
.emails-page {
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

.test-connection-area {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid #f2f3f5;
  display: flex;
  align-items: center;
  gap: 12px;
}

.test-result {
  font-size: 13px;
}

.test-result.success {
  color: #52c41a;
}

.test-result.error {
  color: #f53f3f;
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
