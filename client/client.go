package client

import (
	"context"

	"github.com/graphql-go/graphql"
)

// TraceClient is used to communicate with trace-api.
type TraceClient interface {
	Request(context.Context, string) (*graphql.Result, error)
}
