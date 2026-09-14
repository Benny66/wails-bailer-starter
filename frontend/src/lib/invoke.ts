// 管道契约：Go↔TS 的错误协议前端侧。
//
// 与 Go 侧 internal/apperr 成对：
//   - Go 的 apperr.Format 把错误序列化为 JSON 字符串（经 Wails ErrorFormatter）。
//   - 前端收到的是 Error 实例，其 .message 是那段 JSON。
//   - 本模块负责 JSON.parse 还原为 AppError，并容错退化。
//
// 这是「管道」而非「产品面」：只负责把错误归一化成稳定形状，
// 不做任何 UI（toast / 弹窗由下游按 AppError.code 自行决定）。

/** 归一化后的应用错误。code 供分流，message 可直接展示。 */
export interface AppError {
  code: string
  message: string
  detail?: string
}

/** 与 Go 侧 internal/apperr 的错误码常量保持一致（单一真相在 Go，此处为镜像）。 */
export const ErrorCode = {
  NotFound: 'not_found',
  Validation: 'validation',
  Conflict: 'conflict',
  Internal: 'internal',
} as const

/**
 * 把任意 reject 值归一化为 AppError。
 *
 * 正常路径：ErrorFormatter 产出的 JSON 字符串 → parse 还原。
 * 容错路径：解析失败（非本协议的错误，如 Wails 内部错误）→ 退化为 internal + 原文，
 * 保证下游拿到的永远是稳定形状，不会撞见 "[object Object]"。
 */
export function normalizeError(raw: unknown): AppError {
  const text = raw instanceof Error ? raw.message : String(raw)
  try {
    const parsed = JSON.parse(text)
    if (parsed && typeof parsed.code === 'string' && typeof parsed.message === 'string') {
      return parsed as AppError
    }
  } catch {
    // 落到兜底
  }
  return { code: ErrorCode.Internal, message: text || '未知错误' }
}

/**
 * 包裹一次绑定调用，把错误归一化为 AppError 后重新抛出。
 *
 * 用法（下游范例）：
 *   try {
 *     const data = await invoke(() => ListAssets(req))
 *   } catch (e) {
 *     const err = e as AppError
 *     if (err.code === ErrorCode.NotFound) { ... }  // 按 code 分流
 *   }
 *
 * 不强迫下游使用——直接 await 绑定方法也能跑，只是拿不到归一化形状。
 */
export async function invoke<T>(fn: () => Promise<T>): Promise<T> {
  try {
    return await fn()
  } catch (raw) {
    throw normalizeError(raw)
  }
}
