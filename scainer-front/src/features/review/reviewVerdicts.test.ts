import { describe, it } from "node:test";
import assert from "node:assert/strict";
import { REVIEW_VERDICT_OPTIONS } from "./reviewFilterUtils.ts";
import { formatVerdictLabel, verdictChipTone } from "./reviewVerdicts.ts";

/** Синхронизировано с pkg/ejudge/verdict_filter_test.go */
const EJUDGE_STATUS_BY_VERDICT: Record<string, number> = {
  PR: 16,
  OK: 0,
  WA: 5,
  TL: 3,
  ML: 12,
  CE: 1,
  RJ: 17,
  CF: 6,
  DQ: 10,
};

describe("review filter verdicts", () => {
  for (const code of REVIEW_VERDICT_OPTIONS) {
    it(`${code} has label and tone`, () => {
      const label = formatVerdictLabel(code);
      assert.notEqual(label, "");
      assert.notEqual(label, "—");
      assert.ok(verdictChipTone(code));
      assert.ok(EJUDGE_STATUS_BY_VERDICT[code] != null, `missing ejudge status for ${code}`);
    });
  }

  it("filter compares by exact code", () => {
    assert.equal(REVIEW_VERDICT_OPTIONS.includes("PR"), true);
    assert.equal(REVIEW_VERDICT_OPTIONS.includes("AC" as (typeof REVIEW_VERDICT_OPTIONS)[number]), false);
  });

  it("DQ uses dq tone", () => {
    assert.equal(verdictChipTone("DQ"), "dq");
  });
});
