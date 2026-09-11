#!/usr/bin/env bash
# ==========================================================================
# package.sh — 参数化打包（真安装包）
# --------------------------------------------------------------------------
# 用法：bash scripts/package.sh [os] [scope]
#   - 无参              → 打包当前平台
#   - windows [user|machine] → 交叉编译 Windows .exe 安装器（mac/linux 上可用），
#                         scope 默认 user（免管理员）
#   - macos             → 仅 mac 上可用，产出美化 dmg（含 Applications 软链）
#   - linux             → 仅 linux 上可用，否则报错
#
# 母版防呆：检测到 __APP_NAME__ 占位符残留则拒绝打包（母版不能直接打包）。
# ==========================================================================
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

WAILS="$(go env GOPATH)/bin/wails"
OS="${1:-}"
SCOPE="${2:-user}"   # windows 安装范围：user（默认，免管理员）/ machine

# ---- 母版防呆：占位符残留则拒绝打包 ----
# 注意：检测字符串不能写成裸 __APP_NAME__（会被 init.sh 全局替换成项目名，
# 导致防呆失效）。用下划线拆写规避替换。
PLACEHOLDER="__APP_""NAME__"
if grep -rq "$PLACEHOLDER" go.mod wails.json main.go 2>/dev/null; then
  echo "错误：检测到占位符残留——母版不能直接打包。" >&2
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
    command -v hdiutil >/dev/null 2>&1 || { echo "错误：未找到 hdiutil" >&2; exit 1; }

    APP="build/bin/${APP_NAME}.app"
    DMG="build/bin/${APP_NAME}.dmg"
    [ -d "$APP" ] || { echo "错误：未找到 .app 产物 $APP" >&2; exit 1; }

    # ---- 1. 准备 staging 目录：app + Applications 软链 + 背景图 ----
    STAGING="$(mktemp -d)"
    TMP_DMG="$(mktemp -u).dmg"        # 可读写中间态
    MOUNT_PT="$(mktemp -d)"
    cleanup_dmg() {
      hdiutil detach "$MOUNT_PT" -quiet 2>/dev/null || true
      rm -rf "$STAGING" "$MOUNT_PT" "$TMP_DMG" 2>/dev/null || true
    }
    trap cleanup_dmg EXIT

    cp -R "$APP" "$STAGING/"
    ln -s /Applications "$STAGING/Applications"     # 软链：拖进去即装

    BG_SRC="build/darwin/dmg-background.png"
    if [ -f "$BG_SRC" ]; then
      mkdir -p "$STAGING/.background"
      cp "$BG_SRC" "$STAGING/.background/bg.png"
    fi

    # ---- 2. 生成可读写 dmg（UDRW，才能写 Finder 布局） ----
    hdiutil create -volname "$APP_NAME" -srcfolder "$STAGING" -ov -format UDRW "$TMP_DMG" >/dev/null

    # ---- 3. 挂载 ----
    hdiutil attach "$TMP_DMG" -mountpoint "$MOUNT_PT" -nobrowse -quiet

    # ---- 4. 设置 Finder 布局（美化：窗口尺寸 + 背景图 + 图标定位） ----
    # 布局设置失败不阻断（降级为无美化的 dmg + 软链仍可用）
    if [ -f "$BG_SRC" ]; then
      osascript <<APPLESCRIPT >/dev/null 2>&1 || echo "提示：Finder 布局设置失败，dmg 仍可用（无美化）" >&2
tell application "Finder"
  tell disk "$APP_NAME"
    open
    set current view of container window to icon view
    set toolbar visible of container window to false
    set statusbar visible of container window to false
    set bounds of container window to {100, 100, 700, 500}
    set theViewOptions to icon view options of container window
    set arrangement of theViewOptions to not arranged
    set icon size of theViewOptions to 100
    set background picture of theViewOptions to file ".background:bg.png"
    set position of item "${APP_NAME}.app" of container window to {150, 200}
    set position of item "Applications" of container window to {450, 200}
    update without registering applications
    delay 1
    close
  end tell
end tell
APPLESCRIPT
    fi

    # ---- 5. 卸载 + 压缩为只读最终 dmg ----
    sync
    hdiutil detach "$MOUNT_PT" -quiet
    rm -f "$DMG"
    hdiutil convert "$TMP_DMG" -format UDZO -o "$DMG" >/dev/null

    trap - EXIT
    cleanup_dmg
    echo "✓ 已生成: $DMG（含 Applications 软链 + 拖拽引导）"
    ;;
  linux)
    "$WAILS" build -clean
    echo "✓ 已生成: build/bin/${APP_NAME}"
    ;;
  windows)
    if command -v makensis >/dev/null 2>&1; then
      "$WAILS" build -platform windows/amd64 -nsis -installscope "$SCOPE"
      echo "✓ 已生成: build/bin/（.exe 安装器，安装范围: $SCOPE）"
    else
      echo "提示：未找到 makensis（NSIS 编译器），无法生成 .exe 安装器。" >&2
      echo "  安装方式：macOS → brew install makensis；Windows → 装 NSIS (nsis.sourceforge.io)" >&2
      echo "  当前降级为裸 exe（非安装器）。" >&2
      "$WAILS" build -platform windows/amd64
      echo "✓ 已生成: build/bin/（裸 .exe，非安装器）"
    fi
    ;;
esac

echo ""
echo "打包完成，产物在 build/bin/ 目录。"
