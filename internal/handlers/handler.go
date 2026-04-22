package handlers

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"startup-platform/internal/middleware"
)

type Handler struct {
	DB              *sql.DB
	Templates       *template.Template
	JWTSecret       string
	UploadDir       string
	ActiveChatUsers sync.Map
}

func NewHandler(db *sql.DB, tmplDir, jwtSecret, uploadDir string) *Handler {
	funcMap := template.FuncMap{
		"formatDate": func(t time.Time) string {
			return t.Format("02.01.2006 15:04")
		},
		"timeAgo": func(t time.Time) string {
			d := time.Since(t)
			switch {
			case d.Minutes() < 1:
				return "только что"
			case d.Minutes() < 60:
				m := int(d.Minutes())
				return fmt.Sprintf("%d мин", m)
			case d.Hours() < 24:
				h := int(d.Hours())
				return fmt.Sprintf("%d ч", h)
			case d.Hours() < 24*7:
				days := int(d.Hours() / 24)
				return fmt.Sprintf("%d д", days)
			default:
				return t.Format("02.01.2006")
			}
		},
		"truncate": func(s string, n int) string {
			if len([]rune(s)) <= n {
				return s
			}
			return string([]rune(s)[:n]) + "..."
		},
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
		"seq": func(n int) []int {
			s := make([]int, n)
			for i := range s {
				s[i] = i + 1
			}
			return s
		},
		"dict": func(values ...interface{}) map[string]interface{} {
			m := make(map[string]interface{})
			for i := 0; i < len(values)-1; i += 2 {
				key, ok := values[i].(string)
				if ok {
					m[key] = values[i+1]
				}
			}
			return m
		},
		"isImage": func(ft string) bool {
			return strings.HasPrefix(ft, "image/")
		},
		"isVideo": func(ft string) bool {
			return strings.HasPrefix(ft, "video/")
		},
		"isAudio": func(ft string) bool {
			return strings.HasPrefix(ft, "audio/")
		},
		"json": func(v interface{}) template.JS {
			b, _ := json.Marshal(v)
			return template.JS(b)
		},
		"toJSON": func(v interface{}) string {
			b, _ := json.Marshal(v)
			return string(b)
		},
		"split": func(s, sep string) []string {
			if s == "" {
				return nil
			}
			return strings.Split(s, sep)
		},
		"contains": func(s, sub string) bool {
			return strings.Contains(s, sub)
		},
		"eq": func(a, b interface{}) bool {
			return a == b
		},
		"ne": func(a, b interface{}) bool {
			return a != b
		},
		"gt": func(a, b int) bool {
			return a > b
		},
		"formatSize": func(size int64) string {
			if size < 1024 {
				return fmt.Sprintf("%d Б", size)
			}
			if size < 1024*1024 {
				return fmt.Sprintf("%.1f КБ", float64(size)/1024)
			}
			return fmt.Sprintf("%.1f МБ", float64(size)/(1024*1024))
		},
		"mod": func(a, b int) int {
			return a % b
		},
		"mkSlice": func(args ...string) []string {
			return args
		},
		"intRange": func(start, end int) []int {
			var r []int
			for i := start; i < end; i++ {
				r = append(r, i)
			}
			return r
		},
		"toInt": func(v interface{}) int {
			switch val := v.(type) {
			case int:
				return val
			case int64:
				return int(val)
			case float64:
				return int(val)
			case float32:
				return int(val)
			default:
				return 0
			}
		},
		"toFloat": func(v interface{}) float64 {
			switch val := v.(type) {
			case int:
				return float64(val)
			case int64:
				return float64(val)
			case float64:
				return val
			case float32:
				return float64(val)
			default:
				return 0
			}
		},
		"divFloat": func(a, b float64) float64 {
			if b == 0 {
				return 0
			}
			return a / b
		},
		"mulFloat": func(a, b float64) float64 {
			return a * b
		},
		"chatContent": func(s string) template.HTML {
			safe := template.HTMLEscapeString(s)
			re := regexp.MustCompile(`\[([^\]]+)\]\((/[^)]+)\)`)
			rendered := re.ReplaceAllString(safe, `<a href="$2" style="color:var(--primary);text-decoration:underline">$1</a>`)
			return template.HTML(rendered)
		},
		"roundInt": func(v interface{}) int {
			switch val := v.(type) {
			case float64:
				return int(math.Round(val))
			case float32:
				return int(math.Round(float64(val)))
			case int:
				return val
			case int64:
				return int(val)
			default:
				return 0
			}
		},
	}

	pattern := filepath.Join(tmplDir, "*.html")
	tmpl := template.Must(template.New("").Funcs(funcMap).ParseGlob(pattern))

	return &Handler{
		DB:        db,
		Templates: tmpl,
		JWTSecret: jwtSecret,
		UploadDir: uploadDir,
	}
}

func hashPassword(password string) string {
	h := sha256.New()
	h.Write([]byte(password))
	return hex.EncodeToString(h.Sum(nil))
}

func (h *Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{"error": message})
}

func (h *Handler) renderTemplate(w http.ResponseWriter, name string, data interface{}) {
	var buf bytes.Buffer
	if err := h.Templates.ExecuteTemplate(&buf, name, data); err != nil {
		log.Printf("Template error %s: %v", name, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)
}

func (h *Handler) getCurrentUser(r *http.Request) (int, string) {
	userID := middleware.GetUserID(r)
	role := middleware.GetUserRole(r)
	return userID, role
}

func (h *Handler) logActivity(r *http.Request, userID int, action, targetType string, targetID int, details string) {
	ip := r.RemoteAddr
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		ip = strings.Split(fwd, ",")[0]
	}
	h.DB.Exec(
		"INSERT INTO activity_logs (user_id, action, target_type, target_id, details, ip_address) VALUES ($1, $2, $3, $4, $5, $6)",
		userID, action, targetType, targetID, details, ip,
	)
}

func (h *Handler) getActiveBan(userID int) *map[string]interface{} {
	var banID int
	var reason, restriction, targetLink, adminComment string
	var expiresAt time.Time
	err := h.DB.QueryRow(
		`SELECT id, reason, restriction, target_link, admin_comment, expires_at 
		FROM user_bans WHERE user_id = $1 AND expires_at > NOW() 
		ORDER BY expires_at DESC LIMIT 1`,
		userID,
	).Scan(&banID, &reason, &restriction, &targetLink, &adminComment, &expiresAt)
	if err != nil {
		return nil
	}
	ban := map[string]interface{}{
		"ID":           banID,
		"Reason":       reason,
		"Restriction":  restriction,
		"TargetLink":   targetLink,
		"AdminComment": adminComment,
		"ExpiresAt":    expiresAt,
		"DaysLeft":     int(time.Until(expiresAt).Hours()/24) + 1,
	}
	return &ban
}
