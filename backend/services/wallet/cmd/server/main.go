package main

import (
	"github.com/fairride/shared/server"
)

func main() {
	// Phase B2: domain + gRPC contract foundation only.
	// Repository implementations and handler registration are added in Phase B3
	// when persistence is introduced.
	server.Run("wallet", nil)
}
