<template>
  <div class="home-page">
    <header class="home-header">
      <div class="header-content">
        <h1 class="site-title">{{ siteName }}</h1>
        <p class="site-desc">专业的数字商品发卡平台</p>
        <div class="header-actions">
          <a-button type="primary" size="large" @click="scrollToProducts">
            开始选购
          </a-button>
          <a-button size="large" @click="goQuery">
            查询订单
          </a-button>
        </div>
      </div>
    </header>
    <main class="home-main" ref="productsRef">
      <div class="section-header">
        <h2 class="section-title">商品分类</h2>
      </div>
      <a-row :gutter="[16, 16]" class="category-grid">
        <a-col
          v-for="product in products"
          :key="product.id"
          :xs="24"
          :sm="12"
          :md="8"
          :lg="6"
        >
          <div class="product-card" @click="goDetail(product.id)">
            <div class="product-icon" :style="{ background: product.color || '#2a7fff' }">
              <iconify-icon :icon="product.icon || 'arco:gift'" :size="32" />
            </div>
            <div class="product-info">
              <h3 class="product-name">{{ product.name }}</h3>
              <p class="product-desc">{{ product.description }}</p>
              <div class="product-footer">
                <span class="product-price">{{ product.price_label || ('￥' + product.price) }}</span>
                <a-button type="primary" size="mini">立即购买</a-button>
              </div>
            </div>
          </div>
        </a-col>
      </a-row>
    </main>
    <footer class="home-footer">
      <p>&copy; 2024 {{ siteName }}. All rights reserved.</p>
    </footer>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAdminStore } from '../stores'

const router = useRouter()
const store = useAdminStore()

const siteName = store.siteName
const productsRef = ref(null)

const products = ref([
  { id: 1, name: 'Q币充值', description: '100Q币 官方直充', price: 100, price_label: '￥95', icon: 'arco:gift', color: '#2a7fff' },
  { id: 2, name: '话费充值', description: '100元话费 秒到账', price: 100, price_label: '￥98', icon: 'arco:phone', color: '#52c41a' },
  { id: 3, name: '视频会员', description: '腾讯视频VIP月卡', price: 25, price_label: '￥22', icon: 'arco:play-circle', color: '#fa541c' },
  { id: 4, name: '音乐会员', description: 'QQ音乐绿钻豪华版', price: 18, price_label: '￥15', icon: 'arco:music', color: '#722ed1' },
  { id: 5, name: '游戏点卡', description: '网易一卡通500点', price: 50, price_label: '￥48', icon: 'arco:game', color: '#13c2c2' },
  { id: 6, name: 'steam充值', description: 'Steam钱包充值100元', price: 100, price_label: '￥99', icon: 'arco:steam', color: '#eb2f96' },
  { id: 7, name: '京东E卡', description: '京东E卡500元面值', price: 500, price_label: '￥490', icon: 'arco:shopping-cart', color: '#faad14' },
  { id: 8, name: '加油卡', description: '中石化加油卡1000元', price: 1000, price_label: '￥985', icon: 'arco:car', color: '#2a7fff' }
])

const scrollToProducts = () => {
  productsRef.value?.scrollIntoView({ behavior: 'smooth' })
}

const goQuery = () => {
  router.push('/query')
}

const goDetail = (id) => {
  router.push(`/product/${id}`)
}

onMounted(() => {
  import('../api').then(({ api }) => {
    api.getProducts().then(res => {
      if (res.data && res.data.length > 0) {
        products.value = res.data
      }
    }).catch(() => {})
  })
})
</script>

<style scoped>
.home-page {
  min-height: 100vh;
  background: #f5f7fa;
}

.home-header {
  background: linear-gradient(135deg, #2a7fff 0%, #1a5fcc 100%);
  padding: 80px 20px;
  text-align: center;
  color: #fff;
}

.header-content {
  max-width: 800px;
  margin: 0 auto;
}

.site-title {
  font-size: 48px;
  font-weight: 700;
  margin: 0 0 16px 0;
  letter-spacing: 2px;
}

.site-desc {
  font-size: 18px;
  margin: 0 0 32px 0;
  opacity: 0.9;
}

.header-actions {
  display: flex;
  justify-content: center;
  gap: 16px;
}

.header-actions .arco-btn {
  min-width: 140px;
  height: 44px;
  font-size: 16px;
}

.home-main {
  max-width: 1200px;
  margin: -40px auto 0;
  padding: 0 20px 60px;
  position: relative;
  z-index: 10;
}

.section-header {
  margin-bottom: 24px;
}

.section-title {
  font-size: 24px;
  font-weight: 600;
  color: #1d2129;
  margin: 0;
}

.category-grid {
  margin: 0 !important;
}

.product-card {
  background: #fff;
  border-radius: 12px;
  padding: 20px;
  cursor: pointer;
  transition: all 0.3s ease;
  border: 1px solid #e5e6eb;
  height: 100%;
  display: flex;
  flex-direction: column;
}

.product-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 40px rgba(42, 127, 255, 0.15);
  border-color: #2a7fff;
}

.product-icon {
  width: 64px;
  height: 64px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  margin-bottom: 16px;
}

.product-info {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.product-name {
  font-size: 16px;
  font-weight: 600;
  color: #1d2129;
  margin: 0 0 8px 0;
}

.product-desc {
  font-size: 13px;
  color: #86909c;
  margin: 0 0 16px 0;
  flex: 1;
}

.product-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.product-price {
  font-size: 18px;
  font-weight: 600;
  color: #f53f3f;
}

.home-footer {
  text-align: center;
  padding: 24px;
  color: #86909c;
  font-size: 13px;
  background: #fff;
  border-top: 1px solid #e5e6eb;
}
</style>
