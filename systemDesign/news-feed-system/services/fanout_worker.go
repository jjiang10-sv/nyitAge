package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gocql/gocql"
	"github.com/google/uuid"
)

// FanoutWorker handles the distribution of posts to user timelines
type FanoutWorker struct {
	session            *gocql.Session
	redis              *redis.Client
	socialService      *SocialGraphService
	userService        *UserProfileService
	celebrityThreshold int64
	batchSize          int
	maxRetries         int
	workerID           string
	isRunning          bool
	stopChan           chan struct{}
	wg                 sync.WaitGroup
}

// FanoutStats represents fanout processing statistics
type FanoutStats struct {
	ProcessedJobs   int64         `json:"processed_jobs"`
	SuccessfulJobs  int64         `json:"successful_jobs"`
	FailedJobs      int64         `json:"failed_jobs"`
	SkippedJobs     int64         `json:"skipped_jobs"`
	TotalInserts    int64         `json:"total_inserts"`
	ProcessingTime  time.Duration `json:"processing_time"`
	LastProcessedAt time.Time     `json:"last_processed_at"`
}

// NewFanoutWorker creates a new fanout worker
func NewFanoutWorker(session *gocql.Session, redisClient *redis.Client,
	socialService *SocialGraphService, userService *UserProfileService, workerID string) *FanoutWorker {
	return &FanoutWorker{
		session:            session,
		redis:              redisClient,
		socialService:      socialService,
		userService:        userService,
		celebrityThreshold: 100000, // 100K followers = celebrity
		batchSize:          5000,   // Timeline inserts per batch
		maxRetries:         3,
		workerID:           workerID,
		stopChan:           make(chan struct{}),
	}
}

// Start begins the fanout worker processing loop
func (w *FanoutWorker) Start(ctx context.Context) error {
	if w.isRunning {
		return fmt.Errorf("worker is already running")
	}

	w.isRunning = true
	w.wg.Add(1)

	go func() {
		defer w.wg.Done()
		w.processLoop(ctx)
	}()

	fmt.Printf("Fanout worker %s started\n", w.workerID)
	return nil
}

// Stop gracefully stops the fanout worker
func (w *FanoutWorker) Stop() error {
	if !w.isRunning {
		return fmt.Errorf("worker is not running")
	}

	close(w.stopChan)
	w.wg.Wait()
	w.isRunning = false

	fmt.Printf("Fanout worker %s stopped\n", w.workerID)
	return nil
}

// processLoop is the main processing loop for the worker
func (w *FanoutWorker) processLoop(ctx context.Context) {
	ticker := time.NewTicker(500 * time.Millisecond) // Poll every 500ms
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stopChan:
			return
		case <-ticker.C:
			if err := w.processBatch(ctx); err != nil {
				fmt.Printf("Worker %s error: %v\n", w.workerID, err)
				// Exponential backoff on error
				time.Sleep(time.Second * 2)
			}
		}
	}
}

// processBatch processes a batch of fanout jobs
func (w *FanoutWorker) processBatch(ctx context.Context) error {
	// Get pending jobs from queue
	jobs, err := w.claimJobs(ctx, 10) // Process up to 10 jobs per batch
	if err != nil {
		return fmt.Errorf("failed to claim jobs: %w", err)
	}

	if len(jobs) == 0 {
		return nil // No jobs to process
	}

	stats := &FanoutStats{
		LastProcessedAt: time.Now(),
	}

	// Process each job
	for _, job := range jobs {
		startTime := time.Now()

		err := w.processJob(ctx, job, stats)

		processingTime := time.Since(startTime)
		stats.ProcessingTime += processingTime
		stats.ProcessedJobs++

		if err != nil {
			fmt.Printf("Worker %s failed to process job %s: %v\n", w.workerID, job.QueueID, err)
			w.markJobFailed(ctx, job, err.Error())
			stats.FailedJobs++
		} else {
			w.markJobCompleted(ctx, job)
			stats.SuccessfulJobs++
		}
	}

	// Log stats
	if stats.ProcessedJobs > 0 {
		fmt.Printf("Worker %s processed %d jobs (%d successful, %d failed, %d inserts) in %v\n",
			w.workerID, stats.ProcessedJobs, stats.SuccessfulJobs, stats.FailedJobs,
			stats.TotalInserts, stats.ProcessingTime)
	}

	return nil
}

// claimJobs claims pending jobs from the fanout queue
func (w *FanoutWorker) claimJobs(ctx context.Context, limit int) ([]*FanoutQueueItem, error) {
	// Query pending jobs ordered by priority and creation time
	query := `SELECT status, created_at, queue_id, post_id, author_id, priority, retry_count
	          FROM fanout_queue_by_status 
	          WHERE status = 'pending' 
	          ORDER BY priority DESC, created_at ASC 
	          LIMIT ?`

	iter := w.session.Query(query, limit).WithContext(ctx).Iter()
	defer iter.Close()

	var jobs []*FanoutQueueItem
	var status string
	var createdAt time.Time
	var queueID uuid.UUID
	var postID uuid.UUID
	var authorID int64
	var priority, retryCount int

	for iter.Scan(&status, &createdAt, &queueID, &postID, &authorID, &priority, &retryCount) {
		job := &FanoutQueueItem{
			QueueID:    queueID,
			PostID:     postID,
			AuthorID:   authorID,
			CreatedAt:  createdAt,
			Status:     status,
			RetryCount: retryCount,
			Priority:   priority,
		}
		jobs = append(jobs, job)
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	// Mark jobs as processing
	for _, job := range jobs {
		w.markJobProcessing(ctx, job)
	}

	return jobs, nil
}

// processJob processes a single fanout job
func (w *FanoutWorker) processJob(ctx context.Context, job *FanoutQueueItem, stats *FanoutStats) error {
	// Get author information
	author, err := w.userService.GetUserByID(ctx, job.AuthorID)
	if err != nil {
		return fmt.Errorf("failed to get author: %w", err)
	}

	// Get post information
	post, err := w.getPostFromDB(ctx, job.PostID)
	if err != nil {
		return fmt.Errorf("failed to get post: %w", err)
	}

	// Check if author is celebrity
	if author.IsCelebrity || author.FollowersCount >= w.celebrityThreshold {
		// For celebrities, store in celebrity_posts table instead of fanout
		return w.processCelebrityPost(ctx, post, stats)
	}

	// For regular users, fanout to followers
	return w.processRegularUserPost(ctx, post, stats)
}

// processCelebrityPost handles posts from celebrity users
func (w *FanoutWorker) processCelebrityPost(ctx context.Context, post *Post, stats *FanoutStats) error {
	// Insert into celebrity_posts table for pull-based feed generation
	query := `INSERT INTO celebrity_posts (author_id, created_at, post_id, content, 
	                                      engagement_score, like_count, comment_count, share_count) 
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	err := w.session.Query(query, post.AuthorID, post.CreatedAt, post.PostID, post.Content,
		post.EngagementScore, post.LikeCount, post.CommentCount, post.ShareCount).WithContext(ctx).Exec()

	if err != nil {
		return fmt.Errorf("failed to insert celebrity post: %w", err)
	}

	stats.SkippedJobs++ // Skipped fanout, but processed successfully
	return nil
}

// processRegularUserPost handles posts from regular users with push-based fanout
func (w *FanoutWorker) processRegularUserPost(ctx context.Context, post *Post, stats *FanoutStats) error {
	// Get followers in batches
	offset := 0
	batchSize := 1000 // Followers per batch

	for {
		followers, err := w.socialService.GetFollowers(ctx, post.AuthorID, batchSize, offset)
		if err != nil {
			return fmt.Errorf("failed to get followers: %w", err)
		}

		if len(followers) == 0 {
			break // No more followers
		}

		// Insert into timelines in batches
		err = w.insertTimelineBatch(ctx, followers, post)
		if err != nil {
			return fmt.Errorf("failed to insert timeline batch: %w", err)
		}

		stats.TotalInserts += int64(len(followers))
		offset += len(followers)

		// If we got fewer followers than requested, we're done
		if len(followers) < batchSize {
			break
		}
	}

	return nil
}

// insertTimelineBatch inserts posts into multiple user timelines in batches
func (w *FanoutWorker) insertTimelineBatch(ctx context.Context, followers []*UserConnection, post *Post) error {
	// Process followers in smaller batches for database efficiency
	for i := 0; i < len(followers); i += w.batchSize {
		end := i + w.batchSize
		if end > len(followers) {
			end = len(followers)
		}

		batch := followers[i:end]
		if err := w.insertTimelineSubBatch(ctx, batch, post); err != nil {
			return err
		}
	}

	return nil
}

// insertTimelineSubBatch inserts a sub-batch of timeline entries
func (w *FanoutWorker) insertTimelineSubBatch(ctx context.Context, followers []*UserConnection, post *Post) error {
	// Use batch insert for better performance
	batch := w.session.NewBatch(gocql.LoggedBatch)

	insertedAt := time.Now()
	query := `INSERT INTO user_timeline (user_id, created_at, post_id, author_id, content, 
	                                    engagement_score, inserted_at) 
	          VALUES (?, ?, ?, ?, ?, ?, ?)`

	for _, follower := range followers {
		batch.Query(query, follower.UserID, post.CreatedAt, post.PostID, post.AuthorID,
			post.Content, post.EngagementScore, insertedAt)
	}

	// Execute batch with retry logic
	var err error
	for retry := 0; retry < w.maxRetries; retry++ {
		err = w.session.ExecuteBatch(batch.WithContext(ctx))
		if err == nil {
			break
		}

		// Exponential backoff
		time.Sleep(time.Duration(retry+1) * time.Second)
	}

	if err != nil {
		return fmt.Errorf("failed to execute timeline batch after %d retries: %w", w.maxRetries, err)
	}

	// Update Redis cache for active users
	go w.updateTimelineCache(context.Background(), followers, post)

	return nil
}

// updateTimelineCache updates Redis cache for active users
func (w *FanoutWorker) updateTimelineCache(ctx context.Context, followers []*UserConnection, post *Post) {
	pipe := w.redis.Pipeline()

	for _, follower := range followers {
		// Add to user's timeline cache (keep last 100 posts)
		timelineKey := fmt.Sprintf("timeline:%d", follower.UserID)

		// Add post with timestamp as score
		pipe.ZAdd(ctx, timelineKey, &redis.Z{
			Score:  float64(post.CreatedAt.Unix()),
			Member: post.PostID.String(),
		})

		// Keep only last 100 posts
		pipe.ZRemRangeByRank(ctx, timelineKey, 0, -101)

		// Set expiration
		pipe.Expire(ctx, timelineKey, 24*time.Hour)
	}

	// Execute pipeline
	_, err := pipe.Exec(ctx)
	if err != nil {
		fmt.Printf("Failed to update timeline cache: %v\n", err)
	}
}

// markJobProcessing marks a job as being processed
func (w *FanoutWorker) markJobProcessing(ctx context.Context, job *FanoutQueueItem) {
	// Update main queue table
	query := `UPDATE fanout_queue SET status = 'processing', processed_at = ? WHERE queue_id = ?`
	w.session.Query(query, time.Now(), job.QueueID).WithContext(ctx).Exec()

	// Remove from pending status table
	deleteQuery := `DELETE FROM fanout_queue_by_status WHERE status = 'pending' AND created_at = ? AND queue_id = ?`
	w.session.Query(deleteQuery, job.CreatedAt, job.QueueID).WithContext(ctx).Exec()

	// Add to processing status table
	insertQuery := `INSERT INTO fanout_queue_by_status (status, created_at, queue_id, post_id, author_id, priority, retry_count) 
	                VALUES (?, ?, ?, ?, ?, ?, ?)`
	w.session.Query(insertQuery, "processing", time.Now(), job.QueueID, job.PostID,
		job.AuthorID, job.Priority, job.RetryCount).WithContext(ctx).Exec()
}

// markJobCompleted marks a job as completed
func (w *FanoutWorker) markJobCompleted(ctx context.Context, job *FanoutQueueItem) {
	// Update main queue table
	query := `UPDATE fanout_queue SET status = 'completed', processed_at = ? WHERE queue_id = ?`
	w.session.Query(query, time.Now(), job.QueueID).WithContext(ctx).Exec()

	// Remove from processing status table
	deleteQuery := `DELETE FROM fanout_queue_by_status WHERE status = 'processing' AND queue_id = ?`
	w.session.Query(deleteQuery, job.QueueID).WithContext(ctx).Exec()

	// Optionally add to completed table with TTL for cleanup
	insertQuery := `INSERT INTO fanout_queue_by_status (status, created_at, queue_id, post_id, author_id, priority, retry_count) 
	                VALUES (?, ?, ?, ?, ?, ?, ?) USING TTL 86400` // 24 hours TTL
	w.session.Query(insertQuery, "completed", time.Now(), job.QueueID, job.PostID,
		job.AuthorID, job.Priority, job.RetryCount).WithContext(ctx).Exec()
}

// markJobFailed marks a job as failed and handles retry logic
func (w *FanoutWorker) markJobFailed(ctx context.Context, job *FanoutQueueItem, errorMessage string) {
	job.RetryCount++

	if job.RetryCount < w.maxRetries {
		// Retry the job with exponential backoff
		retryDelay := time.Duration(job.RetryCount*job.RetryCount) * time.Second
		retryAt := time.Now().Add(retryDelay)

		// Update main queue table
		query := `UPDATE fanout_queue SET status = 'pending', retry_count = ?, error_message = ? WHERE queue_id = ?`
		w.session.Query(query, job.RetryCount, errorMessage, job.QueueID).WithContext(ctx).Exec()

		// Remove from processing status table
		deleteQuery := `DELETE FROM fanout_queue_by_status WHERE status = 'processing' AND queue_id = ?`
		w.session.Query(deleteQuery, job.QueueID).WithContext(ctx).Exec()

		// Add back to pending with retry delay
		insertQuery := `INSERT INTO fanout_queue_by_status (status, created_at, queue_id, post_id, author_id, priority, retry_count) 
		                VALUES (?, ?, ?, ?, ?, ?, ?)`
		w.session.Query(insertQuery, "pending", retryAt, job.QueueID, job.PostID,
			job.AuthorID, job.Priority, job.RetryCount).WithContext(ctx).Exec()

		fmt.Printf("Job %s scheduled for retry %d/%d in %v\n", job.QueueID, job.RetryCount, w.maxRetries, retryDelay)
	} else {
		// Max retries exceeded, mark as failed permanently
		query := `UPDATE fanout_queue SET status = 'failed', processed_at = ?, retry_count = ?, error_message = ? WHERE queue_id = ?`
		w.session.Query(query, time.Now(), job.RetryCount, errorMessage, job.QueueID).WithContext(ctx).Exec()

		// Remove from processing status table
		deleteQuery := `DELETE FROM fanout_queue_by_status WHERE status = 'processing' AND queue_id = ?`
		w.session.Query(deleteQuery, job.QueueID).WithContext(ctx).Exec()

		// Add to failed table
		insertQuery := `INSERT INTO fanout_queue_by_status (status, created_at, queue_id, post_id, author_id, priority, retry_count) 
		                VALUES (?, ?, ?, ?, ?, ?, ?)`
		w.session.Query(insertQuery, "failed", time.Now(), job.QueueID, job.PostID,
			job.AuthorID, job.Priority, job.RetryCount).WithContext(ctx).Exec()

		fmt.Printf("Job %s permanently failed after %d retries: %s\n", job.QueueID, job.RetryCount, errorMessage)
	}
}

// getPostFromDB retrieves post information from database
func (w *FanoutWorker) getPostFromDB(ctx context.Context, postID uuid.UUID) (*Post, error) {
	query := `SELECT post_id, author_id, content, created_at, updated_at, is_deleted,
	                 engagement_score, like_count, comment_count, share_count, visibility
	          FROM posts WHERE post_id = ?`

	var post Post
	err := w.session.Query(query, postID).WithContext(ctx).Scan(
		&post.PostID, &post.AuthorID, &post.Content, &post.CreatedAt, &post.UpdatedAt,
		&post.IsDeleted, &post.EngagementScore, &post.LikeCount, &post.CommentCount,
		&post.ShareCount, &post.Visibility,
	)

	if err != nil {
		return nil, err
	}

	return &post, nil
}

// GetQueueStats returns statistics about the fanout queue
func (w *FanoutWorker) GetQueueStats(ctx context.Context) (map[string]int64, error) {
	stats := make(map[string]int64)

	// Count jobs by status
	statuses := []string{"pending", "processing", "completed", "failed"}

	for _, status := range statuses {
		query := `SELECT COUNT(*) FROM fanout_queue_by_status WHERE status = ?`
		var count int64
		err := w.session.Query(query, status).WithContext(ctx).Scan(&count)
		if err != nil {
			return nil, fmt.Errorf("failed to get %s count: %w", status, err)
		}
		stats[status] = count
	}

	return stats, nil
}

// CleanupCompletedJobs removes old completed and failed jobs
func (w *FanoutWorker) CleanupCompletedJobs(ctx context.Context, olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)

	// Delete old completed jobs
	deleteCompleted := `DELETE FROM fanout_queue_by_status WHERE status = 'completed' AND created_at < ?`
	err := w.session.Query(deleteCompleted, cutoff).WithContext(ctx).Exec()
	if err != nil {
		return fmt.Errorf("failed to cleanup completed jobs: %w", err)
	}

	// Delete old failed jobs
	deleteFailed := `DELETE FROM fanout_queue_by_status WHERE status = 'failed' AND created_at < ?`
	err = w.session.Query(deleteFailed, cutoff).WithContext(ctx).Exec()
	if err != nil {
		return fmt.Errorf("failed to cleanup failed jobs: %w", err)
	}

	fmt.Printf("Cleaned up jobs older than %v\n", olderThan)
	return nil
}
