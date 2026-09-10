import { defineStore } from 'pinia'
import { ref } from 'vue'

type Theme = 'dark' | 'light'

// 应用全局状态：主题 + 侧栏折叠。
export const useAppStore = defineStore('app', () => {
  const theme = ref<Theme>('dark')
  const sidebarCollapsed = ref(false)

  function setTheme(next: Theme) {
    theme.value = next
    document.documentElement.setAttribute('data-theme', next)
  }

  function toggleSidebar() {
    sidebarCollapsed.value = !sidebarCollapsed.value
  }

  return { theme, sidebarCollapsed, setTheme, toggleSidebar }
})
