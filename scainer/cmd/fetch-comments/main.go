// Одноразовая утилита: список комментариев по задаче контеста через ejudge API.
// Запуск из scainer/: go run ./cmd/fetch-comments -contest 50506 -problem C
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"os"
	"strings"
	"time"

	"scainer/internal/configs"
	ejgen "scainer/generated/ejudge"
	"scainer/pkg/ejudge"
	ejauth "scainer/pkg/ejudge/auth"
)

func main() {
	contestID := flag.Int("contest", 50506, "contest id")
	problem := flag.String("problem", "C", "problem short name")
	login := flag.String("login", "khetag_dz", "teacher login")
	envFile := flag.String("env", ".env", "path to .env")
	debug := flag.Bool("debug", false, "print all matching runs even without comments")
	dumpRuns := flag.Bool("dump-runs", false, "print prob fields for all listed runs")
	singleRun := flag.Int("run", 0, "fetch single run id (skips list filter)")
	viaAPI := flag.String("via-api", "", "also fetch submission via scainer API base URL")
	flag.Parse()

	if *viaAPI != "" && *singleRun > 0 {
		_, teachers, err := loadEnv(*envFile)
		if err != nil {
			fatal(err)
		}
		password, ok := teachers[*login]
		if !ok {
			fatal(fmt.Errorf("login %q not in TEACHERS_PASSWORDS", *login))
		}
		token, err := loginScainer(*viaAPI, *login, password)
		if err != nil {
			fatal(err)
		}
		contest := fmt.Sprintf("%d", *contestID)
		if err := ensureScainerContest(*viaAPI, token, contest); err != nil {
			fatal(err)
		}
		fetchViaScainerAPIWithToken(*viaAPI, token, contest, fmt.Sprintf("ejudge:%d:%d", *contestID, *singleRun))
		return
	}

	baseURL, teachers, err := loadEnv(*envFile)
	if err != nil {
		fatal(err)
	}
	password, ok := teachers[*login]
	if !ok {
		fatal(fmt.Errorf("login %q not in TEACHERS_PASSWORDS", *login))
	}

	ctx := context.Background()
	apiKey := strings.TrimSpace(os.Getenv("EJUDGE_API_KEY"))
	if apiKey == "" {
		var err error
		apiKey, err = ejauth.IssueAPIKey(ctx, baseURL, *login, password)
		if err != nil {
			fatal(fmt.Errorf("API key: %w (set EJUDGE_API_KEY to reuse stored key)", err))
		}
	}
	client, err := ejudge.New(baseURL, apiKey, 30*time.Second)
	if err != nil {
		fatal(err)
	}

	if *singleRun > 0 {
		printRunComments(ctx, client, *contestID, *singleRun, "")
		if *viaAPI != "" {
			fetchViaScainerAPI(*viaAPI, *login, password, fmt.Sprintf("%d", *contestID), fmt.Sprintf("ejudge:%d:%d", *contestID, *singleRun))
		}
		return
	}

	reply, err := client.ListRuns(ctx, *contestID, nil, nil)
	if err != nil {
		fatal(err)
	}
	if reply.Result == nil || reply.Result.Runs == nil {
		fatal(fmt.Errorf("no runs in reply"))
	}

	fmt.Printf("listed %d runs in contest %d\n", len(*reply.Result.Runs), *contestID)

	var matched int
	for _, run := range *reply.Result.Runs {
		short := problemShortName(run)
		if *dumpRuns {
			id := 0
			if run.RunId != nil {
				id = *run.RunId
			}
			fmt.Printf("run=%d prob_short=%q prob_name=%q prob_id=%v status=%q\n",
				id, short, deref(run.ProbName), derefInt(run.ProbId), deref(run.StatusStr))
		}
		if short != *problem {
			continue
		}
		id := 0
		if run.RunId != nil {
			id = *run.RunId
		}
		if *debug {
			fmt.Printf("problem %s run %d user=%s status=%s\n", short, id, deref(run.UserName), deref(run.StatusStr))
		}
		msgs, err := client.RunMessages(ctx, *contestID, id)
		if err != nil {
			fmt.Printf("run %d: ERROR %v\n", id, err)
			continue
		}
		if len(msgs) == 0 {
			if *debug {
				fmt.Printf("run %d: 0 comments\n", id)
			}
			continue
		}
		matched++
		fmt.Printf("=== run %d (%s) — %d comment(s) ===\n", id, deref(run.UserName), len(msgs))
		for _, m := range msgs {
			fmt.Printf("  [%d] %s @ %s\n", m.ClarID, m.From, m.Time.Format("2006-01-02 15:04:05"))
			for _, line := range strings.Split(m.Text, "\n") {
				fmt.Printf("      %s\n", line)
			}
		}
	}
	if matched == 0 {
		fmt.Printf("no comments for contest %d problem %s\n", *contestID, *problem)
	}
}

func fetchViaScainerAPI(apiBase, login, password, contest, submission string) {
	token, err := loginScainer(apiBase, login, password)
	if err != nil {
		fmt.Printf("scainer API login: %v\n", err)
		return
	}
	fetchViaScainerAPIWithToken(apiBase, token, contest, submission)
}

func fetchViaScainerAPIWithToken(apiBase, token, contest, submission string) {
	apiURL := fmt.Sprintf("%s/api/contests/%s/submissions/%s/comments", apiBase, contest, neturl.PathEscape(submission))
	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		fmt.Printf("scainer API: %v\n", err)
		return
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("scainer API: %v\n", err)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("\n--- scainer API %s ---\nHTTP %d\n", apiURL, resp.StatusCode)
	var pretty bytes.Buffer
	if json.Indent(&pretty, body, "", "  ") == nil {
		fmt.Println(pretty.String())
	} else {
		fmt.Println(string(body))
	}
}

func ensureScainerContest(apiBase, token, contest string) error {
	req, err := http.NewRequest(http.MethodGet, apiBase+"/api/contests/"+contest+"/problems", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return runScainerImport(apiBase, token, contest)
	}
	if resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("problems HTTP %d", resp.StatusCode)
	}

	payload := fmt.Sprintf(`{"id":%q}`, contest)
	req, err = http.NewRequest(http.MethodPost, apiBase+"/api/contests", strings.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest {
		return fmt.Errorf("register HTTP %d: %s", resp.StatusCode, body)
	}
	return runScainerImport(apiBase, token, contest)
}

func runScainerImport(apiBase, token, contest string) error {
	req, err := http.NewRequest(http.MethodPost, apiBase+"/api/contests/"+contest+"/import", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("import HTTP %d: %s", resp.StatusCode, body)
	}
	var out struct {
		JobID string `json:"jobId"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return err
	}
	if out.JobID == "" {
		return nil
	}
	deadline := time.Now().Add(2 * time.Minute)
	for time.Now().Before(deadline) {
		req, err := http.NewRequest(http.MethodGet, apiBase+"/api/jobs/"+out.JobID, nil)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		var job struct {
			Status string `json:"status"`
			Error  string `json:"error"`
		}
		_ = json.Unmarshal(body, &job)
		switch job.Status {
		case "succeeded":
			return nil
		case "failed":
			return fmt.Errorf("import job failed: %s", job.Error)
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("import job timeout")
}

func loginScainer(base, login, password string) (string, error) {
	payload := fmt.Sprintf(`{"username":%q,"password":%q}`, login, password)
	resp, err := http.Post(base+"/api/auth/login", "application/json", strings.NewReader(payload))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, body)
	}
	var out struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return "", err
	}
	if out.AccessToken == "" {
		return "", fmt.Errorf("empty access_token")
	}
	return out.AccessToken, nil
}

func printRunComments(ctx context.Context, client *ejudge.Client, contestID, runID int, user string) {
	msgs, err := client.RunMessages(ctx, contestID, runID)
	if err != nil {
		fmt.Printf("run %d: ERROR %v\n", runID, err)
		return
	}
	fmt.Printf("=== run %d", runID)
	if user != "" {
		fmt.Printf(" (%s)", user)
	}
	fmt.Printf(" — %d comment(s) ===\n", len(msgs))
	for _, m := range msgs {
		fmt.Printf("  [%d] %s @ %s\n", m.ClarID, m.From, m.Time.Format("2006-01-02 15:04:05"))
		for _, line := range strings.Split(m.Text, "\n") {
			fmt.Printf("      %s\n", line)
		}
	}
}

func loadEnv(path string) (string, map[string]string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", nil, err
	}
	var baseURL, teachersRaw string
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		switch strings.TrimSpace(k) {
		case "EJUDGE_BASE_URL":
			baseURL = strings.TrimSpace(v)
		case "TEACHERS_PASSWORDS":
			teachersRaw = strings.TrimSpace(v)
		}
	}
	if baseURL == "" {
		return "", nil, fmt.Errorf("EJUDGE_BASE_URL not set in %s", path)
	}
	teachers, err := configs.ParseTeachersPasswords(teachersRaw)
	if err != nil {
		return "", nil, err
	}
	return baseURL, teachers, nil
}

func problemShortName(run ejgen.Run) string {
	if run.ProbShortName != nil && *run.ProbShortName != "" {
		return *run.ProbShortName
	}
	if run.ProbName != nil && *run.ProbName != "" {
		return *run.ProbName
	}
	return ""
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func derefInt(p *int) any {
	if p == nil {
		return nil
	}
	return *p
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "fetch-comments: %v\n", err)
	os.Exit(1)
}
