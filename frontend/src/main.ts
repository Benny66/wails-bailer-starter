import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import App from './App.vue'
import router from './router'
import { useAppStore } from './stores/app'
import { createVueErrorHandler, installGlobalErrorHandler, logInfo } from './lib/log'
import './styles/tokens.css'
import './styles/theme.css'
import './styles/element.css'
import './styles/index.css'

const app = createApp(App)

// 全局错误兜底必须在挂载【之前】装好，否则启动期的异常捕获不到。
// 生产态没有 DevTools，不装的话未捕获异常会彻底静默（既不显示也不留痕）。
installGlobalErrorHandler()
// Vue 组件树内的异常（渲染/生命周期）不在 window 事件覆盖范围内，单独挂。
app.config.errorHandler = createVueErrorHandler()

const pinia = createPinia()
app.use(pinia)
app.use(router)
app.use(ElementPlus)

// 挂载前同步初始主题到 <html>（await 确保 config 读取完成后再挂载，避免首屏闪烁）。
await useAppStore(pinia).initTheme()

app.mount('#app')

// 前端就绪标记：与 Go 的「应用启动」首行配对。
// 排查白屏时一眼可辨——日志里有「应用启动」但没有这一行，说明前端根本没起来。
logInfo('前端已挂载')
