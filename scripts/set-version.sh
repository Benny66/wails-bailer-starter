#!/usr/bin/env bash
# ==========================================================================
# set-version.sh — 更新版本号【单一真相】（wails.json 的 info.productVersion）
# --------------------------------------------------------------------------
# 用法：bash scripts/set-version.sh <version>   （如 0.1.0）
#
# 版本号只在这里写入，其余一切由它派生：
#   macOS Info.plist / Windows exe 版本资源 / NSIS 注册表（wails 渲染）
#   + 应用内 app.log 与 GetAppInfo（scripts/build.sh 注入 -ldflags）
#
# 本地 release.sh 与 CI 的发布工作流共用本脚本——两处各写一份版本是漂移的温床。
# ==========================================================================
set -euo pipefail

VERSION="${1:-}"
if [ -z "$VERSION" ]; then
  echo "用法: $0 <version>   （如 0.1.0）" >&2
  exit 1
fi

# 格式校验：NSIS 的 VIProductVersion 形如 "${INFO_PRODUCTVERSION}.0"，要求数字点分。
# 非数字版本（v1.2.3 的前缀、git describe 的 abc1234）会让 Windows 打包在最后一步失败，
# 故在写入前就拒绝，而不是等打包打到一半才炸。
if ! printf '%s' "$VERSION" | grep -qE '^[0-9]+(\.[0-9]+){2,3}$'; then
  echo "错误：版本 '$VERSION' 不是数字点分格式（如 0.1.0）" >&2
  echo "      Windows 的 NSIS 不接受非数字版本，见 build/windows/installer/project.nsi" >&2
  exit 1
fi

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

if [ ! -f wails.json ]; then
  echo "错误：找不到 wails.json（版本号单一真相所在）" >&2
  exit 1
fi

# BSD sed 与 GNU sed 的 -i 语义不同，两条都试（与 gen.sh / init.sh 同一手法）
sed -i '' "s|\"productVersion\": \"[^\"]*\"|\"productVersion\": \"${VERSION}\"|" wails.json 2>/dev/null \
  || sed -i "s|\"productVersion\": \"[^\"]*\"|\"productVersion\": \"${VERSION}\"|" wails.json

# 回读确认：「写没写进去」与「写对了」是两件事——静默失败会让产物带错版本
if ! grep -q "\"productVersion\": \"${VERSION}\"" wails.json; then
  echo "错误：写入 wails.json 的 info.productVersion 失败（版本号未生效）" >&2
  exit 1
fi

echo "✓ 版本单一真相已更新为 ${VERSION}（wails.json 的 info.productVersion）"
