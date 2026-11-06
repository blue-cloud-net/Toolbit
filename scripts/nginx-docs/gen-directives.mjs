#!/usr/bin/env node
/**
 * 生成 Nginx 指令目录快照。
 *
 * 从 nginx.org 官方文档抓取**事实字段**并输出 JSON 快照：
 *   - 指令名、所属模块、语法、默认值、可用上下文
 *   - 首次引入 / 废弃 / 移除版本
 *   - 是否属于商业订阅（NGINX Plus）
 *   - 模块的构建开关与官方文档链接
 *
 * 只提取事实元数据：**不**复制官方文档正文与示例，
 * 简述文字来自 `summaries.json`（人工维护的自有文案）。
 *
 * 用法：node scripts/nginx-docs/gen-directives.mjs [--out <路径>] [--offline <缓存目录>]
 */

import { mkdir, readFile, writeFile } from "node:fs/promises";
import path from "node:path";
import process from "node:process";

const ROOT = path.resolve(import.meta.dirname, "..", "..");
const DEFAULT_OUT = path.join(
  ROOT,
  "backend",
  "internal",
  "tools",
  "system",
  "nginx",
  "data",
  "directives.json",
);
const SUMMARIES = path.join(import.meta.dirname, "summaries.json");

const DOCS_BASE = "https://nginx.org/en/docs/";
const USER_AGENT = "toolbit-directive-snapshot/1.0 (+docs metadata only)";

/** 商业订阅模块的说明短语。 */
const COMMERCIAL_MODULE_MARK = "This module is available as part of our";

/**
 * 仅出现在商业订阅中的模块。
 *
 * 官方文档里这些模块并不都用同一句话声明，
 * 因此这里维护一份显式清单（事实性标注，不含任何实现信息）。
 */
const PLUS_ONLY_MODULES = new Set([
  "ngx_google_perftools_module",
  "ngx_mgmt_module",
  "ngx_http_acme_module",
  "ngx_http_api_module",
  "ngx_http_status_module",
  "ngx_otel_module",
]);

/** 拉取页面，带重试与超时。 */
async function fetchText(url) {
  let lastError;
  for (let attempt = 0; attempt < 3; attempt += 1) {
    try {
      const controller = new AbortController();
      const timer = setTimeout(() => controller.abort(), 30000);
      const res = await fetch(url, {
        headers: { "user-agent": USER_AGENT },
        signal: controller.signal,
      });
      clearTimeout(timer);
      if (!res.ok) {
        throw new Error(`HTTP ${res.status}`);
      }
      return await res.text();
    } catch (err) {
      lastError = err;
      await new Promise((resolve) => setTimeout(resolve, 800 * (attempt + 1)));
    }
  }
  throw new Error(`抓取失败 ${url}: ${lastError?.message ?? "unknown"}`);
}

/** 去标签并折叠空白。 */
function textOf(html) {
  return html
    .replace(/<br\s*\/?>/gi, " ")
    .replace(/<[^>]+>/g, "")
    .replace(/&nbsp;/g, " ")
    .replace(/&lt;/g, "<")
    .replace(/&gt;/g, ">")
    .replace(/&amp;/g, "&")
    .replace(/&quot;/g, '"')
    .replace(/\s+/g, " ")
    .trim();
}

/**
 * 额外抓取的模块页。
 *
 * 全局指令索引只覆盖 http / stream / mail 三套模块，
 * **核心模块**（main 上下文的指令）与少数独立模块不在索引里，
 * 因此显式补上。
 */
const EXTRA_PAGES = [
  { page: "ngx_core_module.html", module: "ngx_core_module" },
];

/** 解析全局指令索引：指令名 → 模块页与模块名。 */
function parseIndex(html) {
  // 条目有两种写法：带模块名后缀的 `...>allow</a> (ngx_http_access_module)`，
  // 与不带后缀的 `...>index</a>`（模块名要另行推断）。
  const withModule = /<a href="([^"#]+\.html)#([A-Za-z0-9_]+)">([A-Za-z0-9_]+)<\/a>\s*\(([A-Za-z0-9_]+)\)/g;
  const withoutModule = /<a href="([^"#]+\.html)#([A-Za-z0-9_]+)">([A-Za-z0-9_]+)<\/a>/g;

  const byPage = new Map();
  const moduleOfPage = new Map();

  const add = (page, anchor, module) => {
    if (!byPage.has(page)) byPage.set(page, { module, anchors: new Set() });
    byPage.get(page).anchors.add(anchor);
    if (module) moduleOfPage.set(page, module);
  };

  let m;
  while ((m = withModule.exec(html)) !== null) {
    add(m[1], m[2], m[4]);
  }
  while ((m = withoutModule.exec(html)) !== null) {
    const [, page, anchor] = m;
    if (!byPage.has(page)) {
      // 模块名从文件名推断：`http/ngx_http_index_module.html` → `ngx_http_index_module`。
      const file = page.split("/").pop() ?? page;
      const inferred = file.replace(/\.html$/, "");
      add(page, anchor, moduleOfPage.get(page) ?? inferred);
    }
  }

  // 修正：同一页可能先被无后缀条目登记，此时模块名是推断值；
  // 若后续出现带后缀条目，以官方标注为准。
  for (const [page, info] of byPage.entries()) {
    const official = moduleOfPage.get(page);
    if (official) {
      info.module = official;
    }
  }

  return byPage;
}

/** 取出指令定义块的三个字段与说明文本。 */
function parseDirectiveBlocks(html) {
  const re = /<a name="([A-Za-z0-9_]+)"><\/a><div class="directive">([\s\S]*?)(?=<a name="[A-Za-z0-9_]+"><\/a><div class="directive">|$)/g;
  const out = [];
  let m;
  while ((m = re.exec(html)) !== null) {
    const [, anchor, body] = m;
    const cell = (label) => {
      const c = new RegExp(
        `<th>\\s*${label}:\\s*</th>\\s*<td>([\\s\\S]*?)</td>`,
        "i",
      ).exec(body);
      return c ? c[1] : "";
    };

    const syntaxHtml = cell("Syntax");
    const defaultHtml = cell("Default");
    const contextHtml = cell("Context");

    // 上下文单元格里的每一项都是 <code>…</code>。
    const contexts = [];
    const ctxRe = /<code>([\s\S]*?)<\/code>/gi;
    let cm;
    while ((cm = ctxRe.exec(contextHtml)) !== null) {
      const value = textOf(cm[1]);
      if (value) contexts.push(value);
    }

    // 语法单元格首项是 <strong>指令名</strong>，其后是参数。
    const syntax = textOf(syntaxHtml);
    const defaultValue = textOf(defaultHtml);

    out.push({
      anchor,
      syntax,
      default: defaultValue,
      contexts,
      body,
    });
  }
  return out;
}

/** 从说明文本里取版本事实与商业版参数标记。 */
function parseFacts(body) {
  const plain = textOf(body);
  const since = /This directive appeared in version (\d+\.\d+\.\d+)/.exec(plain);
  const deprecated = /This directive was made obsolete in version (\d+\.\d+\.\d+)/.exec(plain);
  const removed = /was removed in version (\d+\.\d+\.\d+)/.exec(plain);
  // 官方文档里的 “available as part of our commercial subscription”
  // 通常出现在**某个参数**的说明中，而不是整条指令不可用，
  // 因此这里记录为「含商业版参数」，不表示指令本身仅商业版提供。
  const commercial = /commercial subscription/.test(plain);
  return {
    since: since ? since[1] : "",
    deprecated: deprecated ? deprecated[1] : "",
    removed: removed ? removed[1] : "",
    commercial,
  };
}

/** 解析模块页元信息。 */
function parseModuleMeta(html) {
  const h1 = /<h1[^>]*>([\s\S]*?)<\/h1>/i.exec(html);
  const title = h1 ? textOf(h1[1]) : "";
  const head = textOf(html.slice(0, html.indexOf("<a name=\"directives\"")));

  let buildOption = "";
  let defaultBuild = true;
  const buildMatch = /not built by default[^.]*?(--with-[a-z0-9_]+|--without-[a-z0-9_]+)/i.exec(head);
  if (buildMatch) {
    buildOption = buildMatch[1];
    defaultBuild = false;
  } else {
    const withMatch = /enabled with the\s*(--with-[a-z0-9_]+)/i.exec(head);
    if (withMatch) buildOption = withMatch[1];
  }

  const commercial = head.includes(COMMERCIAL_MODULE_MARK);

  return { title, buildOption, defaultBuild, commercial };
}
/** 比较版本号。 */
function cmpVersion(a, b) {
  const pa = a.split(".").map(Number);
  const pb = b.split(".").map(Number);
  for (let i = 0; i < 3; i += 1) {
    const va = pa[i] ?? 0;
    const vb = pb[i] ?? 0;
    if (va !== vb) return va - vb;
  }
  return 0;
}

/** 主流程。 */
async function main() {
  const args = process.argv.slice(2);
  let out = DEFAULT_OUT;
  for (let i = 0; i < args.length; i += 1) {
    if (args[i] === "--out" && args[i + 1]) {
      out = path.resolve(args[i + 1]);
      i += 1;
    }
  }

  console.log("→ 抓取全局指令索引");
  const indexHtml = await fetchText(new URL("dirindex.html", DOCS_BASE).href);
  const byPage = parseIndex(indexHtml);
  console.log(`  模块页 ${byPage.size} 个`);

  // 核心模块不在索引里，显式补上。
  for (const extra of EXTRA_PAGES) {
    if (!byPage.has(extra.page)) {
      byPage.set(extra.page, { module: extra.module, anchors: new Set() });
    }
  }

  const summaries = JSON.parse(await readFile(SUMMARIES, "utf8"));

  const modules = [];
  const directives = [];

  for (const [page, info] of [...byPage.entries()].sort((a, b) =>
    a[0].localeCompare(b[0]),
  )) {
    const url = new URL(page, DOCS_BASE).href;
    let html;
    try {
      html = await fetchText(url);
    } catch (err) {
      console.warn(`  跳过 ${page}：${err.message}`);
      continue;
    }

    const meta = parseModuleMeta(html);
    modules.push({
      name: info.module,
      title: meta.title || info.module,
      docUrl: url,
      buildOption: meta.buildOption,
      defaultBuild: meta.defaultBuild,
      commercial: meta.commercial || PLUS_ONLY_MODULES.has(info.module),
    });

    for (const block of parseDirectiveBlocks(html)) {
      const facts = parseFacts(block.body);
      const summary = summaries[block.anchor] ?? {};
      directives.push({
        id: `${info.module}::${block.anchor}`,
        name: block.anchor,
        module: info.module,
        syntax: block.syntax,
        default: block.default,
        contexts: block.contexts,
        block: /\{\s*\.\.\.\s*\}/.test(block.syntax),
        since: facts.since,
        deprecated: facts.deprecated,
        removed: facts.removed,
        commercialParams: facts.commercial,
        docUrl: `${url}#${block.anchor}`,
        summaryZh: summary.zh ?? "",
        summaryEn: summary.en ?? "",
      });
    }
    console.log(`  ${info.module}：累计指令 ${directives.length}`);
  }

  directives.sort((a, b) => a.name.localeCompare(b.name) || a.module.localeCompare(b.module));
  modules.sort((a, b) => a.name.localeCompare(b.name));

  // 目标版本：主线版与稳定版（来自官方下载页）。
  console.log("→ 抓取版本列表");
  const downloadHtml = await fetchText(new URL("../download.html", DOCS_BASE).href);
  const found = [...downloadHtml.matchAll(/nginx-(\d+\.\d+\.\d+)/g)].map((m) => m[1]);
  const versions = [...new Set(found)].sort(cmpVersion);
  const mainline = versions.at(-1) ?? "";
  const stable = versions.at(-2) ?? mainline;

  const snapshot = {
    docsVersion: mainline,
    stableVersion: stable,
    mainlineVersion: mainline,
    source: DOCS_BASE,
    modules,
    directives,
  };

  await mkdir(path.dirname(out), { recursive: true });
  await writeFile(out, `${JSON.stringify(snapshot, null, 2)}\n`, "utf8");
  console.log(`✓ 已写入 ${path.relative(ROOT, out)}：${directives.length} 条指令，${modules.length} 个模块`);
  console.log(`  稳定版 ${stable} ／ 主线版 ${mainline}`);
}

main().catch((err) => {
  console.error("生成指令快照失败：", err);
  process.exit(1);
});
