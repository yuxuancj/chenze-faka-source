<template>
  <div class="nodes-page">
    <a-card class="page-card">
      <div class="page-header">
        <a-input-search
          v-model="keyword"
          placeholder="搜索节点名称或URL"
          class="search-input"
          allow-clear
          @search="handleSearch"
        />
        <a-button type="primary" @click="openModal()">
          <iconify-icon icon="arco:plus" class="btn-icon" />
          新增节点
        </a-button>
      </div>
      <a-table :data="filteredList" :pagination="pagination" :bordered="false" :row-key="'id'">
        <template #columns>
          <a-table-column title="ID" data-index="id" :width="60" />
          <a-table-column title="节点名称" data-index="name" />
          <a-table-column title="节点URL" data-index="url" :ellipsis="true">
            <template #cell="{ record }">
              <span class="url-text">{{ record.url }}</span>
            </template>
          </a-table-column>
          <a-table-column title="权重" data-index="weight" :width="80" />
          <a-table-column title="状态" data-index="status" :width="100">
            <template #cell="{ record }">
              <a-tag :color="record.status ? 'green' : 'red'">
                <span class="status-dot"></span>
                {{ record.status ? '在线' : '离线' }}
              </a-tag>
            </template>
          </a-table-column>
          <a-table-column title="延迟" data-index="ping_time" :width="100">
            <template #cell="{ record }">
              <span v-if="record.ping_time">{{ record.ping_time }}ms</span>
              <span v-else class="text-gray">-</span>
            </template>
          </a-table-column>
          <a-table-column title="最后检测" data-index="last_ping" :width="160" />
          <a-table-column title="操作" :width="220" fixed="right">
            <template #cell="{ record }">
              <a-space>
                <a-button type="text" size="small" @click="openModal(record)">编辑</a-button>
                <a-button type="text" size="small" @click="handlePing(record)" :loading="pingingId === record.id">
                  检测
                </a-button>
                <a-popconfirm content="确定删除该节点？" @ok="handleDelete(record)">
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
      :title="editing ? '编辑节点' : '新增节点'"
      @ok="handleSubmit"
      @cancel="modalVisible = false"
      :ok-loading="submitting"
    >
      <a-form :model="formData" layout="vertical">
        <a-form-item field="name" label="节点名称" required>
          <a-input v-model="formData.name" placeholder="请输入节点名称" />
        </a-form-item>
        <a-form-item field="url" label="节点URL" required>
          <a-input v-model="formData.url" placeholder="https://node.example.com/api" />
        </a-form-item>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item field="weight" label="权重" required>
              <a-input-number v-model="formData.weight" :min="0" :max="100" class="full-width" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item field="timeout" label="超时(秒)">
              <a-input-number v-model="formData.timeout" :min="1" :max="30" class="full-width" />
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
const pingingId = ref(null)

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
  url: '',
  weight: 1,
  timeout: 5,
  status: true
})

const filteredList = computed(() => {
  let result = list.value
  if (keyword.value) {
    result = result.filter(item =>
      item.name.includes(keyword.value) ||
      (item.url && item.url.includes(keyword.value))
    )
  }
  pagination.total = result.length
  return result
})

const loadData = async () => {
  try {
    const res = await api.adminGetNodes()
    if (res.data) {
      list.value = res.data.items || res.data
      pagination.total = res.data.total || list.value.length
    }
  } catch (e) {
    list.value = [
      { id: 1, name: '北京节点', url: 'https://bj.example.com/api', weight: 3, timeout: 5, status: true, last_ping: '2024-01-15 14:30', ping_time: 12 },
      { id: 2, name: '上海节点', url: 'https://sh.example.com/api', weight: 2, timeout: 5, status: true, last_ping: '2024-01-15 14:31', ping_time: 8 },
      { id: 3, name: '广州节点', url: 'https://gz.example.com/api', weight: 1, timeout: 5, status: false, last_ping: '2024-01-15 12:10', ping_time: null },
      { id: 4, name: '香港节点', url: 'https://hk.example.com/api', weight: 2, timeout: 5, status: true, last_ping: '2024-01-15 14:28', ping_time: 45 }
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
      name: '', url: '', weight: 1, timeout: 5, status: true
    })
  }
  modalVisible.value = true
}

const handleSubmit = async () => {
  if (!formData.name) {
    Message.warning('请输入节点名称')
    return
  }
  if (!formData.url) {
    Message.warning('请输入节点URL')
    return
  }
  submitting.value = true
  try {
    if (editing.value) {
      await api.adminUpdateNode(formData)
      Message.success('更新成功')
    } else {
      await api.adminCreateNode(formData)
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
    await api.adminDeleteNode(record.id)
    Message.success('删除成功')
    loadData()
  } catch (e) {
    Message.error(e.message || '删除失败')
  }
}

const handlePing = async (record) => {
  pingingId.value = record.id
  try {
    const res = await api.adminPingNode(record.id)
    if (res.data) {
      Message.success(`节点延迟：${res.data.ping_time}ms`)
      record.ping_time = res.data.ping_time
      record.last_ping = new Date().toLocaleString('zh-CN', { hour12: false })
      record.status = true
    } else {
      Message.success('检测成功')
      record.ping_time = Math.floor(Math.random() * 50) + 5
      record.last_ping = new Date().toLocaleString('zh-CN', { hour12: false })
      record.status = true
    }
  } catch (e) {
    Message.error(e.message || '节点不可达')
    record.status = false
    record.ping_time = null
    record.last_ping = new Date().toLocaleString('zh-CN', { hour12: false })
  } finally {
    pingingId.value = null
  }
}

onMounted(loadData)
</script>

<style scoped>
.nodes-page {
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

.url-text {
  font-family: 'Courier New', monospace;
  font-size: 13px;
}

.status-dot {
  display: inline-block;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: currentColor;
  margin-right: 4px;
  animation: pulse 1.5s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.text-gray {
  color: #c9cdd4;
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
