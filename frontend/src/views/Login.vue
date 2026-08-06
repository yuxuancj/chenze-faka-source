<template>
  <div class="login-page">
    <div class="login-container">
      <div class="login-left">
        <div class="brand">
          <h1 class="brand-title">{{ siteName }}</h1>
          <p class="brand-slogan">专业的数字商品发卡平台</p>
        </div>
        <div class="features">
          <div class="feature-item">
            <iconify-icon icon="arco:check-circle" />
            <span>7x24小时自动发货</span>
          </div>
          <div class="feature-item">
            <iconify-icon icon="arco:check-circle" />
            <span>海量商品源头直供</span>
          </div>
          <div class="feature-item">
            <iconify-icon icon="arco:check-circle" />
            <span>安全可靠交易保障</span>
          </div>
        </div>
      </div>
      <div class="login-right">
        <a-card class="login-card" :bordered="false">
          <h2 class="login-title">管理员登录</h2>
          <p class="login-subtitle">欢迎回来，请登录您的账号</p>
          <a-form :model="loginForm" layout="vertical" class="login-form" @submit="handleLogin">
            <a-form-item field="username" label="用户名">
              <a-input
                v-model="loginForm.username"
                placeholder="请输入用户名"
                size="large"
              >
                <template #prefix>
                  <iconify-icon icon="arco:user" />
                </template>
              </a-input>
            </a-form-item>
            <a-form-item field="password" label="密码">
              <a-input-password
                v-model="loginForm.password"
                placeholder="请输入密码"
                size="large"
              >
                <template #prefix>
                  <iconify-icon icon="arco:lock" />
                </template>
              </a-input-password>
            </a-form-item>
            <a-form-item field="captcha" label="验证码">
              <a-row :gutter="8">
                <a-col :span="16">
                  <a-input v-model="loginForm.captcha" placeholder="请输入验证码" size="large">
                    <template #prefix>
                      <iconify-icon icon="arco:shield" />
                    </template>
                  </a-input>
                </a-col>
                <a-col :span="8">
                  <div class="captcha-box" @click="refreshCaptcha">
                    <canvas ref="captchaCanvas" width="100" height="40"></canvas>
                  </div>
                </a-col>
              </a-row>
            </a-form-item>
            <a-button
              type="primary"
              size="large"
              long
              :loading="loading"
              html-type="submit"
            >
              登录
            </a-button>
          </a-form>
        </a-card>
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
const siteName = ref('晨泽发卡')
const loading = ref(false)
const captchaCanvas = ref(null)

const loginForm = reactive({
  username: '',
  password: '',
  captcha: ''
})

const generateCaptcha = () => {
  const canvas = captchaCanvas.value
  if (!canvas) return
  const ctx = canvas.getContext('2d')
  ctx.fillStyle = '#f0f0f0'
  ctx.fillRect(0, 0, canvas.width, canvas.height)
  const chars = 'ABCDEFGHJKMNPQRSTUVWXYZ23456789'
  let captcha = ''
  for (let i = 0; i < 4; i++) {
    captcha += chars[Math.floor(Math.random() * chars.length)]
  }
  ctx.font = 'bold 24px Arial'
  ctx.fillStyle = '#2a7fff'
  ctx.textBaseline = 'middle'
  for (let i = 0; i < captcha.length; i++) {
    ctx.save()
    const x = 20 + i * 20
    const y = canvas.height / 2
    const angle = (Math.random() - 0.5) * 0.4
    ctx.translate(x, y)
    ctx.rotate(angle)
    ctx.fillText(captcha[i], 0, 0)
    ctx.restore()
  }
  for (let i = 0; i < 3; i++) {
    ctx.beginPath()
    ctx.moveTo(Math.random() * canvas.width, Math.random() * canvas.height)
    ctx.lineTo(Math.random() * canvas.width, Math.random() * canvas.height)
    ctx.strokeStyle = '#ccc'
    ctx.stroke()
  }
  loginForm._captcha = captcha
}

const refreshCaptcha = () => {
  generateCaptcha()
}

const handleLogin = async () => {
  if (!loginForm.username) {
    Message.warning('请输入用户名')
    return
  }
  if (!loginForm.password) {
    Message.warning('请输入密码')
    return
  }
  if (loginForm.captcha.toUpperCase() !== loginForm._captcha) {
    Message.warning('验证码错误')
    refreshCaptcha()
    return
  }
  loading.value = true
  try {
    const res = await api.login({
      username: loginForm.username,
      password: loginForm.password
    })
    const data = res.data || res
    if (data && data.token) {
      localStorage.setItem('admin_token', data.token)
      if (data.user) {
        localStorage.setItem('admin_user', JSON.stringify(data.user))
      }
      Message.success('登录成功')
      router.push('/admin/dashboard')
    } else {
      Message.error('登录响应数据异常')
    }
  } catch (e) {
    Message.error(e.message || '登录失败')
    refreshCaptcha()
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  generateCaptcha()
})
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  background: #f5f7fa;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}

.login-container {
  width: 100%;
  max-width: 900px;
  background: #fff;
  border-radius: 16px;
  box-shadow: 0 8px 40px rgba(0, 0, 0, 0.08);
  overflow: hidden;
  display: flex;
}

.login-left {
  flex: 1;
  background: linear-gradient(135deg, #2a7fff 0%, #1a5fcc 100%);
  padding: 60px 40px;
  color: #fff;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.brand-title {
  font-size: 32px;
  font-weight: 700;
  margin: 0 0 8px 0;
}

.brand-slogan {
  font-size: 16px;
  opacity: 0.9;
  margin: 0 0 40px 0;
}

.features {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.feature-item {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 14px;
}

.feature-item iconify-icon {
  font-size: 20px;
}

.login-right {
  flex: 1;
  padding: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.login-card {
  width: 100%;
  max-width: 360px;
  background: transparent;
  box-shadow: none;
}

.login-title {
  font-size: 24px;
  font-weight: 600;
  color: #1d2129;
  margin: 0 0 8px 0;
}

.login-subtitle {
  font-size: 14px;
  color: #86909c;
  margin: 0 0 32px 0;
}

.login-form .arco-input,
.login-form .arco-input-password {
  height: 44px;
}

.captcha-box {
  border: 1px solid #e5e6eb;
  border-radius: 4px;
  cursor: pointer;
  overflow: hidden;
  height: 40px;
}

.captcha-box canvas {
  display: block;
}

/* 移动端适配 */
@media (max-width: 768px) {
  .login-page {
    padding: 0;
    align-items: stretch;
  }

  .login-container {
    flex-direction: column;
    border-radius: 0;
    box-shadow: none;
    min-height: 100vh;
  }

  .login-left {
    flex: 0 0 auto;
    padding: 32px 24px;
    text-align: center;
  }

  .brand-title {
    font-size: 24px;
  }

  .brand-slogan {
    font-size: 14px;
    margin-bottom: 20px;
  }

  .features {
    display: none;
  }

  .login-right {
    flex: 1;
    padding: 24px 20px;
    align-items: flex-start;
  }

  .login-card {
    max-width: 100%;
  }

  .login-title {
    font-size: 20px;
  }

  .login-subtitle {
    font-size: 13px;
    margin-bottom: 24px;
  }
}
</style>
