// Entry point di root agar `go build -o server .` (Jenkins) menghasilkan binary.
// Logika sebenarnya ada di internal/server.
package main

import (
	"log"

	"app/internal/server"
)

func main() {
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
