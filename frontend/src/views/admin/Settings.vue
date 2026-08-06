<template>
  <div class="settings-page">
    <a-card title="基本设置" class="page-card">
      <a-form :model="siteForm" layout="vertical" class="settings-form">
        <a-row :gutter="24">
          <a-col :xs="24" :md="12">
            <a-form-item field="site_name" label="网站名称">
              <a-input v-model="siteForm.site_name" placeholder="请输入网站名称" />
            </a-form-item>
          </a-col>
          <a-col :xs="24" :md="12">
            <a-form-item field="site_logo" label="网站Logo">
              <a-input v-model="siteForm.site_logo" placeholder="Logo图片URL" />
            </a-form-item>
          </a-col>
        </a-row>
        <a-form-item field="site_description" label="网站描述">
          <a-textarea
            v-model="siteForm.site_description"
            placeholder="网站描述信息"
            :auto-size="{ minRows: 2, maxRows: 4 }"
          />
        </a-form-item>
        <a-divider />
        <h3 class="section-title">支付配置</h3>
        <a-row :gutter="24">
          <a-col :xs="24" :md="12">
            <a-form-item field="alipay_enabled" label="支付宝">
              <a-switch v-model="siteForm.alipay_enabled" />
              <span class="status-label">{{ siteForm.alipay_enabled ? '已开启' : '已关闭' }}</span>
            </a-form-item>
          </a-col>
          <a-col :xs="24" :md="12">
            <a-form-item field="wechat_enabled" label="微信支付">
              <a-switch v-model="siteForm.wechat_enabled" />
              <span class="status-label">{{ siteForm.wechat_enabled ? '已开启' : '已关闭' }}</span>
            </a-form-item>
          </a-col>
        </a-row>
        <a-row :gutter="24" v-if="siteForm.alipay_enabled">
          <a-col :xs="24" :md="12">
            <a-form-item field="alipay_app_id" label="支付宝AppID">
              <a-input v-model="siteForm.alipay_app_id" placeholder="请输入支付宝AppID" />
            </a-form-item>
          </a-col>
          <a-col :xs="24" :md="12">
            <a-form-item field="alipay_private_key" label="支付宝私钥">
              <a-input-password v-model="siteForm.alipay_private_key" placeholder="请输入应用私钥" />
            </a-form-item>
          </a-col>
        </a-row>
        <a-row :gutter="24" v-if="siteForm.wechat_enabled">
          <a-col :xs="24" :md="12">
            <a-form-item field="wechat_app_id" label="微信AppID">
              <a-input v-model="siteForm.wechat_app_id" placeholder="请输入微信AppID" />
            </a-form-item>
          </a-col>
          <a-col :xs="24" :md="12">
            <a-form-item field="wechat_mch_id" label="微信商户号">
              <a-input v-model="siteForm.wechat_mch_id" placeholder="请输入微信商户号" />
            </a-form-item>
          </a-col>
        </a-row>
        <a-divider />
        <h3 class="section-title">其他设置</h3>
        <a-row :gutter="24">
          <a-col :xs="24" :md="12">
            <a-form-item field="order_expire" label="订单过期时间（分钟）">
              <a-input-number v-model="siteForm.order_expire" :min="5" :max="1440" class="full-width" />
            </a-form-item>
          </a-col>
          <a-col :xs="24" :md="12">
            <a-form-item field="card_prefix" label="卡密前缀">
              <a-input v-model="siteForm.card_prefix" placeholder="可选" />
            </a-form-item>
          </a-col>
        </a-row>
        <a-form-item field="maintenance" label="维护模式">
          <a-switch v-model="siteForm.maintenance" />
          <span class="status-label">{{ siteForm.maintenance ? '已开启' : '已关闭' }}</span>
        </a-form-item>
        <a-button type="primary" size="large" :loading="submitting" @click="handleSave">
          保存设置
        </a-button>
      </a-form>
    </a-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { Message } from '@arco-design/web-vue'
import { api } from '../../api'

const submitting = ref(false)

const siteForm = reactive({
  site_name: '晨泽发卡',
  site_logo: '',
  site_description: '专业的数字商品发卡平台',
  alipay_enabled: true,
  alipay_app_id: '',
  alipay_private_key: '',
  wechat_enabled: false,
  wechat_app_id: '',
  wechat_mch_id: '',
  order_expire: 30,
  card_prefix: '',
  maintenance: false
})

const handleSave = async () => {
  submitting.value = true
  try {
    await api.updateSiteConfig(siteForm)
    localStorage.setItem('site_config', JSON.stringify(siteForm))
    Message.success('保存成功')
  } catch (e) {
    Message.error(e.message || '保存失败')
  } finally {
    submitting.value = false
  }
}

const loadConfig = async () => {
  try {
    const res = await api.getSiteConfig()
    if (res.data) {
      Object.assign(siteForm, res.data)
    }
  } catch (e) {
    try {
      const cached = JSON.parse(localStorage.getItem('site_config') || '{}')
      if (Object.keys(cached).length > 0) {
        Object.assign(siteForm, cached)
      }
    } catch {}
  }
}

onMounted(loadConfig)
</script>

<style scoped>
.settings-page {
  padding: 4px;
  max-width: 900px;
}

.page-card {
  border-radius: 8px;
}

.settings-form {
  padding-top: 8px;
}

.section-title {
  font-size: 16px;
  font-weight: 600;
  color: #1d2129;
  margin: 0 0 16px 0;
}

.status-label {
  margin-left: 12px;
  font-size: 13px;
  color: #86909c;
}

.full-width {
  width: 100%;
}
</style>
