#!/usr/bin/env bash
# ==========================================================================
# verify-gen.sh — 生成器端到端验证
# --------------------------------------------------------------------------
# 目的：把「make gen 能不能产出可编译的模块」变成会失败的检查。
#
# 为什么需要它：gen.sh 是管道的入口，但本仓长期从未真跑过一次——
# 实测曾发现三个真实 bug：
#   1. 绑定片段用了 page.Request，但 gen.sh 只注入 model import → 缺 import 编译失败。
#   2. app.go import 注入非幂等 → 生成第 2 个模块即重复 import（redeclared）编译失败。
#   3. import 追加在 import 块尾部 → 不满足 gofmt 字母序 → make gen 后 make lint 必红，
#      且报错指向 app.go，看不出是生成器的锅。
# 三者都不是靠读代码能可靠发现的，必须真跑。
#
# 流程（全程在临时副本里，不污染工作区；清理由 trap 在任何退出路径上完成）：
#   1. 把当前工作区复制到临时目录（替换 __APP_NAME__，母版也能跑）
#   2. 连续生成两个模块（第 2 个专测幂等）+ 断言 import 幂等
#   3. wails build 生成 bindings（新绑定方法需要它，否则护栏会报绑定过期）
#   4. go build 断言编译通过
#   5. gofmt 断言格式合规（gofmt 范围是【整个工作区】，不限于生成物）
#   6. go test 断言护栏通过（含前端镜像一致性）
#
# 用法：bash scripts/verify-gen.sh    （或 make verify-gen）
# ==========================================================================
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SANDBOX="$(mktemp -d)/verifygen"
cleanup() { rm -rf "$(dirname "$SANDBOX")" 2>/dev/null || true; }
trap cleanup EXIT

echo "==> 1/6 复制工作区到临时沙箱 ..."
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

echo "==> 2/6 连续生成两个模块（第 2 个专测幂等）..."
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

echo "==> 3/6 生成 wails bindings（新绑定方法需要它，否则护栏会报绑定过期）..."
if ! "$(go env GOPATH)/bin/wails" build >/dev/null 2>&1; then
  echo "错误：wails build 失败——生成物或绑定生成有问题" >&2
  exit 1
fi
echo "  ✓ bindings 生成"

echo "==> 4/6 断言生成物可编译 ..."
if ! go build ./... 2>&1; then
  echo "错误：生成物编译失败——gen.sh 产出的代码有问题" >&2
  exit 1
fi
echo "  ✓ 编译通过"

echo "==> 5/6 断言格式合规（对齐 make lint 的 gofmt 检查）..."
# 曾经的 bug：锚点注入把 import 追加到块尾部 → 顺序不符 gofmt → make gen 后
# make lint 必红，且报错指向 app.go，很难联想到是生成器的锅。
# 范围是整个工作区（与 make lint 一致），故列出文件也未必是生成器造的
# ——底下的提示把两种可能都说清楚，避免误导排查方向。
unformatted="$(gofmt -l . 2>/dev/null | grep -v node_modules || true)"
if [ -n "$unformatted" ]; then
  echo "错误：以下文件未通过 gofmt（生成物或工作区原有文件），make lint 会红：" >&2
  echo "$unformatted" >&2
  echo "若为生成器注入所致，请修 scripts/gen.sh；否则运行 gofmt -w 修复。" >&2
  exit 1
fi
echo "  ✓ 格式合规"

echo "==> 6/6 断言护栏通过 ..."
if ! go test ./internal/guard/ 2>&1; then
  echo "错误：生成物未通过契约护栏" >&2
  exit 1
fi
echo "  ✓ 护栏通过"

echo ""
echo "生成器端到端验证通过 ✓"
