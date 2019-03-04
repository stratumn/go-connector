package client

import (
	"context"

	"github.com/graphql-go/graphql"
)

type TraceClient interface {
	TraceRequest(context.Context, string) (*graphql.Result, error)
}

type AccountClient interface {
	AccountRequest(context.Context, string) (*graphql.Result, error)
}


// StratumnClient is used to communicate with trace-api or account-api.
type StratumnClient interface {
	TraceClient
	AccountClient
}
