package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

func main() {
	// Get the current directory
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("Failed to get current file path")
	}
	dir := filepath.Dir(filename)
	
	// List plugin files in the plugins directory
	pluginsDir := filepath.Join(dir, "plugins")
	entries, err := os.ReadDir(pluginsDir)
	if err != nil {
		log.Fatalf("Failed to read plugins directory: %v", err)
	}
	
	// Count plugin executables
	var pluginCount int
	for _, entry := range entries {
		if !entry.IsDir() && isExecutable(filepath.Join(pluginsDir, entry.Name())) {
			pluginCount++
			fmt.Printf("Found plugin: %s\n", entry.Name())
		}
	}
	
	fmt.Printf("Found %d plugin executables\n", pluginCount)
}

// isExecutable checks if a file is executable
func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.Mode()&0111 != 0
}
