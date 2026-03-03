package heybox

import (
	"context"
	"time"

	"github.com/go-rod/rod"
	"github.com/sirupsen/logrus"
	apperrors "github.com/yourusername/heybox-mcp/errors"
)

const (
	// 小黑盒相关 URL
	heyboxBaseURL     = "https://www.xiaoheihe.cn"
	heyboxExploreURL  = "https://www.xiaoheihe.cn/home"
	heyboxPublishURL  = "https://www.xiaoheihe.cn/publish"
	heyboxLoginURL    = "https://www.xiaoheihe.cn/login"
	heyboxCommunityURL = "https://xiaoheihe.cn/app/bbs/home"  // 社区页面
	heyboxUserURL     = "https://www.xiaoheihe.cn/user"
)

// Navigator 页面导航器
type Navigator struct {
	page *rod.Page
}

// NewNavigator 创建导航器
func NewNavigator(page *rod.Page) *Navigator {
	return &Navigator{page: page}
}

// NavigateToPage 导航到指定 URL
func (n *Navigator) NavigateToPage(ctx context.Context, url string) error {
	pp := n.page.Context(ctx).Timeout(60 * time.Second)

	// 注入反检测脚本
	_, _ = pp.EvalOnNewDocument(`() => {
		Object.defineProperty(navigator, 'webdriver', { get: () => undefined });
		Object.defineProperty(navigator, 'plugins', { get: () => [1, 2, 3, 4, 5] });
		Object.defineProperty(navigator, 'languages', { get: () => ['zh-CN', 'zh', 'en'] });
		window.chrome = { runtime: {} };
		const originalQuery = window.navigator.permissions.query;
		window.navigator.permissions.query = (parameters) => (
			parameters.name === 'notifications' ?
				Promise.resolve({ state: Notification.permission }) :
				originalQuery(parameters)
		);
	}`)

	// 随机延迟后导航（模拟人类行为）
	human := NewHumanBehavior(n.page)
	human.RandomDelay(500, 1500)

	if err := pp.Navigate(url); err != nil {
		return apperrors.ErrNetwork("导航失败", err)
	}

	pp.MustWaitLoad()

	// 模拟页面加载后的自然行为
	human.SimulatePageInteraction()

	return nil
}

// NavigateToHome 导航到首页
func (n *Navigator) NavigateToHome(ctx context.Context) error {
	logrus.Info("导航到小黑盒首页")
	return n.NavigateToPage(ctx, heyboxExploreURL)
}

// NavigateToCommunity 导航到社区页面
func (n *Navigator) NavigateToCommunity(ctx context.Context) error {
	logrus.Info("导航到小黑盒社区页面")
	return n.NavigateToPage(ctx, heyboxCommunityURL)
}

// NavigateToPublish 导航到发布页面
func (n *Navigator) NavigateToPublish(ctx context.Context) error {
	logrus.Info("导航到发布页面")
	return n.NavigateToPage(ctx, heyboxPublishURL)
}

// NavigateToFeed 导航到动态详情页
func (n *Navigator) NavigateToFeed(ctx context.Context, itemID string) error {
	// 如果已经是完整 URL，直接使用
	url := itemID
	if len(itemID) < 10 || itemID[:4] != "http" {
		url = "https://www.xiaoheihe.cn/app/bbs/link/" + itemID
	}
	logrus.Infof("导航到动态详情页: %s", url)
	return n.NavigateToPage(ctx, url)
}

// NavigateToUserProfile 导航到用户主页
func (n *Navigator) NavigateToUserProfile(ctx context.Context, userID string) error {
	url := heyboxBaseURL + "/user/" + userID
	logrus.Infof("导航到用户主页: %s", url)
	return n.NavigateToPage(ctx, url)
}

// GetCurrentURL 获取当前页面 URL
func (n *Navigator) GetCurrentURL() string {
	return n.page.MustInfo().URL
}

// WaitForElement 等待元素出现
func (n *Navigator) WaitForElement(ctx context.Context, selector string, timeout time.Duration) error {
	pp := n.page.Context(ctx).Timeout(timeout)

	// 使用 MustElements 等待元素出现
	_, err := pp.Element(selector)
	if err != nil {
		return apperrors.ErrElementNotFound(selector, err)
	}

	return nil
}
