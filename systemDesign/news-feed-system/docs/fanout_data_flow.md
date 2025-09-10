# Fanout Data Flow - Step by Step

## Scenario: Alice follows Bob (regular user) and Celebrity (100M followers)

### **Step 1: Bob Creates a Post (Regular User)**

```
Bob (1K followers) creates: "Hello world!"
```

**Data Flow:**
```
1. POST /posts
   ↓
2. PostService.CreatePost()
   ↓
3. Insert into posts table
   ↓
4. Insert into posts_by_author table  
   ↓
5. Enqueue fanout job → fanout_queue
   ↓
6. Return success to Bob
```

**Database State:**
```sql
-- posts table
post_id: uuid1, author_id: bob_id, content: "Hello world!"

-- fanout_queue table  
queue_id: job1, post_id: uuid1, author_id: bob_id, status: "pending"
```

### **Step 2: Fanout Worker Processes Bob's Post**

```
FanoutWorker picks up job1
```

**Worker Logic:**
```go
// Worker processes the job
job := claimJob("job1")  // Bob's post

// Check if Bob is celebrity
bob := getUserByID(bob_id)
if bob.FollowersCount < 100000 {  // Bob has 1K followers
    // PUSH strategy - fanout to all followers
    processRegularUserPost(bob_post)
}
```

**Fanout Process:**
```go
// Get Bob's followers
followers := [alice_id, charlie_id, diana_id, ...]  // 1000 followers

// Insert into each follower's timeline
for follower_id in followers {
    INSERT INTO user_timeline (
        user_id: follower_id,
        post_id: uuid1, 
        author_id: bob_id,
        content: "Hello world!",
        created_at: now()
    )
}
```

**Result - Timeline Tables:**
```sql
-- Alice's timeline (user_timeline where user_id = alice_id)
user_id   | post_id | author_id | content       | created_at
----------|---------|-----------|---------------|------------
alice_id  | uuid1   | bob_id    | "Hello world!"| 2024-01-01

-- Charlie's timeline  
user_id     | post_id | author_id | content       | created_at
------------|---------|-----------|---------------|------------
charlie_id  | uuid1   | bob_id    | "Hello world!"| 2024-01-01

-- Diana's timeline
user_id   | post_id | author_id | content       | created_at  
----------|---------|-----------|---------------|------------
diana_id  | uuid1   | bob_id    | "Hello world!"| 2024-01-01
```

### **Step 3: Celebrity Creates a Post**

```
Celebrity (100M followers) creates: "New movie announcement!"
```

**Data Flow:**
```
1. POST /posts
   ↓
2. PostService.CreatePost()
   ↓  
3. Insert into posts table
   ↓
4. Enqueue fanout job → fanout_queue
   ↓
5. Return success to Celebrity
```

### **Step 4: Fanout Worker Processes Celebrity's Post**

```
FanoutWorker picks up celebrity job
```

**Worker Logic:**
```go
// Worker processes the celebrity job
celebrity := getUserByID(celebrity_id)
if celebrity.FollowersCount >= 100000 {  // Celebrity has 100M followers
    // PULL strategy - NO fanout, store in celebrity table
    processCelebrityPost(celebrity_post)
}
```

**Celebrity Process:**
```go
// NO timeline fanout - just store in celebrity_posts table
INSERT INTO celebrity_posts (
    author_id: celebrity_id,
    post_id: uuid2,
    content: "New movie announcement!",
    created_at: now()
)

// NO inserts into user_timeline table!
// Saves 100M database writes
```

**Result - Celebrity Table:**
```sql
-- celebrity_posts table
author_id    | post_id | content                  | created_at
-------------|---------|--------------------------|------------
celebrity_id | uuid2   | "New movie announcement!"| 2024-01-01
```

**Notice**: Alice's timeline is **NOT** updated with celebrity post!

### **Step 5: Alice Requests Her Feed**

```
GET /users/alice_id/feed
```

**Feed Generation Process:**
```go
func GetUserFeed(alice_id) {
    // 1. Get push-based content (pre-computed timeline)
    pushItems := query(`
        SELECT * FROM user_timeline 
        WHERE user_id = alice_id 
        ORDER BY created_at DESC
    `)
    // Returns: [Bob's "Hello world!" post]
    
    // 2. Get celebrities Alice follows
    celebrities := getCelebritiesFollowed(alice_id)
    // Returns: [celebrity_id]
    
    // 3. Get pull-based content (celebrity posts)
    pullItems := []
    for celebrity in celebrities {
        posts := query(`
            SELECT * FROM celebrity_posts 
            WHERE author_id = celebrity.id 
            ORDER BY created_at DESC 
            LIMIT 5
        `)
        pullItems.append(posts)
    }
    // Returns: [Celebrity's "New movie announcement!" post]
    
    // 4. Merge and rank
    allItems := merge(pushItems, pullItems)
    rankedItems := rankByRelevance(allItems)
    
    return rankedItems[:10]  // Top 10
}
```

**Alice's Feed Result:**
```json
{
  "items": [
    {
      "post_id": "uuid2",
      "author": "Celebrity", 
      "content": "New movie announcement!",
      "source": "pull"
    },
    {
      "post_id": "uuid1", 
      "author": "Bob",
      "content": "Hello world!",
      "source": "push"
    }
  ]
}
```

## Key Insights

### **Why Not Topic-Based?**

**If we used topics (like Kafka):**
```
- Bob's topic: "user_bob_posts"
- Celebrity's topic: "user_celebrity_posts"  
- Alice subscribes to: ["user_bob_posts", "user_celebrity_posts"]

Problem: Alice would need to subscribe to 1000+ topics (all users she follows)
Feed generation = query 1000+ topics = slow!
```

**Our approach:**
```
- Alice's timeline: Pre-computed, single table query = fast!
- Celebrity posts: Separate table, queried on-demand
- Feed generation: 2 queries max (timeline + celebrity posts)
```

### **Storage Comparison**

**Topic Model:**
```
- Bob's post stored once in "user_bob_posts" topic
- Each follower must query Bob's topic
- Feed = N queries (N = number of followed users)
```

**Our Fanout Model:**
```
- Bob's post copied to 1000 follower timelines  
- Each follower queries their own timeline once
- Feed = 1 query (+ celebrity query if needed)
```

### **Write vs Read Trade-off**

| Approach | Writes | Reads |
|----------|--------|-------|
| **Topic Model** | 1 write per post | N reads per feed (N = following count) |
| **Our Fanout** | N writes per post (N = follower count) | 1-2 reads per feed |

Since **reads >> writes** in social media (people read feeds more than post), our approach optimizes for the common case!

## Summary

The fanout system is **NOT** topic-based. Instead:

1. **Regular users**: Posts are **pushed** into each follower's personal timeline
2. **Celebrities**: Posts are stored separately and **pulled** during feed generation  
3. **Feed generation**: Combines pre-computed timeline (fast) with on-demand celebrity posts
4. **Result**: Fast feed reads with intelligent celebrity handling

This hybrid approach gives us **sub-200ms feed generation** even at 100M user scale!