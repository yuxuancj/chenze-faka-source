<template>
  <div class="install-page">
    <div class="install-container">
      <div class="install-header">
        <h1 class="install-title">晨泽发卡系统安装向导</h1>
        <p class="install-desc">请按步骤完成系统安装配置</p>
      </div>

      <!-- Desktop steps -->
      <a-steps v-if="!isMobile" :current="currentStep" class="install-steps">
        <a-step title="许可协议" description="阅读并接受协议" />
        <a-step title="环境检测" description="检查运行环境" />
        <a-step title="数据库配置" description="配置数据库连接" />
        <a-step title="授权验证" description="验证授权密钥" />
        <a-step title="管理员账号" description="设置管理员信息" />
        <a-step title="完成安装" description="系统初始化" />
      </a-steps>

      <!-- Mobile step indicator -->
      <div v-else class="mobile-steps">
        <div class="mobile-step-info">
          <span class="mobile-step-num">{{ currentStep + 1 }}</span>
          <span class="mobile-step-total">/ {{ stepLabels.length }}</span>
          <span class="mobile-step-title">{{ stepLabels[currentStep].title }}</span>
        </div>
        <div class="mobile-step-bar">
          <div
            v-for="(_, index) in stepLabels"
            :key="index"
            class="mobile-step-dot"
            :class="{
              active: index === currentStep,
              done: index < currentStep
            }"
          />
        </div>
      </div>

      <a-card class="install-content">
        <div v-if="currentStep === 0" class="step-content">
          <h3>软件许可协议</h3>
          <div class="agreement">
            <p>欢迎使用晨泽发卡系统。在安装前，请仔细阅读以下协议：</p>
            <p>1. 本软件受版权法保护，未经授权不得复制、修改或分发。</p>
            <p>2. 您承诺不会将本软件用于任何违法用途。</p>
            <p>3. 本软件按"现状"提供，不对任何直接或间接损失承担责任。</p>
            <p>4. 我们保留随时更新本协议的权利。</p>
          </div>
          <a-checkbox v-model="agreed" class="agreement-checkbox">
            我已阅读并同意以上协议
          </a-checkbox>
        </div>

        <div v-else-if="currentStep === 1" class="step-content">
          <h3>环境检测</h3>
          <div class="env-check-list">
            <div class="env-check-item" v-for="item in envItems" :key="item.label">
              <span class="env-check-label">{{ item.label }}</span>
              <span class="env-check-value">{{ item.value }}</span>
              <a-tag :color="getTagColor(item.status)" size="small">
                {{ getStatusText(item.status) }}
              </a-tag>
            </div>
          </div>
        </div>

        <div v-else-if="currentStep === 2" class="step-content">
          <h3>数据库配置</h3>
          <a-form :model="dbForm" layout="vertical" class="db-form">
            <a-row :gutter="16">
              <a-col :xs="24" :sm="12">
                <a-form-item field="host" label="数据库主机">
                  <a-input v-model="dbForm.host" placeholder="localhost" />
                </a-form-item>
              </a-col>
              <a-col :xs="24" :sm="12">
                <a-form-item field="port" label="端口">
                  <a-input-number v-model="dbForm.port" :min="1" :max="65535" style="width: 100%" />
                </a-form-item>
              </a-col>
            </a-row>
            <a-row :gutter="16">
              <a-col :xs="24" :sm="12">
                <a-form-item field="database" label="数据库名">
                  <a-input v-model="dbForm.database" placeholder="chenze_faka" />
                </a-form-item>
              </a-col>
              <a-col :xs="24" :sm="12">
                <a-form-item field="username" label="用户名">
                  <a-input v-model="dbForm.username" placeholder="root" />
                </a-form-item>
              </a-col>
            </a-row>
            <a-form-item field="password" label="密码">
              <a-input-password v-model="dbForm.password" placeholder="请输入数据库密码" />
            </a-form-item>
          </a-form>
          <div class="test-db-section">
            <a-button
              type="outline"
              :loading="testing"
              @click="testDatabaseConnection"
            >
              测试连接
            </a-button>
            <div v-if="dbTestResult" class="test-result" :class="dbTestResult.success ? 'success' : 'error'">
              <iconify-icon :icon="dbTestResult.success ? 'arco:check-circle-fill' : 'arco:close-circle-fill'" />
              {{ dbTestResult.message }}
            </div>
          </div>
        </div>

        <div v-else-if="currentStep === 3" class="step-content">
          <h3>授权验证</h3>
          <a-form :model="licenseForm" layout="vertical">
            <a-form-item field="license_key" label="授权码">
              <a-input-password
                v-model="licenseForm.license_key"
                placeholder="请输入您的授权码"
                allow-clear
              />
            </a-form-item>
            <a-form-item>
              <a-button
                type="outline"
                :loading="verifying"
                :disabled="!licenseForm.license_key"
                @click="verifyLicense"
              >
                验证授权码
              </a-button>
              <span v-if="licenseVerified" class="verified-tip">
                <iconify-icon icon="arco:check-circle-fill" /> 授权验证通过
              </span>
            </a-form-item>
          </a-form>
          <div class="license-tips">
            <p>授权码用于验证软件合法性，请妥善保管。</p>
            <p>如需获取授权码，请联系系统管理员。</p>
          </div>
        </div>

        <div v-else-if="currentStep === 4" class="step-content">
          <h3>管理员账号</h3>
          <a-form :model="adminForm" layout="vertical">
            <a-form-item field="username" label="管理员用户名">
              <a-input v-model="adminForm.username" placeholder="admin" />
            </a-form-item>
            <a-form-item field="password" label="管理员密码">
              <a-input-password v-model="adminForm.password" placeholder="至少6位" />
            </a-form-item>
            <a-form-item field="confirmPassword" label="确认密码">
              <a-input-password v-model="adminForm.confirmPassword" placeholder="再次输入密码" />
            </a-form-item>
          </a-form>
        </div>

        <div v-else-if="currentStep === 5" class="step-content">
          <div v-if="installing" class="installing">
            <a-spin size="40" />
            <p class="installing-text">正在安装，请稍候...</p>
            <p class="installing-desc">系统正在初始化数据库和配置文件</p>
          </div>
          <div v-else class="install-success">
            <iconify-icon icon="arco:check-circle" class="success-icon" />
            <h3>安装完成！</h3>
            <p>系统已成功安装，即将跳转至管理后台。</p>
          </div>
        </div>
      </a-card>

      <div class="install-actions">
        <a-button
          v-if="currentStep > 0 && currentStep < 5"
          size="large"
          @click="prevStep"
        >
          上一步
        </a-button>
        <a-button
          v-if="currentStep < 5"
          type="primary"
          size="large"
          :loading="currentStep === 5 && installing"
          :disabled="currentStep === 2 && !dbTestSuccess"
          @click="nextStep"
        >
          {{ currentStep === 4 ? '开始安装' : '下一步' }}
        </a-button>
      </div>
    </div>

    <a-modal
      v-model:visible="showDataDialog"
      title="检测到数据库已有数据"
      :footer="false"
      :mask-closable="false"
      unmount-on-close
    >
      <div class="data-dialog-content">
        <iconify-icon icon="arco:exclamation-circle-fill" :size="40" class="warning-icon" />
        <p class="dialog-message">
          检测到数据库已有 {{ dbTableCount }} 张数据表，请选择操作方式：
        </p>
        <div class="dialog-actions">
          <a-button type="outline" size="large" @click="handleDataDialogAction('skip')">
            跳过安装（保留现有数据）
          </a-button>
          <a-button type="primary" size="large" status="warning" @click="handleDataDialogAction('overwrite')">
            覆盖安装（清空数据）
          </a-button>
        </div>
      </div>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed, onBeforeUnmount, watch } from 'vue'
import { useRouter } from 'vue-router'
import { Message } from '@arco-design/web-vue'
import { api } from '../api'

const router = useRouter()

const currentStep = ref(0)
const agreed = ref(false)
const installing = ref(false)
const verifying = ref(false)
const licenseVerified = ref(false)
const testing = ref(false)
const dbTestSuccess = ref(false)
const dbTestResult = ref(null)
const dbHasData = ref(false)
const dbTableCount = ref(0)
const showDataDialog = ref(false)
const pendingDataAction = ref(null)

const stepLabels = [
  { title: '许可协议', desc: '阅读并接受协议' },
  { title: '环境检测', desc: '检查运行环境' },
  { title: '数据库配置', desc: '配置数据库连接' },
  { title: '授权验证', desc: '验证授权密钥' },
  { title: '管理员账号', desc: '设置管理员信息' },
  { title: '完成安装', desc: '系统初始化' }
]

const mysqlStatus = ref('检测中')
const mysqlVersion = ref('')
const envChecked = ref(false)

const checkMysqlVersion = (version) => {
  if (!version) return { status: 'error', text: '未知版本' }
  const match = version.match(/^(\d+)\.(\d+)/)
  if (!match) return { status: 'error', text: version }
  const major = parseInt(match[1], 10)
  const minor = parseInt(match[2], 10)
  if (major >= 8) return { status: 'normal', text: version }
  if (major === 5 && minor >= 7) return { status: 'normal', text: version }
  return { status: 'error', text: version }
}

const envItems = computed(() => [
  { label: '操作系统', value: 'Linux', status: 'normal' },
  { label: 'Go版本', value: '1.21+', status: 'normal' },
  { label: 'MySQL', value: mysqlVersion.value || '待检测', status: mysqlStatus.value },
  { label: '内存', value: '2GB+', status: 'normal' }
])

const getTagColor = (status) => {
  if (status === 'normal' || status === '正常') return 'green'
  if (status === '检测中') return 'orange'
  if (status === 'not-installed' || status === '未安装') return 'gray'
  return 'red'
}

const getStatusText = (status) => {
  const map = {
    'normal': '正常',
    'error': '异常（建议升级至5.7+）',
    'not-installed': '未安装',
    '检测中': '检测中',
    '正常': '正常',
    '异常': '异常',
    '未安装': '未安装'
  }
  return map[status] || status
}

const runEnvCheck = async () => {
  if (envChecked.value) return
  envChecked.value = true
  mysqlStatus.value = '检测中'
  try {
    const res = await api.checkEnv()
    if (res.code === 0 && res.data) {
      const v = res.data.mysql_version || ''
      const s = res.data.mysql_status || 'not-installed'
      if (s === 'normal') {
        const result = checkMysqlVersion(v)
        mysqlVersion.value = result.text
        mysqlStatus.value = '正常'
      } else if (s === 'not-installed' || v === '未安装') {
        mysqlVersion.value = '未安装'
        mysqlStatus.value = '未安装'
      } else {
        mysqlVersion.value = v || '未知版本'
        mysqlStatus.value = '异常'
      }
    } else {
      mysqlVersion.value = '未安装'
      mysqlStatus.value = '未安装'
    }
  } catch (e) {
    mysqlVersion.value = '未安装'
    mysqlStatus.value = '未安装'
  }
}

const isMobile = ref(window.innerWidth < 768)

const handleResize = () => {
  isMobile.value = window.innerWidth < 768
}

onMounted(() => {
  window.addEventListener('resize', handleResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
})

const dbForm = reactive({
  host: 'localhost',
  port: 3306,
  database: 'chenze_faka',
  username: 'root',
  password: ''
})

const licenseForm = reactive({
  license_key: ''
})

const adminForm = reactive({
  username: 'admin',
  password: '',
  confirmPassword: ''
})

const testDatabaseConnection = async () => {
  if (!dbForm.host) {
    Message.warning('请输入数据库主机')
    return
  }
  if (!dbForm.port) {
    Message.warning('请输入端口')
    return
  }
  if (!dbForm.database) {
    Message.warning('请输入数据库名')
    return
  }
  if (!dbForm.username) {
    Message.warning('请输入用户名')
    return
  }

  testing.value = true
  dbTestResult.value = null
  dbTestSuccess.value = false
  try {
    const res = await api.testDatabase({
      host: dbForm.host,
      port: dbForm.port,
      database: dbForm.database,
      username: dbForm.username,
      password: dbForm.password
    })
    if (res.code === 0) {
      const ver = res.data?.version || ''
      const result = checkMysqlVersion(ver)
      mysqlVersion.value = result.text
      mysqlStatus.value = result.status

      const hasData = res.data?.has_data || false
      const tableCount = res.data?.table_count || 0
      dbHasData.value = hasData
      dbTableCount.value = tableCount

      if (hasData) {
        dbTestResult.value = {
          success: true,
          message: `连接正常，检测到数据库已有 ${tableCount} 张数据表`
        }
        showDataDialog.value = true
        pendingDataAction.value = null
      } else {
        dbTestSuccess.value = true
        dbTestResult.value = { success: true, message: '连接正常，数据库为空' }
        Message.success('数据库连接成功，数据库为空')
      }
    } else {
      dbTestSuccess.value = false
      dbTestResult.value = { success: false, message: res.message || '数据库连接失败' }
      mysqlVersion.value = ''
      mysqlStatus.value = '异常'
      Message.error(res.message || '数据库连接失败')
    }
  } catch (e) {
    dbTestSuccess.value = false
    dbTestResult.value = { success: false, message: e.message || '数据库连接失败' }
    mysqlVersion.value = ''
    mysqlStatus.value = '异常'
    Message.error(e.message || '数据库连接失败，请检查账号密码是否正确')
  } finally {
    testing.value = false
  }
}

const handleDataDialogAction = (action) => {
  showDataDialog.value = false
  pendingDataAction.value = action
  if (action === 'skip') {
    dbTestSuccess.value = true
    dbTestResult.value = {
      success: true,
      message: '连接正常，将跳过安装（保留现有数据）'
    }
    Message.info('已选择跳过安装，保留现有数据')
  } else if (action === 'overwrite') {
    dbTestSuccess.value = true
    dbTestResult.value = {
      success: true,
      message: '连接正常，将覆盖安装（清空数据）'
    }
    Message.warning('已选择覆盖安装，将清空现有数据')
  }
}

const handleSkipInstall = async () => {
  try {
    localStorage.setItem('site_config', JSON.stringify({ site_name: '晨泽发卡' }))
    await fetch('/api/install', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ skip: true })
    })
  } catch {}
  Message.success('已跳过安装，使用现有数据')
  router.push('/login')
}

const verifyLicense = async () => {
  if (!licenseForm.license_key) {
    Message.warning('请输入授权码')
    return
  }
  verifying.value = true
  try {
    const res = await api.verifyLicense({ license_key: licenseForm.license_key })
    if (res.data && res.data.verified) {
      licenseVerified.value = true
      Message.success('授权码验证通过')
    } else {
      licenseVerified.value = false
      Message.error('授权码无效或已过期')
    }
  } catch (e) {
    licenseVerified.value = false
    Message.error(e.message || '授权验证失败，请检查网络或授权码是否正确')
  } finally {
    verifying.value = false
  }
}

const nextStep = async () => {
  if (currentStep.value === 0 && !agreed.value) {
    Message.warning('请先同意许可协议')
    return
  }

  if (currentStep.value === 1 && !envChecked.value) {
    await runEnvCheck()
    if (mysqlStatus.value === '异常') {
      Message.warning('环境检测异常，请检查MySQL版本是否为5.7+')
    }
  }

  if (currentStep.value === 2 && !dbTestSuccess.value) {
    Message.warning('请先测试数据库连接成功')
    return
  }

  if (currentStep.value === 2 && pendingDataAction.value === 'skip') {
    try {
      await api.install({ skip: true })
      Message.success('跳过安装成功，正在跳转...')
      setTimeout(() => {
        router.push('/login')
      }, 1500)
      return
    } catch (e) {
      Message.error(e.message || '操作失败')
      return
    }
  }

  if (currentStep.value === 3 && !licenseVerified.value) {
    if (!licenseForm.license_key) {
      Message.warning('请先输入授权码')
      return
    }
    await verifyLicense()
    if (!licenseVerified.value) {
      return
    }
  }

  if (currentStep.value === 4) {
    if (!adminForm.username || !adminForm.password) {
      Message.warning('请填写完整信息')
      return
    }
    if (adminForm.password !== adminForm.confirmPassword) {
      Message.warning('两次密码不一致')
      return
    }
    if (adminForm.password.length < 6) {
      Message.warning('密码至少6位')
      return
    }

    installing.value = true
    try {
      await api.install({
        site_name: '晨泽发卡',
        license_key: licenseForm.license_key,
        database: {
          host: dbForm.host,
          port: dbForm.port,
          database: dbForm.database,
          username: dbForm.username,
          password: dbForm.password
        },
        jwt: {
          secret: '',
          expire_time: 72
        },
        pay: {
          type: 'epay',
          app_id: '',
          app_key: '',
          notify_url: ''
        },
        username: adminForm.username,
        password: adminForm.password,
        force: pendingDataAction.value === 'overwrite'
      })
      Message.success('安装成功')
      currentStep.value = 5
      setTimeout(() => {
        router.push('/login')
      }, 3000)
    } catch (e) {
      Message.error(e.message || '安装失败，请检查数据库配置和授权码')
      installing.value = false
    }
    return
  }

  currentStep.value++
}

const prevStep = () => {
  if (currentStep.value > 0) {
    currentStep.value--
    if (currentStep.value < 3) {
      licenseVerified.value = false
    }
  }
}

onMounted(async () => {
  try {
    const res = await api.getLicenseStatus()
    if (res.data && res.data.installed) {
      Message.info('系统已安装，正在跳转...')
      router.push('/login')
    }
  } catch (e) {
    // 未安装，继续显示安装页
  }
})

watch(currentStep, (step) => {
  if (step === 1 && !envChecked.value) {
    runEnvCheck()
  }
})
</script>

<style scoped>
.install-page {
  min-height: 100vh;
  background: linear-gradient(135deg, #f5f7fa 0%, #e4ecf7 100%);
  padding: 32px 20px;
  box-sizing: border-box;
}

.install-container {
  max-width: 800px;
  margin: 0 auto;
}

.install-header {
  text-align: center;
  margin-bottom: 28px;
}

.install-title {
  font-size: 28px;
  font-weight: 700;
  color: #1d2129;
  margin: 0 0 8px 0;
}

.install-desc {
  font-size: 15px;
  color: #86909c;
  margin: 0;
}

.install-steps {
  margin-bottom: 28px;
}

.mobile-steps {
  margin-bottom: 24px;
  padding: 16px;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
}

.mobile-step-info {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-bottom: 12px;
}

.mobile-step-num {
  font-size: 28px;
  font-weight: 700;
  color: #165dff;
  line-height: 1;
}

.mobile-step-total {
  font-size: 16px;
  color: #86909c;
  line-height: 1;
}

.mobile-step-title {
  margin-left: 8px;
  font-size: 16px;
  font-weight: 600;
  color: #1d2129;
}

.mobile-step-bar {
  display: flex;
  gap: 6px;
}

.mobile-step-dot {
  flex: 1;
  height: 4px;
  border-radius: 2px;
  background: #e5e6eb;
  transition: all 0.3s;
}

.mobile-step-dot.active {
  background: #165dff;
  height: 6px;
}

.mobile-step-dot.done {
  background: #00b42a;
}

.install-content {
  min-height: 280px;
  border-radius: 12px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.06);
  padding: 24px;
}

.step-content h3 {
  font-size: 18px;
  font-weight: 600;
  color: #1d2129;
  margin: 0 0 20px 0;
}

.env-check-list {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.env-check-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid #f2f3f5;
  flex-wrap: wrap;
  gap: 8px;
}

.env-check-item:last-child {
  border-bottom: none;
}

.env-check-label {
  font-size: 14px;
  color: #86909c;
  flex: 0 0 auto;
}

.env-check-value {
  font-size: 14px;
  color: #1d2129;
  flex: 1;
  text-align: center;
  min-width: 100px;
}

.agreement {
  background: #f7f8fa;
  padding: 16px;
  border-radius: 8px;
  margin-bottom: 20px;
  max-height: 200px;
  overflow-y: auto;
}

.agreement p {
  font-size: 14px;
  color: #4e5969;
  line-height: 1.8;
  margin: 0 0 8px 0;
}

.agreement-checkbox {
  font-size: 14px;
}

.test-db-section {
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid #f2f3f5;
}

.test-result {
  margin-top: 12px;
  padding: 10px 14px;
  border-radius: 6px;
  font-size: 14px;
  display: flex;
  align-items: center;
  gap: 6px;
}

.test-result.success {
  background: #e8ffea;
  color: #00a02b;
  border: 1px solid #b7eb8f;
}

.test-result.error {
  background: #ffece8;
  color: #f53f3f;
  border: 1px solid #ffa39e;
}

.verified-tip {
  margin-left: 12px;
  color: #00b42a;
  font-size: 14px;
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.license-tips {
  margin-top: 16px;
  padding: 12px;
  background: #f7f8fa;
  border-radius: 8px;
  color: #86909c;
  font-size: 13px;
}

.license-tips p {
  margin: 0 0 4px 0;
  line-height: 1.6;
}

.installing,
.install-success {
  text-align: center;
  padding: 40px 0;
}

.installing-text {
  margin-top: 16px;
  color: #4e5969;
  font-size: 15px;
}

.installing-desc {
  margin-top: 8px;
  color: #86909c;
  font-size: 13px;
}

.success-icon {
  font-size: 64px;
  color: #52c41a;
}

.install-actions {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  margin-top: 24px;
}

.install-actions .arco-btn {
  min-height: 44px;
  font-size: 15px;
}

@media (max-width: 768px) {
  .install-page {
    padding: 16px 12px;
  }

  .install-container {
    max-width: 100%;
  }

  .install-header {
    margin-bottom: 16px;
  }

  .install-title {
    font-size: 20px;
  }

  .install-desc {
    font-size: 13px;
  }

  .mobile-steps {
    margin-bottom: 16px;
    padding: 12px;
  }

  .mobile-step-num {
    font-size: 24px;
  }

  .mobile-step-total {
    font-size: 14px;
  }

  .mobile-step-title {
    font-size: 15px;
  }

  .install-content {
    padding: 16px;
    min-height: auto;
    border-radius: 8px;
  }

  .step-content h3 {
    font-size: 17px;
    margin-bottom: 16px;
  }

  .agreement {
    max-height: 150px;
    padding: 12px;
  }

  .agreement p {
    font-size: 13px;
  }

  .env-check-item {
    padding: 10px 12px;
    font-size: 13px;
  }

  .env-check-label {
    font-size: 13px;
    flex: 1 1 100%;
    width: 100%;
    margin-bottom: 4px;
  }

  .env-check-value {
    font-size: 13px;
    text-align: left;
    flex: 1;
    min-width: 0;
  }

  .db-form .arco-form-item {
    margin-bottom: 16px;
  }

  .install-actions {
    flex-direction: column;
    gap: 12px;
  }

  .install-actions .arco-btn {
    width: 100%;
    min-height: 48px;
    font-size: 16px;
  }

  .verified-tip {
    display: block;
    margin-top: 8px;
    margin-left: 0;
  }
}

@media (max-width: 480px) {
  .install-title {
    font-size: 18px;
  }

  .install-content {
    padding: 14px 12px;
  }

  .step-content h3 {
    font-size: 16px;
  }
}
</style>
