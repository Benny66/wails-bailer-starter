#!/usr/bin/env bash
# ==========================================================================
# gen.sh — 新模块生成器
# --------------------------------------------------------------------------
# 用法：make gen name=<module>   （如 name=asset）
# 从 _example/ 模板生成四段：model + service + 绑定方法 + 前端页面，
# 并用锚点注入到 AllModels() / app.go / 路由 / 菜单。
#
# 原则（详见 openspec/changes/example-module/design.md）：
#   - fail-fast：动任何文件前，先校验锚点存在 + 目标未被占用。
#   - 幂等：目标文件已存在则报错退出，绝不覆盖业务代码。
#   - 占位符替换：Example/example → 目标模块的 PascalCase/snake_case。
#   - 注入统一用 awk 读临时文件（不依赖 sed -i / sed 多行，跨 macOS/GNU）。
# ==========================================================================
set -euo pipefail

NAME="${1:-}"
if [ -z "$NAME" ]; then
  echo "用法: $0 <module-name>   （小写单数，如 asset）" >&2
  exit 1
fi

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

PASCAL="$(echo "$NAME" | awk '{print toupper(substr($0,1,1)) substr($0,2)}')"
SNAKE="$NAME"

# ---- fail-fast 前置校验：锚点必须存在 ----
ANCHORS=(
  "internal/model/model.go|// gen:model"
  "app.go|// gen:import"
  "app.go|// gen:bind"
  "frontend/src/router/index.ts|// gen:route"
  "frontend/src/layouts/AppShell.vue|// gen:menu"
)
for entry in "${ANCHORS[@]}"; do
  file="${entry%%|*}"
  marker="${entry#*|}"
  if ! grep -qF "$marker" "$file" 2>/dev/null; then
    echo "错误：锚点 '$marker' 在 $file 中不存在——写法可能已变更，请同步更新生成器" >&2
    exit 1
  fi
done

# ---- fail-fast 前置校验：目标不得已存在（幂等） ----
TARGETS=(
  "internal/model/${SNAKE}.go"
  "internal/service/${SNAKE}_service.go"
  "frontend/src/views/${SNAKE}/${PASCAL}List.vue"
)
for t in "${TARGETS[@]}"; do
  if [ -e "$t" ]; then
    echo "错误：目标文件 $t 已存在，拒绝覆盖业务代码" >&2
    exit 1
  fi
done

# ---- 工具函数 ----
# 占位符替换：读 stdin，替换 Example/example，写 stdout
replace() {
  sed -e "s/Example/${PASCAL}/g" -e "s/example/${SNAKE}/g"
}

# 在目标文件的锚点行之前，插入内容文件的所有行（awk 读文件，POSIX 可靠）
insert_before_anchor() {
  local target="$1" marker="$2" content_file="$3"
  awk '{
    if (index($0, marker) > 0) {
      while ((getline line < CONTENT) > 0) print line
      close(CONTENT)
    }
    print
  }' CONTENT="$content_file" marker="$marker" "$target" > "$target.tmp" && mv "$target.tmp" "$target"
}

# ---- 生成文件（从 _example 模板复制 + 占位符替换） ----
replace < "_example/model/example.go" > "internal/model/${SNAKE}.go"
replace < "_example/service/example_service.go" > "internal/service/${SNAKE}_service.go"
mkdir -p "frontend/src/views/${SNAKE}"
replace < "_example/frontend/ExampleList.vue" > "frontend/src/views/${SNAKE}/${PASCAL}List.vue"

# ---- 锚点注入（内容先写临时文件，再 awk 插入） ----

# 0. app.go import：在 `// gen:import` 前插入 internal/model
TMP=$(mktemp)
printf '\t"wails-bailer-starter/internal/model"\n' > "$TMP"
insert_before_anchor "app.go" "// gen:import" "$TMP"
rm -f "$TMP"

# 1. AllModels() 注册模型
TMP=$(mktemp)
printf '\t\t&%s{},\n' "$PASCAL" > "$TMP"
insert_before_anchor "internal/model/model.go" "// gen:model" "$TMP"
rm -f "$TMP"

# 2. app.go 绑定方法
TMP=$(mktemp)
replace < "_example/bindings/app_example.go.txt" > "$TMP"
insert_before_anchor "app.go" "// gen:bind" "$TMP"
rm -f "$TMP"

# 3. 路由
TMP=$(mktemp)
cat > "$TMP" <<EOF
  {
    path: '/${SNAKE}',
    name: '${SNAKE}',
    component: () => import('../views/${SNAKE}/${PASCAL}List.vue'),
    meta: { title: '${PASCAL}' },
  },
EOF
insert_before_anchor "frontend/src/router/index.ts" "// gen:route" "$TMP"
rm -f "$TMP"

# 4. 菜单
TMP=$(mktemp)
cat > "$TMP" <<EOF
  { path: '/${SNAKE}', label: '${PASCAL}', icon: '◇' },
EOF
insert_before_anchor "frontend/src/layouts/AppShell.vue" "// gen:menu" "$TMP"
rm -f "$TMP"

echo "已生成模块 ${NAME}（PascalCase: ${PASCAL}）"
echo "  - model:        internal/model/${SNAKE}.go"
echo "  - service:      internal/service/${SNAKE}_service.go"
echo "  - 前端页面:      frontend/src/views/${SNAKE}/${PASCAL}List.vue"
echo "  - 已注入:       AllModels() / app.go 绑定 / 路由 / 菜单"
echo "下一步：填 TODO 处业务逻辑，然后 make dev 重新生成 bindings。"
