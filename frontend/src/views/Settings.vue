<script lang="ts" setup>
import { storeToRefs } from 'pinia'
import { useAppStore } from '../stores/app'

// 设置页：占位 + 主题切换演示（验证 data-theme 生效）。
const store = useAppStore()
// 用 storeToRefs 建立真响应式，避免 ref(store.theme) 的断链快照。
const { theme } = storeToRefs(store)

function onThemeChange(val: 'dark' | 'light') {
  store.setTheme(val)
}
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
}
.setting__label {
  color: var(--text-2);
}
.setting__btn {
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
