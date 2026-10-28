# Git 与协作规范

> 状态：Draft ｜ 适用版本：v0.1.0 ｜ 最后更新：2026-10-27

---

## 1. 分支模型

| 分支 | 用途 | 保护 |
| --- | --- | --- |
| `main` | 稳定分支 | 禁止直推；需 PR + CI 全绿 |
| `feat/<scope>` | 功能分支 | — |
| `fix/<scope>` | 修复分支 | — |
| `docs/<scope>` | 文档分支 | — |
| `refactor/<scope>` | 重构分支 | — |
| `chore/<scope>` | 工程杂项 | — |

**必须**：`main` 保持**随时可发布**。

**禁止**在 `main` 上直接提交。

---

## 2. 提交信息

采用 Conventional Commits：

```text
<type>(<scope>): <subject>

[body]

[footer]
```

### 2.1 type

| type | 用途 |
| --- | --- |
| `feat` | 新功能（含新工具） |
| `fix` | 缺陷修复 |
| `refactor` | 不改变行为的重构 |
| `perf` | 性能优化 |
| `docs` | 文档 |
| `test` | 测试 |
| `build` | 构建 / 依赖 |
| `ci` | CI 配置 |
| `chore` | 其它杂项 |
| `revert` | 回滚 |

### 2.2 scope

`scope` **必须**用目录或领域名：

`api` `cryptosvc` `catalog` `tools` `web` `i18n` `docker` `ci` `docs` `deps`

### 2.3 subject

- **必须**使用中文或英文祈使句，不超过 72 字符；
- **必须**以动词开头，不用「修改了」「更新」；
- **禁止**句末加句号。

**示例**

```text
feat(cryptosvc): 新增 SM2 密文格式互转与标准向量测试
fix(api): 修正超过 64 张证书时的链验证边界
docs(standards): 补充加密容器 TBFC 的 nonce 派生规则
```

### 2.4 body

**必须**在 body 中说明**为什么**，而不是「改了什么」（diff 已经说了）。

**必须**在 body 中列出破坏性变更与迁移步骤。

### 2.5 footer

**必须**在涉及问题时引用 issue：

```text
Closes #123
Refs #45
```

**禁止**在 footer 中放无关注释。

---

## 3. 提交粒度

**必须**

- 一个提交**只做一件事**；
- 一个提交**只涉及一个变更点**；
- 格式化、依赖升级、代码修改**必须**分开提交；
- 文档**应当**与对应代码同提交（除非文档滞后有明确理由）。

**禁止**

- 把「格式化整个文件」与「功能改动」混在一个提交；
- 在功能分支夹带依赖升级；
- 提交编译不过的中间状态（**必须**每个提交都能通过 CI）。

---

## 4. 敏感信息

**禁止**提交：

- 私钥、证书私钥、`.pem` / `.key`（除测试夹具白名单）；
- `.env`；
- 数据库文件；
- 真实用户数据；
- 内部构建产物。

`.gitignore` **必须**覆盖：`*.pem` `*.key` `.env` `*.db` `/data/` `node_modules/` `bin/` `frontend/dist/`。

**必须**在提交前检查：

```bash
git diff --cached --name-only
git diff --cached | grep -iE '(BEGIN .*PRIVATE KEY|api[_-]?key|secret\s*=)'
```

---

## 5. Pull Request

**必须**

- PR 描述包含：变更目的、影响面、验证方式、关联 issue；
- 勾选对应的 checklist（测试、文档、错误码、i18n、依赖许可证）；
- 至少 1 个 review 通过；
- CI 全绿（含 `-race`、lint、docs 检查）；
- 合并方式：**Squash merge**（保持 `main` 历史线性）。

**禁止**

- 绕过 CI 合并；
- 混合多个不相关变更到一个 PR。

### 5.1 变更类型对应的检查项

| 变更 | **必须**检查 |
| --- | --- |
| 新增工具 | catalog、API 文档、tools 文档、i18n、契约测试、前端页面 |
| 新增错误码 | registry、errors 文档、i18n、测试 |
| 新增配置 | config、deploy 文档 |
| 依赖升级 | changelog、许可证、漏洞扫描 |
| 容器相关 | 镜像扫描、非 root、digest 固定 |

---

## 6. 提交前自检

```bash
gofmt -l .                       # Go 格式
go vet ./...                     # 静态检查
go test ./... -race              # 测试
pnpm lint && pnpm test           # 前端
pnpm build                       # 前端构建
scripts/check-docs.mjs           # 文档一致性
git diff --cached --stat         # 变更面
```

**必须**在提交前至少执行 `go test ./... -race` 与 `pnpm build`。

---

## 7. 分支与标签

- 分支名**必须**小写，短横线分隔；
- 标签**必须**用语义化版本（`v0.1.0`），**禁止**用日期或随机串；
- 发布标签**必须**指向 `main` 的某个 commit，且该 commit 的 CI 全绿。

**禁止**给已发布的标签重新打（必须用新版本号）。

---

## 8. 版本号

语义化版本 `MAJOR.MINOR.PATCH`：

| 变更 | 版本位 |
| --- | --- |
| 破坏性 API 变化 | MAJOR |
| 新功能 / 新工具 | MINOR |
| 缺陷修复 / 文档 | PATCH |

**必须**在 [CHANGELOG.md](../../CHANGELOG.md) 记录每个版本的变更。

**禁止**在已有版本上追加内容（已发布版本不可变）。

---

## 9. 冲突处理

**必须**

- 合并前 rebase 到最新 `main`；
- 冲突**必须**手工解决，**禁止**无脑 `ours` / `theirs`；
- 解决后**必须**重跑完整测试；
- 冲突解决**应当**是**独立**提交。

---

## 10. 禁止清单

- ❌ 直接向 `main` 推送。
- ❌ 一个提交混合多类变更。
- ❌ 提交 `.env`、私钥、数据库文件。
- ❌ 提交编译不过的中间状态。
- ❌ 使用无结构的提交信息。
- ❌ 绕过 CI 合并。
- ❌ 修改已发布标签。
