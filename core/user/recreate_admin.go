package user

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
)

// RecreateAdminUser deletes the existing admin user (if any) and creates a new one with a fresh password hash
func (m *Manager) RecreateAdminUser(ctx context.Context) error {
	log.Println("Recreating admin user...")
	
	// First, try to find any users with admin role
	adminUsers, err := m.findAdminUsers(ctx)
	if err != nil {
		return fmt.Errorf("failed to find admin users: %w", err)
	}
	
	// Delete any existing admin users
	for _, adminUser := range adminUsers {
		log.Printf("Deleting existing admin user: %s (%s)", adminUser.Username, adminUser.ID)
		// Use direct DB delete to bypass the "can't delete last admin" check
		_, err := m.pool.Exec(ctx, "DELETE FROM users WHERE id = $1", adminUser.ID)
		if err != nil {
			return fmt.Errorf("failed to delete existing admin user: %w", err)
		}
	}
	
	// Create new admin user with fresh password hash
	adminID := uuid.New().String()
	adminUsername := "admin"
	adminPassword := "admin123"
	
	// Log admin creation but not credentials
	log.Printf("Creating new admin user with default credentials")
	
	passwordHash, err := hashPassword(adminPassword)
	if err != nil {
		return fmt.Errorf("failed to hash admin password: %w", err)
	}
	
	// Log only the hash length for security
	log.Printf("Generated password hash (length %d)", len(passwordHash))
	
	_, err = m.pool.Exec(ctx, `
		INSERT INTO users (id, username, password_hash, email, full_name, role, created_at, active)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), true)
	`, adminID, adminUsername, passwordHash, "admin@example.com", "Administrator", RoleAdmin)
	
	if err != nil {
		return fmt.Errorf("failed to insert new admin user: %w", err)
	}
	
	log.Println("===================================================================")
	log.Println("Successfully recreated admin user")
	log.Println("Default credentials have been reset")
	log.Println("===================================================================")
	
	return nil
}

// findAdminUsers finds all users with admin role
func (m *Manager) findAdminUsers(ctx context.Context) ([]*User, error) {
	rows, err := m.pool.Query(ctx, `
		SELECT id, username, password_hash, email, full_name, role, created_at, last_login, active
		FROM users
		WHERE role = $1
	`, RoleAdmin)
	if err != nil {
		return nil, fmt.Errorf("failed to query admin users: %w", err)
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
