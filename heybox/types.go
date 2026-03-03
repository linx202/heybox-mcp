package heybox

// UserInfo 用户信息
type UserInfo struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// FeedItem 动态项
type FeedItem struct {
	ItemID       string   `json:"item_id"`       // 动态 ID
	Title        string   `json:"title"`         // 标题
	Content      string   `json:"content"`       // 内容
	Images       []string `json:"images"`        // 图片列表
	Author       UserInfo `json:"author"`        // 作者信息
	LikeCount    int64    `json:"like_count"`    // 点赞数
	CommentCount int64    `json:"comment_count"` // 评论数
	CreatedAt    string   `json:"created_at"`    // 创建时间
}

// FeedListResponse 动态列表响应
type FeedListResponse struct {
	Feeds      []FeedItem `json:"feeds"`
	HasMore    bool       `json:"has_more"`
	NextCursor string     `json:"next_cursor"`
}

// Comment 评论
type Comment struct {
	CommentID   string     `json:"comment_id"`   // 评论 ID
	Content     string     `json:"content"`      // 评论内容
	Author      UserInfo   `json:"author"`       // 评论作者
	LikeCount   int        `json:"like_count"`   // 点赞数
	CreatedAt   string     `json:"created_at"`   // 创建时间
	ReplyTo     *Comment   `json:"reply_to"`     // 回复的评论（如果是回复）
}

// PublishResult 发布结果
type PublishResult struct {
	Success bool   `json:"success"`
	ItemID  string `json:"item_id"`   // 发布成功的动态 ID
	Message string `json:"message"`   // 提示信息
}

// SearchResults 搜索结果
type SearchResults struct {
	Query  string     `json:"query"`   // 搜索关键词
	Feeds  []FeedItem `json:"feeds"`   // 动态结果
	Users  []UserInfo `json:"users"`   // 用户结果
	HasMore bool      `json:"has_more"`
}

// UserProfile 用户主页
type UserProfile struct {
	UserID      string    `json:"user_id"`
	Username    string    `json:"username"`
	Nickname    string    `json:"nickname"`
	Avatar      string    `json:"avatar"`
	Bio         string    `json:"bio"`          // 个人简介
	Followers   int       `json:"followers"`    // 粉丝数
	Following   int       `json:"following"`    // 关注数
	Feeds       []FeedItem `json:"feeds"`       // 用户动态
}
