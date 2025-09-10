package services

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gocql/gocql"
	"github.com/google/uuid"
)

// FeedItem represents an item in a user's feed
type FeedItem struct {
	PostID          uuid.UUID `json:"post_id"`
	AuthorID        int64     `json:"author_id"`
	AuthorUsername  string    `json:"author_username"`
	Content         string    `json:"content"`
	CreatedAt       time.Time `json:"created_at"`
	EngagementScore float64   `json:"engagement_score"`
	LikeCount       int64     `json:"like_count"`
	CommentCount    int64     `json:"comment_count"`
	ShareCount      int64     `json:"share_count"`
	IsLiked         bool      `json:"is_liked,omitempty"`
	RankingScore    float64   `json:"ranking_score,omitempty"`
}

// FeedResponse represents a paginated feed response
type FeedResponse struct {
	Items         []*FeedItem `json:"items"`
	NextPageToken string      `json:"next_page_token,omitempty"`
	HasMore       bool        `json:"has_more"`
	GeneratedAt   time.Time   `json:"generated_at"`
	Source        string      `json:"source"` // "cache", "push", "pull", "hybrid"
}

// FeedCursor represents pagination cursor for feeds
type FeedCursor struct {
	UserID     int64     `json:"user_id"`
	LastReadAt time.Time `json:"last_read_at"`
	LastPostID uuid.UUID `json:"last_post_id"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// FeedService handles feed generation and management
type FeedService struct {
	session            *gocql.Session
	redis              *redis.Client
	postService        *PostService
	socialService      *SocialGraphService
	userService        *UserProfileService
	celebrityThreshold int64
}

// NewFeedService creates a new feed service
func NewFeedService(session *gocql.Session, redisClient *redis.Client,
	postService *PostService, socialService *SocialGraphService,
	userService *UserProfileService) *FeedService {
	return &FeedService{
		session:            session,
		redis:              redisClient,
		postService:        postService,
		socialService:      socialService,
		userService:        userService,
		celebrityThreshold: 100000,
	}
}

// GetUserFeed generates a personalized feed for a user
func (s *FeedService) GetUserFeed(ctx context.Context, userID int64, limit int, pageToken string) (*FeedResponse, error) {
	if limit <= 0 || limit > 50 {
		limit = 10 // Default to top 10 as per requirement
	}

	// Try to get feed from cache first
	if pageToken == "" {
		if cachedFeed, err := s.getFeedFromCache(ctx, userID, limit); err == nil {
			return cachedFeed, nil
		}
	}

	// Get user's following list to determine feed strategy
	following, err := s.socialService.GetFollowing(ctx, userID, 1000, 0) // Get up to 1000 following
	if err != nil {
		return nil, fmt.Errorf("failed to get following list: %w", err)
	}

	// Separate celebrities from regular users
	var celebrities, regularUsers []*UserConnection
	for _, user := range following {
		if user.IsCelebrity {
			celebrities = append(celebrities, user)
		} else {
			regularUsers = append(regularUsers, user)
		}
	}

	// Generate hybrid feed
	feedItems, nextToken, err := s.generateHybridFeed(ctx, userID, celebrities, regularUsers, limit, pageToken)
	if err != nil {
		return nil, fmt.Errorf("failed to generate feed: %w", err)
	}

	// Rank and personalize the feed
	rankedItems := s.rankFeedItems(ctx, userID, feedItems)

	// Ensure we don't exceed the limit
	if len(rankedItems) > limit {
		rankedItems = rankedItems[:limit]
	}

	response := &FeedResponse{
		Items:         rankedItems,
		NextPageToken: nextToken,
		HasMore:       nextToken != "",
		GeneratedAt:   time.Now(),
		Source:        "hybrid",
	}

	// Cache the first page
	if pageToken == "" {
		s.cacheFeed(ctx, userID, response)
	}

	return response, nil
}

// generateHybridFeed combines push-based (precomputed) and pull-based (on-demand) feeds
func (s *FeedService) generateHybridFeed(ctx context.Context, userID int64, celebrities, regularUsers []*UserConnection,
	limit int, pageToken string) ([]*FeedItem, string, error) {

	var allItems []*FeedItem
	var nextPageToken string

	// 1. Get precomputed timeline (push-based for regular users)
	pushItems, pushToken, err := s.getPushBasedFeed(ctx, userID, limit*2, pageToken) // Get more for better ranking
	if err != nil {
		fmt.Printf("Failed to get push-based feed: %v\n", err)
	} else {
		allItems = append(allItems, pushItems...)
		if pushToken != "" {
			nextPageToken = pushToken
		}
	}

	// 2. Get celebrity posts (pull-based)
	if len(celebrities) > 0 {
		pullItems, err := s.getPullBasedFeed(ctx, userID, celebrities, limit)
		if err != nil {
			fmt.Printf("Failed to get pull-based feed: %v\n", err)
		} else {
			allItems = append(allItems, pullItems...)
		}
	}

	// 3. Merge and sort by creation time
	sort.Slice(allItems, func(i, j int) bool {
		return allItems[i].CreatedAt.After(allItems[j].CreatedAt)
	})

	// 4. Deduplicate posts
	allItems = s.deduplicateFeedItems(allItems)

	return allItems, nextPageToken, nil
}

// getPushBasedFeed retrieves precomputed timeline items
func (s *FeedService) getPushBasedFeed(ctx context.Context, userID int64, limit int, pageToken string) ([]*FeedItem, string, error) {
	var query string
	var args []interface{}

	if pageToken == "" {
		query = `SELECT user_id, created_at, post_id, author_id, content, engagement_score, inserted_at
		         FROM user_timeline 
		         WHERE user_id = ? 
		         ORDER BY created_at DESC, post_id ASC
		         LIMIT ?`
		args = []interface{}{userID, limit + 1}
	} else {
		// Parse page token
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

		query = `SELECT user_id, created_at, post_id, author_id, content, engagement_score, inserted_at
		         FROM user_timeline 
		         WHERE user_id = ? AND (created_at, post_id) < (?, ?)
		         ORDER BY created_at DESC, post_id ASC
		         LIMIT ?`
		args = []interface{}{userID, timestamp, postID, limit + 1}
	}

	iter := s.session.Query(query, args...).WithContext(ctx).Iter()
	defer iter.Close()

	var items []*FeedItem
	var userIDResult int64
	var createdAt time.Time
	var postID uuid.UUID
	var authorID int64
	var content string
	var engagementScore float64
	var insertedAt time.Time

	for iter.Scan(&userIDResult, &createdAt, &postID, &authorID, &content, &engagementScore, &insertedAt) {
		// Get author info (cached)
		author, err := s.userService.GetUserByID(ctx, authorID)
		if err != nil {
			continue // Skip if author not found
		}

		item := &FeedItem{
			PostID:          postID,
			AuthorID:        authorID,
			AuthorUsername:  author.Username,
			Content:         content,
			CreatedAt:       createdAt,
			EngagementScore: engagementScore,
		}

		items = append(items, item)
	}

	if err := iter.Close(); err != nil {
		return nil, "", fmt.Errorf("failed to get push-based feed: %w", err)
	}

	// Determine next page token
	var nextPageToken string
	if len(items) > limit {
		lastItem := items[limit-1]
		items = items[:limit]
		nextPageToken = fmt.Sprintf("%s:%s",
			lastItem.CreatedAt.Format(time.RFC3339),
			lastItem.PostID.String())
	}

	return items, nextPageToken, nil
}

// getPullBasedFeed retrieves recent posts from celebrities on-demand
func (s *FeedService) getPullBasedFeed(ctx context.Context, userID int64, celebrities []*UserConnection, limit int) ([]*FeedItem, error) {
	var allItems []*FeedItem

	// Get recent posts from each celebrity
	for _, celebrity := range celebrities {
		posts, _, err := s.postService.GetPostsByAuthor(ctx, celebrity.UserID, limit/len(celebrities)+1, "")
		if err != nil {
			continue // Skip if error getting posts
		}

		for _, post := range posts {
			// Only include recent posts (last 24 hours for celebrities)
			if time.Since(post.CreatedAt) > 24*time.Hour {
				continue
			}

			item := &FeedItem{
				PostID:          post.PostID,
				AuthorID:        post.AuthorID,
				AuthorUsername:  celebrity.Username,
				Content:         post.Content,
				CreatedAt:       post.CreatedAt,
				EngagementScore: post.EngagementScore,
				LikeCount:       post.LikeCount,
				CommentCount:    post.CommentCount,
				ShareCount:      post.ShareCount,
			}

			allItems = append(allItems, item)
		}
	}

	// Sort by creation time (newest first)
	sort.Slice(allItems, func(i, j int) bool {
		return allItems[i].CreatedAt.After(allItems[j].CreatedAt)
	})

	// Limit results
	if len(allItems) > limit {
		allItems = allItems[:limit]
	}

	return allItems, nil
}

// rankFeedItems applies ranking algorithm to feed items
func (s *FeedService) rankFeedItems(ctx context.Context, userID int64, items []*FeedItem) []*FeedItem {
	now := time.Now()

	for _, item := range items {
		// Calculate ranking score based on multiple factors
		score := 0.0

		// 1. Recency score (decay over time)
		ageHours := now.Sub(item.CreatedAt).Hours()
		recencyScore := 1.0 / (1.0 + ageHours/24.0) // Decay over days

		// 2. Engagement score (normalized)
		engagementScore := item.EngagementScore / 100.0 // Normalize to 0-1 range

		// 3. Social proof (likes, comments, shares)
		socialScore := float64(item.LikeCount+item.CommentCount*2+item.ShareCount*3) / 100.0

		// 4. Author affinity (simplified - in production, use ML model)
		affinityScore := s.getUserAffinityScore(ctx, userID, item.AuthorID)

		// 5. Content quality (simplified - length, readability, etc.)
		qualityScore := s.getContentQualityScore(item.Content)

		// Weighted combination
		score = recencyScore*0.3 + engagementScore*0.2 + socialScore*0.2 + affinityScore*0.2 + qualityScore*0.1

		item.RankingScore = score
	}

	// Sort by ranking score (highest first)
	sort.Slice(items, func(i, j int) bool {
		return items[i].RankingScore > items[j].RankingScore
	})

	return items
}

// getUserAffinityScore calculates user affinity based on interaction history
func (s *FeedService) getUserAffinityScore(ctx context.Context, userID, authorID int64) float64 {
	// Check Redis cache for interaction history
	key := fmt.Sprintf("affinity:%d:%d", userID, authorID)
	result := s.redis.Get(ctx, key)
	if result.Err() == nil {
		if score, err := strconv.ParseFloat(result.Val(), 64); err == nil {
			return score
		}
	}

	// Calculate affinity based on:
	// - Recent interactions (likes, comments, shares)
	// - Frequency of interactions
	// - Mutual connections

	// Simplified calculation (in production, use ML model)
	baseScore := 0.5 // Default affinity

	// Check if users follow each other (mutual follow = higher affinity)
	isFollowing, _ := s.socialService.IsFollowing(ctx, userID, authorID)
	isFollower, _ := s.socialService.IsFollowing(ctx, authorID, userID)

	if isFollowing && isFollower {
		baseScore += 0.3 // Mutual follow bonus
	} else if isFollowing {
		baseScore += 0.2 // Following bonus
	}

	// Cache the result
	s.redis.Set(ctx, key, baseScore, 1*time.Hour)

	return baseScore
}

// getContentQualityScore evaluates content quality
func (s *FeedService) getContentQualityScore(content string) float64 {
	score := 0.5 // Base score

	// Length factor (not too short, not too long)
	length := len(content)
	if length >= 50 && length <= 500 {
		score += 0.2
	} else if length < 20 {
		score -= 0.1
	}

	// Check for URLs (might be more engaging)
	if strings.Contains(content, "http") {
		score += 0.1
	}

	// Check for hashtags (might be more discoverable)
	if strings.Contains(content, "#") {
		score += 0.1
	}

	// Check for mentions (social interaction)
	if strings.Contains(content, "@") {
		score += 0.1
	}

	// Ensure score is between 0 and 1
	if score > 1.0 {
		score = 1.0
	}
	if score < 0.0 {
		score = 0.0
	}

	return score
}

// deduplicateFeedItems removes duplicate posts from feed
func (s *FeedService) deduplicateFeedItems(items []*FeedItem) []*FeedItem {
	seen := make(map[uuid.UUID]bool)
	var deduplicated []*FeedItem

	for _, item := range items {
		if !seen[item.PostID] {
			seen[item.PostID] = true
			deduplicated = append(deduplicated, item)
		}
	}

	return deduplicated
}

// RefreshUserFeed forces a refresh of user's feed
func (s *FeedService) RefreshUserFeed(ctx context.Context, userID int64) error {
	// Invalidate cache
	s.invalidateFeedCache(ctx, userID)

	// Optionally trigger background refresh
	go func() {
		_, err := s.GetUserFeed(context.Background(), userID, 10, "")
		if err != nil {
			fmt.Printf("Failed to refresh feed for user %d: %v\n", userID, err)
		}
	}()

	return nil
}

// UpdateFeedCursor updates user's feed reading position
func (s *FeedService) UpdateFeedCursor(ctx context.Context, userID int64, lastPostID uuid.UUID) error {
	cursor := &FeedCursor{
		UserID:     userID,
		LastReadAt: time.Now(),
		LastPostID: lastPostID,
		UpdatedAt:  time.Now(),
	}

	query := `INSERT INTO user_feed_cursors (user_id, last_read_at, last_post_id, updated_at) 
	          VALUES (?, ?, ?, ?) 
	          USING TTL 2592000` // 30 days TTL

	err := s.session.Query(query, cursor.UserID, cursor.LastReadAt,
		cursor.LastPostID, cursor.UpdatedAt).WithContext(ctx).Exec()

	if err != nil {
		return fmt.Errorf("failed to update feed cursor: %w", err)
	}

	// Cache the cursor
	key := fmt.Sprintf("cursor:%d", userID)
	s.redis.HSet(ctx, key,
		"last_read_at", cursor.LastReadAt.Format(time.RFC3339),
		"last_post_id", cursor.LastPostID.String(),
	)
	s.redis.Expire(ctx, key, 24*time.Hour)

	return nil
}

// Cache operations
func (s *FeedService) cacheFeed(ctx context.Context, userID int64, feed *FeedResponse) {
	key := fmt.Sprintf("feed:%d", userID)

	// Cache feed items as JSON (simplified)
	// In production, use proper serialization
	s.redis.Set(ctx, key, fmt.Sprintf("%d_items", len(feed.Items)), 5*time.Minute)
}

func (s *FeedService) getFeedFromCache(ctx context.Context, userID int64, limit int) (*FeedResponse, error) {
	key := fmt.Sprintf("feed:%d", userID)
	result := s.redis.Get(ctx, key)
	if result.Err() != nil {
		return nil, fmt.Errorf("feed not in cache")
	}

	// In production, deserialize proper feed data
	// This is a simplified implementation
	return nil, fmt.Errorf("cache miss")
}

func (s *FeedService) invalidateFeedCache(ctx context.Context, userID int64) {
	key := fmt.Sprintf("feed:%d", userID)
	s.redis.Del(ctx, key)
}

// GetTrendingPosts returns trending posts for discovery
func (s *FeedService) GetTrendingPosts(ctx context.Context, limit int, timeWindow time.Duration) ([]*FeedItem, error) {
	// Get posts from the last time window with high engagement
	since := time.Now().Add(-timeWindow)

	query := `SELECT post_id, author_id, content, created_at, engagement_score, 
	                 like_count, comment_count, share_count
	          FROM posts 
	          WHERE created_at >= ? AND is_deleted = false 
	          ORDER BY engagement_score DESC 
	          LIMIT ?`

	iter := s.session.Query(query, since, limit).WithContext(ctx).Iter()
	defer iter.Close()

	var items []*FeedItem
	var postID uuid.UUID
	var authorID int64
	var content string
	var createdAt time.Time
	var engagementScore float64
	var likeCount, commentCount, shareCount int64

	for iter.Scan(&postID, &authorID, &content, &createdAt, &engagementScore,
		&likeCount, &commentCount, &shareCount) {

		// Get author info
		author, err := s.userService.GetUserByID(ctx, authorID)
		if err != nil {
			continue
		}

		item := &FeedItem{
			PostID:          postID,
			AuthorID:        authorID,
			AuthorUsername:  author.Username,
			Content:         content,
			CreatedAt:       createdAt,
			EngagementScore: engagementScore,
			LikeCount:       likeCount,
			CommentCount:    commentCount,
			ShareCount:      shareCount,
		}

		items = append(items, item)
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("failed to get trending posts: %w", err)
	}

	return items, nil
}
