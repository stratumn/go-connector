package main

import (
	"fmt"
)

// APIModule implements Module.
// It can be used to expose custom handlers on the connector.
type APIModule struct {
}

func main() {
	fmt.Println("api plugin")
}
