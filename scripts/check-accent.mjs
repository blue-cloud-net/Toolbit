#!/usr/bin/env node
/**
 * 主色与图标色调对比度校验
 *
 * 校验项：
 *   1. frontend/src/styles/tokens.css 中每个 [data-accent] 的 primary，
 *      对同作用域的 --tb-bg-elevated 对比度 >= 4.5:1
 *   2. 每个 primary 上的前景色 --tb-primary-contrast 对 primary 对比度 >= 4.5:1
 *   3. frontend/src/config/accents.ts 的色值与 tokens.css 中同名色板一致
 *   4. 语义色（success / warning / danger / info）在所有色板下取值不变
 *   5. 图标色调 --tb-tone-<名> 对 --tb-tone-<名>-soft 对比度 >= 4.5:1，
 *      且在所有主色色板下取值不变（色调不随主色漂移）
 *
 * 退出码：0 = 通过；1 = 存在问题（输出差异清单）
 *
 * 用法：node scripts/check-accent.mjs
 */

import { readFile } from "node:fs/promises";
import path from "node:path";
import process from "node:process";

const ROOT = path.resolve(import.meta.dirname, "..");
const TOKENS_CSS = path.join(ROOT, "frontend", "src", "styles", "tokens.css");
const ACCENTS_TS = path.join(ROOT, "frontend", "src", "config", "accents.ts");

/** 对比度下限（WCAG AA，正文级别）。 */
const MIN_RATIO = 4.5;

/** 语义色：这些变量在任何色板下都不得改变。 */
const SEMANTIC_VARS = ["--tb-success", "--tb-warning", "--tb-danger", "--tb-info"];

/** 图标色调前缀：色调固定，不随主色切换。 */
const TONE_PREFIX = "--tb-tone-";

const problems = [];
const notes = [];

// ── 颜色工具 ────────────────────────────────────────────────────────────

/** 十六进制色值 → [r, g, b]。 */
function toRgb(hex) {
  const h = hex.trim().replace(/^#/, "");
  if (h.length === 3) {
    return [0, 1, 2].map((i) => parseInt(h[i] + h[i], 16));
  }
  return [0, 2, 4].map((i) => parseInt(h.slice(i, i + 2), 16));
}

/** 相对亮度（WCAG 2.x）。 */
function luminance(hex) {
  const [r, g, b] = toRgb(hex).map((c) => {
    const v = c / 255;
    return v <= 0.03928 ? v / 12.92 : ((v + 0.055) / 1.055) ** 2.4;
  });
  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
}

/** 对比度。 */
function contrast(a, b) {
  const [l1, l2] = [luminance(a), luminance(b)].sort((x, y) => y - x);
  return (l1 + 0.05) / (l2 + 0.05);
}

// ── 解析 ────────────────────────────────────────────────────────────────

/**
 * 解析 tokens.css 中的选择器块。
 * 返回 [{ selector, vars: {name: value} }]
 */
function parseBlocks(css) {
  const blocks = [];
  // 只取以 :root 开头的选择器（主色与主题规则）
  const re = /(:root[^{]*)\{([^}]*)\}/g;
  let m;
  while ((m = re.exec(css)) !== null) {
    const selector = m[1].trim();
    const body = m[2];
    const vars = {};
    const vre = /(--tb-[a-z0-9-]+)\s*:\s*([^;]+);/gi;
    let v;
    while ((v = vre.exec(body)) !== null) {
      vars[v[1]] = v[2].trim();
    }
    blocks.push({ selector, vars });
  }
  return blocks;
}

/** 从选择器推断作用域。 */
function scopeOf(selector) {
  const isDark = selector.includes("[data-theme='dark']");
  const accentMatch = selector.match(/\[data-accent='([^']+)'\]/);
  return {
    isDark,
    accent: accentMatch ? accentMatch[1] : null,
  };
}

// ── 主流程 ──────────────────────────────────────────────────────────────

async function main() {
  const css = await readFile(TOKENS_CSS, "utf8");
  const ts = await readFile(ACCENTS_TS, "utf8");

  const blocks = parseBlocks(css);

  // 收集每个作用域的变量
  const scopes = new Map(); // key → { isDark, accent, vars }
  for (const block of blocks) {
    const { isDark, accent } = scopeOf(block.selector);
    const key = `${isDark ? "dark" : "light"}:${accent ?? "default"}`;
    if (!scopes.has(key)) {
      scopes.set(key, { isDark, accent, vars: {} });
    }
    Object.assign(scopes.get(key).vars, block.vars);
  }

  // 1) 基础作用域的 --tb-bg-elevated
  const lightBg = scopes.get("light:default")?.vars["--tb-bg-elevated"];
  const darkBg = scopes.get("dark:default")?.vars["--tb-bg-elevated"];
  if (!lightBg || !darkBg) {
    problems.push("未能从 tokens.css 解析出 --tb-bg-elevated");
    return;
  }

  let checked = 0;

  // 2) 每个色板 × 明暗：primary 对 bg-elevated，contrast 对 primary
  for (const [key, scope] of scopes) {
    const primary = scope.vars["--tb-primary"];
    const fg = scope.vars["--tb-primary-contrast"];
    if (!primary || !fg) {
      continue;
    }
    const bg = scope.isDark ? darkBg : lightBg;

    const c1 = contrast(primary, bg);
    if (c1 < MIN_RATIO) {
      problems.push(
        `[${key}] --tb-primary ${primary} 对 --tb-bg-elevated ${bg} 的对比度为 ${c1.toFixed(2)}:1，低于 ${MIN_RATIO}:1`,
      );
    }

    const c2 = contrast(fg, primary);
    if (c2 < MIN_RATIO) {
      problems.push(
        `[${key}] --tb-primary-contrast ${fg} 对 --tb-primary ${primary} 的对比度为 ${c2.toFixed(2)}:1，低于 ${MIN_RATIO}:1`,
      );
    }
    checked += 2;
  }

  // 3) 语义色在所有色板下必须一致
  //    注：深色模式允许对语义色**整体提亮**（否则深底上不可读），
  //    但**不得随主色变化** —— 因此只在同明暗的色板之间比较。
  const semanticBase = scopes.get("light:default")?.vars ?? {};
  const semanticBaseDark = scopes.get("dark:default")?.vars ?? {};
  for (const name of SEMANTIC_VARS) {
    if (!semanticBase[name]) {
      problems.push(`tokens.css 中缺少语义色 ${name}`);
      continue;
    }
    if (!semanticBaseDark[name]) {
      problems.push(`tokens.css 的深色作用域中缺少语义色 ${name}`);
      continue;
    }
    for (const [key, scope] of scopes) {
      const value = scope.vars[name];
      if (!value) continue;
      const expected = scope.isDark ? semanticBaseDark[name] : semanticBase[name];
      if (value !== expected) {
        problems.push(
          `[${key}] 语义色 ${name} 被改成了 ${value}，同明暗基准为 ${expected}（语义色不得随主色漂移）`,
        );
      }
    }
  }

  // 4) config/accents.ts 的色值必须与 tokens.css 一致
  const tsSwatches = [...ts.matchAll(/key:\s*'([^']+)',\s*swatch:\s*'(#[0-9a-fA-F]{6})'/g)].map(
    (m) => ({ key: m[1], swatch: m[2] }),
  );
  if (tsSwatches.length === 0) {
    problems.push("未能从 config/accents.ts 解析出色板定义");
  }
  for (const { key, swatch } of tsSwatches) {
    const scope = scopes.get(`light:${key}`);
    if (!scope) {
      problems.push(`config/accents.ts 中的色板 ${key} 在 tokens.css 中没有对应的 [data-accent='${key}'] 规则`);
      continue;
    }
    const cssPrimary = scope.vars["--tb-primary"];
    if (cssPrimary && cssPrimary.toLowerCase() !== swatch.toLowerCase()) {
      problems.push(
        `色板 ${key} 的色值不一致：config/accents.ts 为 ${swatch}，tokens.css 为 ${cssPrimary}`,
      );
    }
  }

  // 5) 图标色调：前景色对底色 >= 4.5:1，且不随主色漂移
  const toneNames = new Set();
  for (const scope of scopes.values()) {
    for (const name of Object.keys(scope.vars)) {
      if (name.startsWith(TONE_PREFIX) && !name.endsWith("-soft")) {
        toneNames.add(name.slice(TONE_PREFIX.length));
      }
    }
  }
  if (toneNames.size === 0) {
    problems.push("tokens.css 中未找到任何 --tb-tone-* 图标色调");
  }
  let toneChecked = 0;
  for (const tone of [...toneNames].sort()) {
    const fgName = `${TONE_PREFIX}${tone}`;
    const softName = `${fgName}-soft`;
    for (const [key, scope] of scopes) {
      const fg = scope.vars[fgName];
      const soft = scope.vars[softName];
      if (!fg || !soft) {
        continue;
      }
      const ratio = contrast(fg, soft);
      if (ratio < MIN_RATIO) {
        problems.push(
          `[${key}] ${fgName} ${fg} 对 ${softName} ${soft} 的对比度为 ${ratio.toFixed(2)}:1，低于 ${MIN_RATIO}:1`,
        );
      }
      toneChecked += 1;
    }
  }
  // 色调必须在所有主色色板下保持一致
  for (const tone of [...toneNames].sort()) {
    for (const suffix of ["", "-soft"]) {
      const name = `${TONE_PREFIX}${tone}${suffix}`;
      const lightValue = scopes.get("light:default")?.vars[name];
      const darkValue = scopes.get("dark:default")?.vars[name];
      if (!lightValue || !darkValue) {
        problems.push(`tokens.css 缺少图标色调 ${name}（明暗两套都要定义）`);
        continue;
      }
      for (const [key, scope] of scopes) {
        const value = scope.vars[name];
        if (!value) continue;
        const expected = scope.isDark ? darkValue : lightValue;
        if (value !== expected) {
          problems.push(
            `[${key}] 图标色调 ${name} 被改成了 ${value}，同明暗基准为 ${expected}（色调不得随主色漂移）`,
          );
        }
      }
    }
  }

  notes.push(`对比度：${checked} 项校验（${scopes.size} 个作用域）`);
  notes.push(`色板一致性：${tsSwatches.length} 组`);
  notes.push(`语义色：${SEMANTIC_VARS.length} 项（${scopes.size} 个作用域）`);
  notes.push(`图标色调：${toneNames.size} 组（${toneChecked} 项对比度）`);

  console.log("── 主色检查 ──────────────────────────────");
  for (const n of notes) console.log(`  · ${n}`);

  if (problems.length === 0) {
    console.log("\n✅ 主色检查通过");
    process.exit(0);
  }

  console.log(`\n❌ 发现 ${problems.length} 个问题：\n`);
  for (const msg of problems) console.log(`  - ${msg}`);
  process.exit(1);
}

main().catch((err) => {
  console.error("检查脚本异常：", err);
  process.exit(1);
});
