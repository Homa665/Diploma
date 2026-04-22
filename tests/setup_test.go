package tests

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	"startup-platform/internal/database"
	"startup-platform/internal/handlers"
	"startup-platform/internal/middleware"
)

var (
	testDB      *sql.DB
	testHandler *handlers.Handler
	testSecret  = "test_secret_key_for_jwt"
	testUploadDir string
)

func TestMain(m *testing.M) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:12345678@localhost:5432/startup_platform_test?sslmode=disable"
	}

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		panic("cannot open test db: " + err.Error())
	}
	if err := db.Ping(); err != nil {
		panic("cannot ping test db: " + err.Error())
	}
	testDB = db

	if err := database.Migrate(testDB); err != nil {
		panic("migration failed: " + err.Error())
	}

	tmpDir, _ := os.MkdirTemp("", "uploads_test")
	testUploadDir = tmpDir
	testHandler = handlers.NewHandler(testDB, "../web/templates", testSecret, tmpDir)

	code := m.Run()

	cleanupDB()
	testDB.Close()
	os.RemoveAll(tmpDir)
	os.Exit(code)
}

func cleanupDB() {
	tables := []string{
		"poll_votes", "poll_options", "polls",
		"comment_likes", "reposts", "blocked_users",
		"expert_applications", "subscriptions",
		"team_responses", "team_requests",
		"complaints", "notifications",
		"chat_messages", "friendships",
		"ratings", "comments", "files", "posts", "users",
	}
	for _, t := range tables {
		testDB.Exec("DELETE FROM " + t)
	}
}

func resetDB(t *testing.T) {
	t.Helper()
	cleanupDB()
}

func createTestUser(t *testing.T, email, nickname, role string) int {
	t.Helper()
	var id int
	err := testDB.QueryRow(
		`INSERT INTO users (email, nickname, password_hash, role, name)
		VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		email, nickname, "5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8", role, nickname,
	).Scan(&id)
	if err != nil {
		t.Fatal("createTestUser:", err)
	}
	return id
}

func createTestPost(t *testing.T, authorID int, title, category string) int {
	t.Helper()
	var id int
	err := testDB.QueryRow(
		"INSERT INTO posts (author_id, title, description, category) VALUES ($1, $2, $3, $4) RETURNING id",
		authorID, title, "Test description", category,
	).Scan(&id)
	if err != nil {
		t.Fatal("createTestPost:", err)
	}
	return id
}

func authRequest(r *http.Request, userID int, role string) *http.Request {
	token, _ := middleware.GenerateJWT(testSecret, userID, role)
	r.AddCookie(&http.Cookie{Name: "token", Value: token})
	return r
}

func executeRequest(handler http.HandlerFunc, req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

func withAuth(handler http.HandlerFunc, secret string) http.HandlerFunc {
	auth := middleware.APIAuthMiddleware(secret, testDB)
	return func(w http.ResponseWriter, r *http.Request) {
		auth(http.HandlerFunc(handler)).ServeHTTP(w, r)
	}
}

func withPageAuth(handler http.HandlerFunc, secret string) http.HandlerFunc {
	auth := middleware.AuthMiddleware(secret, testDB)
	return func(w http.ResponseWriter, r *http.Request) {
		auth(http.HandlerFunc(handler)).ServeHTTP(w, r)
	}
}

func adminAuth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth := middleware.APIAuthMiddleware(testSecret, testDB)
		auth(middleware.AdminMiddleware(http.HandlerFunc(handler))).ServeHTTP(w, r)
	}
}

func optionalAuth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth := middleware.OptionalAuthMiddleware(testSecret, testDB)
		auth(http.HandlerFunc(handler)).ServeHTTP(w, r)
	}
}

func itoa(i int) string {
	return strconv.Itoa(i)
}
