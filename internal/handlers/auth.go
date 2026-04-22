package handlers

import (
	"net/http"
	"strings"
	"time"

	"startup-platform/internal/middleware"
)

func (h *Handler) HandleRegisterPage(w http.ResponseWriter, r *http.Request) {
	h.renderTemplate(w, "register.html", nil)
}

func (h *Handler) HandleLoginPage(w http.ResponseWriter, r *http.Request) {
	h.renderTemplate(w, "login.html", nil)
}

func (h *Handler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.respondError(w, http.StatusBadRequest, "Некорректные данные")
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))
	phone := strings.TrimSpace(r.FormValue("phone"))
	nickname := strings.TrimSpace(r.FormValue("nickname"))
	password := r.FormValue("password")
	name := strings.TrimSpace(r.FormValue("name"))

	if email == "" || nickname == "" || password == "" {
		h.respondError(w, http.StatusBadRequest, "Email, никнейм и пароль обязательны")
		return
	}

	if len(password) < 6 {
		h.respondError(w, http.StatusBadRequest, "Пароль должен быть не менее 6 символов")
		return
	}

	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		h.respondError(w, http.StatusBadRequest, "Некорректный email")
		return
	}

	var exists bool
	err := h.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", email).Scan(&exists)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Ошибка сервера")
		return
	}
	if exists {
		h.respondError(w, http.StatusConflict, "Email уже зарегистрирован")
		return
	}

	err = h.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE nickname = $1)", nickname).Scan(&exists)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Ошибка сервера")
		return
	}
	if exists {
		h.respondError(w, http.StatusConflict, "Никнейм уже занят")
		return
	}

	passwordHash := hashPassword(password)

	var userID int
	err = h.DB.QueryRow(
		`INSERT INTO users (email, phone, nickname, password_hash, name, role) 
		VALUES ($1, $2, $3, $4, $5, 'user') RETURNING id`,
		email, phone, nickname, passwordHash, name,
	).Scan(&userID)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Ошибка создания аккаунта")
		return
	}

	token, err := middleware.GenerateJWT(h.JWTSecret, userID, "user")
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Ошибка авторизации")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   259200,
		SameSite: http.SameSiteLaxMode,
	})

	h.logActivity(r, userID, "register", "user", userID, email)

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"redirect": "/feed",
	})
}

func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.respondError(w, http.StatusBadRequest, "Некорректные данные")
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")

	if email == "" || password == "" {
		h.respondError(w, http.StatusBadRequest, "Email и пароль обязательны")
		return
	}

	passwordHash := hashPassword(password)

	var userID int
	var role string
	var isBlocked bool
	err := h.DB.QueryRow(
		"SELECT id, role, is_blocked FROM users WHERE email = $1 AND password_hash = $2",
		email, passwordHash,
	).Scan(&userID, &role, &isBlocked)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "Неверный email или пароль")
		return
	}

	if isBlocked {
		h.respondError(w, http.StatusForbidden, "Аккаунт заблокирован")
		return
	}

	token, err := middleware.GenerateJWT(h.JWTSecret, userID, role)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Ошибка авторизации")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   259200,
		SameSite: http.SameSiteLaxMode,
	})

	h.logActivity(r, userID, "login", "user", userID, "")

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"redirect": "/feed",
	})
}

func (h *Handler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	userID, _ := h.getCurrentUser(r)
	if userID > 0 {
		h.logActivity(r, userID, "logout", "user", userID, "")
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
