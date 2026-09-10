import js from '@eslint/js'
import vueParser from 'vue-eslint-parser'
import tsParser from '@typescript-eslint/parser'

// ==========================================================================
// 前端护栏（flat config）
// --------------------------------------------------------------------------
// 两条自定义规则把宪法铁律编译成会红的检查：
//   - no-node-imports:    渲染层禁 import node 能力，必须走 wails bindings。
//   - no-hardcoded-brand: 禁硬编码 hex 色值（应引 tokens）与品牌字符串。
// ==========================================================================

/** @type {import('eslint').Linter.FlatConfig[]} */
export default [
  {
    ignores: ['dist/**', 'wailsjs/**', 'node_modules/**'],
  },

  // 基础 JS 规则
  js.configs.recommended,

  // Vue 文件用 vue-eslint-parser 解析，script 块用 TS parser
  {
    files: ['**/*.vue'],
    languageOptions: {
      parser: vueParser,
      parserOptions: {
        parser: tsParser,
        sourceType: 'module',
        ecmaVersion: 'latest',
      },
      globals: {
        window: 'readonly',
        document: 'readonly',
        navigator: 'readonly',
        console: 'readonly',
        localStorage: 'readonly',
      },
    },
  },
  {
    files: ['**/*.ts', '**/*.js'],
    languageOptions: {
      parser: tsParser,
      sourceType: 'module',
      ecmaVersion: 'latest',
      globals: {
        window: 'readonly',
        document: 'readonly',
        navigator: 'readonly',
        console: 'readonly',
        localStorage: 'readonly',
      },
    },
  },

  // 自定义规则
  {
    files: ['src/**/*.vue', 'src/**/*.ts', 'src/**/*.js'],
    plugins: {
      guard: {
        rules: {
          'no-node-imports': { create: noNodeImportsRule },
          'no-hardcoded-brand': { create: noHardcodedBrandRule },
        },
      },
    },
    rules: {
      'guard/no-node-imports': 'error',
      'guard/no-hardcoded-brand': 'error',
    },
  },
]

// ---------------------------------------------------------------------------
// 规则 1：渲染层禁 import node 能力
// ---------------------------------------------------------------------------
function noNodeImportsRule(context) {
  const forbidden = /^(node:|\b(fs|child_process|path|os|crypto|http|https|net|tls|dns|dgram|cluster|worker_threads|process)\b)/

  return {
    ImportDeclaration(node) {
      const src = node.source.value
      if (typeof src === 'string' && forbidden.test(src)) {
        context.report({
          node,
          message: `渲染层禁止 import node 能力 "${src}"，必须经 wails bindings（preload 受控 API）`,
        })
      }
    },
  }
}

// ---------------------------------------------------------------------------
// 规则 2：禁硬编码 hex 色值（应引 tokens）
//   - 豁免 src/styles/ 目录（那是令牌真相源，本就该有 hex 定义）
//   - 全文件源码扫描（含 .vue 的 <style> 块）
// ---------------------------------------------------------------------------
function noHardcodedBrandRule(context) {
  const filename = context.filename || context.getPhysicalFilename?.() || ''

  return {
    Program(node) {
      // 令牌/主题定义文件豁免
      if (filename.includes('/styles/')) return

      const source = context.sourceCode.getText()
      const hexRegex = /#(?:[0-9a-fA-F]{3,4}|[0-9a-fA-F]{6}|[0-9a-fA-F]{8})\b/g

      let m
      while ((m = hexRegex.exec(source)) !== null) {
        context.report({
          node,
          loc: context.sourceCode.getLocFromIndex(m.index),
          message: `禁止硬编码色值 "${m[0]}"，应引用设计令牌 var(--token-*)`,
        })
      }
    },
  }
}
