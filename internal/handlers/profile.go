package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"startup-platform/internal/models"
)

func (h *Handler) HandleProfilePage(w http.ResponseWriter, r *http.Request) {
	userID, role := h.getCurrentUser(r)
	profileIDStr := strings.TrimPrefix(r.URL.Path, "/profile/")
	var profileID int
	var err error

	if profileIDStr == "" || profileIDStr == "me" {
		profileID = userID
	} else {
		profileID, err = strconv.Atoi(profileIDStr)
		if err != nil {
			http.NotFound(w, r)
			return
		}
	}

	var blockedByThem bool
	if userID > 0 && userID != profileID {
		h.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM blocked_users WHERE user_id = $1 AND blocked_id = $2)", profileID, userID).Scan(&blockedByThem)
	}

	if blockedByThem {
		var nickname string
		h.DB.QueryRow("SELECT nickname FROM users WHERE id = $1", userID).Scan(&nickname)
		ban := h.getActiveBan(userID)
		data := map[string]interface{}{
			"Blocked":   true,
			"UserID":    userID,
			"UserRole":  role,
			"Nickname":  nickname,
			"ActiveBan": ban,
		}
		h.renderTemplate(w, "profile.html", data)
		return
	}

	var u models.User
	err = h.DB.QueryRow(
		`SELECT id, email, phone, nickname, name, city, bio, role, interests, user_role2, avatar_url, is_blocked, created_at
		FROM users WHERE id = $1`,
		profileID,
	).Scan(&u.ID, &u.Email, &u.Phone, &u.Nickname, &u.Name, &u.City, &u.Bio,
		&u.Role, &u.Interests, &u.UserRole2, &u.AvatarURL, &u.IsBlocked, &u.CreatedAt)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	postRows, _ := h.DB.Query(
		`SELECT p.id, p.title, p.description, p.category, p.view_count, p.created_at,
		COALESCE(AVG(r.score), 0), COUNT(DISTINCT r.id)
		FROM posts p LEFT JOIN ratings r ON r.post_id = p.id
		WHERE p.author_id = $1 AND p.is_hidden = FALSE
		GROUP BY p.id ORDER BY p.created_at DESC`,
		profileID,
	)

	var userPosts []models.Post
	if postRows != nil {
		for postRows.Next() {
			var p models.Post
			postRows.Scan(&p.ID, &p.Title, &p.Description, &p.Category, &p.ViewCount,
				&p.CreatedAt, &p.AvgRating, &p.RatingCount)

			fRows, _ := h.DB.Query("SELECT id, post_id, filename, file_path, file_type, file_size FROM files WHERE post_id = $1", p.ID)
			if fRows != nil {
				for fRows.Next() {
					var f models.File
					fRows.Scan(&f.ID, &f.PostID, &f.Filename, &f.FilePath, &f.FileType, &f.FileSize)
					p.Files = append(p.Files, f)
				}
				fRows.Close()
			}

			userPosts = append(userPosts, p)
		}
		postRows.Close()
	}

	var isFollowing bool
	if userID > 0 && userID != profileID {
		h.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM follows WHERE follower_id = $1 AND following_id = $2)", userID, profileID).Scan(&isFollowing)
	}

	var followersCount, followingCount int
	h.DB.QueryRow("SELECT COUNT(*) FROM follows WHERE following_id = $1", profileID).Scan(&followersCount)
	h.DB.QueryRow("SELECT COUNT(*) FROM follows WHERE follower_id = $1", profileID).Scan(&followingCount)

	var isBlockedByMe bool
	if userID > 0 && userID != profileID {
		h.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM blocked_users WHERE user_id = $1 AND blocked_id = $2)", userID, profileID).Scan(&isBlockedByMe)
	}

	projectRows, _ := h.DB.Query(
		`SELECT p.id, p.title, p.description, p.created_at,
		(SELECT COUNT(*) FROM project_posts pp WHERE pp.project_id = p.id) as post_count
		FROM projects p WHERE p.user_id = $1 ORDER BY p.created_at DESC`,
		profileID,
	)
	var projects []models.Project
	if projectRows != nil {
		for projectRows.Next() {
			var pr models.Project
			projectRows.Scan(&pr.ID, &pr.Title, &pr.Description, &pr.CreatedAt, &pr.PostCount)
			projects = append(projects, pr)
		}
		projectRows.Close()
	}

	var nickname string
	h.DB.QueryRow("SELECT nickname FROM users WHERE id = $1", userID).Scan(&nickname)

	canViewStats := role == "premium" || role == "expert" || role == "admin"
	if !canViewStats {
		var hasActiveSub bool
		h.DB.QueryRow(
			"SELECT EXISTS(SELECT 1 FROM subscriptions WHERE user_id = $1 AND status = 'active' AND expires_at > NOW())",
			userID,
		).Scan(&hasActiveSub)
		canViewStats = hasActiveSub
	}

	ban := h.getActiveBan(userID)

	data := map[string]interface{}{
		"Profile":        u,
		"Posts":          userPosts,
		"Projects":       projects,
		"PostCount":      len(userPosts),
		"UserID":         userID,
		"UserRole":       role,
		"Nickname":       nickname,
		"IsOwner":        userID == profileID,
		"IsFollowing":    isFollowing,
		"FollowersCount": followersCount,
		"FollowingCount": followingCount,
		"IsBlockedByMe":  isBlockedByMe,
		"CanViewStats":   canViewStats,
		"ActiveBan":      ban,
	}
	h.renderTemplate(w, "profile.html", data)
}

func (h *Handler) HandleSettingsPage(w http.ResponseWriter, r *http.Request) {
	userID, role := h.getCurrentUser(r)

	var u models.User
	h.DB.QueryRow(
		`SELECT id, email, phone, nickname, name, city, bio, role, interests, user_role2, avatar_url
		FROM users WHERE id = $1`, userID,
	).Scan(&u.ID, &u.Email, &u.Phone, &u.Nickname, &u.Name, &u.City, &u.Bio,
		&u.Role, &u.Interests, &u.UserRole2, &u.AvatarURL)

	blockedRows, _ := h.DB.Query(
		`SELECT bu.blocked_id, u.nickname, u.avatar_url, bu.created_at 
		FROM blocked_users bu JOIN users u ON bu.blocked_id = u.id 
		WHERE bu.user_id = $1 ORDER BY bu.created_at DESC`, userID,
	)
	type BlockedInfo struct {
		ID        int    `json:"id"`
		Nickname  string `json:"nickname"`
		AvatarURL string `json:"avatar_url"`
	}
	var blockedUsers []BlockedInfo
	if blockedRows != nil {
		for blockedRows.Next() {
			var bi BlockedInfo
			var ca interface{}
			blockedRows.Scan(&bi.ID, &bi.Nickname, &bi.AvatarURL, &ca)
			blockedUsers = append(blockedUsers, bi)
		}
		blockedRows.Close()
	}

	likedRows, _ := h.DB.Query(
		`SELECT DISTINCT p.id, p.title, p.category, p.created_at, u.nickname
		FROM ratings r JOIN posts p ON r.post_id = p.id JOIN users u ON p.author_id = u.id
		WHERE r.user_id = $1 ORDER BY p.created_at DESC LIMIT 50`, userID,
	)
	type HistoryPost struct {
		ID       int    `json:"id"`
		Title    string `json:"title"`
		Category string `json:"category"`
		Author   string `json:"author"`
	}
	var likedPosts []HistoryPost
	if likedRows != nil {
		for likedRows.Next() {
			var hp HistoryPost
			var ca interface{}
			likedRows.Scan(&hp.ID, &hp.Title, &hp.Category, &ca, &hp.Author)
			likedPosts = append(likedPosts, hp)
		}
		likedRows.Close()
	}

	commentedRows, _ := h.DB.Query(
		`SELECT DISTINCT p.id, p.title, p.category, p.created_at, u.nickname
		FROM comments c JOIN posts p ON c.post_id = p.id JOIN users u ON p.author_id = u.id
		WHERE c.author_id = $1 ORDER BY p.created_at DESC LIMIT 50`, userID,
	)
	var commentedPosts []HistoryPost
	if commentedRows != nil {
		for commentedRows.Next() {
			var hp HistoryPost
			var ca interface{}
			commentedRows.Scan(&hp.ID, &hp.Title, &hp.Category, &ca, &hp.Author)
			commentedPosts = append(commentedPosts, hp)
		}
		commentedRows.Close()
	}

	ban := h.getActiveBan(userID)

	canViewStats := role == "premium" || role == "expert" || role == "admin"
	if !canViewStats {
		var hasActiveSub bool
		h.DB.QueryRow(
			"SELECT EXISTS(SELECT 1 FROM subscriptions WHERE user_id = $1 AND status = 'active' AND expires_at > NOW())",
			userID,
		).Scan(&hasActiveSub)
		canViewStats = hasActiveSub
	}

	data := map[string]interface{}{
		"Profile":        u,
		"BlockedUsers":   blockedUsers,
		"LikedPosts":     likedPosts,
		"CommentedPosts": commentedPosts,
		"CanViewStats":   canViewStats,
		"UserID":         userID,
		"UserRole":       role,
		"Nickname":       u.Nickname,
		"ActiveBan":      ban,
	}
	h.renderTemplate(w, "settings.html", data)
}

func (h *Handler) HandleEditProfile(w http.ResponseWriter, r *http.Request) {
	userID, _ := h.getCurrentUser(r)
	r.ParseForm()

	name := strings.TrimSpace(r.FormValue("name"))
	city := strings.TrimSpace(r.FormValue("city"))
	bio := strings.TrimSpace(r.FormValue("bio"))
	interests := strings.TrimSpace(r.FormValue("interests"))
	userRole2 := strings.TrimSpace(r.FormValue("user_role2"))
	phone := strings.TrimSpace(r.FormValue("phone"))

	h.DB.Exec(
		"UPDATE users SET name=$1, city=$2, bio=$3, interests=$4, user_role2=$5, phone=$6, updated_at=NOW() WHERE id=$7",
		name, city, bio, interests, userRole2, phone, userID,
	)

	h.logActivity(r, userID, "edit_profile", "user", userID, "")

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}

func (h *Handler) HandleAvatarUpload(w http.ResponseWriter, r *http.Request) {
	userID, _ := h.getCurrentUser(r)

	r.ParseMultipartForm(5 << 20) // 5MB max
	file, header, err := r.FormFile("avatar")
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Файл не найден")
		return
	}
	defer file.Close()

	ct := header.Header.Get("Content-Type")
	if ct != "image/jpeg" && ct != "image/png" && ct != "image/webp" && ct != "image/gif" {
		h.respondError(w, http.StatusBadRequest, "Допустимы только изображения (JPEG, PNG, WebP, GIF)")
		return
	}

	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	filename := fmt.Sprintf("avatar_%d%s", userID, ext)
	savePath := filepath.Join(h.UploadDir, filename)

	out, err := os.Create(savePath)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Ошибка сохранения файла")
		return
	}
	defer out.Close()
	io.Copy(out, file)

	avatarURL := "/uploads/" + filename
	h.DB.Exec("UPDATE users SET avatar_url=$1, updated_at=NOW() WHERE id=$2", avatarURL, userID)

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}

func (h *Handler) HandleFollow(w http.ResponseWriter, r *http.Request) {
	userID, _ := h.getCurrentUser(r)
	r.ParseForm()

	targetIDStr := r.FormValue("user_id")
	targetID, err := strconv.Atoi(targetIDStr)
	if err != nil || targetID == userID {
		h.respondError(w, http.StatusBadRequest, "Некорректный запрос")
		return
	}

	h.DB.Exec("INSERT INTO follows (follower_id, following_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", userID, targetID)

	var nickname string
	h.DB.QueryRow("SELECT nickname FROM users WHERE id = $1", userID).Scan(&nickname)
	h.DB.Exec("INSERT INTO notifications (user_id, type, content, link) VALUES ($1, 'follow', $2, $3)",
		targetID, nickname+" подписался на вас", "/profile/"+strconv.Itoa(userID))

	h.logActivity(r, userID, "follow", "user", targetID, "")

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}

func (h *Handler) HandleUnfollow(w http.ResponseWriter, r *http.Request) {
	userID, _ := h.getCurrentUser(r)
	r.ParseForm()

	targetIDStr := r.FormValue("user_id")
	targetID, _ := strconv.Atoi(targetIDStr)

	h.DB.Exec("DELETE FROM follows WHERE follower_id = $1 AND following_id = $2", userID, targetID)
	h.logActivity(r, userID, "unfollow", "user", targetID, "")
	h.respondJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}

func (h *Handler) HandleFollowersList(w http.ResponseWriter, r *http.Request) {
	profileIDStr := r.URL.Query().Get("user_id")
	profileID, _ := strconv.Atoi(profileIDStr)

	rows, _ := h.DB.Query(
		`SELECT u.id, u.nickname, u.avatar_url, u.bio FROM follows f 
		JOIN users u ON f.follower_id = u.id WHERE f.following_id = $1 ORDER BY f.created_at DESC`,
		profileID,
	)

	type FollowUser struct {
		ID        int    `json:"id"`
		Nickname  string `json:"nickname"`
		AvatarURL string `json:"avatar_url"`
		Bio       string `json:"bio"`
	}
	var users []FollowUser
	if rows != nil {
		for rows.Next() {
			var fu FollowUser
			rows.Scan(&fu.ID, &fu.Nickname, &fu.AvatarURL, &fu.Bio)
			users = append(users, fu)
		}
		rows.Close()
	}
	if users == nil {
		users = []FollowUser{}
	}

	h.respondJSON(w, http.StatusOK, users)
}

func (h *Handler) HandleFollowingList(w http.ResponseWriter, r *http.Request) {
	profileIDStr := r.URL.Query().Get("user_id")
	profileID, _ := strconv.Atoi(profileIDStr)

	rows, _ := h.DB.Query(
		`SELECT u.id, u.nickname, u.avatar_url, u.bio FROM follows f 
		JOIN users u ON f.following_id = u.id WHERE f.follower_id = $1 ORDER BY f.created_at DESC`,
		profileID,
	)

	type FollowUser struct {
		ID        int    `json:"id"`
		Nickname  string `json:"nickname"`
		AvatarURL string `json:"avatar_url"`
		Bio       string `json:"bio"`
	}
	var users []FollowUser
	if rows != nil {
		for rows.Next() {
			var fu FollowUser
			rows.Scan(&fu.ID, &fu.Nickname, &fu.AvatarURL, &fu.Bio)
			users = append(users, fu)
		}
		rows.Close()
	}
	if users == nil {
		users = []FollowUser{}
	}

	h.respondJSON(w, http.StatusOK, users)
}

func (h *Handler) HandleAddFriend(w http.ResponseWriter, r *http.Request) {
	userID, _ := h.getCurrentUser(r)
	r.ParseForm()

	friendIDStr := r.FormValue("friend_id")
	friendID, err := strconv.Atoi(friendIDStr)
	if err != nil || friendID == userID {
		h.respondError(w, http.StatusBadRequest, "Некорректный запрос")
		return
	}

	var exists bool
	h.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM friendships WHERE user_id = $1 AND friend_id = $2)", userID, friendID).Scan(&exists)
	if exists {
		h.respondError(w, http.StatusConflict, "Запрос уже отправлен")
		return
	}

	var incoming string
	h.DB.QueryRow("SELECT status FROM friendships WHERE user_id = $1 AND friend_id = $2", friendID, userID).Scan(&incoming)
	if incoming == "pending" {
		h.DB.Exec("UPDATE friendships SET status = 'accepted' WHERE user_id = $1 AND friend_id = $2", friendID, userID)
		h.DB.Exec("INSERT INTO friendships (user_id, friend_id, status) VALUES ($1, $2, 'accepted') ON CONFLICT DO NOTHING", userID, friendID)

		var nickname string
		h.DB.QueryRow("SELECT nickname FROM users WHERE id = $1", userID).Scan(&nickname)
		h.DB.Exec("INSERT INTO notifications (user_id, type, content, link) VALUES ($1, 'friend', $2, $3)",
			friendID, nickname+" принял вашу заявку в друзья", "/profile/"+strconv.Itoa(userID))
	} else {
		h.DB.Exec("INSERT INTO friendships (user_id, friend_id, status) VALUES ($1, $2, 'pending')", userID, friendID)

		var nickname string
		h.DB.QueryRow("SELECT nickname FROM users WHERE id = $1", userID).Scan(&nickname)
		h.DB.Exec("INSERT INTO notifications (user_id, type, content, link) VALUES ($1, 'friend_request', $2, $3)",
			friendID, nickname+" хочет добавить вас в друзья", "/profile/"+strconv.Itoa(userID))
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}

func (h *Handler) HandleRemoveFriend(w http.ResponseWriter, r *http.Request) {
	userID, _ := h.getCurrentUser(r)
	r.ParseForm()

	friendIDStr := r.FormValue("friend_id")
	friendID, _ := strconv.Atoi(friendIDStr)

	h.DB.Exec("DELETE FROM friendships WHERE (user_id = $1 AND friend_id = $2) OR (user_id = $2 AND friend_id = $1)", userID, friendID)
	h.respondJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}

func (h *Handler) HandleBlockUser(w http.ResponseWriter, r *http.Request) {
	userID, _ := h.getCurrentUser(r)
	r.ParseForm()

	blockedIDStr := r.FormValue("user_id")
	if blockedIDStr == "" {
		blockedIDStr = r.FormValue("blocked_id")
	}
	blockedID, _ := strconv.Atoi(blockedIDStr)

	if blockedID == userID {
		h.respondError(w, http.StatusBadRequest, "Нельзя заблокировать себя")
		return
	}

	h.DB.Exec("INSERT INTO blocked_users (user_id, blocked_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", userID, blockedID)
	h.DB.Exec("DELETE FROM follows WHERE (follower_id = $1 AND following_id = $2) OR (follower_id = $2 AND following_id = $1)", userID, blockedID)
	h.DB.Exec("DELETE FROM friendships WHERE (user_id = $1 AND friend_id = $2) OR (user_id = $2 AND friend_id = $1)", userID, blockedID)
	h.logActivity(r, userID, "block_user", "user", blockedID, "")
	h.respondJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}

func (h *Handler) HandleUnblockUser(w http.ResponseWriter, r *http.Request) {
	userID, _ := h.getCurrentUser(r)
	r.ParseForm()

	blockedIDStr := r.FormValue("user_id")
	if blockedIDStr == "" {
		blockedIDStr = r.FormValue("blocked_id")
	}
	blockedID, _ := strconv.Atoi(blockedIDStr)

	h.DB.Exec("DELETE FROM blocked_users WHERE user_id = $1 AND blocked_id = $2", userID, blockedID)
	h.logActivity(r, userID, "unblock_user", "user", blockedID, "")
	h.respondJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}

func (h *Handler) HandleRepost(w http.ResponseWriter, r *http.Request) {
	userID, _ := h.getCurrentUser(r)
	r.ParseForm()

	postIDStr := r.FormValue("post_id")
	postID, _ := strconv.Atoi(postIDStr)
	comment := strings.TrimSpace(r.FormValue("comment"))

	h.DB.Exec("INSERT INTO reposts (post_id, user_id, comment) VALUES ($1, $2, $3)", postID, userID, comment)

	var postAuthorID int
	h.DB.QueryRow("SELECT author_id FROM posts WHERE id = $1", postID).Scan(&postAuthorID)
	if postAuthorID != userID {
		var nickname string
		h.DB.QueryRow("SELECT nickname FROM users WHERE id = $1", userID).Scan(&nickname)
		h.DB.Exec("INSERT INTO notifications (user_id, type, content, link) VALUES ($1, 'repost', $2, $3)",
			postAuthorID, nickname+" сделал репост вашего проекта", "/post/"+postIDStr)
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}

func (h *Handler) HandleComplaint(w http.ResponseWriter, r *http.Request) {
	userID, _ := h.getCurrentUser(r)
	r.ParseForm()

	targetType := r.FormValue("target_type")
	targetIDStr := r.FormValue("target_id")
	targetID, _ := strconv.Atoi(targetIDStr)
	category := r.FormValue("category")
	description := strings.TrimSpace(r.FormValue("description"))

	if description == "" {
		h.respondError(w, http.StatusBadRequest, "Описание жалобы обязательно")
		return
	}

	h.DB.Exec(
		"INSERT INTO complaints (author_id, target_type, target_id, category, description) VALUES ($1, $2, $3, $4, $5)",
		userID, targetType, targetID, category, description,
	)

	h.logActivity(r, userID, "complaint", targetType, targetID, category)

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}

func (h *Handler) HandleTeamRequestCreate(w http.ResponseWriter, r *http.Request) {
	userID, _ := h.getCurrentUser(r)
	r.ParseForm()

	postIDStr := r.FormValue("post_id")
	postID, _ := strconv.Atoi(postIDStr)
	title := strings.TrimSpace(r.FormValue("title"))
	description := strings.TrimSpace(r.FormValue("description"))
	skills := strings.TrimSpace(r.FormValue("skills"))
	roleNeeded := strings.TrimSpace(r.FormValue("role_needed"))

	if title == "" {
		h.respondError(w, http.StatusBadRequest, "Название запроса обязательно")
		return
	}

	h.DB.Exec(
		"INSERT INTO team_requests (post_id, author_id, title, description, skills, role_needed) VALUES ($1, $2, $3, $4, $5, $6)",
		postID, userID, title, description, skills, roleNeeded,
	)

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}

func (h *Handler) HandleTeamRespond(w http.ResponseWriter, r *http.Request) {
	userID, _ := h.getCurrentUser(r)
	r.ParseForm()

	requestIDStr := r.FormValue("request_id")
	requestID, _ := strconv.Atoi(requestIDStr)
	message := strings.TrimSpace(r.FormValue("message"))

	h.DB.Exec(
		"INSERT INTO team_responses (request_id, user_id, message) VALUES ($1, $2, $3)",
		requestID, userID, message,
	)

	var authorID int
	h.DB.QueryRow("SELECT author_id FROM team_requests WHERE id = $1", requestID).Scan(&authorID)
	var nickname string
	h.DB.QueryRow("SELECT nickname FROM users WHERE id = $1", userID).Scan(&nickname)
	h.DB.Exec("INSERT INTO notifications (user_id, type, content, link) VALUES ($1, 'team_response', $2, $3)",
		authorID, nickname+" откликнулся на ваш запрос команды", "/profile/"+strconv.Itoa(userID))

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}

func (h *Handler) HandleTeamSearchPage(w http.ResponseWriter, r *http.Request) {
	userID, role := h.getCurrentUser(r)

	rows, _ := h.DB.Query(
		`SELECT tr.id, tr.post_id, tr.author_id, tr.title, tr.description, tr.skills, tr.role_needed, tr.is_open,
		p.title as post_title
		FROM team_requests tr JOIN posts p ON tr.post_id = p.id
		WHERE tr.is_open = TRUE ORDER BY tr.created_at DESC`,
	)

	var requests []models.TeamRequest
	if rows != nil {
		for rows.Next() {
			var tr models.TeamRequest
			rows.Scan(&tr.ID, &tr.PostID, &tr.AuthorID, &tr.Title, &tr.Description,
				&tr.Skills, &tr.RoleNeeded, &tr.IsOpen, &tr.PostTitle)
			requests = append(requests, tr)
		}
		rows.Close()
	}

	var nickname string
	h.DB.QueryRow("SELECT nickname FROM users WHERE id = $1", userID).Scan(&nickname)

	ban := h.getActiveBan(userID)

	data := map[string]interface{}{
		"TeamRequests": requests,
		"UserID":       userID,
		"UserRole":     role,
		"Nickname":     nickname,
		"ActiveBan":    ban,
	}
	h.renderTemplate(w, "teams.html", data)
}

func (h *Handler) HandleExpertApplyPage(w http.ResponseWriter, r *http.Request) {
	userID, role := h.getCurrentUser(r)
	var nickname string
	h.DB.QueryRow("SELECT nickname FROM users WHERE id = $1", userID).Scan(&nickname)

	var appStatus string
	err := h.DB.QueryRow("SELECT status FROM expert_applications WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1", userID).Scan(&appStatus)
	if err != nil {
		appStatus = ""
	}

	data := map[string]interface{}{
		"UserID":    userID,
		"UserRole":  role,
		"Nickname":  nickname,
		"AppStatus": appStatus,
		"IsExpert":  role == "expert",
	}
	h.renderTemplate(w, "expert_apply.html", data)
}

func (h *Handler) HandleExpertApply(w http.ResponseWriter, r *http.Request) {
	userID, _ := h.getCurrentUser(r)
	r.ParseForm()

	portfolio := strings.TrimSpace(r.FormValue("portfolio"))
	description := strings.TrimSpace(r.FormValue("description"))

	if portfolio == "" || description == "" {
		h.respondError(w, http.StatusBadRequest, "Портфолио и описание обязательны")
		return
	}

	var exists bool
	h.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM expert_applications WHERE user_id = $1 AND status = 'pending')", userID).Scan(&exists)
	if exists {
		h.respondError(w, http.StatusConflict, "Заявка уже подана")
		return
	}

	h.DB.Exec("INSERT INTO expert_applications (user_id, portfolio, description) VALUES ($1, $2, $3)",
		userID, portfolio, description)

	h.logActivity(r, userID, "expert_apply", "user", userID, "")

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}

func (h *Handler) HandlePricingPage(w http.ResponseWriter, r *http.Request) {
	userID, role := h.getCurrentUser(r)
	var nickname string
	if userID > 0 {
		h.DB.QueryRow("SELECT nickname FROM users WHERE id = $1", userID).Scan(&nickname)
	}
	data := map[string]interface{}{
		"UserID":    userID,
		"UserRole":  role,
		"Nickname":  nickname,
		"IsPremium": role == "premium" || role == "admin",
	}
	h.renderTemplate(w, "pricing.html", data)
}

func (h *Handler) HandleSubscribe(w http.ResponseWriter, r *http.Request) {
	userID, _ := h.getCurrentUser(r)
	r.ParseForm()

	plan := r.FormValue("plan")
	if plan != "monthly" && plan != "yearly" {
		h.respondError(w, http.StatusBadRequest, "Некорректный план")
		return
	}

	var days int
	if plan == "monthly" {
		days = 30
	} else {
		days = 365
	}

	h.DB.Exec(
		"INSERT INTO subscriptions (user_id, plan, status, expires_at) VALUES ($1, $2, 'active', NOW() + INTERVAL '1 day' * $3)",
		userID, plan, days,
	)
	h.DB.Exec("UPDATE users SET role = 'premium' WHERE id = $1 AND role = 'user'", userID)

	h.logActivity(r, userID, "subscribe", "subscription", 0, plan)

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}

func (h *Handler) HandleCreateProject(w http.ResponseWriter, r *http.Request) {
	userID, _ := h.getCurrentUser(r)
	r.ParseForm()

	title := strings.TrimSpace(r.FormValue("title"))
	description := strings.TrimSpace(r.FormValue("description"))

	if title == "" {
		h.respondError(w, http.StatusBadRequest, "Название проекта обязательно")
		return
	}

	var projectID int
	h.DB.QueryRow(
		"INSERT INTO projects (user_id, title, description) VALUES ($1, $2, $3) RETURNING id",
		userID, title, description,
	).Scan(&projectID)

	h.logActivity(r, userID, "create_project", "project", projectID, title)

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"success": true, "project_id": projectID})
}

func (h *Handler) HandleDeleteProject(w http.ResponseWriter, r *http.Request) {
	userID, _ := h.getCurrentUser(r)
	r.ParseForm()

	projectIDStr := r.FormValue("project_id")
	projectID, _ := strconv.Atoi(projectIDStr)

	var ownerID int
	h.DB.QueryRow("SELECT user_id FROM projects WHERE id = $1", projectID).Scan(&ownerID)
	if ownerID != userID {
		h.respondError(w, http.StatusForbidden, "Нет доступа")
		return
	}

	h.DB.Exec("DELETE FROM projects WHERE id = $1", projectID)
	h.respondJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}

func (h *Handler) HandleAddPostToProject(w http.ResponseWriter, r *http.Request) {
	userID, _ := h.getCurrentUser(r)
	r.ParseForm()

	projectIDStr := r.FormValue("project_id")
	projectID, _ := strconv.Atoi(projectIDStr)
	postIDStr := r.FormValue("post_id")
	postID, _ := strconv.Atoi(postIDStr)

	var ownerID int
	h.DB.QueryRow("SELECT user_id FROM projects WHERE id = $1", projectID).Scan(&ownerID)
	if ownerID != userID {
		h.respondError(w, http.StatusForbidden, "Нет доступа")
		return
	}

	h.DB.Exec("INSERT INTO project_posts (project_id, post_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", projectID, postID)
	h.respondJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}

func (h *Handler) HandleRemovePostFromProject(w http.ResponseWriter, r *http.Request) {
	userID, _ := h.getCurrentUser(r)
	r.ParseForm()

	projectIDStr := r.FormValue("project_id")
	projectID, _ := strconv.Atoi(projectIDStr)
	postIDStr := r.FormValue("post_id")
	postID, _ := strconv.Atoi(postIDStr)

	var ownerID int
	h.DB.QueryRow("SELECT user_id FROM projects WHERE id = $1", projectID).Scan(&ownerID)
	if ownerID != userID {
		h.respondError(w, http.StatusForbidden, "Нет доступа")
		return
	}

	h.DB.Exec("DELETE FROM project_posts WHERE project_id = $1 AND post_id = $2", projectID, postID)
	h.respondJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}

func (h *Handler) HandleProjectPage(w http.ResponseWriter, r *http.Request) {
	userID, role := h.getCurrentUser(r)
	projectIDStr := strings.TrimPrefix(r.URL.Path, "/project/")
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	var pr models.Project
	var ownerNickname string
	err = h.DB.QueryRow(
		`SELECT p.id, p.user_id, p.title, p.description, p.created_at, u.nickname
		FROM projects p JOIN users u ON p.user_id = u.id WHERE p.id = $1`,
		projectID,
	).Scan(&pr.ID, &pr.UserID, &pr.Title, &pr.Description, &pr.CreatedAt, &ownerNickname)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	postRows, _ := h.DB.Query(
		`SELECT p.id, p.title, p.description, p.category, p.view_count, p.created_at,
		u.nickname, u.avatar_url,
		COALESCE(AVG(r.score), 0), COUNT(DISTINCT r.id)
		FROM project_posts pp
		JOIN posts p ON pp.post_id = p.id
		JOIN users u ON p.author_id = u.id
		LEFT JOIN ratings r ON r.post_id = p.id
		WHERE pp.project_id = $1 AND p.is_hidden = FALSE
		GROUP BY p.id, u.nickname, u.avatar_url
		ORDER BY pp.sort_order, pp.added_at DESC`,
		projectID,
	)

	var posts []models.Post
	if postRows != nil {
		for postRows.Next() {
			var p models.Post
			postRows.Scan(&p.ID, &p.Title, &p.Description, &p.Category, &p.ViewCount,
				&p.CreatedAt, &p.AuthorNickname, &p.AuthorAvatarURL, &p.AvgRating, &p.RatingCount)

			fRows, _ := h.DB.Query("SELECT id, post_id, filename, file_path, file_type, file_size FROM files WHERE post_id = $1", p.ID)
			if fRows != nil {
				for fRows.Next() {
					var f models.File
					fRows.Scan(&f.ID, &f.PostID, &f.Filename, &f.FilePath, &f.FileType, &f.FileSize)
					p.Files = append(p.Files, f)
				}
				fRows.Close()
			}

			posts = append(posts, p)
		}
		postRows.Close()
	}

	var nickname string
	h.DB.QueryRow("SELECT nickname FROM users WHERE id = $1", userID).Scan(&nickname)

	data := map[string]interface{}{
		"Project":       pr,
		"Posts":         posts,
		"OwnerNickname": ownerNickname,
		"UserID":        userID,
		"UserRole":      role,
		"Nickname":      nickname,
		"IsOwner":       userID == pr.UserID,
	}
	h.renderTemplate(w, "project.html", data)
}

func (h *Handler) HandlePremiumStats(w http.ResponseWriter, r *http.Request) {
	userID, role := h.getCurrentUser(r)

	canViewStats := role == "premium" || role == "admin" || role == "expert"
	if !canViewStats {
		var hasActiveSub bool
		h.DB.QueryRow(
			"SELECT EXISTS(SELECT 1 FROM subscriptions WHERE user_id = $1 AND status = 'active' AND expires_at > NOW())",
			userID,
		).Scan(&hasActiveSub)
		canViewStats = hasActiveSub
	}

	if !canViewStats {
		http.Redirect(w, r, "/pricing", http.StatusSeeOther)
		return
	}

	var totalViews int
	h.DB.QueryRow("SELECT COALESCE(SUM(view_count), 0) FROM posts WHERE author_id = $1", userID).Scan(&totalViews)

	var totalRatings int
	h.DB.QueryRow("SELECT COUNT(*) FROM ratings r JOIN posts p ON r.post_id = p.id WHERE p.author_id = $1", userID).Scan(&totalRatings)

	var avgScore float64
	h.DB.QueryRow("SELECT COALESCE(AVG(r.score), 0) FROM ratings r JOIN posts p ON r.post_id = p.id WHERE p.author_id = $1", userID).Scan(&avgScore)

	var followersCount int
	h.DB.QueryRow("SELECT COUNT(*) FROM follows WHERE following_id = $1", userID).Scan(&followersCount)

	var postsCount int
	h.DB.QueryRow("SELECT COUNT(*) FROM posts WHERE author_id = $1 AND is_hidden = FALSE", userID).Scan(&postsCount)

	var commentsReceived int
	h.DB.QueryRow("SELECT COUNT(*) FROM comments c JOIN posts p ON c.post_id = p.id WHERE p.author_id = $1", userID).Scan(&commentsReceived)

	categoryRows, _ := h.DB.Query(
		`SELECT p.category, COUNT(*) as cnt FROM posts p 
		JOIN ratings r ON r.post_id = p.id 
		WHERE p.author_id = $1 AND p.category != '' 
		GROUP BY p.category ORDER BY cnt DESC LIMIT 5`, userID,
	)
	type CatStat struct {
		Category string
		Count    int
	}
	var topCategories []CatStat
	if categoryRows != nil {
		for categoryRows.Next() {
			var cs CatStat
			categoryRows.Scan(&cs.Category, &cs.Count)
			topCategories = append(topCategories, cs)
		}
		categoryRows.Close()
	}

	var newFollowersMonth int
	h.DB.QueryRow("SELECT COUNT(*) FROM follows WHERE following_id = $1 AND created_at >= NOW() - INTERVAL '30 days'", userID).Scan(&newFollowersMonth)

	var viewsMonth int
	h.DB.QueryRow(`SELECT COALESCE(SUM(view_count), 0) FROM posts WHERE author_id = $1 AND created_at >= NOW() - INTERVAL '30 days'`, userID).Scan(&viewsMonth)

	var nickname string
	h.DB.QueryRow("SELECT nickname FROM users WHERE id = $1", userID).Scan(&nickname)

	data := map[string]interface{}{
		"TotalViews":       totalViews,
		"TotalRatings":     totalRatings,
		"AvgScore":         fmt.Sprintf("%.1f", avgScore),
		"FollowersCount":   followersCount,
		"PostsCount":       postsCount,
		"CommentsReceived": commentsReceived,
		"TopCategories":    topCategories,
		"NewFollowersMonth": newFollowersMonth,
		"ViewsMonth":       viewsMonth,
		"UserID":           userID,
		"UserRole":         role,
		"Nickname":         nickname,
	}
	h.renderTemplate(w, "premium_stats.html", data)
}

func (h *Handler) HandleExpertConsultation(w http.ResponseWriter, r *http.Request) {
	userID, role := h.getCurrentUser(r)
	if role != "expert" {
		h.respondError(w, http.StatusForbidden, "Только для экспертов")
		return
	}
	r.ParseForm()

	postIDStr := r.FormValue("post_id")
	postID, _ := strconv.Atoi(postIDStr)
	title := strings.TrimSpace(r.FormValue("title"))
	description := strings.TrimSpace(r.FormValue("description"))
	priceStr := r.FormValue("price")
	price, _ := strconv.Atoi(priceStr)

	var consultID int
	h.DB.QueryRow(
		"INSERT INTO expert_consultations (expert_id, client_id, post_id, title, description, price) VALUES ($1, 0, $2, $3, $4, $5) RETURNING id",
		userID, postID, title, description, price,
	).Scan(&consultID)

	var postAuthorID int
	h.DB.QueryRow("SELECT author_id FROM posts WHERE id = $1", postID).Scan(&postAuthorID)
	var nickname string
	h.DB.QueryRow("SELECT nickname FROM users WHERE id = $1", userID).Scan(&nickname)
	h.DB.Exec("INSERT INTO notifications (user_id, type, content, link) VALUES ($1, 'consultation', $2, $3)",
		postAuthorID, nickname+" предлагает экспертную консультацию", "/post/"+postIDStr)

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}

func (h *Handler) HandleConsultationRespond(w http.ResponseWriter, r *http.Request) {
	userID, _ := h.getCurrentUser(r)
	r.ParseForm()

	consultIDStr := r.FormValue("consultation_id")
	consultID, _ := strconv.Atoi(consultIDStr)
	action := r.FormValue("action")

	var expertID, clientID, postID int
	var status string
	err := h.DB.QueryRow("SELECT expert_id, client_id, post_id, status FROM expert_consultations WHERE id = $1", consultID).Scan(&expertID, &clientID, &postID, &status)
	if err != nil {
		h.respondError(w, http.StatusNotFound, "Консультация не найдена")
		return
	}

	var postAuthorID int
	h.DB.QueryRow("SELECT author_id FROM posts WHERE id = $1", postID).Scan(&postAuthorID)
	if userID != postAuthorID {
		h.respondError(w, http.StatusForbidden, "Нет доступа")
		return
	}

	if status != "pending" {
		h.respondError(w, http.StatusBadRequest, "Консультация уже обработана")
		return
	}

	if action == "accept" {
		h.DB.Exec("UPDATE expert_consultations SET status = 'active', client_id = $1 WHERE id = $2", userID, consultID)
		var nick string
		h.DB.QueryRow("SELECT nickname FROM users WHERE id = $1", userID).Scan(&nick)
		h.DB.Exec("INSERT INTO notifications (user_id, type, content, link) VALUES ($1, 'consultation', $2, $3)",
			expertID, nick+" принял предложение консультации", fmt.Sprintf("/consultation/%d", consultID))
	} else {
		h.DB.Exec("UPDATE expert_consultations SET status = 'declined' WHERE id = $1", consultID)
		var nick string
		h.DB.QueryRow("SELECT nickname FROM users WHERE id = $1", userID).Scan(&nick)
		h.DB.Exec("INSERT INTO notifications (user_id, type, content, link) VALUES ($1, 'consultation', $2, $3)",
			expertID, nick+" отклонил предложение консультации", "/profile/me")
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}

func (h *Handler) HandleConsultationPage(w http.ResponseWriter, r *http.Request) {
	userID, role := h.getCurrentUser(r)
	consultIDStr := strings.TrimPrefix(r.URL.Path, "/consultation/")
	consultID, err := strconv.Atoi(consultIDStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	var c models.ExpertConsultation
	var postTitle string
	err = h.DB.QueryRow(
		`SELECT ec.id, ec.expert_id, ec.client_id, ec.post_id, ec.title, ec.description, ec.price, ec.status, ec.created_at,
		eu.nickname, COALESCE(cu.nickname,''), p.title
		FROM expert_consultations ec
		JOIN users eu ON ec.expert_id = eu.id
		LEFT JOIN users cu ON ec.client_id = cu.id
		JOIN posts p ON ec.post_id = p.id
		WHERE ec.id = $1`, consultID,
	).Scan(&c.ID, &c.ExpertID, &c.ClientID, &c.PostID, &c.Title, &c.Description, &c.Price, &c.Status, &c.CreatedAt,
		&c.ExpertNickname, &c.ClientNickname, &postTitle)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	c.PostTitle = postTitle

	if userID != c.ExpertID && userID != c.ClientID {
		http.Redirect(w, r, "/feed", http.StatusSeeOther)
		return
	}

	msgRows, _ := h.DB.Query(
		`SELECT cm.id, cm.consultation_id, cm.sender_id, cm.content, COALESCE(cm.file_path,''), COALESCE(cm.file_name,''), cm.created_at, u.nickname
		FROM consultation_messages cm JOIN users u ON cm.sender_id = u.id
		WHERE cm.consultation_id = $1 ORDER BY cm.created_at ASC`, consultID,
	)
	var messages []models.ConsultationMessage
	if msgRows != nil {
		for msgRows.Next() {
			var m models.ConsultationMessage
			msgRows.Scan(&m.ID, &m.ConsultationID, &m.SenderID, &m.Content, &m.FilePath, &m.FileName, &m.CreatedAt, &m.SenderNickname)
			messages = append(messages, m)
		}
		msgRows.Close()
	}

	var nickname string
	h.DB.QueryRow("SELECT nickname FROM users WHERE id = $1", userID).Scan(&nickname)

	data := map[string]interface{}{
		"Consultation": c,
		"Messages":     messages,
		"UserID":       userID,
		"UserRole":     role,
		"Nickname":     nickname,
	}
	h.renderTemplate(w, "consultation.html", data)
}

func (h *Handler) HandleConsultationSendMessage(w http.ResponseWriter, r *http.Request) {
	userID, _ := h.getCurrentUser(r)

	maxSize := int64(100 << 20)
	r.ParseMultipartForm(maxSize)

	consultIDStr := r.FormValue("consultation_id")
	consultID, _ := strconv.Atoi(consultIDStr)
	content := strings.TrimSpace(r.FormValue("content"))

	var expertID, clientID int
	var status string
	err := h.DB.QueryRow("SELECT expert_id, client_id, status FROM expert_consultations WHERE id = $1", consultID).Scan(&expertID, &clientID, &status)
	if err != nil || status != "active" {
		h.respondError(w, http.StatusBadRequest, "Консультация неактивна")
		return
	}
	if userID != expertID && userID != clientID {
		h.respondError(w, http.StatusForbidden, "Нет доступа")
		return
	}

	var fileSavePath, fileName string
	file, fileHeader, err := r.FormFile("file")
	if err == nil {
		defer file.Close()
		ext := filepath.Ext(fileHeader.Filename)
		newFilename := fmt.Sprintf("consult_%d_%d%s", consultID, time.Now().UnixNano(), ext)
		dst, err := os.Create(filepath.Join(h.UploadDir, newFilename))
		if err == nil {
			defer dst.Close()
			io.Copy(dst, file)
			fileSavePath = "/uploads/" + newFilename
			fileName = fileHeader.Filename
		}
	}

	if content == "" && fileSavePath == "" {
		h.respondError(w, http.StatusBadRequest, "Напишите сообщение или прикрепите файл")
		return
	}

	h.DB.Exec(
		"INSERT INTO consultation_messages (consultation_id, sender_id, content, file_path, file_name) VALUES ($1, $2, $3, $4, $5)",
		consultID, userID, content, fileSavePath, fileName,
	)

	receiverID := expertID
	if userID == expertID {
		receiverID = clientID
	}
	var nick string
	h.DB.QueryRow("SELECT nickname FROM users WHERE id = $1", userID).Scan(&nick)
	h.DB.Exec("INSERT INTO notifications (user_id, type, content, link) VALUES ($1, 'consultation', $2, $3)",
		receiverID, "Новое сообщение от "+nick+" в консультации", fmt.Sprintf("/consultation/%d", consultID))

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}

func (h *Handler) HandleConsultationComplete(w http.ResponseWriter, r *http.Request) {
	userID, _ := h.getCurrentUser(r)
	r.ParseForm()
	consultIDStr := r.FormValue("consultation_id")
	consultID, _ := strconv.Atoi(consultIDStr)

	var expertID int
	h.DB.QueryRow("SELECT expert_id FROM expert_consultations WHERE id = $1", consultID).Scan(&expertID)
	if userID != expertID {
		h.respondError(w, http.StatusForbidden, "Только эксперт может завершить")
		return
	}
	h.DB.Exec("UPDATE expert_consultations SET status = 'completed' WHERE id = $1", consultID)
	h.respondJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}

func (h *Handler) HandleConsultationGetMessages(w http.ResponseWriter, r *http.Request) {
	userID, _ := h.getCurrentUser(r)
	consultIDStr := r.URL.Query().Get("consultation_id")
	consultID, _ := strconv.Atoi(consultIDStr)

	var expertID, clientID int
	h.DB.QueryRow("SELECT expert_id, client_id FROM expert_consultations WHERE id = $1", consultID).Scan(&expertID, &clientID)
	if userID != expertID && userID != clientID {
		h.respondError(w, http.StatusForbidden, "Нет доступа")
		return
	}

	msgRows, _ := h.DB.Query(
		`SELECT cm.id, cm.consultation_id, cm.sender_id, cm.content, COALESCE(cm.file_path,''), COALESCE(cm.file_name,''), cm.created_at, u.nickname
		FROM consultation_messages cm JOIN users u ON cm.sender_id = u.id
		WHERE cm.consultation_id = $1 ORDER BY cm.created_at ASC`, consultID,
	)
	var messages []map[string]interface{}
	if msgRows != nil {
		for msgRows.Next() {
			var id, senderID, cID int
			var cont, fPath, fName, sNick string
			var createdAt time.Time
			msgRows.Scan(&id, &cID, &senderID, &cont, &fPath, &fName, &createdAt, &sNick)
			messages = append(messages, map[string]interface{}{
				"id": id, "sender_id": senderID, "content": cont,
				"file_path": fPath, "file_name": fName,
				"created_at": createdAt.Format("15:04"), "sender_nickname": sNick,
			})
		}
		msgRows.Close()
	}
	if messages == nil {
		messages = []map[string]interface{}{}
	}
	h.respondJSON(w, http.StatusOK, map[string]interface{}{"messages": messages})
}

func (h *Handler) HandleMyConsultations(w http.ResponseWriter, r *http.Request) {
	userID, role := h.getCurrentUser(r)

	rows, _ := h.DB.Query(
		`SELECT ec.id, ec.expert_id, ec.client_id, ec.post_id, ec.title, ec.description, ec.price, ec.status, ec.created_at,
		eu.nickname, COALESCE(cu.nickname,''), p.title
		FROM expert_consultations ec
		JOIN users eu ON ec.expert_id = eu.id
		LEFT JOIN users cu ON ec.client_id = cu.id
		JOIN posts p ON ec.post_id = p.id
		WHERE ec.expert_id = $1 OR ec.client_id = $1
		ORDER BY ec.created_at DESC`, userID,
	)
	var consultations []models.ExpertConsultation
	if rows != nil {
		for rows.Next() {
			var c models.ExpertConsultation
			rows.Scan(&c.ID, &c.ExpertID, &c.ClientID, &c.PostID, &c.Title, &c.Description, &c.Price, &c.Status, &c.CreatedAt,
				&c.ExpertNickname, &c.ClientNickname, &c.PostTitle)
			consultations = append(consultations, c)
		}
		rows.Close()
	}

	var nickname string
	h.DB.QueryRow("SELECT nickname FROM users WHERE id = $1", userID).Scan(&nickname)

	data := map[string]interface{}{
		"Consultations": consultations,
		"UserID":        userID,
		"UserRole":      role,
		"Nickname":      nickname,
	}
	h.renderTemplate(w, "consultations.html", data)
}
