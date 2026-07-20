import type { FindingView, SubjectView } from "@/client/types.gen";

export type GroupBy = "problem" | "participant";

export type FindingGroup = {
  key: string;
  contest: string;
  contestName: string;
  problem: string;
  problemName: string;
  findings: FindingView[];
};

export function scoreLevel(score: number): string {
  if (score >= 0.85) return "level-crit";
  if (score >= 0.6) return "level-high";
  if (score >= 0.3) return "level-mid";
  if (score > 0) return "level-low";
  return "level-0";
}

export function fmtScore(score: number): string {
  return (Math.round(score * 100) / 100).toFixed(2);
}

export function formatSignalCount(n: number): string {
  const mod10 = n % 10;
  const mod100 = n % 100;
  let word: string;
  if (mod10 === 1 && mod100 !== 11) word = "сигнал";
  else if (mod10 >= 2 && mod10 <= 4 && (mod100 < 12 || mod100 > 14)) word = "сигнала";
  else word = "сигналов";
  return `${n} ${word}`;
}

export function problemDisplay(id: string, shortName?: string | null): string {
  if (shortName && id && shortName !== id) return `${shortName} - ${id}`;
  return shortName || id || "";
}

export function problemGroupKey(subject: SubjectView): string {
  const problem = subject.problem || "(без задачи)";
  if (subject.contest) return `${subject.contest} · ${problem}`;
  return problem;
}

export function subjectTitle(subject: SubjectView, groupBy: GroupBy, groupKey: string): string {
  const parts = subject.participants ?? [];
  if (groupBy === "problem") {
    if (parts.length) return parts.join(" · ");
    if (subject.submission) return String(subject.submission);
    return subject.kind || "finding";
  }
  const others = parts.filter((p) => p !== groupKey);
  let label = others.length
    ? `с ${others.join(" · ")}`
    : parts.join(" · ") || subject.kind || "finding";
  if (!subject.problem && subject.submission) label += ` · ${subject.submission}`;
  return label;
}

export function searchBlob(finding: FindingView): string {
  const parts: Array<string | undefined | null> = [
    finding.key,
    finding.subject.kind,
    finding.subject.contest,
    finding.subject.contest_name,
    finding.subject.problem,
    finding.subject.problem_name,
    finding.subject.submission,
  ];
  (finding.subject.participants ?? []).forEach((p) => parts.push(p));
  finding.signals.forEach((sig) => {
    parts.push(sig.detector);
    if (sig.ai) parts.push("ai");
    (sig.evidence ?? []).forEach((ev) => {
      (ev.spans ?? []).forEach((sp) => parts.push(sp.submission));
    });
  });
  return parts.filter(Boolean).join(" ").toLowerCase();
}

export function uniqueDetectors(signals: FindingView["signals"]): string[] {
  const seen = new Set<string>();
  const out: string[] = [];
  for (const sig of signals ?? []) {
    if (!sig.detector || seen.has(sig.detector)) continue;
    seen.add(sig.detector);
    out.push(sig.detector);
  }
  return out;
}

export function groupFindings(findings: FindingView[], groupBy: GroupBy): FindingGroup[] {
  type Acc = {
    contest: string;
    contestName: string;
    problem: string;
    problemName: string;
    findings: FindingView[];
  };
  const map = new Map<string, Acc>();
  const order: string[] = [];

  const ensure = (key: string, meta?: Partial<Acc>) => {
    let g = map.get(key);
    if (!g) {
      g = {
        contest: meta?.contest ?? "",
        contestName: meta?.contestName ?? "",
        problem: meta?.problem ?? "",
        problemName: meta?.problemName ?? "",
        findings: [],
      };
      map.set(key, g);
      order.push(key);
    }
    return g;
  };

  for (const f of findings) {
    if (groupBy === "problem") {
      ensure(problemGroupKey(f.subject), {
        contest: f.subject.contest || "",
        contestName: f.subject.contest_name || "",
        problem: f.subject.problem || "(без задачи)",
        problemName: f.subject.problem_name || "",
      }).findings.push(f);
      continue;
    }
    const parts = f.subject.participants ?? [];
    if (!parts.length) {
      ensure("(без участника)").findings.push(f);
      continue;
    }
    for (const p of parts) ensure(p).findings.push(f);
  }

  order.sort((a, b) => (a < b ? -1 : a > b ? 1 : 0));

  return order.map((key) => {
    const g = map.get(key)!;
    const items = g.findings.slice().sort((a, b) => b.score - a.score);
    return {
      key,
      contest: g.contest,
      contestName: g.contestName,
      problem: g.problem,
      problemName: g.problemName,
      findings: items,
    };
  });
}
