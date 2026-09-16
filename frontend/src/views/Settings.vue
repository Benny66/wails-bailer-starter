<script lang="ts" setup>
import { onMounted, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { useAppStore } from '../stores/app'
import { ExportDatabase, GetAppInfo, GetDataDir, OpenDataDir, SelectSaveFile } from '../../wailsjs/go/main/App'
import { invoke, type AppError } from '../lib/invoke'
import { logInfo } from '../lib/log'

// 设置页：主题切换 + 数据目录 + 应用信息与数据导出。
// 演示三件事：响应式状态落盘、原生目录能力、绑定方法的组合调用（选路径 → 导出）。
const store = useAppStore()
// 用 storeToRefs 建立真响应式，避免 ref(store.theme) 的断链快照。
const { theme } = storeToRefs(store)

function onThemeChange(val: 'dark' | 'light') {
  store.setTheme(val)
}

// 应用信息：版本/平台/路径。类型直接取自绑定返回值，避免手写镜像漂移。
type AppInfo = Awaited<ReturnType<typeof GetAppInfo>>
const info = ref<AppInfo | null>(null)
const actionError = ref<AppError | null>(null)
const exportHint = ref('')

/** 默认导出文件名：带日期，避免用户多次导出互相覆盖。 */
function defaultBackupName(): string {
  const d = new Date()
  const pad = (n: number) => String(n).padStart(2, '0')
  return `backup-${d.getFullYear()}${pad(d.getMonth() + 1)}${pad(d.getDate())}.db`
}

// 导出数据库：先经系统「保存」对话框取路径（用户在那里确认覆盖），再交给后端导出。
// 两个绑定各自单一职责，前端按流程组合——比做一个「一步到位」的绑定更灵活。
async function exportDatabase() {
  actionError.value = null
  exportHint.value = ''
  try {
    const target = await invoke(() => SelectSaveFile('导出数据库', defaultBackupName()))
    if (!target) return // 用户取消
    await invoke(() => ExportDatabase(target))
    exportHint.value = `已导出到 ${target}`
    logInfo('数据库已导出', { target })
  } catch (e) {
    actionError.value = e as AppError
  }
}

// 数据目录：数据库/配置/日志同目录。出问题时用户能自己把日志捞出来。
async function openDataDir() {
  actionError.value = null
  try {
    await invoke(() => OpenDataDir())
  } catch (e) {
    actionError.value = e as AppError
  }
}

// 双保险：AppInfo 已含数据目录，若它整条拿不到（极端情况），退回单独查。
onMounted(async () => {
  try {
    info.value = await invoke(() => GetAppInfo())
  } catch (e) {
    actionError.value = e as AppError
    try {
      info.value = { data_dir: await invoke(() => GetDataDir()) } as AppInfo
    } catch {
      // 两次都失败：错误已记录，界面展示占位即可
    }
  }
})
</script>

<template>
  <section class="page">
    <h1 class="page__title">设置</h1>

    <div class="setting">
      <span class="setting__label">主题</span>
      <button
        class="setting__btn"
        :class="{ 'setting__btn--active': theme === 'dark' }"
        @click="onThemeChange('dark')"
      >
        暗色
      </button>
      <button
        class="setting__btn"
        :class="{ 'setting__btn--active': theme === 'light' }"
        @click="onThemeChange('light')"
      >
        亮色
      </button>
    </div>

    <div class="setting">
      <span class="setting__label">数据目录</span>
      <span class="setting__path" :title="info?.data_dir">{{ info?.data_dir || '—' }}</span>
      <button class="setting__btn" @click="openDataDir">打开</button>
    </div>

    <div class="setting">
      <span class="setting__label">版本</span>
      <span class="setting__path">{{ info?.version || '—' }}</span>
      <span class="setting__muted">{{ info?.platform || '' }}</span>
    </div>

    <div class="setting">
      <span class="setting__label">日志文件</span>
      <span class="setting__path" :title="info?.log_file">{{ info?.log_file || '—' }}</span>
    </div>

    <div class="setting">
      <span class="setting__label">数据备份</span>
      <button class="setting__btn" @click="exportDatabase">导出数据库…</button>
      <span v-if="exportHint" class="setting__muted" :title="exportHint">已导出</span>
    </div>

    <p v-if="actionError" class="setting__error">[{{ actionError.code }}] {{ actionError.message }}</p>
  </section>
</template>

<style scoped>
.page__title {
  font-size: var(--font-size-xl);
  margin: 0 0 var(--space-4);
}
.setting {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  margin-bottom: var(--space-3);
}
.setting__label {
  flex: none;
  width: 72px;
  color: var(--text-2);
}
.setting__path {
  max-width: 420px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--text-3);
  font-size: var(--font-size-sm);
}
.setting__muted {
  color: var(--text-3);
  font-size: var(--font-size-sm);
}
.setting__error {
  color: var(--color-danger);
}
.setting__btn {
  flex: none;
  border: 1px solid var(--border-subtle);
  background: var(--bg-raised);
  color: var(--text-2);
  padding: var(--space-2) var(--space-4);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all var(--motion-fast) var(--ease-standard);
}
.setting__btn--active {
  border-color: var(--color-primary);
  color: var(--color-primary);
  background: var(--color-primary-soft);
}
</style>
