package ejudge

import (
	"context"
	"fmt"
	"strconv"

	ejgen "scainer/generated/ejudge"
)

type ContestInfo struct {
	ID       int
	Name     string
	Problems []ProblemBrief
}

type ProblemBrief struct {
	ID           int
	ShortName    string
	Name         string
	InternalName string
}

func (c *Client) ContestStatus(ctx context.Context, contestID int) (ContestInfo, error) {
	body, err := c.masterJSON(ctx, &ejgen.MasterJSONParams{
		Json:      ejgen.MasterJSONParamsJsonN1,
		Action:    ejgen.ContestStatusJson,
		ContestId: contestID,
	})
	if err != nil {
		return ContestInfo{}, err
	}
	reply, err := decodeReply[ejgen.ContestStatusReply](body)
	if err != nil {
		return ContestInfo{}, fmt.Errorf("ejudge contest-status: json: %w", err)
	}
	if err := EnsureOK(reply.Ok, reply.Error); err != nil {
		return ContestInfo{}, err
	}
	if reply.Result == nil || reply.Result.Contest == nil {
		return ContestInfo{}, fmt.Errorf("ejudge contest-status: пустой result.contest")
	}
	info := ContestInfo{ID: contestID}
	if reply.Result.Contest.Id != nil {
		info.ID = *reply.Result.Contest.Id
	}
	if reply.Result.Contest.Name != nil {
		info.Name = *reply.Result.Contest.Name
	}
	if reply.Result.Problems != nil {
		info.Problems = make([]ProblemBrief, 0, len(*reply.Result.Problems))
		for _, p := range *reply.Result.Problems {
			info.Problems = append(info.Problems, problemBriefFromAPI(p))
		}
	}
	return info, nil
}

func problemBriefFromAPI(p ejgen.ProblemBrief) ProblemBrief {
	out := ProblemBrief{}
	if p.Id != nil {
		out.ID = *p.Id
	}
	if p.ShortName != nil {
		out.ShortName = *p.ShortName
	}
	if p.Name != nil {
		out.Name = *p.Name
	}
	if p.InternalName != nil {
		out.InternalName = *p.InternalName
	}
	return out
}

func (p ProblemBrief) ProblemKey() string {
	if p.InternalName != "" {
		return p.InternalName
	}
	if p.ShortName != "" {
		return p.ShortName
	}
	if p.ID != 0 {
		return strconv.Itoa(p.ID)
	}
	return ""
}

func (p ProblemBrief) DisplayName() string {
	if p.ShortName != "" {
		return p.ShortName
	}
	return p.Name
}

func (c *Client) ListRuns(ctx context.Context, contestID int, firstRun, lastRun *int) (*ejgen.ListRunsReply, error) {
	body, err := c.masterJSON(ctx, &ejgen.MasterJSONParams{
		Json:      ejgen.MasterJSONParamsJsonN1,
		Action:    ejgen.ListRunsJson,
		ContestId: contestID,
		FirstRun:  firstRun,
		LastRun:   lastRun,
	})
	if err != nil {
		return nil, err
	}
	reply, err := decodeReply[ejgen.ListRunsReply](body)
	if err != nil {
		return nil, fmt.Errorf("ejudge list-runs: json: %w", err)
	}
	if err := EnsureOK(reply.Ok, reply.Error); err != nil {
		return nil, err
	}
	return &reply, nil
}
