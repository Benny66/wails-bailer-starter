#!/usr/bin/env bash
# ==========================================================================
# init.sh — 脚手架实例化脚本
# --------------------------------------------------------------------------
# 用法：bash scripts/init.sh <name>   （如 myapp）
# 从母版生成一个干净、可编译的新项目：
#   1. 校验 name 合法性
#   2. git clone 母版（天然排除 .git / settings.local.json / node_modules / build 产物）
#   3. 全局替换 __APP_NAME__ → name（Go import / go.mod / wails.json / index.html / _example / scripts）
#   4. 清空母版归档历史（openspec/changes/archive），保留 specs 能力基线
#   5. 重新 git init + 首次提交
#   6. 断言 __APP_NAME__ 残留为空
# ==========================================================================
set -euo pipefail

NAME="${1:-}"
if [ -z "$NAME" ]; then
  echo "用法: $0 <name>   （小写字母开头，可含数字/下划线/连字符）" >&2
  exit 1
fi

# 校验 name 合法性：小写字母开头，仅含小写字母/数字/下划线/连字符
if ! echo "$NAME" | grep -qE '^[a-z][a-z0-9_-]*$'; then
  echo "错误：项目名 '$NAME' 不合法，须匹配 ^[a-z][a-z0-9_-]*$" >&2
  exit 1
fi

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DEST="$ROOT/../$NAME"

if [ -e "$DEST" ]; then
  echo "错误：目标目录 $DEST 已存在，拒绝覆盖" >&2
  exit 1
fi

echo "==> 从母版生成项目 '$NAME' ..."

# 1. git clone 母版到目标目录（排除 .git 历史与全局 ignore 的 settings.local.json）
git clone "$ROOT" "$DEST" >/dev/null 2>&1

# 2. 全局替换 __APP_NAME__ → name
#    显式列出需要替换的文件类型（不用 grep -r 盲扫 node_modules）
find "$DEST" -type f \
  \( -name '*.go' -o -name '*.mod' -o -name '*.json' -o -name '*.html' \
     -o -name '*.sh' -o -name '*.md' -o -name '*.vue' -o -name '*.ts' \) \
  -not -path '*/node_modules/*' -not -path '*/dist/*' -not -path '*/.git/*' \
  -print0 | while IFS= read -r -d '' f; do
    if grep -q '__APP_NAME__' "$f" 2>/dev/null; then
      sed -i '' "s/__APP_NAME__/${NAME}/g" "$f" 2>/dev/null || sed -i "s/__APP_NAME__/${NAME}/g" "$f"
    fi
  done

# 3. 清空母版归档历史，保留 specs 能力基线
rm -rf "$DEST/openspec/changes/archive"
mkdir -p "$DEST/openspec/changes/archive"

# 4. 重新 git init + 首次提交
rm -rf "$DEST/.git"
cd "$DEST"
git init -q
git add -A
git -c user.name="scaffold" -c user.email="scaffold@local" commit -q -m "初始化项目 ${NAME}（由 scaffold-init 生成）" 2>/dev/null || echo "（首次提交跳过：无提交者信息，不影响项目）"

# 5. 断言 __APP_NAME__ 残留为空
RESIDUE=$(grep -rn '__APP_NAME__' "$DEST" 2>/dev/null | grep -vE 'node_modules|/dist/|\.git/' || true)
if [ -n "$RESIDUE" ]; then
  echo "错误：以下文件仍残留 __APP_NAME__，替换遗漏：" >&2
  echo "$RESIDUE" >&2
  exit 1
fi

echo ""
echo "✓ 项目已生成：$DEST"
echo ""
echo "下一步："
echo "  cd $NAME"
echo "  make dev              # 启动开发态"
echo "  make gen name=asset   # 生成你的第一个模块"
