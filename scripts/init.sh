#!/usr/bin/env bash
# ==========================================================================
# init.sh — 脚手架实例化脚本
# --------------------------------------------------------------------------
# 用法：bash scripts/init.sh <name> [bundle-prefix]
#   <name>          项目名（如 myapp）
#   [bundle-prefix] macOS bundle id 前缀（如 com.mycompany），默认 com.example
# 从母版生成一个干净、可编译的新项目：
#   1. 校验 name 合法性
#   2. git clone 母版（天然排除 .git / settings.local.json / node_modules / build 产物）
#   3. 全局替换「项目名占位符」→ name（Go import / go.mod / wails.json / index.html / _example / scripts）
#   4. 替换「bundle 前缀占位符」→ bundle 前缀（build/darwin/Info.plist 的 bundle id）
#   5. 清空母版归档历史（openspec/changes/archive），保留 specs 能力基线
#   6. 重新 git init + 首次提交
#   7. 断言两个占位符残留均为空 + 回读 bundle id 校验
#
# 占位符一律经下方 PH_APP / PH_BUNDLE 变量使用，不写裸字面量——原因见那两行的注释。
# ==========================================================================
set -euo pipefail

NAME="${1:-}"
if [ -z "$NAME" ]; then
  echo "用法: $0 <name> [bundle-prefix]   （小写字母开头，可含数字/下划线/连字符）" >&2
  exit 1
fi

# bundle id 前缀：macOS 用它区分应用（登录项/权限授予/文件关联都按它记）。
# 默认 com.example 是 IANA 保留给示例的域名，明显是占位符——好过沿用 Wails 的
# 默认 com.wails.<name>（那等于宣称 Wails 是厂商，且与其他 Wails 项目撞名）。
BUNDLE_PREFIX="${2:-com.example}"
if ! printf '%s' "$BUNDLE_PREFIX" | grep -qE '^[a-z][a-z0-9]*(\.[a-z0-9]+)+$'; then
  echo "错误：bundle 前缀 '$BUNDLE_PREFIX' 不合法，须形如 com.mycompany（小写、点分）" >&2
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

# 占位符拼写：必须【拆写】，不能写成完整的占位符字面量。
# 原因：本脚本自己也在替换范围内，裸写会被自己的替换改掉，断言随即失效
# （与 package.sh 的防呆同因，见其 PLACEHOLDER 注释）。
# 拆写后文件里不存在连续的占位符字面量，替换不会碰它，断言在母版与实例化副本里都成立。
# 注：此前能侥幸工作，是因为 bash 在替换发生前已把整个脚本读进内存——
# 依赖读取块大小的行为不可依赖，脚本一变大就可能读到改写后的内容。
PH_APP="__APP_""NAME__"
PH_BUNDLE="__BUNDLE_""PREFIX__"

# 2. 全局替换占位符
#    显式列出需要替换的文件类型（不用 grep -r 盲扫 node_modules）
find "$DEST" -type f \
  \( -name '*.go' -o -name '*.mod' -o -name '*.json' -o -name '*.html' \
     -o -name '*.sh' -o -name '*.md' -o -name '*.vue' -o -name '*.ts' \
     -o -name '*.plist' \) \
  -not -path '*/node_modules/*' -not -path '*/dist/*' -not -path '*/.git/*' \
  -print0 | while IFS= read -r -d '' f; do
    if grep -q -- "$PH_APP" "$f" 2>/dev/null; then
      sed -i '' "s/${PH_APP}/${NAME}/g" "$f" 2>/dev/null || sed -i "s/${PH_APP}/${NAME}/g" "$f"
    fi
    # bundle 前缀单独替换（取值来自第二个参数，与项目名无关）
    if grep -q -- "$PH_BUNDLE" "$f" 2>/dev/null; then
      sed -i '' "s/${PH_BUNDLE}/${BUNDLE_PREFIX}/g" "$f" 2>/dev/null || sed -i "s/${PH_BUNDLE}/${BUNDLE_PREFIX}/g" "$f"
    fi
  done

# .gitignore（无扩展名）单独替换：项目名占位符 → name（忽略项目名命名的裸二进制）
if grep -q -- "$PH_APP" "$DEST/.gitignore" 2>/dev/null; then
  sed -i '' "s/${PH_APP}/${NAME}/g" "$DEST/.gitignore" 2>/dev/null || sed -i "s/${PH_APP}/${NAME}/g" "$DEST/.gitignore"
fi

# 3. 清空母版归档历史，保留 specs 能力基线
rm -rf "$DEST/openspec/changes/archive"
mkdir -p "$DEST/openspec/changes/archive"

# 4. 重新 git init + 首次提交
rm -rf "$DEST/.git"
cd "$DEST"
git init -q
git add -A
git -c user.name="scaffold" -c user.email="scaffold@local" commit -q -m "初始化项目 ${NAME}（由 scaffold-init 生成）" 2>/dev/null || echo "（首次提交跳过：无提交者信息，不影响项目）"

# 5. 断言占位符残留为空（-I 忽略二进制文件，避免误报裸二进制产物）
#    两个占位符都要查：漏掉 bundle 前缀会让产物 bundle id 变成「占位符.myapp」，
#    而这种错误要到 macOS 侧栏/权限授予时才显形。
for PH in "$PH_APP" "$PH_BUNDLE"; do
  RESIDUE=$(grep -rIn -- "$PH" "$DEST" 2>/dev/null | grep -vE 'node_modules|/dist/|\.git/' || true)
  if [ -n "$RESIDUE" ]; then
    echo "错误：以下文件仍残留 $PH，替换遗漏：" >&2
    echo "$RESIDUE" >&2
    exit 1
  fi
done

# 6. 回读校验：bundle id 必须真的进了 Info.plist（占位符替换「没生效」比「替换错」更难发现）
PLIST="$DEST/build/darwin/Info.plist"
if ! grep -q "<string>${BUNDLE_PREFIX}\." "$PLIST"; then
  echo "错误：$PLIST 的 bundle id 未按 '${BUNDLE_PREFIX}.' 生成" >&2
  exit 1
fi

echo ""
echo "✓ 项目已生成：$DEST"
echo ""
echo "下一步："
echo "  cd $NAME"
echo "  make dev              # 启动开发态"
echo "  make gen name=asset   # 生成你的第一个模块"
echo ""
echo "注意："
echo "  · macOS bundle id 现为 ${BUNDLE_PREFIX}.${NAME}"
if [ "$BUNDLE_PREFIX" = "com.example" ]; then
  echo "    用的是默认占位前缀 com.example —— 发布前请改成你的公司域名："
  echo "    bash scripts/init.sh 已生成完毕，直接改 build/darwin/Info.plist 即可"
fi
echo "  · 版本号在 wails.json 的 info.productVersion（当前 0.1.0），发布时用"
echo "    bash scripts/release.sh <version> 更新，它会同步进产物与应用内"
