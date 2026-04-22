package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Connect(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func Migrate(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			phone VARCHAR(50) DEFAULT '',
			nickname VARCHAR(100) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			name VARCHAR(200) DEFAULT '',
			city VARCHAR(100) DEFAULT '',
			bio TEXT DEFAULT '',
			role VARCHAR(20) DEFAULT 'user',
			interests TEXT DEFAULT '',
			user_role2 VARCHAR(100) DEFAULT '',
			avatar_url VARCHAR(500) DEFAULT '',
			is_blocked BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS posts (
			id SERIAL PRIMARY KEY,
			author_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			title VARCHAR(500) NOT NULL,
			description TEXT NOT NULL,
			category VARCHAR(100) DEFAULT '',
			tags TEXT DEFAULT '',
			is_pinned BOOLEAN DEFAULT FALSE,
			pinned_until TIMESTAMP,
			comments_off BOOLEAN DEFAULT FALSE,
			is_hidden BOOLEAN DEFAULT FALSE,
			view_count INTEGER DEFAULT 0,
			is_premium BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS files (
			id SERIAL PRIMARY KEY,
			post_id INTEGER REFERENCES posts(id) ON DELETE CASCADE,
			filename VARCHAR(500) NOT NULL,
			file_path VARCHAR(1000) NOT NULL,
			file_type VARCHAR(100) DEFAULT '',
			file_size BIGINT DEFAULT 0,
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS comments (
			id SERIAL PRIMARY KEY,
			post_id INTEGER REFERENCES posts(id) ON DELETE CASCADE,
			author_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			parent_id INTEGER REFERENCES comments(id) ON DELETE CASCADE,
			content TEXT NOT NULL,
			like_count INTEGER DEFAULT 0,
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS ratings (
			id SERIAL PRIMARY KEY,
			post_id INTEGER REFERENCES posts(id) ON DELETE CASCADE,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			score INTEGER NOT NULL CHECK (score >= 1 AND score <= 10),
			review TEXT DEFAULT '',
			is_expert BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT NOW(),
			UNIQUE(post_id, user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS chat_messages (
			id SERIAL PRIMARY KEY,
			sender_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			receiver_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			content TEXT NOT NULL,
			is_read BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS notifications (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			type VARCHAR(50) NOT NULL,
			content TEXT NOT NULL,
			link VARCHAR(500) DEFAULT '',
			is_read BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS team_requests (
			id SERIAL PRIMARY KEY,
			post_id INTEGER REFERENCES posts(id) ON DELETE CASCADE,
			author_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			title VARCHAR(300) NOT NULL,
			description TEXT DEFAULT '',
			skills TEXT DEFAULT '',
			role_needed VARCHAR(100) DEFAULT '',
			is_open BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS team_responses (
			id SERIAL PRIMARY KEY,
			request_id INTEGER REFERENCES team_requests(id) ON DELETE CASCADE,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			message TEXT DEFAULT '',
			status VARCHAR(20) DEFAULT 'pending',
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS complaints (
			id SERIAL PRIMARY KEY,
			author_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			target_type VARCHAR(20) NOT NULL,
			target_id INTEGER NOT NULL,
			category VARCHAR(100) DEFAULT '',
			description TEXT NOT NULL,
			status VARCHAR(20) DEFAULT 'pending',
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS subscriptions (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			plan VARCHAR(50) NOT NULL,
			status VARCHAR(20) DEFAULT 'active',
			expires_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS friendships (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			friend_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			status VARCHAR(20) DEFAULT 'pending',
			created_at TIMESTAMP DEFAULT NOW(),
			UNIQUE(user_id, friend_id)
		)`,
		`CREATE TABLE IF NOT EXISTS follows (
			id SERIAL PRIMARY KEY,
			follower_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			following_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			created_at TIMESTAMP DEFAULT NOW(),
			UNIQUE(follower_id, following_id)
		)`,
		`CREATE TABLE IF NOT EXISTS polls (
			id SERIAL PRIMARY KEY,
			post_id INTEGER REFERENCES posts(id) ON DELETE CASCADE,
			question TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS poll_options (
			id SERIAL PRIMARY KEY,
			poll_id INTEGER REFERENCES polls(id) ON DELETE CASCADE,
			text VARCHAR(500) NOT NULL,
			vote_count INTEGER DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS poll_votes (
			id SERIAL PRIMARY KEY,
			poll_id INTEGER REFERENCES polls(id) ON DELETE CASCADE,
			option_id INTEGER REFERENCES poll_options(id) ON DELETE CASCADE,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			UNIQUE(poll_id, user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS comment_likes (
			id SERIAL PRIMARY KEY,
			comment_id INTEGER REFERENCES comments(id) ON DELETE CASCADE,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			UNIQUE(comment_id, user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS reposts (
			id SERIAL PRIMARY KEY,
			post_id INTEGER REFERENCES posts(id) ON DELETE CASCADE,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			comment TEXT DEFAULT '',
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS blocked_users (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			blocked_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			created_at TIMESTAMP DEFAULT NOW(),
			UNIQUE(user_id, blocked_id)
		)`,
		`CREATE TABLE IF NOT EXISTS expert_applications (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			portfolio TEXT DEFAULT '',
			description TEXT DEFAULT '',
			status VARCHAR(20) DEFAULT 'pending',
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS activity_logs (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
			action VARCHAR(100) NOT NULL,
			target_type VARCHAR(50) DEFAULT '',
			target_id INTEGER DEFAULT 0,
			details TEXT DEFAULT '',
			ip_address VARCHAR(50) DEFAULT '',
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS user_bans (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			admin_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
			reason TEXT NOT NULL,
			restriction VARCHAR(50) NOT NULL,
			target_link VARCHAR(500) DEFAULT '',
			admin_comment TEXT DEFAULT '',
			expires_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS projects (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			title VARCHAR(300) NOT NULL,
			description TEXT DEFAULT '',
			cover_url VARCHAR(500) DEFAULT '',
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS project_posts (
			id SERIAL PRIMARY KEY,
			project_id INTEGER REFERENCES projects(id) ON DELETE CASCADE,
			post_id INTEGER REFERENCES posts(id) ON DELETE CASCADE,
			sort_order INTEGER DEFAULT 0,
			added_at TIMESTAMP DEFAULT NOW(),
			UNIQUE(project_id, post_id)
		)`,
		`CREATE TABLE IF NOT EXISTS expert_consultations (
			id SERIAL PRIMARY KEY,
			expert_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			client_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			post_id INTEGER REFERENCES posts(id) ON DELETE SET NULL,
			title VARCHAR(300) NOT NULL,
			description TEXT DEFAULT '',
			price INTEGER DEFAULT 0,
			status VARCHAR(20) DEFAULT 'pending',
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS consultation_messages (
			id SERIAL PRIMARY KEY,
			consultation_id INTEGER REFERENCES expert_consultations(id) ON DELETE CASCADE,
			sender_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			content TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS user_category_views (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			category VARCHAR(100) NOT NULL,
			view_count INTEGER DEFAULT 1,
			UNIQUE(user_id, category)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_posts_author ON posts(author_id)`,
		`CREATE INDEX IF NOT EXISTS idx_posts_category ON posts(category)`,
		`CREATE INDEX IF NOT EXISTS idx_posts_created ON posts(created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_comments_post ON comments(post_id)`,
		`CREATE INDEX IF NOT EXISTS idx_ratings_post ON ratings(post_id)`,
		`CREATE INDEX IF NOT EXISTS idx_chat_sender ON chat_messages(sender_id)`,
		`CREATE INDEX IF NOT EXISTS idx_chat_receiver ON chat_messages(receiver_id)`,
		`CREATE INDEX IF NOT EXISTS idx_notifications_user ON notifications(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_team_requests_post ON team_requests(post_id)`,
		`CREATE INDEX IF NOT EXISTS idx_complaints_status ON complaints(status)`,
		`CREATE INDEX IF NOT EXISTS idx_follows_follower ON follows(follower_id)`,
		`CREATE INDEX IF NOT EXISTS idx_follows_following ON follows(following_id)`,
		`CREATE INDEX IF NOT EXISTS idx_activity_logs_user ON activity_logs(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_activity_logs_created ON activity_logs(created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_user_bans_user ON user_bans(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_projects_user ON projects(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_project_posts_project ON project_posts(project_id)`,
		`CREATE INDEX IF NOT EXISTS idx_user_category_views_user ON user_category_views(user_id)`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return fmt.Errorf("migration failed: %w\nQuery: %s", err, q)
		}
	}

	// Migrate TIMESTAMP columns to TIMESTAMPTZ for correct timezone handling
	tzMigrations := []string{
		`ALTER TABLE users ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE current_setting('TimeZone')`,
		`ALTER TABLE users ALTER COLUMN updated_at TYPE TIMESTAMPTZ USING updated_at AT TIME ZONE current_setting('TimeZone')`,
		`ALTER TABLE posts ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE current_setting('TimeZone')`,
		`ALTER TABLE posts ALTER COLUMN updated_at TYPE TIMESTAMPTZ USING updated_at AT TIME ZONE current_setting('TimeZone')`,
		`ALTER TABLE posts ALTER COLUMN pinned_until TYPE TIMESTAMPTZ USING pinned_until AT TIME ZONE current_setting('TimeZone')`,
		`ALTER TABLE files ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE current_setting('TimeZone')`,
		`ALTER TABLE comments ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE current_setting('TimeZone')`,
		`ALTER TABLE ratings ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE current_setting('TimeZone')`,
		`ALTER TABLE chat_messages ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE current_setting('TimeZone')`,
		`ALTER TABLE notifications ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE current_setting('TimeZone')`,
		`ALTER TABLE team_requests ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE current_setting('TimeZone')`,
		`ALTER TABLE team_responses ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE current_setting('TimeZone')`,
		`ALTER TABLE complaints ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE current_setting('TimeZone')`,
		`ALTER TABLE subscriptions ALTER COLUMN expires_at TYPE TIMESTAMPTZ USING expires_at AT TIME ZONE current_setting('TimeZone')`,
		`ALTER TABLE subscriptions ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE current_setting('TimeZone')`,
		`ALTER TABLE friendships ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE current_setting('TimeZone')`,
		`ALTER TABLE follows ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE current_setting('TimeZone')`,
		`ALTER TABLE polls ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE current_setting('TimeZone')`,
		`ALTER TABLE reposts ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE current_setting('TimeZone')`,
		`ALTER TABLE blocked_users ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE current_setting('TimeZone')`,
		`ALTER TABLE expert_applications ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE current_setting('TimeZone')`,
		`ALTER TABLE activity_logs ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE current_setting('TimeZone')`,
		`ALTER TABLE user_bans ALTER COLUMN expires_at TYPE TIMESTAMPTZ USING expires_at AT TIME ZONE current_setting('TimeZone')`,
		`ALTER TABLE user_bans ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE current_setting('TimeZone')`,
		`ALTER TABLE projects ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE current_setting('TimeZone')`,
		`ALTER TABLE projects ALTER COLUMN updated_at TYPE TIMESTAMPTZ USING updated_at AT TIME ZONE current_setting('TimeZone')`,
		`ALTER TABLE project_posts ALTER COLUMN added_at TYPE TIMESTAMPTZ USING added_at AT TIME ZONE current_setting('TimeZone')`,
		`ALTER TABLE expert_consultations ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE current_setting('TimeZone')`,
		`ALTER TABLE consultation_messages ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE current_setting('TimeZone')`,
	}
	for _, q := range tzMigrations {
		db.Exec(q) // ignore errors - column may already be TIMESTAMPTZ
	}

	// Migrate ratings from 1-10 to 1-5 scale
	ratingMigrations := []string{
		`ALTER TABLE ratings DROP CONSTRAINT IF EXISTS ratings_score_check`,
		`UPDATE ratings SET score = GREATEST(1, ROUND(score / 2.0)::int) WHERE score > 5`,
		`ALTER TABLE ratings ADD CONSTRAINT ratings_score_check CHECK (score >= 1 AND score <= 5)`,
	}
	for _, q := range ratingMigrations {
		db.Exec(q)
	}

	// Add file columns to consultation_messages
	db.Exec(`ALTER TABLE consultation_messages ADD COLUMN IF NOT EXISTS file_path TEXT DEFAULT ''`)
	db.Exec(`ALTER TABLE consultation_messages ADD COLUMN IF NOT EXISTS file_name TEXT DEFAULT ''`)

	return nil
}
