package user

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrUserNotFound is returned when a user is not found
	ErrUserNotFound = errors.New("user not found")

	// ErrUserAlreadyExists is returned when a user with the same username already exists
	ErrUserAlreadyExists = errors.New("user already exists")

	// ErrInvalidCredentials is returned when login credentials are invalid
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// Role represents a user role
type Role string

const (
	// RoleAdmin is the administrator role
	RoleAdmin Role = "admin"

	// RoleUser is the standard user role
	RoleUser Role = "user"
)

// User represents a user in the system
type User struct {
	ID           string     `json:"id"`
	Username     string     `json:"username"`
	Email        string     `json:"email,omitempty"`
	FullName     string     `json:"full_name,omitempty"`
	PasswordHash string     `json:"-"` // Never expose in JSON
	Role         Role       `json:"role"`
	CreatedAt    time.Time  `json:"created_at"`
	LastLogin    *time.Time `json:"last_login,omitempty"`
	Active       bool       `json:"active"`
}

// CreateUserRequest represents a request to create a new user
type CreateUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Role     Role   `json:"role"`
}

// UpdateUserRequest represents a request to update an existing user
type UpdateUserRequest struct {
	Email    *string `json:"email"`
	FullName *string `json:"full_name"`
	Password *string `json:"password"`
	Role     *Role   `json:"role"`
	Active   *bool   `json:"active"`
}

// Manager provides user management functionality
type Manager struct {
	pool *pgxpool.Pool
}

// NewManager creates a new user manager
func NewManager(pool *pgxpool.Pool) *Manager {
	return &Manager{
		pool: pool,
	}
}

// InitSchema initializes the user schema
func (m *Manager) InitSchema(ctx context.Context) error {
	// Create users table if it doesn't exist
	_, err := m.pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			email TEXT UNIQUE,
			full_name TEXT,
			role TEXT NOT NULL DEFAULT 'user',
			created_at TIMESTAMPTZ DEFAULT NOW(),
			last_login TIMESTAMPTZ,
			active BOOLEAN DEFAULT TRUE
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// Check if there are any users
	var userCount int
	err = m.pool.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&userCount)
	if err != nil {
		return fmt.Errorf("failed to count users: %w", err)
	}

	// Create default admin user if there are no users
	if userCount == 0 {
		adminID := uuid.New().String()
		passwordHash, err := hashPassword("admin123")
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}

		_, err = m.pool.Exec(ctx, `
			INSERT INTO users (id, username, password_hash, email, full_name, role)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, adminID, "admin", passwordHash, "admin@example.com", "Administrator", RoleAdmin)
		if err != nil {
			return fmt.Errorf("failed to create admin user: %w", err)
		}

		log.Println("===================================================================")
		log.Println("Created default admin user: username=admin, password=admin123")
		log.Println("Make sure to change this password after your first login!")
		log.Println("===================================================================")
	}

	return nil
}

// CreateUser creates a new user
func (m *Manager) CreateUser(ctx context.Context, req CreateUserRequest) (*User, error) {
	// Check if username is already taken
	var count int
	err := m.pool.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE username = $1", req.Username).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to check username: %w", err)
	}
	if count > 0 {
		return nil, ErrUserAlreadyExists
	}

	// Hash the password
	passwordHash, err := hashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Generate a new ID
	userID := uuid.New().String()

	// Set default role if not provided
	if req.Role == "" {
		req.Role = RoleUser
	}

	// Create the user
	user := &User{
		ID:           userID,
		Username:     req.Username,
		Email:        req.Email,
		FullName:     req.FullName,
		PasswordHash: passwordHash,
		Role:         req.Role,
		CreatedAt:    time.Now(),
		Active:       true,
	}

	// Insert the user into the database
	_, err = m.pool.Exec(ctx, `
		INSERT INTO users (id, username, password_hash, email, full_name, role, created_at, active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, user.ID, user.Username, user.PasswordHash, user.Email, user.FullName, user.Role, user.CreatedAt, user.Active)
	if err != nil {
		return nil, fmt.Errorf("failed to insert user: %w", err)
	}

	return user, nil
}

// GetUser retrieves a user by ID
func (m *Manager) GetUser(ctx context.Context, id string) (*User, error) {
	var user User
	err := m.pool.QueryRow(ctx, `
		SELECT id, username, password_hash, email, full_name, role, created_at, last_login, active
		FROM users WHERE id = $1
	`, id).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.Email, &user.FullName,
		&user.Role, &user.CreatedAt, &user.LastLogin, &user.Active,
	)
	if err != nil {
		return nil, ErrUserNotFound
	}

	return &user, nil
}

// GetUserByUsername retrieves a user by username
func (m *Manager) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	var user User
	err := m.pool.QueryRow(ctx, `
		SELECT id, username, password_hash, email, full_name, role, created_at, last_login, active
		FROM users WHERE username = $1
	`, username).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.Email, &user.FullName,
		&user.Role, &user.CreatedAt, &user.LastLogin, &user.Active,
	)
	if err != nil {
		return nil, ErrUserNotFound
	}

	return &user, nil
}

// ListUsers retrieves all users
func (m *Manager) ListUsers(ctx context.Context) ([]*User, error) {
	rows, err := m.pool.Query(ctx, `
		SELECT id, username, password_hash, email, full_name, role, created_at, last_login, active
		FROM users
		ORDER BY username
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var user User
		err := rows.Scan(
			&user.ID, &user.Username, &user.PasswordHash, &user.Email, &user.FullName,
			&user.Role, &user.CreatedAt, &user.LastLogin, &user.Active,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}

// UpdateUser updates an existing user
func (m *Manager) UpdateUser(ctx context.Context, id string, req UpdateUserRequest) (*User, error) {
	// Get the current user
	user, err := m.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.FullName != nil {
		user.FullName = *req.FullName
	}
	if req.Role != nil {
		user.Role = *req.Role
	}
	if req.Active != nil {
		user.Active = *req.Active
	}
	if req.Password != nil {
		passwordHash, err := hashPassword(*req.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		user.PasswordHash = passwordHash
	}

	// Update the user in the database
	_, err = m.pool.Exec(ctx, `
		UPDATE users
		SET email = $1, full_name = $2, role = $3, active = $4, password_hash = $5
		WHERE id = $6
	`, user.Email, user.FullName, user.Role, user.Active, user.PasswordHash, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// DeleteUser deletes a user
func (m *Manager) DeleteUser(ctx context.Context, id string) error {
	result, err := m.pool.Exec(ctx, "DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// Check if any rows were affected
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// Authenticate authenticates a user
func (m *Manager) Authenticate(ctx context.Context, username, password string) (*User, error) {
	// Get the user
	user, err := m.GetUserByUsername(ctx, username)
	if err != nil {
		// To prevent username enumeration, return the same error for invalid username and invalid password
		return nil, ErrInvalidCredentials
	}

	// Log password details for debugging
	log.Printf("Attempting to verify password for user: %s", username)
	log.Printf("Stored password hash length: %d", len(user.PasswordHash))
	
	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		log.Printf("Password verification failed: %v", err)
		return nil, ErrInvalidCredentials
	}
	
	log.Printf("Password verification successful for user: %s", username)

	// Update last login time
	now := time.Now()
	user.LastLogin = &now
	_, err = m.pool.Exec(ctx, "UPDATE users SET last_login = $1 WHERE id = $2", now, user.ID)
	if err != nil {
		// Don't fail the authentication if updating last login fails
		fmt.Printf("Failed to update last login time: %v\n", err)
	}

	return user, nil
}

// hashPassword hashes a password using bcrypt
func hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}
