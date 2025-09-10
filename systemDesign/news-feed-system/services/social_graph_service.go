package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

// FollowRelationship represents a follow relationship
type FollowRelationship struct {
	FollowerID          int64     `json:"follower_id"`
	FolloweeID          int64     `json:"followee_id"`
	Since               time.Time `json:"since"`
	IsActive            bool      `json:"is_active"`
	NotificationEnabled bool      `json:"notification_enabled"`
}

// UserConnection represents a user in the social graph
type UserConnection struct {
	UserID      int64  `json:"user_id"`
	Username    string `json:"username"`
	IsCelebrity bool   `json:"is_celebrity"`
	IsFollowing bool   `json:"is_following,omitempty"`
	IsFollower  bool   `json:"is_follower,omitempty"`
}

// SocialStats represents social statistics for a user
type SocialStats struct {
	UserID         int64 `json:"user_id"`
	FollowersCount int64 `json:"followers_count"`
	FollowingCount int64 `json:"following_count"`
	MutualCount    int64 `json:"mutual_count,omitempty"`
}

// SocialGraphService handles social graph operations
type SocialGraphService struct {
	neo4j              neo4j.Driver
	redis              *redis.Client
	userService        *UserProfileService
	celebrityThreshold int64
}

// NewSocialGraphService creates a new social graph service
func NewSocialGraphService(neo4jDriver neo4j.Driver, redisClient *redis.Client, userService *UserProfileService) *SocialGraphService {
	return &SocialGraphService{
		neo4j:              neo4jDriver,
		redis:              redisClient,
		userService:        userService,
		celebrityThreshold: 100000, // 100K followers = celebrity
	}
}

// FollowUser creates a follow relationship
func (s *SocialGraphService) FollowUser(ctx context.Context, followerID, followeeID int64) error {
	if followerID == followeeID {
		return fmt.Errorf("users cannot follow themselves")
	}

	// Check if relationship already exists
	exists, err := s.IsFollowing(ctx, followerID, followeeID)
	if err != nil {
		return fmt.Errorf("failed to check existing relationship: %w", err)
	}
	if exists {
		return fmt.Errorf("already following this user")
	}

	// Create relationship in Neo4j
	session := s.neo4j.NewSession(neo4j.SessionConfig{})
	defer session.Close()

	_, err = session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		query := `
			MATCH (follower:User {user_id: $follower_id}), (followee:User {user_id: $followee_id})
			CREATE (follower)-[:FOLLOWS {
				since: datetime(),
				is_active: true,
				notification_enabled: true
			}]->(followee)
			RETURN follower.user_id, followee.user_id`

		result, err := tx.Run(query, map[string]interface{}{
			"follower_id": followerID,
			"followee_id": followeeID,
		})
		if err != nil {
			return nil, err
		}

		if !result.Next() {
			return nil, fmt.Errorf("failed to create follow relationship")
		}

		return result.Record(), nil
	})

	if err != nil {
		return fmt.Errorf("failed to create follow relationship: %w", err)
	}

	// Update Redis cache
	now := time.Now().Unix()
	pipe := s.redis.Pipeline()

	// Add to follower's following list
	pipe.ZAdd(ctx, fmt.Sprintf("following:%d", followerID), &redis.Z{
		Score:  float64(now),
		Member: followeeID,
	})

	// Add to followee's followers list
	pipe.ZAdd(ctx, fmt.Sprintf("followers:%d", followeeID), &redis.Z{
		Score:  float64(now),
		Member: followerID,
	})

	// Update relationship cache
	pipe.HSet(ctx, fmt.Sprintf("follows:%d", followerID), followeeID, 1)

	// Increment counts
	pipe.Incr(ctx, fmt.Sprintf("following_count:%d", followerID))
	pipe.Incr(ctx, fmt.Sprintf("follower_count:%d", followeeID))

	_, err = pipe.Exec(ctx)
	if err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to update Redis cache: %v\n", err)
	}

	// Update counts in user profile service (async)
	go s.updateUserCounts(context.Background(), followerID, followeeID)

	return nil
}

// UnfollowUser removes a follow relationship
func (s *SocialGraphService) UnfollowUser(ctx context.Context, followerID, followeeID int64) error {
	// Check if relationship exists
	exists, err := s.IsFollowing(ctx, followerID, followeeID)
	if err != nil {
		return fmt.Errorf("failed to check existing relationship: %w", err)
	}
	if !exists {
		return fmt.Errorf("not following this user")
	}

	// Remove relationship from Neo4j
	session := s.neo4j.NewSession(neo4j.SessionConfig{})
	defer session.Close()

	_, err = session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		query := `
			MATCH (follower:User {user_id: $follower_id})-[r:FOLLOWS]->(followee:User {user_id: $followee_id})
			DELETE r
			RETURN count(r) as deleted`

		result, err := tx.Run(query, map[string]interface{}{
			"follower_id": followerID,
			"followee_id": followeeID,
		})
		if err != nil {
			return nil, err
		}

		if result.Next() {
			deleted, _ := result.Record().Get("deleted")
			if deleted.(int64) == 0 {
				return nil, fmt.Errorf("relationship not found")
			}
		}

		return nil, nil
	})

	if err != nil {
		return fmt.Errorf("failed to remove follow relationship: %w", err)
	}

	// Update Redis cache
	pipe := s.redis.Pipeline()

	// Remove from follower's following list
	pipe.ZRem(ctx, fmt.Sprintf("following:%d", followerID), followeeID)

	// Remove from followee's followers list
	pipe.ZRem(ctx, fmt.Sprintf("followers:%d", followeeID), followerID)

	// Update relationship cache
	pipe.HDel(ctx, fmt.Sprintf("follows:%d", followerID), strconv.FormatInt(followeeID, 10))

	// Decrement counts
	pipe.Decr(ctx, fmt.Sprintf("following_count:%d", followerID))
	pipe.Decr(ctx, fmt.Sprintf("follower_count:%d", followeeID))

	_, err = pipe.Exec(ctx)
	if err != nil {
		fmt.Printf("Failed to update Redis cache: %v\n", err)
	}

	// Update counts in user profile service (async)
	go s.updateUserCounts(context.Background(), followerID, followeeID)

	return nil
}

// IsFollowing checks if follower follows followee
func (s *SocialGraphService) IsFollowing(ctx context.Context, followerID, followeeID int64) (bool, error) {
	// Check Redis cache first
	result := s.redis.HGet(ctx, fmt.Sprintf("follows:%d", followerID), strconv.FormatInt(followeeID, 10))
	if result.Err() == nil {
		return result.Val() == "1", nil
	}

	// Query Neo4j
	session := s.neo4j.NewSession(neo4j.SessionConfig{})
	defer session.Close()

	result2, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		query := `
			MATCH (follower:User {user_id: $follower_id})-[r:FOLLOWS]->(followee:User {user_id: $followee_id})
			RETURN r IS NOT NULL as is_following`

		result, err := tx.Run(query, map[string]interface{}{
			"follower_id": followerID,
			"followee_id": followeeID,
		})
		if err != nil {
			return false, err
		}

		if result.Next() {
			isFollowing, _ := result.Record().Get("is_following")
			return isFollowing.(bool), nil
		}

		return false, nil
	})

	if err != nil {
		return false, fmt.Errorf("failed to check follow relationship: %w", err)
	}

	isFollowing := result2.(bool)

	// Cache the result
	s.redis.HSet(ctx, fmt.Sprintf("follows:%d", followerID), followeeID, map[bool]int{true: 1, false: 0}[isFollowing])

	return isFollowing, nil
}

// GetFollowers returns a list of users following the given user
func (s *SocialGraphService) GetFollowers(ctx context.Context, userID int64, limit, offset int) ([]*UserConnection, error) {
	// Try Redis cache first for recent followers
	if offset == 0 && limit <= 100 {
		followers, err := s.getFollowersFromCache(ctx, userID, limit)
		if err == nil && len(followers) > 0 {
			return followers, nil
		}
	}

	// Query Neo4j
	session := s.neo4j.NewSession(neo4j.SessionConfig{})
	defer session.Close()

	result, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		query := `
			MATCH (follower:User)-[:FOLLOWS]->(user:User {user_id: $user_id})
			RETURN follower.user_id as user_id, follower.username as username, follower.is_celebrity as is_celebrity
			ORDER BY follower.username
			SKIP $offset LIMIT $limit`

		result, err := tx.Run(query, map[string]interface{}{
			"user_id": userID,
			"offset":  offset,
			"limit":   limit,
		})
		if err != nil {
			return nil, err
		}

		var followers []*UserConnection
		for result.Next() {
			record := result.Record()
			followerID, _ := record.Get("user_id")
			username, _ := record.Get("username")
			isCelebrity, _ := record.Get("is_celebrity")

			followers = append(followers, &UserConnection{
				UserID:      followerID.(int64),
				Username:    username.(string),
				IsCelebrity: isCelebrity.(bool),
				IsFollower:  true,
			})
		}

		return followers, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get followers: %w", err)
	}

	return result.([]*UserConnection), nil
}

// GetFollowing returns a list of users that the given user follows
func (s *SocialGraphService) GetFollowing(ctx context.Context, userID int64, limit, offset int) ([]*UserConnection, error) {
	// Try Redis cache first
	if offset == 0 && limit <= 100 {
		following, err := s.getFollowingFromCache(ctx, userID, limit)
		if err == nil && len(following) > 0 {
			return following, nil
		}
	}

	// Query Neo4j
	session := s.neo4j.NewSession(neo4j.SessionConfig{})
	defer session.Close()

	result, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		query := `
			MATCH (user:User {user_id: $user_id})-[:FOLLOWS]->(following:User)
			RETURN following.user_id as user_id, following.username as username, following.is_celebrity as is_celebrity
			ORDER BY following.username
			SKIP $offset LIMIT $limit`

		result, err := tx.Run(query, map[string]interface{}{
			"user_id": userID,
			"offset":  offset,
			"limit":   limit,
		})
		if err != nil {
			return nil, err
		}

		var following []*UserConnection
		for result.Next() {
			record := result.Record()
			followingID, _ := record.Get("user_id")
			username, _ := record.Get("username")
			isCelebrity, _ := record.Get("is_celebrity")

			following = append(following, &UserConnection{
				UserID:      followingID.(int64),
				Username:    username.(string),
				IsCelebrity: isCelebrity.(bool),
				IsFollowing: true,
			})
		}

		return following, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get following: %w", err)
	}

	return result.([]*UserConnection), nil
}

// GetCelebritiesFollowed returns celebrities that the user follows (for pull-based feed)
func (s *SocialGraphService) GetCelebritiesFollowed(ctx context.Context, userID int64) ([]*UserConnection, error) {
	session := s.neo4j.NewSession(neo4j.SessionConfig{})
	defer session.Close()

	result, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		query := `
			MATCH (user:User {user_id: $user_id})-[:FOLLOWS]->(celebrity:User {is_celebrity: true})
			RETURN celebrity.user_id as user_id, celebrity.username as username
			ORDER BY celebrity.username`

		result, err := tx.Run(query, map[string]interface{}{
			"user_id": userID,
		})
		if err != nil {
			return nil, err
		}

		var celebrities []*UserConnection
		for result.Next() {
			record := result.Record()
			celebrityID, _ := record.Get("user_id")
			username, _ := record.Get("username")

			celebrities = append(celebrities, &UserConnection{
				UserID:      celebrityID.(int64),
				Username:    username.(string),
				IsCelebrity: true,
				IsFollowing: true,
			})
		}

		return celebrities, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get celebrities followed: %w", err)
	}

	return result.([]*UserConnection), nil
}

// GetSocialStats returns social statistics for a user
func (s *SocialGraphService) GetSocialStats(ctx context.Context, userID int64) (*SocialStats, error) {
	// Try Redis cache first
	pipe := s.redis.Pipeline()
	followersCountCmd := pipe.Get(ctx, fmt.Sprintf("follower_count:%d", userID))
	followingCountCmd := pipe.Get(ctx, fmt.Sprintf("following_count:%d", userID))
	_, err := pipe.Exec(ctx)

	if err == nil {
		followersCount, _ := strconv.ParseInt(followersCountCmd.Val(), 10, 64)
		followingCount, _ := strconv.ParseInt(followingCountCmd.Val(), 10, 64)

		if followersCount > 0 || followingCount > 0 {
			return &SocialStats{
				UserID:         userID,
				FollowersCount: followersCount,
				FollowingCount: followingCount,
			}, nil
		}
	}

	// Query Neo4j for accurate counts
	session := s.neo4j.NewSession(neo4j.SessionConfig{})
	defer session.Close()

	result, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		query := `
			MATCH (user:User {user_id: $user_id})
			OPTIONAL MATCH (follower:User)-[:FOLLOWS]->(user)
			OPTIONAL MATCH (user)-[:FOLLOWS]->(following:User)
			RETURN count(DISTINCT follower) as followers_count, count(DISTINCT following) as following_count`

		result, err := tx.Run(query, map[string]interface{}{
			"user_id": userID,
		})
		if err != nil {
			return nil, err
		}

		if result.Next() {
			record := result.Record()
			followersCount, _ := record.Get("followers_count")
			followingCount, _ := record.Get("following_count")

			return &SocialStats{
				UserID:         userID,
				FollowersCount: followersCount.(int64),
				FollowingCount: followingCount.(int64),
			}, nil
		}

		return &SocialStats{UserID: userID}, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get social stats: %w", err)
	}

	stats := result.(*SocialStats)

	// Cache the results
	s.redis.Set(ctx, fmt.Sprintf("follower_count:%d", userID), stats.FollowersCount, 30*time.Minute)
	s.redis.Set(ctx, fmt.Sprintf("following_count:%d", userID), stats.FollowingCount, 30*time.Minute)

	return stats, nil
}

// Helper methods for Redis cache operations
func (s *SocialGraphService) getFollowersFromCache(ctx context.Context, userID int64, limit int) ([]*UserConnection, error) {
	key := fmt.Sprintf("followers:%d", userID)
	result := s.redis.ZRevRange(ctx, key, 0, int64(limit-1))
	if result.Err() != nil {
		return nil, result.Err()
	}

	var followers []*UserConnection
	for _, member := range result.Val() {
		followerID, _ := strconv.ParseInt(member, 10, 64)
		// In a real implementation, you'd fetch user details from cache or database
		followers = append(followers, &UserConnection{
			UserID:     followerID,
			IsFollower: true,
		})
	}

	return followers, nil
}

func (s *SocialGraphService) getFollowingFromCache(ctx context.Context, userID int64, limit int) ([]*UserConnection, error) {
	key := fmt.Sprintf("following:%d", userID)
	result := s.redis.ZRevRange(ctx, key, 0, int64(limit-1))
	if result.Err() != nil {
		return nil, result.Err()
	}

	var following []*UserConnection
	for _, member := range result.Val() {
		followingID, _ := strconv.ParseInt(member, 10, 64)
		following = append(following, &UserConnection{
			UserID:      followingID,
			IsFollowing: true,
		})
	}

	return following, nil
}

// updateUserCounts updates user profile counts asynchronously
func (s *SocialGraphService) updateUserCounts(ctx context.Context, followerID, followeeID int64) {
	// Get updated counts
	followerStats, _ := s.GetSocialStats(ctx, followerID)
	followeeStats, _ := s.GetSocialStats(ctx, followeeID)

	// Update user profile service
	if followerStats != nil {
		s.userService.UpdateSocialCounts(ctx, followerID, followerStats.FollowersCount, followerStats.FollowingCount)
	}
	if followeeStats != nil {
		s.userService.UpdateSocialCounts(ctx, followeeID, followeeStats.FollowersCount, followeeStats.FollowingCount)
	}
}
