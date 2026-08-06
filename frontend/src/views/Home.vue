<template>
  <div class="home-page">
    <header class="home-header">
      <div class="header-inner">
        <h1 class="site-title">{{ siteName }}</h1>
        <a class="query-link" @click.prevent="goQuery">
          <iconify-icon icon="arco:search" :size="16" />
          查单
        </a>
      </div>
    </header>

    <main class="home-main">
      <template v-if="groupedProducts.length > 0">
        <section v-for="(group, gi) in groupedProducts" :key="gi" class="product-group">
          <h2 class="group-title">{{ group.category || '商品列表' }}</h2>
          <div class="product-grid">
            <div
              v-for="product in group.products"
              :key="product.id"
              class="product-card"
              @click="goDetail(product.id)"
            >
              <div class="product-name">{{ product.name }}</div>
              <div class="product-price">
                <span class="currency">¥</span>
                <span class="price-num">{{ product.price }}</span>
              </div>
              <div class="product-stock">库存: {{ getStockText(product.stock) }}</div>
              <a-button type="primary" size="small" class="buy-btn">立即购买</a-button>
            </div>
          </div>
        </section>
      </template>

      <div v-else class="empty-state">
        <iconify-icon icon="arco:file" :size="48" />
        <p>暂无商品，请先登录后台添加商品</p>
      </div>
    </main>

    <footer class="home-footer">
      <p>&copy; 2024 {{ siteName }}. All rights reserved.</p>
    </footer>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../api'

const router = useRouter()
const siteName = ref('晨泽发卡')
const groupedProducts = ref([])

const getStockText = (product) => {
  if (product.stock_text) return product.stock_text
  if (product.stock <= 0) return '缺货'
  if (product.stock <= 5) return '少量'
  if (product.stock <= 20) return '充足'
  return '库存充足'
}

const goQuery = () => {
  router.push('/query')
}

const goDetail = (id) => {
  router.push(`/product/${id}`)
}

onMounted(async () => {
  try {
    const res = await api.getProductsGrouped()
    if (res.data && res.data.length > 0) {
      groupedProducts.value = res.data
    }
  } catch (e) {
    console.warn('加载商品失败:', e)
  }
})
</script>

<style scoped>
.home-page {
  min-height: 100vh;
  background: #f7f8fa;
}

.home-header {
  background: #fff;
  border-bottom: 1px solid #e5e6eb;
  padding: 16px 20px;
  position: sticky;
  top: 0;
  z-index: 100;
}

.header-inner {
  max-width: 900px;
  margin: 0 auto;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.site-title {
  font-size: 22px;
  font-weight: 700;
  color: #1d2129;
  margin: 0;
}

.query-link {
  display: flex;
  align-items: center;
  gap: 4px;
  color: #4e5969;
  cursor: pointer;
  font-size: 14px;
  text-decoration: none;
  padding: 4px 10px;
  border-radius: 4px;
  transition: all 0.2s;
}

.query-link:hover {
  color: #2a7fff;
  background: #e8f3ff;
}

.home-main {
  max-width: 900px;
  margin: 0 auto;
  padding: 20px 16px 40px;
}

.product-group {
  margin-bottom: 32px;
}

.group-title {
  font-size: 18px;
  font-weight: 600;
  color: #1d2129;
  margin: 0 0 16px;
  padding-left: 10px;
  border-left: 4px solid #2a7fff;
}

.product-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}

.product-card {
  background: #fff;
  border-radius: 8px;
  padding: 16px;
  cursor: pointer;
  transition: all 0.2s ease;
  border: 1px solid #e5e6eb;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.product-card:hover {
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.08);
  border-color: #2a7fff;
}

.product-name {
  font-size: 15px;
  font-weight: 600;
  color: #1d2129;
}

.product-price {
  color: #f53f3f;
  display: flex;
  align-items: baseline;
  gap: 2px;
}

.product-price .currency {
  font-size: 16px;
  font-weight: 500;
}

.product-price .price-num {
  font-size: 28px;
  font-weight: 700;
  line-height: 1;
}

.product-stock {
  font-size: 13px;
  color: #86909c;
  margin-bottom: 4px;
}

.buy-btn {
  align-self: flex-end;
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
  color: #86909c;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}

.home-footer {
  text-align: center;
  padding: 20px;
  color: #86909c;
  font-size: 12px;
  background: #fff;
  border-top: 1px solid #e5e6eb;
}

@media (min-width: 768px) {
  .product-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media (min-width: 1024px) {
  .product-grid {
    grid-template-columns: repeat(4, 1fr);
  }
}

@media (max-width: 480px) {
  .home-header {
    padding: 12px 16px;
  }
  
  .site-title {
    font-size: 18px;
  }
  
  .home-main {
    padding: 16px 12px 40px;
  }
  
  .product-grid {
    gap: 10px;
  }
  
  .product-card {
    padding: 14px 12px;
  }
  
  .product-price .price-num {
    font-size: 24px;
  }
  
  .product-name {
    font-size: 14px;
  }
}
</style>
