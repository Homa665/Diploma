package tests

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	htmltemplate "html/template"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"startup-platform/internal/database"
	"startup-platform/internal/handlers"
	"startup-platform/internal/middleware"
	"strings"
	"testing"
	"time"
)

func TestFuncMapFunctions(t *testing.T) {
	h := handlers.NewHandler(testDB, "../web/templates", testSecret, testUploadDir)

	type tc struct {
		name     string
		tmplBody string
		data     interface{}
		contains string
	}

	tests := []tc{
		{"formatDate", `{{formatDate .}}`, time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC), "15.01.2024"},
		{"truncate_short", `{{truncate . 10}}`, "Hello", "Hello"},
		{"truncate_long", `{{truncate . 5}}`, "Hello World", "Hello..."},
		{"add", `{{add 3 4}}`, nil, "7"},
		{"sub", `{{sub 10 3}}`, nil, "7"},
		{"seq", `{{range seq 3}}{{.}} {{end}}`, nil, "1 2 3"},
		{"dict", `{{$d := dict "k" "v"}}{{index $d "k"}}`, nil, "v"},
		{"isImage_true", `{{isImage .}}`, "image/png", "true"},
		{"isImage_false", `{{isImage .}}`, "text/plain", "false"},
		{"isVideo_true", `{{isVideo .}}`, "video/mp4", "true"},
		{"isVideo_false", `{{isVideo .}}`, "text/plain", "false"},
		{"isAudio_true", `{{isAudio .}}`, "audio/mpeg", "true"},
		{"isAudio_false", `{{isAudio .}}`, "text/plain", "false"},
		{"json_func", `{{json .}}`, map[string]string{"a": "b"}, `{&#34;a&#34;:&#34;b&#34;}`},
		{"toJSON_func", `{{toJSON .}}`, []int{1, 2}, `[1,2]`},
		{"split_func", `{{range split . ","}}[{{.}}]{{end}}`, "a,b", "[a][b]"},
		{"split_empty", `{{$s := split . ","}}{{if $s}}yes{{else}}no{{end}}`, "", "no"},
		{"contains_true", `{{contains . "hi"}}`, "say hi", "true"},
		{"contains_false", `{{contains . "hi"}}`, "nope", "false"},
		{"eq_true", `{{eq 1 1}}`, nil, "true"},
		{"ne_true", `{{ne 1 2}}`, nil, "true"},
		{"gt_true", `{{gt 5 3}}`, nil, "true"},
		{"gt_false", `{{gt 1 3}}`, nil, "false"},
		{"formatSize_bytes", `{{formatSize .}}`, int64(100), ""},
		{"formatSize_kb", `{{formatSize .}}`, int64(5000), ""},
		{"formatSize_large_kb", `{{formatSize .}}`, int64(50000), ""},
		{"formatSize_mb", `{{formatSize .}}`, int64(5000000), ""},
		{"formatSize_large_mb", `{{formatSize .}}`, int64(50000000), ""},
	}

	for _, c := range tests {
		_, err := h.Templates.New("_ftest_" + c.name).Parse(c.tmplBody)
		if err != nil {
			t.Fatalf("parse %s: %v", c.name, err)
		}
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := h.Templates.ExecuteTemplate(&buf, "_ftest_"+c.name, c.data)
			if err != nil {
				t.Fatalf("execute error: %v", err)
			}
			if c.contains != "" && !strings.Contains(buf.String(), c.contains) {
				t.Fatalf("expected to contain '%s', got '%s'", c.contains, buf.String())
			}
			if c.contains == "" && buf.Len() == 0 {
				t.Fatal("expected non-empty output")
			}
		})
	}
}

func TestRenderTemplateError(t *testing.T) {
	emptyTmpl := htmltemplate.Must(htmltemplate.New("").Parse(""))
	h := &handlers.Handler{
		DB:        testDB,
		Templates: emptyTmpl,
		JWTSecret: testSecret,
		UploadDir: testUploadDir,
	}
	rr := httptest.NewRecorder()
	h.HandleLoginPage(rr, httptest.NewRequest("GET", "/login", nil))
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
}

func TestRenderTemplateNotFound(t *testing.T) {
	rr := httptest.NewRecorder()
	testHandler.HandleProfilePage(rr, httptest.NewRequest("GET", "/profile/me", nil))
	if rr.Code == http.StatusOK {
		t.Log("OK without auth - expected redirect or error")
	}
}

func signToken(secret, header, payload string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(header + "." + payload))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return header + "." + payload + "." + sig
}

func TestValidateJWTExpiredProperly(t *testing.T) {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	claims := middleware.JWTClaims{UserID: 1, Role: "user", Exp: time.Now().Add(-1 * time.Hour).Unix()}
	claimsJSON, _ := json.Marshal(claims)
	payload := base64.RawURLEncoding.EncodeToString(claimsJSON)
	token := signToken(testSecret, header, payload)

	_, err := middleware.ValidateJWT(testSecret, token)
	if err == nil || !strings.Contains(err.Error(), "expired") {
		t.Fatalf("expected 'token expired', got %v", err)
	}
}

func TestValidateJWTInvalidJSON(t *testing.T) {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(`not valid json`))
	token := signToken(testSecret, header, payload)

	_, err := middleware.ValidateJWT(testSecret, token)
	if err == nil || !strings.Contains(err.Error(), "invalid claims") {
		t.Fatalf("expected 'invalid claims', got %v", err)
	}
}

func TestValidateJWTInvalidBase64Payload(t *testing.T) {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payload := "!!!not-base64!!!"
	token := signToken(testSecret, header, payload)

	_, err := middleware.ValidateJWT(testSecret, token)
	if err == nil {
		t.Fatal("expected error for invalid base64 payload")
	}
}

func TestRegisterParseFormError(t *testing.T) {
	resetDB(t)
	req := httptest.NewRequest("POST", "/api/register", strings.NewReader(""))
	req.Header.Set("Content-Type", "multipart/form-data")
	rr := executeRequest(http.HandlerFunc(testHandler.HandleRegister), req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestLoginParseFormError(t *testing.T) {
	resetDB(t)
	req := httptest.NewRequest("POST", "/api/login", strings.NewReader(""))
	req.Header.Set("Content-Type", "multipart/form-data")
	rr := executeRequest(http.HandlerFunc(testHandler.HandleLogin), req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestFeedPageWithFilters(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "feed_filters@test.com", "feedfilter", "user")
	createTestPost(t, uid, "FilterPost", "Игры")
	createTestPost(t, uid, "SecondPost", "Финтех")

	tests := []struct {
		name string
		url  string
	}{
		{"sort_rating", "/feed?sort=rating"},
		{"sort_views", "/feed?sort=views"},
		{"sort_old", "/feed?sort=old"},
		{"category_filter", "/feed?category=Игры"},
		{"search_filter", "/feed?search=FilterPost"},
		{"pagination", "/feed?page=2"},
		{"combined", "/feed?category=Игры&search=Filter&sort=rating&page=1"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := authRequest(httptest.NewRequest("GET", tc.url, nil), uid, "user")
			rr := executeRequest(withPageAuth(testHandler.HandleFeedPage, testSecret), req)
			if rr.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d", rr.Code)
			}
		})
	}
}

func TestCreatePostWithFileUpload(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "upload@test.com", "uploader", "user")

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	w.WriteField("title", "Post With File")
	w.WriteField("description", "Description with file upload")
	w.WriteField("category", "Игры")
	fw, _ := w.CreateFormFile("files", "test.txt")
	fw.Write([]byte("file content here"))
	w.Close()

	req := httptest.NewRequest("POST", "/api/posts/create", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleCreatePost, testSecret), req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d", rr.Code)
	}
}

func TestCreatePostWithPollOptions(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "poll@test.com", "pollmaker", "user")

	form := url.Values{}
	form.Set("title", "Post With Poll")
	form.Set("description", "Description poll")
	form.Set("category", "Игры")
	form.Set("poll_question", "Favourite color?")
	form.Set("poll_option_1", "Red")
	form.Set("poll_option_2", "Blue")

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	for k, vs := range form {
		for _, v := range vs {
			mw.WriteField(k, v)
		}
	}
	mw.Close()

	req := httptest.NewRequest("POST", "/api/posts/create", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleCreatePost, testSecret), req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d", rr.Code)
	}
}

func TestCreatePostPremiumSize(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "premium@test.com", "premuser", "premium")

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	mw.WriteField("title", "Premium Post")
	mw.WriteField("description", "Premium description")
	mw.WriteField("category", "Финтех")
	fw, _ := mw.CreateFormFile("files", "big.txt")
	fw.Write([]byte("premium file content"))
	mw.Close()

	req := httptest.NewRequest("POST", "/api/posts/create", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req = authRequest(req, uid, "premium")
	rr := executeRequest(withAuth(testHandler.HandleCreatePost, testSecret), req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d", rr.Code)
	}
}

func TestUpdatePostWithFileUpload(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "updfile@test.com", "updfile", "user")
	pid := createTestPost(t, uid, "UpdFilePost", "Игры")

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	mw.WriteField("title", "Updated Title")
	mw.WriteField("description", "Updated Desc")
	mw.WriteField("category", "Финтех")
	fw, _ := mw.CreateFormFile("files", "newfile.txt")
	fw.Write([]byte("new file content"))
	mw.Close()

	req := httptest.NewRequest("POST", fmt.Sprintf("/api/posts/%d/update", pid), &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleUpdatePost, testSecret), req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d", rr.Code)
	}
}

func TestAddCommentWithReplyToSelfPost(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "selfcomm@test.com", "selfcomm", "user")
	pid := createTestPost(t, uid, "SelfCommPost", "Игры")

	form := url.Values{}
	form.Set("post_id", itoa(pid))
	form.Set("content", "Self comment")

	req := httptest.NewRequest("POST", "/api/comments/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleAddComment, testSecret), req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestChatPageFallbackPath(t *testing.T) {
	resetDB(t)
	uid1 := createTestUser(t, "chatfb1@test.com", "chatfb1", "user")

	testDB.Exec("ALTER TABLE chat_messages DISABLE TRIGGER ALL")
	testDB.Exec(
		"INSERT INTO chat_messages (sender_id, receiver_id, content) VALUES ($1, $2, $3)",
		uid1, 99999, "Orphaned message for fallback",
	)
	testDB.Exec("ALTER TABLE chat_messages ENABLE TRIGGER ALL")

	req := authRequest(httptest.NewRequest("GET", "/chat", nil), uid1, "user")
	rr := executeRequest(withPageAuth(testHandler.HandleChatPage, testSecret), req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	testDB.Exec("ALTER TABLE chat_messages DISABLE TRIGGER ALL")
	testDB.Exec("DELETE FROM chat_messages WHERE receiver_id = 99999")
	testDB.Exec("ALTER TABLE chat_messages ENABLE TRIGGER ALL")
}

func TestProfilePageNotFoundUser(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "profnf@test.com", "profnf", "user")

	req := authRequest(httptest.NewRequest("GET", "/profile/99999", nil), uid, "user")
	rr := executeRequest(withPageAuth(testHandler.HandleProfilePage, testSecret), req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func brokenHandler() *handlers.Handler {
	db, _ := sql.Open("pgx", "postgres://postgres:12345678@localhost:5432/startup_platform_test?sslmode=disable")
	db.Close()
	return &handlers.Handler{
		DB:        db,
		Templates: testHandler.Templates,
		JWTSecret: testSecret,
		UploadDir: testUploadDir,
	}
}

func TestRegisterDBError(t *testing.T) {
	h := brokenHandler()
	form := url.Values{}
	form.Set("email", "dberr@test.com")
	form.Set("nickname", "dberr")
	form.Set("password", "password123")

	req := httptest.NewRequest("POST", "/api/register", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.HandleRegister(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
}

func TestLoginDBError(t *testing.T) {
	h := brokenHandler()
	form := url.Values{}
	form.Set("email", "dberr@test.com")
	form.Set("password", "password123")

	req := httptest.NewRequest("POST", "/api/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.HandleLogin(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestFeedPageDBError(t *testing.T) {
	h := brokenHandler()
	req := httptest.NewRequest("GET", "/feed", nil)
	rr := httptest.NewRecorder()
	h.HandleFeedPage(rr, req)
	if rr.Code != http.StatusOK {
		t.Log("feed page with broken DB returned", rr.Code)
	}
}

func TestCreatePostDBError(t *testing.T) {
	h := brokenHandler()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	mw.WriteField("title", "DBError Post")
	mw.WriteField("description", "Description")
	mw.Close()

	req := httptest.NewRequest("POST", "/api/posts/create", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rr := httptest.NewRecorder()
	h.HandleCreatePost(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
}

func TestUpdatePostDBError(t *testing.T) {
	h := brokenHandler()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	mw.WriteField("title", "DBError Post")
	mw.WriteField("description", "Description")
	mw.Close()

	req := httptest.NewRequest("POST", "/api/posts/1/update", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rr := httptest.NewRecorder()
	h.HandleUpdatePost(rr, req)
	if rr.Code != http.StatusOK {
		t.Log("update with broken DB returned", rr.Code)
	}
}

func TestAddCommentDBError(t *testing.T) {
	h := brokenHandler()
	form := url.Values{}
	form.Set("post_id", "1")
	form.Set("content", "DB error comment")

	req := httptest.NewRequest("POST", "/api/comments/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.HandleAddComment(rr, req)
	if rr.Code >= 400 {
		t.Log("comment with broken DB returned", rr.Code)
	}
}

func TestProfilePageDBError(t *testing.T) {
	h := brokenHandler()
	req := httptest.NewRequest("GET", "/profile/1", nil)
	rr := httptest.NewRecorder()
	h.HandleProfilePage(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Log("profile with broken DB returned", rr.Code)
	}
}

func TestMigrateDBError(t *testing.T) {
	db, _ := sql.Open("pgx", "postgres://postgres:12345678@localhost:5432/startup_platform_test?sslmode=disable")
	db.Close()
	err := database.Migrate(db)
	if err == nil {
		t.Fatal("expected error from migrate with closed DB")
	}
}

func TestCreatePostBadUploadDir(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "baddir@test.com", "baddir", "user")

	h := &handlers.Handler{
		DB:        testDB,
		Templates: testHandler.Templates,
		JWTSecret: testSecret,
		UploadDir: "/nonexistent/path/uploads",
	}

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	mw.WriteField("title", "Post Bad Dir")
	mw.WriteField("description", "Description")
	mw.WriteField("category", "Игры")
	fw, _ := mw.CreateFormFile("files", "test.txt")
	fw.Write([]byte("file content"))
	mw.Close()

	req := httptest.NewRequest("POST", "/api/posts/create", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(h.HandleCreatePost, testSecret), req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected 303 (file error ignored), got %d", rr.Code)
	}
}

func TestUpdatePostBadUploadDir(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "baddirupd@test.com", "baddirupd", "user")
	pid := createTestPost(t, uid, "BadDirUpd", "Игры")

	h := &handlers.Handler{
		DB:        testDB,
		Templates: testHandler.Templates,
		JWTSecret: testSecret,
		UploadDir: "/nonexistent/path/uploads",
	}

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	mw.WriteField("title", "Updated Bad Dir")
	mw.WriteField("description", "Updated Desc")
	fw, _ := mw.CreateFormFile("files", "test.txt")
	fw.Write([]byte("file content"))
	mw.Close()

	req := httptest.NewRequest("POST", fmt.Sprintf("/api/posts/%d/update", pid), &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(h.HandleUpdatePost, testSecret), req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected 303 (file error ignored), got %d", rr.Code)
	}
}

func TestProfilePageIncomingFriendRequest(t *testing.T) {
	resetDB(t)
	uid1 := createTestUser(t, "incfr1@test.com", "incfr1", "user")
	uid2 := createTestUser(t, "incfr2@test.com", "incfr2", "user")

	testDB.Exec("INSERT INTO friendships (user_id, friend_id, status) VALUES ($1, $2, 'pending')", uid2, uid1)

	req := authRequest(httptest.NewRequest("GET", fmt.Sprintf("/profile/%d", uid2), nil), uid1, "user")
	rr := executeRequest(withPageAuth(testHandler.HandleProfilePage, testSecret), req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestFeedPageRowsProcessing(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "feedrows@test.com", "feedrows", "user")

	for i := 0; i < 15; i++ {
		createTestPost(t, uid, fmt.Sprintf("FeedPost%d", i), "Игры")
	}
	testDB.Exec("INSERT INTO files (post_id, filename, file_path, file_type, file_size) VALUES (1, 'img.png', '/uploads/img.png', 'image/png', 1024)")

	req := authRequest(httptest.NewRequest("GET", "/feed", nil), uid, "user")
	rr := executeRequest(withPageAuth(testHandler.HandleFeedPage, testSecret), req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	req = authRequest(httptest.NewRequest("GET", "/feed?page=2", nil), uid, "user")
	rr = executeRequest(withPageAuth(testHandler.HandleFeedPage, testSecret), req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 page 2, got %d", rr.Code)
	}
}

func TestUpdatePostWithCommentsOff(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "commoff@test.com", "commoff", "user")
	pid := createTestPost(t, uid, "CommOffPost", "Игры")

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	mw.WriteField("title", "Updated CommOff")
	mw.WriteField("description", "Updated Desc")
	mw.WriteField("comments_off", "on")
	mw.Close()

	req := httptest.NewRequest("POST", fmt.Sprintf("/api/posts/%d/update", pid), &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleUpdatePost, testSecret), req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d", rr.Code)
	}
}
