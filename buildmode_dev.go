//go:build !production

package main

// debugBuild 标识当前为开发构建（未带 production 构建标签）。
//
// 判定依据是【构建标签】而非环境变量：Wails 在 `wails build`（默认 Production 模式）
// 时加上 production tag（wails v2 的 pkg/commands/build/base.go），`wails dev` 则不加。
//
// 历史 bug：此处原读环境变量 WAILS_PRODUCTION，而该变量在 Wails 全库中并不存在，
// 判断永远不成立 → 生产态日志级别从未生效（一直按 Debug 输出）。Debug 噪音会快速
// 吃掉 5MB×3 的轮转窗口，把真正的错误冲走。改用构建标签后 dev/build 各自正确。
const debugBuild = true
