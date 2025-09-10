package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gocql/gocql"
	_ "github.com/lib/pq"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"

	"news-feed-system/services"
)

func main() {
	ctx := context.Background()

	// Initialize all database connections
	db, redisClient, neo4jDriver, cassandraSession := initializeAllConnections()
	defer db.Close()
	defer redisClient.Close()
	defer neo4jDriver.Close()
	defer cassandraSession.Close()

	// Initialize all services
	userService := services.NewUserProfileService(db, redisClient)
	socialService := services.NewSocialGraphService(neo4jDriver, redisClient, userService)
	postService := services.NewPostService(cassandraSession, redisClient, userService)
	feedService := services.NewFeedService(cassandraSession, redisClient, postService, socialService, userService)

	// Initialize and start fanout workers
	worker1 := services.NewFanoutWorker(cassandraSession, redisClient, socialService, userService, "worker-1")
	worker2 := services.NewFanoutWorker(cassandraSession, redisClient, socialService, userService, "worker-2")

	// Start workers
	worker1.Start(ctx)
	worker2.Start(ctx)
	defer worker1.Stop()
	defer worker2.Stop()

	// Demonstrate complete system workflow
	demonstrateCompleteWorkflow(ctx, userService, socialService, postService, feedService)

	// Keep workers running for a bit to process fanout
	fmt.Println("Waiting for fanout processing...")
	time.Sleep(5 * time.Second)

	// Show final feed results
	demonstrateFeedGeneration(ctx, feedService)
}

func demonstrateCompleteWorkflow(ctx context.Context, userService *services.UserProfileService,
	socialService *services.SocialGraphService, postService *services.PostService,
	feedService *services.FeedService) {

	fmt.Println("=== Complete News Feed System Demo ===\n")

	// 1. Create users
	fmt.Println("1. Creating users...")
	users := createDemoUsers(ctx, userService)

	// 2. Create social relationships
	fmt.Println("\n2. Creating social relationships...")
	createSocialRelationships(ctx, socialService, users)

	// 3. Create posts
	fmt.Println("\n3. Creating posts...")
	posts := createDemoPosts(ctx, postService, users)

	// 4. Show user stats
	fmt.Println("\n4. User statistics:")
	showUserStats(ctx, socialService, users)

	fmt.Printf("\nCreated %d users, established relationships, and created %d posts\n", len(users), len(posts))
}

func createDemoUsers(ctx context.Context, userService *services.UserProfileService) []*services.User {
	userRequests := []*services.CreateUserRequest{
		{
			Username:    "alice_blogger",
			Email:       "alice@example.com",
			Password:    "password123",
			DisplayName: "Alice the Blogger",
			Bio:         "Tech blogger and coffee enthusiast ☕",
		},
		{
			Username:    "bob_celebrity",
			Email:       "bob@example.com",
			Password:    "password123",
			DisplayName: "Bob Celebrity",
			Bio:         "Famous actor and philanthropist ⭐",
		},
		{
			Username:    "charlie_dev",
			Email:       "charlie@example.com",
			Password:    "password123",
			DisplayName: "Charlie Developer",
			Bio:         "Full-stack developer, open source contributor 💻",
		},
		{
			Username:    "diana_artist",
			Email:       "diana@example.com",
			Password:    "password123",
			DisplayName: "Diana Artist",
			Bio:         "Digital artist and designer 🎨",
		},
		{
			Username:    "eve_scientist",
			Email:       "eve@example.com",
			Password:    "password123",
			DisplayName: "Dr. Eve Scientist",
			Bio:         "Researcher in AI and machine learning 🧬",
		},
	}

	var users []*services.User
	for _, req := range userRequests {
		user, err := userService.CreateUser(ctx, req)
		if err != nil {
			log.Printf("Failed to create user %s: %v", req.Username, err)
			continue
		}
		users = append(users, user)
		fmt.Printf("  Created user: %s (%s)\n", user.Username, user.DisplayName)
	}

	// Simulate Bob as a celebrity by updating his follower count
	if len(users) >= 2 {
		userService.UpdateSocialCounts(ctx, users[1].UserID, 150000, 100) // Bob becomes celebrity
		fmt.Printf("  Updated %s to celebrity status\n", users[1].Username)
	}

	return users
}

func createSocialRelationships(ctx context.Context, socialService *services.SocialGraphService, users []*services.User) {
	if len(users) < 5 {
		fmt.Println("  Not enough users to create relationships")
		return
	}

	relationships := []struct {
		follower, followee int
		description        string
	}{
		{0, 1, "Alice follows Bob (celebrity)"},
		{0, 2, "Alice follows Charlie"},
		{0, 3, "Alice follows Diana"},
		{2, 0, "Charlie follows Alice (mutual)"},
		{2, 1, "Charlie follows Bob (celebrity)"},
		{2, 4, "Charlie follows Eve"},
		{3, 0, "Diana follows Alice"},
		{3, 1, "Diana follows Bob (celebrity)"},
		{3, 4, "Diana follows Eve"},
		{4, 1, "Eve follows Bob (celebrity)"},
		{4, 2, "Eve follows Charlie"},
	}

	for _, rel := range relationships {
		if rel.follower < len(users) && rel.followee < len(users) {
			err := socialService.FollowUser(ctx, users[rel.follower].UserID, users[rel.followee].UserID)
			if err != nil {
				log.Printf("Failed to create relationship: %v", err)
				continue
			}
			fmt.Printf("  %s\n", rel.description)
		}
	}
}

func createDemoPosts(ctx context.Context, postService *services.PostService, users []*services.User) []*services.Post {
	postRequests := []struct {
		authorIndex int
		content     string
	}{
		{0, "Just published a new blog post about microservices architecture! Check it out 📝 #tech #microservices"},
		{1, "Excited to announce my new movie coming out next month! 🎬 Thanks to all my fans for the support ❤️"},
		{2, "Working on a new open source project. Contributions welcome! 💻 #opensource #golang"},
		{3, "New digital art piece completed! Inspired by the beauty of nature 🌸 #digitalart #nature"},
		{4, "Published research paper on neural networks. Link in bio 🧠 #AI #research #machinelearning"},
		{1, "Behind the scenes from today's photoshoot 📸 #celebrity #photoshoot"},
		{0, "Coffee and code - perfect combination for a productive morning ☕💻 #coding #coffee"},
		{2, "Code review tip: Always consider edge cases! 🐛 #programming #tips"},
		{3, "Working on a commission piece. The creative process is so rewarding! 🎨"},
		{4, "Attending AI conference next week. Looking forward to the presentations! 🤖 #AI #conference"},
	}

	var posts []*services.Post
	for i, req := range postRequests {
		if req.authorIndex >= len(users) {
			continue
		}

		createReq := &services.CreatePostRequest{
			Content:    req.content,
			Visibility: "public",
		}

		post, err := postService.CreatePost(ctx, users[req.authorIndex].UserID, createReq)
		if err != nil {
			log.Printf("Failed to create post: %v", err)
			continue
		}

		posts = append(posts, post)
		fmt.Printf("  Post %d by %s: %.50s...\n", i+1, users[req.authorIndex].Username, post.Content)

		// Add some engagement to posts
		if i%2 == 0 {
			// Add likes to every other post
			for j := 0; j < len(users); j++ {
				if j != req.authorIndex {
					postService.AddEngagement(ctx, post.PostID, users[j].UserID, "like", "")
				}
			}
		}

		// Small delay to ensure different timestamps
		time.Sleep(100 * time.Millisecond)
	}

	return posts
}

func showUserStats(ctx context.Context, socialService *services.SocialGraphService, users []*services.User) {
	for _, user := range users {
		stats, err := socialService.GetSocialStats(ctx, user.UserID)
		if err != nil {
			log.Printf("Failed to get stats for %s: %v", user.Username, err)
			continue
		}

		fmt.Printf("  %s: %d followers, %d following\n",
			user.Username, stats.FollowersCount, stats.FollowingCount)
	}
}

func demonstrateFeedGeneration(ctx context.Context, feedService *services.FeedService) {
	fmt.Println("\n=== Feed Generation Demo ===")

	// Simulate getting feeds for different users
	userIDs := []int64{1, 3, 5} // Alice, Charlie, Eve

	for _, userID := range userIDs {
		fmt.Printf("\nGenerating feed for user ID %d:\n", userID)

		feed, err := feedService.GetUserFeed(ctx, userID, 10, "")
		if err != nil {
			log.Printf("Failed to get feed for user %d: %v", userID, err)
			continue
		}

		fmt.Printf("  Feed contains %d items (source: %s)\n", len(feed.Items), feed.Source)

		for i, item := range feed.Items {
			fmt.Printf("  %d. @%s: %.60s... (score: %.2f)\n",
				i+1, item.AuthorUsername, item.Content, item.RankingScore)
		}

		if feed.HasMore {
			fmt.Printf("  ... and more (next page token: %s)\n", feed.NextPageToken)
		}
	}

	// Demonstrate trending posts
	fmt.Println("\nTrending posts (last 24 hours):")
	trending, err := feedService.GetTrendingPosts(ctx, 5, 24*time.Hour)
	if err != nil {
		log.Printf("Failed to get trending posts: %v", err)
	} else {
		for i, item := range trending {
			fmt.Printf("  %d. @%s: %.60s... (engagement: %.1f)\n",
				i+1, item.AuthorUsername, item.Content, item.EngagementScore)
		}
	}
}

func initializeAllConnections() (*sql.DB, *redis.Client, neo4j.Driver, *gocql.Session) {
	// PostgreSQL connection
	db, err := sql.Open("postgres", "postgres://user:password@localhost/newsfeed?sslmode=disable")
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL:", err)
	}

	// Redis connection
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	// Neo4j connection
	neo4jDriver, err := neo4j.NewDriver("bolt://localhost:7687", neo4j.BasicAuth("neo4j", "password", ""))
	if err != nil {
		log.Fatal("Failed to connect to Neo4j:", err)
	}

	// Cassandra connection
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = "newsfeed"
	cluster.Consistency = gocql.Quorum
	cluster.ProtoVersion = 4
	cluster.ConnectTimeout = time.Second * 10

	cassandraSession, err := cluster.CreateSession()
	if err != nil {
		log.Fatal("Failed to connect to Cassandra:", err)
	}

	return db, redisClient, neo4jDriver, cassandraSession
}

// Additional utility functions for testing and monitoring

func monitorFanoutQueue(ctx context.Context, worker *services.FanoutWorker) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			stats, err := worker.GetQueueStats(ctx)
			if err != nil {
				log.Printf("Failed to get queue stats: %v", err)
				continue
			}

			fmt.Printf("Queue stats - Pending: %d, Processing: %d, Completed: %d, Failed: %d\n",
				stats["pending"], stats["processing"], stats["completed"], stats["failed"])
		}
	}
}

func simulateUserActivity(ctx context.Context, postService *services.PostService,
	feedService *services.FeedService, userIDs []int64) {

	// Simulate users creating posts and engaging with content
	for i := 0; i < 10; i++ {
		userID := userIDs[i%len(userIDs)]

		// Create a post
		createReq := &services.CreatePostRequest{
			Content:    fmt.Sprintf("Simulated post #%d from user %d at %s", i+1, userID, time.Now().Format("15:04:05")),
			Visibility: "public",
		}

		post, err := postService.CreatePost(ctx, userID, createReq)
		if err != nil {
			log.Printf("Failed to create simulated post: %v", err)
			continue
		}

		fmt.Printf("User %d created post: %s\n", userID, post.Content)

		// Simulate engagement from other users
		for _, otherUserID := range userIDs {
			if otherUserID != userID && i%3 == 0 { // 1/3 chance of engagement
				postService.AddEngagement(ctx, post.PostID, otherUserID, "like", "")
			}
		}

		time.Sleep(2 * time.Second)
	}
}

func benchmarkFeedGeneration(ctx context.Context, feedService *services.FeedService, userID int64, iterations int) {
	fmt.Printf("Benchmarking feed generation for user %d (%d iterations)...\n", userID, iterations)

	start := time.Now()
	var totalItems int

	for i := 0; i < iterations; i++ {
		feed, err := feedService.GetUserFeed(ctx, userID, 10, "")
		if err != nil {
			log.Printf("Benchmark iteration %d failed: %v", i+1, err)
			continue
		}
		totalItems += len(feed.Items)
	}

	duration := time.Since(start)
	avgLatency := duration / time.Duration(iterations)

	fmt.Printf("Benchmark results:\n")
	fmt.Printf("  Total time: %v\n", duration)
	fmt.Printf("  Average latency: %v\n", avgLatency)
	fmt.Printf("  Average items per feed: %.1f\n", float64(totalItems)/float64(iterations))
	fmt.Printf("  Feeds per second: %.1f\n", float64(iterations)/duration.Seconds())
}
