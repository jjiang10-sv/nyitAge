package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gocql/gocql"
	"github.com/google/uuid"
)

// Post represents a user post
type Post struct {
	PostID          uuid.UUID `json:"post_id" cql:"post_id"`
	AuthorID        int64     `json:"author_id" cql:"author_id"`
	Content         string    `json:"content" cql:"content"`
	CreatedAt       time.Time `json:"created_at" cql:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" cql:"updated_at"`
	IsDeleted       bool      `json:"is_deleted" cql:"is_deleted"`
	EngagementScore float64   `json:"engagement_score" cql:"engagement_score"`
	LikeCount       int64     `json:"like_count" cql:"like_count"`
	CommentCount    int64     `json:"comment_count" cql:"comment_count"`
	ShareCount      int64     `json:"share_count" cql:"share_count"`
	Visibility      string    `json:"visibility" cql:"visibility"`
}

// CreatePostRequest represents post creation data
type CreatePostRequest struct {
	Content    string `json:"content" validate:"required,max=1024"`
	Visibility string `json:"visibility" validate:"oneof=public private followers"`
}

// UpdatePostRequest represents post update data
type UpdatePostRequest struct {
	Content    string `json:"content" validate:"max=1024"`
	Visibility string `json:"visibility" validate:"oneof=public private followers"`
}

// PostEngagement represents user engagement with a post
type PostEngagement struct {
	PostID         uuid.UUID `json:"post_id" cql:"post_id"`
	UserID         int64     `json:"user_id" cql:"user_id"`
	EngagementType string    `json:"engagement_type" cql:"engagement_type"`
	CreatedAt      time.Time `json:"created_at" cql:"created_at"`
	Metadata       string    `json:"metadata" cql:"metadata"`
}

// FanoutQueueItem represents an item in the fanout queue
type FanoutQueueItem struct {
	QueueID      uuid.UUID `json:"queue_id" cql:"queue_id"`
	PostID       uuid.UUID `json:"post_id" cql:"post_id"`
	AuthorID     int64     `json:"author_id" cql:"author_id"`
	CreatedAt    time.Time `json:"created_at" cql:"created_at"`
	ProcessedAt  time.Time `json:"processed_at" cql:"processed_at"`
	Status       string    `json:"status" cql:"status"`
	RetryCount   int       `json:"retry_count" cql:"retry_count"`
	ErrorMessage string    `json:"error_message" cql:"error_message"`
	Priority     int       `json:"priority" cql:"priority"`
}

// PostService handles post operations
type PostService struct {
	session     *gocql.Session
	redis       *redis.Client
	userService *UserProfileService
}

// NewPostService creates a new post service
func NewPostService(session *gocql.Session, redisClient *redis.Client, userService *UserProfileService) *PostService {
	return &PostService{
		session:     session,
		redis:       redisClient,
		userService: userService,
	}
}

// CreatePost creates a new post and triggers fanout
func (s *PostService) CreatePost(ctx context.Context, authorID int64, req *CreatePostRequest) (*Post, error) {
	// Validate content length
	if len(req.Content) > 1024 {
		return nil, fmt.Errorf("content exceeds maximum length of 1024 characters")
	}

	// Validate content (basic profanity/spam check)
	if err := s.validateContent(req.Content); err != nil {
		return nil, fmt.Errorf("content validation failed: %w", err)
	}

	// Get author info for validation
	author, err := s.userService.GetUserByID(ctx, authorID)
	if err != nil {
		return nil, fmt.Errorf("author not found: %w", err)
	}

	if !author.IsActive {
		return nil, fmt.Errorf("author account is not active")
	}

	// Create post
	post := &Post{
		PostID:          uuid.New(),
		AuthorID:        authorID,
		Content:         strings.TrimSpace(req.Content),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		IsDeleted:       false,
		EngagementScore: 0.0,
		LikeCount:       0,
		CommentCount:    0,
		ShareCount:      0,
		Visibility:      req.Visibility,
	}

	if post.Visibility == "" {
		post.Visibility = "public"
	}

	// Insert into posts table
	query := `INSERT INTO posts (post_id, author_id, content, created_at, updated_at, 
	                            is_deleted, engagement_score, like_count, comment_count, 
	                            share_count, visibility) 
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	err = s.session.Query(query,
		post.PostID, post.AuthorID, post.Content, post.CreatedAt, post.UpdatedAt,
		post.IsDeleted, post.EngagementScore, post.LikeCount, post.CommentCount,
		post.ShareCount, post.Visibility,
	).WithContext(ctx).Exec()

	if err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	// Insert into posts_by_author table
	authorQuery := `INSERT INTO posts_by_author (author_id, created_at, post_id, content, 
	                                            engagement_score, like_count, comment_count, 
	                                            share_count, is_deleted) 
	                VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	err = s.session.Query(authorQuery,
		post.AuthorID, post.CreatedAt, post.PostID, post.Content,
		post.EngagementScore, post.LikeCount, post.CommentCount,
		post.ShareCount, post.IsDeleted,
	).WithContext(ctx).Exec()

	if err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to insert into posts_by_author: %v\n", err)
	}

	// Cache the post
	s.cachePost(ctx, post)

	// Trigger fanout process
	err = s.enqueueFanout(ctx, post)
	if err != nil {
		// Log error but don't fail post creation
		fmt.Printf("Failed to enqueue fanout: %v\n", err)
	}

	// Update author's post count
	go s.updateAuthorPostCount(context.Background(), authorID)

	return post, nil
}

// GetPost retrieves a post by ID
func (s *PostService) GetPost(ctx context.Context, postID uuid.UUID) (*Post, error) {
	// Try cache first
	if post, err := s.getPostFromCache(ctx, postID); err == nil && post != nil {
		return post, nil
	}

	// Query database
	query := `SELECT post_id, author_id, content, created_at, updated_at, is_deleted,
	                 engagement_score, like_count, comment_count, share_count, visibility
	          FROM posts WHERE post_id = ? AND is_deleted = false`

	var post Post
	err := s.session.Query(query, postID).WithContext(ctx).Scan(
		&post.PostID, &post.AuthorID, &post.Content, &post.CreatedAt, &post.UpdatedAt,
		&post.IsDeleted, &post.EngagementScore, &post.LikeCount, &post.CommentCount,
		&post.ShareCount, &post.Visibility,
	)

	if err != nil {
		if err == gocql.ErrNotFound {
			return nil, fmt.Errorf("post not found")
		}
		return nil, fmt.Errorf("failed to get post: %w", err)
	}

	// Cache the result
	s.cachePost(ctx, &post)

	return &post, nil
}

// GetPostsByAuthor retrieves posts by author with pagination
func (s *PostService) GetPostsByAuthor(ctx context.Context, authorID int64, limit int, pageToken string) ([]*Post, string, error) {
	var query string
	var args []interface{}

	if pageToken == "" {
		// First page
		query = `SELECT author_id, created_at, post_id, content, engagement_score, 
		                like_count, comment_count, share_count, is_deleted
		         FROM posts_by_author 
		         WHERE author_id = ? AND is_deleted = false 
		         LIMIT ?`
		args = []interface{}{authorID, limit + 1} // +1 to check if there are more pages
	} else {
		// Parse page token (timestamp:post_id)
		parts := strings.Split(pageToken, ":")
		if len(parts) != 2 {
			return nil, "", fmt.Errorf("invalid page token")
		}

		timestamp, err := time.Parse(time.RFC3339, parts[0])
		if err != nil {
			return nil, "", fmt.Errorf("invalid page token timestamp")
		}

		postID, err := uuid.Parse(parts[1])
		if err != nil {
			return nil, "", fmt.Errorf("invalid page token post ID")
		}

		query = `SELECT author_id, created_at, post_id, content, engagement_score, 
		                like_count, comment_count, share_count, is_deleted
		         FROM posts_by_author 
		         WHERE author_id = ? AND is_deleted = false 
		         AND (created_at, post_id) < (?, ?)
		         LIMIT ?`
		args = []interface{}{authorID, timestamp, postID, limit + 1}
	}

	iter := s.session.Query(query, args...).WithContext(ctx).Iter()
	defer iter.Close()

	var posts []*Post
	var authorIDResult int64
	var createdAt time.Time
	var postID uuid.UUID
	var content string
	var engagementScore float64
	var likeCount, commentCount, shareCount int64
	var isDeleted bool

	for iter.Scan(&authorIDResult, &createdAt, &postID, &content, &engagementScore,
		&likeCount, &commentCount, &shareCount, &isDeleted) {

		post := &Post{
			PostID:          postID,
			AuthorID:        authorIDResult,
			Content:         content,
			CreatedAt:       createdAt,
			EngagementScore: engagementScore,
			LikeCount:       likeCount,
			CommentCount:    commentCount,
			ShareCount:      shareCount,
			IsDeleted:       isDeleted,
		}

		posts = append(posts, post)
	}

	if err := iter.Close(); err != nil {
		return nil, "", fmt.Errorf("failed to get posts by author: %w", err)
	}

	// Determine next page token
	var nextPageToken string
	if len(posts) > limit {
		// Remove the extra post and create next page token
		lastPost := posts[limit-1]
		posts = posts[:limit]
		nextPageToken = fmt.Sprintf("%s:%s",
			lastPost.CreatedAt.Format(time.RFC3339),
			lastPost.PostID.String())
	}

	return posts, nextPageToken, nil
}

// UpdatePost updates an existing post
func (s *PostService) UpdatePost(ctx context.Context, postID uuid.UUID, authorID int64, req *UpdatePostRequest) (*Post, error) {
	// Get existing post
	existingPost, err := s.GetPost(ctx, postID)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if existingPost.AuthorID != authorID {
		return nil, fmt.Errorf("unauthorized: user does not own this post")
	}

	// Validate content if provided
	if req.Content != "" {
		if len(req.Content) > 1024 {
			return nil, fmt.Errorf("content exceeds maximum length of 1024 characters")
		}
		if err := s.validateContent(req.Content); err != nil {
			return nil, fmt.Errorf("content validation failed: %w", err)
		}
		existingPost.Content = strings.TrimSpace(req.Content)
	}

	if req.Visibility != "" {
		existingPost.Visibility = req.Visibility
	}

	existingPost.UpdatedAt = time.Now()

	// Update in database
	query := `UPDATE posts SET content = ?, visibility = ?, updated_at = ? 
	          WHERE post_id = ?`

	err = s.session.Query(query, existingPost.Content, existingPost.Visibility,
		existingPost.UpdatedAt, postID).WithContext(ctx).Exec()

	if err != nil {
		return nil, fmt.Errorf("failed to update post: %w", err)
	}

	// Update cache
	s.cachePost(ctx, existingPost)
	s.invalidatePostCache(ctx, postID)

	return existingPost, nil
}

// DeletePost soft deletes a post
func (s *PostService) DeletePost(ctx context.Context, postID uuid.UUID, authorID int64) error {
	// Get existing post
	existingPost, err := s.GetPost(ctx, postID)
	if err != nil {
		return err
	}

	// Verify ownership
	if existingPost.AuthorID != authorID {
		return fmt.Errorf("unauthorized: user does not own this post")
	}

	// Soft delete
	query := `UPDATE posts SET is_deleted = true, updated_at = ? WHERE post_id = ?`
	err = s.session.Query(query, time.Now(), postID).WithContext(ctx).Exec()
	if err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}

	// Update posts_by_author
	authorQuery := `UPDATE posts_by_author SET is_deleted = true 
	                WHERE author_id = ? AND created_at = ? AND post_id = ?`
	s.session.Query(authorQuery, existingPost.AuthorID, existingPost.CreatedAt, postID).WithContext(ctx).Exec()

	// Remove from cache
	s.invalidatePostCache(ctx, postID)

	// Update author's post count
	go s.updateAuthorPostCount(context.Background(), authorID)

	return nil
}

// AddEngagement adds user engagement to a post
func (s *PostService) AddEngagement(ctx context.Context, postID uuid.UUID, userID int64, engagementType string, metadata string) error {
	// Validate engagement type
	validTypes := map[string]bool{
		"like": true, "comment": true, "share": true, "view": true,
	}
	if !validTypes[engagementType] {
		return fmt.Errorf("invalid engagement type: %s", engagementType)
	}

	// Insert engagement
	engagement := &PostEngagement{
		PostID:         postID,
		UserID:         userID,
		EngagementType: engagementType,
		CreatedAt:      time.Now(),
		Metadata:       metadata,
	}

	query := `INSERT INTO post_engagement (post_id, user_id, engagement_type, created_at, metadata) 
	          VALUES (?, ?, ?, ?, ?)`

	err := s.session.Query(query, engagement.PostID, engagement.UserID,
		engagement.EngagementType, engagement.CreatedAt, engagement.Metadata).WithContext(ctx).Exec()

	if err != nil {
		return fmt.Errorf("failed to add engagement: %w", err)
	}

	// Update engagement counts asynchronously
	go s.updateEngagementCounts(context.Background(), postID, engagementType, 1)

	return nil
}

// enqueueFanout adds a post to the fanout queue
func (s *PostService) enqueueFanout(ctx context.Context, post *Post) error {
	queueItem := &FanoutQueueItem{
		QueueID:    uuid.New(),
		PostID:     post.PostID,
		AuthorID:   post.AuthorID,
		CreatedAt:  time.Now(),
		Status:     "pending",
		RetryCount: 0,
		Priority:   0, // Normal priority
	}

	// Insert into fanout_queue
	query := `INSERT INTO fanout_queue (queue_id, post_id, author_id, created_at, status, retry_count, priority) 
	          VALUES (?, ?, ?, ?, ?, ?, ?)`

	err := s.session.Query(query, queueItem.QueueID, queueItem.PostID, queueItem.AuthorID,
		queueItem.CreatedAt, queueItem.Status, queueItem.RetryCount, queueItem.Priority).WithContext(ctx).Exec()

	if err != nil {
		return fmt.Errorf("failed to enqueue fanout: %w", err)
	}

	// Insert into fanout_queue_by_status for efficient polling
	statusQuery := `INSERT INTO fanout_queue_by_status (status, created_at, queue_id, post_id, author_id, priority, retry_count) 
	                VALUES (?, ?, ?, ?, ?, ?, ?)`

	err = s.session.Query(statusQuery, queueItem.Status, queueItem.CreatedAt, queueItem.QueueID,
		queueItem.PostID, queueItem.AuthorID, queueItem.Priority, queueItem.RetryCount).WithContext(ctx).Exec()

	if err != nil {
		// Log error but don't fail
		fmt.Printf("Failed to insert into fanout_queue_by_status: %v\n", err)
	}

	return nil
}

// validateContent performs basic content validation
func (s *PostService) validateContent(content string) error {
	content = strings.TrimSpace(content)

	if content == "" {
		return fmt.Errorf("content cannot be empty")
	}

	// Basic profanity check (in production, use a proper service)
	bannedWords := []string{"spam", "scam", "fake"}
	contentLower := strings.ToLower(content)

	for _, word := range bannedWords {
		if strings.Contains(contentLower, word) {
			return fmt.Errorf("content contains prohibited words")
		}
	}

	return nil
}

// Cache operations
func (s *PostService) cachePost(ctx context.Context, post *Post) {
	key := fmt.Sprintf("post:%s", post.PostID.String())
	s.redis.HSet(ctx, key,
		"post_id", post.PostID.String(),
		"author_id", post.AuthorID,
		"content", post.Content,
		"created_at", post.CreatedAt.Format(time.RFC3339),
		"engagement_score", post.EngagementScore,
		"like_count", post.LikeCount,
		"visibility", post.Visibility,
	)
	s.redis.Expire(ctx, key, 1*time.Hour)
}

func (s *PostService) getPostFromCache(ctx context.Context, postID uuid.UUID) (*Post, error) {
	key := fmt.Sprintf("post:%s", postID.String())
	result := s.redis.HGetAll(ctx, key)
	if result.Err() != nil || len(result.Val()) == 0 {
		return nil, fmt.Errorf("post not in cache")
	}

	data := result.Val()
	createdAt, _ := time.Parse(time.RFC3339, data["created_at"])

	post := &Post{
		PostID:          postID,
		AuthorID:        parseInt64(data["author_id"]),
		Content:         data["content"],
		CreatedAt:       createdAt,
		EngagementScore: parseFloat64(data["engagement_score"]),
		LikeCount:       parseInt64(data["like_count"]),
		Visibility:      data["visibility"],
	}

	return post, nil
}

func (s *PostService) invalidatePostCache(ctx context.Context, postID uuid.UUID) {
	key := fmt.Sprintf("post:%s", postID.String())
	s.redis.Del(ctx, key)
}

// updateEngagementCounts updates post engagement counts
func (s *PostService) updateEngagementCounts(ctx context.Context, postID uuid.UUID, engagementType string, delta int64) {
	var column string
	switch engagementType {
	case "like":
		column = "like_count"
	case "comment":
		column = "comment_count"
	case "share":
		column = "share_count"
	default:
		return
	}

	// Update posts table
	query := fmt.Sprintf("UPDATE posts SET %s = %s + ? WHERE post_id = ?", column, column)
	s.session.Query(query, delta, postID).WithContext(ctx).Exec()

	// Update posts_by_author table
	// Note: This requires getting the post first to get author_id and created_at
	post, err := s.GetPost(ctx, postID)
	if err == nil {
		authorQuery := fmt.Sprintf("UPDATE posts_by_author SET %s = %s + ? WHERE author_id = ? AND created_at = ? AND post_id = ?", column, column)
		s.session.Query(authorQuery, delta, post.AuthorID, post.CreatedAt, postID).WithContext(ctx).Exec()
	}

	// Update cache
	s.invalidatePostCache(ctx, postID)
}

// updateAuthorPostCount updates the author's post count
func (s *PostService) updateAuthorPostCount(ctx context.Context, authorID int64) {
	// Count active posts
	query := `SELECT COUNT(*) FROM posts_by_author WHERE author_id = ? AND is_deleted = false`
	var count int64
	err := s.session.Query(query, authorID).WithContext(ctx).Scan(&count)
	if err != nil {
		fmt.Printf("Failed to count posts for author %d: %v\n", authorID, err)
		return
	}

	// Update user profile (this would call the user service)
	// s.userService.UpdatePostCount(ctx, authorID, count)
}

// Helper functions
func parseFloat64(s string) float64 {
	// In real implementation, use strconv.ParseFloat with error handling
	return 0.0
}
