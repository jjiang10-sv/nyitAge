// Social Graph Schema (Neo4j)

// Create constraints for better performance
CREATE CONSTRAINT user_id_unique IF NOT EXISTS FOR (u:User) REQUIRE u.user_id IS UNIQUE;
CREATE INDEX user_username_index IF NOT EXISTS FOR (u:User) ON (u.username);

// User nodes (minimal data, full profile in PostgreSQL)
CREATE (u1:User {
    user_id: 1, 
    username: "john_doe",
    is_celebrity: false,
    created_at: datetime()
});

CREATE (u2:User {
    user_id: 2, 
    username: "celebrity_user",
    is_celebrity: true,
    created_at: datetime()
});

CREATE (u3:User {
    user_id: 3, 
    username: "jane_smith",
    is_celebrity: false,
    created_at: datetime()
});

CREATE (u4:User {
    user_id: 4, 
    username: "tech_influencer",
    is_celebrity: true,
    created_at: datetime()
});

// Follow relationships
MATCH (u1:User {user_id: 1}), (u2:User {user_id: 2})
CREATE (u1)-[:FOLLOWS {
    since: datetime(),
    is_active: true,
    notification_enabled: true
}]->(u2);

MATCH (u1:User {user_id: 1}), (u3:User {user_id: 3})
CREATE (u1)-[:FOLLOWS {
    since: datetime(),
    is_active: true,
    notification_enabled: true
}]->(u3);

MATCH (u3:User {user_id: 3}), (u1:User {user_id: 1})
CREATE (u3)-[:FOLLOWS {
    since: datetime(),
    is_active: true,
    notification_enabled: false
}]->(u1);

MATCH (u3:User {user_id: 3}), (u4:User {user_id: 4})
CREATE (u3)-[:FOLLOWS {
    since: datetime(),
    is_active: true,
    notification_enabled: true
}]->(u4);

// Block relationships (for privacy/safety)
// MATCH (u1:User {user_id: 1}), (u5:User {user_id: 5})
// CREATE (u1)-[:BLOCKS {since: datetime(), reason: "spam"}]->(u5);

// Common queries for social graph

// 1. Get all followers of a user
// MATCH (follower:User)-[:FOLLOWS]->(user:User {user_id: $user_id})
// RETURN follower.user_id, follower.username
// ORDER BY follower.username;

// 2. Get all users that a user follows
// MATCH (user:User {user_id: $user_id})-[:FOLLOWS]->(following:User)
// RETURN following.user_id, following.username, following.is_celebrity
// ORDER BY following.username;

// 3. Get mutual follows (friends)
// MATCH (user1:User {user_id: $user1_id})-[:FOLLOWS]->(mutual:User)<-[:FOLLOWS]-(user2:User {user_id: $user2_id})
// RETURN mutual.user_id, mutual.username;

// 4. Get follower count
// MATCH (follower:User)-[:FOLLOWS]->(user:User {user_id: $user_id})
// RETURN count(follower) as follower_count;

// 5. Get following count
// MATCH (user:User {user_id: $user_id})-[:FOLLOWS]->(following:User)
// RETURN count(following) as following_count;

// 6. Check if user A follows user B
// MATCH (userA:User {user_id: $user_a_id})-[r:FOLLOWS]->(userB:User {user_id: $user_b_id})
// RETURN r IS NOT NULL as is_following;

// 7. Get celebrities that user follows (for pull-based feed)
// MATCH (user:User {user_id: $user_id})-[:FOLLOWS]->(celebrity:User {is_celebrity: true})
// RETURN celebrity.user_id, celebrity.username;

// 8. Get non-celebrity follows (for push-based feed)
// MATCH (user:User {user_id: $user_id})-[:FOLLOWS]->(regular:User {is_celebrity: false})
// RETURN regular.user_id, regular.username;

// 9. Suggest users to follow (friends of friends, not already following)
// MATCH (user:User {user_id: $user_id})-[:FOLLOWS]->(friend:User)-[:FOLLOWS]->(suggestion:User)
// WHERE NOT (user)-[:FOLLOWS]->(suggestion) AND suggestion.user_id <> $user_id
// RETURN suggestion.user_id, suggestion.username, count(*) as mutual_friends
// ORDER BY mutual_friends DESC
// LIMIT 10;