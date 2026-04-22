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

func (h *Handler) HandleFeedPage(w http.ResponseWriter, r *http.Request) {
	userID, role := h.getCurrentUser(r)
	category := r.URL.Query().Get("category")
	search := r.URL.Query().Get("search")
	sort := r.URL.Query().Get("sort")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit := 12
	offset := (page - 1) * limit

	query := `SELECT p.id, p.author_id, p.title, p.description, p.category, p.tags,
		p.is_pinned, p.pinned_until, p.comments_off, p.is_hidden, p.view_count,
		p.is_premium, p.created_at, p.updated_at,
		u.nickname, u.avatar_url, u.role,
		COALESCE(AVG(r.score), 0) as avg_rating,
		COUNT(DISTINCT r.id) as rating_count,
		(SELECT COUNT(*) FROM comments c WHERE c.post_id = p.id) as comment_count
		FROM posts p
		JOIN users u ON p.author_id = u.id
		LEFT JOIN ratings r ON r.post_id = p.id
		WHERE p.is_hidden = FALSE`

	args := []interface{}{}
	argIdx := 1

	if category == "recommended" && userID > 0 {
		query += fmt.Sprintf(` AND p.category IN (SELECT ucv.category FROM user_category_views ucv WHERE ucv.user_id = $%d ORDER BY ucv.view_count DESC LIMIT 5)`, argIdx)
		args = append(args, userID)
		argIdx++
	} else if category != "" && category != "recommended" {
		query += fmt.Sprintf(" AND p.category = $%d", argIdx)
		args = append(args, category)
		argIdx++
	}

	if search != "" {
		query += fmt.Sprintf(" AND (LOWER(p.title) LIKE LOWER($%d) OR LOWER(p.description) LIKE LOWER($%d) OR LOWER(p.tags) LIKE LOWER($%d))", argIdx, argIdx+1, argIdx+2)
		searchTerm := "%" + search + "%"
		args = append(args, searchTerm, searchTerm, searchTerm)
		argIdx += 3
	}

	query += " GROUP BY p.id, u.nickname, u.avatar_url, u.role"

	if sort == "" && userID > 0 && category == "" && search == "" {
		query += fmt.Sprintf(` ORDER BY 
			CASE WHEN p.category IN (SELECT ucv.category FROM user_category_views ucv WHERE ucv.user_id = $%d ORDER BY ucv.view_count DESC LIMIT 3) THEN 0 ELSE 1 END,
			p.is_pinned DESC, p.created_at DESC`, argIdx)
		args = append(args, userID)
		argIdx++
	} else {
		switch sort {
		case "rating":
			query += " ORDER BY avg_rating DESC, p.created_at DESC"
		case "views":
			query += " ORDER BY p.view_count DESC, p.created_at DESC"
		case "old":
			query += " ORDER BY p.created_at ASC"
		default:
			query += " ORDER BY p.is_pinned DESC, p.created_at DESC"
		}
	}

	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := h.DB.Query(query, args...)
	if err != nil {
		h.renderTemplate(w, "feed.html", map[string]interface{}{
			"Error": "Ошибка загрузки ленты",
		})
		return
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var p models.Post
		var pinnedUntil *time.Time
		err := rows.Scan(&p.ID, &p.AuthorID, &p.Title, &p.Description, &p.Category,
			&p.Tags, &p.IsPinned, &pinnedUntil, &p.CommentsOff, &p.IsHidden,
			&p.ViewCount, &p.IsPremium, &p.CreatedAt, &p.UpdatedAt,
			&p.AuthorNickname, &p.AuthorAvatarURL, &p.AuthorRole,
			&p.AvgRating, &p.RatingCount, &p.CommentCount)
		if err != nil {
			continue
		}
		p.PinnedUntil = pinnedUntil

		fileRows, err := h.DB.Query("SELECT id, post_id, filename, file_path, file_type, file_size FROM files WHERE post_id = $1", p.ID)
		if err == nil {
			for fileRows.Next() {
				var f models.File
				fileRows.Scan(&f.ID, &f.PostID, &f.Filename, &f.FilePath, &f.FileType, &f.FileSize)
				p.Files = append(p.Files, f)
			}
			fileRows.Close()
		}

		posts = append(posts, p)
	}

	var totalPosts int
	countQuery := "SELECT COUNT(*) FROM posts WHERE is_hidden = FALSE"
	countArgs := []interface{}{}
	if category != "" {
		countQuery += " AND category = $1"
		countArgs = append(countArgs, category)
	}
	h.DB.QueryRow(countQuery, countArgs...).Scan(&totalPosts)
	totalPages := (totalPosts + limit - 1) / limit

	categories := []string{"Игры", "Веб-сервисы", "Мобильные приложения", "Логистика", "Финтех", "Образование", "Здоровье", "Социальные сети", "ИИ и ML", "Другое"}

	var nickname string
	if userID > 0 {
		h.DB.QueryRow("SELECT nickname FROM users WHERE id = $1", userID).Scan(&nickname)
	}

	ban := h.getActiveBan(userID)

	data := map[string]interface{}{
		"Posts":      posts,
		"UserID":     userID,
		"UserRole":   role,
		"Nickname":   nickname,
		"Category":   category,
		"Search":     search,
		"Sort":       sort,
		"Page":       page,
		"TotalPages": totalPages,
		"Categories": categories,
		"ActiveBan":  ban,
	}

	h.renderTemplate(w, "feed.html", data)
}

func (h *Handler) HandleFeedAPI(w http.ResponseWriter, r *http.Request) {
	userID, _ := h.getCurrentUser(r)
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	category := r.URL.Query().Get("category")
	limit := 12
	offset := (page - 1) * limit

	query := `SELECT p.id, p.author_id, p.title, p.description, p.category, p.tags,
		p.view_count, p.created_at, u.nickname, u.avatar_url, u.role,
		COALESCE(AVG(r.score), 0),
		COUNT(DISTINCT r.id),
		(SELECT COUNT(*) FROM comments c WHERE c.post_id = p.id)
		FROM posts p JOIN users u ON p.author_id = u.id
		LEFT JOIN ratings r ON r.post_id = p.id
		WHERE p.is_hidden = FALSE`

	args := []interface{}{}
	argIdx := 1

	if category == "recommended" && userID > 0 {
		query += fmt.Sprintf(` AND p.category IN (SELECT ucv.category FROM user_category_views ucv WHERE ucv.user_id = $%d ORDER BY ucv.view_count DESC LIMIT 5)`, argIdx)
		args = append(args, userID)
		argIdx++
	} else if category != "" && category != "recommended" {
		query += fmt.Sprintf(" AND p.category = $%d", argIdx)
		args = append(args, category)
		argIdx++
	}

	query += " GROUP BY p.id, u.nickname, u.avatar_url, u.role"

	if userID > 0 && category == "" {
		query += fmt.Sprintf(` ORDER BY 
			CASE WHEN p.category IN (SELECT ucv.category FROM user_category_views ucv WHERE ucv.user_id = $%d ORDER BY ucv.view_count DESC LIMIT 3) THEN 0 ELSE 1 END,
			p.created_at DESC`, argIdx)
		args = append(args, userID)
		argIdx++
	} else {
		query += " ORDER BY p.created_at DESC"
	}

	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := h.DB.Query(query, args...)
	if err != nil {
		h.respondJSON(w, http.StatusOK, map[string]interface{}{"posts": []interface{}{}, "has_more": false})
		return
	}
	defer rows.Close()

	var posts []map[string]interface{}
	for rows.Next() {
		var id, authorID, viewCount, ratingCount, commentCount int
		var title, desc, cat, tags, nickname, avatarURL string
		var role models.UserRole
		var avgRating float64
		var createdAt time.Time
		rows.Scan(&id, &authorID, &title, &desc, &cat, &tags, &viewCount, &createdAt,
			&nickname, &avatarURL, &role, &avgRating, &ratingCount, &commentCount)

		var files []map[string]interface{}
		fRows, _ := h.DB.Query("SELECT id, file_path, file_type FROM files WHERE post_id = $1", id)
		if fRows != nil {
			for fRows.Next() {
				var fid int
				var fp, ft string
				fRows.Scan(&fid, &fp, &ft)
				files = append(files, map[string]interface{}{"id": fid, "file_path": fp, "file_type": ft})
			}
			fRows.Close()
		}

		posts = append(posts, map[string]interface{}{
			"id": id, "author_id": authorID, "title": title, "description": desc,
			"category": cat, "tags": tags, "view_count": viewCount,
			"created_at": createdAt, "author_nickname": nickname,
			"author_avatar_url": avatarURL, "author_role": string(role),
			"avg_rating": avgRating, "rating_count": ratingCount,
			"comment_count": commentCount, "files": files,
		})
	}

	if posts == nil {
		posts = []map[string]interface{}{}
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"posts":    posts,
		"has_more": len(posts) >= limit,
	})
}

func (h *Handler) HandleCreatePostPage(w http.ResponseWriter, r *http.Request) {
	userID, role := h.getCurrentUser(r)
	categories := []string{"Игры", "Веб-сервисы", "Мобильные приложения", "Логистика", "Финтех", "Образование", "Здоровье", "Социальные сети", "ИИ и ML", "Другое"}

	var nickname string
	h.DB.QueryRow("SELECT nickname FROM users WHERE id = $1", userID).Scan(&nickname)

	ban := h.getActiveBan(userID)

	data := map[string]interface{}{
		"UserID":     userID,
		"UserRole":   role,
		"Nickname":   nickname,
		"Categories": categories,
		"ActiveBan":  ban,
	}
	h.renderTemplate(w, "create_post.html", data)
}

func (h *Handler) HandleCreatePost(w http.ResponseWriter, r *http.Request) {
	userID, role := h.getCurrentUser(r)

	isHTMLFormRequest := strings.Contains(strings.ToLower(r.Header.Get("Content-Type")), "multipart/form-data") ||
		!strings.Contains(strings.ToLower(r.Header.Get("Accept")), "application/json")

	renderCreatePostError := func(status int, message string, activeBan *map[string]interface{}) {
		categories := []string{"Игры", "Веб-сервисы", "Мобильные приложения", "Логистика", "Финтех", "Образование", "Здоровье", "Социальные сети", "ИИ и ML", "Другое"}
		var nickname string
		h.DB.QueryRow("SELECT nickname FROM users WHERE id = $1", userID).Scan(&nickname)

		w.WriteHeader(status)
		h.renderTemplate(w, "create_post.html", map[string]interface{}{
			"UserID":     userID,
			"UserRole":   role,
			"Nickname":   nickname,
			"Categories": categories,
			"ActiveBan":  activeBan,
			"Error":      message,
		})
	}

	ban := h.getActiveBan(userID)
	if ban != nil {
		restriction := (*ban)["Restriction"].(string)
		if restriction == "posts" || restriction == "full" {
			if isHTMLFormRequest {
				renderCreatePostError(http.StatusForbidden, "Вы не можете создавать посты во время блокировки", ban)
				return
			}
			h.respondError(w, http.StatusForbidden, "Вы не можете создавать посты во время блокировки")
			return
		}
	}

	maxSize := int64(50 << 20)
	if role == "premium" || role == "admin" {
		maxSize = 100 << 20
	}
	r.ParseMultipartForm(maxSize)

	title := strings.TrimSpace(r.FormValue("title"))
	description := strings.TrimSpace(r.FormValue("description"))
	category := strings.TrimSpace(r.FormValue("category"))
	tags := strings.TrimSpace(r.FormValue("tags"))

	if title == "" || description == "" {
		if isHTMLFormRequest {
			renderCreatePostError(http.StatusBadRequest, "Заголовок и описание обязательны", ban)
			return
		}
		h.respondError(w, http.StatusBadRequest, "Заголовок и описание обязательны")
		return
	}

	isPremium := role == "premium" || role == "admin"

	var postID int
	err := h.DB.QueryRow(
		`INSERT INTO posts (author_id, title, description, category, tags, is_premium)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		userID, title, description, category, tags, isPremium,
	).Scan(&postID)
	if err != nil {
		if isHTMLFormRequest {
			renderCreatePostError(http.StatusInternalServerError, "Ошибка создания поста", ban)
			return
		}
		h.respondError(w, http.StatusInternalServerError, "Ошибка создания поста")
		return
	}

	if r.MultipartForm != nil && r.MultipartForm.File["files"] != nil {
		files := r.MultipartForm.File["files"]
		for _, fileHeader := range files {
			if fileHeader.Size > maxSize {
				continue
			}

			file, err := fileHeader.Open()
			if err != nil {
				continue
			}

			ext := filepath.Ext(fileHeader.Filename)
			newFilename := fmt.Sprintf("%d_%d%s", postID, time.Now().UnixNano(), ext)
			filePath := filepath.Join(h.UploadDir, newFilename)

			dst, err := os.Create(filePath)
			if err != nil {
				file.Close()
				continue
			}

			io.Copy(dst, file)
			dst.Close()
			file.Close()

			contentType := fileHeader.Header.Get("Content-Type")
			h.DB.Exec(
				`INSERT INTO files (post_id, filename, file_path, file_type, file_size)
				VALUES ($1, $2, $3, $4, $5)`,
				postID, fileHeader.Filename, "/uploads/"+newFilename, contentType, fileHeader.Size,
			)
		}
	}

	pollQuestion := strings.TrimSpace(r.FormValue("poll_question"))
	if pollQuestion != "" {
		var pollID int
		err := h.DB.QueryRow(
			"INSERT INTO polls (post_id, question) VALUES ($1, $2) RETURNING id",
			postID, pollQuestion,
		).Scan(&pollID)
		if err == nil {
			for i := 1; i <= 10; i++ {
				optText := strings.TrimSpace(r.FormValue(fmt.Sprintf("poll_option_%d", i)))
				if optText != "" {
					h.DB.Exec("INSERT INTO poll_options (poll_id, text) VALUES ($1, $2)", pollID, optText)
				}
			}
		}
	}

	h.logActivity(r, userID, "create_post", "post", postID, title)

	if strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") ||
		r.Header.Get("Accept") != "application/json" {
		http.Redirect(w, r, fmt.Sprintf("/post/%d", postID), http.StatusSeeOther)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"redirect": fmt.Sprintf("/post/%d", postID),
	})
}

func (h *Handler) HandlePostPage(w http.ResponseWriter, r *http.Request) {
	userID, role := h.getCurrentUser(r)
	postIDStr := strings.TrimPrefix(r.URL.Path, "/post/")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	h.DB.Exec("UPDATE posts SET view_count = view_count + 1 WHERE id = $1", postID)

	var p models.Post
	var pinnedUntil *time.Time
	err = h.DB.QueryRow(
		`SELECT p.id, p.author_id, p.title, p.description, p.category, p.tags,
		p.is_pinned, p.pinned_until, p.comments_off, p.is_hidden, p.view_count,
		p.is_premium, p.created_at, p.updated_at,
		u.nickname, u.avatar_url, u.role
		FROM posts p JOIN users u ON p.author_id = u.id WHERE p.id = $1`,
		postID,
	).Scan(&p.ID, &p.AuthorID, &p.Title, &p.Description, &p.Category,
		&p.Tags, &p.IsPinned, &pinnedUntil, &p.CommentsOff, &p.IsHidden,
		&p.ViewCount, &p.IsPremium, &p.CreatedAt, &p.UpdatedAt,
		&p.AuthorNickname, &p.AuthorAvatarURL, &p.AuthorRole)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	p.PinnedUntil = pinnedUntil

	if userID > 0 && p.Category != "" {
		h.DB.Exec(
			`INSERT INTO user_category_views (user_id, category, view_count) VALUES ($1, $2, 1)
			ON CONFLICT (user_id, category) DO UPDATE SET view_count = user_category_views.view_count + 1`,
			userID, p.Category,
		)
	}

	fileRows, _ := h.DB.Query("SELECT id, post_id, filename, file_path, file_type, file_size FROM files WHERE post_id = $1", postID)
	if fileRows != nil {
		for fileRows.Next() {
			var f models.File
			fileRows.Scan(&f.ID, &f.PostID, &f.Filename, &f.FilePath, &f.FileType, &f.FileSize)
			p.Files = append(p.Files, f)
		}
		fileRows.Close()
	}

	var avgRating float64
	var ratingCount int
	h.DB.QueryRow(
		"SELECT COALESCE(AVG(score), 0), COUNT(*) FROM ratings WHERE post_id = $1",
		postID,
	).Scan(&avgRating, &ratingCount)
	p.AvgRating = avgRating
	p.RatingCount = ratingCount

	commentRows, _ := h.DB.Query(
		`SELECT c.id, c.post_id, c.author_id, c.parent_id, c.content, c.like_count, c.created_at,
		u.nickname, u.avatar_url, u.role
		FROM comments c JOIN users u ON c.author_id = u.id
		WHERE c.post_id = $1 ORDER BY c.created_at ASC`,
		postID,
	)

	var allComments []models.Comment
	if commentRows != nil {
		for commentRows.Next() {
			var c models.Comment
			commentRows.Scan(&c.ID, &c.PostID, &c.AuthorID, &c.ParentID, &c.Content,
				&c.LikeCount, &c.CreatedAt, &c.AuthorNickname, &c.AuthorAvatarURL, &c.AuthorRole)

			if userID > 0 {
				var liked bool
				h.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM comment_likes WHERE comment_id = $1 AND user_id = $2)", c.ID, userID).Scan(&liked)
				c.LikedByUser = liked
			}

			allComments = append(allComments, c)
		}
		commentRows.Close()
	}

	var topComments []models.Comment
	commentMap := make(map[int]*models.Comment)
	for i := range allComments {
		commentMap[allComments[i].ID] = &allComments[i]
	}
	for i := range allComments {
		if allComments[i].ParentID == nil {
			topComments = append(topComments, allComments[i])
		} else {
			parent, ok := commentMap[*allComments[i].ParentID]
			if ok {
				parent.Replies = append(parent.Replies, allComments[i])
			}
		}
	}

	var userRating *models.Rating
	if userID > 0 {
		var rt models.Rating
		err := h.DB.QueryRow(
			"SELECT id, post_id, user_id, score, review, is_expert FROM ratings WHERE post_id = $1 AND user_id = $2",
			postID, userID,
		).Scan(&rt.ID, &rt.PostID, &rt.UserID, &rt.Score, &rt.Review, &rt.IsExpert)
		if err == nil {
			userRating = &rt
		}
	}

	ratingRows, _ := h.DB.Query(
		`SELECT r.id, r.score, r.review, r.is_expert, r.created_at, u.nickname, u.avatar_url, u.role
		FROM ratings r JOIN users u ON r.user_id = u.id
		WHERE r.post_id = $1 AND r.review != '' ORDER BY r.is_expert DESC, r.created_at DESC`,
		postID,
	)
	var reviews []map[string]interface{}
	if ratingRows != nil {
		for ratingRows.Next() {
			var id, score int
			var review string
			var isExpert bool
			var createdAt time.Time
			var nickname, avatarURL, userRole string
			ratingRows.Scan(&id, &score, &review, &isExpert, &createdAt, &nickname, &avatarURL, &userRole)
			reviews = append(reviews, map[string]interface{}{
				"ID":        id,
				"Score":     score,
				"Review":    review,
				"IsExpert":  isExpert,
				"CreatedAt": createdAt,
				"Nickname":  nickname,
				"AvatarURL": avatarURL,
				"UserRole":  userRole,
			})
		}
		ratingRows.Close()
	}

	var poll *models.Poll
	var pollRow models.Poll
	err = h.DB.QueryRow("SELECT id, post_id, question FROM polls WHERE post_id = $1", postID).Scan(&pollRow.ID, &pollRow.PostID, &pollRow.Question)
	if err == nil {
		optRows, _ := h.DB.Query("SELECT id, poll_id, text, vote_count FROM poll_options WHERE poll_id = $1", pollRow.ID)
		if optRows != nil {
			for optRows.Next() {
				var opt models.PollOption
				optRows.Scan(&opt.ID, &opt.PollID, &opt.Text, &opt.VoteCount)
				pollRow.Options = append(pollRow.Options, opt)
			}
			optRows.Close()
		}
		if userID > 0 {
			var votedOptID int
			h.DB.QueryRow("SELECT option_id FROM poll_votes WHERE poll_id = $1 AND user_id = $2", pollRow.ID, userID).Scan(&votedOptID)
			pollRow.VotedBy = votedOptID
		}
		poll = &pollRow
	}

	teamRows, _ := h.DB.Query(
		"SELECT id, post_id, author_id, title, description, skills, role_needed, is_open FROM team_requests WHERE post_id = $1 AND is_open = TRUE",
		postID,
	)
	var teamRequests []models.TeamRequest
	if teamRows != nil {
		for teamRows.Next() {
			var tr models.TeamRequest
			teamRows.Scan(&tr.ID, &tr.PostID, &tr.AuthorID, &tr.Title, &tr.Description, &tr.Skills, &tr.RoleNeeded, &tr.IsOpen)
			teamRequests = append(teamRequests, tr)
		}
		teamRows.Close()
	}

	var nickname string
	if userID > 0 {
		h.DB.QueryRow("SELECT nickname FROM users WHERE id = $1", userID).Scan(&nickname)
	}

	var projectID int
	h.DB.QueryRow("SELECT pp.project_id FROM project_posts pp JOIN projects pr ON pp.project_id = pr.id WHERE pp.post_id = $1 AND pr.user_id = $2 LIMIT 1", postID, p.AuthorID).Scan(&projectID)

	type FollowingUser struct {
		ID       int    `json:"id"`
		Nickname string `json:"nickname"`
	}
	var following []FollowingUser
	if userID > 0 {
		fRows, _ := h.DB.Query(
			`SELECT u.id, u.nickname FROM follows f JOIN users u ON f.following_id = u.id WHERE f.follower_id = $1 ORDER BY u.nickname`,
			userID,
		)
		if fRows != nil {
			for fRows.Next() {
				var fu FollowingUser
				fRows.Scan(&fu.ID, &fu.Nickname)
				following = append(following, fu)
			}
			fRows.Close()
		}
	}

	ban := h.getActiveBan(userID)

	// Pending consultations for this post (visible to post owner)
	var pendingConsultations []map[string]interface{}
	if userID == p.AuthorID {
		pcRows, _ := h.DB.Query(
			`SELECT ec.id, ec.title, ec.description, ec.price, u.nickname
			FROM expert_consultations ec JOIN users u ON ec.expert_id = u.id
			WHERE ec.post_id = $1 AND ec.status = 'pending'`, postID,
		)
		if pcRows != nil {
			for pcRows.Next() {
				var cID int
				var cTitle, cDesc, cNick string
				var cPrice float64
				pcRows.Scan(&cID, &cTitle, &cDesc, &cPrice, &cNick)
				pendingConsultations = append(pendingConsultations, map[string]interface{}{
					"ID": cID, "Title": cTitle, "Description": cDesc, "Price": cPrice, "ExpertNickname": cNick,
				})
			}
			pcRows.Close()
		}
	}

	// Compute poll total votes and options for template
	var pollTotalVotes int
	var pollOptions []models.PollOption
	if poll != nil {
		for _, opt := range poll.Options {
			pollTotalVotes += opt.VoteCount
		}
		pollOptions = poll.Options
	}

	data := map[string]interface{}{
		"Post":                  p,
		"Comments":              topComments,
		"UserRating":            userRating,
		"Reviews":               reviews,
		"Poll":                  poll,
		"PollTotalVotes":        pollTotalVotes,
		"PollOptions":           pollOptions,
		"TeamRequests":          teamRequests,
		"Following":             following,
		"UserID":                userID,
		"UserRole":              role,
		"Nickname":              nickname,
		"IsOwner":               userID == p.AuthorID,
		"ProjectID":             projectID,
		"ActiveBan":             ban,
		"PendingConsultations":  pendingConsultations,
	}

	h.renderTemplate(w, "post.html", data)
}

func (h *Handler) HandleEditPostPage(w http.ResponseWriter, r *http.Request) {
	userID, role := h.getCurrentUser(r)
	// URL: /post/ID/edit
	path := strings.TrimPrefix(r.URL.Path, "/post/")
	path = strings.TrimSuffix(path, "/edit")
	postID, err := strconv.Atoi(path)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	var p models.Post
	err = h.DB.QueryRow(
		"SELECT id, author_id, title, description, category, tags, comments_off FROM posts WHERE id = $1",
		postID,
	).Scan(&p.ID, &p.AuthorID, &p.Title, &p.Description, &p.Category, &p.Tags, &p.CommentsOff)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if p.AuthorID != userID && role != "admin" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	fileRows, _ := h.DB.Query("SELECT id, post_id, filename, file_path, file_type, file_size FROM files WHERE post_id = $1", postID)
	if fileRows != nil {
		for fileRows.Next() {
			var f models.File
			fileRows.Scan(&f.ID, &f.PostID, &f.Filename, &f.FilePath, &f.FileType, &f.FileSize)
			p.Files = append(p.Files, f)
		}
		fileRows.Close()
	}

	categories := []string{"Игры", "Веб-сервисы", "Мобильные приложения", "Логистика", "Финтех", "Образование", "Здоровье", "Социальные сети", "ИИ и ML", "Другое"}

	var nickname string
	h.DB.QueryRow("SELECT nickname FROM users WHERE id = $1", userID).Scan(&nickname)

	data := map[string]interface{}{
		"Post":       p,
		"UserID":     userID,
		"UserRole":   role,
		"Nickname":   nickname,
		"Categories": categories,
	}
	h.renderTemplate(w, "edit_post.html", data)
}

func (h *Handler) HandleUpdatePost(w http.ResponseWriter, r *http.Request) {
	userID, role := h.getCurrentUser(r)
	postIDStr := strings.TrimPrefix(r.URL.Path, "/api/posts/")
	postIDStr = strings.TrimSuffix(postIDStr, "/update")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Некорректный ID")
		return
	}

	var authorID int
	h.DB.QueryRow("SELECT author_id FROM posts WHERE id = $1", postID).Scan(&authorID)
	if authorID != userID && role != "admin" {
		h.respondError(w, http.StatusForbidden, "Нет доступа")
		return
	}

	maxSize := int64(50 << 20)
	if role == "premium" || role == "admin" {
		maxSize = 100 << 20
	}
	r.ParseMultipartForm(maxSize)

	title := strings.TrimSpace(r.FormValue("title"))
	description := strings.TrimSpace(r.FormValue("description"))
	category := strings.TrimSpace(r.FormValue("category"))
	tags := strings.TrimSpace(r.FormValue("tags"))
	commentsOff := r.FormValue("comments_off") == "on"

	if title == "" || description == "" {
		h.respondError(w, http.StatusBadRequest, "Заголовок и описание обязательны")
		return
	}

	h.DB.Exec(
		"UPDATE posts SET title=$1, description=$2, category=$3, tags=$4, comments_off=$5, updated_at=NOW() WHERE id=$6",
		title, description, category, tags, commentsOff, postID,
	)

	if r.MultipartForm != nil && r.MultipartForm.File["files"] != nil {
		files := r.MultipartForm.File["files"]
		for _, fileHeader := range files {
			if fileHeader.Size > maxSize {
				continue
			}
			file, err := fileHeader.Open()
			if err != nil {
				continue
			}
			ext := filepath.Ext(fileHeader.Filename)
			newFilename := fmt.Sprintf("%d_%d%s", postID, time.Now().UnixNano(), ext)
			filePath := filepath.Join(h.UploadDir, newFilename)
			dst, err := os.Create(filePath)
			if err != nil {
				file.Close()
				continue
			}
			io.Copy(dst, file)
			dst.Close()
			file.Close()
			contentType := fileHeader.Header.Get("Content-Type")
			h.DB.Exec(
				"INSERT INTO files (post_id, filename, file_path, file_type, file_size) VALUES ($1, $2, $3, $4, $5)",
				postID, fileHeader.Filename, "/uploads/"+newFilename, contentType, fileHeader.Size,
			)
		}
	}

	h.logActivity(r, userID, "update_post", "post", postID, title)

	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		http.Redirect(w, r, fmt.Sprintf("/post/%d", postID), http.StatusSeeOther)
		return
	}
	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"redirect": fmt.Sprintf("/post/%d", postID),
	})
}

func (h *Handler) HandleDeletePost(w http.ResponseWriter, r *http.Request) {
	userID, role := h.getCurrentUser(r)
	postIDStr := strings.TrimPrefix(r.URL.Path, "/api/posts/")
	postIDStr = strings.TrimSuffix(postIDStr, "/delete")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Некорректный ID")
		return
	}

	var authorID int
	h.DB.QueryRow("SELECT author_id FROM posts WHERE id = $1", postID).Scan(&authorID)
	if authorID != userID && role != "admin" {
		h.respondError(w, http.StatusForbidden, "Нет доступа")
		return
	}

	h.DB.Exec("DELETE FROM posts WHERE id = $1", postID)
	h.logActivity(r, userID, "delete_post", "post", postID, "")

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"redirect": "/feed",
	})
}

func (h *Handler) HandleDeleteFile(w http.ResponseWriter, r *http.Request) {
	userID, role := h.getCurrentUser(r)
	fileIDStr := strings.TrimPrefix(r.URL.Path, "/api/files/")
	fileIDStr = strings.TrimSuffix(fileIDStr, "/delete")
	fileID, err := strconv.Atoi(fileIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Некорректный ID")
		return
	}

	var postID int
	var filePath string
	h.DB.QueryRow("SELECT post_id, file_path FROM files WHERE id = $1", fileID).Scan(&postID, &filePath)

	var authorID int
	h.DB.QueryRow("SELECT author_id FROM posts WHERE id = $1", postID).Scan(&authorID)
	if authorID != userID && role != "admin" {
		h.respondError(w, http.StatusForbidden, "Нет доступа")
		return
	}

	h.DB.Exec("DELETE FROM files WHERE id = $1", fileID)

	if filePath != "" {
		os.Remove("." + filePath)
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}
