# install-experience — 实施任务

## 1. mac dmg 拖拽安装 + 美化

- [x] 1.1 package.sh mac 分支：staging 目录放 `.app` + `ln -s /Applications` 软链
- [x] 1.2 用 `hdiutil create -format UDRW` 做可读写中间 dmg + 挂载
- [x] 1.3 `osascript` 设置 Finder 布局：窗口尺寸、背景图、app/Applications 图标定位
- [x] 1.4 `hdiutil convert -format UDZO` 压缩为最终只读 dmg + detach 清理
- [x] 1.5 加 `trap` 清理中间态挂载点
- [x] 1.6 美化失败容忍（不阻断 dmg 生成）
- [x] 1.7 提供中性占位背景图 `build/darwin/dmg-background.png`

## 2. Windows 安装范围

- [x] 2.1 package.sh windows 分支支持 `scope` 参数，透传 `-installscope`
- [x] 2.2 Makefile 透传 scope
- [x] 2.3 makensis 缺失时明确提示安装方式

## 3. 文档

- [x] 3.1 README 讲清 mac（拖拽）/win（双击向导）安装模型差异

## 4. 验证

- [x] 4.1 `make package`（mac）产出 dmg，挂载后含 app + Applications 软链（已验证）
- [x] 4.2 打开 dmg 有美化布局（背景图已入 dmg；Finder 布局需 GUI 会话，无 GUI 时降级不阻断）
- [x] 4.3 `make package os=windows scope=user` 参数拼接正确（makensis 缺失时明确提示并降级；本次环境无法装 makensis，NSIS 完整生成未验证）
- [x] 4.4 `openspec validate install-experience` 通过
