import SourceCode from "@/features/code/SourceCode";
import {
  REVIEW_LEGEND_DEMO_PROBLEMS,
  reviewLegendProblemLabel,
} from "@/features/review/reviewLegendDemo";

function ChevronDownIcon({ size = 28 }: { size?: number }) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width={size}
      height={size}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden
    >
      <path d="m6 9 6 6 6-6" />
    </svg>
  );
}

const DEMO_LINES = [
  "#include <bits/stdc++.h>",
  "using namespace std;",
  "const int INF = 1e9;",
  "void dijkstra(int s,",
  "  vector<vector<pair<int,int>>>& g,",
  "  vector<int>& d) {",
  "  priority_queue<pair<int,int>,",
  "    vector<pair<int,int>>, greater<>> q;",
  "  d.assign(g.size(), INF);",
  "  d[s] = 0;",
  "  q.push({0, s});",
  "  while (!q.empty()) {",
  "    auto [du, u] = q.top(); q.pop();",
  "    if (du > d[u]) continue;",
  "    for (auto [v, w] : g[u])",
  "      if (d[u] + w < d[v]) {",
  "        d[v] = d[u] + w;",
  "        q.push({d[v], v});",
  "      }",
  "  }",
  "}",
  "int main() {",
  "  int n, m;",
  "  cin >> n >> m;",
  "  vector<vector<pair<int,int>>> g(n);",
  "  for (int i = 0; i < m; ++i) {",
  "    int a, b, w;",
  "    cin >> a >> b >> w;",
  "    --a; --b;",
  "    g[a].push_back({b, w});",
  "    g[b].push_back({a, w});",
  "  }",
  "  int s;",
  "  cin >> s; --s;",
  "  vector<int> d;",
  "  dijkstra(s, g, d);",
  "  for (int i = 0; i < n; ++i)",
  "    cout << (d[i] == INF ? -1 : d[i])",
  "         << (i + 1 == n ? '\\n' : ' ');",
  "  return 0;",
  "}",
];

type Props = {
  preview?: boolean;
  className?: string;
};

export default function ReviewLegendPreview({ preview = true, className }: Props) {
  return (
    <div
      className={`review-legend-preview${className ? ` ${className}` : ""}`}
      aria-hidden={preview}
    >
      <div className="review-picker-row">
        <ul className="review-picker" aria-hidden>
          {REVIEW_LEGEND_DEMO_PROBLEMS.map((p) => (
            <li key={p.id} className="review-picker__item">
              <span
                className={`chip problem-chip review-picker__chip${p.active ? " is-active" : ""}${p.pr === 0 ? " is-empty" : ""}`}
              >
                {reviewLegendProblemLabel(p.id, p.name)}
                <span className="review-picker__pr">{p.pr}</span>
              </span>
            </li>
          ))}
        </ul>
        <label className="review-picker__filter">
          <input
            type="checkbox"
            className="review-picker__filter-input"
            checked
            readOnly
            tabIndex={-1}
          />
          <span className="review-picker__filter-label">PR only</span>
        </label>
      </div>

      <article className="review-detail review-detail--panel">
        <div className="review-workspace">
          <div className="review-workspace__label-row">
            <h3 className="review-workspace__label">Код</h3>
            <div className="review-workspace__more code-pane-bar">
              <div className="code-pane-meta">
                <span className="contest-chip">ejudge:50500:1</span>
              </div>
            </div>
          </div>
          <div className="review-workspace__label-spacer" aria-hidden />

          <div className="review-workspace__code" aria-label="Исходный код">
            <SourceCode
              lines={DEMO_LINES}
              lang="cpp"
              className="review-source__code"
              matchLine={(lineNo) => lineNo >= 15 && lineNo <= 17}
              onLineNumberClick={() => {}}
            />
          </div>

          <div className="review-side-rail">
            <div className="review-side-rail__track">
              <aside className="review-side-panel">
                <div className="review-side-head">
                  <span className="review-side-head__name">Иванов</span>
                  <span className="review-verdict-chip review-verdict-chip--pr">
                    Pending Review
                  </span>
                </div>
                <form
                  className="review-verdict"
                  onSubmit={(e) => {
                    e.preventDefault();
                  }}
                >
                  <textarea
                    className="review-verdict__input"
                    rows={5}
                    readOnly
                    tabIndex={-1}
                    value="Строки 15–17: "
                    aria-hidden
                  />
                  <div className="review-verdict__actions">
                    <div className="review-verdict__ok-rj">
                      <span className="btn btn--icon btn--ok" aria-hidden>
                        AC
                      </span>
                      <span className="btn btn--icon btn--rj" aria-hidden>
                        RJ
                      </span>
                    </div>
                    <span className="btn btn--comment" aria-hidden>
                      Comment
                    </span>
                  </div>
                </form>
                <section className="review-comments" aria-hidden>
                  <h3 className="review-comments__title">
                    Комментарии
                    <span className="review-comments__count">1</span>
                  </h3>
                  <ul className="review-comments__list">
                    <li className="review-comments__item">
                      <div className="review-comments__meta">
                        <span className="review-comments__author">Хет:</span>
                        <time className="review-comments__time">4:20</time>
                      </div>
                      <pre className="review-comments__text">
                        Строки 15–17: проверьте релаксацию
                      </pre>
                    </li>
                  </ul>
                </section>
              </aside>

              <div className="review-side-rail__scroll" aria-hidden>
                <span className="review-scroll-hint" aria-hidden>
                  <span className="review-scroll-hint__label">Следующая посылка</span>
                  <ChevronDownIcon />
                </span>
              </div>
            </div>
          </div>
        </div>
      </article>
    </div>
  );
}
