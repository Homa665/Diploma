package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"startup-platform/internal/config"
	"startup-platform/internal/database"
	"startup-platform/internal/middleware"
)

func TestRegister(t *testing.T) {
	resetDB(t)

	form := url.Values{}
	form.Set("email", "test@test.com")
	form.Set("nickname", "tester")
	form.Set("password", "password123")
	form.Set("name", "Test User")
	form.Set("phone", "+375291234567")

	req := httptest.NewRequest("POST", "/api/register", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := executeRequest(testHandler.HandleRegister, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp["success"] != true {
		t.Fatal("expected success=true")
	}

	cookies := rr.Result().Cookies()
	var found bool
	for _, c := range cookies {
		if c.Name == "token" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected token cookie")
	}
}

func TestRegisterDuplicateEmail(t *testing.T) {
	resetDB(t)
	createTestUser(t, "dup@test.com", "existing", "user")

	form := url.Values{}
	form.Set("email", "dup@test.com")
	form.Set("nickname", "newuser")
	form.Set("password", "password123")

	req := httptest.NewRequest("POST", "/api/register", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := executeRequest(testHandler.HandleRegister, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rr.Code)
	}
}

func TestRegisterDuplicateNickname(t *testing.T) {
	resetDB(t)
	createTestUser(t, "other@test.com", "taken", "user")

	form := url.Values{}
	form.Set("email", "new@test.com")
	form.Set("nickname", "taken")
	form.Set("password", "password123")

	req := httptest.NewRequest("POST", "/api/register", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := executeRequest(testHandler.HandleRegister, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rr.Code)
	}
}

func TestRegisterEmptyFields(t *testing.T) {
	resetDB(t)

	form := url.Values{}
	form.Set("email", "")
	form.Set("nickname", "")
	form.Set("password", "")

	req := httptest.NewRequest("POST", "/api/register", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := executeRequest(testHandler.HandleRegister, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestRegisterShortPassword(t *testing.T) {
	resetDB(t)

	form := url.Values{}
	form.Set("email", "short@test.com")
	form.Set("nickname", "shortpw")
	form.Set("password", "12345")

	req := httptest.NewRequest("POST", "/api/register", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := executeRequest(testHandler.HandleRegister, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestRegisterInvalidEmail(t *testing.T) {
	resetDB(t)

	form := url.Values{}
	form.Set("email", "noemail")
	form.Set("nickname", "noatsign")
	form.Set("password", "password123")

	req := httptest.NewRequest("POST", "/api/register", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := executeRequest(testHandler.HandleRegister, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestRegisterEmailNoAt(t *testing.T) {
	resetDB(t)

	form := url.Values{}
	form.Set("email", "nodot@com")
	form.Set("nickname", "nodotuser")
	form.Set("password", "password123")

	req := httptest.NewRequest("POST", "/api/register", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := executeRequest(testHandler.HandleRegister, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestLogin(t *testing.T) {
	resetDB(t)
	createTestUser(t, "login@test.com", "loginuser", "user")

	form := url.Values{}
	form.Set("email", "login@test.com")
	form.Set("password", "password")

	req := httptest.NewRequest("POST", "/api/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := executeRequest(testHandler.HandleLogin, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp["success"] != true {
		t.Fatal("expected success=true")
	}
}

func TestLoginWrongPassword(t *testing.T) {
	resetDB(t)
	createTestUser(t, "wrong@test.com", "wrongpw", "user")

	form := url.Values{}
	form.Set("email", "wrong@test.com")
	form.Set("password", "wrongpassword")

	req := httptest.NewRequest("POST", "/api/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := executeRequest(testHandler.HandleLogin, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestLoginEmptyFields(t *testing.T) {
	resetDB(t)

	form := url.Values{}
	form.Set("email", "")
	form.Set("password", "")

	req := httptest.NewRequest("POST", "/api/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := executeRequest(testHandler.HandleLogin, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestLoginNonexistentUser(t *testing.T) {
	resetDB(t)

	form := url.Values{}
	form.Set("email", "nobody@test.com")
	form.Set("password", "password123")

	req := httptest.NewRequest("POST", "/api/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := executeRequest(testHandler.HandleLogin, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestLoginBlockedUser(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "blocked@test.com", "blocked", "user")
	testDB.Exec("UPDATE users SET is_blocked = TRUE WHERE id = $1", uid)

	form := url.Values{}
	form.Set("email", "blocked@test.com")
	form.Set("password", "password")

	req := httptest.NewRequest("POST", "/api/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := executeRequest(testHandler.HandleLogin, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestLogout(t *testing.T) {
	req := httptest.NewRequest("GET", "/logout", nil)
	rr := executeRequest(testHandler.HandleLogout, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d", rr.Code)
	}

	cookies := rr.Result().Cookies()
	for _, c := range cookies {
		if c.Name == "token" && c.MaxAge < 0 {
			return
		}
	}
	t.Fatal("expected token cookie to be cleared")
}

func TestJWTGeneration(t *testing.T) {
	token, err := middleware.GenerateJWT(testSecret, 1, "user")
	if err != nil {
		t.Fatal(err)
	}
	if token == "" {
		t.Fatal("empty token")
	}

	claims, err := middleware.ValidateJWT(testSecret, token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.UserID != 1 {
		t.Fatalf("expected user_id=1, got %d", claims.UserID)
	}
	if claims.Role != "user" {
		t.Fatalf("expected role=user, got %s", claims.Role)
	}
}

func TestJWTInvalidSignature(t *testing.T) {
	token, _ := middleware.GenerateJWT(testSecret, 1, "user")
	_, err := middleware.ValidateJWT("wrong_secret", token)
	if err == nil {
		t.Fatal("expected error for invalid signature")
	}
}

func TestJWTInvalidFormat(t *testing.T) {
	_, err := middleware.ValidateJWT(testSecret, "not.a.valid.token")
	if err == nil {
		t.Fatal("expected error for invalid format")
	}
}

func TestJWTMalformedToken(t *testing.T) {
	_, err := middleware.ValidateJWT(testSecret, "abc")
	if err == nil {
		t.Fatal("expected error for malformed token")
	}
}

func TestJWTMalformedPayload(t *testing.T) {
	_, err := middleware.ValidateJWT(testSecret, "abc.!!!invalid_base64.def")
	if err == nil {
		t.Fatal("expected error for malformed payload")
	}
}

func TestJWTExpiredToken(t *testing.T) {
	_, err := middleware.ValidateJWT(testSecret, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJyb2xlIjoidXNlciIsImV4cCI6MH0.xxx")
	if err == nil {
		t.Fatal("expected error for expired/invalid token")
	}
}

func TestAuthMiddlewareNoToken(t *testing.T) {
	resetDB(t)

	req := httptest.NewRequest("GET", "/feed", nil)
	rr := executeRequest(withPageAuth(testHandler.HandleFeedPage, testSecret), req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect 303, got %d", rr.Code)
	}
}

func TestAuthMiddlewareInvalidToken(t *testing.T) {
	resetDB(t)

	req := httptest.NewRequest("GET", "/feed", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: "invalid.token.here"})
	rr := executeRequest(withPageAuth(testHandler.HandleFeedPage, testSecret), req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect 303, got %d", rr.Code)
	}
}

func TestAuthMiddlewareBlockedUser(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "authblocked@test.com", "authblocked", "user")
	testDB.Exec("UPDATE users SET is_blocked = TRUE WHERE id = $1", uid)

	req := httptest.NewRequest("GET", "/feed", nil)
	req = authRequest(req, uid, "user")
	rr := executeRequest(withPageAuth(testHandler.HandleFeedPage, testSecret), req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect 303, got %d", rr.Code)
	}
}

func TestAuthMiddlewareDeletedUser(t *testing.T) {
	resetDB(t)

	req := httptest.NewRequest("GET", "/feed", nil)
	req = authRequest(req, 999999, "user")
	rr := executeRequest(withPageAuth(testHandler.HandleFeedPage, testSecret), req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect 303, got %d", rr.Code)
	}
}

func TestAPIAuthMiddlewareNoToken(t *testing.T) {
	resetDB(t)

	req := httptest.NewRequest("POST", "/api/rate", nil)
	rr := executeRequest(withAuth(testHandler.HandleRate, testSecret), req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestAPIAuthMiddlewareInvalidToken(t *testing.T) {
	resetDB(t)

	req := httptest.NewRequest("POST", "/api/rate", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: "bad.token.value"})
	rr := executeRequest(withAuth(testHandler.HandleRate, testSecret), req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestAPIAuthMiddlewareBlockedUser(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "apiblocked@test.com", "apiblocked", "user")
	testDB.Exec("UPDATE users SET is_blocked = TRUE WHERE id = $1", uid)

	form := url.Values{}
	form.Set("post_id", "1")
	form.Set("score", "5")

	req := httptest.NewRequest("POST", "/api/rate", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleRate, testSecret), req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestAdminMiddlewareDenied(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "notadmin@test.com", "notadmin", "user")

	req := httptest.NewRequest("POST", "/api/admin/block-user", nil)
	req = authRequest(req, uid, "user")
	rr := executeRequest(adminAuth(testHandler.HandleAdminBlockUser), req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestOptionalAuthNoToken(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	optionalAuth(testHandler.HandleHomePage)(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 without token, got %d", rr.Code)
	}
}

func TestOptionalAuthInvalidToken(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: "invalid.token.here"})
	rr := httptest.NewRecorder()
	optionalAuth(testHandler.HandleHomePage)(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 with invalid token, got %d", rr.Code)
	}
}

func TestOptionalAuthBlockedUser(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "optblocked@test.com", "optblocked", "user")
	testDB.Exec("UPDATE users SET is_blocked = TRUE WHERE id = $1", uid)

	req := httptest.NewRequest("GET", "/", nil)
	req = authRequest(req, uid, "user")
	rr := httptest.NewRecorder()
	optionalAuth(testHandler.HandleHomePage)(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for blocked user with optional auth, got %d", rr.Code)
	}
}

func TestOptionalAuthValidToken(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "optvalid@test.com", "optvalid", "user")

	req := httptest.NewRequest("GET", "/", nil)
	req = authRequest(req, uid, "user")
	rr := httptest.NewRecorder()
	optionalAuth(testHandler.HandleHomePage)(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect 303 for authed user on home, got %d", rr.Code)
	}
}

func TestGetUserIDNilContext(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	id := middleware.GetUserID(req)
	if id != 0 {
		t.Fatalf("expected 0, got %d", id)
	}
}

func TestGetUserRoleNilContext(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	role := middleware.GetUserRole(req)
	if role != "" {
		t.Fatalf("expected empty, got %s", role)
	}
}

func TestGetUserIDFromContext(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, 42)
	req = req.WithContext(ctx)
	id := middleware.GetUserID(req)
	if id != 42 {
		t.Fatalf("expected 42, got %d", id)
	}
}

func TestGetUserRoleFromContext(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), middleware.UserRoleKey, "admin")
	req = req.WithContext(ctx)
	role := middleware.GetUserRole(req)
	if role != "admin" {
		t.Fatalf("expected admin, got %s", role)
	}
}

func TestRegisterPage(t *testing.T) {
	req := httptest.NewRequest("GET", "/register", nil)
	rr := executeRequest(testHandler.HandleRegisterPage, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestLoginPage(t *testing.T) {
	req := httptest.NewRequest("GET", "/login", nil)
	rr := executeRequest(testHandler.HandleLoginPage, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestDatabaseConnect(t *testing.T) {
	dbURL := "postgres://postgres:12345678@localhost:5432/startup_platform_test?sslmode=disable"
	db, err := database.Connect(dbURL)
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
}

func TestDatabaseConnectInvalidURL(t *testing.T) {
	_, err := database.Connect("postgres://invalid:invalid@localhost:9999/nonexistent?sslmode=disable&connect_timeout=1")
	if err == nil {
		t.Fatal("expected error for invalid db")
	}
}

func TestDatabaseMigrate(t *testing.T) {
	err := database.Migrate(testDB)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDatabaseMigrateIdempotent(t *testing.T) {
	err := database.Migrate(testDB)
	if err != nil {
		t.Fatal(err)
	}
	err = database.Migrate(testDB)
	if err != nil {
		t.Fatal(err)
	}
}

func TestConfigLoad(t *testing.T) {
	cfg := config.Load()
	if cfg.Port == "" {
		t.Fatal("expected non-empty port")
	}
	if cfg.JWTSecret == "" {
		t.Fatal("expected non-empty jwt secret")
	}
	if cfg.DatabaseURL == "" {
		t.Fatal("expected non-empty db url")
	}
	if cfg.UploadDir == "" {
		t.Fatal("expected non-empty upload dir")
	}
}

func TestConfigLoadWithEnv(t *testing.T) {
	t.Setenv("PORT", "9999")
	t.Setenv("DATABASE_URL", "test_db_url")
	t.Setenv("JWT_SECRET", "test_secret")
	t.Setenv("UPLOAD_DIR", "/tmp/test")

	cfg := config.Load()
	if cfg.Port != "9999" {
		t.Fatalf("expected 9999, got %s", cfg.Port)
	}
	if cfg.DatabaseURL != "test_db_url" {
		t.Fatalf("expected test_db_url, got %s", cfg.DatabaseURL)
	}
	if cfg.JWTSecret != "test_secret" {
		t.Fatalf("expected test_secret, got %s", cfg.JWTSecret)
	}
	if cfg.UploadDir != "/tmp/test" {
		t.Fatalf("expected /tmp/test, got %s", cfg.UploadDir)
	}
}
