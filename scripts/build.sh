#!/usr/bin/env bash
# ==========================================================================
# build.sh — 带版本注入的 wails 构建（make build / package.sh / release.sh 共用）
# --------------------------------------------------------------------------
# 为什么需要它：版本号必须同时到达四处，且四处【一致】——
#   1. macOS Info.plist（CFBundleShortVersionString / CFBundleVersion）
#   2. Windows exe 版本资源（build/windows/info.json）
#   3. Windows NSIS 安装器（VIFileVersion / 注册表 DisplayVersion）
#   4. 应用内（app.log 首行 / GetAppInfo）—— 只能靠 -ldflags 注入
#
# 前三处由 wails 从 wails.json 的 info.productVersion 渲染；第四处 wails 不管。
# 若各脚本各算各的，就会出现「系统显示 1.0.0、日志显示 dev」这类双源漂移
# （本仓曾长期如此：Info.plist 恒为 Wails 的默认 1.0.0）。
#
# 因此：版本号【单一真相 = wails.json 的 info.productVersion】，本脚本是唯一注入点。
#
# 为何版本必须是数字点分格式：NSIS 的 VIProductVersion 形如
# `VIProductVersion "${INFO_PRODUCTVERSION}.0"`，非数字（如 git describe 的
# abc1234）会让 Windows 打包在最后一步失败。追溯性另由注入的提交号承担。
#
# 用法：bash scripts/build.sh [wails build 的其余参数...]
#   bash scripts/build.sh                                  # 等价 make build
#   bash scripts/build.sh -clean
#   bash scripts/build.sh -platform windows/amd64 -nsis -installscope user
# ==========================================================================
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

# ---- 定位 wails CLI（Git Bash 下可执行名带 .exe，故多路兜底）----
WAILS="$(go env GOPATH)/bin/wails"
if [ ! -x "$WAILS" ]; then
  if [ -x "${WAILS}.exe" ]; then
    WAILS="${WAILS}.exe"
  else
    WAILS="$(command -v wails || true)"
  fi
fi
if [ -z "$WAILS" ]; then
  echo "错误：找不到 wails CLI。安装：go install github.com/wailsapp/wails/v2/cmd/wails@latest" >&2
  exit 1
fi

# ---- 版本号：单一真相 + 格式校验（早失败好过打包到最后一步才炸）----
VERSION="$(sed -n 's/.*"productVersion"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' wails.json | head -1)"
if [ -z "$VERSION" ]; then
  echo "错误：wails.json 未声明 info.productVersion —— 版本号单一真相处缺失" >&2
  exit 1
fi
if ! printf '%s' "$VERSION" | grep -qE '^[0-9]+(\.[0-9]+){2,3}$'; then
  echo "错误：版本 '$VERSION' 不是数字点分格式（如 0.1.0）" >&2
  echo "      NSIS 的 VIProductVersion 不接受非数字版本，Windows 打包会失败：" >&2
  echo "      build/windows/installer/project.nsi 的 VIProductVersion \"\${INFO_PRODUCTVERSION}.0\"" >&2
  exit 1
fi

# ---- 注入：版本 + 提交号 ----
# 提交号单独注入是刻意的：版本号受限于数字格式，排查时提交号比版本更精确。
# 链接器的 -X 对【不存在】的符号静默忽略，故这两个符号由
# internal/guard/wiring_test.go 核对真实存在。
MODULE="$(awk '/^module /{print $2; exit}' go.mod)"
if [ -z "$MODULE" ]; then
  echo "错误：无法从 go.mod 读取 module 路径" >&2
  exit 1
fi
COMMIT="$(git rev-parse --short HEAD 2>/dev/null || true)"
LDFLAGS="-X ${MODULE}/internal/appinfo.InjectedVersion=${VERSION}"
if [ -n "$COMMIT" ]; then
  LDFLAGS="${LDFLAGS} -X ${MODULE}/internal/appinfo.InjectedCommit=${COMMIT}"
fi

echo "==> 构建版本 ${VERSION}${COMMIT:+ (提交 ${COMMIT})}"
"$WAILS" build -ldflags "$LDFLAGS" "$@"

# ---- 回读校验（macOS）----
# 只断言「wails.json 被读到且渲染进产物」，这是本脚本存在的全部意义。
# 不做这一步，版本没生效只会在用户看「显示简介」时才发现（本仓曾长期如此）。
# 仅在原生 mac 构建时校验——交叉编译（-platform windows/...）时 build/bin 里可能
# 残留上一次的 .app，读了会得出错误结论。
PICKED_PLATFORM=""
prev=""
for a in "$@"; do
  [ "$prev" = "-platform" ] && PICKED_PLATFORM="$a"
  prev="$a"
done
if [ "$(uname -s)" = "Darwin" ] && { [ -z "$PICKED_PLATFORM" ] || [ "${PICKED_PLATFORM%%/*}" = "darwin" ]; }; then
  OUT="$(sed -n 's/.*"outputfilename"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' wails.json | head -1)"
  PLIST="build/bin/${OUT}.app/Contents/Info.plist"
  if [ -f "$PLIST" ]; then
    GOT="$(plutil -p "$PLIST" 2>/dev/null | sed -n 's/.*"CFBundleShortVersionString" => "\([^"]*\)".*/\1/p' | head -1)"
    if [ "$GOT" != "$VERSION" ]; then
      echo "错误：产物 Info.plist 的版本为 '$GOT'，期望 '$VERSION' —— 版本未生效" >&2
      echo "      请检查 wails.json 的 info.productVersion" >&2
      exit 1
    fi
    echo "  ✓ 版本已生效（Info.plist = ${GOT}）"
  fi
fi
