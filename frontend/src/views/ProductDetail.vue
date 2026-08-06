<template>
  <div class="product-detail-page">
    <div class="page-container">
      <a-page-header
        class="detail-header"
        @back="goBack"
      >
        <template #title>
          <span>{{ product.name || '商品详情' }}</span>
        </template>
      </a-page-header>
      <a-row :gutter="[24, 24]">
        <a-col :xs="24" :md="12">
          <div class="product-image">
            <div class="image-placeholder" :style="{ background: product.color || '#2a7fff' }">
              <iconify-icon :icon="product.icon || 'arco:gift'" :size="80" />
            </div>
          </div>
        </a-col>
        <a-col :xs="24" :md="12">
          <a-card class="detail-card" :bordered="false">
            <h1 class="product-name">{{ product.name }}</h1>
            <p class="product-desc">{{ product.description }}</p>
            <div class="product-price-section">
              <span class="label">售价</span>
              <span class="price">{{ product.price_label || ('￥' + product.price) }}</span>
            </div>
            <a-divider />
            <a-form :model="orderForm" layout="vertical" class="order-form">
              <a-form-item field="quantity" label="购买数量">
                <a-input-number v-model="orderForm.quantity" :min="1" :max="99" />
              </a-form-item>
              <a-form-item field="contact" label="联系方式（手机号/邮箱）">
                <a-input v-model="orderForm.contact" placeholder="请输入接收卡密的联系方式" />
              </a-form-item>
              <a-form-item field="remark" label="备注">
                <a-textarea v-model="orderForm.remark" placeholder="选填" :auto-size="{ minRows: 2, maxRows: 4 }" />
              </a-form-item>
            </a-form>
            <div class="order-summary">
              <div class="summary-row">
                <span>小计</span>
                <span class="summary-price">￥{{ totalPrice }}</span>
              </div>
            </div>
            <a-space class="action-buttons">
              <a-button type="primary" size="large" long @click="handleBuy">
                立即购买
              </a-button>
              <a-button size="large" long @click="handleAddToCart">
                加入购物车
              </a-button>
            </a-space>
          </a-card>
        </a-col>
      </a-row>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Message } from '@arco-design/web-vue'
import { api } from '../api'

const route = useRoute()
const router = useRouter()

const product = ref({})

const orderForm = reactive({
  quantity: 1,
  contact: '',
  remark: ''
})

const totalPrice = computed(() => {
  const price = product.value.price || 0
  return (price * orderForm.quantity).toFixed(2)
})

const goBack = () => {
  router.push('/')
}

const handleBuy = () => {
  if (!orderForm.contact) {
    Message.warning('请填写联系方式')
    return
  }
  router.push({
    name: 'checkout',
    query: {
      product_id: product.value.id,
      quantity: orderForm.quantity,
      contact: orderForm.contact,
      remark: orderForm.remark
    }
  })
}

const handleAddToCart = () => {
  Message.success('已加入购物车')
}

onMounted(() => {
  const id = route.params.id
  if (id) {
    api.getProduct(id).then(res => {
      if (res.data) {
        product.value = res.data
      }
    }).catch(() => {
      product.value = {
        id: id,
        name: '示例商品',
        description: '这是一个示例商品，请通过后台管理添加真实商品数据',
        price: 99,
        price_label: '￥99',
        icon: 'arco:gift',
        color: '#2a7fff'
      }
    })
  }
})
</script>

<style scoped>
.product-detail-page {
  min-height: 100vh;
  background: #f5f7fa;
  padding: 24px 0;
}

.page-container {
  max-width: 1000px;
  margin: 0 auto;
  padding: 0 24px;
}

.detail-header {
  background: transparent;
  padding: 0;
  margin-bottom: 24px;
}

.product-image {
  background: #fff;
  border-radius: 12px;
  padding: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 300px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.06);
}

.image-placeholder {
  width: 160px;
  height: 160px;
  border-radius: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
}

.detail-card {
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.06);
}

.product-name {
  font-size: 24px;
  font-weight: 600;
  color: #1d2129;
  margin: 0 0 12px 0;
}

.product-desc {
  font-size: 14px;
  color: #86909c;
  margin: 0 0 20px 0;
}

.product-price-section {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  background: #fff7e8;
  border-radius: 8px;
  margin-bottom: 16px;
}

.product-price-section .label {
  font-size: 14px;
  color: #86909c;
}

.product-price-section .price {
  font-size: 28px;
  font-weight: 700;
  color: #f53f3f;
}

.order-form {
  margin-bottom: 16px;
}

.order-summary {
  padding: 16px 0;
  border-top: 1px dashed #e5e6eb;
  margin-bottom: 16px;
}

.summary-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 14px;
  color: #4e5969;
}

.summary-price {
  font-size: 20px;
  font-weight: 600;
  color: #f53f3f;
}

.action-buttons {
  width: 100%;
}

.action-buttons .arco-btn {
  flex: 1;
}
</style>
