<script lang="ts" setup>
import { onMounted, ref } from 'vue'

// 标题栏：
//  - 拖拽用 Wails 自定义属性 --wails-draggable: drag（非 Electron 的 -webkit-app-region）。
//  - CSS 自定义属性会继承，故按钮/控件必须显式 no-drag 覆盖，否则点击会误触发拖拽。
//  - Windows 自绘三按钮（最小化/最大化/关闭），通过 wails runtime 控制窗口。
//  - macOS 保留原生交通灯，不自绘按钮（内容顶到顶）。

// 是否 macOS：决定是否隐藏自绘三按钮
const isMac = navigator.userAgent.includes('Macintosh')

// 是否最大化：用于切换最大化/还原图标
const maximised = ref(false)

// 窗口控制：动态 import wails runtime，避免非 wails 环境报错
let runtime: typeof import('../../wailsjs/runtime/runtime') | null = null
async function getRuntime() {
  if (!runtime) {
    runtime = await import('../../wailsjs/runtime/runtime')
  }
  return runtime
}

async function minimise() {
  const rt = await getRuntime()
  rt.WindowMinimise()
}
async function toggleMaximise() {
  const rt = await getRuntime()
  rt.WindowToggleMaximise()
  maximised.value = !maximised.value
}
async function close() {
  const rt = await getRuntime()
  // 关闭按钮默认隐藏窗口（配合 HideWindowOnClose 实现"关到托盘"语义）
  rt.WindowHide()
}

onMounted(async () => {
  if (!isMac) {
    const rt = await getRuntime()
    // 读取初始最大化状态（尽力而为，失败不影响）
    try {
      maximised.value = await rt.WindowIsMaximised()
    } catch {
      /* 忽略 */
    }
  }
})
</script>

<template>
  <header class="titlebar" :class="{ 'titlebar--mac': isMac }">
    <div v-if="isMac" class="titlebar__traffic" />
    <div class="titlebar__drag" />

    <div v-if="!isMac" class="titlebar__controls">
      <button class="titlebar__btn" title="最小化" @click="minimise">—</button>
      <button class="titlebar__btn" :title="maximised ? '还原' : '最大化'" @click="toggleMaximise">
        {{ maximised ? '▣' : '□' }}
      </button>
      <button class="titlebar__btn titlebar__btn--close" title="关闭" @click="close">✕</button>
    </div>
  </header>
</template>

<style scoped>
.titlebar {
  position: relative;
  height: 40px;
  flex: none;
  display: flex;
  align-items: center;
  /* 整条标题栏作为拖拽区（Wails 自定义属性，会继承给子元素） */
  --wails-draggable: drag;
  user-select: none;
}

.titlebar__drag {
  flex: 1;
  height: 100%;
}

/* macOS 交通灯安全区：左上角留空，且不可拖拽（让交通灯可点击） */
.titlebar__traffic {
  flex: none;
  width: 72px;
  height: 100%;
  --wails-draggable: no-drag;
}

.titlebar__controls {
  display: flex;
  align-items: center;
  height: 100%;
  /* 按钮区排除拖拽（覆盖继承来的 drag） */
  --wails-draggable: no-drag;
}

.titlebar__btn {
  width: 46px;
  height: 100%;
  border: none;
  background: transparent;
  color: var(--text-2);
  font-size: 14px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background-color var(--motion-fast) var(--ease-standard);
}

.titlebar__btn:hover {
  background: var(--bg-hover);
  color: var(--text-1);
}

.titlebar__btn--close:hover {
  background: var(--color-danger);
  color: var(--color-on-danger);
}
</style>
