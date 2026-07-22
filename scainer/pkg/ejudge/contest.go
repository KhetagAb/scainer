package ejudge

import (
	"context"
	"fmt"

	ejgen "scainer/generated/ejudge"
)

type ContestInfo struct {
	ID   int
	Name string
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
	return info, nil
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
