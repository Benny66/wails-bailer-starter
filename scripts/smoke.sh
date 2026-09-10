#!/usr/bin/env bash
# ==========================================================================
# smoke.sh — 冒烟测试
# --------------------------------------------------------------------------
# 目标：把"能跑"变成可观察事实，而非嘴上说。
# 流程：构建 → 启动二进制 → 断言进程存活 + 运行时产物落盘 → trap 清理。
#
# 说明：Wails 无 headless 模式，冒烟定位为本地验证（CI 只跑编译不启动 GUI）。
#   - 绑定握手（调一个 Go 方法验证往返）需在 GUI 环境，故这里用"进程存活 +
#     数据库/config 落盘"作为 startup 链路成功的客观证据。
# ==========================================================================
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

APP_NAME="__APP_NAME__"
WAILS="$(go env GOPATH)/bin/wails"
DATA_DIR="$HOME/Library/Application Support/${APP_NAME}"   # macOS；其他平台见下方兜底

# 平台相关的二进制路径与数据目录
BIN=""
case "$(uname -s)" in
  Darwin)
    BIN="build/bin/${APP_NAME}.app/Contents/MacOS/${APP_NAME}"
    ;;
  Linux)
    BIN="build/bin/${APP_NAME}"
    DATA_DIR="$HOME/.config/${APP_NAME}"
    ;;
  MINGW*|MSYS*|CYGWIN*)
    BIN="build/bin/${APP_NAME}.exe"
    DATA_DIR="$APPDATA/${APP_NAME}"
    ;;
  *)
    echo "不支持的平台: $(uname -s)" >&2
    exit 1
    ;;
esac

cleanup() {
  if [ -n "${APP_PID:-}" ] && kill -0 "$APP_PID" 2>/dev/null; then
    kill "$APP_PID" 2>/dev/null || true
    sleep 1
    kill -9 "$APP_PID" 2>/dev/null || true
  fi
  echo "已清理残留进程"
}
trap cleanup EXIT

# ---- 1. 构建 ----
echo "==> 构建 ${APP_NAME} ..."
"$WAILS" build >/dev/null 2>&1

if [ ! -f "$BIN" ]; then
  echo "错误：构建产物不存在 $BIN" >&2
  exit 1
fi

# ---- 2. 启动 ----
echo "==> 启动二进制 ..."
"$BIN" > /tmp/wails_smoke.log 2>&1 &
APP_PID=$!

# ---- 3. 断言进程存活 ----
sleep 4
if ! kill -0 "$APP_PID" 2>/dev/null; then
  echo "错误：进程启动后立即退出，日志如下：" >&2
  cat /tmp/wails_smoke.log >&2
  exit 1
fi
echo "✓ 进程存活 (pid $APP_PID)"

# ---- 4. 断言运行时产物落盘（startup 链路成功的证据） ----
sleep 1
if [ ! -f "$DATA_DIR/config.json" ]; then
  echo "错误：config.json 未落盘，startup 链路可能未走通" >&2
  exit 1
fi
echo "✓ config.json 已落盘"

if [ ! -f "$DATA_DIR/${APP_NAME}.db" ]; then
  echo "错误：数据库文件未落盘" >&2
  exit 1
fi
echo "✓ 数据库文件已落盘"

echo ""
echo "冒烟测试通过 ✓"
