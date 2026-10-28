#!/usr/bin/env node
/**
 * 前端体积预算检查。
 *
 * 预算口径（docs/standards/01-frontend.md §11）：
 *   - 初始 JS（gzip）< 250 KB
 *   - 单个工具 chunk（gzip）< 40 KB
 *
 * 「初始 JS」= index.html 直接引用的脚本 + 它静态预加载（modulepreload）的 chunk。
 * 路由级懒加载的页面与工具 chunk **不计入**初始体积。
 *
 * 退出码：0 = 通过；1 = 超预算或产物缺失。
 *
 * 用法：node scripts/check-budget.mjs [distDir]
 */

import { readFile, readdir } from "node:fs/promises";
import { existsSync } from "node:fs";
import path from "node:path";
import process from "node:process";
import { gzipSync } from "node:zlib";

// ── 预算 ────────────────────────────────────────────────────────────────

/** 初始 JS 上限（字节，gzip 后）。 */
const INITIAL_JS_BUDGET = 250 * 1024;

/** 单个工具 chunk 上限（字节，gzip 后）。 */
const TOOL_CHUNK_BUDGET = 40 * 1024;

/** 工具 chunk 的命名约定（见 vite.config.ts 的 manualChunks）。 */
const TOOL_CHUNK_RE = /^tool-[a-z0-9-]+-[\w-]+\.js$/;

const ROOT = path.resolve(import.meta.dirname, "..");
const DIST = path.resolve(ROOT, process.argv[2] ?? path.join("frontend", "dist"));

// ── 工具函数 ────────────────────────────────────────────────────────────

/** 人类可读的 KB。 */
function kb(bytes) {
  return `${(bytes / 1024).toFixed(1)} KB`;
}

/** 计算文件 gzip 后体积。 */
async function gzipSize(file) {
  const raw = await readFile(file);
  return gzipSync(raw, { level: 9 }).length;
}

/** 从 index.html 中提取直接加载与预加载的 JS 资源。 */
function initialAssets(html) {
  const out = new Set();
  const attrRe = /(?:src|href)="([^"]+\.js)"/g;
  let m;
  while ((m = attrRe.exec(html)) !== null) {
    out.add(path.basename(m[1]));
  }
  return [...out];
}

// ── 主流程 ──────────────────────────────────────────────────────────────

async function main() {
  const indexPath = path.join(DIST, "index.html");
  if (!existsSync(indexPath)) {
    console.error(`❌ 找不到构建产物：${path.relative(ROOT, indexPath)}`);
    console.error("   请先执行 `pnpm build`。");
    process.exit(1);
  }

  const assetsDir = path.join(DIST, "assets");
  const files = (await readdir(assetsDir)).filter((f) => f.endsWith(".js"));

  const html = await readFile(indexPath, "utf8");
  const initial = initialAssets(html).filter((f) => files.includes(f));

  let initialTotal = 0;
  const initialRows = [];
  for (const file of initial) {
    const size = await gzipSize(path.join(assetsDir, file));
    initialTotal += size;
    initialRows.push({ file, size });
  }

  const toolRows = [];
  for (const file of files) {
    if (!TOOL_CHUNK_RE.test(file)) continue;
    toolRows.push({ file, size: await gzipSize(path.join(assetsDir, file)) });
  }

  console.log("── 前端体积 ──────────────────────────────");
  console.log(`  初始 JS（${initial.length} 个 chunk）`);
  for (const row of initialRows) {
    console.log(`    ${row.file}  ${kb(row.size)}`);
  }
  console.log(`    合计 ${kb(initialTotal)} / 预算 ${kb(INITIAL_JS_BUDGET)}`);

  if (toolRows.length > 0) {
    console.log("  工具 chunk");
    for (const row of toolRows) {
      console.log(`    ${row.file}  ${kb(row.size)} / 预算 ${kb(TOOL_CHUNK_BUDGET)}`);
    }
  } else {
    console.log("  工具 chunk：未拆分出独立 chunk（检查 manualChunks 规则）");
  }

  const problems = [];
  if (initialTotal > INITIAL_JS_BUDGET) {
    problems.push(`初始 JS 超出预算：${kb(initialTotal)} > ${kb(INITIAL_JS_BUDGET)}`);
  }
  for (const row of toolRows) {
    if (row.size > TOOL_CHUNK_BUDGET) {
      problems.push(`工具 chunk 超出预算：${row.file} ${kb(row.size)} > ${kb(TOOL_CHUNK_BUDGET)}`);
    }
  }

  if (problems.length === 0) {
    console.log("\n✅ 体积预算通过");
    process.exit(0);
  }

  console.log(`\n❌ 发现 ${problems.length} 个体积问题：\n`);
  for (const p of problems) console.log(`  - ${p}`);
  process.exit(1);
}

main().catch((err) => {
  console.error("预算检查脚本异常：", err);
  process.exit(1);
});
