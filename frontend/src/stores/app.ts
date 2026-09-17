import { defineStore } from 'pinia'
import { ref } from 'vue'
import { GetTheme, SetTheme } from '../../wailsjs/go/main/App'
import { invoke } from '../lib/invoke'
import { logError } from '../lib/log'

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
  // 初始值先取系统偏好占位，真正的"config 优先"在 initTheme 里完成。
  const theme = ref<Theme>(systemPrefersDark() ? 'dark' : 'light')
  const sidebarCollapsed = ref(false)

  function setTheme(next: Theme) {
    theme.value = next
    document.documentElement.setAttribute('data-theme', next)
    // 异步落盘到 config.json（Go 侧），失败不阻断切换。
    // 经 invoke 包装：错误归一化为 AppError 并写进 app.log —— 原来只 console.error，
    // 而打包后的应用没有 DevTools，那条错误等于没记。
    invoke(() => SetTheme(next), 'SetTheme').catch((err) => {
      logError('主题持久化失败', err)
    })
  }

  // 启动时同步主题：config 优先，空串 fallback 到系统偏好。
  // 这是"前端调 Go"的活范例：GetTheme 读 config，空串表示"未设置"。
  async function initTheme() {
    let resolved: Theme
    try {
      // 经 invoke：这一步在挂载之前，是启动时间线上唯一的一次后端往返，
      // 故必须进 IPC 统计（见 main.ts 的启动分段）
      const saved = await invoke(() => GetTheme(), 'GetTheme')
      resolved = saved === 'dark' || saved === 'light' ? saved : (systemPrefersDark() ? 'dark' : 'light')
    } catch {
      // 非 wails 环境（纯 vite dev）或调用失败，fallback 系统偏好
      resolved = systemPrefersDark() ? 'dark' : 'light'
    }
    theme.value = resolved
    document.documentElement.setAttribute('data-theme', resolved)
  }

  function toggleSidebar() {
    sidebarCollapsed.value = !sidebarCollapsed.value
  }

  return { theme, sidebarCollapsed, setTheme, initTheme, toggleSidebar }
})
