import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import axios from 'axios'
import App from './App.vue'
import router from './router'
import store from './store'

const token = localStorage.getItem('aiw-token')
if (token) {
  axios.defaults.headers.common['Authorization'] = `Bearer ${token}`
}

const isLoginPage = () => window.location.pathname.startsWith('/login')
const isSessionAuthRequest = config => {
  const url = String(config?.url || '')
  return url.includes('/auth/me') || url.includes('/auth/change-password')
}

let loginRedirecting = false

const redirectToLogin = () => {
  if (loginRedirecting || isLoginPage()) {
    return
  }
  loginRedirecting = true
  localStorage.removeItem('aiw-token')
  localStorage.removeItem('aiw-user')
  delete axios.defaults.headers.common['Authorization']
  window.location.href = '/login'
}

const handleAuthExpired = err => {
  if (err.response?.status === 401) {
    redirectToLogin()
  }
  return Promise.reject(err)
}

const attachAuthExpiredInterceptor = client => {
  client.interceptors.response.use(r => r, handleAuthExpired)
  return client
}

const createAxios = axios.create.bind(axios)
axios.create = (...args) => attachAuthExpiredInterceptor(createAxios(...args))
attachAuthExpiredInterceptor(axios)

const app = createApp(App)
app.use(ElementPlus)
app.use(router)
app.use(store)
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}
app.mount('#app')
