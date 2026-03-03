package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sirupsen/logrus"
)

// InitMCPServer 初始化 MCP Server
func InitMCPServer(service *HeyboxService) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "heybox-mcp",
		Version: "1.0.0",
	}, nil)

	registerTools(server, service)

	logrus.Info("MCP Server initialized")
	return server
}

// ToolOutput 通用工具输出
type ToolOutput struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

// registerTools 注册 MCP 工具
func registerTools(server *mcp.Server, service *HeyboxService) {
	// 工具 1: 检查登录状态
	mcp.AddTool(server, &mcp.Tool{
		Name:        "check_login_status",
		Description: "检查小黑盒登录状态，返回是否已登录及用户名",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, ToolOutput, error) {
		result, err := service.CheckLoginStatus(ctx)
		if err != nil {
			return nil, ToolOutput{}, fmt.Errorf("检查登录状态失败: %v", err)
		}
		msg := fmt.Sprintf("登录状态: %v, 用户名: %s", result.IsLoggedIn, result.Username)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		}, ToolOutput{Message: msg, Success: true}, nil
	})

	// 工具 2: 获取登录二维码
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_login_qrcode",
		Description: "获取小黑盒登录二维码，返回二维码图片URL",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, ToolOutput, error) {
		result, err := service.GetLoginQrcode(ctx)
		if err != nil {
			return nil, ToolOutput{}, fmt.Errorf("获取二维码失败: %v", err)
		}
		var msg string
		if result.AlreadyLoggedIn {
			msg = "已经登录，无需扫码"
		} else {
			msg = fmt.Sprintf("请扫描二维码登录: %s", result.QrcodeURL)
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		}, ToolOutput{Message: msg, Success: true}, nil
	})

	// 工具 3: 删除 Cookies
	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_cookies",
		Description: "删除保存的小黑盒登录 Cookies，退出登录",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, ToolOutput, error) {
		err := service.DeleteCookies(ctx)
		if err != nil {
			return nil, ToolOutput{}, fmt.Errorf("删除 Cookies 失败: %v", err)
		}
		msg := "Cookies 已删除，已退出登录"
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		}, ToolOutput{Message: msg, Success: true}, nil
	})

	// 工具 4: 发布内容
	mcp.AddTool(server, &mcp.Tool{
		Name:        "publish_content",
		Description: "发布小黑盒动态，支持文字和图片",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args PublishRequest) (*mcp.CallToolResult, ToolOutput, error) {
		result, err := service.PublishContent(ctx, &args)
		if err != nil {
			return nil, ToolOutput{}, fmt.Errorf("发布失败: %v", err)
		}
		msg := fmt.Sprintf("发布成功! 标题: %s, 图片数: %d, 消息: %s",
			result.Title, result.Images, result.Message)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		}, ToolOutput{Message: msg, Success: result.Success}, nil
	})

	// 工具 5: 发布视频（暂未实现）
	mcp.AddTool(server, &mcp.Tool{
		Name:        "publish_with_video",
		Description: "发布小黑盒动态（带视频），功能尚未实现",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args struct {
		Title      string `json:"title"`
		Content    string `json:"content"`
		VideoPath  string `json:"video_path"`
		CoverImage string `json:"cover_image,omitempty"`
	}) (*mcp.CallToolResult, ToolOutput, error) {
		msg := "视频发布功能尚未实现"
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: msg}},
			IsError: true,
		}, ToolOutput{Message: msg, Success: false}, nil
	})

	// 工具 6: 获取动态列表
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_feeds",
		Description: "获取小黑盒动态列表",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args struct {
		Cursor string `json:"cursor,omitempty"`
		Limit  int    `json:"limit,omitempty"`
	}) (*mcp.CallToolResult, *FeedListResponse, error) {
		if args.Limit == 0 {
			args.Limit = 20
		}
		result, err := service.ListFeeds(ctx, args.Cursor, args.Limit)
		if err != nil {
			return nil, nil, fmt.Errorf("获取动态列表失败: %v", err)
		}
		// 序列化完整数据为 JSON
		jsonData, _ := json.MarshalIndent(result, "", "  ")
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(jsonData)}},
		}, result, nil
	})

	// 工具 7: 搜索内容
	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_feeds",
		Description: "搜索小黑盒内容",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args SearchRequest) (*mcp.CallToolResult, *SearchResponse, error) {
		if args.Limit == 0 {
			args.Limit = 20
		}
		result, err := service.SearchFeeds(ctx, &args)
		if err != nil {
			return nil, nil, fmt.Errorf("搜索失败: %v", err)
		}
		// 序列化完整数据为 JSON
		jsonData, _ := json.MarshalIndent(result, "", "  ")
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(jsonData)}},
		}, result, nil
	})

	// 工具 8: 获取动态详情
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_feed_detail",
		Description: "获取小黑盒动态详情",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args struct {
		FeedID string `json:"feed_id"`
	}) (*mcp.CallToolResult, *FeedDetailResponse, error) {
		result, err := service.GetFeedDetail(ctx, args.FeedID)
		if err != nil {
			return nil, nil, fmt.Errorf("获取详情失败: %v", err)
		}
		// 序列化完整数据为 JSON
		jsonData, _ := json.MarshalIndent(result, "", "  ")
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(jsonData)}},
		}, result, nil
	})

	// 工具 9: 获取用户信息
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_user_profile",
		Description: "获取小黑盒用户主页信息",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args struct {
		UserID string `json:"user_id"`
	}) (*mcp.CallToolResult, ToolOutput, error) {
		result, err := service.GetUserProfile(ctx, args.UserID)
		if err != nil {
			return nil, ToolOutput{}, fmt.Errorf("获取用户信息失败: %v", err)
		}
		var msg string
		if result.Profile == nil {
			msg = "用户不存在或功能尚未实现"
		} else {
			msg = fmt.Sprintf("用户: %s (%s)", result.Profile.Nickname, result.Profile.Username)
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		}, ToolOutput{Message: msg, Success: result.Profile != nil}, nil
	})

	// 工具 10: 发表评论
	mcp.AddTool(server, &mcp.Tool{
		Name:        "post_comment_to_feed",
		Description: "对小黑盒动态发表评论",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args CommentRequest) (*mcp.CallToolResult, ToolOutput, error) {
		result, err := service.PostComment(ctx, &args)
		if err != nil {
			return nil, ToolOutput{}, fmt.Errorf("发表评论失败: %v", err)
		}
		msg := fmt.Sprintf("评论结果: %s", result.Message)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		}, ToolOutput{Message: msg, Success: result.Success}, nil
	})

	// 工具 11: 回复评论
	mcp.AddTool(server, &mcp.Tool{
		Name:        "reply_comment_in_feed",
		Description: "回复小黑盒动态的评论",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ReplyRequest) (*mcp.CallToolResult, ToolOutput, error) {
		result, err := service.ReplyComment(ctx, &args)
		if err != nil {
			return nil, ToolOutput{}, fmt.Errorf("回复评论失败: %v", err)
		}
		msg := fmt.Sprintf("回复结果: %s", result.Message)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		}, ToolOutput{Message: msg, Success: result.Success}, nil
	})

	// 工具 12: 点赞
	mcp.AddTool(server, &mcp.Tool{
		Name:        "like_feed",
		Description: "点赞小黑盒动态",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args LikeRequest) (*mcp.CallToolResult, ToolOutput, error) {
		result, err := service.LikeFeed(ctx, &args)
		if err != nil {
			return nil, ToolOutput{}, fmt.Errorf("点赞失败: %v", err)
		}
		msg := fmt.Sprintf("点赞结果: %s", result.Message)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		}, ToolOutput{Message: msg, Success: result.Success}, nil
	})

	// 工具 13: 收藏
	mcp.AddTool(server, &mcp.Tool{
		Name:        "favorite_feed",
		Description: "收藏小黑盒动态",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args LikeRequest) (*mcp.CallToolResult, ToolOutput, error) {
		result, err := service.FavoriteFeed(ctx, &args)
		if err != nil {
			return nil, ToolOutput{}, fmt.Errorf("收藏失败: %v", err)
		}
		msg := fmt.Sprintf("收藏结果: %s", result.Message)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		}, ToolOutput{Message: msg, Success: result.Success}, nil
	})

	logrus.Info("Registered 13 MCP tools")
}
