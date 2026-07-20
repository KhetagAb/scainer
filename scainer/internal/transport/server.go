package transport

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"

	"scainer/internal/contests"
	"scainer/internal/domain"
	"scainer/internal/generated/server"
	"scainer/internal/jobs"
	"scainer/pkg/auth"
	scainermw "scainer/pkg/middleware"
)

type Server struct {
	contests *contests.Service
	auth     auth.Service
}

func New(svc *contests.Service, authSvc auth.Service) *Server {
	return &Server{contests: svc, auth: authSvc}
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
	if !s.auth.ValidateCredentials(req.Username, req.Password) {
		return c.JSON(http.StatusUnauthorized, server.Error{Error: "invalid credentials"})
	}
	token, expiresIn, err := s.auth.IssueToken(req.Username)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, server.Error{Error: "failed to issue token"})
	}
	return c.JSON(http.StatusOK, server.LoginResponse{
		AccessToken: token,
		ExpiresIn:   expiresIn,
	})
}

func (s *Server) GetAuthMe(c echo.Context) error {
	username, _ := c.Get(auth.ContextUsernameKey).(string)
	if username == "" {
		return c.JSON(http.StatusUnauthorized, server.Error{Error: "unauthorized"})
	}
	return c.JSON(http.StatusOK, server.MeResponse{Username: username})
}

func (s *Server) GetContests(c echo.Context) error {
	list, err := s.contests.List(c.Request().Context())
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
		ID:           domain.ContestID(req.Id),
		ParallelID:   domain.ParallelID(ptrToStr(req.ParallelId)),
		ParallelName: ptrToStr(req.ParallelName),
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

	err := s.contests.SetParallel(c.Request().Context(), domain.ContestID(id), domain.ParallelID(ptrToStr(req.ParallelId)), ptrToStr(req.ParallelName))
	if err != nil {
		return mapContestErr(c, err)
	}

	list, err := s.contests.List(c.Request().Context())
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
	problems, err := s.contests.Problems(c.Request().Context(), domain.ContestID(id))
	if err != nil {
		return mapContestErr(c, err)
	}

	out := make([]server.ProblemInfo, 0, len(problems))
	for _, p := range problems {
		out = append(out, server.ProblemInfo{
			Id:       string(p.ID),
			Name:     p.Name,
			Excluded: p.Excluded,
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

func (s *Server) PostContestImport(c echo.Context, id server.ContestID) error {
	jobID, err := s.contests.SubmitImport(c.Request().Context(), domain.ContestID(id))
	if err != nil {
		return mapContestErr(c, err)
	}
	return c.JSON(http.StatusAccepted, server.JobResponse{JobId: jobID})
}

func (s *Server) GetContestFindings(c echo.Context, id server.ContestID) error {
	findings, subs, err := s.contests.GetFindings(c.Request().Context(), domain.ContestID(id))
	return respondFindings(c, findings, subs, err)
}

func (s *Server) GetJob(c echo.Context, jobId string) error {
	st, ok := s.contests.JobStatus(jobId)
	if !ok {
		return c.JSON(http.StatusNotFound, server.Error{Error: "job not found"})
	}
	return c.JSON(http.StatusOK, toJobState(st))
}

func (s *Server) getJobEvents(c echo.Context) error {
	jobID := c.Param("jobId")
	events, cancel, ok := s.contests.SubscribeJob(jobID)
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
	return c.JSON(http.StatusInternalServerError, server.Error{Error: err.Error()})
}

func toContestInfo(contest contests.Contest) server.ContestInfo {
	st := contest.Statistic
	out := server.ContestInfo{
		Id:              string(contest.ID),
		Name:            contest.Name,
		ParallelId:      strToPtr(string(contest.ParallelID)),
		ParallelName:    strToPtr(contest.ParallelName),
		LastImportedAt:  contest.LastImportedAt,
		SubmissionCount: intPtr(st.SubmissionCount),
		ProblemCount:    intPtr(st.ProblemCount),
		FindingsCount:   intPtr(st.FindingsCount),
	}
	if st.WeightedSuspicionPercent != nil {
		v := float32(*st.WeightedSuspicionPercent)
		out.WeightedSuspicionPercent = &v
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
