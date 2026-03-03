package heybox

import (
	"context"
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/sirupsen/logrus"
)

// FeedDetailAction 动态详情操作
type FeedDetailAction struct {
	page      *rod.Page
	navigator *Navigator
}

// NewFeedDetailAction 创建动态详情操作实例
func NewFeedDetailAction(page *rod.Page) *FeedDetailAction {
	return &FeedDetailAction{
		page:      page,
		navigator: NewNavigator(page),
	}
}

// GetFeedDetail 获取动态详情
func (f *FeedDetailAction) GetFeedDetail(ctx context.Context, feedID string) (*FeedItem, error) {
	pp := f.page.Context(ctx).Timeout(90 * time.Second)

	logrus.Infof("获取动态详情: %s", feedID)

	// 导航到动态详情页
	if err := f.navigator.NavigateToFeed(ctx, feedID); err != nil {
		return nil, fmt.Errorf("导航到动态详情页失败: %w", err)
	}

	// 等待页面加载
	logrus.Info("等待页面内容加载...")
	time.Sleep(3 * time.Second)

	// 检查是否有验证码
	checkCaptcha, _ := pp.Eval(`() => {
		const html = document.body.innerHTML;
		return html.includes('captcha') ||
		       html.includes('验证') ||
		       html.includes('请按顺序点击') ||
		       document.body.innerText.includes('请按顺序点击');
	}`)
	hasCaptcha := false
	checkCaptcha.Value.Unmarshal(&hasCaptcha)

	if hasCaptcha {
		logrus.Warn("检测到验证码，请稍后重试或手动完成验证")
		return &FeedItem{
			ItemID: feedID,
			Title:  "触发验证码",
			Content: "请求过于频繁，触发了验证码校验。请稍等几分钟后再试，或使用非无头模式手动完成验证。",
		}, nil
	}

	// 等待加载动画消失（最多等待 20 秒）
	for i := 0; i < 20; i++ {
		hasLoading, _ := pp.Eval(`() => {
			const loading = document.querySelector('.hb-cpt__loading');
			if (!loading) return false;
			return window.getComputedStyle(loading).display !== 'none';
		}`)
		isLoading := false
		hasLoading.Value.Unmarshal(&isLoading)
		if !isLoading {
			logrus.Info("页面加载完成")
			break
		}
		time.Sleep(1 * time.Second)
	}

	time.Sleep(1 * time.Second)

	// 使用 JavaScript 提取动态详情
	result, err := pp.Eval(`() => {
		const feed = {
			item_id: window.location.href,
			title: '',
			content: '',
			images: [],
			author: { nickname: '', user_id: '', avatar: '' },
			like_count: 0,
			comment_count: 0,
			created_at: ''
		};

		// 解析数字
		const parseNum = (str) => {
			if (!str) return 0;
			const num = parseInt(str.toString().replace(/[^0-9]/g, ''));
			if (isNaN(num) || num > 10000000) return 0;
			return num;
		};

		// 获取主容器
		const container = document.querySelector('.hb-bbs-link__container');
		if (!container) return feed;

		// 获取标题 - 在容器内查找
		const titleEl = container.querySelector('h1, [class*="title"]');
		if (titleEl) {
			feed.title = titleEl.textContent.trim();
		}

		// 获取作者 - 查找用户链接
		const authorEl = container.querySelector('a[href*="/user/"]');
		if (authorEl) {
			feed.author.nickname = authorEl.textContent.trim().split(/\s+/)[0];
		}

		// 获取头像
		const avatarEl = container.querySelector('img[class*="avatar"], a[href*="/user/"] img');
		if (avatarEl && avatarEl.src) {
			feed.author.avatar = avatarEl.src;
		}

		// 获取正文内容 - 获取所有段落
		const paragraphs = container.querySelectorAll('p');
		if (paragraphs.length > 0) {
			const texts = [];
			paragraphs.forEach(p => {
				const text = p.textContent.trim();
				if (text.length > 5) texts.push(text);
			});
			feed.content = texts.join('\n').substring(0, 3000);
		}

		// 获取图片
		container.querySelectorAll('img').forEach(img => {
			if (img.src && !img.src.includes('avatar') && !img.src.includes('icon') && !img.src.includes('logo')) {
				if (feed.images.length < 20) feed.images.push(img.src);
			}
		});

		// 获取点赞数和评论数 - 查找操作按钮
		const actionBtns = container.querySelectorAll('button, [class*="btn"], [class*="action"]');
		const nums = [];
		actionBtns.forEach(btn => {
			const num = parseNum(btn.textContent);
			if (num > 0) nums.push(num);
		});
		if (nums.length >= 1) feed.like_count = nums[0];
		if (nums.length >= 3) feed.comment_count = nums[2];

		return feed;
	}`)

	if err != nil {
		return nil, fmt.Errorf("提取动态详情失败: %w", err)
	}

	var feed FeedItem
	if err := result.Value.Unmarshal(&feed); err != nil {
		return nil, fmt.Errorf("解析动态详情失败: %w", err)
	}

	logrus.Infof("获取动态详情成功: %s", feed.Title)
	return &feed, nil
}
