//go:build production

package main

// debugBuild 标识当前为开发构建；生产构建（`wails build` 默认的 Production 模式
// 会自动带上 production 标签）为 false。
const debugBuild = false
