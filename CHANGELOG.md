# Changelog

本文件记录 Blue Cloud Toolbit 的所有显著变更。

格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/)，
版本号遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

## [未发布]

部署产物与容器探活子命令：仓库从「文档声明了部署形态」变为「真的能构建、能跑」。

### Changed

- **结果区全面结构化**：28 个此前把多字段压成 `key: value` 文本行的工具改为结构化展示
  - 新增三个共享组件并登记进 `docs/standards/01-frontend.md` §4.3：
    `DataTable`（多记录表格：列定义 + 行数据 + 逐行复制 + 滚动区）、
    `StatGrid`（少量关键指标的卡片网格）、`FilterChips`（可点选的筛选标签组）
  - **证书类（16 个）**：摘要用 `StatGrid` / `KeyValueTable`，多记录（证书链、吊销条目、
    CSR 请求扩展、各解析器、逐级校验明细）用 `DataTable`，PEM / DER / 十六进制等**产物**
    改走独立页签并保留全文复制
  - **密钥 / 加密类（5 个）**：`key-match`、`keygen`、`ecdh`、`symmetric-crypt`、`hash-calc`
    改为卡片 + 页签；`hash-calc` 的多算法摘要改为表格
  - **网络类（2 个）**：`ip-range-expand` 的摘要改为 `StatGrid`（模式 / 地址总数 /
    起止地址 / 网段数量 / 截断标记），网段与地址列表改为 `DataTable`；
    `device-headers` 的 UA 解析结果改为 `StatGrid`（浏览器 / 版本 / 系统 / 设备类型 /
    渲染引擎 / 是否爬虫），请求头列表改为 `DataTable`
  - **其它**：`caa-gen` 的 CAA 记录改为表格
  - `docs/standards/01-frontend.md` §10 新增硬规则：多字段结果**禁止**压成文本行，
    产物与摘要**必须**分离，页签选中项**禁止**持久化
- **结果区不再重复展示与表格等价的原始文本**
  - `ResultPanel` 新增 `hideText`：结果已由 `DataTable` / `StatGrid` 覆盖时不再渲染 `<pre>` 文本块，
    文本仍作复制 / 下载 / 回填的数据源
  - 首批应用：`mime-lookup`、`http-status`、`hash-calc`、`country-code`、`cn-regions`、
    `ip-range-expand`、`device-headers`（此前表格被上千行文本压在下方）
- **多记录结果补表格**：`mock-data`（类型 / 取值）、`archive-inspect`（条目清单）、
  `password-policy`（检查项）、`eml-parser`（头字段）、`regex-tester`（匹配列表，含捕获组列）、
  `color-convert`（各格式取值）由文本行 / 卡片改为 `DataTable`，摘要改 `StatGrid`
- **复制按钮改为图标式**：`CopyButton` 只保留图标，悬停 / 聚焦在**上方**显示 tooltip（固定「复制」），
  复制结果改由 `ElMessage` 轻提示（已复制 / 复制失败，连续复制合并计数），全站生效；
  `DataTable` 新增 `copyColumns`（多列内嵌复制，按钮贴值右侧）与 `badgeCopy`（点击彩色徽标即复制，不另加按钮），
  并取消尾部整行列；`http-status` 的状态码改为点击徽标复制，原因短语与含义各带内嵌复制按钮
- **HTTP 状态码按类别着色**
  - `DataTable` 新增 `cellLevels`（按单元格着色）与 `badges`（彩色胶囊标签，浅底 + 同色文字 + 同色描边，
    只用 `--tb-*` 语义令牌，不引入 `ElTag`）
  - `http-status` 的状态码以彩色标签呈现（1xx 灰 / 2xx 绿 / 3xx 蓝 / 4xx 橙 / 5xx 红）；
    废弃条目不整行标黄，改为整行次级文字并在说明里附「已废弃」
  - 类别筛选项去掉 `1xx` / `2xx` 之类的数字前缀，直接显示「信息 / 成功 / 重定向 / 客户端错误 / 服务端错误」
- **MIME 类型查询与 HTTP 状态码查询改为「默认全量列表 + 即时过滤」**
  - 后端 `mime-lookup` 请求改为 `{ query, category }`（都可省略，空 = 全量），
    响应改为列表 `{ query, category, entries[], total, dataVersion }`；
    `http-status` 允许空请求（`kind: "all"`）并新增 `category` 过滤
  - 匹配范围扩展到媒体类型、扩展名、历史别名、用途说明与顶层类型；精确命中排在最前
  - 前端首屏拉全量，搜索与类别筛选在**本地即时完成**（零延迟、无 loading 闪烁）
  - 新增 `useLiveFilter` 组合式（一次取数 + 本地筛选）与 `SearchField` 组件
    （自带字段标签、清空按钮、可选自动聚焦），并登记进
    `docs/standards/01-frontend.md` §2 与 §4.3
- **查表类工具改为「键入即筛选」**
  - `mime-lookup` 与 `http-status` **去掉整条运行栏**（运行 / 重置 / 填入示例），
    页面进入即取数，**键入**搜索词即时收敛，搜索框自带 X 一键清空
  - 进入页面**自动聚焦**搜索框，落地即可打字
  - 首屏取数**不再依赖** catalog 的 `autoRun`；`useAutoRun` 移除随之无人使用的
    `immediate` 选项（`useAutoRun` 仍用于「每次输入都打后端」的工具）
- **国家代码与中国行政区划改为「默认全量 + 即时筛选」**
  - `country-code` 与 `cn-regions` 去掉运行栏，改为 `useLiveFilter`、`SearchField`
    与 `FilterChips`：`country-code` 用**首字母 A–Z** 标签筛选，`cn-regions` 用
    **层级**（省级 / 地级 / 县级）标签筛选，标签上带当前条件下的条目数
  - `cn-regions` 的「限定省份」保留为独立输入框，同样在**前端本地**过滤
  - 后端 `cn-regions` 新增 `level` 请求字段（`province` \| `city` \| `district` \| `all`），
    缺省行为与响应结构**不变**；`level=all` 返回全部层级供前端一次性取回
  - 结果区用 `StatGrid`（命中条目 / 数据集条目数 / 数据集版本）+ `DataTable`
    （代码 / 名称 / 层级 / 所属省份 / 上级代码）呈现，代码列为可点击复制的彩色标签
- **修复 `DataTable` 滚动容器失效**
  - UnoCSS 的 `.table` 工具类是 `display: table`，而调用方普遍写
    `<DataTable class="table">`，导致外层变成 table 盒子：
    `max-height` 与 `overflow: auto` 双双失效，长表格会撑开整页而不再内部滚动
  - `DataTable` 的滚动容器改为显式 `display: block`（影响全部 20 个使用方，
    `http-status` 等页面同样从「撑开整页」变回「面板内滚动」）
- **DNS 查询改为多服务器对比**
  - 前端提供 8 家**预设**解析器（Cloudflare / Google / Quad9 / AliDNS / DNSPod /
    114DNS / OpenDNS 与 DoH 入口，多选）与**自定义**服务器逐台追加，合计上限 8 台；
    默认选中 Cloudflare、Google 与 AliDNS
  - 结果用页签呈现：**第一个页签是「汇总」**（成功 / 失败台数、最快 / 最慢、TTL 区间、
    **记录一致性差异表**），其后每台服务器一个页签（协议、耗时、记录表）
  - 后端各台服务器改为**并发**查询（上限 4 路，结果按下标回填保持顺序），
    8 台服务器可在工具 20 秒预算内完成
- **CIDR 计算改为可视化**
  - 结果用页签呈现：**概览**（`StatGrid` 统计卡片 + 地址区间条 + 32/128 位前缀位图）
    与**子网**（标签网格，逐个可复制，超过 64 条折叠）
  - `cidr-calc` 与 `dns-lookup` 的 catalog `wide` 改为 `true`，宽屏下左右分栏展示
- **词条悬空键清理**：60 处 `t('tools.<slug>.<path>')` 引用原本在语言包中**查不到**，
  界面上直接显示原始键名（如 `tools.keygen.views.summary`）
  - 补齐 26 个**页签**词条（`asn1-view`、`cert-chain-fetch`、`cert-chain-repair`、
    `cert-gen`、`crl-gen`、`csr-gen`、`keygen`、`ocsp-request`、`p7b-extract`、
    `pfx-parse`），`keygen` 的 `views` 从误放的 `options.views` 提升到顶层
  - 补齐 34 处**既有**悬空键（`cert-lifecycle`、`password-policy`、`totp`、
    `ssh-keygen`、`html-escape`、`device-headers`、`ip-range-expand`、
    `color-convert`），并把 3 处指向错键名的代码改到正确键上
  - 新增 `frontend/src/i18n/usage.test.ts`：逐个校验工具页的**字面量**键引用在
    语言包中真实存在，防止悬空键再次进入主干
- **二维码工具页（`qr-code`）重构**：结果区从「一坨 SVG 源码」改为「预览优先」
  - 生成模式：结果区用页签呈现——**预览**（白底大图 + PNG 下载）与 **SVG 源码**；
    摘要用 `StatGrid` 给出模块边长、模块像素边长、图片边长与纠错级别
  - 识别模式：上传后显示缩略图、文件名与大小；识别结果可**「作为新输入」**回填到生成输入
  - 前景色 / 背景色改用 `ElColorPicker` 取色器（仍可手输十六进制），非法值即时提示；
    输入区补上此前写了却从未渲染的尺寸、静默区与纠错级别说明，以及内容字节计数
  - 切换操作时清空上一次的结果与页签
  - catalog 的 `wide` 改为 `true`：宽屏下输入在左、二维码在右
- **`DownloadButton` 支持二进制下载**：新增可选 `base64` 入参（与 `content` 二选一），
  二进制产物（PNG 等）不必再绕道文本下载；解码逻辑抽到
  `frontend/src/components/tool/base64.ts` 并附单元测试

### Removed

- **移除 `wifi-qr`（访客 WiFi 二维码）**：工具总数 98 → **97**
  - 该工具与 `qr-code` 的生成能力完全重叠，且结果区会**明文回显**含网络口令的
    `WIFI:T:…;S:…;P:…;` 载荷；访客网络应改用 `qr-code` 生成，避免口令在页面上长期可见
  - 同步删除后端实现与单测、契约测试数据、catalog 条目、前端模块与双语词条；
    数据分类由 14 个工具变为 13 个
  - 候选与取舍理由保留在 [docs/tool-candidates.md](docs/tool-candidates.md) §2

### Fixed

- `device-headers` **完全不可用**：前端把 `customHeaders` 当作**字符串**下发，而后端
  的类型是 `map[string]string`，每次调用都返回 `INVALID_FIELD`；现在前端按
  「每行一条 `名称: 值`」解析成对象后再提交
- `ip-range-expand` 的响应字段名与后端不符（前端写 `cidr` / `total`，后端实际返回
  `cidrs` / `count`），页面的网段与地址总数一直显示为空
- `ip-range-expand` 与 `device-headers` 在结果区显示与可见内容不一致时不再出现
  「（空结果）」：复制文本改为按实际模式取值

### Added

- **`base64` 支持标准变体与 URL 安全变体互转**：新增 `operation: convert`
  - 请求新增 `targetVariant`（`standard` \| `url`，默认 `url`）；
    `variant` 在 `convert` 下可取 `auto`，按输入用到的字符自动判断来源变体
  - 转换在**字节层面**完成：先按来源变体解出字节，再按目标变体重新编码，中间的字节不丢失；
    忽略换行、空白与 `=` 填充
  - 两组字母表混用（既含 `+`/`/` 又含 `-`/`_`）返回 `INVALID_BASE64`；
    `convert` 的输入传输编码只接受 `utf8`
  - 前端参数区新增「目标变体」下拉与**反向转换按钮**（把当前结果作为新输入再转一次，
    标准 → URL 安全 → 标准一键来回），切换操作时自动归一化互斥选项
    （`auto` 变体与 `hex` 输入编码只在各自适用的操作下出现）
  - 补 6 个单元测试用例、3 个契约测试用例、14 个前端表单纯函数用例与双语词条

- 工具目录扩容（工具总数 72 → **98**），新增 **26 个工具**，覆盖十个分类：
  - **文本类**（5）：`text-diff`（文本行级与 JSON 结构化差异对比）、`regex-validate`（只校验正则语法并给出错误位置）、
    `regex-build`（由结构化选项拼装正则并逐段释义）、`data-mask`（手机号 / 证件号 / 邮箱等敏感数据脱敏）、
    `text-stats`（字数、行数与字符构成统计）
  - **数据类**（8）：`id-validator`（身份证 / 信用代码 / 手机号 / 银行卡校验位计算）、`csv-convert`（CSV 与 JSON 互转）、
    `json-to-types`（由 JSON 样本生成 TypeScript / Go / Rust / Java 类型定义）、`mock-data`（按字段类型生成假数据）、
    `archive-inspect`（ZIP / TAR / GZIP 条目清单，含解压体积上限）、`file-analyze`（按魔数识别文件类型并给出摘要）、
    `wifi-qr`（访客 WiFi 接入二维码）、`id-generator`（UUID / ULID / 雪花 ID / NanoID 统一生成）
  - **系统类**（4）：`nginx-config`（全局与站点双模式配置生成）、`firewall-rules`（ufw / firewalld / iptables / nftables 规则生成）、
    `docker-compose-gen`（由服务定义生成 Compose 配置）、`lookup-table`（端口、退出码与时间戳格式速查）
  - **网络类**（2）：`device-headers`（按平台生成客户端请求头，并反向解析 User-Agent）、`ip-range-expand`（IP 段与 CIDR 互转及批量展开）
  - **格式化 / 编码 / 文档类**（3）：`color-convert`（颜色格式转换与 WCAG 对比度检查）、`html-escape`（HTML 实体编解码）、
    `markdown-convert`（Markdown 与 HTML 互转）
  - **加密 / 证书 / 密钥类**（4）：`password-policy`（口令策略合规检查）、`totp`（TOTP 生成与校验，附 RFC 6238 标准向量）、
    `cert-lifecycle`（证书有效期区间与到期倒计时）、`ssh-keygen`（OpenSSH 格式密钥对与 SHA256 指纹）
  - 新工具**零新依赖**：全部复用既有依赖树（含 `goldmark`、`golang.org/x/net/html`、`go-qrcode` 与 `tongsuo-go`）
  - 每个新工具均有**正例 + 反例**单元测试与**契约测试数据**；`totp` 另附 RFC 6238 标准向量，
    `ssh-keygen` 另附 OpenSSH 格式往返校验
  - 前端新增 26 个工具页（含共享组件复用）与齐备的双语词条（`zh-CN` / `en-US` key 集合一致）

- 新增两个**网络类**工具（工具总数 70 → 72），均**不出网**、纯查表：
  - `mime-lookup`（MIME 类型查询）：按扩展名查 MIME 媒体类型，或按媒体类型反查常用扩展名；
    接受 `png` / `.png` / `report.pdf` 与 `image/png`，媒体类型带参数时自动去掉参数；
    历史别名（如 `image/x-png`）归一到规范类型；未收录**不是**错误，返回 `matched: false`
  - `http-status`（HTTP 状态码查询）：按状态码（`100`–`599`）查原因短语与含义，
    或按关键字检索（同时匹配原因短语、中文说明与类别中文名，`not found` / `重定向` 都能用）；
    落在区间内但未收录的状态码仍给出所属类别，越界返回 `INVALID_FIELD` 且不回显输入
  - 两者的数据集是**内置快照**（`backend/internal/tools/network/dataset/*.csv`），
    **不**读取系统 MIME 数据库、**不**联网，响应带 `dataVersion` 便于追溯
  - 前端新增 `frontend/src/tools/network/mime-lookup` 与 `frontend/src/tools/network/http-status`，
    双语词条齐备（`zh-CN` / `en-US` key 集合一致）
- 设置面板新增「清理缓存」：一键恢复默认配置
  - 按 `toolbit.` 前缀删除语言、主题、主色、侧栏、分栏比例、收藏与最近使用，
    同时清空本源 Cache Storage；同域其他应用的存储不受影响
  - 两步操作（点「清理缓存」展开确认区，再点「清理并重载」），避免误触丢收藏
  - 清理后重新加载页面，页面上残留的工具输入与分栏比例一并复位
  - 顺带修正：首屏未把持久化的语言应用到 i18n，刷新后界面语言会回到简体中文
- 品牌标识（Toolbit 徽标）：矢量重绘，贴合品牌名语义
  - 图形为「T 型手柄 + 六角批头」：`T` 既是首字母，也构成 T 型螺丝刀轮廓；
    底端尖头六角是批头（bit），合起来正好是「Tool + Bit」
  - `frontend/public/favicon.svg` 为唯一矢量源；`favicon-32.png` / `favicon.ico` /
    `apple-touch-icon.png` 均由它生成，不手工维护第二套图形
  - `frontend/src/components/common/BrandMark.vue`：顶栏品牌标识组件，内联 SVG
  - 配色走新增的 `--tb-brand-from` / `--tb-brand-to` / `--tb-brand-fg` 令牌，
    **固定靛蓝**，不随主色切换 —— 徽标是品牌锚点
  - 顶栏原有的 `◆` 占位字符与 `.header__logo` 样式一并移除
- `deployments/Dockerfile`：四阶段构建（`tongsuo` → `frontend` → `backend` → `runtime`）
  - Tongsuo 8.4.0 在镜像内编译安装到 `/opt/tongsuo`（`--libdir=lib64 shared enable-ntls`），
    安装路径与 `tongsuo-go` 的 cgo 指令、rpath 严格一致
  - 源码包先按 sha256 校验再解包；基础镜像固定 digest
  - 运行镜像基于 `debian:bookworm-slim`，非 root（UID 10001），含 CA 证书与时区数据，
    只带二进制 + `/app/web` + Tongsuo 运行库，**不**含 Go / Node / 编译工具链
- `deployments/docker-compose.yml`：单容器编排，命名卷 `toolbit-data`、
  `cap_drop: ALL`、`no-new-privileges`、内存上限、日志轮转、`STOPSIGNAL SIGTERM`
- `deployments/.env.example`：16 个 `TOOLBIT_*` 与 6 个编排变量的默认值模板（无任何密钥）
- `.dockerignore`：排除依赖、构建产物、文档、测试与密钥
- `toolbit healthcheck` 子命令（`backend/cmd/toolbit/healthcheck.go`）：
  探 `http://127.0.0.1:<TOOLBIT_ADDR 端口>/readyz`，3 秒超时，非 200 退出 1、未知子命令退出 2，
  不输出响应体与配置值 —— 兑现 `docs/api/admin.md` §9.1 早已写明的契约
- Makefile 新增 `docker-build` / `docker-up` / `docker-smoke` / `docker-logs` / `docker-ps` /
  `docker-shell` / `docker-down` / `docker-clean`
- 前端体验改版：设计令牌、全局检索与工具页交互的一次性改造
  - 12 组主色预设（明暗双套 × 5 令牌），可在设置面板切换；
    语义色（`success` / `warning` / `danger` / `info`）**保持固定不随主色漂移**
  - 主色对比度门禁 `scripts/check-accent.mjs`：52 项 WCAG 对比度校验（阈值 4.5:1）、
    色板与令牌一致性校验、语义色漂移校验；已接入 `make lint`
  - 命令面板：`Ctrl/⌘ + K` 全局唤起，覆盖 70 个工具、10 个分类与 3 个内建页，
    中英文名任一命中即匹配，支持 `↑` `↓` 选择、`Enter` 打开、`Esc` 关闭、Tab 焦点陷阱
  - 全局快捷键：`/` 聚焦搜索、`?` 打开快捷键说明、`Alt+↑/↓` 在最近使用中切换
  - 全屏切换：顶栏按钮经 Fullscreen API 进入 / 退出全屏，不改 URL、不写存储
    （原「专注模式」靠 `?focus=1` 隐藏侧栏，与折叠侧栏按钮职责重叠，已移除）
  - 顶栏三段式布局：品牌贴左并与侧栏菜单对齐、搜索在视口内居中且加长、
    全屏 / 主题 / 设置图标贴右边缘
  - 侧栏字号整体上调：菜单项 14 → 15px，图标胶囊 26 → 28px，侧栏宽 260 → 272px
  - 工具页「使用说明 / 常见问题」改为标题即折叠开关，默认展开
  - 工具页双栏：仅 `wide` 工具在视口 ≥1280 时左右分栏，分隔条可拖拽，比例记忆在 `localStorage`
  - 双栏两列改为**等高**，并去掉左栏的内部滚动条（内容超出时由页面滚动）
  - 侧栏分固定区 / 滚动区：快捷入口、分类标题与过滤框固定，只有分类列表滚动
  - 过滤框下移到分类标题之下（它只筛分类下的工具）；分类标题右侧新增一键收起 / 展开全部
  - 分类下每个工具项右侧新增收藏星标（淡出，hover / 已收藏时可见，与工具页收藏同源）
  - 结果区结论徽章（通过 / 未通过 / 风险 / 未知）与网络类工具的汇总统计条
  - 结果区「作为新输入」按钮：把结果回填输入区（仅内存，不写任何存储）
  - 输入即时校验：纯前端语法检查（JSON / 十六进制 / Base64 / UUID），不发起网络请求
  - 可点击的快速参考面板：正则元字符、cron 表达式、进制写法一键填入
  - 共享组件补齐 10 个：`SensitiveInput` `FileTextInput` `JsonViewer` `ByteView`
    `Asn1Tree` `CertChainTable` `CertDetailDialog` `KeyValueTable` `DownloadButton`
    `ProgressStream`
  - 结果多视图页签：`json-format` 支持「原文 / 结构化」双视图
  - `autoRun` 死字段接通：22 个纯本地工具在参数变更后自动运行（400ms 防抖），
    敏感、出网与高成本工具一律不开启
  - `ssl-check` 综合评级（A+ / A / B / C / T）与逐项检查清单
  - 单元测试：命令面板 9 条、主色偏好 9 条、SSL 评级规则 7 条
- 侧栏导航改版：可扫读的图标轨道与悬浮子菜单
  - 折叠态收成图标轨道：分类只显示图标（不再显示两字缩写）；展开 / 收起带 180ms 宽度过渡，
    顶部标题与过滤框用 `grid-template-rows` `0fr`→`1fr` 同步收高，不再一跳到位；
    顶栏左段随侧栏同宽过渡，折叠按钮不会瞬移
  - 折叠态悬浮或聚焦分类图标弹出子菜单（`Teleport` 到 `body`，按视口坐标定位并做上下边界校正），
    列出该分类全部工具并展示一句话描述，悬停轨道与菜单之间留 120ms 关闭延迟
  - 悬浮子菜单的描述按需加载：`frontend/src/tools/toolDescriptions.ts` 用 `import.meta.glob`
    读取既有 `locales/<lang>/tools/<slug>.json`，只在悬浮某个分类时加载该分类的词条，
    不复制文案、不进初始包
  - 展开态工具项补齐图标：`frontend/src/tools/toolIcons.ts` 按 slug 登记，未知 slug 回落分类图标
  - 工具项 hover / 选中底色覆盖**整行**（含收藏按钮）；收藏星标 14 → 16px，工具页标题旁 17 → 20px
  - 分类子项去掉左侧引导竖线并整体左移；修掉「全部工具」在任意工具页都被判为激活（`/` 是路径前缀）
  - 工具页面包屑改为「首页 / 分类 / 工具名」，中间层可点回分类页
  - 顶栏品牌去掉「70 个 IT 工具」副标题，品牌名放大到 24px、徽标 28 → 30px
  - 顶栏伸缩按钮换成带方向的面板图标（`panel-collapse` / `panel-expand`）：展开态指向左（收起）、
    收起态指向右（展开）；窄屏抽屉同理
  - 「快捷」与「分类」两组之间加分隔线并留出组间距（`space-3` 上边距）；轨道态分隔线收窄为图标宽度
- 顶栏新增 GitHub 入口：项目仓库图标（实心剪影）新标签页打开，悬浮提示两行
  （「GitHub 仓库」+「项目开源，可用 Docker 在局域网内自行部署，数据不出内网」）；
  地址统一取 `frontend/src/config/site.ts` 的 `REPO_URL`，不另写死链接

### Changed（界面与视觉）

- 图标改为**彩色**：不再统一成一个主色，改为每个分类、每个工具各有各的色调
  - 新增 **12 组固定色调**令牌 `--tb-tone-<名>` / `--tb-tone-<名>-soft`
    （`tokens.css`，明暗两套），与主色（`--tb-primary`）**正交**：主色负责交互高亮，
    色调负责图标识别
  - 新增映射表 `frontend/src/tools/tones.ts`（`CATEGORY_TONES` 分类 → 色调、
    `TOOL_TONES` 工具 → 色调，未登记时回落到所属分类）
  - 生效范围：工具卡片、分类标题、侧栏分类胶囊与工具项、分类页标题、工具页标题
  - `scripts/check-accent.mjs` 门禁扩展：新增色调对底色的 **4.5:1 对比度**校验，
    并校验色调**不随主色漂移**（12 组 / 24 项）
- 首页「全部工具」改为**按分类分组**展示：每个分类一节，组标题**可点击**进入该分类的
  工具列表页（`CategoryHeader` → 路由 `category`）；搜索时同样保持分组结构
  - 新增组件 `frontend/src/components/discovery/CategoryHeader.vue`，
    移除 `CategoryCard.vue`（其职责由分组标题承接）
  - 分类页标题增加带色调的分类图标胶囊

- `password-gen` 增强：新增**策略模板**（长度 / 字符类 / 排除项）、可发音口令与助记词口令，
  排除易混淆字符集（`I` `l` `O` `0` `1`）
- `jwt-codec` 增强：新增**非对称签名验证**（`RS256` / `RS384` / `RS512` / `PS256` / `PS384` / `PS512` / `ES256` / `ES384`），
  并输出签名算法与密钥匹配结论

- 部署产物目录由文档中的 `deploy/` 统一为 **`deployments/`**（`AGENTS.md`、
  `docs/architecture.md`、`README.md`、`.gitignore` 同步）
- `Makefile`：`docker-build` 目标补进文件头「常用目标」清单；该目标现在同样会在缺
  `deployments/.env` 时用模板生成一份，并在结束时回显构建出的镜像名
  （`docs/deploy.md` §1.1 与 §8 同步）
- 容器内路径口径统一：前端产物 `/app/web`、统计库 `/app/data/toolbit.db`
  （原先 `deploy.md` 写 `/app/web` + `stats.db`，`architecture.md` 写 `/app/wwwroot` + `toolbit.db`）
- `docs/standards/08-deploy.md`：阶段表、基础镜像选型、Tongsuo `--libdir=lib64` 约定、
  digest 固定口径、探针实现方式、compose 位置与禁止清单更新
- `docs/deploy.md`：Docker 章节改写为 `make docker-up` 为主线，补编排变量表与容器故障排查项
- `docs/api/admin.md` §9.1：补充探针地址来源、退出码与输出约束
- 字号体系整体上调一档（正文基准 14px → 15px），全站硬编码字号清零
- 工具卡片增加分类图标锚点，hover 响应主色并预取对应 chunk
- 侧栏 ≤900px 从「整块隐藏」改为抽屉；新增内联过滤框与筛选后的空分组清理
- 页面标题随路由更新（此前全部页面标题恒为静态 "Toolbit"）
- FAQ 与使用说明改为可折叠；`ToolPanel` 无参数时不再渲染空标题
- `KeyInput` 的口令输入改用 `SensitiveInput`，默认遮蔽并提供显示切换
- 首屏防闪烁：`index.html` 的 `<head>` 内联脚本在 Vue 之前应用已保存的主题与主色
- 新增 `noscript` 兜底；新增两枚 `theme-color` meta
- `CopyButton` 新增 `size` 属性，支持表格内的紧凑场景
- Element Plus 主色梯度改为由 `--tb-primary` 经 `color-mix()` 派生，不再写死

### 新增的本地存储键

- `toolbit.split`：工具页左右分栏的比例（`0.25` ~ `0.75`），属界面偏好

> 工具输入、密钥、证书与历史结果**一律不写**任何浏览器存储（见 [docs/privacy.md](docs/privacy.md) §5）。

## [0.7.0] - 2026-10-29

网络探测：出网能力就继，SSRF 防护全覆盖。
**新增 7 个工具**（累计 43 个已实现），`ct-search` 本版未实现。

本版引入统一的出网防护层 `netguard`，并把 `ssl-check` 做成 WebSocket 流式端点。

### Added

- 出网防护层 `backend/internal/netguard/`
  - 私网 / 回环 / 链路本地 / `0.0.0.0/8` / 运营商级 NAT / 文档与保留网段全部拦截
  - IPv4-mapped IPv6（`::ffff:127.0.0.1`）先还原再判定，无法绕过
  - 建连使用**已校验的 IP**（防 DNS rebinding），不二次解析
  - 端口白名单固定为 43 / 53 / 80 / 443 / 853，用户无法指定任意端口
  - 重定向逐跳校验，最多 3 跳；禁用环境代理
  - `TOOLBIT_NETWORK_ALLOWLIST` 支持精确主机名与 `*.suffix`
- 工具 `cidr-calc`：网段计算、地址范围反推、按前缀拆分；IPv6 大数以十进制字符串表达
- 工具 `dns-lookup`：`udp` / `tcp` / `dot` / `doh` 四种协议，逐服务器结果与耗时
- 工具 `dns-tester`：RDAP 优先、whois（TCP 43）回退；DNSSEC 逐级 DS 检查；PTR + 正向复核
- 工具 `ssl-check`：WebSocket 流式 TLS 探测，四个阶段的进度帧
- 工具 `cert-chain-fetch`：只呈现服务端实发链，附链摘要与验证结论
- 工具 `domain-verify`：`dns-txt` / `dns-cname` / `http-file` 三种方式
- 工具 `websocket-test`：浏览器原生直连，**无**后端端点
- WebSocket 端点 `GET /api/v1/tools/ssl-check/stream`
  - 首帧 5 秒超时、心跳 30 秒、空闲 60 秒兜底、总预算 30 秒
  - 同 IP 同时最多 1 条连接，超出在升级前返回 `429`
  - 关闭码：`1000` / `1008` / `1011`
- 前端 8 个工具页与 16 份双语词条；流式页面自带四态状态机
- 契约测试数据 5 份；`netguard` 与网络工具层单元测试

### Changed

- catalog 传输方式校正
  - `ssl-check` → `transport: websocket`（此前误标为 `json`）
  - `websocket-test` → `transport: client`（与「纯前端」的定义对齐）
  - `cert-chain-fetch` → `transport: json`，请求体上限降到 1 MiB
- 分发器按 `transport` 分流：`client` 工具返回 `404`，`websocket` 工具的 JSON 通道返回 `405`
- 端口**白名单**判定从字段校验下移到 `netguard`：字段层只校验范围，越界端口返回
  `SSRF_BLOCKED`，策略只有一个事实来源
- `toolset.Definitions` 需要注入出网守卫，进程入口与 API 测试共用同一份装配

### Fixed

- 修正若干工具在边界输入下的行为：CSV 转换方向此前调反、IP 区间倒置未被拒、
  CIDR 聚合产出多余网段、HTML 转 Markdown 丢失正文、Markdown 递归参数传递错误
- `totp` 在未显式指定 `digits` / `period` 时被误拒（校验早于默认值填充）
- `ssh-keygen` 的 RSA 私钥因 ASN.1 结构多出一个字段而无法解析

- `cidr-calc` 的地址族字段此前返回位宽（32/128），改为 `4` / `6`

### Known limitations

- `ct-search` **未实现**：公开证书透明度日志查询服务没有稳定 SLA，
  catalog 保留登记，调用返回 `NOT_IMPLEMENTED`
- TLS 客户端使用 Go 的 `crypto/tls`：仅限出网探测，链验证仍走 `pkisvc.VerifyChain`
  （条款见 `docs/standards/03-crypto.md`）
- 本地信任库可能缺失：`ssl-check` 届时只能判定链内自洽，
  响应里用 `verified` / `implicitTrust` 显式区分
- 系统解析器（`/etc/resolv.conf`）属运维配置，不做私网校验；
  用户自定义的 DNS 服务器仍要过出网防护
- TLS 1.3 的密码套件无法通过配置裁剪

---

## [0.6.0] - 2026-10-29

PKI 生成、吊销与容器：能签发证书、处理吊销，并读写 PKCS#12 / PKCS#7 容器。
**新增 12 个工具**（累计 36 个已实现）。

本版把 `pkisvc` 从「X.509 只读侧」扩展到**写侧**：CSR / 证书 / CRL 签发、
离线吊销判定、OCSP 请求与响应、PKCS#12 与 PKCS#7 容器。

### Added

- PKI 服务层新增 8 个文件（`backend/internal/pkisvc/`）
  - `certgen.go`：DN 构造（`TryAdd` 白名单）、随机序列号、
    **分步构建**的证书签发（`NewCertificate` → `Set*` → `Add*` → `Sign`）
  - `csrgen.go`：CSR 生成（`NewEmptyCertificateRequest` 分步构建）
  - `crlgen.go`：CRL 签发与吊销原因码白名单（拒绝保留值 7）
  - `revocation.go`：`RevocationList` 句柄与 `CheckRevocation`（**先验签后判定**）
  - `ocsp.go`：OCSP 请求生成与响应解析（含可选验签）
  - `pkcs12.go`：打包 / 解析 / 改密；口令错误按**操作名**判别
  - `pkcs7.go`：证书袋打包与提取
  - `store.go`：信任库句柄（每请求新建）
- 工具 `csr-gen`：PKCS#10 请求，按证书类型预设 KU / EKU
- 工具 `cert-gen`：自签或 CA 签发，SKID / AKID 自动补齐
- 工具 `crl-gen`：CA 签发 CRL，原因码白名单
- 工具 `revocation-check`：离线吊销判定（未验签时只报 `unverified`）
- 工具 `ocsp-request` / `ocsp-response`：OCSP 请求与响应（可选 `trustedRoots` 验签）
- 工具 `pfx-pack` / `pfx-parse` / `pfx-change-password`：PKCS#12 容器
- 工具 `p7b-pack` / `p7b-extract`：PKCS#7 证书袋
- 工具 `caa-gen`：CAA 记录生成（纯字符串拼装，不进 `pkisvc`）
- 前端 12 个工具页与 24 份双语词条；夹具 `frontend/src/tools/certificates/fixtures-v060.ts`
  （一次性生成的**测试密钥**，非真实密钥）
- 契约测试数据 12 份；服务层与工具层单元测试各新增一批用例
- 生成脚本 `scripts/gen-v060-fixtures.sh` / `gen-v060-contracts.py` / `gen-v060-web-fixtures.py`

### Changed

- **契约扩展**（只加**可选**字段）
  - `revocation-check` 新增 `issuer`（提供时先验签）、`atTime`
  - `ocsp-response` 新增 `trustedRoots`，响应新增 `signatureVerified`
- **契约收窄**
  - `cert-gen.templateCert` **只**复用公钥与有效期，**不**拷贝扩展
  - `crl-gen.revoked[]` **只**接受证书，去掉「按序列号吊销」选项
  - `revocation-check` 的 `status` 新增 `unverified` 取值；未验签时**绝不**返回 `good`
- 新增警告码 `CRL_UNVERIFIED`
- 路线图：`docs/roadmap.md` §9 全部任务勾选，§9.3 补 M-T9 – M-T12

### Fixed

- 口令错误判别不再依赖 OpenSSL 错误栈文本：该栈在成功调用后不会清空，
  同进程内前一次失败的条目会覆盖本次原因，导致「容器损坏」被误报成「口令错误」

### Known limitations

- `cert-gen.templateCert` **不拷贝扩展**：上游读证书只能拿到扩展 DER，
  没有「按 DER 添加扩展」的入口（见 `docs/upstream/tongsuo-go-issues.md` I-19）
- `crl-gen` **无法按序列号吊销**：上游 `CRLBuilder.Revoke` 只接受证书对象（见 I-20）
- 上游的便捷签发入口会**立即签名**，签名后 `Add*` 被「已签名锁」拒绝（见 I-21）；
  Toolbit 因此固定走分步构建
- `pfx-change-password` **只**支持本库打包的容器：上游 `PKCS12_newpass`
  无法对外部实现产出的容器改密（见 I-22）
- `p7b-pack` / `p7b-extract` **只做证书袋**：上游 PKCS#7 尚无签名 / 加密（见 I-11）
- 序列号上限 8 字节：上游用 `int64` 承载（与 I-18 同源），超出返回 `SERIAL_INVALID`
- `ocsp-response` 未提供 `trustedRoots` 时**不验签**，只解读响应内容

---

## [0.5.0] - 2026-10-29

PKI 解析与链验证：能读懂证书，并能判定一条链是否可信。
**新增 8 个工具**（累计 24 个已实现）。

本版把 `pkisvc` 从「密钥与格式」扩展到完整的 **X.509 只读侧**：
证书 / CSR / CRL 解析、链排序、链验证、链修复与 ASN.1 展开。

### Added

- PKI 服务层新增 8 个文件（`backend/internal/pkisvc/`）
  - `cert.go`：证书加载与解析、PEM / DER / base64 / hex 四种输入；
    主体 / 签发者 / 有效期 / 扩展 / SAN / KU / EKU / SKID / AKID；
    指纹支持 `MD5` / `SHA1` / `SHA256` / `SHA384` / `SHA512` / `SM3`
  - `pemder.go`：PEM 块拆分、类型过滤（cert / key / csr / crl）、PEM ↔ DER 互转
  - `csr.go`：CSR 解析与自签名校验；SAN 由扩展手工解析
  - `asn1.go`：DER 展开为 TLV 树 + 十六进制转储；调用上游**之前**自测深度 / 节点 / 体积
  - `chain.go`：AKID → SKID 优先的链排序与链拆分
  - `chainverify.go`：逐层验证、域名匹配、`X509_V_ERR_*` 结构化错误项
  - `chainrepair.go`：从中间证书与受信任根里补链
  - `crl.go`：CRL 解析（签发者 / 时间窗 / 编号 / 吊销条目）
- 工具 `cert-parse`：证书详情、有效性状态、指纹、原始 DER
- 工具 `csr-parse`：主体、SAN、公钥、扩展、挑战口令**存在性**、自签名校验结果
- 工具 `pem-der`：PEM ↔ DER 互转、标签识别、块数
- 工具 `asn1-view`：DER 结构树 + 十六进制转储（catalog 标记 `wide`）
- 工具 `cert-chain-parse`：链排序、是否已排序、是否缺根、是否自签名
- 工具 `cert-chain-verify`：链可信性、错误列表、逐张证书结论、域名匹配、隐含信任锚标志
- 工具 `cert-chain-repair`：补全缺失的中间证书（纯本地）
- 工具 `crl-parse`：CRL 签发者、时间窗、签名算法、编号、吊销条目
- 共享输入类型 `backend/internal/tools/pembundle.go`：`tools.PemBundle`，
  与 `KeyInput` 共用同一套传输层解码逻辑
- 前端 8 个工具页与 16 份双语词条；共享组件
  `CertTextInput.vue`（证书族输入）与 `CodeEditor.vue`（**自渲染**只读树 / hex 视图）
- 测试夹具 `backend/testdata/fixtures/pki/`（自签根 CA + 中间 CA + 叶证书 + CSR + CRL）
  与生成脚本 `scripts/gen-pki-fixtures.sh`；**不含任何私钥**
- 契约测试数据 8 份；服务层与工具层单元测试各 1 个文件

### Changed

- **契约扩展**（只加**可选**字段，不改已有字段语义）
  - `cert-chain-repair` 新增 `intermediates`
  - `cert-chain-verify` 的 `trustedRoots` 语义明确为「受信任根列表」
  - `cert-parse` 的 `fingerprints` 扩展为**全部**哈希算法
- `cert-chain-parse` 的 `ordered: false` **只**表示输入需要重排，**不是**错误
- `asn1-view` 的请求体上限由 8 MB 降为 **1 MB**（DER 展开的内存放大）
- 契约测试与 501 用例改为**从 `toolset.Definitions()` 推导**，
  新增工具时不再需要手改测试清单
- 路线图：`docs/roadmap.md` §8 全部任务勾选；§8.3 补 L-T9 – L-T11 三条技术决策

### Fixed

- `backend/internal/api` 的「未实现工具返回 501」用例改用**推导**出来的未实现 slug，
  避免新工具落地后用例失效

### Known limitations

- `cert-chain-verify` **没有系统信任根可用**：上游未暴露默认信任库
  （见 `docs/upstream/tongsuo-go-issues.md` I-17）。
  未传入受信任根时，链内的自签名根被当作**隐含信任锚**，
  此时 `valid: true` **只**说明链内自洽；响应里的 `implicitTrust` 标出这种情况
- `csr-parse` 的 CSR 句柄**无法显式释放**：上游 `CertificateRequest` 没有公开的
  `Close()`（见 I-15），只能依赖 `finalizer` —— v0.5.0 唯一一处句柄释放例外
- `crl-parse` 的吊销序列号**超过 64 位时输出空字符串**：上游用 64 位整数承载序列号，
  超长时溢出为 `-1`（见 I-18），工具**不**输出溢出的错误值
- `asn1-view` 的深度 / 节点上限由 **Toolbit 侧**测量实现（上游 `asn1.Parse`
  无上限参数，见 I-16）；契约约定的 64 层 / 10000 节点不变
- CSR 的 SAN 由扩展**手工**解析（上游未暴露 `SAN()` 方法），只覆盖
  email / DNS / URI / IP / RegisteredID 五类
- 证书签发、吊销判定、OCSP、PKCS 容器均在 v0.6.0 范围，本版**不含**

## [0.4.0] - 2026-10-29

密钥管理：交付密钥全生命周期（生成 → 转换 → 解析 → 运算 → 配对）。
**新增 6 个工具**（累计 16 个已实现）。

本版同时建立 v0.4.0 – v0.6.0 共用的底座：新增 `pkisvc` 密码学层与
`toolset` 工具装配层，把「密码学原语」与「密钥 / PKI 结构」分开。

### Added

- PKI 服务层 `backend/internal/pkisvc/`：密钥与 **PKI 领域**
  - `Generate`：RSA / SM2 / EC / Ed25519 / Ed448 / X25519 / X448；
    公钥是**独立持有**的句柄，可与私钥分别释放
  - `Convert`：PKCS#1 / PKCS#8 / SPKI / DER / PEM 互转；
    DER 源靠**试加载**识别编码，加密私钥转换后**保持加密**
  - `Inspect`：算法、类型、位数、曲线、公钥指纹（SPKI DER 的 SHA-256）与 RSA 私钥参数
  - `EncryptPrivateKey` / `DecryptPrivateKey` / `ChangePassphrase` / `ExtractPublicKey`
  - `Match`：私钥 ↔ 公钥 / 证书，依据公钥指纹判定
  - `JWKFromPEM` / `PEMFromJWK` / `XMLFromPEM` / `PEMFromXML`：RFC 7517 与 .NET `RSAKeyValue`
- 工具 `keygen`：7 种算法；`privateFormat` 支持 `pkcs8` / `pkcs1`；
  可选用口令加密私钥（AES-256-CBC / DES-EDE3-CBC）；生成走重操作信号量
- 工具 `key-convert`：格式互转 + 源密钥识别（算法 / 编码 / 是否加密 / 位数 / 曲线）
- 工具 `key-parse`：密钥元信息、RSA 私钥参数与规范化 PEM
- 工具 `key-ops`：加密、解密、提取公钥、修改口令
- 工具 `key-match`：私钥 ↔ 公钥 / 证书配对检测，给出双方公钥指纹
- 工具 `jwk-xml-convert`：JWK / PEM / XML 三向互转
- 工具装配层 `backend/internal/toolset/`：汇总各分类 `Definitions()`，
  进程入口与 API 测试共用；测试守住 slug 唯一与 catalog 登记
- 共享输入类型 `backend/internal/tools/bytes.go` `keyinput.go`：
  `tools.Bytes` 与 `tools.KeyInput`，各工具包用类型别名复用
- 前端 6 个工具页与 12 份双语词条；共享夹具 `frontend/src/tools/keys/fixtures.ts`
  （一次性生成的**测试密钥**，非真实密钥）
- 契约测试数据 6 份；服务层与工具层单元测试各 1 个文件

### Changed

- **架构**：新增 `backend/internal/pkisvc` 作为**第二个**允许 `import tongsuo-go` 的层；
  `AGENTS.md` §4.1 与 `docs/architecture.md` 的依赖图同步更新
- **依赖守卫**：`.golangci.yml` 的 `depguard` 规则由「仅 cryptosvc」改为
  「cryptosvc + pkisvc」双白名单；`no-stdlib-crypto` 新增 `crypto/x509` 禁令
- **catalog**：`revocation-check` 由 `rateClass: network` + `requiresNetwork: true`
  改为 `standard` + `false`——其契约自 v0.1.0 起就是**离线**的「证书 + CRL」判定
- **路线图**：`docs/roadmap.md` 新增 §6.8「v0.4.0 起的跨版本公共基座」，
  并把 §7 / §8 / §9 细化为与 v0.3.0 同构的可执行计划（交付物 / 技术决策 /
  分阶段任务表 / 完成标准 / 不在本版范围 / 风险）

### Fixed

- `backend/internal/api` 的两处「未实现工具返回 501」用例改用仍未实现的工具
  （`cert-parse` / `cert-chain-fetch`），避免 `keygen` 落地后用例失效

### Known limitations

- `keygen` 的 `count` 字段**当前只支持 1**：响应结构只能表达一把密钥，
  批量生成需要新增响应字段（属破坏性契约变更），已在
  `docs/api/tools/keys.md` 与 `docs/roadmap.md` §7.7 双处登记
- `keygen` 的 `publicFormat: pkcs1` 与 `key-convert` 的公钥 → `pkcs1`
  **暂不支持**：上游 `tongsuo-go` 缺 PKCS#1 公钥的导出 / 加载入口
  （见 `docs/upstream/tongsuo-go-issues.md` I-14）
- `jwk-xml-convert` 的 JWK 侧**只支持 RSA 与 EC**（上游无 OKP，见 I-12），
  XML 侧**只支持 RSA**（见 I-13）
- `p7b-pack` / `p7b-extract`（v0.6.0）**只做证书袋**：上游 PKCS#7 尚无签名 / 加密（见 I-11）

## [0.3.1] - 2026-10-28

密码学文件工具：交付 `sm4-file-crypt`，让大文件加解密可用。
**新增 1 个工具**（累计 10 个已实现）。

本版打通 multipart 通道：此前分发器只认 JSON，
既无法接收表单，也无法返回二进制流。

### Added

- 文件容器 `backend/internal/cryptosvc/tbfc.go`：`TBFC` 分块 GCM 容器
  - 头（22 字节）：magic `TBFC`(4) + version(1) + flags(1) + chunkSize(4) + baseNonce(12)
  - 块：`length`(4) + ciphertext + `tag`(16)；`length = 0` 表示结束
  - 分块 nonce：`baseNonce[0:8] || uint32be(chunkIndex)`
  - AAD：`容器头 || uint32be(chunkIndex)`，同时绑定容器头与块索引
  - 流式读写，内存占用约 2×chunkSize，与文件大小无关
- multipart 通道 `backend/internal/tools/multipart.go`
  - `MultipartRequest` / `MultipartResult` / `MultipartHandler`
  - `MultipartForm`：把 `multipart.Reader` 的顺序读取包装成「先收小字段、再收文件」
  - `SanitizeFilename` / `ContentDisposition`：文件名来自客户端，**不可信**，必须清洗
- 分发器分流 `backend/internal/api/dispatch_multipart.go`
  - 按 catalog 的 `transport` 分流：`multipart` 走表单 + 二进制流通道
  - 二进制响应不套信封，警告改走 `X-Warnings` 响应头
  - 失败时**仍返回 JSON 信封**：错误在响应头发出之前就能确定
- 工具 `sm4-file-crypt`：TBFC 容器，分块 GCM，100 MB 上限
  - 解密**先同步**校验容器头与第一块：密钥不对、容器被篡改、不是容器
    这三类错误因此返回干净的 JSON 错误，而不是「200 + 被截断的下载」
- 前端 `callToolBinary`：multipart 请求 + 二进制响应，带上传进度
- 共享组件 `FileDropZone`：点击选择 + 拖放，单文件
- 工具页 `sm4-file-crypt`：操作、密钥、基础 nonce、分块大小与拖放区；结果以「下载」给出
- 组合式函数 `useFileToolRun`：与 `useToolRun` 同构，但结果是二进制文件，额外提供上传进度
- 契约测试执行器支持 multipart 分支；契约数据落在 `backend/testdata/contract/sm4-file-crypt.json`

### Changed

- 密码学规范 `docs/standards/03-crypto.md` §11 重写：容器格式逐字节落档，
  补「首块预校验」「multipart 顺序约束」两节
- 加密 API 契约 `docs/api/tools/crypto.md`：`sm4-file-crypt` 章节按实现重写，
  含容器格式与错误码表
- 通用 API 契约 `docs/api/README.md`：`X-Warnings` 的格式明确为 JSON 数组；
  §8 补 `transport` 分流；§4.5 与 §12.4 补「文件必须是最后一个 part」
- 架构文档 `docs/architecture.md`：§4.1 双通道、§4.2 两种工具形态、§4.3 补文件容器
- 路线图 `docs/roadmap.md`：v0.3.1 补齐 8 项任务清单与完成标准

### 破坏性变更

- multipart 工具的**文件部分必须是最后一个 part**：客户端必须先发所有普通字段，
  最后再发文件。服务端读到文件后就无法回头取后面的字段
- multipart 工具的**失败响应仍是 JSON 信封**，不是二进制流；
  客户端必须按 `Content-Type` 分流

## [0.3.0] - 2026-10-28

密码学核心：国密能力就绪，建立 `cryptosvc` 服务层与句柄生命周期规范。
**新增 7 个工具**（累计 9 个已实现）。

### Added

- 密码学服务层 `backend/internal/cryptosvc/`：**唯一**允许 import `tongsuo-go` 的层，所有密码学运算在此转调 Tongsuo
  - `Session` 句柄作用域：LIFO 释放、`Close` 幂等、警告累积、重操作信号量（≤ 2）、orphan 记录
  - Session 经 `ctx` 传递，**不**改变 `tools.Handler` 签名
  - `Digest` / `MAC` / `SymEncrypt` / `SymDecrypt` / `Sign` / `Verify` / `AsymEncrypt` / `AsymDecrypt` / `ECDHSharedSecret` / `ECDHDerive` / `RandomBytes` 服务方法
  - 底层错误统一映射为 `toolerr` 错误码，不回显底层英文文本
  - 密钥封装 `PrivateKey` / `PublicKey`：显式 `Close`，源缓冲用后清零
- 编码服务层 `backend/internal/codecsvc/`：utf8 / hex / base64 / base64url 归一化与 PEM / DER 识别
- 加密类工具 `backend/internal/tools/crypto/`：
  - `hash-calc` —— SM3 / MD5 / SHA1 / SHA224 / SHA256 / SHA384 / SHA512 多算法摘要，弱哈希告警
  - `hmac-gen` —— HMAC-SM3 与 HMAC-SHA 系列
  - `password-gen` —— CSPRNG 密码生成，拒绝采样避免模偏差，附带熵估算
  - `symmetric-crypt` —— SM4 / AES，ECB / CBC / CTR / OFB / CFB / GCM，GCM 自动生成 nonce 并回传
  - `signature` —— SM2 / RSA / ECDSA / Ed25519 / Ed448 签名与验签
  - `asymmetric-crypt` —— SM2 / RSA 公私钥加解密
  - `ecdh` —— P-256 / P-384 / P-521 / X25519 / X448 协商 + HKDF 派生
- 前端 7 个加密工具页 `frontend/src/tools/crypto/`，含共享 `KeyInput` 组件（密钥只存于组件局部状态，**不**写浏览器存储）
- 双语词条：7 个工具的 `zh-CN` / `en-US` 十项字段全覆盖；词条一致性测试改为用 `import.meta.glob` 自动收集，新增工具无需再改测试
- 契约测试执行器 `backend/internal/api/contract_test.go`：遍历 `backend/testdata/contract/*.json` 校验状态码、错误码、响应数据与警告
- 上游能力缺口登记 `docs/upstream/tongsuo-go-issues.md`：10 条缺口的现象 / 期望 / 验收标准 / 下游影响
- CGO 构建基线：`Makefile` 新增 `check-tongsuo`，`CGO_CFLAGS` / `CGO_LDFLAGS` 指向 `/opt/tongsuo`，以 rpath 固定运行期库路径

### Changed

- 密码学规范 `docs/standards/03-crypto.md`：新增 §1.1 运算归属、对称加解密的「算法 × 模式 × 填充」支持矩阵、签名结构预校验规则、GCM nonce 回传约定
- 架构文档 `docs/architecture.md`：§4.3 服务层改为与实现一致的包级函数清单；新增关键决策 D11「密码学运算只转调，不自研」
- 加密 API 契约 `docs/api/tools/crypto.md`：逐字段对齐实现，新增算法支持矩阵与实现状态列
- 产品文档 `docs/tools.md`：加密章节新增状态列，移除未实现算法的承诺
- 路线图 `docs/roadmap.md`：v0.3.0 补齐 6 个阶段的任务清单与验收标准；新增 v0.3.1（`sm4-file-crypt`）
- 工程门禁 `.golangci.yml`：新增 `depguard` 规则 `no-stdlib-crypto`，禁止在 `cryptosvc` 之外导入标准库密码学包
- 协作契约 `AGENTS.md`：新增不变量 §4.15（密码学运算一律经 `cryptosvc` 转调）、两条禁止项与上游 issue 登记流程

### 破坏性变更

- 加密类 API 契约按实际实现收敛，**删除**以下尚未实现的承诺：
  - `symmetric-crypt`：`DES` / `3DES` 算法；AES 的 `OFB` / `CFB` 模式；AES 的 `zero` / `none` 填充
  - `asymmetric-crypt`：`ECIES` 算法；请求参数 `sm2Order` / `sm2Prefix04` / `sm2Format`（SM2 输出恒为 `C1C3C2` + `04` 前缀 + 裸格式）
  - `signature`：请求参数 `format` / `saltLength`；响应中的 `format` 字段（签名值改为 Base64 字符串）
  - `hmac-gen`：算法名改为 `HMAC-SM3` / `HMAC-SHA256` 等带前缀形式；移除文件输入
- `sm4-file-crypt` **未在本版交付**，端点返回 `501 NOT_IMPLEMENTED`，顺延至 v0.3.1
- `tongsuo-go` 能力缺口不做兜底：`DES` / `3DES` / `ECIES` / SM2 压缩点密文 / RSA-PSS 的 DER 形式**均不可用**，缺口登记在上游 issue 文档

### 依赖

- adds: `github.com/blue-cloud-net/tongsuo-go` v0.5.0

## [0.2.1] - 2026-10-28

前端骨架：让 v0.2.0 的后端链路在浏览器里跑起来。**不新增工具**。

### Added

- 前端工程脚手架（Vite 6 + Vue 3.5 + TypeScript 5.7 + Pinia + vue-router + vue-i18n + UnoCSS），Element Plus **按组件路径**引入
- 设计令牌与布局：`tokens.css` / `base.css` / `element-overrides.css`，`header` / `nav` / `main` / `footer` 地标齐备，「跳到主内容」链接可用
- API 客户端 `frontend/src/api/client.ts`：统一拆信封、`AbortController` 取消、超时归一化（`TIMEOUT`）、非 JSON 响应归为 `INVALID_RESPONSE`
- 目录 store 与发现页：首页 70 个入口 + 10 个分类、中英文名搜索、分类页、404
- 工具页骨架 `ToolPageLayout` / `ToolPanel` / `RunBar` / `ResultPanel` / `ErrorAlert` / `CopyButton` / `EncodingSelect`，四态齐备，`aria-live` 播报，`Ctrl` / `⌘ + Enter` 运行
- `base64` 与 `number-base` 两个工具页，含示例数据与「填入示例 → 运行 → 复制」全链路
- 双语词条：`zh-CN` 与 `en-US` 全覆盖（含 81 个错误码与警告码）
- 侧栏导航：`全部工具` / `最近使用` / `我的收藏` 三个特殊快捷入口，分类可折叠并展开工具列表
- 侧栏可折叠成图标轨道（仅图标 + 分类首两字），状态持久化；分类全部展开后菜单超出视口时**侧栏内部滚动且隐藏滚动条**
- 顶栏：主题切换改为图标按钮；语言与「跟随系统」收进设置菜单（Esc 关闭、点击外部关闭、关闭后焦点回到触发器）
- 工具页标题行新增收藏星标；新增 `/recent` 与 `/favorites` 两个页面
- 布局：侧栏为整列直通到底部（贴住视口底、带右边框），页脚只占内容列；页脚文案居中，含数据说明、版权、Apache-2.0 许可证、版本号（链接到发布记录）与 GitHub 仓库链接
- 站点常量集中在 `frontend/src/config/site.ts`（仓库地址、发布地址、版本号、版权年份），组件内不再写死
- 主题（`system` / `light` / `dark`）与语言切换，持久化并同步 `<html lang>`
- 体积预算脚本 `scripts/check-budget.mjs`：初始 JS < 250 KB（gzip）、单工具 chunk < 40 KB（gzip）
- `Makefile` 新增 `lint-web` / `test-web` / `build-web` / `check-budget`，并纳入 `verify`

### 依赖

- adds: `vue` 3.5.44
- adds: `vue-router` 4.6.4
- adds: `pinia` 3.0.4
- adds: `vue-i18n` 11.4.16
- adds: `element-plus` 2.14.7
- adds: `vite` 6.4.1
- adds: `vitest` 3.2.7
- adds: `unocss` 66.12.0

## [0.2.0] - 2026-10-28

后端骨架与基础设施。**不含任何前端产物**，前端骨架见 `v0.2.1`。

### Added

- 仓库拆分为两个独立构建单元：`backend/`（独立 Go 模块 `github.com/blue-cloud-net/Toolbit/backend`）与 `frontend/`（pnpm 项目）
- 工具目录内嵌：`backend/catalog/embed.go` 以 `//go:embed` 携带 `tools.yaml`，`internal/catalog` 负责解析与六类一致性校验
- 错误框架 `internal/toolerr`：67 个错误码 + 14 个警告码的注册表、统一 `Error` 类型与 `toolerr.From` 归一化
- 配置 `internal/config`：16 个 `TOOLBIT_*` 环境变量，非法值 fail-fast 且错误信息不含变量值
- 响应信封与中间件 `internal/httpx`：统一信封、requestId 透传与校验、安全响应头、访问日志、Body 限制、跨站校验、panic 恢复
- HTTP 分发器 `internal/api`：按 10 步顺序校验的统一端点、目录端点（`ETag` 协商）、限流（按 `rateClass` 五档分桶）、运维端点
- 统计落库 `internal/stats`：版本化迁移、异步批量写入、队列满丢弃计数、每日保留期清理
- 静态资源服务 `internal/webui`：缓存头、SPA 回退、路径穿越拦截、`/api/*` JSON 404
- 工具实现：`base64` 与 `number-base` 两个编码类工具，含契约测试数据
- 进程入口 `backend/cmd/toolbit/main.go`：配置装配、`SIGTERM` 优雅退出
- 工程门禁：`Makefile`（`make verify`）、`.golangci.yml`、`.editorconfig`

### Changed

- 重构路线图：[docs/roadmap.md](docs/roadmap.md) 拆分为 19 节，版本序列新增 `v0.2.1`（前端骨架）与 `v0.9.0`（统计体系），原「质量与硬化」顺延为 `v0.10.0+`
- `v0.2.0` 收敛为纯后端骨架（17 个 Issue 级任务），前端工作整体移入 `v0.2.1`
- 新增 `v0.9.0` 统计体系章节（9 个 Issue 级任务），涵盖访客标识、日聚合、公开统计端点与统计页

### 破坏性变更

- 目录布局：后端代码由仓库根移入 `backend/`，`catalog/tools.yaml` 移入 `backend/catalog/`，前端目录由 `web/` 改为 `frontend/`
- Go module 路径改为 `github.com/blue-cloud-net/Toolbit/backend`
- 68 个尚未实现的工具返回 `501 NOT_IMPLEMENTED`（catalog 仍返回全部 70 条）

### 依赖

- adds: `github.com/gin-gonic/gin` v1.12.0
- adds: `gopkg.in/yaml.v3` v3.0.1
- adds: `modernc.org/sqlite` v1.39.0
- adds: `golang.org/x/time` v0.12.0

## [0.1.0] - 2026-10-27

首个开发版本，仅包含工程地基与文档体系，工具尚未上线。

### Added

- 仓库骨架：`backend/cmd/toolbit`、`backend/internal/{api,catalog,codecsvc,cryptosvc,httpx,netguard,stats,toolerr,tools,webui}`、`frontend/`、`deploy/`、`catalog/`
- 文档体系：[README](README.md)、[AGENTS](AGENTS.md)、[架构总览](docs/architecture.md)、[开发规范](docs/standards/00-intro.md)、[API 契约](docs/api/README.md)、[工具规格](docs/tools.md)、[路线图](docs/roadmap.md)
- 工具目录 `backend/catalog/tools.yaml`：登记 70 个工具的分类、传输方式、限流档位与限额
- CI 工作流骨架：`ci`（lint / test）、`builder`（构建 Tongsuo 基础镜像）、`docker`（多架构镜像推 GHCR）
- 文档校验：Markdown lint、相对链接检查、catalog / 错误码 / 环境变量覆盖率检查

[未发布]: https://github.com/blue-cloud-net/Toolbit/compare/v0.3.1...HEAD
[0.3.1]: https://github.com/blue-cloud-net/Toolbit/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/blue-cloud-net/Toolbit/compare/v0.2.1...v0.3.0
[0.2.1]: https://github.com/blue-cloud-net/Toolbit/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/blue-cloud-net/Toolbit/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/blue-cloud-net/Toolbit/releases/tag/v0.1.0
