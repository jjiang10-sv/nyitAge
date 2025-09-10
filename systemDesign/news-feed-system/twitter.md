Design a news feeding system
Feature: 100M users. Each user posts 1 post/day and subscribes 10 others. Text length is 1K. 
Optional: Some users are subscribed by 100M users.
each user can only see the top 10 posts from other users since subscription

Here’s a crisp system-design for a Twitter-style **news feed** at the scale you gave.

---

# 1) Quick sizing (order-of-magnitude)

* Users: **100M**
* Posts: **1 post/user/day → \~100M posts/day ≈ 1,157 writes/sec**
* Follows: **10 follows/user → \~1B follow edges**
* Post size: **\~1 KB text** (assume \~2 KB with metadata/indexing)

**Storage/day (posts only):** \~100M × 2 KB = **\~200 GB/day** (primary). With replication x3 → **\~600 GB/day**.
**Fanout volume (uniform case):** avg followers/user ≈ 10 → **\~1B timeline inserts/day ≈ \~11.6K inserts/sec**.
**Reads:** if DAU=20M and each opens feed 10×/day → **200M home-timeline reads/day ≈ \~2.3K reads/sec** (often much higher at peaks; plan 10× headroom).

**Celebrity tail:** a user with **100M followers** creates **100M** deliveries for a single post if you “push on write” blindly → needs special handling.

---

# 2) High-level architecture

**Edge/API**

* API Gateway + Auth
* Rate limiting / WAF

**Core services**

* **Post Service** (write path): create post, store content, append to author outbox.
* **Social Graph Service**: follow/ unfollow; fanout degree; blocks/mutes.
* **Feed Service**:

  * **Fanout workers** (push on write for normal authors).
  * **Home Timeline store** (per-user inbox).
  * **Hybrid fanout**: push for normal users; **pull on read** for celebrities.
* **Ranking/ML Service**: relevance score, heuristics (freshness, social, quality).
* **Feature Store**: aggregates, embeddings, counters.
* **Search/Indexing**: for discovery (optional to core feed).
* **Notification Service** (optional).

**Data infra**

* **Message bus**: Kafka / Pulsar (post events; fanout tasks).
* **Persistent stores**:

  * **Posts**: time-series KV/column store (Cassandra/Scylla/DynamoDB); blob store if media.
  * **Home timelines**: Cassandra/Scylla/RocksDB-backed service.
  * **Social graph**: write-optimized store (Cassandra/DynamoDB) or graph DB (careful at scale).
* **Caches**:

  * Redis cluster for hot timelines / user metadata / counters.
  * CDN for media.

---

# 3) Data model (simplified)

* **posts**(post\_id PK, author\_id, ts, text\_ptr/content, visibility, counters…)
* **outbox\_by\_author**(author\_id PK, ts DESC → post\_id)  // append-only
* **home\_inbox**(user\_id PK, ts DESC → post\_id, origin\_author\_id, dedupe\_key)
* **follows**(follower\_id PK, followee\_id, ts)
* **mutes/blocks** similarly
* **engagement counters** in a separate table or Redis (write-heavy).

---

# 4) Write path

1. Client → **Post Service** (validate, persist content).
2. Emit **PostCreated** event to Kafka (key = author\_id for partition locality).
3. **Fanout workers** consume:

   * Fetch **followers** of author.
   * If `followers_count < threshold` (e.g., 100k): **push** post\_id into each follower’s **home\_inbox** (batch writes).
   * If `followers_count ≥ threshold` (celebrity): **don’t push**. Only keep in author’s **outbox** and mark the author as “pull-on-read”.
4. Apply dedupe key, idempotent writes (exactly-once semantics not required; idempotency is enough).

**Throughput**:
Uniform case ≈ **11.6K inbox inserts/sec**, easy with batched writes (e.g., 1–5K rows/batch to Cassandra).

---

# 5) Read path (Home timeline)

1. Client → **Feed Service**: `GET /home?user_id&cursor`
2. **Merge sources**:

   * **Precomputed inbox** (pushed posts) from **home\_inbox** (primary).
   * **On-the-fly merge** with:

     * Followed authors marked celebrity: read recent items from their **outbox** and merge by timestamp.
     * Optionally a “global hot” stream (for explore).
3. **Ranking**:

   * Pull feature vectors (author affinity, freshness, embeddings, negative feedback).
   * Score → re-rank → apply demotions (seen, muted, adult, etc.).
4. **Serve page**:

   * Post IDs → hydrate from **posts** store.
   * Cache the assembled page (short TTL) in Redis for repeat opens.

**Hot cache strategy**:

* Keep **top N (e.g., 100–200)** post\_ids per active user in Redis (not all 100M users).
* Full inbox lives in Cassandra; Redis is only for **active** users (e.g., last 7–30 days).

---

# 6) Hybrid fanout (solves celebrity problem)

* **Threshold**: T followers (e.g., 100k).
* **Normal authors** (< T): **push** to followers’ inboxes on write (low read latency).
* **Celebrities** (≥ T): **pull** at read time:

  * Keep their **outbox** only.
  * At read, for each follower, **k-way merge** (their precomputed inbox + celebrity outboxes they follow).
  * Cache merged pages so you don’t recompute for every scroll.

This avoids 100M writes per celebrity post while keeping normal feed snappy.

---

# 7) Ranking & relevance (minimal viable)

Features (cheap first):

* Freshness (age decay).
* Social proof (engagement counts with saturation).
* Author-viewer affinity (past interactions).
* Quality heuristics (language, length, blacklist).
* Diversity: avoid back-to-back same author; interleave sources.

Start with **heuristics** → add ML later. Keep **explainability** logs for debugging.

---

# 8) Consistency & correctness

* **Eventual consistency**: acceptable for feeds.
* **Idempotent fanout**: inbox row key `(user_id, post_id)`; “INSERT IF NOT EXISTS”.
* **Delete/mute**: tombstone in posts; filter at read or sweep fanouts asynchronously.
* **Privacy/block**: check at **read time** (author or viewer state) and/or enforce at fanout time with allowlists.

---

# 9) Failure & scaling patterns

* **Kafka** with many partitions (e.g., 1–5K), key by author\_id for order, or by follower\_id for uniformity in inbox writes (pick one consistently).
* **Backpressure**: if fanout lags, degrade gracefully:

  * Prefer write-drop? No. Instead **switch to pull** after a lag threshold for affected authors.
* **Batch writes** to timeline store; use **async** and **retries with jitter**.
* **Multi-region**:

  * Posts stored regional; replicate asynchronously.
  * Home timelines local to reader region; celebrity outboxes globally replicated.

---

# 10) Capacity notes

* **Social graph (1B edges)**: with \~40–80 bytes raw + index overhead → **\~100–200 GB** (pre-replication).
* **Home inbox**: keep only recent **\~500 IDs**/user primary (older via pagination).

  * As IDs only (8B) + overhead → on disk few hundred GB; with replication → low TB.
* **Redis**: only for **active users’ heads** (e.g., 10–20M users × 200 IDs) → must be **sharded**. Consider Redis Streams or compressed structures, but primary should remain on disk store.

---

# 11) APIs (sketch)

* `POST /posts` → create post
* `GET /home?cursor=…&limit=…`
* `POST /follow`, `POST /unfollow`
* `POST /mute`, `POST /block`
* Internal:

  * `fanout.push(post_id)` (Kafka event)
  * `timeline.bulkInsert(user_id, [post_ids])`
  * `timeline.fetch(user_id, cursor, limit)`
  * `outbox.fetch(author_id, since_ts, limit)`

---

# 12) Optional optimizations

* **Write coalescing**: batch multiple posts to the same follower inbox in one mutation.
* **Cold-start**: prefill new users with “global hot” + followed authors’ last K.
* **Seen-state**: maintain per-user “last\_seen\_ts” in Redis to cheap-filter.
* **Dedup**: `(user_id, origin_author_id, origin_ts)` as secondary dedupe when reposts exist.
* **TTL/compaction**: keep only last **N=10k** items per inbox; long tail fetched from outboxes on demand.

---

# 13) What to pick (tech)

* **Message bus**: Kafka
* **Timeline & posts**: Cassandra/Scylla (LSM, time-series friendly) or DynamoDB
* **Caches**: Redis Cluster
* **Search**: OpenSearch/Elastic (optional)
* **Feature store**: Redis + column store
* **Infra**: Kubernetes with HPA; observability (Prometheus + Grafana + OpenTelemetry); SLOs on read latency (p50/p95).

---

## TL;DR design choice

* **Hybrid fanout** (push for normal users, pull for celebrities).
* **Precomputed per-user inbox** in a disk-backed time-series store; **Redis** only for hot heads.
* **Kafka** to decouple writes from fanout & provide backpressure.
* **Simple heuristics ranking** first, ML later.

If you want, I can sketch exact table schemas and a sequence diagram for write/read paths next.

Awesome—let’s lock in a concrete, production-leaning **PostgreSQL schema** and a **Go implementation** (using `pgx`) that covers: follows, posting, a DB-backed fanout queue, a worker that does hybrid fanout (push for normal authors, skip for celebrities), and a simple read API for the home timeline.

---

# PostgreSQL schema (DDL)

```sql
-- users: keep a cached followers_count (updated by app) for celeb threshold checks
CREATE TABLE IF NOT EXISTS users (
  id                BIGSERIAL PRIMARY KEY,
  username          TEXT UNIQUE NOT NULL,
  followers_count   BIGINT NOT NULL DEFAULT 0,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- follows: 1B-ish edges → composite PK + indexes
CREATE TABLE IF NOT EXISTS follows (
  follower_id   BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  followee_id   BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  since_ts      TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (follower_id, followee_id)
);
-- To fetch followers of an author fast (fanout on write)
CREATE INDEX IF NOT EXISTS idx_follows_followee ON follows(followee_id, follower_id);

-- posts: ~100M/day; text ~1KB
CREATE TABLE IF NOT EXISTS posts (
  id           BIGSERIAL PRIMARY KEY,
  author_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  body         TEXT NOT NULL CHECK (char_length(body) <= 4000), -- 1k text + headroom
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at   TIMESTAMPTZ
);
-- Author outbox access (reverse chronological)
CREATE INDEX IF NOT EXISTS idx_posts_author_time ON posts(author_id, created_at DESC);

-- home_inbox: per-user feed (precomputed), only IDs + origin
CREATE TABLE IF NOT EXISTS home_inbox (
  user_id          BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  post_id          BIGINT NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
  origin_author_id BIGINT NOT NULL,
  created_at       TIMESTAMPTZ NOT NULL, -- copy from post.created_at
  PRIMARY KEY (user_id, post_id)
);
-- Fast page-by-time for a single user
CREATE INDEX IF NOT EXISTS idx_home_inbox_user_time ON home_inbox(user_id, created_at DESC);

-- DB-backed fanout queue (simple, durable)
CREATE TABLE IF NOT EXISTS fanout_queue (
  id           BIGSERIAL PRIMARY KEY,
  post_id      BIGINT NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
  author_id    BIGINT NOT NULL,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  processed_at TIMESTAMPTZ,
  status       SMALLINT NOT NULL DEFAULT 0 -- 0=pending, 1=done, 2=failed
);
CREATE INDEX IF NOT EXISTS idx_fanout_pending ON fanout_queue(status, created_at);

-- Optional: keep only the last N items per inbox (periodic job can truncate)
-- Example helper to trim a single user's inbox to last 10k items:
-- DELETE FROM home_inbox hi
-- USING (
--   SELECT post_id FROM home_inbox
--   WHERE user_id=$1
--   ORDER BY created_at DESC
--   OFFSET 10000
-- ) old
-- WHERE hi.user_id=$1 AND hi.post_id=old.post_id;
```

**Notes**

* We rely on the `posts` table as the author outbox (via `idx_posts_author_time`).
* `home_inbox` stores only IDs + timestamps; we hydrate content from `posts` at read.
* **Hybrid fanout**: a constant threshold (e.g., `CELEB_THRESHOLD = 100_000`). If `followers_count ≥ threshold`, we don’t push; followers will pull on read (merge not shown here for brevity, but you can add it easily by reading the author outbox during `GetHome`).

---

# Go implementation (pgx, single file)

* Runs migrations (the DDL above).
* Provides functions: `Follow`, `Unfollow`, `CreatePost`, `RunFanoutWorker`, `GetHome`.
* Uses batch inserts for fanout.
* Keeps `followers_count` consistent in app logic.

Save as `main.go`.

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	CELEB_THRESHOLD       = 100_000
	FANOUT_BATCH_SIZE     = 5_000   // inbox inserts per batch
	WORKER_FETCH_LIMIT    = 200     // fanout jobs per poll
	WORKER_POLL_INTERVAL  = 500 * time.Millisecond
	HOME_PAGE_LIMIT       = 50
	HOME_MAX_PAGE_LIMIT   = 200
)

var ddl = []string{
	`CREATE TABLE IF NOT EXISTS users (
		id BIGSERIAL PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		followers_count BIGINT NOT NULL DEFAULT 0,
		created_at TIMESTAMPTZ NOT NULL DEFAULT now()
	);`,
	`CREATE TABLE IF NOT EXISTS follows (
		follower_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		followee_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		since_ts TIMESTAMPTZ NOT NULL DEFAULT now(),
		PRIMARY KEY (follower_id, followee_id)
	);`,
	`CREATE INDEX IF NOT EXISTS idx_follows_followee ON follows(followee_id, follower_id);`,
	`CREATE TABLE IF NOT EXISTS posts (
		id BIGSERIAL PRIMARY KEY,
		author_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		body TEXT NOT NULL CHECK (char_length(body) <= 4000),
		created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
		deleted_at TIMESTAMPTZ
	);`,
	`CREATE INDEX IF NOT EXISTS idx_posts_author_time ON posts(author_id, created_at DESC);`,
	`CREATE TABLE IF NOT EXISTS home_inbox (
		user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		post_id BIGINT NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
		origin_author_id BIGINT NOT NULL,
		created_at TIMESTAMPTZ NOT NULL,
		PRIMARY KEY (user_id, post_id)
	);`,
	`CREATE INDEX IF NOT EXISTS idx_home_inbox_user_time ON home_inbox(user_id, created_at DESC);`,
	`CREATE TABLE IF NOT EXISTS fanout_queue (
		id BIGSERIAL PRIMARY KEY,
		post_id BIGINT NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
		author_id BIGINT NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
		processed_at TIMESTAMPTZ,
		status SMALLINT NOT NULL DEFAULT 0
	);`,
	`CREATE INDEX IF NOT EXISTS idx_fanout_pending ON fanout_queue(status, created_at);`,
}

type App struct {
	db *pgxpool.Pool
}

func mustEnv(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}

func main() {
	ctx := context.Background()

	// Example DSN: postgres://user:pass@localhost:5432/newsfeed?pool_max_conns=20
	dsn := mustEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/newsfeed?sslmode=disable")
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	app := &App{db: pool}
	if err := app.migrate(ctx); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	// (Demo) start worker in background. In real use, run it as a separate process.
	go func() {
		if err := app.RunFanoutWorker(ctx); err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("fanout worker stopped: %v", err)
		}
	}()

	// Minimal demo: create two users, follow, post, read
	aliceID, _ := app.CreateUser(ctx, "alice")
	bobID, _ := app.CreateUser(ctx, "bob")

	_ = app.Follow(ctx, aliceID, bobID) // alice follows bob

	postID, err := app.CreatePost(ctx, bobID, "hello from bob")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Created post %d by bob", postID)

	time.Sleep(2 * time.Second) // let worker push

	items, err := app.GetHome(ctx, aliceID, 0, HOME_PAGE_LIMIT)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Alice home feed got %d items", len(items))

	// Block forever (demo)
	select {}
}

func (a *App) migrate(ctx context.Context) error {
	for _, q := range ddl {
		if _, err := a.db.Exec(ctx, q); err != nil {
			return fmt.Errorf("ddl: %w", err)
		}
	}
	return nil
}

/************* Core API *************/

// CreateUser inserts a user (idempotent on username)
func (a *App) CreateUser(ctx context.Context, username string) (int64, error) {
	var id int64
	err := a.db.QueryRow(ctx,
		`INSERT INTO users(username) VALUES($1)
		 ON CONFLICT (username) DO UPDATE SET username=EXCLUDED.username
		 RETURNING id;`, username).Scan(&id)
	return id, err
}

func (a *App) Follow(ctx context.Context, followerID, followeeID int64) error {
	tx, err := a.db.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `INSERT INTO follows(follower_id, followee_id)
		VALUES($1,$2) ON CONFLICT DO NOTHING;`, followerID, followeeID)
	if err != nil { return err }

	// update cached followers_count
	_, err = tx.Exec(ctx, `UPDATE users SET followers_count = followers_count + 1 WHERE id=$1;`, followeeID)
	if err != nil { return err }

	return tx.Commit(ctx)
}

func (a *App) Unfollow(ctx context.Context, followerID, followeeID int64) error {
	tx, err := a.db.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)

	ct, err := tx.Exec(ctx, `DELETE FROM follows WHERE follower_id=$1 AND followee_id=$2;`, followerID, followeeID)
	if err != nil { return err }

	if ct.RowsAffected() > 0 {
		_, err = tx.Exec(ctx, `UPDATE users SET followers_count = GREATEST(0, followers_count - 1) WHERE id=$1;`, followeeID)
		if err != nil { return err }
	}

	return tx.Commit(ctx)
}

// CreatePost persists the post and enqueues a fanout job
func (a *App) CreatePost(ctx context.Context, authorID int64, body string) (int64, error) {
	tx, err := a.db.Begin(ctx)
	if err != nil { return 0, err }
	defer tx.Rollback(ctx)

	var postID int64
	err = tx.QueryRow(ctx, `INSERT INTO posts(author_id, body) VALUES($1,$2) RETURNING id;`, authorID, body).Scan(&postID)
	if err != nil { return 0, err }

	_, err = tx.Exec(ctx, `INSERT INTO fanout_queue(post_id, author_id) VALUES($1,$2);`, postID, authorID)
	if err != nil { return 0, err }

	if err := tx.Commit(ctx); err != nil { return 0, err }
	return postID, nil
}

// GetHome returns hydrated posts from the precomputed inbox (push path).
// Cursor is epoch millis of last seen created_at (0 for newest).
type HomeItem struct {
	PostID     int64
	AuthorID   int64
	Body       string
	CreatedAt  time.Time
}

func (a *App) GetHome(ctx context.Context, userID int64, cursorMillis int64, limit int) ([]HomeItem, error) {
	if limit <= 0 || limit > HOME_MAX_PAGE_LIMIT {
		limit = HOME_PAGE_LIMIT
	}

	var rows pgxpool.Rows
	var err error
	if cursorMillis <= 0 {
		rows, err = a.db.Query(ctx, `
			SELECT p.id, p.author_id, p.body, p.created_at
			FROM home_inbox hi
			JOIN posts p ON p.id = hi.post_id
			WHERE hi.user_id = $1
			ORDER BY hi.created_at DESC
			LIMIT $2;`, userID, limit)
	} else {
		cursor := time.UnixMilli(cursorMillis).UTC()
		rows, err = a.db.Query(ctx, `
			SELECT p.id, p.author_id, p.body, p.created_at
			FROM home_inbox hi
			JOIN posts p ON p.id = hi.post_id
			WHERE hi.user_id = $1 AND hi.created_at < $2
			ORDER BY hi.created_at DESC
			LIMIT $3;`, userID, cursor, limit)
	}
	if err != nil { return nil, err }
	defer rows.Close()

	out := make([]HomeItem, 0, limit)
	for rows.Next() {
		var it HomeItem
		if err := rows.Scan(&it.PostID, &it.AuthorID, &it.Body, &it.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

/************* Fanout worker *************/

func (a *App) RunFanoutWorker(ctx context.Context) error {
	log.Printf("fanout worker: start, celeb threshold=%d", CELEB_THRESHOLD)
	t := time.NewTicker(WORKER_POLL_INTERVAL)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
			if err := a.processFanoutBatch(ctx); err != nil {
				log.Printf("fanout worker: %v", err)
			}
		}
	}
}

type fanoutJob struct {
	ID       int64
	PostID   int64
	AuthorID int64
	Created  time.Time
}

func (a *App) processFanoutBatch(ctx context.Context) error {
	// Claim jobs (simple approach: select pending oldest first)
	tx, err := a.db.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, `
		SELECT id, post_id, author_id, created_at
		FROM fanout_queue
		WHERE status=0
		ORDER BY created_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED;`, WORKER_FETCH_LIMIT)
	if err != nil { return err }
	defer rows.Close()

	jobs := make([]fanoutJob, 0, WORKER_FETCH_LIMIT)
	for rows.Next() {
		var j fanoutJob
		if err := rows.Scan(&j.ID, &j.PostID, &j.AuthorID, &j.Created); err != nil {
			return err
		}
		jobs = append(jobs, j)
	}
	if err := rows.Err(); err != nil { return err }

	if len(jobs) == 0 {
		return tx.Commit(ctx)
	}

	for _, j := range jobs {
		if err := a.fanoutOne(ctx, tx, j); err != nil {
			log.Printf("fanout job %d failed: %v", j.ID, err)
			_, _ = tx.Exec(ctx, `UPDATE fanout_queue SET status=2, processed_at=now() WHERE id=$1;`, j.ID)
			// Continue with other jobs
			continue
		}
		_, _ = tx.Exec(ctx, `UPDATE fanout_queue SET status=1, processed_at=now() WHERE id=$1;`, j.ID)
	}

	return tx.Commit(ctx)
}

func (a *App) fanoutOne(ctx context.Context, tx pgxTx, j fanoutJob) error {
	// celeb check (cached)
	var followersCount int64
	if err := tx.QueryRow(ctx, `SELECT followers_count FROM users WHERE id=$1;`, j.AuthorID).Scan(&followersCount); err != nil {
		return err
	}

	// If celebrity, we do not push (read-time pull is expected).
	if followersCount >= CELEB_THRESHOLD {
		return nil
	}

	// Get post created_at once
	var postCreated time.Time
	if err := tx.QueryRow(ctx, `SELECT created_at FROM posts WHERE id=$1;`, j.PostID).Scan(&postCreated); err != nil {
		return err
	}

	// Iterate followers in chunks; insert into home_inbox with ON CONFLICT DO NOTHING
	const page = 50_000
	var lastFollower int64 = 0
	for {
		followers, err := fetchFollowersPage(ctx, tx, j.AuthorID, lastFollower, page)
		if err != nil { return err }
		if len(followers) == 0 { break }

		if err := insertInboxBatch(ctx, tx, followers, j.PostID, j.AuthorID, postCreated); err != nil {
			return err
		}
		lastFollower = followers[len(followers)-1]
	}
	return nil
}

type pgxTx interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgxpool.Rows, error)
	QueryRow(context.Context, string, ...any) pgxRow
}

type pgxRow interface {
	Scan(...any) error
}

func fetchFollowersPage(ctx context.Context, tx pgxTx, authorID, afterFollowerID int64, limit int) ([]int64, error) {
	rows, err := tx.Query(ctx, `
		SELECT follower_id
		FROM follows
		WHERE followee_id=$1 AND follower_id > $2
		ORDER BY follower_id ASC
		LIMIT $3;`, authorID, afterFollowerID, limit)
	if err != nil { return nil, err }
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func insertInboxBatch(ctx context.Context, tx pgxTx, userIDs []int64, postID, originAuthorID int64, createdAt time.Time) error {
	// Batch inserts in smaller chunks to avoid giant statements
	for start := 0; start < len(userIDs); start += FANOUT_BATCH_SIZE {
		end := start + FANOUT_BATCH_SIZE
		if end > len(userIDs) { end = len(userIDs) }

		b := &pgxBatch{}
		for _, uid := range userIDs[start:end] {
			b.Queue(`INSERT INTO home_inbox(user_id, post_id, origin_author_id, created_at)
					 VALUES($1,$2,$3,$4)
					 ON CONFLICT (user_id, post_id) DO NOTHING;`,
				uid, postID, originAuthorID, createdAt)
		}
		if err := b.Send(ctx, tx); err != nil { return err }
	}
	return nil
}

/************* Minimal Batch helper over pgx *************/

// Tiny wrapper so we don't import pgx batch directly (keeps this single-file simple).
type pgxBatch struct{
	stmts []stmt
}
type stmt struct {
	sql string
	args []any
}
func (b *pgxBatch) Queue(sql string, args ...any) { b.stmts = append(b.stmts, stmt{sql, args}) }
func (b *pgxBatch) Send(ctx context.Context, tx pgxTx) error {
	for _, s := range b.stmts {
		if _, err := tx.Exec(ctx, s.sql, s.args...); err != nil {
			return err
		}
	}
	b.stmts = b.stmts[:0]
	return nil
}
```

---

## How this maps to the design

* **Hybrid fanout**: `RunFanoutWorker` pushes into `home_inbox` only if author’s `followers_count` < threshold; celebrity posts **skip push** (their followers will read from author outbox during `GetHome`—add that merge if/when you need it).
* **Idempotency**: `(user_id, post_id)` is the primary key of `home_inbox`, and inserts use `ON CONFLICT DO NOTHING`.
* **Throughput**: follower pagination + batch inserts avoid giant transactions; tune `FANOUT_BATCH_SIZE` and pool size.
* **Backpressure**: the worker takes `WORKER_FETCH_LIMIT` jobs per poll and uses `FOR UPDATE SKIP LOCKED` to shard across multiple workers safely.

---

## Next steps you can add quickly

* Add **read-time merge** for celebrity follows:

  * On `GetHome`, fetch top K from `posts` where `author_id IN (celebs user follows)` and merge with `home_inbox` by `created_at` before returning.
* Add **HTTP handlers** (e.g., chi/echo) to expose `POST /users`, `POST /follow`, `POST /posts`, `GET /home`.
* Add a periodic **inbox trimming** job to keep only the last N items per user.
* Add **soft delete** handling (filter `deleted_at IS NULL`).

If you want, I can extend this into a tiny HTTP service (handlers + JSON) or split the worker into its own `cmd/worker` binary.
