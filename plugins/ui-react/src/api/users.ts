import { fetchWithAuth } from '../api';

// Get API base URL from environment or use default
const API_BASE = import.meta.env.VITE_API_BASE || '';

/**
 * User role type
 */
export type UserRole = 'admin' | 'user';

/**
 * User interface
 */
export interface User {
  id: string;
  username: string;
  email?: string;
  full_name?: string;
  role: UserRole;
  created_at: string;
  last_login?: string;
  active: boolean;
}

/**
 * Create user request
 */
export interface CreateUserRequest {
  username: string;
  password: string;
  email?: string;
  full_name?: string;
  role: UserRole;
}

/**
 * Update user request
 */
export interface UpdateUserRequest {
  email?: string;
  full_name?: string;
  password?: string;
  role?: UserRole;
  active?: boolean;
}

/**
 * Login request
 */
export interface LoginRequest {
  username: string;
  password: string;
}

/**
 * Login response
 */
export interface LoginResponse {
  token: string;
  user: User;
}

/**
 * Login with username and password
 */
export const login = async (username: string, password: string): Promise<LoginResponse> => {
  try {
    const response = await fetch(`${API_BASE}/api/auth/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ username, password }),
    });

    if (!response.ok) {
      const errorText = await response.text();
      console.error("Login API error:", {
        status: response.status,
        statusText: response.statusText,
        responseText: errorText
      });
      throw new Error(`Login failed: ${response.status} ${response.statusText} - ${errorText}`);
    }

    const data = await response.json();
    console.log("Login API success response:", data);
    return data;
  } catch (error) {
    console.error("Login request error:", error);
    throw error;
  }
};

/**
 * Get all users (admin only)
 */
export const getUsers = async (): Promise<User[]> => {
  const response = await fetchWithAuth('/api/users');

  if (!response.ok) {
    throw new Error(`Failed to fetch users: ${response.statusText}`);
  }

  const data = await response.json();
  return data.users;
};

/**
 * Get a specific user by ID (admin only)
 */
export const getUser = async (id: string): Promise<User> => {
  const response = await fetchWithAuth(`/api/users/${id}`);

  if (!response.ok) {
    throw new Error(`Failed to fetch user: ${response.statusText}`);
  }

  return response.json();
};

/**
 * Create a new user (admin only)
 */
export const createUser = async (userData: CreateUserRequest): Promise<User> => {
  const response = await fetchWithAuth('/api/users', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(userData),
  });

  if (!response.ok) {
    throw new Error(`Failed to create user: ${response.statusText}`);
  }

  return response.json();
};

/**
 * Update an existing user (admin only)
 */
export const updateUser = async (id: string, userData: UpdateUserRequest): Promise<User> => {
  const response = await fetchWithAuth(`/api/users/${id}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(userData),
  });

  if (!response.ok) {
    throw new Error(`Failed to update user: ${response.statusText}`);
  }

  return response.json();
};

/**
 * Delete a user (admin only)
 */
export const deleteUser = async (id: string): Promise<void> => {
  const response = await fetchWithAuth(`/api/users/${id}`, {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error(`Failed to delete user: ${response.statusText}`);
  }
};
