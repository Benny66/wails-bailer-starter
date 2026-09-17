#!/usr/bin/env bash
# ==========================================================================
# release.sh — 版本号管理 + 打包 + 生成 release 说明（可选）
# --------------------------------------------------------------------------
# 用法：bash scripts/release.sh <version>   （如 0.1.0）
# 做四件事：
#   1. 把 <version> 写进 wails.json 的 info.productVersion（版本号单一真相）
#   2. 三平台打包（windows/macos/linux；构建经 scripts/build.sh 注入版本）
#   3. 生成 RELEASE_NOTES.md
#   4. 打印产物清单
# 说明：这是可选辅助脚本，CI 主流程不依赖它。发布需自行上传到托管平台。
# 注意：版本号必须是数字点分格式（不同平台打包对它有格式要求，见下）。
# ==========================================================================
set -euo pipefail

VERSION="${1:-}"
if [ -z "$VERSION" ]; then
  echo "用法: $0 <version>   （如 0.1.0）" >&2
  exit 1
fi

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

APP_NAME="__APP_NAME__"

# 更新版本号单一真相（格式校验 + 写入 + 回读都在 set-version.sh 里，
# 与 CI 的发布工作流共用同一处逻辑——两处各写一份版本是漂移的温床）。
bash scripts/set-version.sh "$VERSION"

echo "==> 打包版本 $VERSION ..."
mkdir -p dist

# 三平台打包（各平台需对应工具链）。构建统一走 build.sh，由它注入版本与提交号。
# Linux 行额外带 webkit2_41：wails 默认按 webkit2gtk-4.0 链接，Ubuntu 24.04+/Debian 13+
# 只有 4.1（详见 scripts/package.sh 的同名说明）。Linux 不在 CI 编译矩阵内。
bash scripts/build.sh -platform windows/amd64 -o "dist/${APP_NAME}-${VERSION}-windows-amd64.exe" 2>&1 | tail -2 || echo "Windows 打包跳过（需在支持的环境）"
bash scripts/build.sh -platform darwin/universal -o "dist/${APP_NAME}-${VERSION}-macos-universal" 2>&1 | tail -2 || echo "macOS 打包跳过（需在支持的环境）"
bash scripts/build.sh -platform linux/amd64 -tags webkit2_41 -o "dist/${APP_NAME}-${VERSION}-linux-amd64" 2>&1 | tail -2 || echo "Linux 打包跳过（需在支持的环境）"

# 生成 release 说明
cat > "RELEASE_NOTES.md" <<EOF
# ${APP_NAME} ${VERSION}

## 变更

- （请在此填写本次发布的变更内容）

## 产物

- \`${APP_NAME}-${VERSION}-windows-amd64.exe\`
- \`${APP_NAME}-${VERSION}-macos-universal\`
- \`${APP_NAME}-${VERSION}-linux-amd64\`

## 校验

- 冒烟测试：\`make smoke\`
- 静态检查：\`make lint\`
EOF

echo ""
echo "✓ 完成，产物在 dist/，release 说明见 RELEASE_NOTES.md"
ls -la dist/ 2>/dev/null || true
