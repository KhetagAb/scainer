package ejudge

import (
	"context"
	"fmt"

	ejudgeapi "scainer/pkg/ejudge"
)

type ClientResolver interface {
	ClientFor(ctx context.Context) (*ejudgeapi.Client, error)
}

func clientFor(ctx context.Context, clientResolver ClientResolver) (*ejudgeapi.Client, error) {
	if clientResolver == nil {
		return nil, fmt.Errorf("ejudge: client resolver не задан")
	}
	return clientResolver.ClientFor(ctx)
}