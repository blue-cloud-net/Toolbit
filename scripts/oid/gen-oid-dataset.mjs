#!/usr/bin/env node
/**
 * 生成 OID 数据集。
 *
 * 输入（三类）：
 *   1. IANA 注册表 CSV（**官方英文**来源，按需联网拉取，缓存在 .cache/）；
 *   2. `scripts/oid/curated/*.csv` —— IANA 未覆盖分支（ITU/ISO 根、X.500 属性、
 *      X.509 扩展、PKCS、NIST 算法、国密、Thawte）的完整条目，含英文与中文；
 *   3. `scripts/oid/zh/*.csv` —— IANA 已覆盖分支的**中文覆盖表**，只写中文两列。
 *
 * 输出：
 *   backend/internal/tools/pki/dataset/oid_nodes.csv —— OID、名称、说明、种类、状态、来源
 *   backend/internal/tools/pki/dataset/oid_zh.csv    —— OID、中文名、中文说明
 *
 * 设计取舍：
 *   - 英文名与来源**一律取官方**：IANA 注册表优先，其次人工维护的表；
 *   - 中文**必须全覆盖**：缺中文即报错，杜绝「半中半英」；
 *   - IANA 的 `Decimal` 布局是**弧**，`OID Value` 布局是**完整 OID**，两者都要支持；
 *   - `10-20`、`Reserved` 这类区间或非弧取值**跳过**（它们不是可定位的节点）。
 *
 * 用法：
 *   node scripts/oid/gen-oid-dataset.mjs            # 生成并写出
 *   node scripts/oid/gen-oid-dataset.mjs --check    # 只校验，不写出
 *   node scripts/oid/gen-oid-dataset.mjs --offline  # 用 .cache/ 的缓存，不联网
 *
 * 退出码：0 = 通过；1 = 校验失败。
 */
import { mkdir, readFile, readdir, writeFile } from "node:fs/promises";
import path from "node:path";

const ROOT = path.resolve(import.meta.dirname, "..", "..");
const CURATED_DIR = path.join(import.meta.dirname, "curated");
const ZH_DIR = path.join(import.meta.dirname, "zh");
const CACHE_DIR = path.join(import.meta.dirname, ".cache");
const OUT_DIR = path.join(ROOT, "backend", "internal", "tools", "pki", "dataset");

/** IANA 注册表清单。 */
const IANA_BASE = "https://www.iana.org/assignments/smi-numbers";
const IANA_REGISTRIES = [
  { file: "smi-numbers-1.3.6.1.5.5.7", oid: "1.3.6.1.5.5.7" },
  { file: "smi-numbers-1.3.6.1.5.5.7.0", oid: "1.3.6.1.5.5.7.0" },
  { file: "smi-numbers-1.3.6.1.5.5.7.1", oid: "1.3.6.1.5.5.7.1" },
  { file: "smi-numbers-1.3.6.1.5.5.7.2", oid: "1.3.6.1.5.5.7.2" },
  { file: "smi-numbers-1.3.6.1.5.5.7.3", oid: "1.3.6.1.5.5.7.3" },
  { file: "smi-numbers-1.3.6.1.5.5.7.4", oid: "1.3.6.1.5.5.7.4" },
  { file: "smi-numbers-1.3.6.1.5.5.7.5", oid: "1.3.6.1.5.5.7.5" },
  { file: "smi-numbers-1.3.6.1.5.5.7.6", oid: "1.3.6.1.5.5.7.6" },
  { file: "smi-numbers-1.3.6.1.5.5.7.7", oid: "1.3.6.1.5.5.7.7" },
  { file: "smi-numbers-1.3.6.1.5.5.7.8", oid: "1.3.6.1.5.5.7.8" },
  { file: "smi-numbers-1.3.6.1.5.5.7.9", oid: "1.3.6.1.5.5.7.9" },
  { file: "smi-numbers-1.3.6.1.5.5.7.10", oid: "1.3.6.1.5.5.7.10" },
  { file: "smi-numbers-1.3.6.1.5.5.7.11", oid: "1.3.6.1.5.5.7.11" },
  { file: "smi-numbers-1.3.6.1.5.5.7.12", oid: "1.3.6.1.5.5.7.12" },
  { file: "smi-numbers-1.3.6.1.5.5.7.13", oid: "1.3.6.1.5.5.7.13" },
  { file: "smi-numbers-1.3.6.1.5.5.7.14", oid: "1.3.6.1.5.5.7.14" },
  { file: "smi-numbers-1.3.6.1.5.5.7.15", oid: "1.3.6.1.5.5.7.15" },
  { file: "smi-numbers-1.3.6.1.5.5.7.16", oid: "1.3.6.1.5.5.7.16" },
  { file: "smi-numbers-1.3.6.1.5.5.7.17", oid: "1.3.6.1.5.5.7.17" },
  { file: "smi-numbers-1.3.6.1.5.5.7.18", oid: "1.3.6.1.5.5.7.18" },
  { file: "smi-numbers-1.3.6.1.5.5.7.19", oid: "1.3.6.1.5.5.7.19" },
  { file: "smi-numbers-1.3.6.1.5.5.7.20", oid: "1.3.6.1.5.5.7.20" },
  { file: "smi-numbers-1.3.6.1.5.5.7.21", oid: "1.3.6.1.5.5.7.21" },
  { file: "smi-numbers-1.3.6.1.5.5.7.22", oid: "1.3.6.1.5.5.7.22" },
  { file: "smi-numbers-1.3.6.1.5.5.7.24", oid: "1.3.6.1.5.5.7.24" },
  { file: "smi-numbers-1.3.6.1.5.5.7.25", oid: "1.3.6.1.5.5.7.25" },
  { file: "smi-numbers-1.3.6.1.5.5.7.48", oid: "1.3.6.1.5.5.7.48" },
  { file: "smi-numbers-1.3.6.1.5.5.7.48.1", oid: "1.3.6.1.5.5.7.48.1" },
  { file: "security-smime-0", oid: "1.2.840.113549.1.9.16.0" },
  { file: "security-smime-1", oid: "1.2.840.113549.1.9.16.1" },
  { file: "security-smime-2", oid: "1.2.840.113549.1.9.16.2" },
  { file: "security-smime-3", oid: "1.2.840.113549.1.9.16.3" },
  { file: "security-smime-4", oid: "1.2.840.113549.1.9.16.4" },
  { file: "security-smime-5", oid: "1.2.840.113549.1.9.16.5" },
  { file: "security-smime-6", oid: "1.2.840.113549.1.9.16.6" },
  { file: "security-smime-7", oid: "1.2.840.113549.1.9.16.7" },
  { file: "security-smime-8", oid: "1.2.840.113549.1.9.16.8" },
  { file: "security-smime-9", oid: "1.2.840.113549.1.9.16.9" },
  { file: "security-smime-10", oid: "1.2.840.113549.1.9.16.10" },
  { file: "security-smime-11", oid: "1.2.840.113549.1.9.16.11" },
  { file: "security-smime-12", oid: "1.2.840.113549.1.9.16.12" },
  { file: "security-smime-13", oid: "1.2.840.113549.1.9.16.13" },
  { file: "ltans-module-identifiers", oid: "1.3.6.1.5.5.11.0" },
  { file: "ltans-cms-content-types", oid: "1.3.6.1.5.5.11.1" },
  { file: "ltans-ers-encryption-methods", oid: "1.3.6.1.5.5.11.2" },
  { file: "smi-numbers-1.3.101", oid: "1.3.101" },
  { file: "smi-numbers-26", oid: "1.3.6.1.5" },
  { file: "smi-numbers-29", oid: null },
  { file: "smi-numbers-2", oid: "1.3.6.1.2" },
  { file: "smi-numbers-24", oid: "1.3.6.1.4" },
  { file: "smi-numbers-64", oid: "1.3.6.1.7" },
  { file: "smi-numbers-65", oid: "1.3.6.1.7.1" },
  { file: "smi-numbers-66", oid: "1.3.6.1.7.1.1" },
  { file: "smi-numbers-67", oid: "1.3.6.1.7.1.2" },
  { file: "smi-numbers-70", oid: "1.3.6.1.7.1.3" },
  { file: "smi-numbers-59", oid: "1.3.6.1.6" },
  { file: "smi-numbers-60", oid: "1.3.6.1.6.1" },
  { file: "smi-numbers-61", oid: "1.3.6.1.6.3" },
];

/** 解析 CSV（含双引号转义）。 */
function parseCsv(text) {
  const rows = [];
  let row = [];
  let field = "";
  let quoted = false;
  for (let i = 0; i < text.length; i += 1) {
    const ch = text[i];
    if (quoted) {
      if (ch === '"') {
        if (text[i + 1] === '"') {
          field += '"';
          i += 1;
        } else {
          quoted = false;
        }
      } else {
        field += ch;
      }
      continue;
    }
    if (ch === '"') {
      quoted = true;
      continue;
    }
    if (ch === ",") {
      row.push(field);
      field = "";
      continue;
    }
    if (ch === "\n") {
      row.push(field);
      rows.push(row);
      row = [];
      field = "";
      continue;
    }
    if (ch === "\r") {
      continue;
    }
    field += ch;
  }
  if (field !== "" || row.length > 0) {
    row.push(field);
    rows.push(row);
  }
  return rows.filter((r) => r.length > 0 && r.some((c) => c.trim() !== ""));
}

/** 转义 CSV 字段。 */
function csvField(value) {
  const text = value ?? "";
  if (/[",\n]/.test(text)) {
    return `"${text.replace(/"/g, '""')}"`;
  }
  return text;
}

/** 校验 OID 语法。 */
function isValidOid(oid) {
  return /^\d+(\.\d+)*$/.test(oid);
}

/** 取父 OID；顶层返回空串。 */
function parentOf(oid) {
  const idx = oid.lastIndexOf(".");
  return idx < 0 ? "" : oid.slice(0, idx);
}

/** 把空白折叠成单个空格：IANA 的说明常跨行，展示时应收成一行。 */
function collapse(text) {
  return text.replace(/\s+/g, " ").trim();
}

/** 清理 IANA 的引用串：去掉方括号，保留 `RFC 5280; RFC 5758` 这样的可读形式。 */
function cleanReference(text) {
  return collapse(text).replace(/\[|\]/g, "").replace(/\s*;\s*/g, "; ");
}

/** 读一个 CSV 目录，返回行数组（不含表头）。 */
async function readCsvDir(dir) {
  const files = (await readdir(dir).catch(() => [])).filter((f) => f.endsWith(".csv")).sort();
  const out = [];
  for (const file of files) {
    const text = await readFile(path.join(dir, file), "utf8");
    const rows = parseCsv(text);
    if (rows.length === 0) {
      continue;
    }
    const header = rows[0].map((h) => h.trim());
    const idx = Object.fromEntries(header.map((h, i) => [h, i]));
    for (let i = 1; i < rows.length; i += 1) {
      const row = rows[i];
      const get = (key) => (idx[key] === undefined ? "" : (row[idx[key]] ?? "").trim());
      const oid = get("oid");
      if (oid === "") {
        continue;
      }
      out.push({ file, oid, get });
    }
  }
  return out;
}

/** 拉取或读取缓存中的 IANA CSV。 */
async function loadRegistry(reg, offline) {
  const cached = path.join(CACHE_DIR, `${reg.file}.csv`);
  if (offline) {
    const text = await readFile(cached, "utf8").catch(() => null);
    if (text === null) {
      return { text: null, note: `缺少缓存 ${reg.file}.csv` };
    }
    return { text, note: "" };
  }
  try {
    const res = await fetch(`${IANA_BASE}/${reg.file}.csv`);
    if (!res.ok) {
      const text = await readFile(cached, "utf8").catch(() => null);
      return { text, note: `HTTP ${res.status}${text ? "（用缓存）" : ""}` };
    }
    const text = await res.text();
    await mkdir(CACHE_DIR, { recursive: true });
    await writeFile(cached, text, "utf8");
    return { text, note: "" };
  } catch (err) {
    const text = await readFile(cached, "utf8").catch(() => null);
    return { text, note: `拉取失败（${err.message}）${text ? "（用缓存）" : ""}` };
  }
}

/** 解析 IANA 注册表，返回条目。 */
function parseRegistry(text, reg) {
  const rows = parseCsv(text);
  if (rows.length < 2) {
    return [];
  }
  const header = rows[0].map((h) => h.trim());
  const oidIdx = header.indexOf("OID Value");
  const decIdx = header.indexOf("Decimal");
  const nameIdx = header.indexOf("Name");
  const descIdx = header.indexOf("Description");
  const refIdx = header.indexOf("Reference");
  const out = [];
  for (let i = 1; i < rows.length; i += 1) {
    const row = rows[i];
    const cell = (idx) => (idx >= 0 ? (row[idx] ?? "").trim() : "");
    let oid = "";
    if (oidIdx >= 0) {
      oid = cell(oidIdx);
    } else if (decIdx >= 0 && reg.oid) {
      const arc = cell(decIdx);
      if (!/^\d+$/.test(arc)) {
        continue;
      }
      oid = `${reg.oid}.${arc}`;
    }
    if (!isValidOid(oid)) {
      continue;
    }
    let name = cell(nameIdx);
    const desc = cell(descIdx);
    if (name === "") {
      name = desc;
    }
    if (name === "") {
      continue;
    }
    out.push({
      oid,
      name: collapse(name),
      description: desc !== name ? collapse(desc) : "",
      source: cleanReference(cell(refIdx)),
    });
  }
  return out;
}

async function main() {
  const args = process.argv.slice(2);
  const checkOnly = args.includes("--check");
  const offline = args.includes("--offline");

  const problems = [];
  const notes = [];
  /** oid → 节点 */
  const nodes = new Map();

  const put = (node, origin) => {
    const existing = nodes.get(node.oid);
    if (existing) {
      // 已存在时只补空字段，不覆盖已有内容。
      for (const key of ["name", "description", "source", "sourceUrl"]) {
        if (!existing[key] && node[key]) {
          existing[key] = node[key];
        }
      }
      return;
    }
    nodes.set(node.oid, { ...node, origin });
  };

  // ── 1. 人工维护的完整条目 ────────────────────────────────────
  const curated = await readCsvDir(CURATED_DIR);
  for (const row of curated) {
    const get = row.get;
    if (!isValidOid(row.oid)) {
      problems.push(`${row.file}: OID 语法非法 —— ${row.oid}`);
      continue;
    }
    const aliases = get("aliases")
      .split(";")
      .map((a) => a.trim())
      .filter((a) => a !== "");
    put(
      {
        oid: row.oid,
        name: get("name"),
        nameZh: get("nameZh"),
        description: get("description"),
        descriptionZh: get("descriptionZh"),
        status: get("status") || "current",
        source: get("source"),
        sourceUrl: get("sourceUrl"),
        aliases,
      },
      row.file,
    );
  }
  notes.push(`人工条目：${curated.length} 条`);

  // ── 2. IANA 官方注册表 ───────────────────────────────────────
  let ianaCount = 0;
  const ianaProblems = [];
  for (const reg of IANA_REGISTRIES) {
    const { text, note } = await loadRegistry(reg, offline);
    if (text === null) {
      ianaProblems.push(`${reg.file}: ${note}`);
      continue;
    }
    const entries = parseRegistry(text, reg);
    for (const entry of entries) {
      put(
        {
          oid: entry.oid,
          name: entry.name,
          nameZh: "",
          description: entry.description,
          descriptionZh: "",
          status: "current",
          source: entry.source,
          sourceUrl: `${IANA_BASE}/${reg.file}.csv`,
          aliases: [],
        },
        `IANA:${reg.file}`,
      );
      ianaCount += 1;
    }
  }
  notes.push(`IANA 条目：${ianaCount} 条`);
  if (ianaProblems.length > 0) {
    notes.push(`IANA 拉取异常：${ianaProblems.length} 个`);
    for (const p of ianaProblems) {
      notes.push(`  ${p}`);
    }
  }

  // ── 3. 中文覆盖表 ────────────────────────────────────────────
  const zhRows = await readCsvDir(ZH_DIR);
  let zhApplied = 0;
  for (const row of zhRows) {
    const node = nodes.get(row.oid);
    if (!node) {
      problems.push(`${row.file}: 中文覆盖表指向不存在的 OID —— ${row.oid}`);
      continue;
    }
    const nameZh = row.get("nameZh");
    const descriptionZh = row.get("descriptionZh");
    if (nameZh === "") {
      problems.push(`${row.file}: 中文名缺失 —— ${row.oid}`);
      continue;
    }
    if (descriptionZh === "") {
      problems.push(`${row.file}: 中文说明缺失 —— ${row.oid}`);
      continue;
    }
    node.nameZh = nameZh;
    node.descriptionZh = descriptionZh;
    zhApplied += 1;
  }
  notes.push(`中文覆盖：${zhApplied} 条`);

  // ── 4. 结构校验 ──────────────────────────────────────────────
  for (const root of ["0", "1", "2"]) {
    if (!nodes.has(root)) {
      problems.push(`缺少根节点 ${root}`);
    }
  }
  for (const node of nodes.values()) {
    const parent = parentOf(node.oid);
    if (parent !== "" && !nodes.has(parent)) {
      problems.push(`父节点缺失 —— ${node.oid}（父 ${parent}）`);
    }
    if (node.name === "") {
      problems.push(`英文名缺失 —— ${node.oid}`);
    }
    if (node.nameZh === "") {
      problems.push(`中文名缺失 —— ${node.oid}`);
    }
    if (node.descriptionZh === "") {
      problems.push(`中文说明缺失 —— ${node.oid}`);
    }
    if (!["current", "reserved", "deprecated", "obsolete"].includes(node.status)) {
      problems.push(`状态非法 —— ${node.oid}（${node.status}）`);
    }
  }

  // ── 5. 输出 ──────────────────────────────────────────────────
  const ordered = [...nodes.values()].sort((a, b) => {
    const pa = a.oid.split(".").map(Number);
    const pb = b.oid.split(".").map(Number);
    for (let i = 0; i < Math.max(pa.length, pb.length); i += 1) {
      const x = pa[i] ?? -1;
      const y = pb[i] ?? -1;
      if (x !== y) {
        return x - y;
      }
    }
    return 0;
  });

  const childCount = new Map();
  for (const node of ordered) {
    const parent = parentOf(node.oid);
    if (parent !== "") {
      childCount.set(parent, (childCount.get(parent) ?? 0) + 1);
    }
  }
  const kindOf = (oid) => {
    if (oid === "0" || oid === "1" || oid === "2") {
      return "root";
    }
    return (childCount.get(oid) ?? 0) > 0 ? "branch" : "leaf";
  };

  const nodeLines = ["oid,name,description,kind,status,source,sourceUrl,aliases"];
  const zhLines = ["oid,nameZh,descriptionZh"];
  for (const node of ordered) {
    nodeLines.push(
      [
        node.oid,
        node.name,
        node.description,
        kindOf(node.oid),
        node.status,
        node.source,
        node.sourceUrl,
        node.aliases.join(";"),
      ]
        .map(csvField)
        .join(","),
    );
    zhLines.push([node.oid, node.nameZh, node.descriptionZh].map(csvField).join(","));
  }
  const nodeText = `${nodeLines.join("\n")}\n`;
  const zhText = `${zhLines.join("\n")}\n`;

  notes.push(`节点总数：${ordered.length}`);

  if (problems.length > 0) {    // `--missing` 打印全部待补的中文条目，便于逐条补齐。
    if (args.includes("--missing")) {
      const missing = [...nodes.values()]
        .filter((n) => n.nameZh === "" || n.descriptionZh === "")
        .sort((a, b) => (a.oid < b.oid ? -1 : 1));
      for (const node of missing) {
        console.log(
          [node.oid, collapse(node.name), collapse(node.description)].join("\t"),
        );
      }
    }    console.error(`❌ 数据集校验失败（${problems.length} 个问题）：`);
    for (const p of problems.slice(0, 60)) {
      console.error(`  · ${p}`);
    }
    if (problems.length > 60) {
      console.error(`  … 另有 ${problems.length - 60} 个问题`);
    }
    process.exit(1);
  }

  if (checkOnly) {
    const current = await readFile(path.join(OUT_DIR, "oid_nodes.csv"), "utf8").catch(() => "");
    const currentZh = await readFile(path.join(OUT_DIR, "oid_zh.csv"), "utf8").catch(() => "");
    console.log("── OID 数据集 ─────────────────────────");
    for (const n of notes) {
      console.log(`  · ${n}`);
    }
    if (current !== nodeText || currentZh !== zhText) {
      console.error("\n❌ 已提交的数据集与源表不一致，请运行 node scripts/oid/gen-oid-dataset.mjs");
      process.exit(1);
    }
    console.log("\n✅ 数据集与源表一致");
    return;
  }

  await mkdir(OUT_DIR, { recursive: true });
  await writeFile(path.join(OUT_DIR, "oid_nodes.csv"), nodeText, "utf8");
  await writeFile(path.join(OUT_DIR, "oid_zh.csv"), zhText, "utf8");

  console.log("── OID 数据集 ─────────────────────────");
  for (const n of notes) {
    console.log(`  · ${n}`);
  }
  console.log(`\n✅ 已写出 ${path.relative(ROOT, path.join(OUT_DIR, "oid_nodes.csv"))}`);
}

await main();
