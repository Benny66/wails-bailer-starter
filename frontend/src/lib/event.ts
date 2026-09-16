// 管道契约：Go↔TS 的事件协议前端侧。
//
// 与 Go 侧 internal/event 成对：
//   - Go 用 event.Name(domain, action) 拼事件名，本模块的 eventName() 是它的镜像；
//   - 动作常量（progress/done/error）的单一真相在 Go，此处为镜像，
//     由 internal/guard/parity_test.go 断言二者一致（漂移即失败）。
//
// 为何必须有本模块：没有它，下游只能手写 EventsOn('asset:progress', ...)——
// 那是第二处真相，拼错了不会有任何提示（订阅失败是静默的）。
//
// 可抄的最小范例（长任务推进度）：
//
//	Go 侧（service 层跑长任务，经 App 绑定方法拿到 ctx）：
//	  runtime.EventsEmit(ctx, event.Name("asset", event.ActionProgress),
//	      event.Progress{Done: done, Total: total})
//	  runtime.EventsEmit(ctx, event.Name("asset", event.ActionDone),
//	      event.Result{Ok: true})
//
//	前端侧（组件里，卸载自动解绑）：
//	  const percent = ref(0)
//	  useEvent<EventProgress>(eventName('asset', EventAction.Progress), (p) => {
//	    percent.value = p.total > 0 ? Math.round((p.done / p.total) * 100) : 0
//	  })

import { EventsOn } from '../../wailsjs/runtime/runtime'

/** 事件动作常量。单一真相在 Go 的 internal/event（Action*），此处为镜像。 */
export const EventAction = {
  /** 进度：长任务推进中。 */
  Progress: 'progress',
  /** 成功结束。 */
  Done: 'done',
  /** 失败结束。 */
  Error: 'error',
} as const

/** 事件动作取值。 */
export type EventActionValue = (typeof EventAction)[keyof typeof EventAction]

/**
 * 拼装符合约定的事件名（`<domain>:<action>`）。
 * domain 与业务模块名一致（如 'asset'），镜像 Go 的 event.Name。
 *
 * action 写成 `EventActionValue | string` 是有意的：事件协议是**约定式**的，
 * 下游可以有自定义动作。联合类型让编辑器给出三个内置动作的补全提示，
 * 同时不禁止其他取值（并非冗余——删掉 `| string` 会把约定变成硬约束）。
 */
export function eventName(domain: string, action: EventActionValue | string): string {
  return `${domain}:${action}`
}

/** 进度载荷（对应 Go 的 event.Progress）。Total 为 0 表示总量未知，应按「不确定进度」处理。 */
export interface EventProgress {
  done: number
  total: number
}

/** 结束载荷（对应 Go 的 event.Result）。失败时 ok 为 false，message 为中文提示。 */
export interface EventResult {
  ok: boolean
  message?: string
}

/**
 * 订阅一个后端事件，返回取消订阅函数。
 *
 * 【务必解除订阅】：Wails 的订阅是累积的，漏解绑会让同一事件在路由来回切换后
 * 触发多次（重复处理 + 内存泄漏）。组件里请改用 composables/useEvent.ts 自动解绑。
 *
 * 载荷约定：Go 侧单参数 emit（推荐）→ handler 收到该对象本身；
 * 多参数 emit → handler 收到参数数组。非 wails 环境（纯 vite dev）下订阅不可用，
 * 返回空取消函数并在控制台留下原因，避免「静默不生效」难以排查。
 */
export function onEvent<T = unknown>(name: string, handler: (payload: T) => void): () => void {
  try {
    return EventsOn(name, (...data: unknown[]) => {
      handler((data.length > 1 ? data : data[0]) as T)
    })
  } catch (err) {
    console.warn(`事件订阅失败（非 wails 环境？）: ${name}`, err)
    return () => {}
  }
}
