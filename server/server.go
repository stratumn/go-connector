package server

import (
	"fmt"
)

// Run runs the server.
// The connector flow goes as follow:
// - Parse a configuration file containing :
//		- trace-api endpoint
// 		- required plugins
// - Load all plugins and call their Bootstrap() method.
// - Expose the custom handlers defined by the modules' Expose() method.
// - Expose a '/graphql' endpoint that forwards the gql queries to trace-api.
// - Apply the PreProcess() method on the query before forwarding it to trace-api.
// - Apply the PostProcess() method on the result after getting the result back from trace-api.
func Run() {
	fmt.Println("running")
}
