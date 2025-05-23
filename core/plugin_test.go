package main

import (
	"log"
	"os"
	"pixie/plugin/loader"
)

func main() {
	// Set plugins directory to the current directory
	os.Setenv("PLUGINS_DIR", "../plugins")

	// Initialize plugin loader
	log.Println("Initializing plugin loader...")
	if err := loader.Init(); err != nil {
		log.Printf("Failed to initialize plugin loader: %v", err)
	}

	// Print the number of loaded plugins
	log.Printf("Loaded %d plugins", len(loader.Registry))

	// Keep the application running for a while
	log.Println("Press Ctrl+C to exit")
	select {}
}
