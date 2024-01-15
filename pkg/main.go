package main

import (
	"github.com/has-ghas/no-phi-ai/pkg/server"
	"github.com/rs/zerolog/log"
)

// main() function for no-phi-ai app is minimal by design
func main() {
	// setup and run the HTTP server for handling GitHub webhook events
	if err := server.Run(); err != nil {
		// panic if anything goes wrong
		log.Fatal().Err(err).Msg("runtime error in HTTP server")
	}
}
