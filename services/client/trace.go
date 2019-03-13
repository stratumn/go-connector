package client

import "context"

// TraceClient defines all the possible interactions with Trace.
type TraceClient interface {
	// CallTraceGql makes a call to the Trace graphql endpoint.
	CallTraceGql(ctx context.Context, query string, variables map[string]interface{}, rsp interface{}) error
}

func (c *client) CallTraceGql(ctx context.Context, query string, variables map[string]interface{}, rsp interface{}) error {
	return c.callGqlEndpoint(ctx, c.urlTrace+"/graphql", query, variables, rsp)
}
