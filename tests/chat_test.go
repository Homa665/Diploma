package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestSendMessage(t *testing.T) {
	resetDB(t)
	uid1 := createTestUser(t, "sender@test.com", "sender", "user")
	uid2 := createTestUser(t, "receiver@test.com", "receiver", "user")

	form := url.Values{}
	form.Set("receiver_id", itoa(uid2))
	form.Set("content", "Hello!")

	req := httptest.NewRequest("POST", "/api/messages/send", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid1, "user")
	rr := executeRequest(withAuth(testHandler.HandleSendMessage, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var count int
	testDB.QueryRow("SELECT COUNT(*) FROM chat_messages WHERE sender_id = $1 AND receiver_id = $2", uid1, uid2).Scan(&count)
	if count != 1 {
		t.Fatal("expected 1 message")
	}
}

func TestSendMessageEmpty(t *testing.T) {
	resetDB(t)
	uid1 := createTestUser(t, "empty1@test.com", "emptymsg1", "user")
	uid2 := createTestUser(t, "empty2@test.com", "emptymsg2", "user")

	form := url.Values{}
	form.Set("receiver_id", itoa(uid2))
	form.Set("content", "")

	req := httptest.NewRequest("POST", "/api/messages/send", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid1, "user")
	rr := executeRequest(withAuth(testHandler.HandleSendMessage, testSecret), req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestSendMessageBlocked(t *testing.T) {
	resetDB(t)
	uid1 := createTestUser(t, "blocked1@test.com", "blockedsender", "user")
	uid2 := createTestUser(t, "blocked2@test.com", "blockedreceiver", "user")
	testDB.Exec("INSERT INTO blocked_users (user_id, blocked_id) VALUES ($1, $2)", uid2, uid1)

	form := url.Values{}
	form.Set("receiver_id", itoa(uid2))
	form.Set("content", "Hello blocked!")

	req := httptest.NewRequest("POST", "/api/messages/send", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid1, "user")
	rr := executeRequest(withAuth(testHandler.HandleSendMessage, testSecret), req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestSendMessageInvalidReceiverID(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "sinv@test.com", "sinvuser", "user")

	form := url.Values{}
	form.Set("receiver_id", "abc")
	form.Set("content", "Hello!")

	req := httptest.NewRequest("POST", "/api/messages/send", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleSendMessage, testSecret), req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestSendMessageToSelf(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "selfmsg@test.com", "selfmsg", "user")

	form := url.Values{}
	form.Set("receiver_id", itoa(uid))
	form.Set("content", "Talking to myself")

	req := httptest.NewRequest("POST", "/api/messages/send", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleSendMessage, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestSendMessageNotification(t *testing.T) {
	resetDB(t)
	uid1 := createTestUser(t, "notif1@test.com", "notifmsg1", "user")
	uid2 := createTestUser(t, "notif2@test.com", "notifmsg2", "user")

	form := url.Values{}
	form.Set("receiver_id", itoa(uid2))
	form.Set("content", "Check notifications!")

	req := httptest.NewRequest("POST", "/api/messages/send", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = authRequest(req, uid1, "user")
	executeRequest(withAuth(testHandler.HandleSendMessage, testSecret), req)

	var count int
	testDB.QueryRow("SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND type = 'message'", uid2).Scan(&count)
	if count != 1 {
		t.Fatal("expected message notification")
	}
}

func TestGetMessages(t *testing.T) {
	resetDB(t)
	uid1 := createTestUser(t, "get1@test.com", "getmsg1", "user")
	uid2 := createTestUser(t, "get2@test.com", "getmsg2", "user")
	testDB.Exec("INSERT INTO chat_messages (sender_id, receiver_id, content) VALUES ($1, $2, 'msg1')", uid1, uid2)
	testDB.Exec("INSERT INTO chat_messages (sender_id, receiver_id, content) VALUES ($1, $2, 'msg2')", uid2, uid1)

	req := httptest.NewRequest("GET", "/api/messages?partner_id="+itoa(uid2), nil)
	req = authRequest(req, uid1, "user")
	rr := executeRequest(withAuth(testHandler.HandleGetMessages, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestGetMessagesInvalidPartnerID(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "ginv@test.com", "ginv", "user")

	req := httptest.NewRequest("GET", "/api/messages?partner_id=abc", nil)
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleGetMessages, testSecret), req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestGetMessagesNoPartnerID(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "gnone@test.com", "gnone", "user")

	req := httptest.NewRequest("GET", "/api/messages", nil)
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleGetMessages, testSecret), req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestGetMessagesMarksRead(t *testing.T) {
	resetDB(t)
	uid1 := createTestUser(t, "read1@test.com", "readmsg1", "user")
	uid2 := createTestUser(t, "read2@test.com", "readmsg2", "user")
	testDB.Exec("INSERT INTO chat_messages (sender_id, receiver_id, content, is_read) VALUES ($1, $2, 'unread', FALSE)", uid2, uid1)

	req := httptest.NewRequest("GET", "/api/messages?partner_id="+itoa(uid2), nil)
	req = authRequest(req, uid1, "user")
	executeRequest(withAuth(testHandler.HandleGetMessages, testSecret), req)

	var isRead bool
	testDB.QueryRow("SELECT is_read FROM chat_messages WHERE sender_id = $1 AND receiver_id = $2", uid2, uid1).Scan(&isRead)
	if !isRead {
		t.Fatal("expected message to be marked read")
	}
}

func TestChatPage(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "chat@test.com", "chatuser", "user")

	req := httptest.NewRequest("GET", "/chat", nil)
	req = authRequest(req, uid, "user")
	rr := executeRequest(withPageAuth(testHandler.HandleChatPage, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestChatPageWithConversations(t *testing.T) {
	resetDB(t)
	uid1 := createTestUser(t, "chatconv1@test.com", "chatconv1", "user")
	uid2 := createTestUser(t, "chatconv2@test.com", "chatconv2", "user")
	testDB.Exec("INSERT INTO chat_messages (sender_id, receiver_id, content) VALUES ($1, $2, 'Hey')", uid1, uid2)
	testDB.Exec("INSERT INTO chat_messages (sender_id, receiver_id, content) VALUES ($1, $2, 'Hi')", uid2, uid1)

	req := httptest.NewRequest("GET", "/chat", nil)
	req = authRequest(req, uid1, "user")
	rr := executeRequest(withPageAuth(testHandler.HandleChatPage, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestChatWithPage(t *testing.T) {
	resetDB(t)
	uid1 := createTestUser(t, "chatwith1@test.com", "chatwith1", "user")
	uid2 := createTestUser(t, "chatwith2@test.com", "chatwith2", "user")

	req := httptest.NewRequest("GET", "/chat/"+itoa(uid2), nil)
	req.URL.Path = "/chat/" + itoa(uid2)
	req = authRequest(req, uid1, "user")
	rr := executeRequest(withPageAuth(testHandler.HandleChatWith, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestChatWithPageInvalidID(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "chatinv@test.com", "chatinv", "user")

	req := httptest.NewRequest("GET", "/chat/abc", nil)
	req.URL.Path = "/chat/abc"
	req = authRequest(req, uid, "user")
	rr := executeRequest(withPageAuth(testHandler.HandleChatWith, testSecret), req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestChatWithPageNotFound(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "chatnf@test.com", "chatnf", "user")

	req := httptest.NewRequest("GET", "/chat/99999", nil)
	req.URL.Path = "/chat/99999"
	req = authRequest(req, uid, "user")
	rr := executeRequest(withPageAuth(testHandler.HandleChatWith, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestChatWithPageMessages(t *testing.T) {
	resetDB(t)
	uid1 := createTestUser(t, "chatwm1@test.com", "chatwm1", "user")
	uid2 := createTestUser(t, "chatwm2@test.com", "chatwm2", "user")
	testDB.Exec("INSERT INTO chat_messages (sender_id, receiver_id, content) VALUES ($1, $2, 'Test msg')", uid1, uid2)

	req := httptest.NewRequest("GET", "/chat/"+itoa(uid2), nil)
	req.URL.Path = "/chat/" + itoa(uid2)
	req = authRequest(req, uid1, "user")
	rr := executeRequest(withPageAuth(testHandler.HandleChatWith, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestGetNotifications(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "notif@test.com", "notifuser", "user")
	testDB.Exec("INSERT INTO notifications (user_id, type, content, link) VALUES ($1, 'test', 'Test notification', '/test')", uid)

	req := httptest.NewRequest("GET", "/api/notifications", nil)
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleGetNotifications, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp["unread_count"] == nil {
		t.Fatal("expected unread_count in response")
	}
}

func TestGetNotificationsEmpty(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "notifemp@test.com", "notifemp", "user")

	req := httptest.NewRequest("GET", "/api/notifications", nil)
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleGetNotifications, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestMarkNotificationRead(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "mark@test.com", "markread", "user")
	var nid int
	testDB.QueryRow("INSERT INTO notifications (user_id, type, content) VALUES ($1, 'test', 'Test') RETURNING id", uid).Scan(&nid)

	req := httptest.NewRequest("POST", "/api/notifications/"+itoa(nid)+"/read", nil)
	req.URL.Path = "/api/notifications/" + itoa(nid) + "/read"
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleMarkNotificationRead, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var isRead bool
	testDB.QueryRow("SELECT is_read FROM notifications WHERE id = $1", nid).Scan(&isRead)
	if !isRead {
		t.Fatal("expected notification to be marked read")
	}
}

func TestMarkNotificationReadInvalidID(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "markinv@test.com", "markinv", "user")

	req := httptest.NewRequest("POST", "/api/notifications/abc/read", nil)
	req.URL.Path = "/api/notifications/abc/read"
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleMarkNotificationRead, testSecret), req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestMarkAllNotificationsRead(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "markall@test.com", "markall", "user")
	testDB.Exec("INSERT INTO notifications (user_id, type, content) VALUES ($1, 'test', 'Test1')", uid)
	testDB.Exec("INSERT INTO notifications (user_id, type, content) VALUES ($1, 'test', 'Test2')", uid)

	req := httptest.NewRequest("POST", "/api/notifications/read-all", nil)
	req = authRequest(req, uid, "user")
	rr := executeRequest(withAuth(testHandler.HandleMarkAllNotificationsRead, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var count int
	testDB.QueryRow("SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = FALSE", uid).Scan(&count)
	if count != 0 {
		t.Fatalf("expected 0 unread, got %d", count)
	}
}

func TestNotificationsPage(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "npage@test.com", "npage", "user")

	req := httptest.NewRequest("GET", "/notifications", nil)
	req = authRequest(req, uid, "user")
	rr := executeRequest(withPageAuth(testHandler.HandleNotificationsPage, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestNotificationsPageWithData(t *testing.T) {
	resetDB(t)
	uid := createTestUser(t, "npdata@test.com", "npdata", "user")
	testDB.Exec("INSERT INTO notifications (user_id, type, content, link) VALUES ($1, 'comment', 'New comment', '/post/1')", uid)
	testDB.Exec("INSERT INTO notifications (user_id, type, content, link, is_read) VALUES ($1, 'message', 'New msg', '/chat/1', TRUE)", uid)

	req := httptest.NewRequest("GET", "/notifications", nil)
	req = authRequest(req, uid, "user")
	rr := executeRequest(withPageAuth(testHandler.HandleNotificationsPage, testSecret), req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}
