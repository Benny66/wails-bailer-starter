<script lang="ts" setup>
import { onMounted, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { useAppStore } from '../stores/app'
import { GetDataDir, OpenDataDir } from '../../wailsjs/go/main/App'
import { invoke, type AppError } from '../lib/invoke'

// 设置页：主题切换演示 + 数据目录入口（验证 data-theme 与绑定调用）。
const store = useAppStore()
// 用 storeToRefs 建立真响应式，避免 ref(store.theme) 的断链快照。
const { theme } = storeToRefs(store)

function onThemeChange(val: 'dark' | 'light') {
  store.setTheme(val)
}

// 数据目录：数据库/配置/日志同目录。展示路径 + 一键在文件管理器中打开，
// 出问题时用户能自己把日志捞出来，不必靠开发者口头教各平台路径。
const dataDir = ref('')
const actionError = ref<AppError | null>(null)

async function openDataDir() {
  actionError.value = null
  try {
    // invoke 把 Go 的业务错误归一化为 {code, message}，按 code 分流。
    await invoke(() => OpenDataDir())
  } catch (e) {
    actionError.value = e as AppError
  }
}

onMounted(async () => {
  try {
    dataDir.value = await invoke(() => GetDataDir())
  } catch (e) {
    actionError.value = e as AppError
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
      <span class="setting__path" :title="dataDir">{{ dataDir || '—' }}</span>
      <button class="setting__btn" @click="openDataDir">打开</button>
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
