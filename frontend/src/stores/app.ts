import { defineStore } from 'pinia'
import { ref } from 'vue'

type Theme = 'dark' | 'light'

// 检测系统偏好（含空值兜底：webview 无 matchMedia 时默认暗色）。
function systemPrefersDark(): boolean {
  try {
    return window.matchMedia?.('(prefers-color-scheme: dark)').matches ?? true
  } catch {
    return true
  }
}

// 应用全局状态：主题 + 侧栏折叠。
export const useAppStore = defineStore('app', () => {
  // 默认跟随系统：初始主题 = 系统偏好，而非硬编码 dark。
  const theme = ref<Theme>(systemPrefersDark() ? 'dark' : 'light')
  const sidebarCollapsed = ref(false)

  function setTheme(next: Theme) {
    theme.value = next
    document.documentElement.setAttribute('data-theme', next)
  }

  // 启动时同步一次 data-theme，消除"store 有值但 DOM 无属性"的脱节。
  function initTheme() {
    document.documentElement.setAttribute('data-theme', theme.value)
  }

  function toggleSidebar() {
    sidebarCollapsed.value = !sidebarCollapsed.value
  }

  return { theme, sidebarCollapsed, setTheme, initTheme, toggleSidebar }
})
