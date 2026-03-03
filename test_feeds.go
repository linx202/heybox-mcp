// +build ignore

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/yourusername/heybox-mcp/browser"
	"github.com/yourusername/heybox-mcp/configs"
	"github.com/yourusername/heybox-mcp/heybox"
)

func main() {
	fmt.Println("=== 测试获取推荐动态列表 ===")

	// 创建浏览器
	b := browser.NewBrowser(
		true, // headless
		browser.WithBinPath(configs.GetBinPath()),
	)
	defer b.Close()

	page := b.MustPage("")
	defer page.Close()

	// 加载 Cookies
	loginAction := heybox.NewLogin(page)
	if err := loginAction.LoadCookies(); err != nil {
		fmt.Printf("加载 Cookies 失败: %v\n", err)
	} else {
		fmt.Println("✅ Cookies 加载成功")
	}

	// 创建动态列表操作
	feedsAction := heybox.NewFeedsAction(page)

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 获取推荐动态列表
	fmt.Println("正在获取推荐动态列表...")
	result, err := feedsAction.GetRecommendedFeeds(ctx, "", 10)
	if err != nil {
		fmt.Printf("❌ 获取动态列表失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 获取到 %d 条动态\n", len(result.Feeds))
	fmt.Printf("   HasMore: %v\n", result.HasMore)

	for i, feed := range result.Feeds {
		fmt.Printf("\n--- 动态 %d ---\n", i+1)
		fmt.Printf("标题: %s\n", feed.Title)
		if len(feed.Content) > 100 {
			fmt.Printf("内容: %s...\n", feed.Content[:100])
		} else {
			fmt.Printf("内容: %s\n", feed.Content)
		}
		fmt.Printf("作者: %s\n", feed.Author.Nickname)
		fmt.Printf("点赞数: %d, 评论数: %d\n", feed.LikeCount, feed.CommentCount)
		fmt.Printf("ID: %s\n", feed.ItemID)
		fmt.Printf("图片数: %d\n", len(feed.Images))
	}
}
