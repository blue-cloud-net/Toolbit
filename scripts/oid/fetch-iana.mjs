#!/usr/bin/env node
/**
 * 下载 IANA 注册表 CSV 到本地缓存，便于离线检视格式。
 *
 * 仅用于**开发期**检视与生成数据集；运行时与 `make verify` 都不依赖网络。
 *
 * 用法：node scripts/oid/fetch-iana.mjs <registry> [...]
 *   node scripts/oid/fetch-iana.mjs smi-numbers-1.3.6.1.5.5.7.3
 */
import { mkdir, writeFile } from "node:fs/promises";
import path from "node:path";

const BASE = "https://www.iana.org/assignments/smi-numbers";
const OUT = path.resolve(import.meta.dirname, ".cache");

async function main() {
  const names = process.argv.slice(2);
  if (names.length === 0) {
    console.error("用法：node scripts/oid/fetch-iana.mjs <registry> [...]");
    process.exit(2);
  }
  await mkdir(OUT, { recursive: true });
  for (const name of names) {
    const url = `${BASE}/${name}.csv`;
    const res = await fetch(url);
    if (!res.ok) {
      console.error(`✗ ${name}: HTTP ${res.status}`);
      continue;
    }
    const text = await res.text();
    await writeFile(path.join(OUT, `${name}.csv`), text, "utf8");
    const lines = text.trim().split("\n");
    console.log(`✓ ${name}: ${lines.length} 行`);
    console.log(`  ${lines[0]}`);
    console.log(`  ${lines[1] ?? ""}`);
  }
}

await main();
