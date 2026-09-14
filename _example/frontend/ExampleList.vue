<script lang="ts" setup>
import { onMounted, ref } from 'vue'
import { ListExamples } from '../../../wailsjs/go/main/App'
import { invoke, type AppError } from '../../../src/lib/invoke'

// TODO: 业务逻辑 —— 表单字段、校验、编辑/删除交互。
// 本页是黄金范例：演示管道契约的正确用法（分页 + 错误归一化）。
// 表格/表单等 UI 属产品面，由下游自行实现，此处的 <ul> 仅示意数据形状。

// 分页状态。page_size 传 0 时 Go 侧会归一化为 DefaultPageSize。
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref<{ id: number }[]>([])
const loading = ref(false)
const error = ref<AppError | null>(null)

async function load() {
  loading.value = true
  error.value = null
  try {
    // invoke 把 Go 的业务错误归一化为 {code, message}，按 code 分流。
    const res = await invoke(() =>
      ListExamples({ page: page.value, page_size: pageSize.value }),
    )
    list.value = res.list
    total.value = res.total
    // Go 侧回填了归一化后的 page/page_size，前端以回显值为准。
    page.value = res.page
    pageSize.value = res.page_size
  } catch (e) {
    error.value = e as AppError
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <section class="page">
    <h1 class="page__title">Example 列表</h1>

    <!-- 错误呈现：据 code 分流（此处仅示意，toast 等 UI 由下游实现） -->
    <p v-if="error" class="page__error">[{{ error.code }}] {{ error.message }}</p>

    <p v-if="loading">加载中…</p>
    <p v-else-if="!list.length" class="page__desc">暂无数据</p>
    <ul v-else>
      <li v-for="item in list" :key="item.id">{{ item.id }}</li>
    </ul>

    <!-- 分页信息：total 来自分页协议，前端据此渲染分页控件 -->
    <p class="page__total">共 {{ total }} 条，第 {{ page }} 页</p>
  </section>
</template>

<style scoped>
.page__title {
  font-size: var(--font-size-xl);
  margin: 0 0 var(--space-4);
}
.page__desc {
  color: var(--text-2);
}
.page__error {
  color: var(--color-danger);
}
.page__total {
  color: var(--text-2);
  font-size: var(--font-size-sm);
}
</style>
