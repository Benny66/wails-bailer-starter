// 管道契约：前端日志出口与全局错误兜底。
//
// 为什么必须有它：打包后的应用【没有 DevTools】。`console.*` 在开发态看着挺好，
// 生产态是一处无人可见的输出——用户遇到的异常就此消失，只能收到一句「它不好使了」。
//
// Wails 提供了 LogDebug/LogInfo/LogWarning/LogError，它们经 WailsInvoke 送到 Go，
// 由 Go 侧 logger 写进 app.log（接线在 main.go 的 options.App.Logger，
// 由 internal/guard/wiring_test.go 强制——漏接不会有任何症状，日志会静默丢失）。
// 因此本模块是前端日志【唯一】的正确出口；console 只在非 wails 环境（纯 vite dev）兜底。
//
// 本文件不放任何 UI：异常如何呈现给用户（toast/弹窗）是产品面，由下游决定。
// 需要接管时，用 installGlobalErrorHandler 的 hook 参数。

import { LogDebug, LogError, LogInfo, LogWarning } from '../../wailsjs/runtime/runtime'
import { normalizeError, type AppError } from './invoke'

/** 日志级别。与 Go 侧 slog 的级别对应（debug/info/warn/error）。 */
export type LogLevel = 'debug' | 'info' | 'warn' | 'error'

/** 单条日志中 detail 的最大长度，超出截断——避免把巨大对象灌进日志文件。 */
const MAX_DETAIL_LENGTH = 2000

/** 把附加上下文序列化为可读文本；循环引用等异常情况退化为 String()。 */
function formatDetail(detail: unknown): string {
  if (detail === undefined || detail === null) return ''
  try {
    const text = typeof detail === 'string' ? detail : JSON.stringify(detail)
    if (text === undefined) return String(detail)
    return text.length > MAX_DETAIL_LENGTH ? `${text.slice(0, MAX_DETAIL_LENGTH)}…(已截断)` : text
  } catch {
    // JSON.stringify 对循环引用会抛错，此时退回最朴素的形式
    return String(detail)
  }
}

/**
 * 组装成单行文本。日志是行式存储，多行内容会破坏解析，故统一压成一行。
 * 不带级别前缀——级别由后端 logger 记录，写进消息里是重复。
 */
function compose(message: string, detail?: unknown): string {
  const tail = formatDetail(detail)
  const line = tail ? `${message} | ${tail}` : message
  return line.replace(/\s*\n\s*/g, ' ⏎ ')
}

/** 写入后端日志文件；非 wails 环境退化为 console，保证日志不丢。 */
function write(level: LogLevel, line: string): void {
  try {
    switch (level) {
      case 'debug':
        LogDebug(line)
        return
      case 'info':
        LogInfo(line)
        return
      case 'warn':
        LogWarning(line)
        return
      case 'error':
        LogError(line)
        return
    }
  } catch {
    // 非 wails 环境（纯 vite dev / 单测）没有注入运行时，退化到 console
  }
  const fallback = level === 'warn' ? console.warn : level === 'error' ? console.error : console.log
  fallback(line)
}

/** 写一条 debug 日志。 */
export function logDebug(message: string, detail?: unknown): void {
  write('debug', compose(message, detail))
}

/** 写一条 info 日志。 */
export function logInfo(message: string, detail?: unknown): void {
  write('info', compose(message, detail))
}

/** 写一条警告日志。 */
export function logWarn(message: string, detail?: unknown): void {
  write('warn', compose(message, detail))
}

/** 写一条错误日志。 */
export function logError(message: string, detail?: unknown): void {
  write('error', compose(message, detail))
}

/**
 * 上报一个异常：归一化为 AppError 后写入日志，并交给可选的 hook。
 *
 * @param raw 任意抛出物（Error / AppError / 字符串 / 其它）
 * @param context 出错位置的人类可读描述（如 'Vue 渲染'、'未处理的 Promise 拒绝'）
 * @param hook 下游接管入口（toast 等），不传则只记日志
 */
export function reportError(
  raw: unknown,
  context: string,
  hook?: (err: AppError, context: string) => void,
): AppError {
  // 复用错误契约的归一化，保证日志里的 code/message 与调用方 catch 到的是同一形状
  const err = normalizeError(raw)
  logError(`[${context}] ${err.message}`, err.detail ? { code: err.code, detail: err.detail } : { code: err.code })
  hook?.(err, context)
  return err
}

// 当前已安装的处理器，供幂等安装与卸载使用。
let onWindowError: ((event: ErrorEvent) => void) | null = null
let onUnhandledRejection: ((event: PromiseRejectionEvent) => void) | null = null

/**
 * 安装全局错误兜底，返回卸载函数。
 *
 * 三处都要挂，缺一不可：
 *   - `error`             ：同步运行时错误、资源加载失败
 *   - `unhandledrejection`：未 catch 的 Promise —— 绑定调用漏 catch 是重灾区
 *   - Vue 的 errorHandler ：组件树内的渲染/生命周期异常，上面两个覆盖不到
 *     （由 main.ts 用 createVueErrorHandler() 挂到 app.config.errorHandler）
 *
 * 幂等：重复调用只保留最后一次安装。
 */
export function installGlobalErrorHandler(
  hook?: (err: AppError, context: string) => void,
): () => void {
  uninstallGlobalErrorHandler()

  onWindowError = (event: ErrorEvent) => {
    // resource 类错误（图片/脚本加载失败）没有 error 对象，退化为消息文本
    reportError(event.error ?? event.message, 'window.onerror', hook)
  }
  onUnhandledRejection = (event: PromiseRejectionEvent) => {
    reportError(event.reason, '未处理的 Promise 拒绝', hook)
  }

  window.addEventListener('error', onWindowError)
  window.addEventListener('unhandledrejection', onUnhandledRejection)

  return uninstallGlobalErrorHandler
}

/** 卸载全局错误兜底（未安装时为空操作）。 */
export function uninstallGlobalErrorHandler(): void {
  if (onWindowError) {
    window.removeEventListener('error', onWindowError)
    onWindowError = null
  }
  if (onUnhandledRejection) {
    window.removeEventListener('unhandledrejection', onUnhandledRejection)
    onUnhandledRejection = null
  }
}

/**
 * 生成 Vue 的 errorHandler，供 main.ts 挂到 `app.config.errorHandler`。
 *
 * 刻意不在这里 import Vue 类型：`lib/` 是管道层，不依赖框架。
 * 返回的函数签名与 Vue 期望的一致（多余的参数忽略即可）。
 */
export function createVueErrorHandler(
  hook?: (err: AppError, context: string) => void,
): (err: unknown, instance: unknown, info: string) => void {
  return (err: unknown, _instance: unknown, info: string) => {
    reportError(err, `Vue ${info || '异常'}`, hook)
  }
}
