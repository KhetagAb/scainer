package transport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"

	"scainer/internal/services/analyze"
	"scainer/internal/services/contests"
	"scainer/internal/domain"
	"scainer/internal/services/ejudge/gateway"
	"scainer/generated/server"
	"scainer/internal/services/review"
	"scainer/internal/services/teachers"
	"scainer/pkg/auth"
	"scainer/pkg/jobs"
	scainermw "scainer/pkg/middleware"
)

type ejudgeGateway interface {
	BrowserLogin(ctx context.Context, contestID int) (gateway.BrowserLogin, bool, error)
	EnsureAPIKey(ctx context.Context) error
}

type Server struct {
	contests *contests.Service
	reader   *contests.ContestReader
	analyze  *analyze.Service
	review   *review.Service
	teachers *teachers.Service
	auth     auth.Service
	ejudge   ejudgeGateway
}

func New(contestsSvc *contests.Service, reader *contests.ContestReader, analyzeSvc *analyze.Service, reviewSvc *review.Service, teachersSvc *teachers.Service, authSvc auth.Service, ejudgeGw ejudgeGateway) *Server {
	return &Server{contests: contestsSvc, reader: reader, analyze: analyzeSvc, review: reviewSvc, teachers: teachersSvc, auth: authSvc, ejudge: ejudgeGw}
}

func (s *Server) Echo() *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	// За nginx / внешним reverse proxy: RealIP из X-Forwarded-For (rate limit login).
	e.IPExtractor = scainermw.ExtractIPFromProxyHeaders
	e.Use(echomiddleware.Recover())
	// Без Authorization — JWT не должен попадать в access log.
	e.Use(echomiddleware.LoggerWithConfig(echomiddleware.LoggerConfig{
		Format: `${time_rfc3339} ${remote_ip} ${method} ${uri} ${status} ${latency_human}` + "\n",
	}))
	e.Use(scainermw.LoginRateLimit(5, time.Minute))
	e.Use(scainermw.RequireJWT(s.auth))
	server.RegisterHandlers(e, s)
	// SSE вне OpenAPI/codegen (text/event-stream не в ServerInterface).
	e.GET("/api/jobs/:jobId/events", s.getJobEvents)
	return e
}

var _ server.ServerInterface = (*Server)(nil)

func (s *Server) PostAuthLogin(c echo.Context) error {
	var req server.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, server.Error{Error: "invalid request body"})
	}
	if s.teachers == nil {
		return c.JSON(http.StatusInternalServerError, server.Error{Error: "authentication is not configured"})
	}
	if err := s.teachers.Login(c.Request().Context(), req.Username, req.Password); err != nil {
		if errors.Is(err, teachers.ErrNotFound) {
			return c.JSON(http.StatusUnauthorized, server.Error{Error: "teacher not found"})
		}
		if errors.Is(err, teachers.ErrInvalidCredentials) {
			return c.JSON(http.StatusUnauthorized, server.Error{Error: "invalid credentials"})
		}
		return c.JSON(http.StatusInternalServerError, server.Error{Error: "login failed"})
	}
	token, expiresIn, err := s.auth.IssueToken(req.Username)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, server.Error{Error: "failed to issue token"})
	}
	s.warmupAPIKeyAsync(req.Username)
	return c.JSON(http.StatusOK, server.LoginResponse{
		AccessToken: token,
		ExpiresIn:   expiresIn,
	})
}

func (s *Server) warmupAPIKeyAsync(login string) {
	if s.ejudge == nil || login == "" {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(auth.WithLogin(context.Background(), login), 30*time.Second)
		defer cancel()
		_ = s.ejudge.EnsureAPIKey(ctx)
	}()
}

func (s *Server) GetAuthMe(c echo.Context) error {
	username, _ := c.Get(auth.ContextUsernameKey).(string)
	if username == "" {
		return c.JSON(http.StatusUnauthorized, server.Error{Error: "unauthorized"})
	}
	return c.JSON(http.StatusOK, server.MeResponse{Username: username})
}

func (s *Server) GetContestEjudgeLogin(c echo.Context, id server.ContestID) error {
	if s.ejudge == nil {
		return c.JSON(http.StatusServiceUnavailable, server.Error{Error: "ejudge is not configured"})
	}
	contestID, err := strconv.Atoi(string(id))
	if err != nil || contestID <= 0 {
		return c.JSON(http.StatusNotFound, server.Error{Error: "contest not found"})
	}
	login, ok, err := s.ejudge.BrowserLogin(c.Request().Context(), contestID)
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, server.Error{Error: "ejudge login unavailable"})
	}
	if !ok {
		return c.JSON(http.StatusServiceUnavailable, server.Error{Error: "ejudge login unavailable"})
	}
	return c.JSON(http.StatusOK, server.EjudgeBrowserLogin{
		BaseUrl:   login.BaseURL,
		Login:     login.Login,
		Password:  login.Password,
		ContestId: login.ContestID,
	})
}

func (s *Server) GetContests(c echo.Context) error {
	list, err := s.reader.List(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, server.Error{Error: err.Error()})
	}
	out := make([]server.ContestInfo, 0, len(list))
	for _, info := range list {
		out = append(out, toContestInfo(info))
	}
	return c.JSON(http.StatusOK, out)
}

func (s *Server) PostContests(c echo.Context) error {
	var req server.RegisterContestRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, server.Error{Error: "invalid request body"})
	}

	reg := contests.Registration{
		ID:         domain.ContestID(req.Id),
		ParallelID: ptrToStr(req.ParallelId),
	}

	contest, err := s.contests.Register(c.Request().Context(), reg)
	if err != nil {
		if errors.Is(err, contests.ErrDuplicateContest) {
			return c.JSON(http.StatusBadRequest, server.Error{Error: "contest already registered"})
		}
		return c.JSON(http.StatusInternalServerError, server.Error{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, toContestInfo(contest))
}

func (s *Server) PatchContest(c echo.Context, id server.ContestID) error {
	var req server.UpdateContestRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, server.Error{Error: "invalid request body"})
	}

	err := s.contests.SetParallel(c.Request().Context(), domain.ContestID(id), ptrToStr(req.ParallelId))
	if err != nil {
		return mapContestErr(c, err)
	}

	list, err := s.reader.List(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, server.Error{Error: err.Error()})
	}

	for _, info := range list {
		if info.ID == domain.ContestID(id) {
			return c.JSON(http.StatusOK, toContestInfo(info))
		}
	}

	return c.JSON(http.StatusNotFound, server.Error{Error: "contest not found"})
}

func (s *Server) DeleteContest(c echo.Context, id server.ContestID) error {
	err := s.contests.RemoveContest(c.Request().Context(), domain.ContestID(id))
	if err != nil {
		return mapContestErr(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (s *Server) GetContestProblems(c echo.Context, id server.ContestID) error {
	problems, err := s.reader.Problems(c.Request().Context(), domain.ContestID(id))
	if err != nil {
		return mapContestErr(c, err)
	}

	out := make([]server.ProblemInfo, 0, len(problems))
	for _, p := range problems {
		out = append(out, server.ProblemInfo{
			Id:              string(p.ID),
			Name:            p.Name,
			Excluded:        p.Excluded,
			SubmissionCount: p.SubmissionCount,
			PendingCount:    p.PendingCount,
		})
	}
	return c.JSON(http.StatusOK, out)
}

func (s *Server) PutExcludedProblems(c echo.Context, id server.ContestID) error {
	var req server.ExcludedProblemsRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, server.Error{Error: "invalid request body"})
	}

	problems := make([]domain.ProblemID, 0, len(req.Problems))
	for _, p := range req.Problems {
		problems = append(problems, domain.ProblemID(p))
	}

	err := s.contests.SetExcludedProblems(c.Request().Context(), domain.ContestID(id), problems)
	if err != nil {
		return mapContestErr(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

func (s *Server) GetContestSubmissions(c echo.Context, id server.ContestID) error {
	if err := s.requireReview(c); err != nil {
		return err
	}
	if _, err := s.reader.Problems(c.Request().Context(), domain.ContestID(id)); err != nil {
		return mapContestErr(c, err)
	}
	items, err := s.review.List(c.Request().Context(), domain.ContestID(id))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, server.Error{Error: err.Error()})
	}
	out := make([]server.SubmissionListItem, 0, len(items))
	for _, it := range items {
		out = append(out, server.SubmissionListItem{
			Id:          string(it.ID),
			Problem:     string(it.Problem),
			Participant: string(it.Participant),
			Lang:        string(it.Lang),
			SubmittedAt: it.SubmittedAt,
			Verdict:     string(it.Verdict),
		})
	}
	return c.JSON(http.StatusOK, out)
}

func (s *Server) GetSubmissionComments(c echo.Context, id server.ContestID, submissionId server.SubmissionID) error {
	if err := s.requireReview(c); err != nil {
		return err
	}
	if _, err := s.reader.Problems(c.Request().Context(), domain.ContestID(id)); err != nil {
		return mapContestErr(c, err)
	}
	res, err := s.review.LoadComments(c.Request().Context(), domain.SubmissionID(submissionId))
	if err != nil {
		return mapReviewErr(c, err)
	}
	out := server.SubmissionCommentsResponse{
		Verdict:     string(res.Verdict),
		StatusStale: res.StatusStale,
		Source:      res.Source,
		Comments:    make([]server.RunComment, 0, len(res.Comments)),
	}
	if out.Source == nil {
		out.Source = []string{}
	}
	if res.StatusError != "" {
		out.StatusError = &res.StatusError
	}
	if res.CommentsError != "" {
		out.CommentsError = &res.CommentsError
	}
	for _, cm := range res.Comments {
		rc := server.RunComment{
			Id:   cm.ID,
			From: cm.From,
			Text: cm.Text,
			Time: cm.Time,
		}
		if cm.Subject != "" {
			rc.Subject = strToPtr(cm.Subject)
		}
		out.Comments = append(out.Comments, rc)
	}
	return c.JSON(http.StatusOK, out)
}

func (s *Server) PostSubmissionComment(c echo.Context, id server.ContestID, submissionId server.SubmissionID) error {
	if err := s.requireReview(c); err != nil {
		return err
	}
	var req server.PostCommentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, server.Error{Error: "invalid request body"})
	}
	if _, err := s.reader.Problems(c.Request().Context(), domain.ContestID(id)); err != nil {
		return mapContestErr(c, err)
	}
	err := s.review.Comment(c.Request().Context(), domain.SubmissionID(submissionId), req.Text)
	if err != nil {
		return mapReviewErr(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (s *Server) PostSubmissionVerdict(c echo.Context, id server.ContestID, submissionId server.SubmissionID) error {
	if err := s.requireReview(c); err != nil {
		return err
	}
	var req server.PostVerdictRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, server.Error{Error: "invalid request body"})
	}
	if !req.Verdict.Valid() {
		return c.JSON(http.StatusBadRequest, server.Error{Error: "verdict must be OK or RJ"})
	}
	if _, err := s.reader.Problems(c.Request().Context(), domain.ContestID(id)); err != nil {
		return mapContestErr(c, err)
	}
	comment := ""
	if req.Comment != nil {
		comment = *req.Comment
	}
	err := s.review.Decide(c.Request().Context(), domain.SubmissionID(submissionId), review.DecideRequest{
		Verdict: domain.Verdict(req.Verdict),
		Comment: comment,
	})
	if err != nil {
		return mapReviewErr(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (s *Server) requireReview(c echo.Context) error {
	if s.review == nil {
		return c.JSON(http.StatusServiceUnavailable, server.Error{Error: "review is not configured (ejudge credentials)"})
	}
	return nil
}

func (s *Server) PostContestImport(c echo.Context, id server.ContestID) error {
	jobID, err := s.analyze.ImportThenAnalyze(c.Request().Context(), domain.ContestID(id))
	if err != nil {
		return mapContestErr(c, err)
	}
	return c.JSON(http.StatusAccepted, server.JobResponse{JobId: jobID})
}

func (s *Server) GetContestFindings(c echo.Context, id server.ContestID) error {
	findings, subs, err := s.reader.GetFindings(c.Request().Context(), domain.ContestID(id))
	return respondFindings(c, findings, subs, err)
}

func (s *Server) GetJob(c echo.Context, jobId string) error {
	st, ok := s.analyze.JobStatus(jobId)
	if !ok {
		return c.JSON(http.StatusNotFound, server.Error{Error: "job not found"})
	}
	return c.JSON(http.StatusOK, toJobState(st))
}

func (s *Server) getJobEvents(c echo.Context) error {
	jobID := c.Param("jobId")
	events, cancel, ok := s.analyze.SubscribeJob(jobID)
	if !ok {
		return c.JSON(http.StatusNotFound, server.Error{Error: "job not found"})
	}
	defer cancel()

	w := c.Response()
	w.Header().Set(echo.HeaderContentType, "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // иначе nginx буферизует SSE
	w.WriteHeader(http.StatusOK)

	// Analyze может идти минуты без байт — nginx рвёт тихое соединение; ":" — SSE-комментарий.
	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	ctx := c.Request().Context()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-heartbeat.C:
			if _, err := fmt.Fprint(w, ": ping\n\n"); err != nil {
				return err
			}
			w.Flush()
		case st, ok := <-events:
			if !ok {
				return nil
			}
			data, err := json.Marshal(toJobState(st))
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintf(w, "event: progress\ndata: %s\n\n", data); err != nil {
				return err
			}
			w.Flush()
			if st.Status == jobs.StatusSucceeded || st.Status == jobs.StatusFailed {
				return nil
			}
		}
	}
}

func toJobState(st jobs.State) server.JobState {
	out := server.JobState{
		Id:     st.ID,
		Status: server.JobStateStatus(st.Status),
		Progress: server.JobProgress{
			Phase: st.Progress.Phase,
			Done:  st.Progress.Done,
			Total: st.Progress.Total,
		},
	}
	if st.Progress.Label != "" {
		out.Progress.Label = &st.Progress.Label
	}
	if st.Err != "" {
		out.Error = &st.Err
	}
	if !st.StartedAt.IsZero() {
		out.StartedAt = &st.StartedAt
	}
	if !st.FinishedAt.IsZero() {
		out.FinishedAt = &st.FinishedAt
	}
	return out
}

func respondFindings(c echo.Context, findings []domain.Finding, subs map[domain.SubmissionID]domain.Submission, err error) error {
	if err != nil {
		return mapContestErr(c, err)
	}
	return c.JSON(http.StatusOK, BuildData(findings, subs))
}

func mapContestErr(c echo.Context, err error) error {
	if errors.Is(err, contests.ErrContestNotFound) {
		return c.JSON(http.StatusNotFound, server.Error{Error: "contest not found"})
	}
	if errors.Is(err, analyze.ErrJobRunning) {
		return c.JSON(http.StatusConflict, server.Error{Error: err.Error()})
	}
	return c.JSON(http.StatusInternalServerError, server.Error{Error: err.Error()})
}

func mapReviewErr(c echo.Context, err error) error {
	switch {
	case errors.Is(err, domain.ErrSubmissionNotFound):
		return c.JSON(http.StatusNotFound, server.Error{Error: "submission not found"})
	case errors.Is(err, review.ErrInvalidVerdict):
		return c.JSON(http.StatusBadRequest, server.Error{Error: err.Error()})
	case errors.Is(err, gateway.ErrAPIKeyProvision):
		return c.JSON(http.StatusServiceUnavailable, server.Error{Error: "ejudge is temporarily unavailable"})
	case errors.Is(err, gateway.ErrNoLogin):
		return c.JSON(http.StatusUnauthorized, server.Error{Error: "unauthorized"})
	default:
		return c.JSON(http.StatusBadGateway, server.Error{Error: err.Error()})
	}
}

func toContestInfo(contest contests.Contest) server.ContestInfo {
	st := contest.Statistic
	out := server.ContestInfo{
		Id:              string(contest.ID),
		Name:            contest.Name,
		ParallelId:      strToPtr(contest.ParallelID),
		LastImportedAt:  contest.LastImportedAt,
		SubmissionCount: intPtr(st.SubmissionCount),
		ProblemCount:    intPtr(st.ProblemCount),
	}
	return out
}

func intPtr(n int) *int { return &n }

func strToPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func ptrToStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
