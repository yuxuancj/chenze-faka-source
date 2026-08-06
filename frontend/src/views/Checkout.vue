<template>
  <div class="checkout-page">
    <div class="page-container">
      <a-page-header class="checkout-header" @back="goBack">
        <template #title>确认订单</template>
      </a-page-header>
      <a-row :gutter="[24, 24]">
        <a-col :xs="24" :md="16">
          <a-card class="checkout-card" title="订单信息">
            <a-descriptions :column="1" bordered size="large">
              <a-descriptions-item label="商品名称">
                {{ orderData.product_name || '-' }}
              </a-descriptions-item>
              <a-descriptions-item label="购买数量">
                {{ orderData.quantity || 1 }}
              </a-descriptions-item>
              <a-descriptions-item label="联系方式">
                {{ orderData.contact || '-' }}
              </a-descriptions-item>
              <a-descriptions-item label="备注">
                {{ orderData.remark || '无' }}
              </a-descriptions-item>
            </a-descriptions>
          </a-card>
          <a-card class="checkout-card" title="支付方式">
            <a-radio-group v-model="payMethod" class="pay-method-group">
              <a-radio value="alipay">
                <div class="pay-method">
                  <iconify-icon icon="arco:alipay" class="pay-icon alipay" />
                  <span>支付宝</span>
                </div>
              </a-radio>
              <a-radio value="wechat">
                <div class="pay-method">
                  <iconify-icon icon="arco:wechat" class="pay-icon wechat" />
                  <span>微信支付</span>
                </div>
              </a-radio>
            </a-radio-group>
          </a-card>
        </a-col>
        <a-col :xs="24" :md="8">
          <a-card class="summary-card" title="订单金额">
            <div class="amount-display">
              <span class="currency">￥</span>
              <span class="amount">{{ totalAmount }}</span>
            </div>
            <a-divider />
            <div class="summary-detail">
              <div class="summary-item">
                <span>商品金额</span>
                <span>￥{{ totalAmount }}</span>
              </div>
              <div class="summary-item">
                <span>优惠</span>
                <span class="discount">-￥0.00</span>
              </div>
            </div>
            <a-divider />
            <a-button
              type="primary"
              size="large"
              long
              class="pay-button"
              :loading="paying"
              @click="handlePay"
            >
              立即支付
            </a-button>
          </a-card>
        </a-col>
      </a-row>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Message, Modal } from '@arco-design/web-vue'
import { api } from '../api'

const route = useRoute()
const router = useRouter()

const paying = ref(false)
const payMethod = ref('alipay')

const orderData = reactive({
  product_id: route.query.product_id,
  product_name: '',
  quantity: parseInt(route.query.quantity) || 1,
  contact: route.query.contact || '',
  remark: route.query.remark || ''
})

const totalAmount = computed(() => {
  return (99 * orderData.quantity).toFixed(2)
})

const goBack = () => {
  router.back()
}

const handlePay = async () => {
  Modal.confirm({
    title: '确认支付',
    content: `确认使用${payMethod.value === 'alipay' ? '支付宝' : '微信'}支付 ￥${totalAmount.value}？`,
    onOk: async () => {
      paying.value = true
      try {
        const res = await api.createOrder({
          product_id: orderData.product_id,
          quantity: orderData.quantity,
          contact: orderData.contact,
          remark: orderData.remark,
          pay_method: payMethod.value
        })
        if (res.data) {
          Message.success('支付成功！')
          router.push(`/query?order_no=${res.data.order_no}`)
        }
      } catch (e) {
        Message.error(e.message || '支付失败，请重试')
      } finally {
        paying.value = false
      }
    }
  })
}
</script>

<style scoped>
.checkout-page {
  min-height: 100vh;
  background: #f5f7fa;
  padding: 24px 0;
}

.page-container {
  max-width: 1100px;
  margin: 0 auto;
  padding: 0 24px;
}

.checkout-header {
  background: transparent;
  padding: 0;
  margin-bottom: 24px;
}

.checkout-card {
  margin-bottom: 24px;
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.06);
}

.pay-method-group {
  display: flex;
  gap: 40px;
}

.pay-method {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 15px;
}

.pay-icon {
  font-size: 24px;
}

.pay-icon.alipay {
  color: #1677ff;
}

.pay-icon.wechat {
  color: #07c160;
}

.summary-card {
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.06);
}

.amount-display {
  text-align: center;
  padding: 16px 0;
}

.amount-display .currency {
  font-size: 24px;
  font-weight: 600;
  color: #f53f3f;
}

.amount-display .amount {
  font-size: 42px;
  font-weight: 700;
  color: #f53f3f;
}

.summary-detail {
  padding: 8px 0;
}

.summary-item {
  display: flex;
  justify-content: space-between;
  padding: 8px 0;
  font-size: 14px;
  color: #4e5969;
}

.discount {
  color: #52c41a;
}

.pay-button {
  height: 48px;
  font-size: 16px;
}
</style>
