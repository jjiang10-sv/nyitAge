package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"

	"news-feed-system/services"
)

func main() {
	ctx := context.Background()

	// Initialize database connections
	db, redisClient, neo4jDriver := initializeConnections()
	defer db.Close()
	defer redisClient.Close()
	defer neo4jDriver.Close()

	// Initialize services
	userService := services.NewUserProfileService(db, redisClient)
	socialService := services.NewSocialGraphService(neo4jDriver, redisClient, userService)

	// Example usage
	demonstrateUserProfileService(ctx, userService)
	demonstrateSocialGraphService(ctx, socialService)
}

func demonstrateUserProfileService(ctx context.Context, userService *services.UserProfileService) {
	fmt.Println("=== User Profile Service Demo ===")

	// Create a new user
	createReq := &services.CreateUserRequest{
		Username:    "demo_user",
		Email:       "demo@example.com",
		Password:    "securepassword123",
		DisplayName: "Demo User",
		Bio:         "This is a demo user for testing",
	}

	user, err := userService.CreateUser(ctx, createReq)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		return
	}

	fmt.Printf("Created user: %+v\n", user)

	// Get user by ID
	retrievedUser, err := userService.GetUserByID(ctx, user.UserID)
	if err != nil {
		log.Printf("Error retrieving user: %v", err)
		return
	}

	fmt.Printf("Retrieved user: %+v\n", retrievedUser)

	// Update user profile
	updateReq := &services.UpdateUserRequest{
		DisplayName: "Updated Demo User",
		Bio:         "Updated bio for demo user",
	}

	updatedUser, err := userService.UpdateUser(ctx, user.UserID, updateReq)
	if err != nil {
		log.Printf("Error updating user: %v", err)
		return
	}

	fmt.Printf("Updated user: %+v\n", updatedUser)

	// Authenticate user
	authUser, err := userService.AuthenticateUser(ctx, "demo_user", "securepassword123")
	if err != nil {
		log.Printf("Error authenticating user: %v", err)
		return
	}

	fmt.Printf("Authenticated user: %+v\n", authUser)
}

func demonstrateSocialGraphService(ctx context.Context, socialService *services.SocialGraphService) {
	fmt.Println("\n=== Social Graph Service Demo ===")

	// Assume we have users with IDs 1, 2, 3, 4
	userIDs := []int64{1, 2, 3, 4}

	// Create follow relationships
	fmt.Println("Creating follow relationships...")

	// User 1 follows User 2 (celebrity)
	err := socialService.FollowUser(ctx, 1, 2)
	if err != nil {
		log.Printf("Error following user: %v", err)
	} else {
		fmt.Println("User 1 now follows User 2 (celebrity)")
	}

	// User 1 follows User 3
	err = socialService.FollowUser(ctx, 1, 3)
	if err != nil {
		log.Printf("Error following user: %v", err)
	} else {
		fmt.Println("User 1 now follows User 3")
	}

	// User 3 follows User 1 (mutual follow)
	err = socialService.FollowUser(ctx, 3, 1)
	if err != nil {
		log.Printf("Error following user: %v", err)
	} else {
		fmt.Println("User 3 now follows User 1 (mutual)")
	}

	// User 3 follows User 4 (tech influencer)
	err = socialService.FollowUser(ctx, 3, 4)
	if err != nil {
		log.Printf("Error following user: %v", err)
	} else {
		fmt.Println("User 3 now follows User 4 (tech influencer)")
	}

	// Check follow relationships
	fmt.Println("\nChecking follow relationships...")

	isFollowing, err := socialService.IsFollowing(ctx, 1, 2)
	if err != nil {
		log.Printf("Error checking follow relationship: %v", err)
	} else {
		fmt.Printf("User 1 follows User 2: %t\n", isFollowing)
	}

	isFollowing, err = socialService.IsFollowing(ctx, 2, 1)
	if err != nil {
		log.Printf("Error checking follow relationship: %v", err)
	} else {
		fmt.Printf("User 2 follows User 1: %t\n", isFollowing)
	}

	// Get followers and following
	fmt.Println("\nGetting social connections...")

	for _, userID := range userIDs {
		followers, err := socialService.GetFollowers(ctx, userID, 10, 0)
		if err != nil {
			log.Printf("Error getting followers for user %d: %v", userID, err)
			continue
		}

		following, err := socialService.GetFollowing(ctx, userID, 10, 0)
		if err != nil {
			log.Printf("Error getting following for user %d: %v", userID, err)
			continue
		}

		fmt.Printf("User %d has %d followers and follows %d users\n",
			userID, len(followers), len(following))

		if len(followers) > 0 {
			fmt.Printf("  Followers: ")
			for i, follower := range followers {
				if i > 0 {
					fmt.Print(", ")
				}
				fmt.Printf("%s(%d)", follower.Username, follower.UserID)
			}
			fmt.Println()
		}

		if len(following) > 0 {
			fmt.Printf("  Following: ")
			for i, follow := range following {
				if i > 0 {
					fmt.Print(", ")
				}
				fmt.Printf("%s(%d)", follow.Username, follow.UserID)
				if follow.IsCelebrity {
					fmt.Print("⭐")
				}
			}
			fmt.Println()
		}
	}

	// Get celebrities followed by user 1 (for pull-based feed)
	fmt.Println("\nGetting celebrities followed by User 1...")
	celebrities, err := socialService.GetCelebritiesFollowed(ctx, 1)
	if err != nil {
		log.Printf("Error getting celebrities: %v", err)
	} else {
		fmt.Printf("User 1 follows %d celebrities: ", len(celebrities))
		for i, celeb := range celebrities {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Printf("%s(%d)", celeb.Username, celeb.UserID)
		}
		fmt.Println()
	}

	// Get social stats
	fmt.Println("\nGetting social statistics...")
	for _, userID := range userIDs {
		stats, err := socialService.GetSocialStats(ctx, userID)
		if err != nil {
			log.Printf("Error getting stats for user %d: %v", userID, err)
			continue
		}

		fmt.Printf("User %d: %d followers, %d following\n",
			stats.UserID, stats.FollowersCount, stats.FollowingCount)
	}

	// Demonstrate unfollow
	fmt.Println("\nTesting unfollow...")
	err = socialService.UnfollowUser(ctx, 1, 3)
	if err != nil {
		log.Printf("Error unfollowing user: %v", err)
	} else {
		fmt.Println("User 1 unfollowed User 3")

		// Verify unfollow
		isFollowing, _ := socialService.IsFollowing(ctx, 1, 3)
		fmt.Printf("User 1 still follows User 3: %t\n", isFollowing)
	}
}

func initializeConnections() (*sql.DB, *redis.Client, neo4j.Driver) {
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

	return db, redisClient, neo4jDriver
}

// Example of how the services would be used in HTTP handlers
func exampleHTTPHandlers() {
	// This would be in your HTTP handler layer

	// POST /api/v1/users/{user_id}/follow
	// func followUserHandler(w http.ResponseWriter, r *http.Request) {
	//     userID := getUserIDFromToken(r)
	//     targetUserID := getTargetUserIDFromPath(r)
	//
	//     err := socialService.FollowUser(r.Context(), userID, targetUserID)
	//     if err != nil {
	//         http.Error(w, err.Error(), http.StatusBadRequest)
	//         return
	//     }
	//
	//     w.WriteHeader(http.StatusCreated)
	//     json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	// }

	// GET /api/v1/users/{user_id}/followers
	// func getFollowersHandler(w http.ResponseWriter, r *http.Request) {
	//     userID := getUserIDFromPath(r)
	//     limit := getLimitFromQuery(r, 20)
	//     offset := getOffsetFromQuery(r, 0)
	//
	//     followers, err := socialService.GetFollowers(r.Context(), userID, limit, offset)
	//     if err != nil {
	//         http.Error(w, err.Error(), http.StatusInternalServerError)
	//         return
	//     }
	//
	//     json.NewEncoder(w).Encode(map[string]interface{}{
	//         "followers": followers,
	//         "count": len(followers),
	//     })
	// }
}
