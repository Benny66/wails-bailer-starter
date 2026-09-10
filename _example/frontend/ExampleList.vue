<script lang="ts" setup>
import { onMounted, ref } from 'vue'
import { ListExamples } from '../../../wailsjs/go/main/App'
import { model } from '../../../wailsjs/go/models'

// TODO: 业务逻辑 —— 表单字段、校验、编辑/删除交互。
const list = ref<model.Example[]>([])
const loading = ref(false)

async function load() {
  loading.value = true
  try {
    list.value = await ListExamples()
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <section class="page">
    <h1 class="page__title">Example 列表</h1>
    <!-- TODO: 业务逻辑 —— 用 EP 的 el-table / el-form / el-pagination 展示 -->
    <p v-if="loading">加载中…</p>
    <p v-else-if="!list.length" class="page__desc">暂无数据</p>
    <ul v-else>
      <li v-for="item in list" :key="item.id">{{ item.id }}</li>
    </ul>
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
</style>
