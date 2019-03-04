package main

import (
	"fmt"
)

// ValidationModule implements Module and Pluggable.
// It requires a decryption module to be instantiated since links need
// to be decrypted before validation can be applied.
// It needs the Decryption module.
type ValidationModule struct {
}

func main() {
	fmt.Println("validation plugin")
}
