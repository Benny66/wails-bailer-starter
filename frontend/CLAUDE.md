# CLAUDE.md — Vue 域规范

本文件约束 `frontend/` 下 Vue3 代码。AI 在前端目录干活时读本份。
跨端铁律见根 `AGENTS.md`。

## 命名

| 元素 | 规则 | 示例 |
|---|---|---|
| 组件文件 | PascalCase | `UserList.vue`、`Layout.vue` |
| 页面目录 | kebab-case | `views/system/user/` |
| JS 模块 | camelCase | `request.js`、`format.js` |
| 组合式函数 | useXxx | `useAppStore` |
| 变量 | 小驼峰 | `userList`、`loading` |

## 组件风格

- 使用 `<script setup>` 语法。
- 模板里复杂表达式抽成计算属性或方法。
- 每个页面组件：`<template>` → `<script setup>` → `<style scoped>` 三段式。

## 状态管理

MUST 统一使用 `stores/`（Pinia 约定），禁止 `store/` 单复数混用。

## 调用 Go 后端

- 前端调用 Go 方法走 wails 自动生成的绑定 `frontend/wailsjs/go/...`，
  不写 `http` 请求。
- 调用前确认绑定已随 Go 侧方法更新（`make dev` 时 wails 自动重新生成）。

## 模块规范

- 工具函数导出用命名导出，一个文件一个职责。
- API 定义集中，按模块注释分组。
