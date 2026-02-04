package main

// User service demo implementation.

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// User represents a user in the system.
type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserStore is an in-memory user store.
type UserStore struct {
	mu         sync.RWMutex
	users      map[string]*User
	byUsername map[string]*User
}

// NewUserStore creates a new user store with sample data.
func NewUserStore() *UserStore {
	store := &UserStore{
		users:      make(map[string]*User),
		byUsername: make(map[string]*User),
	}

	// Add sample users
	sampleUsers := []*User{
		{
			ID:        "user-001",
			Username:  "zhangsan",
			Email:     "zhangsan@example.com",
			Name:      "张三",
			Phone:     "13800138001",
			CreatedAt: time.Now().Add(-30 * 24 * time.Hour),
			UpdatedAt: time.Now().Add(-7 * 24 * time.Hour),
		},
		{
			ID:        "user-002",
			Username:  "lisi",
			Email:     "lisi@example.com",
			Name:      "李四",
			Phone:     "13800138002",
			CreatedAt: time.Now().Add(-20 * 24 * time.Hour),
			UpdatedAt: time.Now().Add(-3 * 24 * time.Hour),
		},
		{
			ID:        "user-003",
			Username:  "wangwu",
			Email:     "wangwu@example.com",
			Name:      "王五",
			Phone:     "13800138003",
			CreatedAt: time.Now().Add(-10 * 24 * time.Hour),
			UpdatedAt: time.Now(),
		},
	}

	for _, user := range sampleUsers {
		store.users[user.ID] = user
		store.byUsername[user.Username] = user
	}

	return store
}

// QueryUserArgs are the arguments for querying a user.
type QueryUserArgs struct {
	Username string `json:"username,omitempty" description:"Username to search for"`
	UserID   string `json:"user_id,omitempty" description:"User ID to search for"`
	Name     string `json:"name,omitempty" description:"Name to search for (partial match)"`
}

// QueryUser queries a user by various criteria.
func (s *UserStore) QueryUser(ctx context.Context, args *QueryUserArgs) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Query by ID
	if args.UserID != "" {
		if user, ok := s.users[args.UserID]; ok {
			return s.formatUser(user), nil
		}
		return "未找到ID为 " + args.UserID + " 的用户", nil
	}

	// Query by username
	if args.Username != "" {
		if user, ok := s.byUsername[args.Username]; ok {
			return s.formatUser(user), nil
		}
		return "未找到用户名为 " + args.Username + " 的用户", nil
	}

	// Query by name (partial match)
	if args.Name != "" {
		var results []*User
		for _, user := range s.users {
			if containsIgnoreCase(user.Name, args.Name) {
				results = append(results, user)
			}
		}
		if len(results) == 0 {
			return "未找到名字包含 " + args.Name + " 的用户", nil
		}
		return s.formatUsers(results), nil
	}

	return "请提供查询条件：username、user_id 或 name", nil
}

// CreateUserArgs are the arguments for creating a user.
type CreateUserArgs struct {
	Username string `json:"username" description:"Username for the new user"`
	Email    string `json:"email" description:"Email address"`
	Name     string `json:"name" description:"Display name"`
	Phone    string `json:"phone,omitempty" description:"Phone number (optional)"`
}

// CreateUser creates a new user.
func (s *UserStore) CreateUser(ctx context.Context, args *CreateUserArgs) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if username exists
	if _, exists := s.byUsername[args.Username]; exists {
		return "用户名 " + args.Username + " 已存在", nil
	}

	user := &User{
		ID:        fmt.Sprintf("user-%03d", len(s.users)+1),
		Username:  args.Username,
		Email:     args.Email,
		Name:      args.Name,
		Phone:     args.Phone,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	s.users[user.ID] = user
	s.byUsername[user.Username] = user

	return fmt.Sprintf("用户创建成功！\n%s", s.formatUser(user)), nil
}

// ListUsersArgs are the arguments for listing users.
type ListUsersArgs struct {
	Limit int `json:"limit,omitempty" description:"Maximum number of users to return"`
}

// ListUsers lists all users.
func (s *UserStore) ListUsers(ctx context.Context, args *ListUsersArgs) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	limit := args.Limit
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	var users []*User
	count := 0
	for _, user := range s.users {
		if count >= limit {
			break
		}
		users = append(users, user)
		count++
	}

	if len(users) == 0 {
		return "暂无用户数据", nil
	}

	return fmt.Sprintf("共 %d 个用户：\n%s", len(s.users), s.formatUsers(users)), nil
}

func (s *UserStore) formatUser(user *User) string {
	data, _ := json.MarshalIndent(user, "", "  ")
	return string(data)
}

func (s *UserStore) formatUsers(users []*User) string {
	var result string
	for i, user := range users {
		if i > 0 {
			result += "\n---\n"
		}
		result += s.formatUser(user)
	}
	return result
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if toLower(s[i+j]) != toLower(substr[j]) {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func toLower(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c + 32
	}
	return c
}
