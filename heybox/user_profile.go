package heybox

import (
	"context"
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/sirupsen/logrus"
)

// UserProfileAction 用户主页操作
type UserProfileAction struct {
	page      *rod.Page
	navigator *Navigator
}

// NewUserProfileAction 创建用户主页操作实例
func NewUserProfileAction(page *rod.Page) *UserProfileAction {
	return &UserProfileAction{
		page:      page,
		navigator: NewNavigator(page),
	}
}

// GetUserProfile 获取用户主页信息
func (u *UserProfileAction) GetUserProfile(ctx context.Context, userID string) (*UserProfile, error) {
	pp := u.page.Context(ctx).Timeout(60 * time.Second)

	// 构建用户主页 URL
	profileURL := userID
	if len(userID) < 10 || userID[:4] != "http" {
		profileURL = fmt.Sprintf("https://www.xiaoheihe.cn/app/user/profile/%s", userID)
	}

	logrus.Infof("获取用户主页: %s", profileURL)

	// 导航到用户主页
	if err := u.navigator.NavigateToUserProfile(ctx, profileURL); err != nil {
		return nil, fmt.Errorf("导航到用户主页失败: %w", err)
	}

	// 等待页面加载
	time.Sleep(3 * time.Second)

	// 使用 JavaScript 提取用户信息
	result, err := pp.Eval(`() => {
		const profile = {
			user_id: window.location.href,
			username: '',
			nickname: '',
			avatar: '',
			bio: '',
			followers: 0,
			following: 0,
			feeds: []
		};

		// 获取昵称
		const nicknameSelectors = ['.nickname', '.user-name', '[class*="nickname"]', '[class*="username"]', 'h1', 'h2'];
		for (const sel of nicknameSelectors) {
			const el = document.querySelector(sel);
			if (el && el.textContent.trim()) {
				profile.nickname = el.textContent.trim();
				break;
			}
		}

		// 获取头像
		const avatarEl = document.querySelector('img[class*="avatar"], img.avatar, [class*="avatar"] img');
		if (avatarEl && avatarEl.src) {
			profile.avatar = avatarEl.src;
		}

		// 获取简介
		const bioSelectors = ['.bio', '.desc', '[class*="bio"]', '[class*="intro"]'];
		for (const sel of bioSelectors) {
			const el = document.querySelector(sel);
			if (el && el.textContent.trim()) {
				profile.bio = el.textContent.trim();
				break;
			}
		}

		// 获取粉丝数和关注数
		const statEls = document.querySelectorAll('[class*="stat"], [class*="count"], [class*="num"]');
		const stats = [];
		statEls.forEach(el => {
			const num = parseInt(el.textContent.replace(/[^0-9]/g, ''));
			if (!isNaN(num) && num > 0) {
				stats.push(num);
			}
		});
		if (stats.length >= 1) profile.followers = stats[0];
		if (stats.length >= 2) profile.following = stats[1];

		// 获取用户动态列表
		const feedSelectors = ['.feed-item', '.post-item', '[class*="feed"]', '[class*="post"]'];
		let feedEls = [];
		for (const sel of feedSelectors) {
			const found = document.querySelectorAll(sel);
			if (found.length > feedEls.length) {
				feedEls = found;
			}
		}

		feedEls.forEach((el, i) => {
			if (i >= 10) return;
			const feed = {
				item_id: '',
				title: '',
				content: '',
				images: []
			};

			const titleEl = el.querySelector('h1, h2, h3, .title, [class*="title"]');
			if (titleEl) feed.title = titleEl.textContent.trim();

			const linkEl = el.querySelector('a[href*="/bbs/"], a[href*="/post/"]');
			if (linkEl) feed.item_id = linkEl.href;

			el.querySelectorAll('img').forEach(img => {
				if (img.src && !img.src.includes('avatar') && img.width > 50) {
					feed.images.push(img.src);
				}
			});

			if (feed.title) {
				profile.feeds.push(feed);
			}
		});

		return profile;
	}`)

	if err != nil {
		return nil, fmt.Errorf("提取用户信息失败: %w", err)
	}

	var profile UserProfile
	if err := result.Value.Unmarshal(&profile); err != nil {
		return nil, fmt.Errorf("解析用户信息失败: %w", err)
	}

	logrus.Infof("获取用户信息成功: %s", profile.Nickname)
	return &profile, nil
}
