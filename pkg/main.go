package main

import "github.com/has-ghas/no-phi-ai/pkg/server"

// main() function for no-phi-ai app is minimal by design
func main() {
	// setup an HTTP(S) server for handling GitHub webhook events
	manager := server.NewManagerOrDie()
	// use the manager to run the HTTP(S) server
	manager.Serve()
}
