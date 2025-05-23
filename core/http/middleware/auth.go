package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	pluginv1 "pixie/gen/plugin/v1"
	"pixie/plugin/loader"
)

// Auth is a middleware that validates JWT tokens
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if any auth plugins are registered
		if len(loader.Registry) == 0 {
			log.Println("No auth plugins registered, bypassing authentication")
			next.ServeHTTP(w, r)
			return
		}

		// Get the Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Check if it's a Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
			return
		}

		token := parts[1]
		if token == "" {
			http.Error(w, "Token cannot be empty", http.StatusUnauthorized)
			return
		}

		// Create a request for the plugin
		req := &pluginv1.ValidateTokenRequest{
			Token: token,
		}

		// Try each plugin until one succeeds
		var authenticated bool
		var userId string

		err := loader.ForEach(func(plugin pluginv1.PhotoPluginClient) error {
			// Create a context with timeout
			ctx, cancel := context.WithTimeout(r.Context(), 200*time.Millisecond)
			defer cancel()

			// Call the plugin
			resp, err := plugin.ValidateToken(ctx, req)
			if err != nil {
				log.Printf("Error calling ValidateToken: %v", err)
				return nil // Skip this plugin and try the next one
			}

			// Check if the token is valid
			if resp.Ok {
				authenticated = true
				userId = resp.UserId
				return nil // Stop iterating
			}

			// Log the error if the token is invalid
			if resp.Error != "" {
				log.Printf("Token validation failed: %s", resp.Error)
			}

			return nil // Continue to the next plugin
		})

		if err != nil {
			log.Printf("Error iterating through plugins: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// If no plugin authenticated the token, return 401
		if !authenticated {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Set the X-User-Id header
		r.Header.Set("X-User-Id", userId)

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
