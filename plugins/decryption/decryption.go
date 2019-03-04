package main

import (
	"fmt"

	"github.com/stratumn/go-chainscript"
)

// Decryptor provides utility functions to decrypt links and traces.
type Decryptor interface {
	DecryptLink(*chainscript.Link) (*chainscript.Link, error)
}

// DecryptionModule implements Module and Exposer and Decryptor.
// It maintains an internal cache with decrypted links.
// It exposes the Decryptor type.
type DecryptionModule struct {
}

func main() {
	fmt.Println("decryption plugin")
}
