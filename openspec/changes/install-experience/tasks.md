# install-experience — 实施任务

## 1. mac dmg 拖拽安装 + 美化

- [ ] 1.1 package.sh mac 分支：staging 目录放 `.app` + `ln -s /Applications` 软链
- [ ] 1.2 用 `hdiutil create -format UDRW` 做可读写中间 dmg + 挂载
- [ ] 1.3 `osascript` 设置 Finder 布局：窗口尺寸、背景图、app/Applications 图标定位
- [ ] 1.4 `hdiutil convert -format UDZO` 压缩为最终只读 dmg + detach 清理
- [ ] 1.5 加 `trap` 清理中间态挂载点
- [ ] 1.6 美化失败容忍（不阻断 dmg 生成）
- [ ] 1.7 提供中性占位背景图 `build/darwin/dmg-background.png`

## 2. Windows 安装范围

- [ ] 2.1 package.sh windows 分支支持 `scope` 参数，透传 `-installscope`
- [ ] 2.2 Makefile 透传 scope
- [ ] 2.3 makensis 缺失时明确提示安装方式

## 3. 文档

- [ ] 3.1 README 讲清 mac（拖拽）/win（双击向导）安装模型差异

## 4. 验证

- [ ] 4.1 `make package`（mac）产出 dmg，挂载后含 app + Applications 软链
- [ ] 4.2 打开 dmg 有美化布局（背景图 + 图标定位）
- [ ] 4.3 `make package os=windows scope=user` 参数传递正确（无 makensis 时提示）
- [ ] 4.4 `openspec validate install-experience` 通过
