package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/sirupsen/logrus"
	"github.com/yourusername/heybox-mcp/browser"
	"github.com/yourusername/heybox-mcp/configs"
	"github.com/yourusername/heybox-mcp/cookies"
	"github.com/yourusername/heybox-mcp/heybox"
	"github.com/yourusername/heybox-mcp/pkg/downloader"
)

// HeyboxService 小黑盒业务服务
type HeyboxService struct{}

// NewHeyboxService 创建小黑盒服务实例
func NewHeyboxService() *HeyboxService {
	return &HeyboxService{}
}

// LoginStatusResponse 登录状态响应
type LoginStatusResponse struct {
	IsLoggedIn bool   `json:"is_logged_in"`
	Username   string `json:"username,omitempty"`
}

// CheckLoginStatus 检查登录状态
func (s *HeyboxService) CheckLoginStatus(ctx context.Context) (*LoginStatusResponse, error) {
	b := newBrowser()
	defer b.Close()

	page := b.MustPage("")
	defer page.Close()

	loginAction := heybox.NewLogin(page)

	// 先尝试加载已保存的 Cookies
	if err := loginAction.LoadCookies(); err != nil {
		logrus.Warnf("加载 Cookies 失败: %v", err)
	}

	isLoggedIn, err := loginAction.CheckLoginStatus(ctx)
	if err != nil {
		return nil, err
	}

	return &LoginStatusResponse{
		IsLoggedIn: isLoggedIn,
		Username:   configs.GetUsername(),
	}, nil
}

// QrcodeResponse 二维码响应
type QrcodeResponse struct {
	QrcodeURL string `json:"qrcode_url"`
	AlreadyLoggedIn bool `json:"already_logged_in"`
}

// GetLoginQrcode 获取登录二维码
func (s *HeyboxService) GetLoginQrcode(ctx context.Context) (*QrcodeResponse, error) {
	b := newBrowser()
	defer b.Close()

	page := b.MustPage("")
	defer page.Close()

	loginAction := heybox.NewLogin(page)

	qrcodeURL, alreadyLoggedIn, err := loginAction.FetchQrcodeImage(ctx)
	if err != nil {
		return nil, err
	}

	return &QrcodeResponse{
		QrcodeURL: qrcodeURL,
		AlreadyLoggedIn: alreadyLoggedIn,
	}, nil
}

// DeleteCookies 删除 Cookies
func (s *HeyboxService) DeleteCookies(ctx context.Context) error {
	b := newBrowser()
	defer b.Close()

	page := b.MustPage("")
	defer page.Close()

	loginAction := heybox.NewLogin(page)
	return loginAction.ClearCookies()
}

// PublishRequest 发布请求
type PublishRequest struct {
	Title   string   `json:"title"`
	Content string   `json:"content"`
	Images  []string `json:"images"`
	Tags    []string `json:"tags,omitempty"`
}

// PublishResponse 发布响应
type PublishResponse struct {
	Success     bool   `json:"success"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	Images      int    `json:"images"`
	Message     string `json:"message"`
	NeedLogin   bool   `json:"need_login,omitempty"`   // 是否需要登录
	QrcodeURL   string `json:"qrcode_url,omitempty"`   // 二维码 URL
	QrcodeSaved bool   `json:"qrcode_saved,omitempty"` // 二维码是否已保存
}

// PublishContent 发布内容
func (s *HeyboxService) PublishContent(ctx context.Context, req *PublishRequest) (*PublishResponse, error) {
	// 参数校验
	if len(req.Title) == 0 && len(req.Content) == 0 {
		return nil, fmt.Errorf("标题和内容不能同时为空")
	}

	// 处理图片
	imagePaths, err := s.processImages(req.Images)
	if err != nil {
		return nil, err
	}

	// 构建发布内容
	content := heybox.PublishContent{
		Title:      req.Title,
		Content:    req.Content,
		Tags:       req.Tags,
		ImagePaths: imagePaths,
	}

	// 执行发布
	result, err := s.publishContent(ctx, content)
	if err != nil {
		return nil, err
	}

	// 如果返回了结果（可能需要登录），直接返回
	if result != nil {
		return result, nil
	}

	// 发布成功
	return &PublishResponse{
		Success: true,
		Title:   req.Title,
		Content: req.Content,
		Images:  len(imagePaths),
		Message: "发布完成",
	}, nil
}

// processImages 处理图片（下载或使用本地路径）
func (s *HeyboxService) processImages(images []string) ([]string, error) {
	if len(images) == 0 {
		return []string{}, nil
	}

	imageDownloader := downloader.NewImageDownloader(downloader.WithSaveDir(configs.GetImageSaveDir()))
	var imagePaths []string

	for _, img := range images {
		// 判断是 URL 还是本地路径
		if isURL(img) {
			// 下载图片
			path, err := imageDownloader.Download(context.Background(), img)
			if err != nil {
				logrus.Warnf("下载图片失败: %s, %v", img, err)
				continue
			}
			imagePaths = append(imagePaths, path)
		} else {
			// 使用本地路径
			imagePaths = append(imagePaths, img)
		}
	}

	return imagePaths, nil
}

// publishContent 执行发布
func (s *HeyboxService) publishContent(ctx context.Context, content heybox.PublishContent) (*PublishResponse, error) {
	b := newBrowser()
	defer b.Close()

	page := b.MustPage("")
	defer page.Close()

	// 加载 Cookies
	loginAction := heybox.NewLogin(page)
	if err := loginAction.LoadCookies(); err != nil {
		logrus.Warnf("加载 Cookies 失败: %v", err)
	}

	// 创建发布操作（会导航到发布页面）
	publishAction, err := heybox.NewPublishAction(page)
	if err != nil {
		return nil, err
	}

	// 检查登录状态，如果未登录则获取二维码
	loginResult, err := publishAction.CheckLoginAndFetchQrcode(ctx)
	if err != nil {
		return nil, err
	}

	// 如果需要登录，返回二维码信息
	if loginResult.NeedLogin {
		return &PublishResponse{
			Success:     false,
			NeedLogin:   true,
			QrcodeURL:   loginResult.QrcodeURL,
			QrcodeSaved: loginResult.QrcodeSaved,
			Message:     "需要登录，请扫描二维码后重试",
		}, nil
	}

	// 执行发布
	if err := publishAction.Publish(ctx, content); err != nil {
		return nil, err
	}

	return nil, nil
}

// FeedListResponse 动态列表响应
type FeedListResponse struct {
	Feeds      []heybox.FeedItem `json:"feeds"`
	HasMore    bool              `json:"has_more"`
	NextCursor string            `json:"next_cursor"`
}

// ListFeeds 获取动态列表
func (s *HeyboxService) ListFeeds(ctx context.Context, cursor string, limit int) (*FeedListResponse, error) {
	b := newBrowser()
	defer b.Close()

	page := b.MustPage("")
	defer page.Close()

	// 加载 Cookies
	loginAction := heybox.NewLogin(page)
	if err := loginAction.LoadCookies(); err != nil {
		logrus.Warnf("加载 Cookies 失败: %v", err)
	}

	// 创建动态列表操作
	feedsAction := heybox.NewFeedsAction(page)

	// 获取推荐动态列表
	result, err := feedsAction.GetRecommendedFeeds(ctx, cursor, limit)
	if err != nil {
		return nil, fmt.Errorf("获取动态列表失败: %w", err)
	}

	return &FeedListResponse{
		Feeds:      result.Feeds,
		HasMore:    result.HasMore,
		NextCursor: result.NextCursor,
	}, nil
}

// SearchRequest 搜索请求
type SearchRequest struct {
	Query string `json:"query"`
	Limit int    `json:"limit,omitempty"`
}

// SearchResponse 搜索响应
type SearchResponse struct {
	Query   string            `json:"query"`
	Feeds   []heybox.FeedItem `json:"feeds"`
	Users   []heybox.UserInfo `json:"users"`
	HasMore bool              `json:"has_more"`
}

// SearchFeeds 搜索内容
func (s *HeyboxService) SearchFeeds(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	b := newBrowser()
	defer b.Close()

	page := b.MustPage("")
	defer page.Close()

	// 加载 Cookies
	loginAction := heybox.NewLogin(page)
	if err := loginAction.LoadCookies(); err != nil {
		logrus.Warnf("加载 Cookies 失败: %v", err)
	}

	// 创建搜索操作
	searchAction := heybox.NewSearchAction(page)

	// 执行搜索
	searchReq := &heybox.SearchRequest{
		Query: req.Query,
		Limit: req.Limit,
	}
	result, err := searchAction.SearchFeeds(ctx, searchReq)
	if err != nil {
		return nil, fmt.Errorf("搜索失败: %w", err)
	}

	return &SearchResponse{
		Query:   result.Query,
		Feeds:   result.Feeds,
		Users:   result.Users,
		HasMore: result.HasMore,
	}, nil
}

// FeedDetailResponse 动态详情响应
type FeedDetailResponse struct {
	Feed *heybox.FeedItem `json:"feed"`
}

// GetFeedDetail 获取动态详情
func (s *HeyboxService) GetFeedDetail(ctx context.Context, feedID string) (*FeedDetailResponse, error) {
	b := newBrowser()
	defer b.Close()

	page := b.MustPage("")
	defer page.Close()

	// 加载 Cookies
	loginAction := heybox.NewLogin(page)
	if err := loginAction.LoadCookies(); err != nil {
		logrus.Warnf("加载 Cookies 失败: %v", err)
	}

	// 创建动态详情操作
	feedDetailAction := heybox.NewFeedDetailAction(page)

	// 获取动态详情
	feed, err := feedDetailAction.GetFeedDetail(ctx, feedID)
	if err != nil {
		return nil, fmt.Errorf("获取动态详情失败: %w", err)
	}

	return &FeedDetailResponse{
		Feed: feed,
	}, nil
}

// UserProfileResponse 用户主页响应
type UserProfileResponse struct {
	Profile *heybox.UserProfile `json:"profile"`
}

// GetUserProfile 获取用户信息
func (s *HeyboxService) GetUserProfile(ctx context.Context, userID string) (*UserProfileResponse, error) {
	b := newBrowser()
	defer b.Close()

	page := b.MustPage("")
	defer page.Close()

	// 加载 Cookies
	loginAction := heybox.NewLogin(page)
	if err := loginAction.LoadCookies(); err != nil {
		logrus.Warnf("加载 Cookies 失败: %v", err)
	}

	// 创建用户主页操作
	userProfileAction := heybox.NewUserProfileAction(page)

	// 获取用户信息
	profile, err := userProfileAction.GetUserProfile(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	return &UserProfileResponse{
		Profile: profile,
	}, nil
}

// CommentRequest 评论请求
type CommentRequest struct {
	FeedID  string `json:"feed_id"`
	Content string `json:"content"`
}

// CommentResponse 评论响应
type CommentResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// PostComment 发表评论
func (s *HeyboxService) PostComment(ctx context.Context, req *CommentRequest) (*CommentResponse, error) {
	b := newBrowser()
	defer b.Close()

	page := b.MustPage("")
	defer page.Close()

	// 加载 Cookies
	loginAction := heybox.NewLogin(page)
	if err := loginAction.LoadCookies(); err != nil {
		logrus.Warnf("加载 Cookies 失败: %v", err)
	}

	// 检查登录状态
	isLoggedIn, err := loginAction.CheckLoginStatus(ctx)
	if err != nil {
		return nil, err
	}
	if !isLoggedIn {
		return &CommentResponse{Success: false, Message: "未登录"}, nil
	}

	// 创建评论操作
	commentAction := heybox.NewCommentAction(page)

	// 发表评论
	if err := commentAction.PostComment(ctx, req.FeedID, req.Content); err != nil {
		return &CommentResponse{Success: false, Message: err.Error()}, nil
	}

	return &CommentResponse{Success: true, Message: "评论发表成功"}, nil
}

// ReplyRequest 回复请求
type ReplyRequest struct {
	FeedID    string `json:"feed_id"`
	CommentID string `json:"comment_id"`
	Content   string `json:"content"`
}

// ReplyComment 回复评论
func (s *HeyboxService) ReplyComment(ctx context.Context, req *ReplyRequest) (*CommentResponse, error) {
	b := newBrowser()
	defer b.Close()

	page := b.MustPage("")
	defer page.Close()

	// 加载 Cookies
	loginAction := heybox.NewLogin(page)
	if err := loginAction.LoadCookies(); err != nil {
		logrus.Warnf("加载 Cookies 失败: %v", err)
	}

	// 检查登录状态
	isLoggedIn, err := loginAction.CheckLoginStatus(ctx)
	if err != nil {
		return nil, err
	}
	if !isLoggedIn {
		return &CommentResponse{Success: false, Message: "未登录"}, nil
	}

	// 创建评论操作
	commentAction := heybox.NewCommentAction(page)

	// 回复评论
	if err := commentAction.ReplyComment(ctx, req.FeedID, req.CommentID, req.Content); err != nil {
		return &CommentResponse{Success: false, Message: err.Error()}, nil
	}

	return &CommentResponse{Success: true, Message: "回复发表成功"}, nil
}

// LikeRequest 点赞请求
type LikeRequest struct {
	FeedID string `json:"feed_id"`
}

// LikeResponse 点赞响应
type LikeResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// LikeFeed 点赞动态
func (s *HeyboxService) LikeFeed(ctx context.Context, req *LikeRequest) (*LikeResponse, error) {
	b := newBrowser()
	defer b.Close()

	page := b.MustPage("")
	defer page.Close()

	// 加载 Cookies
	loginAction := heybox.NewLogin(page)
	if err := loginAction.LoadCookies(); err != nil {
		logrus.Warnf("加载 Cookies 失败: %v", err)
	}

	// 检查登录状态
	isLoggedIn, err := loginAction.CheckLoginStatus(ctx)
	if err != nil {
		return nil, err
	}
	if !isLoggedIn {
		return &LikeResponse{Success: false, Message: "未登录"}, nil
	}

	// 创建点赞操作
	likeAction := heybox.NewLikeAction(page)

	// 执行点赞
	if err := likeAction.LikeFeed(ctx, req.FeedID); err != nil {
		return &LikeResponse{Success: false, Message: err.Error()}, nil
	}

	return &LikeResponse{Success: true, Message: "点赞成功"}, nil
}

// FavoriteFeed 收藏动态
func (s *HeyboxService) FavoriteFeed(ctx context.Context, req *LikeRequest) (*LikeResponse, error) {
	b := newBrowser()
	defer b.Close()

	page := b.MustPage("")
	defer page.Close()

	// 加载 Cookies
	loginAction := heybox.NewLogin(page)
	if err := loginAction.LoadCookies(); err != nil {
		logrus.Warnf("加载 Cookies 失败: %v", err)
	}

	// 检查登录状态
	isLoggedIn, err := loginAction.CheckLoginStatus(ctx)
	if err != nil {
		return nil, err
	}
	if !isLoggedIn {
		return &LikeResponse{Success: false, Message: "未登录"}, nil
	}

	// 创建点赞操作
	likeAction := heybox.NewLikeAction(page)

	// 执行收藏
	if err := likeAction.FavoriteFeed(ctx, req.FeedID); err != nil {
		return &LikeResponse{Success: false, Message: err.Error()}, nil
	}

	return &LikeResponse{Success: true, Message: "收藏成功"}, nil
}

// newBrowser 创建浏览器实例
func newBrowser() *rod.Browser {
	return browser.NewBrowser(
		configs.IsHeadless(),
		browser.WithBinPath(configs.GetBinPath()),
	)
}

// saveCookies 保存 Cookies
func saveCookies(page *rod.Page) error {
	result, err := proto.NetworkGetCookies{}.Call(page.Browser())
	if err != nil {
		return err
	}

	data, err := json.Marshal(result.Cookies)
	if err != nil {
		return err
	}

	cookieLoader := cookies.NewLoadCookie(cookies.GetCookiesFilePath())
	return cookieLoader.SaveCookies(data)
}

// isURL 判断是否是 URL
func isURL(s string) bool {
	return len(s) > 4 && (s[:4] == "http" || s[:5] == "https")
}
