package plugins

import (
	"context"
	"net/http"

	"github.com/graphql-go/graphql"
	"github.com/vektah/gqlparser/ast"

	"github.com/stratumn/go-connector/client"
)

// Module is the interface a plugin should implement.
// see https://golang.org/pkg/plugin/
type Module interface {

	// ID returns the ID of the module.
	ID() string

	// Bootstrap is called once when starting the module.
	Bootstrap(context.Context) error

	// PreProcess is called prior to forwarding the query to trace-api.
	// It takes the original parsed graphql query.
	// It returns the query that will be forwarded to trace-api.
	PreProcess(context.Context, *ast.QueryDocument) (*ast.QueryDocument, error)

	// PostProcess is called after trace-api has returned the result.
	// It takes the query that was sent to trace-api and the result that was returned.
	// It returns the result that will be returned to the user.
	PostProcess(context.Context, *ast.QueryDocument, *graphql.Result) (*graphql.Result, error)

	// Handlers allows the module to define custom http handlers
	// that will be exposed alongside the graphql endpoint.
	// As the handlers may need to communicate with the Trace API,
	// we pass the TraceClient.
	Handlers(client.StratumnClient) (http.Handler, error)
}

// Requirer connects other modules.
type Requirer interface {
	// Requires returns a set of module identifiers needed before this
	// module can start.
	Requires() map[string]struct{}

	// Resolve is given a map of exposed connected objects, giving the handler
	// a chance to use them. It must check that the types are correct, or
	// return an error.
	Resolve(dependencies map[string]interface{}) error
}

// Exposer exposes a type to other modules.
type Exposer interface {
	// Expose exposes a type to other Modules. modules that depend on
	// this module will receive the returned object in their Resolve method
	// if they have one.
	Expose() interface{}
}
