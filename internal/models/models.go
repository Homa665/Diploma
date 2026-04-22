package models

import "time"

type UserRole string

const (
	RoleUser    UserRole = "user"
	RolePremium UserRole = "premium"
	RoleExpert  UserRole = "expert"
	RoleAdmin   UserRole = "admin"
)

type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone"`
	Nickname     string    `json:"nickname"`
	PasswordHash string    `json:"-"`
	Name         string    `json:"name"`
	City         string    `json:"city"`
	Bio          string    `json:"bio"`
	Role         UserRole  `json:"role"`
	Interests    string    `json:"interests"`
	UserRole2    string    `json:"user_role2"`
	AvatarURL    string    `json:"avatar_url"`
	IsBlocked    bool      `json:"is_blocked"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Post struct {
	ID              int        `json:"id"`
	AuthorID        int        `json:"author_id"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	Category        string     `json:"category"`
	Tags            string     `json:"tags"`
	IsPinned        bool       `json:"is_pinned"`
	PinnedUntil     *time.Time `json:"pinned_until"`
	CommentsOff     bool       `json:"comments_off"`
	IsHidden        bool       `json:"is_hidden"`
	ViewCount       int        `json:"view_count"`
	AvgRating       float64    `json:"avg_rating"`
	RatingCount     int        `json:"rating_count"`
	IsPremium       bool       `json:"is_premium"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	AuthorNickname  string     `json:"author_nickname"`
	AuthorAvatarURL string     `json:"author_avatar_url"`
	AuthorRole      UserRole   `json:"author_role"`
	Files           []File     `json:"files"`
	LikeCount       int        `json:"like_count"`
	CommentCount    int        `json:"comment_count"`
}

type File struct {
	ID        int       `json:"id"`
	PostID    int       `json:"post_id"`
	Filename  string    `json:"filename"`
	FilePath  string    `json:"file_path"`
	FileType  string    `json:"file_type"`
	FileSize  int64     `json:"file_size"`
	CreatedAt time.Time `json:"created_at"`
}

type Comment struct {
	ID              int       `json:"id"`
	PostID          int       `json:"post_id"`
	AuthorID        int       `json:"author_id"`
	ParentID        *int      `json:"parent_id"`
	Content         string    `json:"content"`
	LikeCount       int       `json:"like_count"`
	CreatedAt       time.Time `json:"created_at"`
	AuthorNickname  string    `json:"author_nickname"`
	AuthorAvatarURL string    `json:"author_avatar_url"`
	AuthorRole      UserRole  `json:"author_role"`
	Replies         []Comment `json:"replies"`
	LikedByUser     bool      `json:"liked_by_user"`
}

type Rating struct {
	ID        int       `json:"id"`
	PostID    int       `json:"post_id"`
	UserID    int       `json:"user_id"`
	Score     int       `json:"score"`
	Review    string    `json:"review"`
	IsExpert  bool      `json:"is_expert"`
	CreatedAt time.Time `json:"created_at"`
}

type ChatMessage struct {
	ID              int       `json:"id"`
	SenderID        int       `json:"sender_id"`
	ReceiverID      int       `json:"receiver_id"`
	Content         string    `json:"content"`
	IsRead          bool      `json:"is_read"`
	CreatedAt       time.Time `json:"created_at"`
	SenderNickname  string    `json:"sender_nickname"`
	SenderAvatarURL string    `json:"sender_avatar_url"`
}

type Notification struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Type      string    `json:"type"`
	Content   string    `json:"content"`
	Link      string    `json:"link"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

type TeamRequest struct {
	ID          int       `json:"id"`
	PostID      int       `json:"post_id"`
	AuthorID    int       `json:"author_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Skills      string    `json:"skills"`
	RoleNeeded  string    `json:"role_needed"`
	IsOpen      bool      `json:"is_open"`
	CreatedAt   time.Time `json:"created_at"`
	PostTitle   string    `json:"post_title"`
}

type TeamResponse struct {
	ID            int       `json:"id"`
	RequestID     int       `json:"request_id"`
	UserID        int       `json:"user_id"`
	Message       string    `json:"message"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UserNickname  string    `json:"user_nickname"`
	UserAvatarURL string    `json:"user_avatar_url"`
}

type Complaint struct {
	ID             int       `json:"id"`
	AuthorID       int       `json:"author_id"`
	TargetType     string    `json:"target_type"`
	TargetID       int       `json:"target_id"`
	Category       string    `json:"category"`
	Description    string    `json:"description"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	AuthorNickname string    `json:"author_nickname"`
	TargetContent  string    `json:"target_content"`
	TargetNickname string    `json:"target_nickname"`
}

type Subscription struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Plan      string    `json:"plan"`
	Status    string    `json:"status"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type Friendship struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	FriendID  int       `json:"friend_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type Follow struct {
	ID          int       `json:"id"`
	FollowerID  int       `json:"follower_id"`
	FollowingID int       `json:"following_id"`
	CreatedAt   time.Time `json:"created_at"`
}

type Poll struct {
	ID        int          `json:"id"`
	PostID    int          `json:"post_id"`
	Question  string       `json:"question"`
	CreatedAt time.Time    `json:"created_at"`
	Options   []PollOption `json:"options"`
	VotedBy   int          `json:"voted_by"`
}

type PollOption struct {
	ID        int    `json:"id"`
	PollID    int    `json:"poll_id"`
	Text      string `json:"text"`
	VoteCount int    `json:"vote_count"`
}

type CommentLike struct {
	ID        int `json:"id"`
	CommentID int `json:"comment_id"`
	UserID    int `json:"user_id"`
}

type Repost struct {
	ID        int       `json:"id"`
	PostID    int       `json:"post_id"`
	UserID    int       `json:"user_id"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
}

type BlockedUser struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	BlockedID int       `json:"blocked_id"`
	CreatedAt time.Time `json:"created_at"`
}

type ExpertApplication struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	Portfolio   string    `json:"portfolio"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type ActivityLog struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	Action     string    `json:"action"`
	TargetType string    `json:"target_type"`
	TargetID   int       `json:"target_id"`
	Details    string    `json:"details"`
	IPAddress  string    `json:"ip_address"`
	CreatedAt  time.Time `json:"created_at"`
	Nickname   string    `json:"nickname"`
}

type UserBan struct {
	ID           int       `json:"id"`
	UserID       int       `json:"user_id"`
	AdminID      int       `json:"admin_id"`
	Reason       string    `json:"reason"`
	Restriction  string    `json:"restriction"`
	TargetLink   string    `json:"target_link"`
	AdminComment string    `json:"admin_comment"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
	Nickname     string    `json:"nickname"`
	IsActive     bool      `json:"is_active"`
}

type Project struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CoverURL    string    `json:"cover_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	PostCount   int       `json:"post_count"`
}

type ExpertConsultation struct {
	ID              int       `json:"id"`
	ExpertID        int       `json:"expert_id"`
	ClientID        int       `json:"client_id"`
	PostID          int       `json:"post_id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	Price           int       `json:"price"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	ExpertNickname  string    `json:"expert_nickname"`
	ClientNickname  string    `json:"client_nickname"`
	PostTitle       string    `json:"post_title"`
}

type ConsultationMessage struct {
	ID              int       `json:"id"`
	ConsultationID  int       `json:"consultation_id"`
	SenderID        int       `json:"sender_id"`
	Content         string    `json:"content"`
	FilePath        string    `json:"file_path"`
	FileName        string    `json:"file_name"`
	CreatedAt       time.Time `json:"created_at"`
	SenderNickname  string    `json:"sender_nickname"`
}

type AdminStats struct {
	TotalUsers        int `json:"total_users"`
	TotalPosts        int `json:"total_posts"`
	TotalComments     int `json:"total_comments"`
	TotalComplaints   int `json:"total_complaints"`
	PendingComplaints int `json:"pending_complaints"`
	PendingExpertApps int `json:"pending_expert_apps"`
	TotalExperts      int `json:"total_experts"`
	TotalPremium      int `json:"total_premium"`
	NewUsersToday     int `json:"new_users_today"`
	NewPostsToday     int `json:"new_posts_today"`
}
