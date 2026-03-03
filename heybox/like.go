package heybox

import (
	"context"
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/sirupsen/logrus"
)

// LikeAction 点赞/收藏操作
type LikeAction struct {
	page      *rod.Page
	navigator *Navigator
}

// NewLikeAction 创建点赞操作实例
func NewLikeAction(page *rod.Page) *LikeAction {
	return &LikeAction{
		page:      page,
		navigator: NewNavigator(page),
	}
}

// LikeFeed 点赞动态
func (l *LikeAction) LikeFeed(ctx context.Context, feedID string) error {
	return l.toggleLike(ctx, feedID, true)
}

// UnlikeFeed 取消点赞
func (l *LikeAction) UnlikeFeed(ctx context.Context, feedID string) error {
	return l.toggleLike(ctx, feedID, false)
}

// toggleLike 切换点赞状态
func (l *LikeAction) toggleLike(ctx context.Context, feedID string, isLike bool) error {
	pp := l.page.Context(ctx).Timeout(60 * time.Second)

	// 导航到动态详情页
	detailURL := feedID
	if len(feedID) < 10 || feedID[:4] != "http" {
		detailURL = fmt.Sprintf("https://www.xiaoheihe.cn/app/bbs/link/%s", feedID)
	}

	if err := l.navigator.NavigateToFeed(ctx, detailURL); err != nil {
		return fmt.Errorf("导航失败: %w", err)
	}

	time.Sleep(2 * time.Second)

	// 查找点赞按钮
	likeBtnSelectors := []string{
		"[class*='like']",
		"[class*='zan']",
		"[class*='vote']",
		".like-btn",
		"button:contains('赞')",
	}

	for _, sel := range likeBtnSelectors {
		likeBtn, err := pp.Element(sel)
		if err == nil && likeBtn != nil {
			if err := likeBtn.Click(proto.InputMouseButtonLeft, 1); err == nil {
				action := "点赞"
				if !isLike {
					action = "取消点赞"
				}
				logrus.Infof("%s成功", action)
				time.Sleep(500 * time.Millisecond)
				return nil
			}
		}
	}

	return fmt.Errorf("未找到点赞按钮")
}

// FavoriteFeed 收藏动态
func (l *LikeAction) FavoriteFeed(ctx context.Context, feedID string) error {
	return l.toggleFavorite(ctx, feedID, true)
}

// UnfavoriteFeed 取消收藏
func (l *LikeAction) UnfavoriteFeed(ctx context.Context, feedID string) error {
	return l.toggleFavorite(ctx, feedID, false)
}

// toggleFavorite 切换收藏状态
func (l *LikeAction) toggleFavorite(ctx context.Context, feedID string, isFav bool) error {
	pp := l.page.Context(ctx).Timeout(60 * time.Second)

	// 导航到动态详情页
	detailURL := feedID
	if len(feedID) < 10 || feedID[:4] != "http" {
		detailURL = fmt.Sprintf("https://www.xiaoheihe.cn/app/bbs/link/%s", feedID)
	}

	if err := l.navigator.NavigateToFeed(ctx, detailURL); err != nil {
		return fmt.Errorf("导航失败: %w", err)
	}

	time.Sleep(2 * time.Second)

	// 查找收藏按钮
	favBtnSelectors := []string{
		"[class*='collect']",
		"[class*='favorite']",
		"[class*='star']",
		".fav-btn",
		"button:contains('收藏')",
	}

	for _, sel := range favBtnSelectors {
		favBtn, err := pp.Element(sel)
		if err == nil && favBtn != nil {
			if err := favBtn.Click(proto.InputMouseButtonLeft, 1); err == nil {
				action := "收藏"
				if !isFav {
					action = "取消收藏"
				}
				logrus.Infof("%s成功", action)
				time.Sleep(500 * time.Millisecond)
				return nil
			}
		}
	}

	return fmt.Errorf("未找到收藏按钮")
}
