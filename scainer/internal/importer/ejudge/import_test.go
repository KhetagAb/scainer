package ejudge

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"scainer/internal/domain"
	"scainer/internal/progress"
	"scainer/internal/store"
	ejudgeapi "scainer/pkg/ejudge"
)

func TestImport_FullThenIncremental(t *testing.T) {
	fixture, err := os.ReadFile(filepath.Join("testdata", "list-runs.json"))
	if err != nil {
		t.Fatal(err)
	}

	var listCalls, downloadCalls int
	var lastFirstRun string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/cgi-bin/new-master" && r.URL.Query().Get("action") == "list-runs-json":
			listCalls++
			lastFirstRun = r.URL.Query().Get("first_run")
			first, _ := strconv.Atoi(lastFirstRun)

			var envelope map[string]any
			if err := json.Unmarshal(fixture, &envelope); err != nil {
				t.Errorf("fixture: %v", err)
				http.Error(w, "bad fixture", 500)
				return
			}
			result := envelope["result"].(map[string]any)
			all := result["runs"].([]any)
			var filtered []any
			for _, item := range all {
				run := item.(map[string]any)
				id := int(run["run_id"].(float64))
				if id >= first {
					filtered = append(filtered, run)
				}
			}
			result["runs"] = filtered
			result["listed_runs"] = len(filtered)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(envelope)

		case r.URL.Query().Get("action") == "contest-status-json":
			writeJSON(w, map[string]any{
				"ok": true,
				"result": map[string]any{
					"contest": map[string]any{"id": 50501, "name": "Тестовый контест"},
				},
			})

		case r.URL.Path == "/cgi-bin/master" && r.URL.Query().Get("action") == "download-run":
			downloadCalls++
			runID := r.URL.Query().Get("run_id")
			_, _ = w.Write([]byte("source-of-" + runID))

		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	client, err := ejudgeapi.New(srv.URL, "tok", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	imp := &Importer{
		cfg:    Config{ContestID: 50501},
		client: client,
	}
	st := store.NewMem()
	ctx := context.Background()

	res, err := imp.Import(ctx, st)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Submissions) != 3 {
		t.Fatalf("first import: got %d subs", len(res.Submissions))
	}
	if listCalls != 1 || downloadCalls != 3 {
		t.Fatalf("calls list=%d download=%d", listCalls, downloadCalls)
	}
	if lastFirstRun != "0" {
		t.Fatalf("first_run = %q", lastFirstRun)
	}
	if res.Submissions[0].Participant != "[5] Верхошинский Марк" || res.Submissions[0].Problem != "find-cycle" || res.Submissions[0].Lang != domain.LangCPP {
		t.Fatalf("sub0 = %+v", res.Submissions[0])
	}
	if res.Submissions[0].Verdict != domain.VerdictWA || res.Submissions[1].Verdict != domain.VerdictML {
		t.Fatalf("verdicts = %q %q", res.Submissions[0].Verdict, res.Submissions[1].Verdict)
	}
	if res.Submissions[2].Participant != "[5] Амбарцумян Гордей" || res.Submissions[2].Lang != domain.LangCPP {
		t.Fatalf("sub2 = %+v", res.Submissions[2])
	}
	if string(res.Submissions[1].Source) != "source-of-1" {
		t.Fatalf("source = %q", res.Submissions[1].Source)
	}
	if res.Submissions[0].Meta["contest_name"] != "Тестовый контест" {
		t.Fatalf("contest_name = %#v", res.Submissions[0].Meta["contest_name"])
	}
	if res.Submissions[0].Meta["problem_name"] != "A" {
		t.Fatalf("problem_name = %#v", res.Submissions[0].Meta["problem_name"])
	}
	cur, ok, err := st.GetCursor(ctx, cursorKey(50501))
	if err != nil || !ok || cur != "2" {
		t.Fatalf("cursor = %q ok=%v err=%v", cur, ok, err)
	}

	listCalls, downloadCalls = 0, 0
	res, err = imp.Import(ctx, st)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Submissions) != 0 {
		t.Fatalf("second import: got %d subs", len(res.Submissions))
	}
	if listCalls != 1 || downloadCalls != 0 {
		t.Fatalf("second calls list=%d download=%d", listCalls, downloadCalls)
	}
	if lastFirstRun != "3" {
		t.Fatalf("incremental first_run = %q", lastFirstRun)
	}
	cur, ok, err = st.GetCursor(ctx, cursorKey(50501))
	if err != nil || !ok || cur != "2" {
		t.Fatalf("cursor unchanged = %q ok=%v err=%v", cur, ok, err)
	}
}

func TestImport_ProgressAbsoluteIncludesStored(t *testing.T) {
	fixture, err := os.ReadFile(filepath.Join("testdata", "list-runs.json"))
	if err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/cgi-bin/new-master" && r.URL.Query().Get("action") == "list-runs-json":
			first, _ := strconv.Atoi(r.URL.Query().Get("first_run"))
			var envelope map[string]any
			if err := json.Unmarshal(fixture, &envelope); err != nil {
				http.Error(w, "bad fixture", 500)
				return
			}
			result := envelope["result"].(map[string]any)
			all := result["runs"].([]any)
			var filtered []any
			for _, item := range all {
				run := item.(map[string]any)
				id := int(run["run_id"].(float64))
				if id >= first {
					filtered = append(filtered, run)
				}
			}
			// После курсора — ещё один «новый» ран.
			if first > 0 {
				filtered = []any{
					map[string]any{
						"run_id": 3, "user_login": "u", "prob_internal_name": "A",
						"lang_name": "g++", "status_str": "OK",
					},
				}
			}
			result["runs"] = filtered
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(envelope)
		case r.URL.Query().Get("action") == "contest-status-json":
			writeJSON(w, map[string]any{
				"ok": true,
				"result": map[string]any{
					"contest": map[string]any{"id": 50501, "name": "Тестовый контест"},
				},
			})
		case r.URL.Path == "/cgi-bin/master" && r.URL.Query().Get("action") == "download-run":
			_, _ = w.Write([]byte("src"))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	client, err := ejudgeapi.New(srv.URL, "tok", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	imp := &Importer{
		cfg:    Config{ContestID: 50501},
		client: client,
	}
	st := store.NewMem()
	ctx := context.Background()

	res, err := imp.Import(ctx, st)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.Put(ctx, res.Submissions); err != nil {
		t.Fatal(err)
	}

	var events []struct{ Done, Total int }
	ctx = progress.With(ctx, func(e progress.Event) {
		if e.Phase == "importing" {
			events = append(events, struct{ Done, Total int }{e.Done, e.Total})
		}
	})
	_, err = imp.Import(ctx, st)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) < 2 {
		t.Fatalf("events = %+v", events)
	}
	// Уже 3 в store + 1 новый → старт 3/4, финиш 4/4.
	if events[0].Done != 3 || events[0].Total != 4 {
		t.Fatalf("first event = %+v, want Done=3 Total=4", events[0])
	}
	last := events[len(events)-1]
	if last.Done != 4 || last.Total != 4 {
		t.Fatalf("last event = %+v, want Done=4 Total=4", last)
	}
}

func TestImport_DownloadErrorAborts(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("action") {
		case "list-runs-json":
			writeJSON(w, map[string]any{
				"ok": true,
				"result": map[string]any{
					"runs": []map[string]any{
						{"run_id": 0, "user_login": "u", "prob_internal_name": "A", "lang_name": "g++"},
					},
				},
			})
		case "contest-status-json":
			writeJSON(w, map[string]any{
				"ok":     true,
				"result": map[string]any{"contest": map[string]any{"id": 1, "name": "c"}},
			})
		default:
			http.Error(w, "nope", http.StatusInternalServerError)
		}
	}))
	t.Cleanup(srv.Close)

	client, err := ejudgeapi.New(srv.URL, "tok", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	imp := &Importer{
		cfg: Config{ContestID: 1},
		client: client,
	}
	st := store.NewMem()
	_, err = imp.Import(context.Background(), st)
	if err == nil {
		t.Fatal("expected error")
	}
	if _, ok, _ := st.GetCursor(context.Background(), cursorKey(1)); ok {
		t.Fatal("cursor must not be set on failure")
	}
}

// ejudge list-runs часто отдаёт только user_name без user_login.
func TestImport_FallsBackToUserName(t *testing.T) {
	var downloadCalls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("action") {
		case "list-runs-json":
			writeJSON(w, map[string]any{
				"ok": true,
				"result": map[string]any{
					"runs": []map[string]any{
						{"run_id": 10, "user_name": "Anon", "prob_internal_name": "A", "lang_name": "g++"},
						{"run_id": 11, "user_login": "alice", "prob_internal_name": "B", "lang_name": "g++"},
						{"run_id": 12, "prob_internal_name": "A", "lang_name": "g++"},
					},
				},
			})
		case "contest-status-json":
			writeJSON(w, map[string]any{
				"ok":     true,
				"result": map[string]any{"contest": map[string]any{"id": 1, "name": "c"}},
			})
		case "download-run":
			downloadCalls++
			_, _ = w.Write([]byte("src"))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	client, err := ejudgeapi.New(srv.URL, "tok", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	imp := &Importer{
		cfg: Config{ContestID: 1},
		client: client,
	}
	st := store.NewMem()
	ctx := context.Background()

	res, err := imp.Import(ctx, st)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Submissions) != 2 {
		t.Fatalf("subs = %+v", res.Submissions)
	}
	if res.Submissions[0].Participant != "Anon" || res.Submissions[1].Participant != "alice" {
		t.Fatalf("participants: %q %q", res.Submissions[0].Participant, res.Submissions[1].Participant)
	}
	if downloadCalls != 2 {
		t.Fatalf("downloadCalls = %d, want 2 (run_id=12 без login/name пропущен)", downloadCalls)
	}
	// Курсор — только по успешно импортированным (не по пропущенному run_id=12).
	cur, ok, err := st.GetCursor(ctx, cursorKey(1))
	if err != nil || !ok || cur != "11" {
		t.Fatalf("cursor = %q ok=%v err=%v", cur, ok, err)
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
