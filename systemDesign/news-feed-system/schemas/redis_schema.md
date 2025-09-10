# Redis Schema for Social Graph Hot Data

## Key Patterns

### Follower Lists (Sorted Sets)
```redis
# Followers of a user (sorted by follow timestamp)
ZADD followers:1 1640995200 2    # user 2 follows user 1 at timestamp
ZADD followers:1 1640995300 3    # user 3 follows user 1 at timestamp
ZADD followers:2 1640995400 1    # user 1 follows user 2 at timestamp

# Get all followers of user 1
ZRANGE followers:1 0 -1

# Get recent followers (last 100)
ZREVRANGE followers:1 0 99

# Get follower count
ZCARD followers:1
```

### Following Lists (Sorted Sets)
```redis
# Users that user 1 follows (sorted by follow timestamp)
ZADD following:1 1640995200 2    # user 1 follows user 2
ZADD following:1 1640995300 3    # user 1 follows user 3

# Get all users that user 1 follows
ZRANGE following:1 0 -1

# Get celebrities that user 1 follows
ZINTER 2 following:1 celebrities WEIGHTS 1 1
```

### Cached Counts (Strings)
```redis
# Follower counts (updated from Neo4j)
SET follower_count:1 1250
SET follower_count:2 50000000
SET following_count:1 180
SET following_count:2 100

# Get counts
GET follower_count:1
MGET follower_count:1 follower_count:2 following_count:1
```

### Celebrity Lists (Sets)
```redis
# Set of celebrity user IDs
SADD celebrities 2 4

# Check if user is celebrity
SISMEMBER celebrities 2

# Get all celebrities
SMEMBERS celebrities
```

### User Activity (Hashes)
```redis
# User activity tracking
HSET user_activity:1 last_post_time 1640995200
HSET user_activity:1 last_login_time 1640995300
HSET user_activity:1 posts_today 3

# Get user activity
HGETALL user_activity:1
```

### Follow Relationships Cache (Hashes)
```redis
# Quick follow relationship checks
HSET follows:1 2 1    # user 1 follows user 2 (1 = true)
HSET follows:1 3 1    # user 1 follows user 3
HSET follows:3 1 1    # user 3 follows user 1

# Check if user 1 follows user 2
HGET follows:1 2
```

### Mutual Follows Cache (Sets)
```redis
# Cache mutual follows for quick access
SADD mutual:1:3 2    # users 1 and 3 both follow user 2

# Get mutual follows between user 1 and 3
SMEMBERS mutual:1:3
```

## Sample Redis Commands

```bash
# Setup sample data
ZADD followers:1 1640995200 2 1640995300 3
ZADD followers:2 1640995400 1 1640995500 3 1640995600 4
ZADD following:1 1640995200 2 1640995300 3
ZADD following:3 1640995250 1 1640995350 4

SET follower_count:1 2
SET follower_count:2 3
SET following_count:1 2
SET following_count:3 2

SADD celebrities 2 4

HSET follows:1 2 1 3 1
HSET follows:3 1 1 4 1
HSET follows:2 1 0 3 0 4 0  # celebrity doesn't follow back

# Query examples
ZRANGE followers:1 0 -1        # Get all followers of user 1
ZCARD followers:2              # Count followers of user 2
GET follower_count:1           # Get cached follower count
SISMEMBER celebrities 2        # Check if user 2 is celebrity
HGET follows:1 2              # Check if user 1 follows user 2
```

## TTL Strategy

```redis
# Set TTL for different data types
EXPIRE followers:1 3600        # 1 hour for follower lists
EXPIRE follower_count:1 1800   # 30 minutes for counts
EXPIRE follows:1 7200          # 2 hours for relationship cache
# celebrities set has no TTL (relatively static)
```

## Cache Warming Strategy

```redis
# Warm cache for active users on startup
# This would be done programmatically
ZADD hot_users 1640995200 1 1640995300 2 1640995400 3

# Pipeline operations for efficiency
MULTI
ZADD followers:1 1640995200 2 1640995300 3
SET follower_count:1 2
HSET follows:1 2 1 3 1
EXEC
```