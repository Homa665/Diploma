package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"startup-platform/internal/config"
	"startup-platform/internal/database"
	"startup-platform/internal/handlers"
	"startup-platform/internal/middleware"
	"startup-platform/internal/seed"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Database connection failed: ", err)
	}
	defer db.Close()

	if err := database.Migrate(db); err != nil {
		log.Fatal("Migration failed: ", err)
	}

	if err := seed.Run(db); err != nil {
		log.Printf("Seed warning: %v", err)
	}

	if err := os.MkdirAll(cfg.UploadDir, 0755); err != nil {
		log.Fatal("Failed to create upload directory: ", err)
	}

	h := handlers.NewHandler(db, "./web/templates", cfg.JWTSecret, cfg.UploadDir)

	mux := http.NewServeMux()

	auth := middleware.AuthMiddleware(cfg.JWTSecret, db)
	apiAuth := middleware.APIAuthMiddleware(cfg.JWTSecret, db)
	optAuth := middleware.OptionalAuthMiddleware(cfg.JWTSecret, db)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		optAuth(http.HandlerFunc(h.HandleHomePage)).ServeHTTP(w, r)
	})

	mux.Handle("/register", http.HandlerFunc(h.HandleRegisterPage))
	mux.Handle("/login", http.HandlerFunc(h.HandleLoginPage))
	mux.Handle("/api/register", http.HandlerFunc(h.HandleRegister))
	mux.Handle("/api/login", http.HandlerFunc(h.HandleLogin))
	mux.Handle("/logout", http.HandlerFunc(h.HandleLogout))

	mux.Handle("/feed", auth(http.HandlerFunc(h.HandleFeedPage)))
	mux.Handle("/api/feed", apiAuth(http.HandlerFunc(h.HandleFeedAPI)))
	mux.HandleFunc("/post/new", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			auth(http.HandlerFunc(h.HandleCreatePost)).ServeHTTP(w, r)
			return
		}
		auth(http.HandlerFunc(h.HandleCreatePostPage)).ServeHTTP(w, r)
	})
	mux.Handle("/api/posts/create", apiAuth(http.HandlerFunc(h.HandleCreatePost)))

	mux.HandleFunc("/post/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/edit") {
			auth(http.HandlerFunc(h.HandleEditPostPage)).ServeHTTP(w, r)
			return
		}
		auth(http.HandlerFunc(h.HandlePostPage)).ServeHTTP(w, r)
	})

	mux.HandleFunc("/api/posts/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasSuffix(path, "/delete") {
			apiAuth(http.HandlerFunc(h.HandleDeletePost)).ServeHTTP(w, r)
			return
		}
		if strings.HasSuffix(path, "/update") {
			apiAuth(http.HandlerFunc(h.HandleUpdatePost)).ServeHTTP(w, r)
			return
		}
		http.NotFound(w, r)
	})

	mux.HandleFunc("/api/files/", func(w http.ResponseWriter, r *http.Request) {
		apiAuth(http.HandlerFunc(h.HandleDeleteFile)).ServeHTTP(w, r)
	})

	mux.Handle("/api/comments/add", apiAuth(http.HandlerFunc(h.HandleAddComment)))
	mux.HandleFunc("/api/comments/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if len(path) > 15 {
			if path[len(path)-7:] == "/delete" {
				apiAuth(http.HandlerFunc(h.HandleDeleteComment)).ServeHTTP(w, r)
				return
			}
			if path[len(path)-5:] == "/like" {
				apiAuth(http.HandlerFunc(h.HandleLikeComment)).ServeHTTP(w, r)
				return
			}
		}
		http.NotFound(w, r)
	})

	mux.Handle("/api/rate", apiAuth(http.HandlerFunc(h.HandleRate)))
	mux.Handle("/api/poll/vote", apiAuth(http.HandlerFunc(h.HandleVotePoll)))

	mux.Handle("/chat", auth(http.HandlerFunc(h.HandleChatPage)))
	mux.HandleFunc("/chat/", func(w http.ResponseWriter, r *http.Request) {
		auth(http.HandlerFunc(h.HandleChatWith)).ServeHTTP(w, r)
	})
	mux.Handle("/api/messages/send", apiAuth(http.HandlerFunc(h.HandleSendMessage)))
	mux.Handle("/api/messages", apiAuth(http.HandlerFunc(h.HandleGetMessages)))
	mux.Handle("/api/share-to-chat", apiAuth(http.HandlerFunc(h.HandleShareToChat)))

	mux.Handle("/api/notifications", apiAuth(http.HandlerFunc(h.HandleGetNotifications)))
	mux.HandleFunc("/api/notifications/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/api/notifications/read-all" {
			apiAuth(http.HandlerFunc(h.HandleMarkAllNotificationsRead)).ServeHTTP(w, r)
			return
		}
		if len(path) > 20 && path[len(path)-5:] == "/read" {
			apiAuth(http.HandlerFunc(h.HandleMarkNotificationRead)).ServeHTTP(w, r)
			return
		}
		http.NotFound(w, r)
	})
	mux.Handle("/notifications", auth(http.HandlerFunc(h.HandleNotificationsPage)))

	mux.HandleFunc("/profile/", func(w http.ResponseWriter, r *http.Request) {
		auth(http.HandlerFunc(h.HandleProfilePage)).ServeHTTP(w, r)
	})
	mux.Handle("/settings", auth(http.HandlerFunc(h.HandleSettingsPage)))
	mux.Handle("/premium-stats", auth(http.HandlerFunc(h.HandlePremiumStats)))

	mux.Handle("/api/profile/edit", apiAuth(http.HandlerFunc(h.HandleEditProfile)))
	mux.Handle("/api/profile/avatar", apiAuth(http.HandlerFunc(h.HandleAvatarUpload)))
	mux.Handle("/api/follow", apiAuth(http.HandlerFunc(h.HandleFollow)))
	mux.Handle("/api/unfollow", apiAuth(http.HandlerFunc(h.HandleUnfollow)))
	mux.Handle("/api/followers", apiAuth(http.HandlerFunc(h.HandleFollowersList)))
	mux.Handle("/api/following", apiAuth(http.HandlerFunc(h.HandleFollowingList)))

	mux.Handle("/api/friends/add", apiAuth(http.HandlerFunc(h.HandleAddFriend)))
	mux.Handle("/api/friends/remove", apiAuth(http.HandlerFunc(h.HandleRemoveFriend)))
	mux.Handle("/api/block", apiAuth(http.HandlerFunc(h.HandleBlockUser)))
	mux.Handle("/api/unblock", apiAuth(http.HandlerFunc(h.HandleUnblockUser)))
	mux.Handle("/api/repost", apiAuth(http.HandlerFunc(h.HandleRepost)))
	mux.Handle("/api/complaint", apiAuth(http.HandlerFunc(h.HandleComplaint)))

	mux.Handle("/api/projects/create", apiAuth(http.HandlerFunc(h.HandleCreateProject)))
	mux.Handle("/api/projects/delete", apiAuth(http.HandlerFunc(h.HandleDeleteProject)))
	mux.Handle("/api/projects/add-post", apiAuth(http.HandlerFunc(h.HandleAddPostToProject)))
	mux.Handle("/api/projects/remove-post", apiAuth(http.HandlerFunc(h.HandleRemovePostFromProject)))
	mux.HandleFunc("/project/", func(w http.ResponseWriter, r *http.Request) {
		auth(http.HandlerFunc(h.HandleProjectPage)).ServeHTTP(w, r)
	})

	mux.Handle("/teams", auth(http.HandlerFunc(h.HandleTeamSearchPage)))
	mux.Handle("/api/team/request", apiAuth(http.HandlerFunc(h.HandleTeamRequestCreate)))
	mux.Handle("/api/team/respond", apiAuth(http.HandlerFunc(h.HandleTeamRespond)))

	mux.Handle("/api/expert/apply", apiAuth(http.HandlerFunc(h.HandleExpertApply)))
	mux.Handle("/expert/apply", auth(http.HandlerFunc(h.HandleExpertApplyPage)))
	mux.Handle("/api/expert/consultation", apiAuth(http.HandlerFunc(h.HandleExpertConsultation)))
	mux.Handle("/api/consultation/respond", apiAuth(http.HandlerFunc(h.HandleConsultationRespond)))
	mux.Handle("/api/consultation/send", apiAuth(http.HandlerFunc(h.HandleConsultationSendMessage)))
	mux.Handle("/api/consultation/complete", apiAuth(http.HandlerFunc(h.HandleConsultationComplete)))
	mux.Handle("/api/consultation/messages", apiAuth(http.HandlerFunc(h.HandleConsultationGetMessages)))
	mux.HandleFunc("/consultation/", func(w http.ResponseWriter, r *http.Request) {
		auth(http.HandlerFunc(h.HandleConsultationPage)).ServeHTTP(w, r)
	})
	mux.Handle("/consultations", auth(http.HandlerFunc(h.HandleMyConsultations)))
	mux.Handle("/api/subscribe", apiAuth(http.HandlerFunc(h.HandleSubscribe)))
	mux.Handle("/pricing", auth(http.HandlerFunc(h.HandlePricingPage)))

	mux.Handle("/admin", auth(middleware.AdminMiddleware(http.HandlerFunc(h.HandleAdminPage))))
	mux.Handle("/api/admin/block-user", apiAuth(middleware.AdminMiddleware(http.HandlerFunc(h.HandleAdminBlockUser))))
	mux.Handle("/api/admin/ban-user", apiAuth(middleware.AdminMiddleware(http.HandlerFunc(h.HandleAdminBanUser))))
	mux.Handle("/api/admin/change-role", apiAuth(middleware.AdminMiddleware(http.HandlerFunc(h.HandleAdminChangeRole))))
	mux.HandleFunc("/api/admin/complaints/", func(w http.ResponseWriter, r *http.Request) {
		apiAuth(middleware.AdminMiddleware(http.HandlerFunc(h.HandleAdminComplaint))).ServeHTTP(w, r)
	})
	mux.HandleFunc("/api/admin/expert-apps/", func(w http.ResponseWriter, r *http.Request) {
		apiAuth(middleware.AdminMiddleware(http.HandlerFunc(h.HandleAdminExpertApp))).ServeHTTP(w, r)
	})
	mux.Handle("/api/admin/hide-post", apiAuth(middleware.AdminMiddleware(http.HandlerFunc(h.HandleAdminHidePost))))
	mux.Handle("/api/admin/delete-post", apiAuth(middleware.AdminMiddleware(http.HandlerFunc(h.HandleAdminDeletePost))))
	mux.Handle("/api/admin/delete-user", apiAuth(middleware.AdminMiddleware(http.HandlerFunc(h.HandleAdminDeleteUser))))

	fs := http.FileServer(http.Dir("./web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	uploadFS := http.FileServer(http.Dir(cfg.UploadDir))
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", uploadFS))

	addr := ":" + cfg.Port
	fmt.Printf("Server starting on http://localhost%s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal("Server failed: ", err)
	}
}
