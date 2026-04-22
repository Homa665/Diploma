package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFeedPage(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "feed@test.com", "feeduser", "user")
	createTestPost(t, uid, "Test Post", "Игры")

	req := httptest.NewRequest("GET", "/feed", nil)
	req = authRequest(req, uid, "user")
	rr := executeRequest(withPageAuth(testHandler.HandleFeedPage, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	body := rr.Body.String()
	if !strings.Contains(body, "Test Post") {
		t.Fatal("expected post title in response")
	}
}

func TestFeedPageWithCategory(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "cat@test.com", "catuser", "user")
	createTestPost(t, uid, "Game Post", "Игры")
	createTestPost(t, uid, "Web Post", "Веб-сервисы")

	req := httptest.NewRequest("GET", "/feed?category=Игры", nil)
	req = authRequest(req, uid, "user")
	rr := executeRequest(withPageAuth(testHandler.HandleFeedPage, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "Game Post") {
		t.Fatal("expected Game Post in filtered results")
	}
}

func TestFeedPageWithSearch(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "search@test.com", "searchuser", "user")
	createTestPost(t, uid, "Unique Searchable", "Другое")

	req := httptest.NewRequest("GET", "/feed?search=Unique", nil)
	req = authRequest(req, uid, "user")
	rr := executeRequest(withPageAuth(testHandler.HandleFeedPage, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestFeedPageSortByRating(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "sortrate@test.com", "sortrate", "user")
	createTestPost(t, uid, "Post A", "Другое")

	req := httptest.NewRequest("GET", "/feed?sort=rating", nil)
	req = authRequest(req, uid, "user")
	rr := executeRequest(withPageAuth(testHandler.HandleFeedPage, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestFeedPageSortByViews(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "sortviews@test.com", "sortviews", "user")
	createTestPost(t, uid, "Post B", "Другое")

	req := httptest.NewRequest("GET", "/feed?sort=views", nil)
	req = authRequest(req, uid, "user")
	rr := executeRequest(withPageAuth(testHandler.HandleFeedPage, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestFeedPageSortByOld(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "sortold@test.com", "sortold", "user")
	createTestPost(t, uid, "Post C", "Другое")

	req := httptest.NewRequest("GET", "/feed?sort=old", nil)
	req = authRequest(req, uid, "user")
	rr := executeRequest(withPageAuth(testHandler.HandleFeedPage, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestFeedPagePagination(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "page@test.com", "pageuser", "user")
	for i := 0; i < 15; i++ {
		createTestPost(t, uid, fmt.Sprintf("Post %d", i), "Другое")
	}

	req := httptest.NewRequest("GET", "/feed?page=2", nil)
	req = authRequest(req, uid, "user")
	rr := executeRequest(withPageAuth(testHandler.HandleFeedPage, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestFeedPageCategoryAndSearch(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "both@test.com", "bothuser", "user")
	createTestPost(t, uid, "Target Post", "Игры")

	req := httptest.NewRequest("GET", "/feed?category=Игры&search=Target", nil)
	req = authRequest(req, uid, "user")
	rr := executeRequest(withPageAuth(testHandler.HandleFeedPage, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestCreatePostPage(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "create@test.com", "createuser", "user")

	req := httptest.NewRequest("GET", "/post/new", nil)
	req = authRequest(req, uid, "user")
	rr := executeRequest(withPageAuth(testHandler.HandleCreatePostPage, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestCreatePost(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "newpost@test.com", "newpost", "user")

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("title", "New Post Title")
	writer.WriteField("description", "Post description here")
	writer.WriteField("category", "Игры")
	writer.WriteField("tags", "go,test")
	writer.Close()

	req := httptest.NewRequest("POST", "/api/posts/create", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleCreatePost, testSecret), req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d: %s", rr.Code, rr.Body.String())
	}
	loc := rr.Header().Get("Location")
	if !strings.Contains(loc, "/post/") {
		t.Fatalf("expected redirect to /post/, got %s", loc)
	}
}

func TestCreatePostEmptyFields(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "emptypost@test.com", "emptypost", "user")

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("title", "")
	writer.WriteField("description", "")
	writer.Close()

	req := httptest.NewRequest("POST", "/api/posts/create", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleCreatePost, testSecret), req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestCreatePostBannedRendersHTMLMessage(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "banpost@test.com", "banpost", "user")
	adminID := createTestUser(t, "banadmin@test.com", "banadmin", "admin")

	_, err := testDB.Exec(
		"INSERT INTO user_bans (user_id, admin_id, reason, restriction, expires_at) VALUES ($1, $2, $3, $4, $5)",
		uid, adminID, "spam", "posts", time.Now().Add(48*time.Hour),
	)
	if err != nil {
		t.Fatalf("failed to insert ban: %v", err)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("title", "Blocked post")
	writer.WriteField("description", "Should not be created")
	writer.Close()

	req := httptest.NewRequest("POST", "/post/new", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req = authRequest(req, uid, "user")
	rr := executeRequest(withPageAuth(testHandler.HandleCreatePost, testSecret), req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", rr.Code, rr.Body.String())
	}

	bodyText := rr.Body.String()
	if !strings.Contains(bodyText, "Вы не можете создавать посты во время блокировки") {
		t.Fatalf("expected blocked message in HTML response, got: %s", bodyText)
	}
	if !strings.Contains(strings.ToLower(bodyText), "<!doctype html>") {
		t.Fatalf("expected HTML response instead of raw JSON, got: %s", bodyText)
	}

	var postsCount int
	testDB.QueryRow("SELECT COUNT(*) FROM posts WHERE author_id = $1", uid).Scan(&postsCount)
	if postsCount != 0 {
		t.Fatalf("expected no posts to be created while banned, got %d", postsCount)
	}
}

func TestCreatePostWithFile(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "filpost@test.com", "filpost", "user")

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("title", "Post With File")
	writer.WriteField("description", "Has a file attached")
	writer.WriteField("category", "Другое")

	part, _ := writer.CreateFormFile("files", "test.txt")
	part.Write([]byte("test file content"))
	writer.Close()

	req := httptest.NewRequest("POST", "/api/posts/create", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleCreatePost, testSecret), req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d: %s", rr.Code, rr.Body.String())
	}

	var count int
	testDB.QueryRow("SELECT COUNT(*) FROM files").Scan(&count)
	if count != 1 {
		t.Fatalf("expected 1 file, got %d", count)
	}
}

func TestCreatePostWithPoll(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "pollpost@test.com", "pollpost", "user")

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("title", "Poll Post")
	writer.WriteField("description", "Has a poll")
	writer.WriteField("poll_question", "What do you prefer?")
	writer.WriteField("poll_option_1", "Option A")
	writer.WriteField("poll_option_2", "Option B")
	writer.WriteField("poll_option_3", "Option C")
	writer.Close()

	req := httptest.NewRequest("POST", "/api/posts/create", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleCreatePost, testSecret), req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d: %s", rr.Code, rr.Body.String())
	}

	var count int
	testDB.QueryRow("SELECT COUNT(*) FROM polls").Scan(&count)
	if count != 1 {
		t.Fatalf("expected 1 poll, got %d", count)
	}

	testDB.QueryRow("SELECT COUNT(*) FROM poll_options").Scan(&count)
	if count != 3 {
		t.Fatalf("expected 3 poll options, got %d", count)
	}
}

func TestCreatePostPremiumUser(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "prempost@test.com", "prempost", "premium")

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("title", "Premium Post")
	writer.WriteField("description", "From premium user")
	writer.Close()

	req := httptest.NewRequest("POST", "/api/posts/create", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req = authRequest(req, uid, "premium")
	rr := executeRequest(withAuth(testHandler.HandleCreatePost, testSecret), req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d", rr.Code)
	}

	var isPremium bool
	testDB.QueryRow("SELECT is_premium FROM posts ORDER BY id DESC LIMIT 1").Scan(&isPremium)
	if !isPremium {
		t.Fatal("expected premium post")
	}
}

func TestCreatePostAdminUser(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "admpost@test.com", "admpost", "admin")

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("title", "Admin Post")
	writer.WriteField("description", "From admin user")
	writer.Close()

	req := httptest.NewRequest("POST", "/api/posts/create", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req = authRequest(req, uid, "admin")
	rr := executeRequest(withAuth(testHandler.HandleCreatePost, testSecret), req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d", rr.Code)
	}
}

func TestPostPage(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "post@test.com", "postuser", "user")
	pid := createTestPost(t, uid, "View This Post", "Образование")

	req := httptest.NewRequest("GET", "/post/"+itoa(pid), nil)
	req.URL.Path = "/post/" + itoa(pid)
	req = authRequest(req, uid, "user")
	rr := executeRequest(withPageAuth(testHandler.HandlePostPage, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestPostPageNotFound(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "pnf@test.com", "pnfuser", "user")

	req := httptest.NewRequest("GET", "/post/99999", nil)
	req.URL.Path = "/post/99999"
	req = authRequest(req, uid, "user")
	rr := executeRequest(withPageAuth(testHandler.HandlePostPage, testSecret), req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestPostPageInvalidID(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "pinv@test.com", "pinvuser", "user")

	req := httptest.NewRequest("GET", "/post/abc", nil)
	req.URL.Path = "/post/abc"
	req = authRequest(req, uid, "user")
	rr := executeRequest(withPageAuth(testHandler.HandlePostPage, testSecret), req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestPostPageWithComments(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "pcomm@test.com", "pcommuser", "user")
	pid := createTestPost(t, uid, "Commented Post", "Другое")

	var parentID int
	testDB.QueryRow("INSERT INTO comments (post_id, author_id, content) VALUES ($1, $2, 'Parent comment') RETURNING id", pid, uid).Scan(&parentID)
	testDB.Exec("INSERT INTO comments (post_id, author_id, parent_id, content) VALUES ($1, $2, $3, 'Reply comment')", pid, uid, parentID)
	testDB.Exec("INSERT INTO comment_likes (comment_id, user_id) VALUES ($1, $2)", parentID, uid)

	req := httptest.NewRequest("GET", "/post/"+itoa(pid), nil)
	req.URL.Path = "/post/" + itoa(pid)
	req = authRequest(req, uid, "user")
	rr := executeRequest(withPageAuth(testHandler.HandlePostPage, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestPostPageWithRating(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "prate@test.com", "prateuser", "user")
	uid2 := createTestUser(t, "prate2@test.com", "prater", "expert")
	pid := createTestPost(t, uid, "Rated Post", "Другое")

	testDB.Exec("INSERT INTO ratings (post_id, user_id, score, review, is_expert) VALUES ($1, $2, 4, 'Good project', TRUE)", pid, uid2)
	testDB.Exec("INSERT INTO ratings (post_id, user_id, score, review, is_expert) VALUES ($1, $2, 3, '', FALSE)", pid, uid)

	req := httptest.NewRequest("GET", "/post/"+itoa(pid), nil)
	req.URL.Path = "/post/" + itoa(pid)
	req = authRequest(req, uid, "user")
	rr := executeRequest(withPageAuth(testHandler.HandlePostPage, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestPostPageWithPoll(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "ppoll@test.com", "ppolluser", "user")
	pid := createTestPost(t, uid, "Poll Post", "Другое")

	var pollID int
	testDB.QueryRow("INSERT INTO polls (post_id, question) VALUES ($1, 'Test?') RETURNING id", pid).Scan(&pollID)
	var optID int
	testDB.QueryRow("INSERT INTO poll_options (poll_id, text) VALUES ($1, 'Opt1') RETURNING id", pollID).Scan(&optID)
	testDB.Exec("INSERT INTO poll_votes (poll_id, option_id, user_id) VALUES ($1, $2, $3)", pollID, optID, uid)

	req := httptest.NewRequest("GET", "/post/"+itoa(pid), nil)
	req.URL.Path = "/post/" + itoa(pid)
	req = authRequest(req, uid, "user")
	rr := executeRequest(withPageAuth(testHandler.HandlePostPage, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestPostPageWithTeamRequests(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "pteam@test.com", "pteamuser", "user")
	pid := createTestPost(t, uid, "Team Post", "Другое")

	testDB.Exec("INSERT INTO team_requests (post_id, author_id, title, is_open) VALUES ($1, $2, 'Need dev', TRUE)", pid, uid)

	req := httptest.NewRequest("GET", "/post/"+itoa(pid), nil)
	req.URL.Path = "/post/" + itoa(pid)
	req = authRequest(req, uid, "user")
	rr := executeRequest(withPageAuth(testHandler.HandlePostPage, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestPostPageWithFiles(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "pfile@test.com", "pfileuser", "user")
	pid := createTestPost(t, uid, "File Post", "Другое")

	testDB.Exec("INSERT INTO files (post_id, filename, file_path, file_type, file_size) VALUES ($1, 'test.png', '/uploads/test.png', 'image/png', 1024)", pid)

	req := httptest.NewRequest("GET", "/post/"+itoa(pid), nil)
	req.URL.Path = "/post/" + itoa(pid)
	req = authRequest(req, uid, "user")
	rr := executeRequest(withPageAuth(testHandler.HandlePostPage, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestEditPostPage(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "editp@test.com", "editpuser", "user")
	pid := createTestPost(t, uid, "Edit Me", "Другое")

	req := httptest.NewRequest("GET", "/post/"+itoa(pid)+"/edit", nil)
	req.URL.Path = "/post/" + itoa(pid) + "/edit"
	req = authRequest(req, uid, "user")
	rr := executeRequest(withPageAuth(testHandler.HandleEditPostPage, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestEditPostPageForbidden(t *testing.T) {
	resetDB(t)
	uid1 := createTestUser(t, "editown@test.com", "editown", "user")
	uid2 := createTestUser(t, "editoth@test.com", "editoth", "user")
	pid := createTestPost(t, uid1, "Not Yours", "Другое")

	req := httptest.NewRequest("GET", "/post/"+itoa(pid)+"/edit", nil)
	req.URL.Path = "/post/" + itoa(pid) + "/edit"
	req = authRequest(req, uid2, "user")
	rr := executeRequest(withPageAuth(testHandler.HandleEditPostPage, testSecret), req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestEditPostPageNotFound(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "editnf@test.com", "editnf", "user")

	req := httptest.NewRequest("GET", "/post/99999/edit", nil)
	req.URL.Path = "/post/99999/edit"
	req = authRequest(req, uid, "user")
	rr := executeRequest(withPageAuth(testHandler.HandleEditPostPage, testSecret), req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestEditPostPageInvalidID(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "editinv@test.com", "editinv", "user")

	req := httptest.NewRequest("GET", "/post/abc/edit", nil)
	req.URL.Path = "/post/abc/edit"
	req = authRequest(req, uid, "user")
	rr := executeRequest(withPageAuth(testHandler.HandleEditPostPage, testSecret), req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestEditPostPageAdmin(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "editadm@test.com", "editadm", "user")
	adminID := createTestUser(t, "admedit@test.com", "admedit", "admin")
	pid := createTestPost(t, uid, "Admin Can Edit", "Другое")

	req := httptest.NewRequest("GET", "/post/"+itoa(pid)+"/edit", nil)
	req.URL.Path = "/post/" + itoa(pid) + "/edit"
	req = authRequest(req, adminID, "admin")
	rr := executeRequest(withPageAuth(testHandler.HandleEditPostPage, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestUpdatePost(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "upd@test.com", "upduser", "user")
	pid := createTestPost(t, uid, "Old Title", "Другое")

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("title", "New Title")
	writer.WriteField("description", "New description")
	writer.WriteField("category", "Игры")
	writer.WriteField("tags", "updated")
	writer.WriteField("comments_off", "on")
	writer.Close()

	req := httptest.NewRequest("POST", "/api/posts/"+itoa(pid)+"/update", body)
	req.URL.Path = "/api/posts/" + itoa(pid) + "/update"
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleUpdatePost, testSecret), req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d: %s", rr.Code, rr.Body.String())
	}

	var title string
	testDB.QueryRow("SELECT title FROM posts WHERE id = $1", pid).Scan(&title)
	if title != "New Title" {
		t.Fatalf("expected 'New Title', got '%s'", title)
	}
}

func TestUpdatePostEmptyFields(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "updempty@test.com", "updempty", "user")
	pid := createTestPost(t, uid, "Title", "Другое")

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("title", "")
	writer.WriteField("description", "")
	writer.Close()

	req := httptest.NewRequest("POST", "/api/posts/"+itoa(pid)+"/update", body)
	req.URL.Path = "/api/posts/" + itoa(pid) + "/update"
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleUpdatePost, testSecret), req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestUpdatePostForbidden(t *testing.T) {
	resetDB(t)
	uid1 := createTestUser(t, "updow@test.com", "updow", "user")
	uid2 := createTestUser(t, "updot@test.com", "updot", "user")
	pid := createTestPost(t, uid1, "Title", "Другое")

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("title", "Hacked")
	writer.WriteField("description", "Hacked desc")
	writer.Close()

	req := httptest.NewRequest("POST", "/api/posts/"+itoa(pid)+"/update", body)
	req.URL.Path = "/api/posts/" + itoa(pid) + "/update"
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req = authRequest(req, uid2, "user")
	rr := executeRequest(withAuth(testHandler.HandleUpdatePost, testSecret), req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestUpdatePostInvalidID(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "updinv@test.com", "updinv", "user")

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("title", "T")
	writer.WriteField("description", "D")
	writer.Close()

	req := httptest.NewRequest("POST", "/api/posts/abc/update", body)
	req.URL.Path = "/api/posts/abc/update"
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleUpdatePost, testSecret), req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestUpdatePostWithFile(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "updfile@test.com", "updfile", "user")
	pid := createTestPost(t, uid, "Title", "Другое")

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("title", "Updated Title")
	writer.WriteField("description", "Updated desc")
	part, _ := writer.CreateFormFile("files", "newfile.txt")
	part.Write([]byte("new content"))
	writer.Close()

	req := httptest.NewRequest("POST", "/api/posts/"+itoa(pid)+"/update", body)
	req.URL.Path = "/api/posts/" + itoa(pid) + "/update"
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleUpdatePost, testSecret), req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestDeletePost(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "del@test.com", "deluser", "user")
	pid := createTestPost(t, uid, "Delete Me", "Другое")

	req := httptest.NewRequest("POST", "/api/posts/"+itoa(pid)+"/delete", nil)
	req.URL.Path = "/api/posts/" + itoa(pid) + "/delete"
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleDeletePost, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var count int
	testDB.QueryRow("SELECT COUNT(*) FROM posts WHERE id = $1", pid).Scan(&count)
	if count != 0 {
		t.Fatal("expected post to be deleted")
	}
}

func TestDeletePostForbidden(t *testing.T) {
	resetDB(t)
	uid1 := createTestUser(t, "owner@test.com", "owner", "user")
	uid2 := createTestUser(t, "other@test.com", "other", "user")
	pid := createTestPost(t, uid1, "Not Yours", "Другое")

	req := httptest.NewRequest("POST", "/api/posts/"+itoa(pid)+"/delete", nil)
	req.URL.Path = "/api/posts/" + itoa(pid) + "/delete"
	req = authRequest(req, uid2, "user")
	rr := executeRequest(withAuth(testHandler.HandleDeletePost, testSecret), req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestDeletePostInvalidID(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "delinv@test.com", "delinv", "user")

	req := httptest.NewRequest("POST", "/api/posts/abc/delete", nil)
	req.URL.Path = "/api/posts/abc/delete"
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleDeletePost, testSecret), req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestDeletePostAdmin(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "deladmow@test.com", "deladmow", "user")
	adminID := createTestUser(t, "deladm@test.com", "deladm", "admin")
	pid := createTestPost(t, uid, "Admin Deletes", "Другое")

	req := httptest.NewRequest("POST", "/api/posts/"+itoa(pid)+"/delete", nil)
	req.URL.Path = "/api/posts/" + itoa(pid) + "/delete"
	req = authRequest(req, adminID, "admin")
	rr := executeRequest(withAuth(testHandler.HandleDeletePost, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestDeleteFile(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "delfile@test.com", "delfile", "user")
	pid := createTestPost(t, uid, "File Post", "Другое")

	tmpFile := filepath.Join(testUploadDir, "testdel.txt")
	os.WriteFile(tmpFile, []byte("test"), 0644)

	var fileID int
	testDB.QueryRow(
		"INSERT INTO files (post_id, filename, file_path, file_type, file_size) VALUES ($1, 'testdel.txt', '/uploads/testdel.txt', 'text/plain', 4) RETURNING id",
		pid,
	).Scan(&fileID)

	req := httptest.NewRequest("POST", "/api/files/"+itoa(fileID)+"/delete", nil)
	req.URL.Path = "/api/files/" + itoa(fileID) + "/delete"
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleDeleteFile, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var count int
	testDB.QueryRow("SELECT COUNT(*) FROM files WHERE id = $1", fileID).Scan(&count)
	if count != 0 {
		t.Fatal("expected file record deleted")
	}
}

func TestDeleteFileForbidden(t *testing.T) {
	resetDB(t)
	uid1 := createTestUser(t, "delfown@test.com", "delfown", "user")
	uid2 := createTestUser(t, "delfoth@test.com", "delfoth", "user")
	pid := createTestPost(t, uid1, "Post", "Другое")

	var fileID int
	testDB.QueryRow(
		"INSERT INTO files (post_id, filename, file_path, file_type, file_size) VALUES ($1, 'x.txt', '/uploads/x.txt', 'text/plain', 1) RETURNING id",
		pid,
	).Scan(&fileID)

	req := httptest.NewRequest("POST", "/api/files/"+itoa(fileID)+"/delete", nil)
	req.URL.Path = "/api/files/" + itoa(fileID) + "/delete"
	req = authRequest(req, uid2, "user")
	rr := executeRequest(withAuth(testHandler.HandleDeleteFile, testSecret), req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestDeleteFileInvalidID(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "delfinv@test.com", "delfinv", "user")

	req := httptest.NewRequest("POST", "/api/files/abc/delete", nil)
	req.URL.Path = "/api/files/abc/delete"
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleDeleteFile, testSecret), req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestAddComment(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "comm@test.com", "commuser", "user")
	uid2 := createTestUser(t, "comm2@test.com", "commuser2", "user")
	pid := createTestPost(t, uid2, "Commentable", "Другое")

	form := url.Values{}
	form.Set("post_id", itoa(pid))
	form.Set("content", "Test comment")

	req := httptest.NewRequest("POST", "/api/comments/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleAddComment, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp["success"] != true {
		t.Fatal("expected success")
	}

	var notifCount int
	testDB.QueryRow("SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND type = 'comment'", uid2).Scan(&notifCount)
	if notifCount != 1 {
		t.Fatal("expected notification for comment")
	}
}

func TestAddCommentOwnPost(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "owncomm@test.com", "owncomm", "user")
	pid := createTestPost(t, uid, "My Post", "Другое")

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

func TestAddCommentWithParent(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "reply@test.com", "replyuser", "user")
	pid := createTestPost(t, uid, "Post", "Другое")

	var parentID int
	testDB.QueryRow("INSERT INTO comments (post_id, author_id, content) VALUES ($1, $2, 'Parent') RETURNING id", pid, uid).Scan(&parentID)

	form := url.Values{}
	form.Set("post_id", itoa(pid))
	form.Set("content", "Reply to parent")
	form.Set("parent_id", itoa(parentID))

	req := httptest.NewRequest("POST", "/api/comments/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleAddComment, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestAddCommentInvalidPostID(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "invcomm@test.com", "invcomm", "user")

	form := url.Values{}
	form.Set("post_id", "abc")
	form.Set("content", "Test")

	req := httptest.NewRequest("POST", "/api/comments/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleAddComment, testSecret), req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestAddCommentEmpty(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "empty@test.com", "emptycomm", "user")
	pid := createTestPost(t, uid, "Post", "Другое")

	form := url.Values{}
	form.Set("post_id", itoa(pid))
	form.Set("content", "")

	req := httptest.NewRequest("POST", "/api/comments/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleAddComment, testSecret), req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestAddCommentDisabled(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "off@test.com", "commoff", "user")
	pid := createTestPost(t, uid, "Closed", "Другое")
	testDB.Exec("UPDATE posts SET comments_off = TRUE WHERE id = $1", pid)

	form := url.Values{}
	form.Set("post_id", itoa(pid))
	form.Set("content", "Trying to comment")

	req := httptest.NewRequest("POST", "/api/comments/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleAddComment, testSecret), req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestDeleteComment(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "delcomm@test.com", "delcomm", "user")
	pid := createTestPost(t, uid, "Post", "Другое")

	var cid int
	testDB.QueryRow("INSERT INTO comments (post_id, author_id, content) VALUES ($1, $2, 'Delete me') RETURNING id", pid, uid).Scan(&cid)

	req := httptest.NewRequest("POST", "/api/comments/"+itoa(cid)+"/delete", nil)
	req.URL.Path = "/api/comments/" + itoa(cid) + "/delete"
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleDeleteComment, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestDeleteCommentByPostOwner(t *testing.T) {
	resetDB(t)
	uid1 := createTestUser(t, "postow@test.com", "postow", "user")
	uid2 := createTestUser(t, "commenter@test.com", "commenter", "user")
	pid := createTestPost(t, uid1, "Post", "Другое")

	var cid int
	testDB.QueryRow("INSERT INTO comments (post_id, author_id, content) VALUES ($1, $2, 'Other comment') RETURNING id", pid, uid2).Scan(&cid)

	req := httptest.NewRequest("POST", "/api/comments/"+itoa(cid)+"/delete", nil)
	req.URL.Path = "/api/comments/" + itoa(cid) + "/delete"
	req = authRequest(req, uid1, "user")
	rr := executeRequest(withAuth(testHandler.HandleDeleteComment, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for post owner deleting comment, got %d", rr.Code)
	}
}

func TestDeleteCommentForbidden(t *testing.T) {
	resetDB(t)
	uid1 := createTestUser(t, "commown@test.com", "commown", "user")
	uid2 := createTestUser(t, "commoth@test.com", "commoth", "user")
	uid3 := createTestUser(t, "commrand@test.com", "commrand", "user")
	pid := createTestPost(t, uid1, "Post", "Другое")

	var cid int
	testDB.QueryRow("INSERT INTO comments (post_id, author_id, content) VALUES ($1, $2, 'Not yours') RETURNING id", pid, uid2).Scan(&cid)

	req := httptest.NewRequest("POST", "/api/comments/"+itoa(cid)+"/delete", nil)
	req.URL.Path = "/api/comments/" + itoa(cid) + "/delete"
	req = authRequest(req, uid3, "user")
	rr := executeRequest(withAuth(testHandler.HandleDeleteComment, testSecret), req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestDeleteCommentInvalidID(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "delcinv@test.com", "delcinv", "user")

	req := httptest.NewRequest("POST", "/api/comments/abc/delete", nil)
	req.URL.Path = "/api/comments/abc/delete"
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleDeleteComment, testSecret), req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestRate(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "rate@test.com", "rateuser", "user")
	uid2 := createTestUser(t, "rate2@test.com", "rateuser2", "user")
	pid := createTestPost(t, uid2, "Rate Me", "Другое")

	form := url.Values{}
	form.Set("post_id", itoa(pid))
	form.Set("score", "4")
	form.Set("review", "Great project!")

	req := httptest.NewRequest("POST", "/api/rate", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleRate, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestRateExpert(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "exprate@test.com", "exprate", "expert")
	uid2 := createTestUser(t, "exprate2@test.com", "exprate2", "user")
	pid := createTestPost(t, uid2, "Rate Me Expert", "Другое")

	form := url.Values{}
	form.Set("post_id", itoa(pid))
	form.Set("score", "5")
	form.Set("review", "Expert review")

	req := httptest.NewRequest("POST", "/api/rate", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid, "expert")
	rr := executeRequest(withAuth(testHandler.HandleRate, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var isExpert bool
	testDB.QueryRow("SELECT is_expert FROM ratings WHERE user_id = $1", uid).Scan(&isExpert)
	if !isExpert {
		t.Fatal("expected is_expert=true")
	}
}

func TestRateOwnPost(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "selfrate@test.com", "selfrate", "user")
	pid := createTestPost(t, uid, "My Post", "Другое")

	form := url.Values{}
	form.Set("post_id", itoa(pid))
	form.Set("score", "5")

	req := httptest.NewRequest("POST", "/api/rate", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleRate, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestRateInvalidScore(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "badscore@test.com", "badscore", "user")
	pid := createTestPost(t, uid, "Post", "Другое")

	form := url.Values{}
	form.Set("post_id", itoa(pid))
	form.Set("score", "6")

	req := httptest.NewRequest("POST", "/api/rate", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleRate, testSecret), req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestRateZeroScore(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "zscore@test.com", "zscore", "user")
	pid := createTestPost(t, uid, "Post", "Другое")

	form := url.Values{}
	form.Set("post_id", itoa(pid))
	form.Set("score", "0")

	req := httptest.NewRequest("POST", "/api/rate", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleRate, testSecret), req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestRateInvalidPostID(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "badpid@test.com", "badpid", "user")

	form := url.Values{}
	form.Set("post_id", "abc")
	form.Set("score", "5")

	req := httptest.NewRequest("POST", "/api/rate", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleRate, testSecret), req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestRateUpdate(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "update@test.com", "updaterate", "user")
	pid := createTestPost(t, uid, "Post", "Другое")

	for _, score := range []string{"3", "5"} {
		form := url.Values{}
		form.Set("post_id", itoa(pid))
		form.Set("score", score)

		req := httptest.NewRequest("POST", "/api/rate", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req = authRequest(req, uid, "user")
		executeRequest(withAuth(testHandler.HandleRate, testSecret), req)
	}

	var s int
	testDB.QueryRow("SELECT score FROM ratings WHERE post_id = $1 AND user_id = $2", pid, uid).Scan(&s)
	if s != 5 {
		t.Fatalf("expected updated score 5, got %d", s)
	}
}

func TestVotePoll(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "poll@test.com", "polluser", "user")
	pid := createTestPost(t, uid, "Poll Post", "Другое")

	var pollID int
	testDB.QueryRow("INSERT INTO polls (post_id, question) VALUES ($1, 'Q?') RETURNING id", pid).Scan(&pollID)
	var optID int
	testDB.QueryRow("INSERT INTO poll_options (poll_id, text) VALUES ($1, 'Option A') RETURNING id", pollID).Scan(&optID)

	form := url.Values{}
	form.Set("poll_id", itoa(pollID))
	form.Set("option_id", itoa(optID))

	req := httptest.NewRequest("POST", "/api/poll/vote", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleVotePoll, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestVotePollDuplicate(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "dup@test.com", "dupvote", "user")
	pid := createTestPost(t, uid, "Poll Post", "Другое")

	var pollID int
	testDB.QueryRow("INSERT INTO polls (post_id, question) VALUES ($1, 'Q?') RETURNING id", pid).Scan(&pollID)
	var optID int
	testDB.QueryRow("INSERT INTO poll_options (poll_id, text) VALUES ($1, 'Option A') RETURNING id", pollID).Scan(&optID)
	testDB.Exec("INSERT INTO poll_votes (poll_id, option_id, user_id) VALUES ($1, $2, $3)", pollID, optID, uid)

	form := url.Values{}
	form.Set("poll_id", itoa(pollID))
	form.Set("option_id", itoa(optID))

	req := httptest.NewRequest("POST", "/api/poll/vote", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleVotePoll, testSecret), req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rr.Code)
	}
}

func TestVotePollInvalidPollID(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "invpoll@test.com", "invpoll", "user")

	form := url.Values{}
	form.Set("poll_id", "abc")
	form.Set("option_id", "1")

	req := httptest.NewRequest("POST", "/api/poll/vote", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleVotePoll, testSecret), req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestVotePollInvalidOptionID(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "invopt@test.com", "invopt", "user")

	form := url.Values{}
	form.Set("poll_id", "1")
	form.Set("option_id", "abc")

	req := httptest.NewRequest("POST", "/api/poll/vote", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleVotePoll, testSecret), req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestRateInvalidScoreString(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "strrate@test.com", "strrate", "user")
	pid := createTestPost(t, uid, "Post", "Другое")

	form := url.Values{}
	form.Set("post_id", itoa(pid))
	form.Set("score", "abc")

	req := httptest.NewRequest("POST", "/api/rate", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleRate, testSecret), req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestLikeComment(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "like@test.com", "likeuser", "user")
	pid := createTestPost(t, uid, "Post", "Другое")

	var commentID int
	testDB.QueryRow("INSERT INTO comments (post_id, author_id, content) VALUES ($1, $2, 'Comment') RETURNING id", pid, uid).Scan(&commentID)

	req := httptest.NewRequest("POST", "/api/comments/"+itoa(commentID)+"/like", nil)
	req.URL.Path = "/api/comments/" + itoa(commentID) + "/like"
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleLikeComment, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp["liked"] != true {
		t.Fatal("expected liked=true")
	}
}

func TestLikeCommentToggle(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "toggle@test.com", "togglelike", "user")
	pid := createTestPost(t, uid, "Post", "Другое")

	var commentID int
	testDB.QueryRow("INSERT INTO comments (post_id, author_id, content) VALUES ($1, $2, 'Comment') RETURNING id", pid, uid).Scan(&commentID)

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("POST", "/api/comments/"+itoa(commentID)+"/like", nil)
		req.URL.Path = "/api/comments/" + itoa(commentID) + "/like"
		req = authRequest(req, uid, "user")
		executeRequest(withAuth(testHandler.HandleLikeComment, testSecret), req)
	}

	var likeCount int
	testDB.QueryRow("SELECT like_count FROM comments WHERE id = $1", commentID).Scan(&likeCount)
	if likeCount != 0 {
		t.Fatalf("expected like_count=0 after toggle, got %d", likeCount)
	}
}

func TestLikeCommentInvalidID(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "likeinv@test.com", "likeinv", "user")

	req := httptest.NewRequest("POST", "/api/comments/abc/like", nil)
	req.URL.Path = "/api/comments/abc/like"
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleLikeComment, testSecret), req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}
