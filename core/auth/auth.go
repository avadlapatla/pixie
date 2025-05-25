package auth

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/time/rate"
)

var (
	// ErrInvalidToken is returned when the token is invalid
	ErrInvalidToken = errors.New("invalid token")

	// ErrTokenExpired is returned when the token has expired
	ErrTokenExpired = errors.New("token expired")

	// ErrInvalidClaims is returned when the token claims are invalid
	ErrInvalidClaims = errors.New("invalid claims")

	// ErrRateLimitExceeded is returned when the rate limit is exceeded
	ErrRateLimitExceeded = errors.New("rate limit exceeded")

	// TokenBlacklist holds revoked tokens
	tokenBlacklist = make(map[string]time.Time)
	blacklistMutex sync.RWMutex

	// RateLimiter limits the number of authentication requests
	rateLimiter = rate.NewLimiter(rate.Limit(10), 30) // 10 requests per second with burst of 30
)

// Config holds the authentication configuration
type Config struct {
	JWTAlgo           string
	JWTSecret         string
	JWTPublicKeyFile  string
	JWTPrivateKeyFile string
	TokenExpiration   time.Duration
}

// Service provides authentication functionality
type Service struct {
	config Config
	pubKey *rsa.PublicKey
}

// Claims represents the JWT claims
type Claims struct {
	jwt.RegisteredClaims
	CustomClaims map[string]interface{} `json:"custom,omitempty"`
}

// NewService creates a new authentication service
func NewService(config Config) (*Service, error) {
	service := &Service{
		config: config,
	}

	// Load the public key if using RS256
	if config.JWTAlgo == "RS256" && config.JWTPublicKeyFile != "" {
		pubKeyBytes, err := os.ReadFile(config.JWTPublicKeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read public key file: %w", err)
		}

		pubKey, err := jwt.ParseRSAPublicKeyFromPEM(pubKeyBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse public key: %w", err)
		}

		service.pubKey = pubKey
		log.Println("Successfully loaded RSA public key")
	}

	// Start the token cleanup goroutine
	go service.cleanupBlacklist()

	return service, nil
}

// GenerateToken generates a new JWT token
func (s *Service) GenerateToken(subject string, customClaims map[string]interface{}) (string, error) {
	now := time.Now()
	expirationTime := now.Add(s.config.TokenExpiration)

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
		CustomClaims: customClaims,
	}

	var token *jwt.Token
	switch s.config.JWTAlgo {
	case "HS256":
		token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(s.config.JWTSecret))
		if err != nil {
			return "", fmt.Errorf("failed to sign token with HS256: %w", err)
		}
		return tokenString, nil
	case "RS256":
		// For RS256, we need a private key
		if s.config.JWTPrivateKeyFile == "" {
			return "", errors.New("RS256 requires a private key file")
		}
		
		// Load the private key
		privateKeyBytes, err := os.ReadFile(s.config.JWTPrivateKeyFile)
		if err != nil {
			return "", fmt.Errorf("failed to read private key file: %w", err)
		}
		
		privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyBytes)
		if err != nil {
			return "", fmt.Errorf("failed to parse private key: %w", err)
		}
		
		// Create the token
		token = jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		tokenString, err := token.SignedString(privateKey)
		if err != nil {
			return "", fmt.Errorf("failed to sign token with RS256: %w", err)
		}
		return tokenString, nil
	default:
		return "", fmt.Errorf("unsupported algorithm: %s", s.config.JWTAlgo)
	}
}

// ValidateToken validates a JWT token
func (s *Service) ValidateToken(tokenString string) (string, map[string]interface{}, error) {
	// Check rate limit
	if !rateLimiter.Allow() {
		return "", nil, ErrRateLimitExceeded
	}

	// Check if token is blacklisted
	blacklistMutex.RLock()
	if _, blacklisted := tokenBlacklist[tokenString]; blacklisted {
		blacklistMutex.RUnlock()
		return "", nil, ErrInvalidToken
	}
	blacklistMutex.RUnlock()

	// Parse the token
	var token *jwt.Token
	var err error

	switch s.config.JWTAlgo {
	case "HS256":
		token, err = jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			// Validate the alg is what we expect
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(s.config.JWTSecret), nil
		})
	case "RS256":
		token, err = jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			// Validate the alg is what we expect
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return s.pubKey, nil
		})
	default:
		return "", nil, fmt.Errorf("unsupported algorithm: %s", s.config.JWTAlgo)
	}

	// Check for parsing errors
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", nil, ErrTokenExpired
		}
		return "", nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Check if the token is valid
	if !token.Valid {
		return "", nil, ErrInvalidToken
	}

	// Get the claims
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return "", nil, ErrInvalidClaims
	}

	// Check if the token has expired
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return "", nil, ErrTokenExpired
	}

	// Get the subject claim
	sub := claims.Subject
	if sub == "" {
		return "", nil, ErrInvalidClaims
	}

	// Return the subject and custom claims
	return sub, claims.CustomClaims, nil
}

// RevokeToken adds a token to the blacklist
func (s *Service) RevokeToken(tokenString string) {
	blacklistMutex.Lock()
	defer blacklistMutex.Unlock()

	// Parse the token to get the expiration time
	token, _ := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.config.JWTSecret), nil
	})

	var expirationTime time.Time
	if token != nil {
		if claims, ok := token.Claims.(*Claims); ok && claims.ExpiresAt != nil {
			expirationTime = claims.ExpiresAt.Time
		}
	}

	// If we couldn't get the expiration time, use a default
	if expirationTime.IsZero() {
		expirationTime = time.Now().Add(24 * time.Hour)
	}

	// Add the token to the blacklist
	tokenBlacklist[tokenString] = expirationTime
}

// cleanupBlacklist removes expired tokens from the blacklist
func (s *Service) cleanupBlacklist() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		blacklistMutex.Lock()
		for token, expiry := range tokenBlacklist {
			if now.After(expiry) {
				delete(tokenBlacklist, token)
			}
		}
		blacklistMutex.Unlock()
	}
}

// Middleware is an HTTP middleware that validates JWT tokens
func (s *Service) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		// Validate the token
		userId, customClaims, err := s.ValidateToken(token)
		if err != nil {
			log.Printf("Token validation failed: %v", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Set the user ID in the request context
		ctx := context.WithValue(r.Context(), "user_id", userId)
		ctx = context.WithValue(ctx, "custom_claims", customClaims)
		r = r.WithContext(ctx)

		// Set the X-User-Id header
		r.Header.Set("X-User-Id", userId)
		log.Printf("User authenticated: %s", userId)

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// HealthCheck checks if the authentication service is healthy
func (s *Service) HealthCheck() error {
	// In a real implementation, you might check database connections, etc.
	return nil
}
