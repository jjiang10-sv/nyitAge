package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// User represents a user profile
type User struct {
	UserID         int64      `json:"user_id" db:"user_id"`
	Username       string     `json:"username" db:"username"`
	Email          string     `json:"email" db:"email"`
	PasswordHash   string     `json:"-" db:"password_hash"`
	DisplayName    string     `json:"display_name" db:"display_name"`
	Bio            string     `json:"bio" db:"bio"`
	AvatarURL      string     `json:"avatar_url" db:"avatar_url"`
	FollowersCount int64      `json:"followers_count" db:"followers_count"`
	FollowingCount int64      `json:"following_count" db:"following_count"`
	PostsCount     int64      `json:"posts_count" db:"posts_count"`
	IsVerified     bool       `json:"is_verified" db:"is_verified"`
	IsCelebrity    bool       `json:"is_celebrity" db:"is_celebrity"`
	IsActive       bool       `json:"is_active" db:"is_active"`
	IsPrivate      bool       `json:"is_private" db:"is_private"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
	LastLoginAt    *time.Time `json:"last_login_at" db:"last_login_at"`
}

// CreateUserRequest represents user registration data
type CreateUserRequest struct {
	Username    string `json:"username" validate:"required,min=3,max=50"`
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=8"`
	DisplayName string `json:"display_name" validate:"max=100"`
	Bio         string `json:"bio" validate:"max=500"`
}

// UpdateUserRequest represents user profile update data
type UpdateUserRequest struct {
	DisplayName string `json:"display_name" validate:"max=100"`
	Bio         string `json:"bio" validate:"max=500"`
	AvatarURL   string `json:"avatar_url" validate:"url"`
	IsPrivate   *bool  `json:"is_private"`
}

// UserSession represents an active user session
type UserSession struct {
	SessionID  string    `json:"session_id" db:"session_id"`
	UserID     int64     `json:"user_id" db:"user_id"`
	TokenHash  string    `json:"-" db:"token_hash"`
	ExpiresAt  time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	LastUsedAt time.Time `json:"last_used_at" db:"last_used_at"`
	UserAgent  string    `json:"user_agent" db:"user_agent"`
	IPAddress  string    `json:"ip_address" db:"ip_address"`
}

// UserProfileService handles user profile operations
type UserProfileService struct {
	db    *sql.DB
	redis *redis.Client
}

// NewUserProfileService creates a new user profile service
func NewUserProfileService(db *sql.DB, redisClient *redis.Client) *UserProfileService {
	return &UserProfileService{
		db:    db,
		redis: redisClient,
	}
}

// CreateUser creates a new user account
func (s *UserProfileService) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Insert user into database
	query := `
		INSERT INTO users (username, email, password_hash, display_name, bio)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING user_id, username, email, display_name, bio, followers_count, 
		          following_count, posts_count, is_verified, is_celebrity, is_active, 
		          is_private, created_at, updated_at`

	var user User
	err = s.db.QueryRowContext(ctx, query,
		req.Username, req.Email, string(hashedPassword), req.DisplayName, req.Bio,
	).Scan(
		&user.UserID, &user.Username, &user.Email, &user.DisplayName, &user.Bio,
		&user.FollowersCount, &user.FollowingCount, &user.PostsCount,
		&user.IsVerified, &user.IsCelebrity, &user.IsActive, &user.IsPrivate,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique_violation
				if pqErr.Constraint == "users_username_key" {
					return nil, fmt.Errorf("username already exists")
				}
				if pqErr.Constraint == "users_email_key" {
					return nil, fmt.Errorf("email already exists")
				}
			}
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Cache user data in Redis
	s.cacheUser(ctx, &user)

	return &user, nil
}

// GetUserByID retrieves a user by ID
func (s *UserProfileService) GetUserByID(ctx context.Context, userID int64) (*User, error) {
	// Try cache first
	if user, err := s.getUserFromCache(ctx, userID); err == nil && user != nil {
		return user, nil
	}

	// Query database
	query := `
		SELECT user_id, username, email, display_name, bio, avatar_url,
		       followers_count, following_count, posts_count,
		       is_verified, is_celebrity, is_active, is_private,
		       created_at, updated_at, last_login_at
		FROM users 
		WHERE user_id = $1 AND is_active = true`

	var user User
	err := s.db.QueryRowContext(ctx, query, userID).Scan(
		&user.UserID, &user.Username, &user.Email, &user.DisplayName,
		&user.Bio, &user.AvatarURL, &user.FollowersCount, &user.FollowingCount,
		&user.PostsCount, &user.IsVerified, &user.IsCelebrity, &user.IsActive,
		&user.IsPrivate, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Cache the result
	s.cacheUser(ctx, &user)

	return &user, nil
}

// GetUserByUsername retrieves a user by username
func (s *UserProfileService) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	query := `
		SELECT user_id, username, email, display_name, bio, avatar_url,
		       followers_count, following_count, posts_count,
		       is_verified, is_celebrity, is_active, is_private,
		       created_at, updated_at, last_login_at
		FROM users 
		WHERE username = $1 AND is_active = true`

	var user User
	err := s.db.QueryRowContext(ctx, query, username).Scan(
		&user.UserID, &user.Username, &user.Email, &user.DisplayName,
		&user.Bio, &user.AvatarURL, &user.FollowersCount, &user.FollowingCount,
		&user.PostsCount, &user.IsVerified, &user.IsCelebrity, &user.IsActive,
		&user.IsPrivate, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Cache the result
	s.cacheUser(ctx, &user)

	return &user, nil
}

// UpdateUser updates user profile information
func (s *UserProfileService) UpdateUser(ctx context.Context, userID int64, req *UpdateUserRequest) (*User, error) {
	// Build dynamic update query
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.DisplayName != "" {
		setParts = append(setParts, fmt.Sprintf("display_name = $%d", argIndex))
		args = append(args, req.DisplayName)
		argIndex++
	}

	if req.Bio != "" {
		setParts = append(setParts, fmt.Sprintf("bio = $%d", argIndex))
		args = append(args, req.Bio)
		argIndex++
	}

	if req.AvatarURL != "" {
		setParts = append(setParts, fmt.Sprintf("avatar_url = $%d", argIndex))
		args = append(args, req.AvatarURL)
		argIndex++
	}

	if req.IsPrivate != nil {
		setParts = append(setParts, fmt.Sprintf("is_private = $%d", argIndex))
		args = append(args, *req.IsPrivate)
		argIndex++
	}

	if len(setParts) == 0 {
		return s.GetUserByID(ctx, userID)
	}

	// Add updated_at
	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	// Add WHERE clause
	args = append(args, userID)

	query := fmt.Sprintf(`
		UPDATE users 
		SET %s
		WHERE user_id = $%d AND is_active = true
		RETURNING user_id, username, email, display_name, bio, avatar_url,
		          followers_count, following_count, posts_count,
		          is_verified, is_celebrity, is_active, is_private,
		          created_at, updated_at, last_login_at`,
		fmt.Sprintf("%s", setParts[0]), argIndex)

	for i := 1; i < len(setParts); i++ {
		query = fmt.Sprintf("%s, %s", query, setParts[i])
	}

	var user User
	err := s.db.QueryRowContext(ctx, query, args...).Scan(
		&user.UserID, &user.Username, &user.Email, &user.DisplayName,
		&user.Bio, &user.AvatarURL, &user.FollowersCount, &user.FollowingCount,
		&user.PostsCount, &user.IsVerified, &user.IsCelebrity, &user.IsActive,
		&user.IsPrivate, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Update cache
	s.cacheUser(ctx, &user)
	s.invalidateUserCache(ctx, userID)

	return &user, nil
}

// UpdateSocialCounts updates cached social metrics from social graph service
func (s *UserProfileService) UpdateSocialCounts(ctx context.Context, userID int64, followersCount, followingCount int64) error {
	// Determine celebrity status
	isCelebrity := followersCount >= 100000 // Celebrity threshold

	query := `
		UPDATE users 
		SET followers_count = $1, following_count = $2, is_celebrity = $3, updated_at = $4
		WHERE user_id = $5`

	_, err := s.db.ExecContext(ctx, query, followersCount, followingCount, isCelebrity, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update social counts: %w", err)
	}

	// Update cache
	s.redis.HSet(ctx, fmt.Sprintf("user:%d", userID),
		"followers_count", followersCount,
		"following_count", followingCount,
		"is_celebrity", isCelebrity,
	)

	return nil
}

// AuthenticateUser validates user credentials
func (s *UserProfileService) AuthenticateUser(ctx context.Context, username, password string) (*User, error) {
	query := `
		SELECT user_id, username, email, password_hash, display_name, bio, avatar_url,
		       followers_count, following_count, posts_count,
		       is_verified, is_celebrity, is_active, is_private,
		       created_at, updated_at, last_login_at
		FROM users 
		WHERE username = $1 AND is_active = true`

	var user User
	err := s.db.QueryRowContext(ctx, query, username).Scan(
		&user.UserID, &user.Username, &user.Email, &user.PasswordHash,
		&user.DisplayName, &user.Bio, &user.AvatarURL, &user.FollowersCount,
		&user.FollowingCount, &user.PostsCount, &user.IsVerified, &user.IsCelebrity,
		&user.IsActive, &user.IsPrivate, &user.CreatedAt, &user.UpdatedAt,
		&user.LastLoginAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invalid credentials")
		}
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Update last login time
	now := time.Now()
	user.LastLoginAt = &now
	s.db.ExecContext(ctx, "UPDATE users SET last_login_at = $1 WHERE user_id = $2", now, user.UserID)

	// Don't return password hash
	user.PasswordHash = ""

	return &user, nil
}

// Cache operations
func (s *UserProfileService) cacheUser(ctx context.Context, user *User) {
	key := fmt.Sprintf("user:%d", user.UserID)
	s.redis.HSet(ctx, key,
		"user_id", user.UserID,
		"username", user.Username,
		"display_name", user.DisplayName,
		"followers_count", user.FollowersCount,
		"following_count", user.FollowingCount,
		"is_celebrity", user.IsCelebrity,
		"is_private", user.IsPrivate,
	)
	s.redis.Expire(ctx, key, 30*time.Minute)
}

func (s *UserProfileService) getUserFromCache(ctx context.Context, userID int64) (*User, error) {
	key := fmt.Sprintf("user:%d", userID)
	result := s.redis.HGetAll(ctx, key)
	if result.Err() != nil || len(result.Val()) == 0 {
		return nil, fmt.Errorf("user not in cache")
	}

	data := result.Val()
	user := &User{
		UserID:      parseInt64(data["user_id"]),
		Username:    data["username"],
		DisplayName: data["display_name"],
		IsCelebrity: data["is_celebrity"] == "true",
		IsPrivate:   data["is_private"] == "true",
	}

	return user, nil
}

func (s *UserProfileService) invalidateUserCache(ctx context.Context, userID int64) {
	key := fmt.Sprintf("user:%d", userID)
	s.redis.Del(ctx, key)
}

// Helper function to parse int64 from string
func parseInt64(s string) int64 {
	if s == "" {
		return 0
	}
	// In real implementation, use strconv.ParseInt with error handling
	return 0 // Simplified for example
}
