package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"startup-platform/internal/models"
)

func (h *Handler) HandleAdminPage(w http.ResponseWriter, r *http.Request) {
	userID, role := h.getCurrentUser(r)
	if role != "admin" {
		http.Redirect(w, r, "/feed", http.StatusSeeOther)
		return
	}

	tab := r.URL.Query().Get("tab")
	if tab == "" {
		tab = "users"
	}

	var stats models.AdminStats
	h.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&stats.TotalUsers)
	h.DB.QueryRow("SELECT COUNT(*) FROM posts").Scan(&stats.TotalPosts)
	h.DB.QueryRow("SELECT COUNT(*) FROM comments").Scan(&stats.TotalComments)
	h.DB.QueryRow("SELECT COUNT(*) FROM complaints").Scan(&stats.TotalComplaints)
	h.DB.QueryRow("SELECT COUNT(*) FROM complaints WHERE status = 'pending'").Scan(&stats.PendingComplaints)
	h.DB.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'expert'").Scan(&stats.TotalExperts)
	h.DB.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'premium'").Scan(&stats.TotalPremium)
	h.DB.QueryRow("SELECT COUNT(*) FROM users WHERE created_at >= CURRENT_DATE").Scan(&stats.NewUsersToday)
	h.DB.QueryRow("SELECT COUNT(*) FROM posts WHERE created_at >= CURRENT_DATE").Scan(&stats.NewPostsToday)
	h.DB.QueryRow("SELECT COUNT(*) FROM expert_applications WHERE status = 'pending'").Scan(&stats.PendingExpertApps)

	userRows, _ := h.DB.Query(
		"SELECT id, email, nickname, name, role, is_blocked, created_at FROM users ORDER BY created_at DESC LIMIT 100",
	)
	var users []models.User
	if userRows != nil {
		for userRows.Next() {
			var u models.User
			userRows.Scan(&u.ID, &u.Email, &u.Nickname, &u.Name, &u.Role, &u.IsBlocked, &u.CreatedAt)
			users = append(users, u)
		}
		userRows.Close()
	}

	complaintRows, _ := h.DB.Query(
		`SELECT c.id, c.author_id, c.target_type, c.target_id, c.category, c.description, c.status, c.created_at,
		u.nickname as author_nickname
		FROM complaints c JOIN users u ON c.author_id = u.id
		ORDER BY CASE WHEN c.status = 'pending' THEN 0 ELSE 1 END, c.created_at DESC LIMIT 100`,
	)
	var complaints []models.Complaint
	if complaintRows != nil {
		for complaintRows.Next() {
			var c models.Complaint
			complaintRows.Scan(&c.ID, &c.AuthorID, &c.TargetType, &c.TargetID, &c.Category,
				&c.Description, &c.Status, &c.CreatedAt, &c.AuthorNickname)

			switch c.TargetType {
			case "user":
				h.DB.QueryRow("SELECT nickname FROM users WHERE id = $1", c.TargetID).Scan(&c.TargetNickname)
				c.TargetContent = "Профиль пользователя"
			case "post":
				var title, authorNick string
				h.DB.QueryRow("SELECT p.title, u.nickname FROM posts p JOIN users u ON p.author_id = u.id WHERE p.id = $1", c.TargetID).Scan(&title, &authorNick)
				c.TargetContent = title
				c.TargetNickname = authorNick
			case "comment":
				var content, authorNick string
				h.DB.QueryRow("SELECT cm.content, u.nickname FROM comments cm JOIN users u ON cm.author_id = u.id WHERE cm.id = $1", c.TargetID).Scan(&content, &authorNick)
				c.TargetContent = content
				c.TargetNickname = authorNick
			}

			complaints = append(complaints, c)
		}
		complaintRows.Close()
	}

	expertRows, _ := h.DB.Query(
		`SELECT ea.id, ea.user_id, ea.portfolio, ea.description, ea.status, ea.created_at,
		u.nickname, u.email
		FROM expert_applications ea JOIN users u ON ea.user_id = u.id
		ORDER BY CASE WHEN ea.status = 'pending' THEN 0 ELSE 1 END, ea.created_at DESC`,
	)
	type ExpertApp struct {
		ID          int       `json:"id"`
		UserID      int       `json:"user_id"`
		Portfolio   string    `json:"portfolio"`
		Description string    `json:"description"`
		Status      string    `json:"status"`
		CreatedAt   time.Time `json:"created_at"`
		Nickname    string    `json:"nickname"`
		Email       string    `json:"email"`
	}
	var expertApps []ExpertApp
	if expertRows != nil {
		for expertRows.Next() {
			var ea ExpertApp
			expertRows.Scan(&ea.ID, &ea.UserID, &ea.Portfolio, &ea.Description, &ea.Status, &ea.CreatedAt, &ea.Nickname, &ea.Email)
			expertApps = append(expertApps, ea)
		}
		expertRows.Close()
	}

	postRows, _ := h.DB.Query(
		`SELECT p.id, p.title, p.category, p.is_hidden, p.created_at, u.nickname
		FROM posts p JOIN users u ON p.author_id = u.id
		ORDER BY p.created_at DESC LIMIT 100`,
	)
	type AdminPost struct {
		ID       int
		Title    string
		Category string
		IsHidden bool
		Author   string
	}
	var posts []AdminPost
	if postRows != nil {
		for postRows.Next() {
			var ap AdminPost
			var createdAt time.Time
			postRows.Scan(&ap.ID, &ap.Title, &ap.Category, &ap.IsHidden, &createdAt, &ap.Author)
			posts = append(posts, ap)
		}
		postRows.Close()
	}

	logRows, _ := h.DB.Query(
		`SELECT al.id, al.user_id, al.action, al.target_type, al.target_id, al.details, al.ip_address, al.created_at,
		COALESCE(u.nickname, 'удалён')
		FROM activity_logs al LEFT JOIN users u ON al.user_id = u.id
		ORDER BY al.created_at DESC LIMIT 200`,
	)
	var logs []models.ActivityLog
	if logRows != nil {
		for logRows.Next() {
			var l models.ActivityLog
			logRows.Scan(&l.ID, &l.UserID, &l.Action, &l.TargetType, &l.TargetID, &l.Details, &l.IPAddress, &l.CreatedAt, &l.Nickname)
			logs = append(logs, l)
		}
		logRows.Close()
	}

	banRows, _ := h.DB.Query(
		`SELECT ub.id, ub.user_id, ub.reason, ub.restriction, ub.target_link, ub.admin_comment, ub.expires_at, ub.created_at,
		u.nickname
		FROM user_bans ub JOIN users u ON ub.user_id = u.id
		ORDER BY ub.created_at DESC LIMIT 100`,
	)
	var bans []models.UserBan
	if banRows != nil {
		for banRows.Next() {
			var b models.UserBan
			banRows.Scan(&b.ID, &b.UserID, &b.Reason, &b.Restriction, &b.TargetLink, &b.AdminComment, &b.ExpiresAt, &b.CreatedAt, &b.Nickname)
			b.IsActive = b.ExpiresAt.After(time.Now())
			bans = append(bans, b)
		}
		banRows.Close()
	}

	var nickname string
	h.DB.QueryRow("SELECT nickname FROM users WHERE id = $1", userID).Scan(&nickname)

	data := map[string]interface{}{
		"Stats":      stats,
		"Users":      users,
		"Complaints": complaints,
		"ExpertApps": expertApps,
		"Posts":      posts,
		"Logs":       logs,
		"Bans":       bans,
		"Tab":        tab,
		"UserID":     userID,
		"UserRole":   role,
		"Nickname":   nickname,
	}
	h.renderTemplate(w, "admin.html", data)
}

func (h *Handler) HandleAdminBlockUser(w http.ResponseWriter, r *http.Request) {
	adminID, _ := h.getCurrentUser(r)
	r.ParseForm()
	targetIDStr := r.FormValue("user_id")
	targetID, _ := strconv.Atoi(targetIDStr)
	block := r.FormValue("blocked") == "true"

	h.DB.Exec("UPDATE users SET is_blocked = $1 WHERE id = $2", block, targetID)
	h.logActivity(r, adminID, "admin_block_user", "user", targetID, r.FormValue("blocked"))
	h.respondJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}

func (h *Handler) HandleAdminBanUser(w http.ResponseWriter, r *http.Request) {
	adminID, _ := h.getCurrentUser(r)
	r.ParseForm()

	userIDStr := r.FormValue("user_id")
	userBanID, _ := strconv.Atoi(userIDStr)
	reason := strings.TrimSpace(r.FormValue("reason"))
	restriction := r.FormValue("restriction")
	daysStr := r.FormValue("days")
	days, _ := strconv.Atoi(daysStr)
	targetLink := strings.TrimSpace(r.FormValue("target_link"))
	adminComment := strings.TrimSpace(r.FormValue("admin_comment"))

	if reason == "" || days < 1 {
		h.respondError(w, http.StatusBadRequest, "Укажите причину и срок")
		return
	}

	validRestrictions := map[string]bool{"comments": true, "posts": true, "full": true}
	if !validRestrictions[restriction] {
		h.respondError(w, http.StatusBadRequest, "Некорректный тип ограничения")
		return
	}

	h.DB.Exec(
		`INSERT INTO user_bans (user_id, admin_id, reason, restriction, target_link, admin_comment, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW() + INTERVAL '1 day' * $7)`,
		userBanID, adminID, reason, restriction, targetLink, adminComment, days,
	)

	h.logActivity(r, adminID, "admin_ban", "user", userBanID, restriction+" "+daysStr+"d: "+reason)

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}

func (h *Handler) HandleAdminChangeRole(w http.ResponseWriter, r *http.Request) {
	adminID, _ := h.getCurrentUser(r)
	r.ParseForm()
	targetIDStr := r.FormValue("user_id")
	targetID, _ := strconv.Atoi(targetIDStr)
	newRole := r.FormValue("role")

	validRoles := map[string]bool{"user": true, "premium": true, "expert": true, "admin": true}
	if !validRoles[newRole] {
		h.respondError(w, http.StatusBadRequest, "Некорректная роль")
		return
	}

	h.DB.Exec("UPDATE users SET role = $1 WHERE id = $2", newRole, targetID)
	h.logActivity(r, adminID, "admin_change_role", "user", targetID, newRole)
	h.respondJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}

func (h *Handler) HandleAdminComplaint(w http.ResponseWriter, r *http.Request) {
	adminID, _ := h.getCurrentUser(r)
	r.ParseForm()
	complaintIDStr := strings.TrimPrefix(r.URL.Path, "/api/admin/complaints/")
	complaintIDStr = strings.TrimSuffix(complaintIDStr, "/resolve")
	complaintID, _ := strconv.Atoi(complaintIDStr)
	status := r.FormValue("status")

	h.DB.Exec("UPDATE complaints SET status = $1 WHERE id = $2", status, complaintID)
	h.logActivity(r, adminID, "admin_resolve_complaint", "complaint", complaintID, status)
	h.respondJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}

func (h *Handler) HandleAdminExpertApp(w http.ResponseWriter, r *http.Request) {
	adminID, _ := h.getCurrentUser(r)
	r.ParseForm()
	appIDStr := strings.TrimPrefix(r.URL.Path, "/api/admin/expert-apps/")
	appIDStr = strings.TrimSuffix(appIDStr, "/resolve")
	appID, _ := strconv.Atoi(appIDStr)
	action := r.FormValue("status")

	if action == "approved" {
		var uid int
		h.DB.QueryRow("SELECT user_id FROM expert_applications WHERE id = $1", appID).Scan(&uid)
		h.DB.Exec("UPDATE expert_applications SET status = 'approved' WHERE id = $1", appID)
		h.DB.Exec("UPDATE users SET role = 'expert' WHERE id = $1", uid)
		h.DB.Exec("INSERT INTO notifications (user_id, type, content, link) VALUES ($1, 'expert', 'Ваша заявка на статус эксперта одобрена!', '/profile/me')", uid)
		h.logActivity(r, adminID, "admin_approve_expert", "user", uid, "")
	} else {
		h.DB.Exec("UPDATE expert_applications SET status = 'rejected' WHERE id = $1", appID)
		var uid int
		h.DB.QueryRow("SELECT user_id FROM expert_applications WHERE id = $1", appID).Scan(&uid)
		h.DB.Exec("INSERT INTO notifications (user_id, type, content, link) VALUES ($1, 'expert', 'Ваша заявка на статус эксперта отклонена', '/profile/me')", uid)
		h.logActivity(r, adminID, "admin_reject_expert", "user", uid, "")
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}

func (h *Handler) HandleAdminDeleteUser(w http.ResponseWriter, r *http.Request) {
	adminID, _ := h.getCurrentUser(r)
	r.ParseForm()
	targetIDStr := r.FormValue("user_id")
	targetID, _ := strconv.Atoi(targetIDStr)
	if targetID < 1 {
		h.respondError(w, http.StatusBadRequest, "Некорректный пользователь")
		return
	}
	var targetRole string
	h.DB.QueryRow("SELECT role FROM users WHERE id = $1", targetID).Scan(&targetRole)
	if targetRole == "admin" {
		h.respondError(w, http.StatusForbidden, "Нельзя удалить администратора")
		return
	}
	h.DB.Exec("DELETE FROM users WHERE id = $1 AND role != 'admin'", targetID)
	h.logActivity(r, adminID, "admin_delete_user", "user", targetID, "")
	h.respondJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}

func (h *Handler) HandleAdminHidePost(w http.ResponseWriter, r *http.Request) {
	adminID, _ := h.getCurrentUser(r)
	r.ParseForm()
	postIDStr := r.FormValue("post_id")
	postID, _ := strconv.Atoi(postIDStr)
	h.DB.Exec("UPDATE posts SET is_hidden = TRUE WHERE id = $1", postID)
	h.logActivity(r, adminID, "admin_hide_post", "post", postID, "")
	h.respondJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}

func (h *Handler) HandleAdminDeletePost(w http.ResponseWriter, r *http.Request) {
	adminID, _ := h.getCurrentUser(r)
	r.ParseForm()
	postIDStr := r.FormValue("post_id")
	postID, _ := strconv.Atoi(postIDStr)

	h.DB.Exec("DELETE FROM posts WHERE id = $1", postID)
	h.logActivity(r, adminID, "admin_delete_post", "post", postID, "")
	h.respondJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}

func (h *Handler) HandleHomePage(w http.ResponseWriter, r *http.Request) {
	userID, role := h.getCurrentUser(r)
	if userID > 0 {
		http.Redirect(w, r, "/feed", http.StatusSeeOther)
		return
	}

	data := map[string]interface{}{
		"UserID":   userID,
		"UserRole": role,
	}
	h.renderTemplate(w, "home.html", data)
}

func (h *Handler) HandleNotificationsPage(w http.ResponseWriter, r *http.Request) {
	userID, role := h.getCurrentUser(r)

	// Mark all as read when viewing
	h.DB.Exec("UPDATE notifications SET is_read = TRUE WHERE user_id = $1", userID)

	rows, _ := h.DB.Query(
		"SELECT id, user_id, type, content, link, is_read, created_at FROM notifications WHERE user_id = $1 ORDER BY created_at DESC LIMIT 100",
		userID,
	)

	var notifications []models.Notification
	if rows != nil {
		for rows.Next() {
			var n models.Notification
			rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Content, &n.Link, &n.IsRead, &n.CreatedAt)
			notifications = append(notifications, n)
		}
		rows.Close()
	}

	var nickname string
	h.DB.QueryRow("SELECT nickname FROM users WHERE id = $1", userID).Scan(&nickname)

	ban := h.getActiveBan(userID)

	data := map[string]interface{}{
		"Notifications": notifications,
		"UserID":        userID,
		"UserRole":      role,
		"Nickname":      nickname,
		"ActiveBan":     ban,
	}
	h.renderTemplate(w, "notifications.html", data)
}
