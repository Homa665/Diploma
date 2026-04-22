package tests

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestProfilePage(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "profile@test.com", "profuser", "user")

	req := httptest.NewRequest("GET", "/profile/me", nil)
	req.URL.Path = "/profile/me"
	req = authRequest(req, uid, "user")
	rr := executeRequest(withPageAuth(testHandler.HandleProfilePage, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestProfilePageOtherUser(t *testing.T) {
	resetDB(t)
	uid1 := createTestUser(t, "viewer@test.com", "viewer", "user")
	uid2 := createTestUser(t, "viewed@test.com", "viewed", "user")

	req := httptest.NewRequest("GET", "/profile/"+itoa(uid2), nil)
	req.URL.Path = "/profile/" + itoa(uid2)
	req = authRequest(req, uid1, "user")
	rr := executeRequest(withPageAuth(testHandler.HandleProfilePage, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestProfilePageNotFound(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "nf@test.com", "nfuser", "user")

	req := httptest.NewRequest("GET", "/profile/99999", nil)
	req.URL.Path = "/profile/99999"
	req = authRequest(req, uid, "user")
	rr := executeRequest(withPageAuth(testHandler.HandleProfilePage, testSecret), req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestProfilePageInvalidID(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "pinv@test.com", "pinvprof", "user")

	req := httptest.NewRequest("GET", "/profile/abc", nil)
	req.URL.Path = "/profile/abc"
	req = authRequest(req, uid, "user")
	rr := executeRequest(withPageAuth(testHandler.HandleProfilePage, testSecret), req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestProfilePageWithData(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "fullprof@test.com", "fullprof", "user")
	uid2 := createTestUser(t, "friend@test.com", "profriend", "user")
	pid := createTestPost(t, uid, "User Post", "Другое")

	testDB.Exec("INSERT INTO friendships (user_id, friend_id, status) VALUES ($1, $2, 'accepted')", uid, uid2)
	testDB.Exec("INSERT INTO ratings (post_id, user_id, score) VALUES ($1, $2, 8)", pid, uid2)

	req := httptest.NewRequest("GET", "/profile/"+itoa(uid), nil)
	req.URL.Path = "/profile/" + itoa(uid)
	req = authRequest(req, uid2, "user")
	rr := executeRequest(withPageAuth(testHandler.HandleProfilePage, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestProfilePageBlocked(t *testing.T) {
	resetDB(t)
	uid1 := createTestUser(t, "blocker@test.com", "profblocker", "user")
	uid2 := createTestUser(t, "blocked@test.com", "profblocked", "user")
	testDB.Exec("INSERT INTO blocked_users (user_id, blocked_id) VALUES ($1, $2)", uid2, uid1)

	req := httptest.NewRequest("GET", "/profile/"+itoa(uid2), nil)
	req.URL.Path = "/profile/" + itoa(uid2)
	req = authRequest(req, uid1, "user")
	rr := executeRequest(withPageAuth(testHandler.HandleProfilePage, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestEditProfile(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "edit@test.com", "edituser", "user")

	form := url.Values{}
	form.Set("name", "New Name")
	form.Set("city", "Минск")
	form.Set("bio", "Test bio")
	form.Set("interests", "Go, Web")
	form.Set("user_role2", "Developer")
	form.Set("phone", "+375291234567")

	req := httptest.NewRequest("POST", "/api/profile/edit", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleEditProfile, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var name, city string
	testDB.QueryRow("SELECT name, city FROM users WHERE id = $1", uid).Scan(&name, &city)
	if name != "New Name" || city != "Минск" {
		t.Fatalf("expected name='New Name', city='Минск', got name='%s', city='%s'", name, city)
	}
}

func TestEditProfileEmptyName(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "empname@test.com", "empname", "user")

	form := url.Values{}
	form.Set("name", "")

	req := httptest.NewRequest("POST", "/api/profile/edit", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleEditProfile, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestAddFriend(t *testing.T) {
	resetDB(t)
	uid1 := createTestUser(t, "friend1@test.com", "friend1", "user")
	uid2 := createTestUser(t, "friend2@test.com", "friend2", "user")

	form := url.Values{}
	form.Set("friend_id", itoa(uid2))

	req := httptest.NewRequest("POST", "/api/friends/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid1, "user")
	rr := executeRequest(withAuth(testHandler.HandleAddFriend, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var status string
	testDB.QueryRow("SELECT status FROM friendships WHERE user_id = $1 AND friend_id = $2", uid1, uid2).Scan(&status)
	if status != "pending" {
		t.Fatalf("expected pending, got %s", status)
	}
}

func TestAcceptFriend(t *testing.T) {
	resetDB(t)
	uid1 := createTestUser(t, "accept1@test.com", "accept1", "user")
	uid2 := createTestUser(t, "accept2@test.com", "accept2", "user")
	testDB.Exec("INSERT INTO friendships (user_id, friend_id, status) VALUES ($1, $2, 'pending')", uid1, uid2)

	form := url.Values{}
	form.Set("friend_id", itoa(uid1))

	req := httptest.NewRequest("POST", "/api/friends/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid2, "user")
	rr := executeRequest(withAuth(testHandler.HandleAddFriend, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var status string
	testDB.QueryRow("SELECT status FROM friendships WHERE user_id = $1 AND friend_id = $2", uid1, uid2).Scan(&status)
	if status != "accepted" {
		t.Fatalf("expected accepted, got %s", status)
	}
}

func TestAddFriendDuplicate(t *testing.T) {
	resetDB(t)
	uid1 := createTestUser(t, "fdup1@test.com", "fdup1", "user")
	uid2 := createTestUser(t, "fdup2@test.com", "fdup2", "user")
	testDB.Exec("INSERT INTO friendships (user_id, friend_id, status) VALUES ($1, $2, 'pending')", uid1, uid2)

	form := url.Values{}
	form.Set("friend_id", itoa(uid2))

	req := httptest.NewRequest("POST", "/api/friends/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid1, "user")
	rr := executeRequest(withAuth(testHandler.HandleAddFriend, testSecret), req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rr.Code)
	}
}

func TestAddFriendSelf(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "selfadd@test.com", "selfadd", "user")

	form := url.Values{}
	form.Set("friend_id", itoa(uid))

	req := httptest.NewRequest("POST", "/api/friends/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleAddFriend, testSecret), req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestAddFriendInvalidID(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "finv@test.com", "finv", "user")

	form := url.Values{}
	form.Set("friend_id", "abc")

	req := httptest.NewRequest("POST", "/api/friends/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleAddFriend, testSecret), req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestRemoveFriend(t *testing.T) {
	resetDB(t)
	uid1 := createTestUser(t, "rm1@test.com", "rmfriend1", "user")
	uid2 := createTestUser(t, "rm2@test.com", "rmfriend2", "user")
	testDB.Exec("INSERT INTO friendships (user_id, friend_id, status) VALUES ($1, $2, 'accepted')", uid1, uid2)

	form := url.Values{}
	form.Set("friend_id", itoa(uid2))

	req := httptest.NewRequest("POST", "/api/friends/remove", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid1, "user")
	rr := executeRequest(withAuth(testHandler.HandleRemoveFriend, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var count int
	testDB.QueryRow("SELECT COUNT(*) FROM friendships WHERE user_id = $1 AND friend_id = $2", uid1, uid2).Scan(&count)
	if count != 0 {
		t.Fatal("expected friendship removed")
	}
}

func TestRemoveFriendInvalidID(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "rminv@test.com", "rminv", "user")

	form := url.Values{}
	form.Set("friend_id", "abc")

	req := httptest.NewRequest("POST", "/api/friends/remove", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleRemoveFriend, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestBlockUser(t *testing.T) {
	resetDB(t)
	uid1 := createTestUser(t, "block1@test.com", "blocker", "user")
	uid2 := createTestUser(t, "block2@test.com", "blockee", "user")

	form := url.Values{}
	form.Set("blocked_id", itoa(uid2))

	req := httptest.NewRequest("POST", "/api/block", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid1, "user")
	rr := executeRequest(withAuth(testHandler.HandleBlockUser, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var exists bool
	testDB.QueryRow("SELECT EXISTS(SELECT 1 FROM blocked_users WHERE user_id = $1 AND blocked_id = $2)", uid1, uid2).Scan(&exists)
	if !exists {
		t.Fatal("expected blocked user record")
	}
}

func TestBlockSelf(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "self@test.com", "selfblock", "user")

	form := url.Values{}
	form.Set("blocked_id", itoa(uid))

	req := httptest.NewRequest("POST", "/api/block", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleBlockUser, testSecret), req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestBlockInvalidID(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "binv@test.com", "binv", "user")

	form := url.Values{}
	form.Set("blocked_id", "abc")

	req := httptest.NewRequest("POST", "/api/block", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleBlockUser, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestUnblockUser(t *testing.T) {
	resetDB(t)
	uid1 := createTestUser(t, "unblock1@test.com", "unblocker", "user")
	uid2 := createTestUser(t, "unblock2@test.com", "unblockee", "user")
	testDB.Exec("INSERT INTO blocked_users (user_id, blocked_id) VALUES ($1, $2)", uid1, uid2)

	form := url.Values{}
	form.Set("blocked_id", itoa(uid2))

	req := httptest.NewRequest("POST", "/api/unblock", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid1, "user")
	rr := executeRequest(withAuth(testHandler.HandleUnblockUser, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var exists bool
	testDB.QueryRow("SELECT EXISTS(SELECT 1 FROM blocked_users WHERE user_id = $1 AND blocked_id = $2)", uid1, uid2).Scan(&exists)
	if exists {
		t.Fatal("expected blocked record removed")
	}
}

func TestUnblockInvalidID(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "uinv@test.com", "uinv", "user")

	form := url.Values{}
	form.Set("blocked_id", "abc")

	req := httptest.NewRequest("POST", "/api/unblock", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleUnblockUser, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestRepost(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "repost@test.com", "reposter", "user")
	uid2 := createTestUser(t, "repost2@test.com", "repostee", "user")
	pid := createTestPost(t, uid2, "Repost Me", "Другое")

	form := url.Values{}
	form.Set("post_id", itoa(pid))
	form.Set("comment", "Great stuff!")

	req := httptest.NewRequest("POST", "/api/repost", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleRepost, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestRepostInvalidPostID(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "repinv@test.com", "repinv", "user")

	form := url.Values{}
	form.Set("post_id", "abc")

	req := httptest.NewRequest("POST", "/api/repost", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleRepost, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestRepostDuplicate(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "repdup@test.com", "repdup", "user")
	uid2 := createTestUser(t, "repdup2@test.com", "repdup2", "user")
	pid := createTestPost(t, uid2, "Post", "Другое")
	testDB.Exec("INSERT INTO reposts (post_id, user_id) VALUES ($1, $2)", pid, uid)

	form := url.Values{}
	form.Set("post_id", itoa(pid))

	req := httptest.NewRequest("POST", "/api/repost", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleRepost, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestComplaint(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "complaint@test.com", "complainer", "user")
	uid2 := createTestUser(t, "target@test.com", "target", "user")
	pid := createTestPost(t, uid2, "Bad Post", "Другое")

	form := url.Values{}
	form.Set("target_type", "post")
	form.Set("target_id", itoa(pid))
	form.Set("category", "spam")
	form.Set("description", "This is spam")

	req := httptest.NewRequest("POST", "/api/complaint", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleComplaint, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestComplaintEmpty(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "emptycomp@test.com", "emptycomp", "user")

	form := url.Values{}
	form.Set("target_type", "post")
	form.Set("target_id", "1")
	form.Set("category", "spam")
	form.Set("description", "")

	req := httptest.NewRequest("POST", "/api/complaint", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleComplaint, testSecret), req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestComplaintUserType(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "compuser@test.com", "compuser", "user")
	uid2 := createTestUser(t, "comptarget@test.com", "comptarget", "user")

	form := url.Values{}
	form.Set("target_type", "user")
	form.Set("target_id", itoa(uid2))
	form.Set("category", "harassment")
	form.Set("description", "Harassing me")

	req := httptest.NewRequest("POST", "/api/complaint", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleComplaint, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestTeamRequestCreate(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "team@test.com", "teamuser", "user")
	pid := createTestPost(t, uid, "Team Post", "Другое")

	form := url.Values{}
	form.Set("post_id", itoa(pid))
	form.Set("title", "Need Developer")
	form.Set("description", "Looking for Go dev")
	form.Set("skills", "Go, PostgreSQL")
	form.Set("role_needed", "Backend Developer")

	req := httptest.NewRequest("POST", "/api/team/request", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleTeamRequestCreate, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestTeamRequestEmptyTitle(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "emptytitle@test.com", "emptytitle", "user")
	pid := createTestPost(t, uid, "Post", "Другое")

	form := url.Values{}
	form.Set("post_id", itoa(pid))
	form.Set("title", "")

	req := httptest.NewRequest("POST", "/api/team/request", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleTeamRequestCreate, testSecret), req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestTeamRequestInvalidPostID(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "tinv@test.com", "tinv", "user")

	form := url.Values{}
	form.Set("post_id", "abc")
	form.Set("title", "Dev")

	req := httptest.NewRequest("POST", "/api/team/request", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleTeamRequestCreate, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestTeamRespond(t *testing.T) {
	resetDB(t)
	uid1 := createTestUser(t, "teamresp1@test.com", "teamreq", "user")
	uid2 := createTestUser(t, "teamresp2@test.com", "teamresp", "user")
	pid := createTestPost(t, uid1, "Post", "Другое")

	var reqID int
	testDB.QueryRow(
		"INSERT INTO team_requests (post_id, author_id, title) VALUES ($1, $2, 'Need help') RETURNING id",
		pid, uid1,
	).Scan(&reqID)

	form := url.Values{}
	form.Set("request_id", itoa(reqID))
	form.Set("message", "I can help!")

	req := httptest.NewRequest("POST", "/api/team/respond", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid2, "user")
	rr := executeRequest(withAuth(testHandler.HandleTeamRespond, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestTeamRespondInvalidRequestID(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "trinv@test.com", "trinv", "user")

	form := url.Values{}
	form.Set("request_id", "abc")
	form.Set("message", "Test")

	req := httptest.NewRequest("POST", "/api/team/respond", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleTeamRespond, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestTeamRespondEmptyMessage(t *testing.T) {
	resetDB(t)
	uid1 := createTestUser(t, "trem1@test.com", "trem1", "user")
	uid2 := createTestUser(t, "trem2@test.com", "trem2", "user")
	pid := createTestPost(t, uid1, "Post", "Другое")

	var reqID int
	testDB.QueryRow("INSERT INTO team_requests (post_id, author_id, title) VALUES ($1, $2, 'Need') RETURNING id", pid, uid1).Scan(&reqID)

	form := url.Values{}
	form.Set("request_id", itoa(reqID))
	form.Set("message", "")

	req := httptest.NewRequest("POST", "/api/team/respond", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid2, "user")
	rr := executeRequest(withAuth(testHandler.HandleTeamRespond, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestExpertApply(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "expert@test.com", "expertapply", "user")

	form := url.Values{}
	form.Set("portfolio", "https://portfolio.example.com")
	form.Set("description", "10 years of experience")

	req := httptest.NewRequest("POST", "/api/expert/apply", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleExpertApply, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestExpertApplyDuplicate(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "dupexpert@test.com", "dupexpert", "user")
	testDB.Exec("INSERT INTO expert_applications (user_id, portfolio, description) VALUES ($1, 'p', 'd')", uid)

	form := url.Values{}
	form.Set("portfolio", "https://portfolio.example.com")
	form.Set("description", "Experience")

	req := httptest.NewRequest("POST", "/api/expert/apply", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleExpertApply, testSecret), req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rr.Code)
	}
}

func TestExpertApplyEmptyFields(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "emptyexp@test.com", "emptyexp", "user")

	form := url.Values{}
	form.Set("portfolio", "")
	form.Set("description", "")

	req := httptest.NewRequest("POST", "/api/expert/apply", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleExpertApply, testSecret), req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestSubscribe(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "sub@test.com", "subscriber", "user")

	form := url.Values{}
	form.Set("plan", "monthly")

	req := httptest.NewRequest("POST", "/api/subscribe", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleSubscribe, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var role string
	testDB.QueryRow("SELECT role FROM users WHERE id = $1", uid).Scan(&role)
	if role != "premium" {
		t.Fatalf("expected premium role, got %s", role)
	}
}

func TestSubscribeInvalidPlan(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "badplan@test.com", "badplan", "user")

	form := url.Values{}
	form.Set("plan", "weekly")

	req := httptest.NewRequest("POST", "/api/subscribe", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleSubscribe, testSecret), req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestSubscribeYearly(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "yearly@test.com", "yearuser", "user")

	form := url.Values{}
	form.Set("plan", "yearly")

	req := httptest.NewRequest("POST", "/api/subscribe", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleSubscribe, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var count int
	testDB.QueryRow("SELECT COUNT(*) FROM subscriptions WHERE user_id = $1 AND plan = 'yearly'", uid).Scan(&count)
	if count != 1 {
		t.Fatal("expected yearly subscription")
	}
}

func TestSubscribeEmptyPlan(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "emptyplan@test.com", "emptyplan", "user")

	form := url.Values{}
	form.Set("plan", "")

	req := httptest.NewRequest("POST", "/api/subscribe", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleSubscribe, testSecret), req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestPremiumStatsWithActiveSubscription(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "statssub@test.com", "statssub", "user")

	_, err := testDB.Exec(
		"INSERT INTO subscriptions (user_id, plan, status, expires_at) VALUES ($1, 'monthly', 'active', NOW() + INTERVAL '10 days')",
		uid,
	)
	if err != nil {
		t.Fatalf("failed to insert subscription: %v", err)
	}

	req := httptest.NewRequest("GET", "/premium-stats", nil)
	req = authRequest(req, uid, "user")
	rr := executeRequest(withPageAuth(testHandler.HandlePremiumStats, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	if !strings.Contains(rr.Body.String(), "Статистика") {
		t.Fatalf("expected stats page content, got: %s", rr.Body.String())
	}
}

func TestPremiumStatsWithoutAccessRedirectsToSettings(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "statsdeny@test.com", "statsdeny", "user")

	req := httptest.NewRequest("GET", "/premium-stats", nil)
	req = authRequest(req, uid, "user")
	rr := executeRequest(withPageAuth(testHandler.HandlePremiumStats, testSecret), req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d", rr.Code)
	}

	if rr.Header().Get("Location") != "/settings" {
		t.Fatalf("expected redirect to /settings, got %s", rr.Header().Get("Location"))
	}
}

func TestTeamSearchPage(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "tsearch@test.com", "tsearch", "user")

	req := httptest.NewRequest("GET", "/teams", nil)
	req = authRequest(req, uid, "user")
	rr := executeRequest(withPageAuth(testHandler.HandleTeamSearchPage, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestTeamSearchPageWithData(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "tsdata@test.com", "tsdata", "user")
	pid := createTestPost(t, uid, "Post", "Другое")
	testDB.Exec("INSERT INTO team_requests (post_id, author_id, title, skills, role_needed, is_open) VALUES ($1, $2, 'Dev needed', 'Go', 'Backend', TRUE)", pid, uid)

	req := httptest.NewRequest("GET", "/teams?search=Dev", nil)
	req = authRequest(req, uid, "user")
	rr := executeRequest(withPageAuth(testHandler.HandleTeamSearchPage, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestHomePage(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	testHandler.HandleHomePage(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHomePageRedirect(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "home@test.com", "homeuser", "user")

	req := httptest.NewRequest("GET", "/", nil)
	req = authRequest(req, uid, "user")
	rr := executeRequest(optionalAuth(testHandler.HandleHomePage), req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected 303 redirect, got %d", rr.Code)
	}
}

func TestAdminPageAccessDenied(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "nonadmin@test.com", "nonadmin", "user")

	req := httptest.NewRequest("GET", "/admin", nil)
	req = authRequest(req, uid, "user")
	rr := executeRequest(withPageAuth(testHandler.HandleAdminPage, testSecret), req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect 303, got %d", rr.Code)
	}
}

func TestAdminPage(t *testing.T) {
	resetDB(t)
	adminID := createTestUser(t, "adminpg@test.com", "adminpg", "admin")

	req := httptest.NewRequest("GET", "/admin", nil)
	req = authRequest(req, adminID, "admin")
	rr := executeRequest(withPageAuth(testHandler.HandleAdminPage, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestAdminPageWithData(t *testing.T) {
	resetDB(t)
	adminID := createTestUser(t, "admpgd@test.com", "admpgd", "admin")
	uid := createTestUser(t, "admpgdu@test.com", "admpgdu", "user")
	createTestPost(t, uid, "Post", "Другое")
	testDB.Exec("INSERT INTO complaints (author_id, target_type, target_id, category, description) VALUES ($1, 'post', 1, 'spam', 'Spam')", uid)
	testDB.Exec("INSERT INTO expert_applications (user_id, portfolio, description) VALUES ($1, 'https://ex.com', 'Exp')", uid)

	req := httptest.NewRequest("GET", "/admin", nil)
	req = authRequest(req, adminID, "admin")
	rr := executeRequest(withPageAuth(testHandler.HandleAdminPage, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestAdminBlockUser(t *testing.T) {
	resetDB(t)
	adminID := createTestUser(t, "admin@test.com", "adminuser", "admin")
	uid := createTestUser(t, "target@test.com", "targetuser", "user")

	form := url.Values{}
	form.Set("user_id", itoa(uid))
	form.Set("blocked", "true")

	req := httptest.NewRequest("POST", "/api/admin/block-user", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, adminID, "admin")
	rr := executeRequest(adminAuth(testHandler.HandleAdminBlockUser), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var isBlocked bool
	testDB.QueryRow("SELECT is_blocked FROM users WHERE id = $1", uid).Scan(&isBlocked)
	if !isBlocked {
		t.Fatal("expected user to be blocked")
	}
}

func TestAdminUnblockUser(t *testing.T) {
	resetDB(t)
	adminID := createTestUser(t, "adminub@test.com", "adminub", "admin")
	uid := createTestUser(t, "ubuser@test.com", "ubuser", "user")
	testDB.Exec("UPDATE users SET is_blocked = TRUE WHERE id = $1", uid)

	form := url.Values{}
	form.Set("user_id", itoa(uid))
	form.Set("blocked", "false")

	req := httptest.NewRequest("POST", "/api/admin/block-user", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, adminID, "admin")
	rr := executeRequest(adminAuth(testHandler.HandleAdminBlockUser), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var isBlocked bool
	testDB.QueryRow("SELECT is_blocked FROM users WHERE id = $1", uid).Scan(&isBlocked)
	if isBlocked {
		t.Fatal("expected user to be unblocked")
	}
}

func TestAdminBlockUserInvalidID(t *testing.T) {
	resetDB(t)
	adminID := createTestUser(t, "adminbinv@test.com", "adminbinv", "admin")

	form := url.Values{}
	form.Set("user_id", "abc")
	form.Set("blocked", "true")

	req := httptest.NewRequest("POST", "/api/admin/block-user", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, adminID, "admin")
	rr := executeRequest(adminAuth(testHandler.HandleAdminBlockUser), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestAdminChangeRole(t *testing.T) {
	resetDB(t)
	adminID := createTestUser(t, "admin2@test.com", "admin2", "admin")
	uid := createTestUser(t, "role@test.com", "roleuser", "user")

	form := url.Values{}
	form.Set("user_id", itoa(uid))
	form.Set("role", "expert")

	req := httptest.NewRequest("POST", "/api/admin/change-role", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, adminID, "admin")
	rr := executeRequest(adminAuth(testHandler.HandleAdminChangeRole), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var role string
	testDB.QueryRow("SELECT role FROM users WHERE id = $1", uid).Scan(&role)
	if role != "expert" {
		t.Fatalf("expected expert, got %s", role)
	}
}

func TestAdminChangeRoleInvalid(t *testing.T) {
	resetDB(t)
	adminID := createTestUser(t, "admin3@test.com", "admin3", "admin")
	uid := createTestUser(t, "badrole@test.com", "badrole", "user")

	form := url.Values{}
	form.Set("user_id", itoa(uid))
	form.Set("role", "superadmin")

	req := httptest.NewRequest("POST", "/api/admin/change-role", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, adminID, "admin")
	rr := executeRequest(adminAuth(testHandler.HandleAdminChangeRole), req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestAdminChangeRoleInvalidUserID(t *testing.T) {
	resetDB(t)
	adminID := createTestUser(t, "adminrinv@test.com", "adminrinv", "admin")

	form := url.Values{}
	form.Set("user_id", "abc")
	form.Set("role", "expert")

	req := httptest.NewRequest("POST", "/api/admin/change-role", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, adminID, "admin")
	rr := executeRequest(adminAuth(testHandler.HandleAdminChangeRole), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestAdminDeletePost(t *testing.T) {
	resetDB(t)
	adminID := createTestUser(t, "admindel@test.com", "admindel", "admin")
	uid := createTestUser(t, "postowner@test.com", "postowner", "user")
	pid := createTestPost(t, uid, "To Delete", "Другое")

	form := url.Values{}
	form.Set("post_id", itoa(pid))

	req := httptest.NewRequest("POST", "/api/admin/delete-post", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, adminID, "admin")
	rr := executeRequest(adminAuth(testHandler.HandleAdminDeletePost), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var count int
	testDB.QueryRow("SELECT COUNT(*) FROM posts WHERE id = $1", pid).Scan(&count)
	if count != 0 {
		t.Fatal("expected post deleted")
	}
}

func TestAdminDeletePostInvalidID(t *testing.T) {
	resetDB(t)
	adminID := createTestUser(t, "admindinv@test.com", "admindinv", "admin")

	form := url.Values{}
	form.Set("post_id", "abc")

	req := httptest.NewRequest("POST", "/api/admin/delete-post", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, adminID, "admin")
	rr := executeRequest(adminAuth(testHandler.HandleAdminDeletePost), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestAdminHidePost(t *testing.T) {
	resetDB(t)
	adminID := createTestUser(t, "adminhide@test.com", "adminhide", "admin")
	uid := createTestUser(t, "hideowner@test.com", "hideowner", "user")
	pid := createTestPost(t, uid, "To Hide", "Другое")

	form := url.Values{}
	form.Set("post_id", itoa(pid))

	req := httptest.NewRequest("POST", "/api/admin/hide-post", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, adminID, "admin")
	rr := executeRequest(adminAuth(testHandler.HandleAdminHidePost), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var isHidden bool
	testDB.QueryRow("SELECT is_hidden FROM posts WHERE id = $1", pid).Scan(&isHidden)
	if !isHidden {
		t.Fatal("expected post hidden")
	}
}

func TestAdminHidePostInvalidID(t *testing.T) {
	resetDB(t)
	adminID := createTestUser(t, "adminhinv@test.com", "adminhinv", "admin")

	form := url.Values{}
	form.Set("post_id", "abc")

	req := httptest.NewRequest("POST", "/api/admin/hide-post", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, adminID, "admin")
	rr := executeRequest(adminAuth(testHandler.HandleAdminHidePost), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestAdminComplaintResolve(t *testing.T) {
	resetDB(t)
	adminID := createTestUser(t, "admcomp@test.com", "admcomp", "admin")
	uid := createTestUser(t, "compauth@test.com", "compauth", "user")
	testDB.Exec("INSERT INTO complaints (author_id, target_type, target_id, category, description) VALUES ($1, 'post', 1, 'spam', 'Spam content')", uid)

	var cid int
	testDB.QueryRow("SELECT id FROM complaints ORDER BY id DESC LIMIT 1").Scan(&cid)

	form := url.Values{}
	form.Set("status", "resolved")

	req := httptest.NewRequest("POST", "/api/admin/complaints/"+itoa(cid)+"/resolve", strings.NewReader(form.Encode()))
	req.URL.Path = "/api/admin/complaints/" + itoa(cid) + "/resolve"
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, adminID, "admin")
	rr := executeRequest(adminAuth(testHandler.HandleAdminComplaint), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var status string
	testDB.QueryRow("SELECT status FROM complaints WHERE id = $1", cid).Scan(&status)
	if status != "resolved" {
		t.Fatalf("expected resolved, got %s", status)
	}
}

func TestAdminComplaintReject(t *testing.T) {
	resetDB(t)
	adminID := createTestUser(t, "admrej@test.com", "admrej", "admin")
	uid := createTestUser(t, "rejauth@test.com", "rejauth", "user")
	testDB.Exec("INSERT INTO complaints (author_id, target_type, target_id, category, description) VALUES ($1, 'user', 1, 'harassment', 'Bad')", uid)

	var cid int
	testDB.QueryRow("SELECT id FROM complaints ORDER BY id DESC LIMIT 1").Scan(&cid)

	form := url.Values{}
	form.Set("status", "rejected")

	req := httptest.NewRequest("POST", "/api/admin/complaints/"+itoa(cid)+"/resolve", strings.NewReader(form.Encode()))
	req.URL.Path = "/api/admin/complaints/" + itoa(cid) + "/resolve"
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, adminID, "admin")
	rr := executeRequest(adminAuth(testHandler.HandleAdminComplaint), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestAdminComplaintInvalidID(t *testing.T) {
	resetDB(t)
	adminID := createTestUser(t, "admcinv@test.com", "admcinv", "admin")

	form := url.Values{}
	form.Set("status", "resolved")

	req := httptest.NewRequest("POST", "/api/admin/complaints/abc/resolve", strings.NewReader(form.Encode()))
	req.URL.Path = "/api/admin/complaints/abc/resolve"
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, adminID, "admin")
	rr := executeRequest(adminAuth(testHandler.HandleAdminComplaint), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestAdminExpertAppApprove(t *testing.T) {
	resetDB(t)
	adminID := createTestUser(t, "admexp@test.com", "admexp", "admin")
	uid := createTestUser(t, "expapp@test.com", "expapp", "user")
	testDB.Exec("INSERT INTO expert_applications (user_id, portfolio, description) VALUES ($1, 'https://ex.com', 'Expert')", uid)

	var appID int
	testDB.QueryRow("SELECT id FROM expert_applications WHERE user_id = $1", uid).Scan(&appID)

	form := url.Values{}
	form.Set("status", "approved")

	req := httptest.NewRequest("POST", "/api/admin/expert-apps/"+itoa(appID)+"/resolve", strings.NewReader(form.Encode()))
	req.URL.Path = "/api/admin/expert-apps/" + itoa(appID) + "/resolve"
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, adminID, "admin")
	rr := executeRequest(adminAuth(testHandler.HandleAdminExpertApp), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var role string
	testDB.QueryRow("SELECT role FROM users WHERE id = $1", uid).Scan(&role)
	if role != "expert" {
		t.Fatalf("expected expert, got %s", role)
	}
}

func TestAdminExpertAppReject(t *testing.T) {
	resetDB(t)
	adminID := createTestUser(t, "admexprej@test.com", "admexprej", "admin")
	uid := createTestUser(t, "exprej@test.com", "exprej", "user")
	testDB.Exec("INSERT INTO expert_applications (user_id, portfolio, description) VALUES ($1, 'https://ex.com', 'Expert')", uid)

	var appID int
	testDB.QueryRow("SELECT id FROM expert_applications WHERE user_id = $1", uid).Scan(&appID)

	form := url.Values{}
	form.Set("status", "rejected")

	req := httptest.NewRequest("POST", "/api/admin/expert-apps/"+itoa(appID)+"/resolve", strings.NewReader(form.Encode()))
	req.URL.Path = "/api/admin/expert-apps/" + itoa(appID) + "/resolve"
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, adminID, "admin")
	rr := executeRequest(adminAuth(testHandler.HandleAdminExpertApp), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var role string
	testDB.QueryRow("SELECT role FROM users WHERE id = $1", uid).Scan(&role)
	if role != "user" {
		t.Fatalf("expected user (not changed), got %s", role)
	}
}

func TestAdminExpertAppInvalidID(t *testing.T) {
	resetDB(t)
	adminID := createTestUser(t, "admeinv@test.com", "admeinv", "admin")

	form := url.Values{}
	form.Set("status", "approved")

	req := httptest.NewRequest("POST", "/api/admin/expert-apps/abc/resolve", strings.NewReader(form.Encode()))
	req.URL.Path = "/api/admin/expert-apps/abc/resolve"
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, adminID, "admin")
	rr := executeRequest(adminAuth(testHandler.HandleAdminExpertApp), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}
