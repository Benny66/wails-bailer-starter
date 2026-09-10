#!/usr/bin/env bash
# ==========================================================================
# package.sh — 参数化打包（真安装包）
# --------------------------------------------------------------------------
# 用法：bash scripts/package.sh [os]
#   - 无参        → 打包当前平台
#   - windows     → 交叉编译 Windows .exe 安装器（mac/linux 上可用）
#   - macos       → 仅 mac 上可用，否则报错
#   - linux       → 仅 linux 上可用，否则报错
#
# 母版防呆：检测到 __APP_NAME__ 占位符残留则拒绝打包（母版不能直接打包）。
# ==========================================================================
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

WAILS="$(go env GOPATH)/bin/wails"
OS="${1:-}"

# ---- 母版防呆：占位符残留则拒绝打包 ----
if grep -rq '__APP_NAME__' go.mod wails.json main.go 2>/dev/null; then
  echo "错误：检测到 __APP_NAME__ 占位符残留——母版不能直接打包。" >&2
  echo "请先实例化：bash scripts/init.sh <your-app-name>" >&2
  exit 1
fi

# 应用名（来自 wails.json 的 outputfilename，用 python 精确解析 JSON）
APP_NAME="$(python3 -c 'import json; print(json.load(open("wails.json"))["outputfilename"])' 2>/dev/null)"
if [ -z "$APP_NAME" ]; then
  echo "错误：无法从 wails.json 读取 outputfilename" >&2
  exit 1
fi

# ---- 确定目标平台 ----
CURRENT="$(uname -s)"
case "$CURRENT" in
  Darwin) CUR_PLATFORM="macos" ;;
  Linux)  CUR_PLATFORM="linux" ;;
  MINGW*|MSYS*|CYGWIN*) CUR_PLATFORM="windows" ;;
  *) echo "不支持的当前平台: $CURRENT" >&2; exit 1 ;;
esac

TARGET="${OS:-$CUR_PLATFORM}"

# ---- 校验目标平台在当前机器是否可打包 ----
case "$TARGET" in
  macos)
    if [ "$CUR_PLATFORM" != "macos" ]; then
      echo "错误：macOS 包只能在 mac 上打包（Wails 不支持交叉编译 mac）。" >&2
      exit 1
    fi
    ;;
  linux)
    if [ "$CUR_PLATFORM" != "linux" ]; then
      echo "错误：Linux 包只能在 linux 上打包（Wails 不支持交叉编译 linux）。" >&2
      exit 1
    fi
    ;;
  windows)
    # windows 是唯一支持交叉编译的目标，mac/linux 上都可打
    ;;
  *)
    echo "用法: make package [windows|macos|linux]" >&2
    exit 1
    ;;
esac

# ---- 打包 ----
echo "==> 打包目标: $TARGET (当前: $CUR_PLATFORM) ..."

case "$TARGET" in
  macos)
    "$WAILS" build -clean
    # 用 hdiutil 封装 .dmg（macOS 系统自带，零依赖）
    command -v hdiutil >/dev/null 2>&1 || { echo "错误：未找到 hdiutil" >&2; exit 1; }
    APP="build/bin/${APP_NAME}.app"
    DMG="build/bin/${APP_NAME}.dmg"
    [ -d "$APP" ] || { echo "错误：未找到 .app 产物 $APP" >&2; exit 1; }
    hdiutil create -volname "$APP_NAME" -srcfolder "$APP" -ov -format UDZO "$DMG" >/dev/null
    echo "✓ 已生成: $DMG"
    ;;
  linux)
    "$WAILS" build -clean
    echo "✓ 已生成: build/bin/${APP_NAME}"
    ;;
  windows)
    "$WAILS" build -platform windows/amd64 -nsis
    echo "✓ 已生成: build/bin/（.exe 安装器）"
    ;;
esac

echo ""
echo "打包完成，产物在 build/bin/ 目录。"
