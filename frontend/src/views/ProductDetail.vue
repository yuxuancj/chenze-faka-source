<template>
  <div class="product-detail-page">
    <div class="page-container">
      <a-page-header class="detail-header" @back="goBack">
        <template #title>
          <span>{{ product.name || '商品详情' }}</span>
        </template>
      </a-page-header>
      <div class="detail-layout">
        <div class="product-image">
          <div class="image-placeholder">
            <iconify-icon :icon="product.icon || 'arco:gift'" :size="80" />
          </div>
        </div>
        <div class="detail-info">
          <h1 class="product-name">{{ product.name }}</h1>
          <p class="product-desc">{{ product.description }}</p>
          <div class="product-price-section">
            <span class="label">售价</span>
            <span class="price">¥{{ product.price }}</span>
          </div>
          <div class="stock-info">
            <span>库存: {{ getStockText(product.stock) }}</span>
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
              <span class="summary-price">¥{{ totalPrice }}</span>
            </div>
          </div>
          <a-button type="primary" size="large" long class="buy-btn" @click="handleBuy">
            立即购买
          </a-button>
        </div>
      </div>
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

const getStockText = (stock) => {
  if (stock <= 0) return '缺货'
  if (stock <= 5) return '少量'
  if (stock <= 20) return '充足'
  return '库存充足'
}

const goBack = () => {
  router.push('/')
}

const handleBuy = () => {
  if (!orderForm.contact) {
    Message.warning('请填写联系方式')
    return
  }
  if (product.value.stock <= 0) {
    Message.warning('商品已缺货')
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
        stock: 99,
        icon: 'arco:gift'
      }
    })
  }
})
</script>

<style scoped>
.product-detail-page {
  min-height: 100vh;
  background: #f5f7fa;
  padding: 20px 0;
}

.page-container {
  max-width: 1000px;
  margin: 0 auto;
  padding: 0 20px;
}

.detail-header {
  background: transparent;
  padding: 0;
  margin-bottom: 20px;
}

.detail-layout {
  display: grid;
  grid-template-columns: 1fr;
  gap: 20px;
}

.product-image {
  background: #fff;
  border-radius: 12px;
  padding: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 240px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.06);
}

.image-placeholder {
  width: 140px;
  height: 140px;
  border-radius: 20px;
  background: linear-gradient(135deg, #2a7fff, #1a5fcc);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
}

.detail-info {
  background: #fff;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.06);
}

.product-name {
  font-size: 22px;
  font-weight: 600;
  color: #1d2129;
  margin: 0 0 10px;
}

.product-desc {
  font-size: 14px;
  color: #86909c;
  margin: 0 0 16px;
}

.product-price-section {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 16px;
  background: #fff7e8;
  border-radius: 8px;
  margin-bottom: 8px;
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

.stock-info {
  font-size: 13px;
  color: #86909c;
  padding: 0 4px;
  margin-bottom: 12px;
}

.order-form {
  margin-bottom: 12px;
}

.order-summary {
  padding: 12px 0;
  border-top: 1px dashed #e5e6eb;
  margin-bottom: 12px;
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

.buy-btn {
  height: 48px;
  font-size: 16px;
}

@media (min-width: 768px) {
  .detail-layout {
    grid-template-columns: 1fr 1fr;
  }
  
  .product-image {
    min-height: 400px;
  }
  
  .image-placeholder {
    width: 160px;
    height: 160px;
  }
  
  .detail-info {
    padding: 24px;
  }
}

@media (max-width: 480px) {
  .page-container {
    padding: 0 12px;
  }
  
  .product-image {
    padding: 24px;
    min-height: 200px;
  }
  
  .image-placeholder {
    width: 110px;
    height: 110px;
  }
  
  .detail-info {
    padding: 16px;
  }
  
  .product-name {
    font-size: 18px;
  }
  
  .product-price-section .price {
    font-size: 24px;
  }
}
</style>
