package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

/*
This is a self-contained newsfeed service with:
- PostgreSQL schema & migrations
- Hybrid fanout (push for normal authors; pull/merge for celebrities at read time)
- HTTP API: users, follow/unfollow, posts (create/delete), home timeline
- Backgrounds: fanout worker + inbox trimming job

Env:
  DATABASE_URL=postgres://postgres:postgres@localhost:5432/newsfeed?sslmode=disable
  CELEB_THRESHOLD=100000
  HOME_MAX_PER_USER=10000 (inbox trimming threshold)
*/

const (
	DefaultCelebThreshold   = 100_000
	FanoutBatchSize         = 5_000
	WorkerFetchLimit        = 200
	WorkerPollInterval      = 500 * time.Millisecond
	HomePageLimit           = 50
	HomeMaxPageLimit        = 200
	InboxTrimEvery          = 2 * time.Minute
	DefaultHomeMaxPerUser   = 10_000
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
	db              *pgxpool.Pool
	celebThreshold  int64
	inboxKeepMax    int
}

func mustEnv(k string, def string) string {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return def
	}
	return v
}

func mustEnvInt(k string, def int) int {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil { return def }
	return i
}

func mustEnvInt64(k string, def int64) int64 {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" { return def }
	i, err := strconv.ParseInt(v, 10, 64)
	if err != nil { return def }
	return i
}

func main() {
	ctx := context.Background()

	dsn := mustEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/newsfeed?sslmode=disable")
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	app := &App{db: pool,
		celebThreshold: mustEnvInt64("CELEB_THRESHOLD", DefaultCelebThreshold),
		inboxKeepMax:   mustEnvInt("HOME_MAX_PER_USER", DefaultHomeMaxPerUser),
	}
	if err := app.migrate(ctx); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	// Backgrounds
	go func() {
		if err := app.RunFanoutWorker(ctx); err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("fanout worker stopped: %v", err)
		}
	}()
	go app.periodicInboxTrim(ctx, InboxTrimEvery)

	// HTTP server
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Logger, middleware.Recoverer)
	app.mountRoutes(r)

	addr := mustEnv("ADDR", ":8080")
	log.Printf("http listening on %s (celeb_threshold=%d keep_max=%d)", addr, app.celebThreshold, app.inboxKeepMax)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}

func (a *App) migrate(ctx context.Context) error {
	for _, q := range ddl {
		if _, err := a.db.Exec(ctx, q); err != nil {
			return fmt.Errorf("ddl: %w", err)
		}
	}
	return nil
}

/************* HTTP *************/

func (a *App) mountRoutes(r chi.Router) {
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK); w.Write([]byte("ok")) })

	r.Route("/users", func(r chi.Router) {
		r.Post("/", a.handleCreateUser)
	})

	r.Post("/follow", a.handleFollow)
	r.Post("/unfollow", a.handleUnfollow)

	r.Route("/posts", func(r chi.Router) {
		r.Post("/", a.handleCreatePost)
		r.Delete("/{postID}", a.handleDeletePost)
	})

	r.Get("/home", a.handleGetHome)
}

type jsonResp map[string]any

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func parseID(s string) (int64, error) {
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid id")
	}
	return id, nil
}

/************* Core API (HTTP handlers -> methods) *************/

type createUserReq struct { Username string `json:"username"` }
func (a *App) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req createUserReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Username) == "" {
		writeJSON(w, 400, jsonResp{"error": "username required"}); return
	}
	id, err := a.CreateUser(ctx, strings.TrimSpace(req.Username))
	if err != nil { writeJSON(w, 500, jsonResp{"error": err.Error()}); return }
	writeJSON(w, 201, jsonResp{"id": id})
}

type followReq struct { FollowerID int64 `json:"follower_id"`; FolloweeID int64 `json:"followee_id"` }
func (a *App) handleFollow(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req followReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.FollowerID<=0 || req.FolloweeID<=0 {
		writeJSON(w, 400, jsonResp{"error":"invalid follower_id/followee_id"}); return
	}
	if err := a.Follow(ctx, req.FollowerID, req.FolloweeID); err != nil {
		writeJSON(w, 500, jsonResp{"error": err.Error()}); return
	}
	writeJSON(w, 200, jsonResp{"status":"ok"})
}
func (a *App) handleUnfollow(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req followReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.FollowerID<=0 || req.FolloweeID<=0 {
		writeJSON(w, 400, jsonResp{"error":"invalid follower_id/followee_id"}); return
	}
	if err := a.Unfollow(ctx, req.FollowerID, req.FolloweeID); err != nil {
		writeJSON(w, 500, jsonResp{"error": err.Error()}); return
	}
	writeJSON(w, 200, jsonResp{"status":"ok"})
}

type createPostReq struct { AuthorID int64 `json:"author_id"`; Body string `json:"body"` }
func (a *App) handleCreatePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req createPostReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.AuthorID<=0 || strings.TrimSpace(req.Body)=="" {
		writeJSON(w, 400, jsonResp{"error":"author_id and body required"}); return
	}
	postID, err := a.CreatePost(ctx, req.AuthorID, req.Body)
	if err != nil { writeJSON(w, 500, jsonResp{"error": err.Error()}); return }
	writeJSON(w, 201, jsonResp{"post_id": postID})
}

func (a *App) handleDeletePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pidStr := chi.URLParam(r, "postID")
	pid, err := parseID(pidStr)
	if err != nil { writeJSON(w, 400, jsonResp{"error":"invalid post id"}); return }
	if err := a.DeletePost(ctx, pid); err != nil {
		writeJSON(w, 500, jsonResp{"error": err.Error()}); return
	}
	writeJSON(w, 200, jsonResp{"status":"deleted"})
}

func (a *App) handleGetHome(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid, _ := strconv.ParseInt(r.URL.Query().Get("user_id"), 10, 64)
	if uid <= 0 { writeJSON(w, 400, jsonResp{"error":"user_id required"}); return }
	cursorMillis, _ := strconv.ParseInt(r.URL.Query().Get("cursor"), 10, 64)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	items, nextCursor, err := a.GetHome(ctx, uid, cursorMillis, limit)
	if err != nil { writeJSON(w, 500, jsonResp{"error": err.Error()}); return }
	writeJSON(w, 200, jsonResp{"items": items, "next_cursor": nextCursor})
}

/************* DB methods *************/

type HomeItem struct {
	PostID    int64     `json:"post_id"`
	AuthorID  int64     `json:"author_id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

func (a *App) CreateUser(ctx context.Context, username string) (int64, error) {
	var id int64
	err := a.db.QueryRow(ctx,
		`INSERT INTO users(username) VALUES($1)
		 ON CONFLICT (username) DO UPDATE SET username=EXCLUDED.username
		 RETURNING id;`, username).Scan(&id)
	return id, err
}

func (a *App) Follow(ctx context.Context, followerID, followeeID int64) error {
	tx, err := a.db.Begin(ctx); if err != nil { return err }
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `INSERT INTO follows(follower_id, followee_id)
		VALUES($1,$2) ON CONFLICT DO NOTHING;`, followerID, followeeID)
	if err != nil { return err }
	_, err = tx.Exec(ctx, `UPDATE users SET followers_count = followers_count + 1 WHERE id=$1;`, followeeID)
	if err != nil { return err }
	return tx.Commit(ctx)
}

func (a *App) Unfollow(ctx context.Context, followerID, followeeID int64) error {
	tx, err := a.db.Begin(ctx); if err != nil { return err }
	defer tx.Rollback(ctx)

	ct, err := tx.Exec(ctx, `DELETE FROM follows WHERE follower_id=$1 AND followee_id=$2;`, followerID, followeeID)
	if err != nil { return err }
	if ct.RowsAffected() > 0 {
		_, err = tx.Exec(ctx, `UPDATE users SET followers_count = GREATEST(0, followers_count - 1) WHERE id=$1;`, followeeID)
		if err != nil { return err }
	}
	return tx.Commit(ctx)
}

func (a *App) CreatePost(ctx context.Context, authorID int64, body string) (int64, error) {
	tx, err := a.db.Begin(ctx); if err != nil { return 0, err }
	defer tx.Rollback(ctx)

	var postID int64
	err = tx.QueryRow(ctx, `INSERT INTO posts(author_id, body) VALUES($1,$2) RETURNING id;`, authorID, body).Scan(&postID)
	if err != nil { return 0, err }
	_, err = tx.Exec(ctx, `INSERT INTO fanout_queue(post_id, author_id) VALUES($1,$2);`, postID, authorID)
	if err != nil { return 0, err }
	if err := tx.Commit(ctx); err != nil { return 0, err }
	return postID, nil
}

func (a *App) DeletePost(ctx context.Context, postID int64) error {
	tx, err := a.db.Begin(ctx); if err != nil { return err }
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `UPDATE posts SET deleted_at=now() WHERE id=$1 AND deleted_at IS NULL;`, postID)
	if err != nil { return err }
	// Best-effort removal from inboxes to save space; not strictly required if filtered on read
	_, _ = tx.Exec(ctx, `DELETE FROM home_inbox WHERE post_id=$1;`, postID)
	return tx.Commit(ctx)
}

// GetHome merges two sources:
// 1) Precomputed inbox (push path, for normal authors)
// 2) Celebrity posts (pull path) from authors the user follows whose followers_count>=threshold
// Cursor is epoch millis of the merged stream's created_at; returns nextCursor for pagination
func (a *App) GetHome(ctx context.Context, userID int64, cursorMillis int64, limit int) ([]HomeItem, int64, error) {
	if limit <= 0 || limit > HomeMaxPageLimit { limit = HomePageLimit }
	cursor := time.UnixMilli(0)
	useCursor := false
	if cursorMillis > 0 { cursor = time.UnixMilli(cursorMillis).UTC(); useCursor = true }

	// Fetch inbox items (already filtered to user's follows by construction)
	inboxRows, err := func() ([]HomeItem, error) {
		if !useCursor {
			rows, err := a.db.Query(ctx, `
				SELECT p.id, p.author_id, p.body, p.created_at
				FROM home_inbox hi
				JOIN posts p ON p.id = hi.post_id
				WHERE hi.user_id=$1 AND p.deleted_at IS NULL
				ORDER BY hi.created_at DESC
				LIMIT $2;`, userID, limit*2) // grab extra to merge
			if err != nil { return nil, err }
			defer rows.Close()
			return scanHomeRows(rows)
		} else {
			rows, err := a.db.Query(ctx, `
				SELECT p.id, p.author_id, p.body, p.created_at
				FROM home_inbox hi
				JOIN posts p ON p.id = hi.post_id
				WHERE hi.user_id=$1 AND hi.created_at < $2 AND p.deleted_at IS NULL
				ORDER BY hi.created_at DESC
				LIMIT $3;`, userID, cursor, limit*2)
			if err != nil { return nil, err }
			defer rows.Close()
			return scanHomeRows(rows)
		}
	}()
	if err != nil { return nil, 0, err }

	// Fetch celeb posts for authors followed by user
	celebRows, err := func() ([]HomeItem, error) {
		if !useCursor {
			rows, err := a.db.Query(ctx, `
				SELECT p.id, p.author_id, p.body, p.created_at
				FROM posts p
				JOIN follows f ON f.followee_id=p.author_id AND f.follower_id=$1
				JOIN users u ON u.id=p.author_id
				WHERE u.followers_count >= $2 AND p.deleted_at IS NULL
				ORDER BY p.created_at DESC
				LIMIT $3;`, userID, a.celebThreshold, limit*2)
			if err != nil { return nil, err }
			defer rows.Close()
			return scanHomeRows(rows)
		} else {
			rows, err := a.db.Query(ctx, `
				SELECT p.id, p.author_id, p.body, p.created_at
				FROM posts p
				JOIN follows f ON f.followee_id=p.author_id AND f.follower_id=$1
				JOIN users u ON u.id=p.author_id
				WHERE u.followers_count >= $2 AND p.created_at < $3 AND p.deleted_at IS NULL
				ORDER BY p.created_at DESC
				LIMIT $4;`, userID, a.celebThreshold, cursor, limit*2)
			if err != nil { return nil, err }
			defer rows.Close()
			return scanHomeRows(rows)
		}
	}()
	if err != nil { return nil, 0, err }

	// Merge by created_at DESC
	merged := mergeByTimeDesc(inboxRows, celebRows, limit)
	var next int64 = 0
	if len(merged) > 0 {
		next = merged[len(merged)-1].CreatedAt.UnixMilli()
	}
	return merged, next, nil
}

func scanHomeRows(rows pgxRows) ([]HomeItem, error) {
	var out []HomeItem
	for rows.Next() {
		var it HomeItem
		if err := rows.Scan(&it.PostID, &it.AuthorID, &it.Body, &it.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func mergeByTimeDesc(a, b []HomeItem, limit int) []HomeItem {
	out := make([]HomeItem, 0, limit)
	i, j := 0, 0
	for len(out) < limit && (i < len(a) || j < len(b)) {
		if j >= len(b) || (i < len(a) && a[i].CreatedAt.After(b[j].CreatedAt)) {
			out = append(out, a[i])
			i++
		} else {
			out = append(out, b[j])
			j++
		}
	}
	return out
}

/************* Fanout worker *************/

type fanoutJob struct {
	ID       int64
	PostID   int64
	AuthorID int64
	Created  time.Time
}

func (a *App) RunFanoutWorker(ctx context.Context) error {
	log.Printf("fanout worker: start, celeb threshold=%d", a.celebThreshold)
	t := time.NewTicker(WorkerPollInterval)
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

func (a *App) processFanoutBatch(ctx context.Context) error {
	tx, err := a.db.Begin(ctx); if err != nil { return err }
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, `
		SELECT id, post_id, author_id, created_at
		FROM fanout_queue
		WHERE status=0
		ORDER BY created_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED;`, WorkerFetchLimit)
	if err != nil { return err }
	defer rows.Close()

	var jobs []fanoutJob
	for rows.Next() {
		var j fanoutJob
		if err := rows.Scan(&j.ID, &j.PostID, &j.AuthorID, &j.Created); err != nil { return err }
		jobs = append(jobs, j)
	}
	if err := rows.Err(); err != nil { return err }
	if len(jobs)==0 { return tx.Commit(ctx) }

	for _, j := range jobs {
		if err := a.fanoutOne(ctx, tx, j); err != nil {
			log.Printf("fanout job %d failed: %v", j.ID, err)
			_, _ = tx.Exec(ctx, `UPDATE fanout_queue SET status=2, processed_at=now() WHERE id=$1;`, j.ID)
			continue
		}
		_, _ = tx.Exec(ctx, `UPDATE fanout_queue SET status=1, processed_at=now() WHERE id=$1;`, j.ID)
	}
	return tx.Commit(ctx)
}

func (a *App) fanoutOne(ctx context.Context, tx pgxTx, j fanoutJob) error {
	var followersCount int64
	if err := tx.QueryRow(ctx, `SELECT followers_count FROM users WHERE id=$1;`, j.AuthorID).Scan(&followersCount); err != nil {
		return err
	}
	if followersCount >= a.celebThreshold { return nil }

	var postCreated time.Time
	if err := tx.QueryRow(ctx, `SELECT created_at FROM posts WHERE id=$1;`, j.PostID).Scan(&postCreated); err != nil { return err }

	const page = 50_000
	var lastFollower int64 = 0
	for {
		followers, err := fetchFollowersPage(ctx, tx, j.AuthorID, lastFollower, page)
		if err != nil { return err }
		if len(followers) == 0 { break }

		if err := insertInboxBatch(ctx, tx, followers, j.PostID, j.AuthorID, postCreated); err != nil { return err }
		lastFollower = followers[len(followers)-1]
	}
	return nil
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
	for rows.Next() { var id int64; if err := rows.Scan(&id); err != nil { return nil, err }; ids = append(ids, id) }
	return ids, rows.Err()
}

func insertInboxBatch(ctx context.Context, tx pgxTx, userIDs []int64, postID, originAuthorID int64, createdAt time.Time) error {
	for start := 0; start < len(userIDs); start += FanoutBatchSize {
		end := start + FanoutBatchSize
		if end > len(userIDs) { end = len(userIDs) }
		b := &pgxBatch{}
		for _, uid := range userIDs[start:end] {
			b.Queue(`INSERT INTO home_inbox(user_id, post_id, origin_author_id, created_at)
				VALUES($1,$2,$3,$4)
				ON CONFLICT (user_id, post_id) DO NOTHING;`, uid, postID, originAuthorID, createdAt)
		}
		if err := b.Send(ctx, tx); err != nil { return err }
	}
	return nil
}

/************* Inbox trimming *************/

func (a *App) periodicInboxTrim(ctx context.Context, every time.Duration) {
	t := time.NewTicker(every)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if err := a.trimAllInboxes(ctx, a.inboxKeepMax); err != nil {
				log.Printf("inbox trim: %v", err)
			}
		}
	}
}

func (a *App) trimAllInboxes(ctx context.Context, keep int) error {
	_, err := a.db.Exec(ctx, `
		WITH ranked AS (
			SELECT user_id, post_id, created_at,
				row_number() over (PARTITION BY user_id ORDER BY created_at DESC) as rn
			FROM home_inbox
		), old AS (
			SELECT user_id, post_id FROM ranked WHERE rn > $1
		)
		DELETE FROM home_inbox hi
		USING old
		WHERE hi.user_id=old.user_id AND hi.post_id=old.post_id;`, keep)
	return err
}

/************* pgx helper types *************/

type pgxTx interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgxRows, error)
	QueryRow(context.Context, string, ...any) pgxRow
}

type pgxRow interface { Scan(...any) error }

type pgxRows interface {
	Next() bool
	Scan(...any) error
	Err() error
	Close()
}

type pgxBatch struct{ stmts []stmt }

type stmt struct { sql string; args []any }

func (b *pgxBatch) Queue(sql string, args ...any) { b.stmts = append(b.stmts, stmt{sql, args}) }
func (b *pgxBatch) Send(ctx context.Context, tx pgxTx) error {
	for _, s := range b.stmts {
		if _, err := tx.Exec(ctx, s.sql, s.args...); err != nil { return err }
	}
	b.stmts = b.stmts[:0]
	return nil
}
