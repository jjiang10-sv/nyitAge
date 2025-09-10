-- Cassandra Schema for Posts and Timeline Storage

-- Posts table (main content storage)
CREATE TABLE IF NOT EXISTS posts (
    post_id UUID PRIMARY KEY,
    author_id BIGINT,
    content TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    is_deleted BOOLEAN DEFAULT FALSE,
    engagement_score DOUBLE DEFAULT 0.0,
    like_count BIGINT DEFAULT 0,
    comment_count BIGINT DEFAULT 0,
    share_count BIGINT DEFAULT 0,
    visibility VARCHAR DEFAULT 'public'  -- public, private, followers
);

-- Posts by author (for author timeline/outbox)
CREATE TABLE IF NOT EXISTS posts_by_author (
    author_id BIGINT,
    created_at TIMESTAMP,
    post_id UUID,
    content TEXT,
    engagement_score DOUBLE,
    like_count BIGINT,
    comment_count BIGINT,
    share_count BIGINT,
    is_deleted BOOLEAN,
    PRIMARY KEY (author_id, created_at, post_id)
) WITH CLUSTERING ORDER BY (created_at DESC, post_id ASC);

-- User timeline (precomputed feed for push-based delivery)
CREATE TABLE IF NOT EXISTS user_timeline (
    user_id BIGINT,
    created_at TIMESTAMP,
    post_id UUID,
    author_id BIGINT,
    content TEXT,
    engagement_score DOUBLE,
    inserted_at TIMESTAMP,
    PRIMARY KEY (user_id, created_at, post_id)
) WITH CLUSTERING ORDER BY (created_at DESC, post_id ASC);

-- Celebrity posts (for pull-based feed generation)
CREATE TABLE IF NOT EXISTS celebrity_posts (
    author_id BIGINT,
    created_at TIMESTAMP,
    post_id UUID,
    content TEXT,
    engagement_score DOUBLE,
    like_count BIGINT,
    comment_count BIGINT,
    share_count BIGINT,
    PRIMARY KEY (author_id, created_at, post_id)
) WITH CLUSTERING ORDER BY (created_at DESC, post_id ASC);

-- Fanout queue (for processing post distribution)
CREATE TABLE IF NOT EXISTS fanout_queue (
    queue_id UUID PRIMARY KEY,
    post_id UUID,
    author_id BIGINT,
    created_at TIMESTAMP,
    processed_at TIMESTAMP,
    status TEXT DEFAULT 'pending',  -- pending, processing, completed, failed
    retry_count INT DEFAULT 0,
    error_message TEXT,
    priority INT DEFAULT 0  -- higher number = higher priority
);

-- Fanout queue by status (for efficient worker polling)
CREATE TABLE IF NOT EXISTS fanout_queue_by_status (
    status TEXT,
    created_at TIMESTAMP,
    queue_id UUID,
    post_id UUID,
    author_id BIGINT,
    priority INT,
    retry_count INT,
    PRIMARY KEY (status, created_at, queue_id)
) WITH CLUSTERING ORDER BY (created_at ASC, queue_id ASC);

-- Post engagement tracking
CREATE TABLE IF NOT EXISTS post_engagement (
    post_id UUID,
    user_id BIGINT,
    engagement_type TEXT,  -- like, comment, share, view
    created_at TIMESTAMP,
    metadata TEXT,  -- JSON for additional data
    PRIMARY KEY (post_id, user_id, engagement_type)
);

-- User feed cursor tracking (for pagination)
CREATE TABLE IF NOT EXISTS user_feed_cursors (
    user_id BIGINT PRIMARY KEY,
    last_read_at TIMESTAMP,
    last_post_id UUID,
    updated_at TIMESTAMP
);

-- Sample data
INSERT INTO posts (post_id, author_id, content, created_at, engagement_score, like_count) VALUES
(uuid(), 1, 'Hello world! This is my first post.', '2024-01-01 10:00:00', 15.5, 10);

INSERT INTO posts (post_id, author_id, content, created_at, engagement_score, like_count) VALUES
(uuid(), 2, 'Celebrity announcement: New project coming soon!', '2024-01-01 11:00:00', 1500.0, 50000);

INSERT INTO posts (post_id, author_id, content, created_at, engagement_score, like_count) VALUES
(uuid(), 3, 'Beautiful sunset today 🌅', '2024-01-01 12:00:00', 25.3, 20);

INSERT INTO posts (post_id, author_id, content, created_at, engagement_score, like_count) VALUES
(uuid(), 4, 'Tech tip: Always validate your inputs!', '2024-01-01 13:00:00', 85.7, 150);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS posts_author_created_idx ON posts (author_id, created_at);
CREATE INDEX IF NOT EXISTS posts_created_idx ON posts (created_at);
CREATE INDEX IF NOT EXISTS fanout_queue_status_idx ON fanout_queue (status, created_at);
CREATE INDEX IF NOT EXISTS post_engagement_user_idx ON post_engagement (user_id, created_at);