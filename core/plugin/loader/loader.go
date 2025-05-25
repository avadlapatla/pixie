package loader

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"

	pluginv1 "pixie/gen/plugin/v1"
)

var (
	// Registry holds all loaded plugin clients
	Registry []pluginv1.PhotoPluginClient

	// pluginProcesses holds references to all plugin processes for cleanup
	pluginProcesses []*os.Process

	// portRegex is used to extract the port number from plugin output
	portRegex = regexp.MustCompile(`PORT=(\d+)`)

	// ErrPluginTimeout is returned when a plugin fails to start within the timeout period
	ErrPluginTimeout = errors.New("plugin failed to start within timeout period")

	// mutex for thread-safe registry operations
	mutex sync.Mutex
)

// Init initializes the plugin loader
func Init() error {
	log.Println("Initializing plugin loader...")

	// Get plugins directory from environment or use default
	pluginsDir := getEnv("PLUGINS_DIR", "./plugins")
	log.Printf("Using plugins directory: %s", pluginsDir)

	// Setup cleanup on interrupt/termination
	setupCleanup()

	// Load plugins
	if err := loadPlugins(pluginsDir); err != nil {
		return fmt.Errorf("failed to load plugins: %w", err)
	}

	log.Printf("Successfully loaded %d plugins", len(Registry))
	return nil
}

// ForEach executes the provided function for each plugin in the registry
func ForEach(fn func(pluginv1.PhotoPluginClient) error) error {
	mutex.Lock()
	defer mutex.Unlock()

	for _, plugin := range Registry {
		if err := fn(plugin); err != nil {
			return err
		}
	}
	return nil
}

// loadPlugins loads all plugins from the specified directory
func loadPlugins(dir string) error {
	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		log.Printf("Plugins directory %s does not exist, creating it", dir)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create plugins directory: %w", err)
		}
		return nil // No plugins to load
	}

	// Walk through the directory and load each executable file
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Printf("Error accessing path %s: %v", path, err)
			return nil // Continue walking
		}

		// Skip directories and node_modules
		if d.IsDir() {
			// Skip node_modules directories
			if d.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if file is executable
		info, err := d.Info()
		if err != nil {
			log.Printf("Error getting file info for %s: %v", path, err)
			return nil // Continue walking
		}

		// Check if file is executable (on Unix-like systems)
		if info.Mode()&0111 != 0 {
			// Load the plugin
			if err := loadPlugin(path); err != nil {
				log.Printf("Failed to load plugin %s: %v", path, err)
				// Continue loading other plugins
				return nil
			}
		}

		return nil
	})
}

// loadPlugin loads a single plugin
func loadPlugin(path string) error {
	log.Printf("Loading plugin: %s", path)

	// Start the plugin process
	cmd := exec.Command(path, "--port=0")
	
	// Pass environment variables to the plugin process
	cmd.Env = os.Environ()
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start plugin: %w", err)
	}

	// Store process for cleanup
	pluginProcesses = append(pluginProcesses, cmd.Process)

	// Read port from stdout
	scanner := bufio.NewScanner(stdout)
	portChan := make(chan int, 1)
	errChan := make(chan error, 1)

	// Start goroutine to scan stdout for PORT line
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			log.Printf("Plugin output: %s", line)

			// Check if line contains PORT=
			matches := portRegex.FindStringSubmatch(line)
			if len(matches) == 2 {
				port, err := strconv.Atoi(matches[1])
				if err != nil {
					errChan <- fmt.Errorf("invalid port number: %s", matches[1])
					return
				}
				portChan <- port
				return
			}
		}
		if err := scanner.Err(); err != nil {
			errChan <- fmt.Errorf("error reading plugin output: %w", err)
		}
	}()

	// Start goroutine to log stderr
	go func() {
		stderrScanner := bufio.NewScanner(stderr)
		for stderrScanner.Scan() {
			log.Printf("Plugin error: %s", stderrScanner.Text())
		}
	}()

	// Wait for port or timeout
	var port int
	select {
	case port = <-portChan:
		log.Printf("Plugin listening on port %d", port)
	case err := <-errChan:
		return fmt.Errorf("plugin error: %w", err)
	case <-time.After(5 * time.Second):
		return ErrPluginTimeout
	}

	// Connect to the plugin
	addr := fmt.Sprintf("localhost:%d", port)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to plugin: %w", err)
	}

	// Check health
	healthClient := grpc_health_v1.NewHealthClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = healthClient.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	// Create plugin client
	client := pluginv1.NewPhotoPluginClient(conn)

	// Add to registry
	mutex.Lock()
	Registry = append(Registry, client)
	mutex.Unlock()

	log.Printf("Successfully loaded plugin: %s", path)
	return nil
}

// setupCleanup registers signal handlers to clean up plugin processes
func setupCleanup() {
	// This is a simplified version - in a real application, you would use a proper
	// signal handling mechanism to catch SIGINT/SIGTERM
	go func() {
		// Wait for process to exit
		<-context.Background().Done()
		cleanup()
	}()
}

// cleanup terminates all plugin processes
func cleanup() {
	log.Println("Cleaning up plugin processes...")
	for _, process := range pluginProcesses {
		if process != nil {
			log.Printf("Terminating plugin process %d", process.Pid)
			if err := process.Kill(); err != nil {
				log.Printf("Failed to kill plugin process %d: %v", process.Pid, err)
			}
		}
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
