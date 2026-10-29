# 前端规范

> 状态：Draft ｜ 适用版本：v0.2.1 ｜ 最后更新：2026-10-29

适用目录：`frontend/`。

---

## 1. 技术栈与版本

- Vue 3.5（`<script setup>` + Composition API，**禁止** Options API）
- TypeScript 5.7，`strict: true`
- Vite 6
- Element Plus 2.14（**按需引入**，不整体引入）
- Pinia 3 / vue-router 4 / vue-i18n 11
- UnoCSS 66
- Node 22 LTS，pnpm 10

> Element Plus 的引入方式：
>
> - 允许按组件路径显式引入（`element-plus/es/components/button/index`）；
> - 允许经 `unplugin-vue-components` 的 `ElementPlusResolver` 自动引入；
> - **禁止**从 `element-plus` 根 barrel 或 `element-plus/es` 根入口引入 ——
>   实测会把整包打进 chunk（766 KB / 242 KB gzip），直接击穿体积预算。

新增依赖必须在 PR 描述中写明 `adds: <pkg> <version>` 与引入理由。

---

## 2. 目录结构

```text
frontend/src/
├── api/            # 与后端通信的唯一入口
├── components/
│   ├── common/     # AppIcon / BrandMark / ShortcutHelp
│   ├── layout/     # Header / Sidebar / Footer / SettingsMenu
│   ├── discovery/  # ToolCard / CategoryCard
│   ├── search/     # SearchTrigger / CommandPalette
│   └── tool/       # 工具页专用共享组件（见下方清单）
├── composables/    # use* 组合式函数
├── config/         # accents.ts（主色预设的唯一真源）
├── layouts/        # DefaultLayout
├── locales/        # zh-CN/ 与 en-US/ 目录（common.json、home.json、tool.json、errors.json、tools/<slug>.json）
├── router/
├── stores/         # catalog / prefs
├── styles/         # tokens.css / element-overrides.css / base.css / forms.css
├── tools/<category>/<slug>/{index.ts,Tool.vue,example.ts}
├── types/          # 与后端 DTO 对齐的类型
└── views/          # 路由级页面组件
```

**共享组件清单（`components/tool/`）**

`Asn1Tree` `Base64FileInput` `ByteView` `CertChainTable` `CertDetailDialog`
`CertTextInput` `CodeEditor` `CodeViewer` `CopyButton` `DownloadButton`
`EmptyState` `EncodingSelect` `ErrorAlert` `FileDropZone` `FileTextInput`
`JsonViewer` `KeyInput` `KeyValueTable` `LoadingSkeleton` `ProgressStream`
`QuickReference` `ResultPanel` `RunBar` `SensitiveInput` `ToolPageLayout`
`ToolPanel` `ValidationHint`

**组合式函数清单（`composables/`）**

`useAutoRun` `useCommandPalette` `useFileToolRun` `useGlobalHotkeys`
`useToolRun` `useValidation`

**必须**

- 工具组件只放在 `frontend/src/tools/<category>/<slug>/`；
- 跨工具复用的组件放 `frontend/src/components/tool/`，**禁止**在工具目录内建「公共」组件；
- 与后端 DTO 对齐的类型放 `frontend/src/types/`，**禁止**在组件内重复定义请求/响应结构。

---

## 3. 命名

| 对象 | 约定 | 示例 |
| --- | --- | --- |
| 组件文件 | PascalCase | `ToolCard.vue`、`ResultPanel.vue` |
| 组合式函数 | `useXxx` | `useTheme.ts`、`useToolRun.ts` |
| Store | 小写单数或 `xxxStore` | `catalog.ts`、`prefs.ts` |
| 类型 | PascalCase + 语义后缀 | `ToolMeta`、`CertParseResult` |
| 常量 | SCREAMING_SNAKE | `MAX_FILE_BYTES` |
| CSS 变量 | `--tb-` 前缀 | `--tb-primary`、`--tb-bg` |
| i18n key | 点分层级 | `tools.symmetric-crypt.name` |

**禁止**在组件里出现中文标识符或拼音标识符。

---

## 4. 组件规范

### 4.1 通用

- **必须**为纯展示组件使用 `defineProps` + 类型化 props；禁止 `any`。
- **必须**提供 `aria-label` 或可见文本，图标按钮尤其如此。
- 交互元素**必须**可聚焦，焦点样式使用 `:focus-visible`。
- 列表**必须**使用语义标签（`ul/li`、`table`）或补齐 `role`。

### 4.2 工具页骨架

所有工具页**必须**复用同一骨架，不允许各写各的：

```text
ToolPageLayout
├── 标题行：图标 / 名称 / 描述 / 徽章 / 收藏 / 复制链接
├── ToolPanel
│   ├── 输入区（FieldRow / FileDropZone / KeyInput / CertTextInput ...）
│   ├── 参数区（按工具）
│   ├── RunBar：运行 / 重置 / 填入示例
│   └── ResultPanel：idle / loading / success / error
└── 内容区：使用说明 / 输入输出说明 / FAQ / 相关工具 / 安全提示
```

- 布局默认**单列上下**；catalog 中标记 `wide: true` 的工具在视口 ≥1280 时切换为左右分栏。
- 双栏时两列**等高**，且列内不设滚动条 —— 内容超出时由页面整体滚动。
- 工具页的「使用说明」与「常见问题」**默认展开**，标题本身即折叠开关（不再单列一个「展开 / 收起」按钮）。
- 侧栏分「固定区」与「滚动区」：快捷入口、分类标题、过滤框固定不动，**只有分类列表滚动**。
  过滤框只筛分类下的工具，因此置于分类标题下方；分类标题右侧提供一键收起 / 展开全部。
- 分类下的每个工具项右侧提供收藏按钮（淡出，hover / 已收藏时可见），星标状态与工具页标题旁的收藏一致。
- 工具项的 hover 与选中底色**必须**覆盖整行（含收藏按钮），不得只染链接部分。
- 分类下的工具项**必须**带图标：图标表是 `frontend/src/tools/toolIcons.ts`（按 slug 登记，
  未登记时回落到所属分类图标）；分类图标仍来自 catalog 的 `icon` 字段。
- 图标颜色**必须**按分类 / 工具取各自的色调，**禁止**统一成一个主色；
  色调表是 `frontend/src/tools/tones.ts`（见 §8.6）。
- 分类展开后的子项相对分类项缩进一级（**不画**引导竖线），层级靠子项自己的图标与缩进体现。
- 折叠态**必须**收成图标轨道（分类显示图标，不显示文字缩写），宽度变化带过渡；
  折叠时顶部标题 / 过滤框同步收高，不得整块瞬间消失。
- 顶栏的侧栏伸缩按钮**必须**用带方向的面板图标（指向动作结果：展开态指向左、收起态指向右），
  **禁止**再用汉堡图标 —— 汉堡表达「打开菜单」，无法区分收起与展开。
- 「快捷」与「分类」两组之间**必须**有分隔线与组间距，不得只用标题文字分隔。
- 折叠态悬浮或聚焦分类图标**必须**弹出子菜单：按视口坐标定位（`Teleport` 到 `body`，
  避免被侧栏 `overflow` 裁掉），列出该分类全部工具并显示一句话描述。
  描述**按需加载**，来源是既有 `locales/<lang>/tools/<slug>.json`，**禁止**另建一份文案。
- 工具页面包屑格式为「首页 / 分类 / 工具名」，中间层可点回分类页。
- 顶栏右段固定为 GitHub、全屏、主题、设置四个入口；GitHub 是**外链**
  （新标签页 + `rel="noopener noreferrer"`），地址取 `frontend/src/config/site.ts` 的 `REPO_URL`，
  悬浮提示**必须**说明「可自部署」（局域网 Docker 部署、数据不出内网）。
- 移动端运行按钮**必须**吸底。

### 4.3 共享组件清单

| 组件 | 用途 |
| --- | --- |
| `EncodingSelect` | utf8 / hex / base64 / base64url 切换 |
| `SensitiveInput` | 遮蔽 + 显示切换（密钥、口令） |
| `KeyInput` | 私钥 / 公钥输入，自动识别 PEM/DER |
| `CertTextInput` | 证书 / CSR / CRL 文本输入，自动嗅探格式 |
| `FileTextInput` | 文件或文本二选一 |
| `FileDropZone` | 拖拽上传区，含大小与类型提示 |
| `CodeEditor` | 等宽多行编辑（Tab 处理、换行开关、可选行号） |
| `JsonViewer` | 折叠树形 JSON 展示 |
| `ByteView` | 十六进制转储 |
| `Asn1Tree` | ASN.1 结构树 |
| `CertChainTable` | 证书链表格（主题 / 签发者 / 有效期 / 类型 / 指纹） |
| `CertDetailDialog` | 证书详情弹窗 |
| `KeyValueTable` | 键值对展示 |
| `CopyButton` / `DownloadButton` | 复制 / 下载 |
| `EmptyState` / `ErrorAlert` / `LoadingSkeleton` | 状态展示 |
| `ProgressStream` | 流式进度（ssl-check） |

**禁止**在工具组件内重复实现上述能力。

---

## 5. 状态管理

- **Pinia** 只放**跨路由共享**的状态：`catalog`、`prefs`（主题、主色、语言、侧栏折叠、收藏、最近）。
- **禁止**把工具的输入内容放进 Pinia 或任何全局 store。
- **收藏与最近使用只存 slug 列表**（`toolbit.favorites` / `toolbit.recent`），最近使用上限 12 条；**禁止**在其中附带任何输入、结果或时间戳以外的元数据。
- 所有写入 `localStorage` / `sessionStorage` 的键**必须**以 `toolbit.` 为前缀：设置面板的「清理缓存」按前缀批量删除，前缀不符的键不会被清理。
- 「清理缓存」**必须**二次确认，**必须**在清理后重新加载页面，让分栏比例等页面内状态一并回到默认。
- 工具内部状态用 `ref` / `reactive` / `computed` 局部管理。
- 组合式函数**必须**在卸载时清理副作用（定时器、事件监听、AbortController）。

---

## 6. API 调用

- 后端通信**只能**经过 `frontend/src/api/client.ts`。
- **禁止**在组件内直接 `fetch` / `XMLHttpRequest`。
- 客户端**必须**统一拆信封：读 `code` / `data` / `error` / `meta`，并把 `error.code` 映射为本地化文案。
- 请求**必须**携带 `AbortController`；组件卸载或用户重跑时**必须**中止上一次请求。
- 文件类工具用 `FormData`；二进制响应走 `blob()`。

---

## 7. 国际化

- **禁止**在模板或脚本中硬编码面向用户的文案。
- 文案一律走 `t('key')`；`zh-CN` 与 `en-US` **必须**成对存在。
- 工具文案 key 固定为 `tools.<slug>.*`，字段（**必须**与 [../i18n.md §4](../i18n.md) 一致）：

  | 字段 | 类型 | 说明 |
  | --- | --- | --- |
  | `name` | string | 工具名 |
  | `description` | string | 一句话描述 |
  | `category` | string | 分类名（与 catalog 一致） |
  | `actions` | object | 操作按钮文案（键 = `operation` 取值） |
  | `fields` | object | 输入字段标签（键 = 请求字段名） |
  | `options` | object | 下拉 / 分组选项文案 |
  | `results` | object | 结果区标题与说明 |
  | `examples` | array | 示例数据（对应工具的 `example.ts`） |
  | `hints` | object | 字段级提示与安全提示 |
  | `faq` | `{q,a}[]` | 2–3 条常见问题 |

- 插值使用 `t('key', { n })` 形式，**禁止**字符串拼接。
- 详见 [docs/i18n.md](../i18n.md)。

---

## 8. 样式与主题

### 8.1 设计令牌

所有颜色、间距、圆角、字号**必须**来自 `frontend/src/styles/tokens.css` 的 CSS 变量，**禁止**在组件里写死十六进制颜色。

| 组 | 变量 |
| --- | --- |
| 主色 | `--tb-primary` / `-hover` / `-active` / `-soft` / `-contrast` |
| 背景 | `--tb-bg` / `--tb-bg-elevated` / `--tb-bg-sunken` |
| 边框 | `--tb-border` / `--tb-border-strong` |
| 文字 | `--tb-text` / `--tb-text-secondary` / `--tb-text-tertiary` |
| 语义 | `--tb-success` / `--tb-warning` / `--tb-danger` / `--tb-info` |
| 品牌标识 | `--tb-brand-from` / `--tb-brand-to` / `--tb-brand-fg` |

- 浅色：`--tb-bg` = `#F7F8FA`、elevated `#FFFFFF`、border `#E5E7EB`、text `#111827 / #4B5563 / #9CA3AF`
- 深色：`--tb-bg` = `#0B0F17`、elevated `#111827`、sunken `#0F1520`、border `#1F2937`、text `#E5E7EB / #9CA3AF`

### 8.2 主题

- 模式：`system` / `light` / `dark`，默认 `system`，浅色优先。
- 切换**必须**即时生效并持久化（`toolbit.theme`）。
- **必须**监听系统主题变化并实时跟随。

### 8.3 主色

- 共 12 组预设：靛蓝 `#4F46E5`（默认）、蓝 `#2563EB`、GitLab 蓝 `#1F75CB`、天蓝 `#0369A1`、
  青绿 `#0F766E`、绿 `#15803D`、黄绿 `#4D7C0F`、琥珀 `#B45309`、橙 `#C2410C`、
  玫红 `#E11D48`、紫 `#7C3AED`、石墨 `#475569`。
- **禁止**提供任意 HEX 取色。
- 每组预设**必须**提供明暗两套令牌（`--tb-primary` / `-hover` / `-active` / `-soft` / `-contrast`），
  由 `:root[data-accent]` 与 `:root[data-theme='dark'][data-accent]` 两个作用域定义。
- 运行时写入 `--el-color-primary` 及 `light-3/5/7/8/9`、`dark-2`，使 Element Plus 组件跟随；
  派生值**必须**用 `color-mix()` 从 `--tb-primary` 计算，不得写死。
- 语义色（`success` / `warning` / `danger` / `info`）**必须**保持固定，**不随主色漂移** ——
  语义色承载状态语义，与品牌无关。
- 预设色**必须**通过 4.5:1 对比度校验（`scripts/check-accent.mjs` 门禁，检查明暗两套共 52 项）。
- 图标色调（§8.6）**必须**同样通过 4.5:1 对比度校验，且**不得随主色漂移**（同一门禁一并校验）。

### 8.4 其他

- 字体：系统无衬线（PingFang SC / Microsoft YaHei / Segoe UI）；等宽 `ui-monospace / Menlo / Consolas`。
- 字号（`--tb-font-*`）：11 / 13 / 14 / 15（基准）/ 16 / 18 / 20 / 24 / 30。
  **禁止**在组件里硬编码字号（§8.1）。
- 间距（`--tb-space-*`）：4 / 8 / 12 / 16 / 20 / 24 / 32 / 40 / 48。
- 圆角（`--tb-radius-*`）：6（控件）/ 10（卡片）/ 14（面板）/ 999（胶囊）。
- 阴影（`--tb-shadow-sm/md/lg`）：元素与浮层的层次依据，**禁止**硬编码 `box-shadow`。
- 断点约定：640 / 768 / 900 / 1024 / 1280 / 1536；
  **禁止**出现未登记的魔法数字。
- 动效：150–200ms `ease-out`；**必须**响应 `prefers-reduced-motion`。

### 8.5 品牌标识

- 徽标图形**必须**是矢量（`frontend/public/favicon.svg` + `frontend/src/components/common/BrandMark.vue`），
  **禁止**使用位图、图标字体或外链图片。
- 图形语义**必须**贴合品牌名 Toolbit：`T`（首字母，同时构成 T 型手柄轮廓）+ 六角批头（bit）。
- 配色**必须**走 `--tb-brand-*` 令牌，**固定靛蓝**，**不随主色（§8.3）切换** ——
  徽标是品牌锚点，在任意主色与明暗主题下都保持同一识别度。
- 栅格产物（`favicon-32.png` / `favicon.ico` / `apple-touch-icon.png`）**必须**由同一份矢量源生成，
  **禁止**手工维护两套图形。
- `index.html` **必须**声明 `rel="icon"`（SVG + PNG）与 `rel="apple-touch-icon"`。

### 8.6 图标色调

图标颜色是**展示层**选择，与主色（§8.3）**正交**：主色负责交互高亮（hover、选中、按钮），
色调负责图标识别。图标**禁止**统一成主色，**必须**每个分类、每个工具各有各的颜色。

- 色调共 12 组固定色板（浅色前景色）：
  靛蓝 `#4F46E5`、蓝 `#1D4ED8`、天蓝 `#0369A1`、青 `#0E7490`、青绿 `#0F766E`、绿 `#166534`、
  黄绿 `#3F6212`、琥珀 `#92400E`、橙 `#9A3412`、玫红 `#BE123C`、紫 `#7C3AED`、石墨 `#475569`。
- 每组提供两个令牌：`--tb-tone-<名>`（图标前景色）与 `--tb-tone-<名>-soft`（图标底色），
  明暗两套分别定义在 `tokens.css` 的 `:root` 与 `:root[data-theme='dark']`。
- 色调**固定**，**不随主色切换** —— 色调承载「这个图标是什么颜色」的身份信息，
  与品牌主色是两回事。
- 色调到分类 / 工具的映射表是 `frontend/src/tools/tones.ts`：
  - `CATEGORY_TONES` —— 分类 slug → 色调；
  - `TOOL_TONES` —— 工具 slug → 色调（未登记时回落到所属分类，再回落到 `slate`）。
- 组件**禁止**写死颜色值；用 `toneStyle(tone)` 把 `--tb-tone` / `--tb-tone-soft`
  绑定到 `:style`，再在样式里用 `var(--tb-tone)` 取色。
- 新增工具**必须**在 `TOOL_TONES` 登记色调，否则会沿用分类色，与同分类其他工具撞色。

---

## 9. 表单与输入

- 标签**必须**与输入框关联（`el-form-item` 的 `label` + `for`）。
- 校验错误**必须**显示在字段附近，**禁止**只用 `el-message`。
- 数字输入**必须**限制范围与步长，并在 UI 上体现默认值。
- 密钥、口令类输入**必须**使用 `SensitiveInput`，默认遮蔽。
- **禁止**把任何工具输入写入 `localStorage` / `sessionStorage` / IndexedDB。
- `Ctrl/⌘ + Enter` 运行快捷键由 `RunBar` **统一**实现：它在 `window` 捕获阶段监听该组合键
  并 `stopPropagation()`，因此焦点在工具页任意控件（文本域、`CodeEditor`、普通输入框）内都有效，
  且一次按键只触发一次运行。
- 工具页**禁止**再各自绑定 `Ctrl/⌘ + Enter`；历史上页面级 / `CodeEditor` 内的同名实现会被上述拦截，
  不再执行（`frontend/src/components/tool/RunBar.test.ts` 覆盖该行为）。
- 文本域内 `Tab` **必须**插入缩进而不是切换焦点（长文本编辑场景）。

---

## 10. 结果展示

- 结果区有四种状态：`idle` / `loading` / `success` / `error`。
- 成功后**必须**滚动到结果区，并通过 `aria-live` 播报结果摘要。
- 错误**必须**内联展示（`ErrorAlert`）：错误码 chip + 本地化文案 + 重试按钮；**禁止**只用弹窗。
- 结果**必须**提供复制；文件类结果**必须**提供下载并显示文件名与大小。
- `meta.warnings` **必须**以醒目的非阻断提示展示（如 `INSECURE_ECB`）。

---

## 11. 性能预算

| 项 | 预算 |
| --- | --- |
| 初始 JS（gzip） | < 250 KB |
| 单工具 chunk（gzip） | < 40 KB |
| 首屏可交互 | < 2 s（本地网络） |

- **必须**路由级懒加载。
- **必须**按分类拆 chunk（`manualChunks`）。
- **禁止**引入图表库、图标字体、Web 字体。
- 工具卡片 hover **应当**预取对应 chunk。
- CI **必须**检查 `dist` 体积预算。

---

## 12. 可访问性

- **必须**提供 `header` / `nav` / `main` / `footer` 地标与「跳到主内容」链接。
- 命令面板与抽屉**必须**实现焦点陷阱与关闭后焦点返回。
- 搜索结果**必须**支持 `↑` `↓` 选择、`Enter` 打开、`Esc` 关闭。
- 所有图标按钮**必须**有 `aria-label`。
- 状态变更**必须**通过 `aria-live` 播报。
- 快捷键：`Ctrl/⌘ + K` 打开搜索、`Ctrl/⌘ + Enter` 运行、`Esc` 关闭、`/` 聚焦搜索。

---

## 13. 测试

- 组合式函数、工具组件、API 客户端**必须**有单测（Vitest + `@vue/test-utils`）。
- 组件测试**必须**覆盖：正常渲染、错误态、loading 态。
- **禁止**在测试中依赖真实网络；用 `vi.mock` 拦截 `api/client`。
- 端到端（Playwright）**必须**覆盖：全部工具路由可达、首页搜索→工具页→运行→复制链路、语言与主题切换。

---

## 14. 禁止清单

- ❌ Options API。
- ❌ `any`（除确有必要的外部类型断言）。
- ❌ 在组件内 `fetch`。
- ❌ 硬编码颜色值与用户可见文案。
- ❌ 把工具输入写入任何本地存储。
- ❌ 引入图表库 / 图标字体 / Web 字体。
- ❌ 复制外部项目的代码、样式或文案。
- ❌ 在本仓库任何文件中提及参考项目名称。
