#!/usr/bin/env bash
# ==========================================================================
# verify-gen.sh — 生成器端到端验证
# --------------------------------------------------------------------------
# 目的：把「make gen 能不能产出可编译的模块」变成会失败的检查。
#
# 为什么需要它：gen.sh 是管道的入口，但本仓长期从未真跑过一次——
# 实测曾发现两个真实 bug：
#   1. 绑定片段用了 page.Request，但 gen.sh 只注入 model import → 缺 import 编译失败。
#   2. app.go import 注入非幂等 → 生成第 2 个模块即重复 import（redeclared）编译失败。
# 二者都不是靠读代码能可靠发现的，必须真跑。
#
# 流程（全程在临时副本里，不污染工作区）：
#   1. 把当前工作区复制到临时目录（替换 __APP_NAME__，母版也能跑）
#   2. 连续生成两个模块（第 2 个专测幂等）
#   3. go build 断言编译通过
#   4. go test 断言护栏通过
#   5. 清理
#
# 用法：bash scripts/verify-gen.sh    （或 make verify-gen）
# ==========================================================================
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SANDBOX="$(mktemp -d)/verifygen"
cleanup() { rm -rf "$(dirname "$SANDBOX")" 2>/dev/null || true; }
trap cleanup EXIT

echo "==> 1/5 复制工作区到临时沙箱 ..."
mkdir -p "$(dirname "$SANDBOX")"
# 排除 .git（无必要且大）与 build 产物（编译会重建）
rsync -a --exclude '.git' --exclude 'build/bin' "$ROOT/" "$SANDBOX/" 2>/dev/null

cd "$SANDBOX"

# 母版含占位符，gen.sh/编译都不认；实例化为一个合法项目名。
# 非母版（已实例化）时无占位符，此步等价 no-op。
find "$SANDBOX" -type f \
  \( -name '*.go' -o -name '*.mod' -o -name '*.json' -o -name '*.html' \
     -o -name '*.sh' -o -name '*.vue' -o -name '*.ts' \) \
  -not -path '*/node_modules/*' -not -path '*/dist/*' -print0 2>/dev/null |
  while IFS= read -r -d '' f; do
    if grep -q '__APP_NAME__' "$f" 2>/dev/null; then
      sed -i '' 's/__APP_NAME__/verifygen/g' "$f" 2>/dev/null || sed -i 's/__APP_NAME__/verifygen/g' "$f"
    fi
  done

echo "==> 2/5 连续生成两个模块（第 2 个专测幂等）..."
bash scripts/gen.sh asset  >/dev/null
bash scripts/gen.sh widget >/dev/null

# 断言：import 幂等——每个 import 只应出现一次
for imp in 'internal/model' 'internal/page'; do
  n=$(grep -c "verifygen/$imp\"" app.go || true)
  if [ "$n" -ne 1 ]; then
    echo "错误：import $imp 在 app.go 出现 $n 次（期望 1）——gen.sh 的 import 注入非幂等" >&2
    exit 1
  fi
done
echo "  ✓ import 幂等"

echo "==> 3/5 生成 wails bindings（新绑定方法需要它，否则护栏会报绑定过期）..."
if ! "$(go env GOPATH)/bin/wails" build >/dev/null 2>&1; then
  echo "错误：wails build 失败——生成物或绑定生成有问题" >&2
  exit 1
fi
echo "  ✓ bindings 生成"

echo "==> 4/5 断言生成物可编译 ..."
if ! go build ./... 2>&1; then
  echo "错误：生成物编译失败——gen.sh 产出的代码有问题" >&2
  exit 1
fi
echo "  ✓ 编译通过"

echo "==> 5/5 断言护栏通过 ..."
if ! go test ./internal/guard/ 2>&1; then
  echo "错误：生成物未通过契约护栏" >&2
  exit 1
fi
echo "  ✓ 护栏通过"

echo ""
echo "生成器端到端验证通过 ✓"
