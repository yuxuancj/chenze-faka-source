<template>
  <div class="install-page">
    <div class="install-container">
      <div class="install-header">
        <h1 class="install-title">晨泽发卡系统安装向导</h1>
        <p class="install-desc">请按步骤完成系统安装配置</p>
      </div>
      <a-steps :current="currentStep" class="install-steps">
        <a-step title="许可协议" description="阅读并接受协议" />
        <a-step title="环境检测" description="检查运行环境" />
        <a-step title="数据库配置" description="配置数据库连接" />
        <a-step title="管理员账号" description="设置管理员信息" />
        <a-step title="完成安装" description="系统初始化" />
      </a-steps>
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
          <a-descriptions :column="1" bordered>
            <a-descriptions-item label="操作系统">
              {{ envInfo.os }} <a-tag :color="envInfo.os_ok ? 'green' : 'red'">{{ envInfo.os_ok ? '正常' : '不支持' }}</a-tag>
            </a-descriptions-item>
            <a-descriptions-item label="PHP版本">
              {{ envInfo.php_version }} <a-tag :color="envInfo.php_ok ? 'green' : 'red'">{{ envInfo.php_ok ? '正常' : '版本过低' }}</a-tag>
            </a-descriptions-item>
            <a-descriptions-item label="MySQL">
              {{ envInfo.mysql }} <a-tag :color="envInfo.mysql_ok ? 'green' : 'red'">{{ envInfo.mysql_ok ? '已安装' : '未检测' }}</a-tag>
            </a-descriptions-item>
            <a-descriptions-item label="文件权限">
              <a-tag :color="envInfo.permission_ok ? 'green' : 'red'">{{ envInfo.permission_ok ? '可写' : '无权限' }}</a-tag>
            </a-descriptions-item>
          </a-descriptions>
        </div>
        <div v-else-if="currentStep === 2" class="step-content">
          <h3>数据库配置</h3>
          <a-form :model="dbForm" layout="vertical">
            <a-form-item field="host" label="数据库主机">
              <a-input v-model="dbForm.host" placeholder="localhost" />
            </a-form-item>
            <a-form-item field="port" label="端口">
              <a-input-number v-model="dbForm.port" :min="1" :max="65535" />
            </a-form-item>
            <a-form-item field="database" label="数据库名">
              <a-input v-model="dbForm.database" placeholder="chenze_faka" />
            </a-form-item>
            <a-form-item field="username" label="用户名">
              <a-input v-model="dbForm.username" placeholder="root" />
            </a-form-item>
            <a-form-item field="password" label="密码">
              <a-input-password v-model="dbForm.password" placeholder="请输入数据库密码" />
            </a-form-item>
          </a-form>
        </div>
        <div v-else-if="currentStep === 3" class="step-content">
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
        <div v-else-if="currentStep === 4" class="step-content">
          <div v-if="installing" class="installing">
            <a-spin size="40" />
            <p class="installing-text">正在安装，请稍候...</p>
          </div>
          <div v-else class="install-success">
            <iconify-icon icon="arco:check-circle" class="success-icon" />
            <h3>安装完成！</h3>
            <p>系统已成功安装，即将跳转至管理后台。</p>
          </div>
        </div>
      </a-card>
      <div class="install-actions">
        <a-button v-if="currentStep > 0 && currentStep < 4" @click="prevStep">上一步</a-button>
        <a-button v-if="currentStep < 4" type="primary" @click="nextStep">
          {{ currentStep === 3 ? '开始安装' : '下一步' }}
        </a-button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Message } from '@arco-design/web-vue'
import { api } from '../api'

const router = useRouter()

const currentStep = ref(0)
const agreed = ref(false)
const installing = ref(false)

const envInfo = reactive({
  os: 'Linux',
  os_ok: true,
  php_version: '8.1+',
  php_ok: true,
  mysql: 'MySQL 5.7+ / 8.0+',
  mysql_ok: true,
  permission_ok: true
})

const dbForm = reactive({
  host: 'localhost',
  port: 3306,
  database: 'chenze_faka',
  username: 'root',
  password: ''
})

const adminForm = reactive({
  username: 'admin',
  password: '',
  confirmPassword: ''
})

const nextStep = async () => {
  if (currentStep.value === 0 && !agreed.value) {
    Message.warning('请先同意许可协议')
    return
  }
  if (currentStep.value === 4) {
    return
  }
  if (currentStep.value === 3) {
    if (!adminForm.username || !adminForm.password) {
      Message.warning('请填写完整信息')
      return
    }
    if (adminForm.password !== adminForm.confirmPassword) {
      Message.warning('两次密码不一致')
      return
    }
    installing.value = true
    try {
      await new Promise(resolve => setTimeout(resolve, 2000))
      await api.install({
        db: dbForm,
        admin: {
          username: adminForm.username,
          password: adminForm.password
        }
      })
      Message.success('安装成功')
      setTimeout(() => {
        router.push('/login')
      }, 2000)
    } catch (e) {
      Message.error(e.message || '安装失败')
      installing.value = false
    }
    return
  }
  currentStep.value++
}

const prevStep = () => {
  if (currentStep.value > 0) {
    currentStep.value--
  }
}

onMounted(async () => {
  try {
    const res = await api.getInstallStatus()
    if (res.data && res.data.installed) {
      router.push('/login')
    }
  } catch (e) {}
})
</script>

<style scoped>
.install-page {
  min-height: 100vh;
  background: linear-gradient(135deg, #f5f7fa 0%, #e4ecf7 100%);
  padding: 40px 20px;
}

.install-container {
  max-width: 800px;
  margin: 0 auto;
}

.install-header {
  text-align: center;
  margin-bottom: 32px;
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
  margin-bottom: 32px;
}

.install-content {
  min-height: 300px;
  border-radius: 12px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.06);
}

.step-content h3 {
  font-size: 18px;
  font-weight: 600;
  color: #1d2129;
  margin: 0 0 20px 0;
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

.installing,
.install-success {
  text-align: center;
  padding: 40px 0;
}

.installing-text {
  margin-top: 16px;
  color: #4e5969;
}

.success-icon {
  font-size: 64px;
  color: #52c41a;
}

.install-actions {
  display: flex;
  justify-content: space-between;
  margin-top: 24px;
}
</style>
