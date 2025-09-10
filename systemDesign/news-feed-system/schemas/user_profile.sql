-- User Profile Database Schema (PostgreSQL)

-- Users table with profile information and cached social metrics
CREATE TABLE users (
    user_id BIGSERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    display_name VARCHAR(100),
    bio TEXT,
    avatar_url VARCHAR(500),
    
    -- Cached social metrics (updated from social graph)
    followers_count BIGINT DEFAULT 0,
    following_count BIGINT DEFAULT 0,
    posts_count BIGINT DEFAULT 0,
    
    -- Account status
    is_verified BOOLEAN DEFAULT FALSE,
    is_celebrity BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    is_private BOOLEAN DEFAULT FALSE,
    
    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    last_login_at TIMESTAMPTZ
);

-- Indexes for performance
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_celebrity ON users(is_celebrity) WHERE is_celebrity = TRUE;
CREATE INDEX idx_users_active ON users(is_active) WHERE is_active = TRUE;
CREATE INDEX idx_users_created_at ON users(created_at);

-- User sessions for authentication
CREATE TABLE user_sessions (
    session_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id BIGINT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    last_used_at TIMESTAMPTZ DEFAULT NOW(),
    user_agent TEXT,
    ip_address INET
);

CREATE INDEX idx_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_sessions_expires_at ON user_sessions(expires_at);

-- User preferences
CREATE TABLE user_preferences (
    user_id BIGINT PRIMARY KEY REFERENCES users(user_id) ON DELETE CASCADE,
    email_notifications BOOLEAN DEFAULT TRUE,
    push_notifications BOOLEAN DEFAULT TRUE,
    privacy_level VARCHAR(20) DEFAULT 'public' CHECK (privacy_level IN ('public', 'friends', 'private')),
    language VARCHAR(10) DEFAULT 'en',
    timezone VARCHAR(50) DEFAULT 'UTC',
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Sample data
INSERT INTO users (username, email, password_hash, display_name, bio, followers_count, following_count, is_celebrity) VALUES
('john_doe', 'john@example.com', '$2a$10$hash1', 'John Doe', 'Software engineer and coffee enthusiast', 1250, 180, FALSE),
('celebrity_user', 'celeb@example.com', '$2a$10$hash2', 'Celebrity User', 'Famous person with millions of followers', 50000000, 100, TRUE),
('jane_smith', 'jane@example.com', '$2a$10$hash3', 'Jane Smith', 'Designer and artist', 890, 220, FALSE),
('tech_influencer', 'tech@example.com', '$2a$10$hash4', 'Tech Influencer', 'Technology trends and reviews', 150000, 500, TRUE);

INSERT INTO user_preferences (user_id, email_notifications, privacy_level, language) VALUES
(1, TRUE, 'public', 'en'),
(2, FALSE, 'public', 'en'),
(3, TRUE, 'friends', 'en'),
(4, TRUE, 'public', 'en');