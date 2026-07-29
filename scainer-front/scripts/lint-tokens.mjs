#!/usr/bin/env node
/**
 * Verifies code contract token parity across all 4 theme blocks in tokens.css.
 */
import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { dirname, join } from "node:path";

const __dirname = dirname(fileURLToPath(import.meta.url));
const tokensPath = join(__dirname, "../src/app/styles/foundation/tokens.css");
const css = readFileSync(tokensPath, "utf8");

const REQUIRED = [
  "--code-bg",
  "--code-border",
  "--code-suspicion-bg",
  "--code-suspicion-mark",
  "--code-suspicion-flash",
  "--code-match-bg",
  "--code-match-mark",
  "--code-highlight",
  "--code-diff-add",
  "--code-diff-remove",
  "--code-diff-add-mark",
  "--code-diff-remove-mark",
  "--prism-bg",
  "--prism-fg",
  "--prism-comment",
  "--prism-keyword",
  "--prism-string",
  "--prism-function",
  "--prism-number",
  "--prism-operator",
  "--prism-punctuation",
  "--verdict-ok-bg",
  "--verdict-ok-border",
  "--verdict-ok-on",
  "--verdict-rj-bg",
  "--verdict-rj-border",
  "--verdict-rj-on",
];

const BLOCK_NAMES = [
  "Arctic · light",
  "Arctic · dark",
  "Creme · light",
  "Creme · dark",
];

const sections = css.split(/\/\* ───/).slice(1);
const themeSections = sections.filter((s) => s.includes("--code-bg")).slice(0, 4);

if (themeSections.length !== 4) {
  console.error(`Expected 4 theme blocks with --code-bg, found ${themeSections.length}`);
  process.exit(1);
}

let failed = false;

for (let i = 0; i < themeSections.length; i++) {
  const block = themeSections[i];
  const name = BLOCK_NAMES[i];
  const defined = new Set([...block.matchAll(/(--[\w-]+)\s*:/g)].map((m) => m[1]));
  const missing = REQUIRED.filter((t) => !defined.has(t));

  if (missing.length) {
    failed = true;
    console.error(`\n${name}: missing ${missing.length} token(s):`);
    for (const t of missing) console.error(`  ${t}`);
  } else {
    console.log(`✓ ${name}: ${REQUIRED.length} code contract tokens`);
  }
}

const sets = themeSections.map(
  (block) =>
    new Set(
      [...block.matchAll(/(--(?:code|prism|verdict-(?:ok|rj))-[\w-]+)\s*:/g)].map((m) => m[1]),
    ),
);

const reference = sets[0];
for (let i = 1; i < sets.length; i++) {
  const extra = [...sets[i]].filter((t) => !reference.has(t));
  const missing = [...reference].filter((t) => !sets[i].has(t));
  if (extra.length || missing.length) {
    failed = true;
    console.error(`\nToken set mismatch: ${BLOCK_NAMES[0]} vs ${BLOCK_NAMES[i]}`);
    if (missing.length) console.error(`  missing in ${BLOCK_NAMES[i]}: ${missing.join(", ")}`);
    if (extra.length) console.error(`  extra in ${BLOCK_NAMES[i]}: ${extra.join(", ")}`);
  }
}

if (failed) process.exit(1);
console.log("\nAll theme blocks have matching code contract tokens.");
