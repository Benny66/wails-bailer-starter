import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import App from './App.vue'
import router from './router'
import { useAppStore } from './stores/app'
import './styles/tokens.css'
import './styles/theme.css'
import './styles/element.css'
import './styles/index.css'

const app = createApp(App)
const pinia = createPinia()
app.use(pinia)
app.use(router)
app.use(ElementPlus)

// 挂载前同步初始主题到 <html>（await 确保 config 读取完成后再挂载，避免首屏闪烁）。
await useAppStore(pinia).initTheme()

app.mount('#app')
