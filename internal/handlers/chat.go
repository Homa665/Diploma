package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"startup-platform/internal/models"
)

func (h *Handler) HandleChatPage(w http.ResponseWriter, r *http.Request) {
	userID, role := h.getCurrentUser(r)

	rows, err := h.DB.Query(
		`SELECT DISTINCT ON (partner_id) partner_id, nickname, avatar_url, last_msg, last_time, unread_count
		FROM (
			SELECT 
				CASE WHEN sender_id = $1 THEN receiver_id ELSE sender_id END as partner_id,
				content as last_msg,
				created_at as last_time
			FROM chat_messages 
			WHERE sender_id = $1 OR receiver_id = $1
			ORDER BY created_at DESC
		) sub
		JOIN users u ON u.id = sub.partner_id
		LEFT JOIN LATERAL (
			SELECT COUNT(*) as unread_count FROM chat_messages 
			WHERE sender_id = sub.partner_id AND receiver_id = $1 AND is_read = FALSE
		) uc ON TRUE
		ORDER BY partner_id, last_time DESC`,
		userID,
	)

	type ChatItem struct {
		PartnerID       int       `json:"partner_id"`
		PartnerNickname string    `json:"partner_nickname"`
		AvatarURL       string    `json:"avatar_url"`
		LastMessage     string    `json:"last_message"`
		LastMessageAt   time.Time `json:"last_message_at"`
		UnreadCount     int       `json:"unread_count"`
	}

	var chats []ChatItem
	if err == nil {
		for rows.Next() {
			var ci ChatItem
			var lastMsg string
			var lastTime time.Time
			var unread int
			rows.Scan(&ci.PartnerID, &ci.PartnerNickname, &ci.AvatarURL, &lastMsg, &lastTime, &unread)
			ci.LastMessage = lastMsg
			ci.LastMessageAt = lastTime
			ci.UnreadCount = unread
			chats = append(chats, ci)
		}
		rows.Close()
	}

	if chats == nil {
		rows2, err2 := h.DB.Query(
			`SELECT DISTINCT 
				CASE WHEN sender_id = $1 THEN receiver_id ELSE sender_id END as partner_id
			FROM chat_messages WHERE sender_id = $1 OR receiver_id = $1`,
			userID,
		)
		if err2 == nil {
			for rows2.Next() {
				var pid int
				rows2.Scan(&pid)
				var ci ChatItem
				ci.PartnerID = pid
				h.DB.QueryRow("SELECT nickname, avatar_url FROM users WHERE id = $1", pid).Scan(&ci.PartnerNickname, &ci.AvatarURL)
				var unread int
				h.DB.QueryRow("SELECT COUNT(*) FROM chat_messages WHERE sender_id = $1 AND receiver_id = $2 AND is_read = FALSE", pid, userID).Scan(&unread)
				ci.UnreadCount = unread
				h.DB.QueryRow("SELECT content, created_at FROM chat_messages WHERE (sender_id = $1 AND receiver_id = $2) OR (sender_id = $2 AND receiver_id = $1) ORDER BY created_at DESC LIMIT 1", userID, pid).Scan(&ci.LastMessage, &ci.LastMessageAt)
				chats = append(chats, ci)
			}
			rows2.Close()
		}
	}

	var nickname string
	h.DB.QueryRow("SELECT nickname FROM users WHERE id = $1", userID).Scan(&nickname)

	ban := h.getActiveBan(userID)

	data := map[string]interface{}{
		"Chats":     chats,
		"UserID":    userID,
		"UserRole":  role,
		"Nickname":  nickname,
		"ActiveBan": ban,
	}
	h.renderTemplate(w, "chat.html", data)
}

func (h *Handler) HandleChatWith(w http.ResponseWriter, r *http.Request) {
	userID, role := h.getCurrentUser(r)
	partnerIDStr := strings.TrimPrefix(r.URL.Path, "/chat/")
	partnerID, err := strconv.Atoi(partnerIDStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	var blockedByPartner bool
	h.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM blocked_users WHERE user_id = $1 AND blocked_id = $2)", partnerID, userID).Scan(&blockedByPartner)

	var blockedByMe bool
	h.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM blocked_users WHERE user_id = $1 AND blocked_id = $2)", userID, partnerID).Scan(&blockedByMe)

	h.DB.Exec("UPDATE chat_messages SET is_read = TRUE WHERE sender_id = $1 AND receiver_id = $2", partnerID, userID)

	key := fmt.Sprintf("%d:%d", userID, partnerID)
	h.ActiveChatUsers.Store(key, time.Now())

	rows, err := h.DB.Query(
		`SELECT cm.id, cm.sender_id, cm.receiver_id, cm.content, cm.is_read, cm.created_at,
		u.nickname, u.avatar_url
		FROM chat_messages cm JOIN users u ON cm.sender_id = u.id
		WHERE (cm.sender_id = $1 AND cm.receiver_id = $2) OR (cm.sender_id = $2 AND cm.receiver_id = $1)
		ORDER BY cm.created_at ASC`,
		userID, partnerID,
	)

	var messages []models.ChatMessage
	if err == nil {
		for rows.Next() {
			var m models.ChatMessage
			rows.Scan(&m.ID, &m.SenderID, &m.ReceiverID, &m.Content, &m.IsRead, &m.CreatedAt,
				&m.SenderNickname, &m.SenderAvatarURL)
			messages = append(messages, m)
		}
		rows.Close()
	}

	var partnerNickname, partnerAvatar string
	h.DB.QueryRow("SELECT nickname, avatar_url FROM users WHERE id = $1", partnerID).Scan(&partnerNickname, &partnerAvatar)

	var nickname string
	h.DB.QueryRow("SELECT nickname FROM users WHERE id = $1", userID).Scan(&nickname)

	ban := h.getActiveBan(userID)

	data := map[string]interface{}{
		"Messages":         messages,
		"PartnerID":        partnerID,
		"PartnerNickname":  partnerNickname,
		"PartnerAvatar":    partnerAvatar,
		"UserID":           userID,
		"UserRole":         role,
		"Nickname":         nickname,
		"BlockedByPartner": blockedByPartner,
		"BlockedByMe":      blockedByMe,
		"ActiveBan":        ban,
	}
	h.renderTemplate(w, "chat_with.html", data)
}

func (h *Handler) HandleSendMessage(w http.ResponseWriter, r *http.Request) {
	userID, _ := h.getCurrentUser(r)
	r.ParseForm()

	receiverIDStr := r.FormValue("receiver_id")
	receiverID, err := strconv.Atoi(receiverIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Некорректный получатель")
		return
	}

	content := strings.TrimSpace(r.FormValue("content"))
	if content == "" {
		h.respondError(w, http.StatusBadRequest, "Сообщение не может быть пустым")
		return
	}

	var blocked bool
	h.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM blocked_users WHERE user_id = $1 AND blocked_id = $2)", receiverID, userID).Scan(&blocked)
	if blocked {
		h.respondError(w, http.StatusForbidden, "Этот пользователь вас заблокировал. Вы не можете отправить ему сообщение.")
		return
	}

	var blockedByMe bool
	h.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM blocked_users WHERE user_id = $1 AND blocked_id = $2)", userID, receiverID).Scan(&blockedByMe)
	if blockedByMe {
		h.respondError(w, http.StatusForbidden, "Вы заблокировали этого пользователя. Разблокируйте для отправки сообщений.")
		return
	}

	var msgID int
	h.DB.QueryRow(
		"INSERT INTO chat_messages (sender_id, receiver_id, content) VALUES ($1, $2, $3) RETURNING id",
		userID, receiverID, content,
	).Scan(&msgID)

	shouldNotify := true
	key := fmt.Sprintf("%d:%d", receiverID, userID)
	if val, ok := h.ActiveChatUsers.Load(key); ok {
		lastActive := val.(time.Time)
		if time.Since(lastActive) < 30*time.Second {
			shouldNotify = false
		}
	}

	if shouldNotify {
		var nickname string
		h.DB.QueryRow("SELECT nickname FROM users WHERE id = $1", userID).Scan(&nickname)
		h.DB.Exec(
			"INSERT INTO notifications (user_id, type, content, link) VALUES ($1, 'message', $2, $3)",
			receiverID, "Новое сообщение от "+nickname, "/chat/"+strconv.Itoa(userID),
		)
	}

	var createdAt time.Time
	h.DB.QueryRow("SELECT created_at FROM chat_messages WHERE id = $1", msgID).Scan(&createdAt)

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"success":    true,
		"message_id": msgID,
		"created_at": createdAt.Format("15:04"),
	})
}

func (h *Handler) HandleGetMessages(w http.ResponseWriter, r *http.Request) {
	userID, _ := h.getCurrentUser(r)
	partnerIDStr := r.URL.Query().Get("partner_id")
	partnerID, err := strconv.Atoi(partnerIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Некорректный ID")
		return
	}

	h.DB.Exec("UPDATE chat_messages SET is_read = TRUE WHERE sender_id = $1 AND receiver_id = $2", partnerID, userID)

	key := fmt.Sprintf("%d:%d", userID, partnerID)
	h.ActiveChatUsers.Store(key, time.Now())

	rows, err := h.DB.Query(
		`SELECT cm.id, cm.sender_id, cm.receiver_id, cm.content, cm.is_read, cm.created_at,
		u.nickname, u.avatar_url
		FROM chat_messages cm JOIN users u ON cm.sender_id = u.id
		WHERE (cm.sender_id = $1 AND cm.receiver_id = $2) OR (cm.sender_id = $2 AND cm.receiver_id = $1)
		ORDER BY cm.created_at ASC`,
		userID, partnerID,
	)

	var messages []models.ChatMessage
	if err == nil {
		for rows.Next() {
			var m models.ChatMessage
			rows.Scan(&m.ID, &m.SenderID, &m.ReceiverID, &m.Content, &m.IsRead, &m.CreatedAt,
				&m.SenderNickname, &m.SenderAvatarURL)
			messages = append(messages, m)
		}
		rows.Close()
	}

	if messages == nil {
		messages = []models.ChatMessage{}
	}

	h.respondJSON(w, http.StatusOK, messages)
}

func (h *Handler) HandleGetNotifications(w http.ResponseWriter, r *http.Request) {
	userID, _ := h.getCurrentUser(r)

	rows, err := h.DB.Query(
		"SELECT id, user_id, type, content, link, is_read, created_at FROM notifications WHERE user_id = $1 ORDER BY created_at DESC LIMIT 50",
		userID,
	)

	var notifications []models.Notification
	if err == nil {
		for rows.Next() {
			var n models.Notification
			rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Content, &n.Link, &n.IsRead, &n.CreatedAt)
			notifications = append(notifications, n)
		}
		rows.Close()
	}

	if notifications == nil {
		notifications = []models.Notification{}
	}

	var unread int
	for _, n := range notifications {
		if !n.IsRead {
			unread++
		}
	}
	h.respondJSON(w, http.StatusOK, map[string]interface{}{"notifications": notifications, "unread_count": unread})
}

func (h *Handler) HandleMarkNotificationRead(w http.ResponseWriter, r *http.Request) {
	userID, _ := h.getCurrentUser(r)
	notifIDStr := strings.TrimPrefix(r.URL.Path, "/api/notifications/")
	notifIDStr = strings.TrimSuffix(notifIDStr, "/read")
	notifID, err := strconv.Atoi(notifIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Некорректный ID")
		return
	}

	h.DB.Exec("UPDATE notifications SET is_read = TRUE WHERE id = $1 AND user_id = $2", notifID, userID)
	h.respondJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}

func (h *Handler) HandleMarkAllNotificationsRead(w http.ResponseWriter, r *http.Request) {
	userID, _ := h.getCurrentUser(r)
	h.DB.Exec("UPDATE notifications SET is_read = TRUE WHERE user_id = $1", userID)
	h.respondJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}

func (h *Handler) HandleShareToChat(w http.ResponseWriter, r *http.Request) {
	userID, _ := h.getCurrentUser(r)
	r.ParseForm()

	postIDStr := r.FormValue("post_id")
	postID, _ := strconv.Atoi(postIDStr)
	receiverIDStr := r.FormValue("receiver_id")
	receiverID, _ := strconv.Atoi(receiverIDStr)

	var title string
	h.DB.QueryRow("SELECT title FROM posts WHERE id = $1", postID).Scan(&title)

	content := fmt.Sprintf("[Поделился постом: %s](/post/%d)", title, postID)

	h.DB.Exec(
		"INSERT INTO chat_messages (sender_id, receiver_id, content) VALUES ($1, $2, $3)",
		userID, receiverID, content,
	)

	var nickname string
	h.DB.QueryRow("SELECT nickname FROM users WHERE id = $1", userID).Scan(&nickname)
	h.DB.Exec(
		"INSERT INTO notifications (user_id, type, content, link) VALUES ($1, 'message', $2, $3)",
		receiverID, nickname+" поделился с вами постом", "/chat/"+strconv.Itoa(userID),
	)

	h.logActivity(r, userID, "share_post", "post", postID, "to user "+receiverIDStr)

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}
