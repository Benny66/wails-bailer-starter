# ci-release — 实施任务

## 1. 冒烟测试

- [x] 1.1 建 `scripts/smoke.sh`：构建 → 启动 → 绑定握手 → 断言 → `trap` 清理
- [x] 1.2 `make smoke` 从占位改为真实冒烟

## 2. CI

- [x] 2.1 建 `.github/workflows/build.yml`：静态检查 job + 三平台编译矩阵
- [x] 2.2 静态检查 job 跑 `make test` + `make lint`
- [x] 2.3 编译矩阵产出 exe/app/二进制 并 `upload-artifact`

## 3. 文档

- [x] 3.1 写 `README.md`：环境准备/开发/打包/目录结构/命令表/换肤/WebView2 部署
- [x] 3.2 可选 `scripts/release.sh` 版本号 + 打包 + release 说明

## 4. 验证

- [x] 4.1 本地 `make smoke` 通过
- [x] 4.2 `openspec validate ci-release` 通过
