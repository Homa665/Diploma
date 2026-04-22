package handlers

import (
	"net/http"
	"strconv"
	"strings"
)

func (h *Handler) HandleAddComment(w http.ResponseWriter, r *http.Request) {
	userID, _ := h.getCurrentUser(r)
	r.ParseForm()

	ban := h.getActiveBan(userID)
	if ban != nil {
		restriction := (*ban)["Restriction"].(string)
		if restriction == "comments" || restriction == "full" {
			h.respondError(w, http.StatusForbidden, "Вы не можете писать комментарии во время блокировки")
			return
		}
	}

	postIDStr := r.FormValue("post_id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Некорректный ID поста")
		return
	}

	content := strings.TrimSpace(r.FormValue("content"))
	if content == "" {
		h.respondError(w, http.StatusBadRequest, "Комментарий не может быть пустым")
		return
	}

	var commentsOff bool
	h.DB.QueryRow("SELECT comments_off FROM posts WHERE id = $1", postID).Scan(&commentsOff)
	if commentsOff {
		h.respondError(w, http.StatusForbidden, "Комментарии отключены")
		return
	}

	parentIDStr := r.FormValue("parent_id")
	var parentID *int
	if parentIDStr != "" {
		pid, err := strconv.Atoi(parentIDStr)
		if err == nil {
			parentID = &pid
		}
	}

	var commentID int
	err = h.DB.QueryRow(
		"INSERT INTO comments (post_id, author_id, parent_id, content) VALUES ($1, $2, $3, $4) RETURNING id",
		postID, userID, parentID, content,
	).Scan(&commentID)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Ошибка добавления комментария")
		return
	}

	var postAuthorID int
	h.DB.QueryRow("SELECT author_id FROM posts WHERE id = $1", postID).Scan(&postAuthorID)
	if postAuthorID != userID {
		var nickname string
		h.DB.QueryRow("SELECT nickname FROM users WHERE id = $1", userID).Scan(&nickname)
		h.DB.Exec(
			"INSERT INTO notifications (user_id, type, content, link) VALUES ($1, 'comment', $2, $3)",
			postAuthorID, nickname+" оставил комментарий к вашему посту", "/post/"+postIDStr,
		)
	}

	h.logActivity(r, userID, "add_comment", "comment", commentID, "")

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"success": true, "comment_id": commentID})
}

func (h *Handler) HandleDeleteComment(w http.ResponseWriter, r *http.Request) {
	userID, role := h.getCurrentUser(r)
	commentIDStr := strings.TrimPrefix(r.URL.Path, "/api/comments/")
	commentIDStr = strings.TrimSuffix(commentIDStr, "/delete")
	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Некорректный ID")
		return
	}

	var authorID, postID int
	h.DB.QueryRow("SELECT author_id, post_id FROM comments WHERE id = $1", commentID).Scan(&authorID, &postID)

	var postAuthorID int
	h.DB.QueryRow("SELECT author_id FROM posts WHERE id = $1", postID).Scan(&postAuthorID)

	if authorID != userID && postAuthorID != userID && role != "admin" {
		h.respondError(w, http.StatusForbidden, "Нет доступа")
		return
	}

	h.DB.Exec("DELETE FROM comments WHERE id = $1", commentID)
	h.logActivity(r, userID, "delete_comment", "comment", commentID, "")
	h.respondJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}

func (h *Handler) HandleLikeComment(w http.ResponseWriter, r *http.Request) {
	userID, _ := h.getCurrentUser(r)
	commentIDStr := strings.TrimPrefix(r.URL.Path, "/api/comments/")
	commentIDStr = strings.TrimSuffix(commentIDStr, "/like")
	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Некорректный ID")
		return
	}

	var exists bool
	h.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM comment_likes WHERE comment_id = $1 AND user_id = $2)", commentID, userID).Scan(&exists)

	if exists {
		h.DB.Exec("DELETE FROM comment_likes WHERE comment_id = $1 AND user_id = $2", commentID, userID)
		h.DB.Exec("UPDATE comments SET like_count = like_count - 1 WHERE id = $1", commentID)
	} else {
		h.DB.Exec("INSERT INTO comment_likes (comment_id, user_id) VALUES ($1, $2)", commentID, userID)
		h.DB.Exec("UPDATE comments SET like_count = like_count + 1 WHERE id = $1", commentID)
	}

	var likeCount int
	h.DB.QueryRow("SELECT like_count FROM comments WHERE id = $1", commentID).Scan(&likeCount)

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"success":    true,
		"liked":      !exists,
		"like_count": likeCount,
	})
}

func (h *Handler) HandleRate(w http.ResponseWriter, r *http.Request) {
	userID, role := h.getCurrentUser(r)
	r.ParseForm()

	postIDStr := r.FormValue("post_id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Некорректный ID поста")
		return
	}

	scoreStr := r.FormValue("score")
	score, err := strconv.Atoi(scoreStr)
	if err != nil || score < 1 || score > 5 {
		h.respondError(w, http.StatusBadRequest, "Оценка должна быть от 1 до 5")
		return
	}

	review := strings.TrimSpace(r.FormValue("review"))
	isExpert := role == "expert"

	var existingID int
	err = h.DB.QueryRow("SELECT id FROM ratings WHERE post_id = $1 AND user_id = $2", postID, userID).Scan(&existingID)
	if err == nil {
		h.DB.Exec("UPDATE ratings SET score = $1, review = $2, is_expert = $3 WHERE id = $4",
			score, review, isExpert, existingID)
	} else {
		h.DB.Exec(
			"INSERT INTO ratings (post_id, user_id, score, review, is_expert) VALUES ($1, $2, $3, $4, $5)",
			postID, userID, score, review, isExpert,
		)
	}

	var postAuthorID int
	h.DB.QueryRow("SELECT author_id FROM posts WHERE id = $1", postID).Scan(&postAuthorID)
	if postAuthorID != userID {
		var nickname string
		h.DB.QueryRow("SELECT nickname FROM users WHERE id = $1", userID).Scan(&nickname)
		h.DB.Exec(
			"INSERT INTO notifications (user_id, type, content, link) VALUES ($1, 'rating', $2, $3)",
			postAuthorID, nickname+" оценил ваш проект на "+scoreStr, "/post/"+postIDStr,
		)
	}

	h.logActivity(r, userID, "rate_post", "post", postID, scoreStr)

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}

func (h *Handler) HandleVotePoll(w http.ResponseWriter, r *http.Request) {
	userID, _ := h.getCurrentUser(r)
	r.ParseForm()

	pollIDStr := r.FormValue("poll_id")
	pollID, err := strconv.Atoi(pollIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Некорректный ID опроса")
		return
	}

	optionIDStr := r.FormValue("option_id")
	optionID, err := strconv.Atoi(optionIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Некорректный вариант")
		return
	}

	var exists bool
	h.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM poll_votes WHERE poll_id = $1 AND user_id = $2)", pollID, userID).Scan(&exists)
	if exists {
		h.respondError(w, http.StatusConflict, "Вы уже голосовали")
		return
	}

	h.DB.Exec("INSERT INTO poll_votes (poll_id, option_id, user_id) VALUES ($1, $2, $3)", pollID, optionID, userID)
	h.DB.Exec("UPDATE poll_options SET vote_count = vote_count + 1 WHERE id = $1", optionID)

	h.respondJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}
