#!/usr/bin/env bash
# ==========================================================================
# release.sh — 版本号管理 + 打包 + 生成 release 说明（可选）
# --------------------------------------------------------------------------
# 用法：bash scripts/release.sh <version>   （如 0.1.0）
# 做三件事：
#   1. 三平台打包（windows/macos/linux）
#   2. 生成 RELEASE_NOTES.md
#   3. 打印产物清单
# 说明：这是可选辅助脚本，CI 主流程不依赖它。发布需自行上传到托管平台。
# ==========================================================================
set -euo pipefail

VERSION="${1:-}"
if [ -z "$VERSION" ]; then
  echo "用法: $0 <version>   （如 0.1.0）" >&2
  exit 1
fi

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

WAILS="$(go env GOPATH)/bin/wails"
APP_NAME="wails-bailer-starter"

echo "==> 打包版本 $VERSION ..."
mkdir -p dist

# 三平台打包（各平台需对应工具链）
"$WAILS" build -platform windows/amd64 -o "dist/${APP_NAME}-${VERSION}-windows-amd64.exe" 2>&1 | tail -2 || echo "Windows 打包跳过（需在支持的环境）"
"$WAILS" build -platform darwin/universal -o "dist/${APP_NAME}-${VERSION}-macos-universal" 2>&1 | tail -2 || echo "macOS 打包跳过（需在支持的环境）"
"$WAILS" build -platform linux/amd64 -o "dist/${APP_NAME}-${VERSION}-linux-amd64" 2>&1 | tail -2 || echo "Linux 打包跳过（需在支持的环境）"

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
