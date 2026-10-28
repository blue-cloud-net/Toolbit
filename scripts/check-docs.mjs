#!/usr/bin/env node
/**
 * Toolbit 文档一致性检查
 *
 * 校验项：
 *   1. catalog/tools.yaml 中的每个 slug 都在 docs/api/tools/ 与 docs/tools.md 中有记录
 *   2. docs/api/errors.md 覆盖 internal/toolerr/registry.go 中的全部错误码
 *   3. docs/deploy.md 覆盖 internal/config 中的全部环境变量
 *   4. 全部 Markdown 文档都带状态行
 *   5. 禁用词检查
 *
 * 退出码：0 = 通过；1 = 存在问题（输出差异清单）
 *
 * 用法：node scripts/check-docs.mjs
 */

import { readFile, readdir, stat } from "node:fs/promises";
import { existsSync } from "node:fs";
import path from "node:path";
import process from "node:process";

// ── 配置 ────────────────────────────────────────────────────────────────

const ROOT = path.resolve(import.meta.dirname, "..");
const CATALOG = path.join(ROOT, "backend", "catalog", "tools.yaml");
const API_TOOLS_DIR = path.join(ROOT, "docs", "api", "tools");
const ERRORS_DOC = path.join(ROOT, "docs", "api", "errors.md");
const DEPLOY_DOC = path.join(ROOT, "docs", "deploy.md");
const TOOLS_DOC = path.join(ROOT, "docs", "tools.md");
const REGISTRY_DIR = path.join(ROOT, "backend", "internal", "toolerr");
const CONFIG_GO = path.join(ROOT, "backend", "internal", "config", "config.go");
const DOCS_DIR = path.join(ROOT, "docs");
const README = path.join(ROOT, "README.md");

/** 禁用词：不得出现在任何文档中 */
const FORBIDDEN_TERMS = [
  "ittools",
  "IT-Tools",
  "toolbox365",
  "toolbit365",
];

/** 状态行格式 */
const STATUS_LINE = /^> 状态：.+ ｜ 适用版本：.+ ｜ 最后更新：\d{4}-\d{2}-\d{2}$/m;

// ── 工具函数 ────────────────────────────────────────────────────────────

/** 递归收集所有 Markdown 文件 */
async function collectMarkdown(dir) {
  const out = [];
  if (!existsSync(dir)) return out;
  const entries = await readdir(dir, { withFileTypes: true });
  for (const entry of entries) {
    const full = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      if (["node_modules", "dist", ".git"].includes(entry.name)) continue;
      out.push(...(await collectMarkdown(full)));
    } else if (entry.isFile() && entry.name.endsWith(".md")) {
      out.push(full);
    }
  }
  return out;
}

/** 极简 YAML 列表项解析：提取 `- slug: xxx` 与 `  key: value` */
function parseCatalog(text) {
  const tools = [];
  const categories = [];
  let inTools = false;
  let inCategories = false;
  let current = null;
  let currentCategory = null;

  for (const raw of text.split("\n")) {
    if (/^tools:\s*$/.test(raw)) {
      inTools = true;
      inCategories = false;
      continue;
    }
    if (/^categories:\s*$/.test(raw)) {
      inCategories = true;
      inTools = false;
      continue;
    }
    if (/^\S/.test(raw) && raw.trim() !== "" && !raw.startsWith("#")) {
      // 新的顶层键，结束当前段落
      inTools = false;
      inCategories = false;
      current = null;
      currentCategory = null;
      continue;
    }

    if (inCategories) {
      const m = raw.match(/^\s*-\s+slug:\s*(\S+)/);
      if (m) {
        currentCategory = { slug: m[1] };
        categories.push(currentCategory);
        continue;
      }
      const cm = raw.match(/^\s+name:\s*(.+)$/);
      if (cm && currentCategory) currentCategory.name = cm[1].trim();
      continue;
    }

    if (inTools) {
      const m = raw.match(/^\s*-\s+slug:\s*(\S+)/);
      if (m) {
        current = { slug: m[1] };
        tools.push(current);
        continue;
      }
      const km = raw.match(/^\s+([A-Za-z][A-Za-z0-9]*):\s*(.+)$/);
      if (km && current) {
        current[km[1]] = km[2].trim().replace(/^["']|["']$/g, "");
      }
    }
  }
  return { tools, categories };
}

/** 从 Go 源码中提取错误码常量名 */
function parseErrorCodes(text) {
  const codes = new Set();
  // 先剥离行注释，避免注释中的示例代码被当成真实声明。
  const source = text
    .split("\n")
    .map((line) => line.replace(/\/\/.*$/, ""))
    .join("\n");
  const re = /Code([A-Z][A-Za-z0-9]*)\s*ErrorCode\s*=/g;
  let m;
  while ((m = re.exec(source)) !== null) {
    codes.add(m[1].toUpperCase());
  }
  // 兜底：匹配 "CODE": http.Status... 形式的注册表条目
  const re2 = /"([A-Z][A-Z0-9_]{3,})"\s*:\s*\{/g;
  while ((m = re2.exec(source)) !== null) {
    codes.add(m[1]);
  }
  return [...codes].sort();
}

/** 从 Go 源码中提取环境变量名 */
function parseEnvVars(text) {
  const vars = new Set();
  const re = /"TOOLBIT_[A-Z0-9_]+"/g;
  let m;
  while ((m = re.exec(text)) !== null) {
    vars.add(m[0].replace(/"/g, ""));
  }
  return [...vars].sort();
}

// ── 检查实现 ────────────────────────────────────────────────────────────

const problems = [];
const notes = [];

function fail(scope, message) {
  problems.push({ scope, message });
}

async function checkCatalogCoverage() {
  if (!existsSync(CATALOG)) {
    fail("catalog", "catalog/tools.yaml 不存在");
    return;
  }
  const { tools, categories } = parseCatalog(await readFile(CATALOG, "utf8"));
  if (tools.length === 0) {
    fail("catalog", "catalog/tools.yaml 未解析到任何工具");
    return;
  }
  notes.push(`catalog: ${tools.length} 个工具，${categories.length} 个分类`);

  // slug 唯一性
  const seen = new Set();
  for (const t of tools) {
    if (seen.has(t.slug)) fail("catalog", `slug 重复：${t.slug}`);
    seen.add(t.slug);
    if (!/^[a-z0-9]+(-[a-z0-9]+)*$/.test(t.slug)) {
      fail("catalog", `slug 不符合 kebab-case：${t.slug}`);
    }
    for (const required of ["category", "order", "name", "nameEn", "transport", "rateClass", "maxBodyBytes", "timeoutMs", "i18nKey"]) {
      if (t[required] === undefined) fail("catalog", `工具 ${t.slug} 缺少字段 ${required}`);
    }
  }

  // 分类引用有效性
  const catSlugs = new Set(categories.map((c) => c.slug));
  for (const t of tools) {
    if (!catSlugs.has(t.category)) {
      fail("catalog", `工具 ${t.slug} 引用了不存在的分类 ${t.category}`);
    }
  }

  // API 文档覆盖
  if (!existsSync(API_TOOLS_DIR)) {
    fail("api", "docs/api/tools/ 目录不存在");
  } else {
    const files = (await readdir(API_TOOLS_DIR)).filter((f) => f.endsWith(".md"));
    const apiText = (
      await Promise.all(files.map((f) => readFile(path.join(API_TOOLS_DIR, f), "utf8")))
    ).join("\n");
    for (const t of tools) {
      if (!apiText.includes(t.slug)) {
        fail("api", `工具 ${t.slug} 在 docs/api/tools/ 中没有出现`);
      }
    }
    notes.push(`api/tools: ${files.length} 个分类文件`);
  }

  // 产品文档覆盖
  if (!existsSync(TOOLS_DOC)) {
    fail("docs", "docs/tools.md 不存在");
  } else {
    const text = await readFile(TOOLS_DOC, "utf8");
    for (const t of tools) {
      if (!text.includes(t.slug)) {
        fail("docs", `工具 ${t.slug} 在 docs/tools.md 中没有出现`);
      }
    }
  }
}

async function checkErrorCodes() {
  if (!existsSync(REGISTRY_DIR)) {
    notes.push("errors: backend/internal/toolerr 尚不存在，跳过（代码未实现）");
    return;
  }
  const files = (await readdir(REGISTRY_DIR)).filter((f) => f.endsWith(".go"));
  if (files.length === 0) {
    notes.push("errors: backend/internal/toolerr 下没有 .go 文件，跳过");
    return;
  }
  const source = (
    await Promise.all(files.map((f) => readFile(path.join(REGISTRY_DIR, f), "utf8")))
  ).join("\n");
  const codes = parseErrorCodes(source);
  if (codes.length === 0) {
    notes.push("errors: 未从 toolerr 包解析到错误码，跳过");
    return;
  }
  const doc = existsSync(ERRORS_DOC) ? await readFile(ERRORS_DOC, "utf8") : "";
  if (!doc) {
    fail("errors", "docs/api/errors.md 不存在");
    return;
  }
  for (const code of codes) {
    if (!doc.includes(code)) {
      fail("errors", `错误码 ${code} 未在 docs/api/errors.md 中记录`);
    }
  }
  notes.push(`errors: ${codes.length} 个错误码已校验（来自 ${files.length} 个源文件）`);
}

async function checkConfigVars() {
  if (!existsSync(CONFIG_GO)) {
    notes.push("config: backend/internal/config/config.go 尚不存在，跳过（代码未实现）");
    return;
  }
  const vars = parseEnvVars(await readFile(CONFIG_GO, "utf8"));
  if (vars.length === 0) {
    notes.push("config: 未从 config.go 解析到环境变量，跳过");
    return;
  }
  const doc = existsSync(DEPLOY_DOC) ? await readFile(DEPLOY_DOC, "utf8") : "";
  for (const v of vars) {
    if (!doc.includes(v)) {
      fail("config", `环境变量 ${v} 未在 docs/deploy.md 中记录`);
    }
  }
  notes.push(`config: ${vars.length} 个环境变量已校验`);
}

async function checkStatusLines() {
  const files = await collectMarkdown(DOCS_DIR);
  files.push(README);
  for (const file of files) {
    const rel = path.relative(ROOT, file);
    if (rel.startsWith("docs/_templates")) continue;
    const text = await readFile(file, "utf8");
    if (!STATUS_LINE.test(text)) {
      fail("status", `${rel} 缺少或不合规的状态行`);
    }
  }
  notes.push(`status: ${files.length} 个文档已校验状态行`);
}

async function checkForbiddenTerms() {
  const files = await collectMarkdown(DOCS_DIR);
  files.push(README);
  const extra = [path.join(ROOT, "AGENTS.md"), path.join(ROOT, "CHANGELOG.md")];
  for (const f of extra) if (existsSync(f)) files.push(f);

  for (const file of files) {
    const rel = path.relative(ROOT, file);
    const text = await readFile(file, "utf8");
    for (const term of FORBIDDEN_TERMS) {
      if (text.toLowerCase().includes(term.toLowerCase())) {
        fail("forbidden", `${rel} 含有禁用词「${term}」`);
      }
    }
  }
  notes.push(`forbidden: ${files.length} 个文件已检查禁用词`);
}

// ── 主流程 ──────────────────────────────────────────────────────────────

async function main() {
  await checkCatalogCoverage();
  await checkErrorCodes();
  await checkConfigVars();
  await checkStatusLines();
  await checkForbiddenTerms();

  console.log("── 文档检查 ──────────────────────────────");
  for (const n of notes) console.log(`  · ${n}`);

  if (problems.length === 0) {
    console.log("\n✅ 文档检查通过");
    process.exit(0);
  }

  console.log(`\n❌ 发现 ${problems.length} 个问题：\n`);
  const byScope = new Map();
  for (const p of problems) {
    if (!byScope.has(p.scope)) byScope.set(p.scope, []);
    byScope.get(p.scope).push(p.message);
  }
  for (const [scope, list] of byScope) {
    console.log(`  [${scope}] ${list.length} 项`);
    for (const msg of list) console.log(`    - ${msg}`);
  }
  process.exit(1);
}

main().catch((err) => {
  console.error("检查脚本异常：", err);
  process.exit(1);
});
